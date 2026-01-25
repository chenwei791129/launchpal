package backup

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Backup represents a single backup entry
type Backup struct {
	ID        string    `json:"id"`
	Service   string    `json:"service"`
	Timestamp time.Time `json:"timestamp"`
	Path      string    `json:"path"`
}

// BackupManager handles backup operations for service plists
type BackupManager struct {
	baseDir string
}

// NewBackupManager creates a new BackupManager with default base directory
func NewBackupManager() *BackupManager {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}
	return &BackupManager{
		baseDir: filepath.Join(home, ".launchpal", "backups"),
	}
}

// getBackupDir returns the backup directory for a specific service
func (m *BackupManager) getBackupDir(serviceName string) string {
	return filepath.Join(m.baseDir, serviceName)
}

// Create creates a backup of the given plist file
func (m *BackupManager) Create(serviceName, plistPath string) (*Backup, error) {
	// Ensure the backup directory exists
	backupDir := m.getBackupDir(serviceName)
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create backup directory: %w", err)
	}

	// Generate backup ID with timestamp
	timestamp := time.Now()
	id := timestamp.Format("20060102-150405")
	backupPath := filepath.Join(backupDir, id+".plist")

	// Copy the plist file to backup location
	if err := copyFile(plistPath, backupPath); err != nil {
		return nil, fmt.Errorf("failed to copy plist to backup: %w", err)
	}

	// Prune old backups to keep only the 10 most recent
	m.pruneBackups(serviceName)

	return &Backup{
		ID:        id,
		Service:   serviceName,
		Timestamp: timestamp,
		Path:      backupPath,
	}, nil
}

// List returns all backups for a service sorted by time (newest first)
func (m *BackupManager) List(serviceName string) ([]Backup, error) {
	backupDir := m.getBackupDir(serviceName)

	entries, err := os.ReadDir(backupDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []Backup{}, nil
		}
		return nil, fmt.Errorf("failed to read backup directory: %w", err)
	}

	var backups []Backup
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".plist") {
			continue
		}

		id := strings.TrimSuffix(entry.Name(), ".plist")
		timestamp, err := time.Parse("20060102-150405", id)
		if err != nil {
			// Skip files that don't match our naming convention
			continue
		}

		backups = append(backups, Backup{
			ID:        id,
			Service:   serviceName,
			Timestamp: timestamp,
			Path:      filepath.Join(backupDir, entry.Name()),
		})
	}

	// Sort by timestamp descending (newest first)
	sort.Slice(backups, func(i, j int) bool {
		return backups[i].Timestamp.After(backups[j].Timestamp)
	})

	return backups, nil
}

// Get returns a specific backup by service name and backup ID
func (m *BackupManager) Get(serviceName, backupID string) (*Backup, error) {
	backupPath := filepath.Join(m.getBackupDir(serviceName), backupID+".plist")

	info, err := os.Stat(backupPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("backup not found: %s", backupID)
		}
		return nil, fmt.Errorf("failed to stat backup: %w", err)
	}

	timestamp, err := time.Parse("20060102-150405", backupID)
	if err != nil {
		timestamp = info.ModTime()
	}

	return &Backup{
		ID:        backupID,
		Service:   serviceName,
		Timestamp: timestamp,
		Path:      backupPath,
	}, nil
}

// GetContent returns the content of a backup file
func (m *BackupManager) GetContent(serviceName, backupID string) (string, error) {
	backup, err := m.Get(serviceName, backupID)
	if err != nil {
		return "", err
	}

	content, err := os.ReadFile(backup.Path)
	if err != nil {
		return "", fmt.Errorf("failed to read backup content: %w", err)
	}

	return string(content), nil
}

// Restore restores a backup to the target path
func (m *BackupManager) Restore(serviceName, backupID, targetPath string) error {
	backup, err := m.Get(serviceName, backupID)
	if err != nil {
		return err
	}

	if err := copyFile(backup.Path, targetPath); err != nil {
		return fmt.Errorf("failed to restore backup: %w", err)
	}

	return nil
}

// pruneBackups keeps only the 10 most recent backups for a service
func (m *BackupManager) pruneBackups(serviceName string) error {
	backups, err := m.List(serviceName)
	if err != nil {
		return err
	}

	// Keep only the 10 most recent backups
	const maxBackups = 10
	if len(backups) <= maxBackups {
		return nil
	}

	// Remove older backups
	for _, backup := range backups[maxBackups:] {
		if err := os.Remove(backup.Path); err != nil {
			// Log but don't fail on cleanup errors
			continue
		}
	}

	return nil
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	// Get source file info for permissions
	srcInfo, err := srcFile.Stat()
	if err != nil {
		return err
	}

	dstFile, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, srcInfo.Mode())
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}
