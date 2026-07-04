package launchctl

import "fmt"

// AppleSystemManager manages Apple system LaunchDaemons in /System/Library/LaunchDaemons (read-only)
type AppleSystemManager struct {
	readOnlyManager
}

// NewAppleSystemManager creates a new AppleSystemManager
func NewAppleSystemManager() *AppleSystemManager {
	return &AppleSystemManager{
		readOnlyManager: readOnlyManager{
			basePath:    "/System/Library/LaunchDaemons",
			serviceType: "apple-system",
		},
	}
}

func (m *AppleSystemManager) List() ([]Service, error)             { return m.list() }
func (m *AppleSystemManager) Get(name string) (*Service, error)    { return m.get(name) }
func (m *AppleSystemManager) GetPlist(name string) (string, error) { return m.getPlist(name) }
func (m *AppleSystemManager) GetLogs(name string, logType string) (LogsResult, error) {
	return m.getLogs(name, logType)
}

// GetLogClearStatus returns the resolved log path / existence / writability
// for an apple-system service. The Clear Logs control is hidden in this
// domain, but the status query stays available so the UI can render the
// same surface regardless of service type.
func (m *AppleSystemManager) GetLogClearStatus(name string, logType string) (LogClearStatus, error) {
	return m.getLogClearStatus(name, logType)
}

// ClearLogs is rejected for apple-system services. The Wails layer should
// never route here — but the manager refuses regardless because SIP would
// block the truncate even if Admin Mode were on.
func (m *AppleSystemManager) ClearLogs(name string, logType string) error {
	return fmt.Errorf("apple-system services are read-only: cannot clear logs (%w)", ErrReadOnlyManager)
}

// Start returns ErrReadOnlyManager as Apple system services are read-only.
func (m *AppleSystemManager) Start(name string) error { return ErrReadOnlyManager }

// Stop returns ErrReadOnlyManager as Apple system services are read-only.
func (m *AppleSystemManager) Stop(name string) error { return ErrReadOnlyManager }

// Restart returns ErrReadOnlyManager as Apple system services are read-only.
func (m *AppleSystemManager) Restart(name string) error { return ErrReadOnlyManager }

// Create returns ErrReadOnlyManager as Apple system services are read-only.
func (m *AppleSystemManager) Create(config *ServiceConfig) error { return ErrReadOnlyManager }

// Update returns ErrReadOnlyManager as Apple system services are read-only.
func (m *AppleSystemManager) Update(name string, config *ServiceConfig) error {
	return ErrReadOnlyManager
}
func (m *AppleSystemManager) Delete(name string) error { return ErrReadOnlyManager }

// Ensure AppleSystemManager implements Manager interface
var _ Manager = (*AppleSystemManager)(nil)
