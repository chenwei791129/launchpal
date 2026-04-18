package backup

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestBackupManager_GetBackupDir(t *testing.T) {
	m := NewBackupManager()
	dir := m.getBackupDir("com.test.service")

	home, _ := os.UserHomeDir()
	expected := filepath.Join(home, ".launchpal", "backups", "com.test.service")

	if dir != expected {
		t.Errorf("getBackupDir() = %v, want %v", dir, expected)
	}
}

func TestBackupManager_Create(t *testing.T) {
	tmpDir := t.TempDir()
	m := &BackupManager{baseDir: tmpDir}

	// Create a temp plist file as the source
	plistContent := `<?xml version="1.0" encoding="UTF-8"?>
<plist version="1.0">
<dict><key>Label</key><string>com.test.service</string></dict>
</plist>`
	srcFile := filepath.Join(t.TempDir(), "com.test.service.plist")
	if err := os.WriteFile(srcFile, []byte(plistContent), 0644); err != nil {
		t.Fatalf("failed to create source plist: %v", err)
	}

	backup, err := m.Create("com.test.service", srcFile)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	// Verify returned Backup is non-nil with correct service name
	if backup == nil {
		t.Fatal("Create() returned nil backup")
	}
	if backup.Service != "com.test.service" {
		t.Errorf("Create() Service = %v, want %v", backup.Service, "com.test.service")
	}

	// Verify .plist file exists in backup dir
	backupDir := filepath.Join(tmpDir, "com.test.service")
	plistPath := filepath.Join(backupDir, backup.ID+".plist")
	if _, err := os.Stat(plistPath); os.IsNotExist(err) {
		t.Errorf("backup plist file does not exist at %v", plistPath)
	}

	// Verify .meta.json file exists and contains the original path
	metaPath := filepath.Join(backupDir, backup.ID+".meta.json")
	metaData, err := os.ReadFile(metaPath)
	if err != nil {
		t.Fatalf("failed to read meta file: %v", err)
	}
	var meta struct {
		OriginalPath string `json:"originalPath"`
	}
	if err := json.Unmarshal(metaData, &meta); err != nil {
		t.Fatalf("failed to parse meta json: %v", err)
	}
	if meta.OriginalPath != srcFile {
		t.Errorf("meta.OriginalPath = %v, want %v", meta.OriginalPath, srcFile)
	}
}

func TestBackupManager_List(t *testing.T) {
	tmpDir := t.TempDir()
	m := &BackupManager{baseDir: tmpDir}

	// Create backup files directly with synthetic timestamps to avoid time.Sleep
	backupDir := filepath.Join(tmpDir, "com.test.list")
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		t.Fatalf("failed to create backup dir: %v", err)
	}

	base := time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC)
	ids := make([]string, 3)
	for i := range 3 {
		ts := base.Add(time.Duration(i) * time.Second)
		ids[i] = ts.Format("20060102-150405")
		if err := os.WriteFile(filepath.Join(backupDir, ids[i]+".plist"), []byte("<plist/>"), 0644); err != nil {
			t.Fatalf("failed to write backup #%d: %v", i, err)
		}
	}

	backups, err := m.List("com.test.list")
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if len(backups) != 3 {
		t.Fatalf("List() returned %d backups, want 3", len(backups))
	}

	// Verify sorted newest first
	if backups[0].ID != ids[2] {
		t.Errorf("backups[0].ID = %s, want %s (newest)", backups[0].ID, ids[2])
	}
	if backups[2].ID != ids[0] {
		t.Errorf("backups[2].ID = %s, want %s (oldest)", backups[2].ID, ids[0])
	}

	// Verify empty list for nonexistent service (no error)
	emptyList, err := m.List("com.nonexistent.service")
	if err != nil {
		t.Fatalf("List() for nonexistent service error = %v", err)
	}
	if len(emptyList) != 0 {
		t.Errorf("List() for nonexistent service returned %d backups, want 0", len(emptyList))
	}
}

