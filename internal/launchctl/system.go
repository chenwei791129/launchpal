package launchctl

// SystemManager manages system LaunchDaemons in /Library/LaunchDaemons (read-only)
type SystemManager struct {
	readOnlyManager
}

// NewSystemManager creates a new SystemManager
func NewSystemManager() *SystemManager {
	return &SystemManager{
		readOnlyManager: readOnlyManager{
			basePath:    "/Library/LaunchDaemons",
			serviceType: "system",
		},
	}
}

func (m *SystemManager) List() ([]Service, error)                  { return m.list() }
func (m *SystemManager) Get(name string) (*Service, error)         { return m.get(name) }
func (m *SystemManager) GetPlist(name string) (string, error)      { return m.getPlist(name) }
func (m *SystemManager) GetLogs(name string, logType string) (string, error) {
	return m.getLogs(name, logType)
}

// Write operations return ErrReadOnlyManager
func (m *SystemManager) Start(name string) error                      { return ErrReadOnlyManager }
func (m *SystemManager) Stop(name string) error                       { return ErrReadOnlyManager }
func (m *SystemManager) Restart(name string) error                    { return ErrReadOnlyManager }
func (m *SystemManager) Create(config *ServiceConfig) error           { return ErrReadOnlyManager }
func (m *SystemManager) Update(name string, config *ServiceConfig) error { return ErrReadOnlyManager }
func (m *SystemManager) Delete(name string) error                     { return ErrReadOnlyManager }

// Ensure SystemManager implements Manager interface
var _ Manager = (*SystemManager)(nil)
