package main

import (
	"context"
	"fmt"
	"os"

	"launchpal/internal/backup"
	"launchpal/internal/launchctl"
)

// App struct
type App struct {
	ctx            context.Context
	version        string
	manager        *launchctl.UserManager
	systemManager  *launchctl.SystemManager
	appleSystemMgr *launchctl.AppleSystemManager
	backup         *backup.BackupManager
}

// NewApp creates a new App application struct with default version
func NewApp() *App {
	return NewAppWithVersion("dev")
}

// NewAppWithVersion creates a new App with the specified version
func NewAppWithVersion(version string) *App {
	return &App{
		version:        version,
		manager:        launchctl.NewUserManager(),
		systemManager:  launchctl.NewSystemManager(),
		appleSystemMgr: launchctl.NewAppleSystemManager(),
		backup:         backup.NewBackupManager(),
	}
}

// GetVersion returns the application version
func (a *App) GetVersion() string {
	return a.version
}

// startup is called when the app starts
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// ListServices returns all LaunchAgent services
func (a *App) ListServices() ([]launchctl.Service, error) {
	return a.manager.List()
}

// GetService returns a single service by name
func (a *App) GetService(name string) (*launchctl.Service, error) {
	return a.manager.Get(name)
}

// StartService starts a service
func (a *App) StartService(name string) error {
	return a.manager.Start(name)
}

// StopService stops a service
func (a *App) StopService(name string) error {
	return a.manager.Stop(name)
}

// RestartService restarts a service
func (a *App) RestartService(name string) error {
	return a.manager.Restart(name)
}

// GetPlist returns the raw plist content
func (a *App) GetPlist(name string) (string, error) {
	return a.manager.GetPlist(name)
}

// GetLogs returns log content
func (a *App) GetLogs(name string, logType string) (string, error) {
	return a.manager.GetLogs(name, logType)
}

// CreateService creates a new service
func (a *App) CreateService(config launchctl.ServiceConfig) error {
	return a.manager.Create(&config)
}

// UpdateService updates an existing service (auto-backup before update)
func (a *App) UpdateService(name string, config launchctl.ServiceConfig) error {
	// Auto-backup before updating
	svc, err := a.manager.Get(name)
	if err == nil && svc.Path != "" {
		// Ignore backup errors - update should proceed regardless
		_, _ = a.backup.Create(name, svc.Path)
	}
	return a.manager.Update(name, &config)
}

// DeleteService deletes a service (auto-backup before delete)
func (a *App) DeleteService(name string) error {
	// Auto-backup before deleting
	svc, err := a.manager.Get(name)
	if err == nil {
		// Ignore backup errors - delete should proceed regardless
		_, _ = a.backup.Create(name, svc.Path)
	}
	return a.manager.Delete(name)
}

// ListAllBackups returns all backups for all services
func (a *App) ListAllBackups() ([]backup.Backup, error) {
	return a.backup.ListAll()
}

// ListBackups returns all backups for a service
func (a *App) ListBackups(serviceName string) ([]backup.Backup, error) {
	return a.backup.List(serviceName)
}

// GetBackupContent returns the content of a backup
func (a *App) GetBackupContent(serviceName, backupID string) (string, error) {
	return a.backup.GetContent(serviceName, backupID)
}

// RestoreBackup restores a backup to the service's plist path
func (a *App) RestoreBackup(serviceName, backupID string) error {
	// Try to get existing service path first
	svc, err := a.manager.Get(serviceName)
	if err == nil {
		return a.backup.Restore(serviceName, backupID, svc.Path)
	}

	// Service doesn't exist, use original path from backup metadata
	backups, err := a.backup.List(serviceName)
	if err != nil {
		return err
	}
	for _, b := range backups {
		if b.ID == backupID && b.OriginalPath != "" {
			return a.backup.Restore(serviceName, backupID, b.OriginalPath)
		}
	}

	return fmt.Errorf("cannot restore: service not found and no original path in backup")
}

// ListSystemServices returns all LaunchDaemon services from /Library
func (a *App) ListSystemServices() ([]launchctl.Service, error) {
	return a.systemManager.List()
}

// ListAppleSystemServices returns all LaunchDaemon services from /System/Library
func (a *App) ListAppleSystemServices() ([]launchctl.Service, error) {
	return a.appleSystemMgr.List()
}

// GetSystemService returns a system service by name and type
func (a *App) GetSystemService(name string, serviceType string) (*launchctl.Service, error) {
	switch serviceType {
	case "system":
		return a.systemManager.Get(name)
	case "apple-system":
		return a.appleSystemMgr.Get(name)
	default:
		return nil, fmt.Errorf("invalid service type: %s", serviceType)
	}
}

// GetSystemPlist returns raw plist content for system services
func (a *App) GetSystemPlist(name string, serviceType string) (string, error) {
	switch serviceType {
	case "system":
		return a.systemManager.GetPlist(name)
	case "apple-system":
		return a.appleSystemMgr.GetPlist(name)
	default:
		return "", fmt.Errorf("invalid service type: %s", serviceType)
	}
}

// GetSystemLogs returns log content for system services
func (a *App) GetSystemLogs(name string, serviceType string, logType string) (string, error) {
	switch serviceType {
	case "system":
		return a.systemManager.GetLogs(name, logType)
	case "apple-system":
		return a.appleSystemMgr.GetLogs(name, logType)
	default:
		return "", fmt.Errorf("invalid service type: %s", serviceType)
	}
}

// CheckPermissions returns permission status for each service domain
func (a *App) CheckPermissions() map[string]bool {
	canReadDir := func(path string) bool {
		_, err := os.ReadDir(path)
		return err == nil
	}

	return map[string]bool{
		"user":         true,
		"system":       canReadDir("/Library/LaunchDaemons"),
		"apple-system": canReadDir("/System/Library/LaunchDaemons"),
	}
}
