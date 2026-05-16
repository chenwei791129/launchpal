package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"launchpal/internal/backup"
	"launchpal/internal/launchctl"
	"launchpal/internal/plistutil"
	"launchpal/internal/settings"
)

// App struct
type App struct {
	ctx            context.Context
	version        string
	manager        *launchctl.UserManager
	systemManager  *launchctl.SystemManager
	appleSystemMgr *launchctl.AppleSystemManager
	backup         *backup.BackupManager
	admin          *adminModeManager
}

// NewApp creates a new App application struct with default version
func NewApp() *App {
	return NewAppWithVersion("dev")
}

// NewAppWithVersion creates a new App with the specified version
func NewAppWithVersion(version string) *App {
	sysMgr := launchctl.NewSystemManager()
	return &App{
		version:        version,
		manager:        launchctl.NewUserManager(),
		systemManager:  sysMgr,
		appleSystemMgr: launchctl.NewAppleSystemManager(),
		backup:         backup.NewBackupManager(),
		admin:          newAdminModeManager(sysMgr),
	}
}

// GetVersion returns the application version
func (a *App) GetVersion() string {
	return a.version
}

// startup is called when the app starts
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	// Wire the Wails event emitter so state transitions reach the frontend
	// without requiring a poll.
	a.admin.event = &wailsEventEmitter{ctx: ctx}
}

// EnableAdminMode launches the privileged helper (prompts for user
// authorization via osascript) and installs it as the write backend for
// SystemManager. Returns nil when Admin Mode is already Enabled or when the
// user cancelled the authorization prompt; returns an error only when the
// helper could not be started.
func (a *App) EnableAdminMode() error {
	return a.admin.Enable(a.ctx)
}

// DisableAdminMode gracefully shuts the helper down and reverts SystemManager
// to read-only. Safe to call when Admin Mode is already Disabled.
func (a *App) DisableAdminMode() error {
	return a.admin.Disable(a.ctx)
}

// GetAdminModeStatus returns the current Admin Mode state and any recent
// error code. The frontend polls this (and listens to the
// "admin_mode:state" event) to drive the UI.
func (a *App) GetAdminModeStatus() AdminModeStatus {
	return a.admin.status()
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

// GetBackupContent returns the normalized content of a backup plist (binary
// plists are auto-converted to XML).
func (a *App) GetBackupContent(serviceName, backupID string) (*plistutil.Content, error) {
	return a.backup.GetContent(serviceName, backupID)
}

// GetCurrentPlist returns the current plist content of a user service,
// normalized to XML. Used by the backup diff preview. Returns an empty Content
// (not an error) when the service is absent or its plist file is missing so
// the diff view can render the full backup as additions.
func (a *App) GetCurrentPlist(name string) (*plistutil.Content, error) {
	return a.manager.GetPlistContent(name)
}

// GetCurrentSystemPlist is the system-domain counterpart of GetCurrentPlist:
// it reads /Library/LaunchDaemons/<name>.plist for diffing against a backup.
// Same empty-on-missing convention as GetCurrentPlist.
func (a *App) GetCurrentSystemPlist(name string) (*plistutil.Content, error) {
	return a.systemManager.GetPlistContent(name)
}

// RestoreBackup restores a backup to the service's plist path. Dispatches on
// the backup's target location: user-domain backups (under ~/Library/LaunchAgents
// or any path writable by the GUI user) use a direct file copy; system-domain
// backups (paths under /Library/LaunchDaemons) are routed through the
// privileged helper, which fails with ErrReadOnlyManager if Admin Mode is off.
// The target path comes from the service's current location when available and
// falls back to the OriginalPath stored in the backup metadata.
func (a *App) RestoreBackup(serviceName, backupID string) error {
	targetPath := ""
	if svc, err := a.manager.Get(serviceName); err == nil && svc.Path != "" {
		targetPath = svc.Path
	}
	if targetPath == "" {
		b, err := a.backup.Get(serviceName, backupID)
		if err != nil {
			return err
		}
		targetPath = b.OriginalPath
	}
	if targetPath == "" {
		return fmt.Errorf("cannot restore: service not found and no original path in backup")
	}

	// Paths under /Library/LaunchDaemons are root-owned; a direct copyFile
	// would fail with EACCES. Route through the helper instead.
	if isSystemDaemonPath(targetPath) {
		content, err := a.backup.GetContent(serviceName, backupID)
		if err != nil {
			return err
		}
		return a.systemManager.RestorePlist(targetPath, []byte(content.Data))
	}
	return a.backup.Restore(serviceName, backupID, targetPath)
}

// isSystemDaemonPath reports whether path lives under /Library/LaunchDaemons.
// The comparison uses a clean absolute-path + trailing-separator check so
// look-alike directories (e.g. /Library/LaunchDaemonsX) don't match.
func isSystemDaemonPath(path string) bool {
	clean := filepath.Clean(path)
	prefix := "/Library/LaunchDaemons/"
	return strings.HasPrefix(clean+"/", prefix)
}

// KickstartService immediately runs a user service via launchctl kickstart
func (a *App) KickstartService(name string) error {
	return a.manager.Kickstart(name)
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
	case launchctl.ServiceTypeSystem:
		return a.systemManager.Get(name)
	case launchctl.ServiceTypeAppleSystem:
		return a.appleSystemMgr.Get(name)
	default:
		return nil, fmt.Errorf("invalid service type: %s", serviceType)
	}
}

