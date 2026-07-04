package launchctl

// Manager defines the interface for managing LaunchAgents
type Manager interface {
	// List returns all services in the managed directory
	List() ([]Service, error)

	// Get returns a single service by name
	Get(name string) (*Service, error)

	// Start loads and starts a service
	Start(name string) error

	// Stop stops and unloads a service
	Stop(name string) error

	// Restart stops and starts a service
	Restart(name string) error

	// Create creates a new service with the given config
	Create(config *ServiceConfig) error

	// Update updates an existing service
	Update(name string, config *ServiceConfig) error

	// Delete removes a service
	Delete(name string) error

	// GetPlist returns the raw plist content
	GetPlist(name string) (string, error)

	// GetLogs returns stdout or stderr log content classified as a
	// LogsResult. Structural states (no path configured, file not created
	// yet) travel in LogsResult.Status with a nil error; only real
	// failures (invalid log type, missing service, permission, I/O) use
	// the error channel.
	GetLogs(name string, logType string) (LogsResult, error)

	// ClearLogs truncates the configured stdout or stderr log file for the
	// given service to 0 bytes. The file's inode, owner, group, and mode
	// are preserved. Returns an error if logType is invalid, no log path
	// is configured, the file does not exist, or the truncate fails.
	ClearLogs(name string, logType string) error

	// GetLogClearStatus returns information for the UI to decide whether
	// the Clear Logs control should be enabled. The structure is returned
	// even for empty log paths or missing files; an error is only returned
	// when the service itself does not exist.
	GetLogClearStatus(name string, logType string) (LogClearStatus, error)
}