func TestBackupManager_GetContent_XMLPlist(t *testing.T) {
	tmpDir := t.TempDir()
	m := &BackupManager{baseDir: tmpDir}

	plistContent := `<?xml version="1.0" encoding="UTF-8"?>
<plist version="1.0">
<dict><key>Label</key><string>com.test.content</string></dict>
</plist>`
	srcFile := filepath.Join(t.TempDir(), "test.plist")
	if err := os.WriteFile(srcFile, []byte(plistContent), 0644); err != nil {
		t.Fatalf("failed to create source plist: %v", err)
	}

	backup, err := m.Create("com.test.content", srcFile)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	got, err := m.GetContent("com.test.content", backup.ID)
	if err != nil {
		t.Fatalf("GetContent() error = %v", err)
	}
	if got.Data != plistContent {
		t.Errorf("Data = %q, want %q", got.Data, plistContent)
	}
	if got.Format != "xml" {
		t.Errorf("Format = %q, want xml", got.Format)
	}
	if got.ConvertFailed {
		t.Errorf("ConvertFailed = true, want false")
	}

	_, err = m.GetContent("com.test.content", "99999999-999999")
	if err == nil {
		t.Error("GetContent() with nonexistent ID should return error")
	}
}

func TestBackupManager_GetContent_BinaryPlistConvertedToXML(t *testing.T) {
	tmpDir := t.TempDir()
	m := &BackupManager{baseDir: tmpDir}

	xmlSource := `<?xml version="1.0" encoding="UTF-8"?>
<plist version="1.0">
<dict><key>Label</key><string>com.test.binary.backup</string></dict>
</plist>`
	xmlPath := filepath.Join(t.TempDir(), "source.plist")
	if err := os.WriteFile(xmlPath, []byte(xmlSource), 0644); err != nil {
		t.Fatalf("write xml: %v", err)
	}
	binaryPath := filepath.Join(t.TempDir(), "binary.plist")
	cmd := exec.Command("plutil", "-convert", "binary1", "-o", binaryPath, xmlPath)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Skipf("plutil not available: %v (output: %s)", err, out)
	}

	backup, err := m.Create("com.test.binary", binaryPath)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	got, err := m.GetContent("com.test.binary", backup.ID)
	if err != nil {
		t.Fatalf("GetContent() error = %v", err)
	}
	if got.Format != "binary" {
		t.Errorf("Format = %q, want binary", got.Format)
	}
	if got.ConvertFailed {
		t.Errorf("ConvertFailed = true, want false")
	}
	if !strings.Contains(got.Data, "com.test.binary.backup") {
		t.Errorf("converted Data missing expected label: %q", got.Data)
	}
	if !strings.Contains(got.Data, "<?xml") {
		t.Errorf("converted Data not XML; got prefix: %q", got.Data)
	}
}

func TestBackupManager_GetContent_CorruptedBinaryFallsBackToRaw(t *testing.T) {
	tmpDir := t.TempDir()
	m := &BackupManager{baseDir: tmpDir}

	corrupt := append([]byte("bplist00"), 0xFF, 0xFE, 0xFD, 0xFC, 0xFB, 0xFA)
	srcFile := filepath.Join(t.TempDir(), "corrupt.plist")
	if err := os.WriteFile(srcFile, corrupt, 0644); err != nil {
		t.Fatalf("write corrupt: %v", err)
	}

	backup, err := m.Create("com.test.corrupt", srcFile)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	got, err := m.GetContent("com.test.corrupt", backup.ID)
	if err != nil {
		t.Fatalf("GetContent() error = %v", err)
	}
	if got.Format != "binary" {
		t.Errorf("Format = %q, want binary", got.Format)
	}
	if !got.ConvertFailed {
		t.Errorf("ConvertFailed = false, want true")
	}
	if got.Data != string(corrupt) {
		t.Errorf("fallback Data does not match raw bytes")
	}
}

