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

	// GetLogs returns stdout or stderr log content
	GetLogs(name string, logType string) (string, error)
}
