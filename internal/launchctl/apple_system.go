package launchctl

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

func (m *AppleSystemManager) List() ([]Service, error)                  { return m.list() }
func (m *AppleSystemManager) Get(name string) (*Service, error)         { return m.get(name) }
func (m *AppleSystemManager) GetPlist(name string) (string, error)      { return m.getPlist(name) }
func (m *AppleSystemManager) GetLogs(name string, logType string) (string, error) {
	return m.getLogs(name, logType)
}

// Write operations return ErrReadOnlyManager
func (m *AppleSystemManager) Start(name string) error                      { return ErrReadOnlyManager }
func (m *AppleSystemManager) Stop(name string) error                       { return ErrReadOnlyManager }
func (m *AppleSystemManager) Restart(name string) error                    { return ErrReadOnlyManager }
func (m *AppleSystemManager) Create(config *ServiceConfig) error           { return ErrReadOnlyManager }
func (m *AppleSystemManager) Update(name string, config *ServiceConfig) error { return ErrReadOnlyManager }
func (m *AppleSystemManager) Delete(name string) error                     { return ErrReadOnlyManager }

// Ensure AppleSystemManager implements Manager interface
var _ Manager = (*AppleSystemManager)(nil)