func TestBackupManager_Restore(t *testing.T) {
	tmpDir := t.TempDir()
	m := &BackupManager{baseDir: tmpDir}

	originalContent := `<?xml version="1.0" encoding="UTF-8"?>
<plist version="1.0">
<dict><key>Label</key><string>com.test.restore</string></dict>
</plist>`
	srcFile := filepath.Join(t.TempDir(), "test.plist")
	if err := os.WriteFile(srcFile, []byte(originalContent), 0644); err != nil {
		t.Fatalf("failed to create source plist: %v", err)
	}

	backup, err := m.Create("com.test.restore", srcFile)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	// Restore to a new path and verify content matches
	restorePath := filepath.Join(t.TempDir(), "restored.plist")
	if err := m.Restore("com.test.restore", backup.ID, restorePath); err != nil {
		t.Fatalf("Restore() error = %v", err)
	}

	restoredContent, err := os.ReadFile(restorePath)
	if err != nil {
		t.Fatalf("failed to read restored file: %v", err)
	}
	if string(restoredContent) != originalContent {
		t.Errorf("restored content = %v, want %v", string(restoredContent), originalContent)
	}

	// Verify error for nonexistent backup ID
	err = m.Restore("com.test.restore", "99999999-999999", restorePath)
	if err == nil {
		t.Error("Restore() with nonexistent ID should return error")
	}
}

func TestBackupManager_Prune(t *testing.T) {
	tmpDir := t.TempDir()
	m := &BackupManager{baseDir: tmpDir}

	// Create 12 backup files directly with synthetic timestamps to avoid time.Sleep
	backupDir := filepath.Join(tmpDir, "com.test.prune")
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		t.Fatalf("failed to create backup dir: %v", err)
	}

	base := time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC)
	allIDs := make([]string, 12)
	for i := range 12 {
		ts := base.Add(time.Duration(i) * time.Second)
		allIDs[i] = ts.Format("20060102-150405")
		plistPath := filepath.Join(backupDir, allIDs[i]+".plist")
		metaPath := filepath.Join(backupDir, allIDs[i]+".meta.json")
		if err := os.WriteFile(plistPath, []byte("<plist/>"), 0644); err != nil {
			t.Fatalf("failed to write backup #%d: %v", i, err)
		}
		if err := os.WriteFile(metaPath, []byte(fmt.Sprintf(`{"originalPath":"/tmp/test-%d.plist"}`, i)), 0644); err != nil {
			t.Fatalf("failed to write meta #%d: %v", i, err)
		}
	}

	// Trigger prune
	if err := m.pruneBackups("com.test.prune"); err != nil {
		t.Fatalf("pruneBackups() error = %v", err)
	}

	// Verify only 10 backups remain
	backups, err := m.List("com.test.prune")
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(backups) != 10 {
		t.Fatalf("after pruning 12 backups, List() returned %d, want 10", len(backups))
	}

	// Verify the 2 oldest are removed (along with their meta files)
	for _, oldID := range allIDs[:2] {
		plistPath := filepath.Join(backupDir, oldID+".plist")
		if _, err := os.Stat(plistPath); !os.IsNotExist(err) {
			t.Errorf("oldest backup %s should have been pruned", oldID)
		}
		metaPath := filepath.Join(backupDir, oldID+".meta.json")
		if _, err := os.Stat(metaPath); !os.IsNotExist(err) {
			t.Errorf("metadata for %s should have been pruned", oldID)
		}
	}
}

func TestBackupManager_ListAll(t *testing.T) {
	tmpDir := t.TempDir()
	m := &BackupManager{baseDir: tmpDir}

	// Create backups for two different services with synthetic timestamps
	base := time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC)
	services := []string{"com.test.svc1", "com.test.svc2"}
	for i, svc := range services {
		backupDir := filepath.Join(tmpDir, svc)
		if err := os.MkdirAll(backupDir, 0755); err != nil {
			t.Fatalf("failed to create dir for %s: %v", svc, err)
		}
		for j := range 2 {
			ts := base.Add(time.Duration(i*10+j) * time.Second)
			id := ts.Format("20060102-150405")
			if err := os.WriteFile(filepath.Join(backupDir, id+".plist"), []byte("<plist/>"), 0644); err != nil {
				t.Fatalf("failed to write backup: %v", err)
			}
		}
	}

	all, err := m.ListAll()
	if err != nil {
		t.Fatalf("ListAll() error = %v", err)
	}
	if len(all) != 4 {
		t.Fatalf("ListAll() returned %d backups, want 4", len(all))
	}

	// Verify globally sorted newest first
	for i := 1; i < len(all); i++ {
		if all[i].Timestamp.After(all[i-1].Timestamp) {
			t.Errorf("ListAll() not sorted: [%d]=%v after [%d]=%v", i, all[i].Timestamp, i-1, all[i-1].Timestamp)
		}
	}
}