// GetSystemPlist returns raw plist content for system services
func (a *App) GetSystemPlist(name string, serviceType string) (string, error) {
	switch serviceType {
	case launchctl.ServiceTypeSystem:
		return a.systemManager.GetPlist(name)
	case launchctl.ServiceTypeAppleSystem:
		return a.appleSystemMgr.GetPlist(name)
	default:
		return "", fmt.Errorf("invalid service type: %s", serviceType)
	}
}

// GetSystemLogs returns log content for system services
func (a *App) GetSystemLogs(name string, serviceType string, logType string) (string, error) {
	switch serviceType {
	case launchctl.ServiceTypeSystem:
		return a.systemManager.GetLogs(name, logType)
	case launchctl.ServiceTypeAppleSystem:
		return a.appleSystemMgr.GetLogs(name, logType)
	default:
		return "", fmt.Errorf("invalid service type: %s", serviceType)
	}
}

// ClearLogs truncates the configured stdout or stderr log file for a user
// service to 0 bytes. The file's inode and mode are preserved.
func (a *App) ClearLogs(name string, logType string) error {
	return a.manager.ClearLogs(name, logType)
}

// ClearSystemLogs truncates a system daemon's configured log file via
// SystemManager. apple-system is rejected at the binding gate — SIP would
// block the truncate even with Admin Mode on, and we want the destructive
// surface to fail before reaching any manager.
func (a *App) ClearSystemLogs(name string, serviceType string, logType string) error {
	switch serviceType {
	case launchctl.ServiceTypeSystem:
		return a.systemManager.ClearLogs(name, logType)
	case launchctl.ServiceTypeAppleSystem:
		return fmt.Errorf("apple-system services are read-only: cannot clear logs (%w)", launchctl.ErrReadOnlyManager)
	default:
		return fmt.Errorf("invalid service type: %s", serviceType)
	}
}

// GetLogClearStatus returns the resolved log path, its existence, and
// whether the calling process can write the file. The frontend uses this to
// decide whether the Clear Logs button is enabled and which tooltip to show.
// apple-system services return the same status struct for consistency, but
// the UI hides the Clear Logs button regardless.
func (a *App) GetLogClearStatus(name string, serviceType string, logType string) (launchctl.LogClearStatus, error) {
	switch serviceType {
	case launchctl.ServiceTypeUser:
		return a.manager.GetLogClearStatus(name, logType)
	case launchctl.ServiceTypeSystem:
		return a.systemManager.GetLogClearStatus(name, logType)
	case launchctl.ServiceTypeAppleSystem:
		return a.appleSystemMgr.GetLogClearStatus(name, logType)
	default:
		return launchctl.LogClearStatus{}, fmt.Errorf("invalid service type: %s", serviceType)
	}
}

// StartSystemService starts a system daemon via the privileged helper.
// Requires Admin Mode enabled; returns launchctl.ErrReadOnlyManager otherwise.
func (a *App) StartSystemService(name string) error {
	return a.systemManager.Start(name)
}

// StopSystemService stops a system daemon via the privileged helper.
func (a *App) StopSystemService(name string) error {
	return a.systemManager.Stop(name)
}

// RestartSystemService kickstarts (-k) a system daemon via the helper.
func (a *App) RestartSystemService(name string) error {
	return a.systemManager.Restart(name)
}

// UpdateSystemService writes a new plist for a system daemon via the helper
// (the helper takes a backup before overwriting).
func (a *App) UpdateSystemService(name string, config launchctl.ServiceConfig) error {
	return a.systemManager.Update(name, &config)
}

// CreateSystemService creates a new system daemon plist and bootstraps it.
func (a *App) CreateSystemService(config launchctl.ServiceConfig) error {
	return a.systemManager.Create(&config)
}

// DeleteSystemService boots out and removes a system daemon plist via the
// helper. Returns a non-empty warning string when the plist was deleted
// successfully but log cleanup partially failed; the daemon delete still
// counts as success in that case, so the Wails Promise must NOT reject.
// Hard failures (helper unreachable, Admin Mode off, etc.) come back as
// the error return.
func (a *App) DeleteSystemService(name string, options launchctl.DeleteServiceOptions) (string, error) {
	err := a.systemManager.DeleteWithOptions(name, options)
	if err == nil {
		return "", nil
	}
	var warn *launchctl.LogDeletionWarning
	if errors.As(err, &warn) {
		return warn.Error(), nil
	}
	return "", err
}

// RevealInFinder opens Finder and highlights the file at the given path.
func (a *App) RevealInFinder(path string) error {
	if path == "" {
		return fmt.Errorf("path must not be empty")
	}
	return exec.Command("open", "-R", path).Run()
}

// GetSettings reads ~/.launchpal/settings.json (or returns Default() if the
// file is missing or corrupt) and surfaces it to the frontend.
func (a *App) GetSettings() (settings.Settings, error) {
	return settings.Load()
}

// UpdateSettings validates and atomically persists the supplied Settings.
// Validation errors are returned verbatim; the on-disk settings file is
// left untouched if validation fails.
func (a *App) UpdateSettings(s settings.Settings) error {
	return settings.Save(s)
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
