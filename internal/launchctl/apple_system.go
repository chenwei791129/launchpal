package launchctl

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"howett.net/plist"
)

// AppleSystemManager manages Apple system LaunchDaemons in /System/Library/LaunchDaemons (read-only)
type AppleSystemManager struct {
	launchDaemonsPath string
}

// NewAppleSystemManager creates a new AppleSystemManager
func NewAppleSystemManager() *AppleSystemManager {
	return &AppleSystemManager{
		launchDaemonsPath: "/System/Library/LaunchDaemons",
	}
}

// List returns all services in /System/Library/LaunchDaemons
func (m *AppleSystemManager) List() ([]Service, error) {
	entries, err := os.ReadDir(m.launchDaemonsPath)
	if err != nil {
		if os.IsNotExist(err) || os.IsPermission(err) {
			return []Service{}, nil
		}
		return nil, fmt.Errorf("failed to read LaunchDaemons directory: %w", err)
	}

	var services []Service
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".plist") {
			continue
		}

		name := strings.TrimSuffix(entry.Name(), ".plist")
		service, err := m.Get(name)
		if err != nil {
			// Skip services we can't read
			continue
		}
		services = append(services, *service)
	}

	if services == nil {
		services = []Service{}
	}

	return services, nil
}

// Get returns a single service by name
func (m *AppleSystemManager) Get(name string) (*Service, error) {
	plistPath := filepath.Join(m.launchDaemonsPath, name+".plist")

	data, err := os.ReadFile(plistPath)
	if err != nil {
		if os.IsPermission(err) {
			return nil, fmt.Errorf("permission denied reading %s", name)
		}
		return nil, fmt.Errorf("failed to read plist file: %w", err)
	}

	var pd plistData
	_, err = plist.Unmarshal(data, &pd)
	if err != nil {
		return nil, fmt.Errorf("failed to parse plist: %w", err)
	}

	service := &Service{
		Name:        name,
		Label:       pd.Label,
		Path:        plistPath,
		Program:     pd.Program,
		Arguments:   pd.ProgramArguments,
		RunAtLoad:   pd.RunAtLoad,
		Environment: pd.EnvironmentVariables,
		StdoutPath:  pd.StandardOutPath,
		StderrPath:  pd.StandardErrorPath,
		WorkingDir:  pd.WorkingDirectory,
		Type:        "apple-system",
		ReadOnly:    true,
		PlistFormat: detectPlistFormat(data),
	}

	// Handle KeepAlive
	switch v := pd.KeepAlive.(type) {
	case bool:
		service.KeepAlive = v
	case map[string]interface{}:
		service.KeepAlive = true
	}

	// Handle Schedule
	um := &UserManager{}
	service.Schedule = um.parseSchedule(pd.StartCalendarInterval, pd.StartInterval)

	// Get service status
	status, pid := (&UserManager{}).getServiceStatus(pd.Label)
	service.Status = status
	service.PID = pid

	return service, nil
}

// GetPlist returns the raw plist content
func (m *AppleSystemManager) GetPlist(name string) (string, error) {
	plistPath := filepath.Join(m.launchDaemonsPath, name+".plist")

	// Use plutil to convert binary plist to XML format
	cmd := exec.Command("plutil", "-convert", "xml1", "-o", "-", plistPath)
	output, err := cmd.Output()
	if err != nil {
		if os.IsPermission(err) || strings.Contains(err.Error(), "permission denied") {
			return "", fmt.Errorf("permission denied reading %s", name)
		}
		return "", fmt.Errorf("failed to read plist file: %w", err)
	}

	return string(output), nil
}

// GetLogs returns stdout or stderr log content
func (m *AppleSystemManager) GetLogs(name string, logType string) (string, error) {
	service, err := m.Get(name)
	if err != nil {
		return "", err
	}

	var logPath string
	switch logType {
	case "stdout":
		logPath = service.StdoutPath
	case "stderr":
		logPath = service.StderrPath
	default:
		return "", fmt.Errorf("invalid log type: %s (use 'stdout' or 'stderr')", logType)
	}

	if logPath == "" {
		return "", fmt.Errorf("no %s log path configured for service %s", logType, name)
	}

	data, err := os.ReadFile(logPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("log file not found: %s", logPath)
		}
		if os.IsPermission(err) {
			return "", fmt.Errorf("permission denied reading log file: %s", logPath)
		}
		return "", fmt.Errorf("failed to read log file: %w", err)
	}

	return string(data), nil
}

// Write operations return ErrReadOnlyManager

func (m *AppleSystemManager) Start(name string) error {
	return ErrReadOnlyManager
}

func (m *AppleSystemManager) Stop(name string) error {
	return ErrReadOnlyManager
}

func (m *AppleSystemManager) Restart(name string) error {
	return ErrReadOnlyManager
}

func (m *AppleSystemManager) Create(config *ServiceConfig) error {
	return ErrReadOnlyManager
}

func (m *AppleSystemManager) Update(name string, config *ServiceConfig) error {
	return ErrReadOnlyManager
}

func (m *AppleSystemManager) Delete(name string) error {
	return ErrReadOnlyManager
}

// Ensure AppleSystemManager implements Manager interface
var _ Manager = (*AppleSystemManager)(nil)
