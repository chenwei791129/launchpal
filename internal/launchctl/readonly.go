package launchctl

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"howett.net/plist"

	"launchpal/internal/plistutil"
)

// readOnlyManager provides shared read-only operations for system service managers
type readOnlyManager struct {
	basePath    string
	serviceType string
}

// list returns all services in the managed directory
func (m *readOnlyManager) list() ([]Service, error) {
	entries, err := os.ReadDir(m.basePath)
	if err != nil {
		if os.IsNotExist(err) || os.IsPermission(err) {
			return []Service{}, nil
		}
		return nil, fmt.Errorf("failed to read LaunchDaemons directory: %w", err)
	}

	// Batch-fetch all service statuses with a single launchctl list call
	statusMap := getBatchServiceStatus()

	services := []Service{}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".plist") {
			continue
		}

		name := strings.TrimSuffix(entry.Name(), ".plist")
		service, err := m.getWithStatus(name, statusMap)
		if err != nil {
			continue
		}
		services = append(services, *service)
	}

	return services, nil
}

// get returns a single service by name (queries status individually)
func (m *readOnlyManager) get(name string) (*Service, error) {
	return m.getWithStatus(name, nil)
}

// getWithStatus returns a service, using pre-fetched statusMap if provided
func (m *readOnlyManager) getWithStatus(name string, statusMap map[string]serviceStatus) (*Service, error) {
	plistPath := filepath.Join(m.basePath, name+".plist")

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
		Type:        m.serviceType,
		ReadOnly:    true,
		PlistFormat: plistutil.DetectFormat(data),
	}

	service.KeepAlive = parseKeepAlive(pd.KeepAlive)
	service.WakeSystem = pd.WakeSystem
	service.Schedule = parseSchedule(pd.StartCalendarInterval, pd.StartInterval)

	// Use pre-fetched status if available, otherwise query individually
	if statusMap != nil {
		if s, ok := statusMap[pd.Label]; ok {
			service.Status = s.status
			service.PID = s.pid
		} else {
			service.Status = StatusStopped
		}
	} else {
		service.Status, service.PID = getServiceStatus(pd.Label)
	}

	return service, nil
}

func (m *readOnlyManager) getPlist(name string) (string, error) {
	plistPath := filepath.Join(m.basePath, name+".plist")

	content, err := plistutil.NormalizeFromPath(plistPath)
	if err != nil {
		if os.IsPermission(err) || strings.Contains(err.Error(), "permission denied") {
			return "", fmt.Errorf("permission denied reading %s", name)
		}
		return "", fmt.Errorf("failed to read plist file: %w", err)
	}
	return content.Data, nil
}

// getLogs returns stdout or stderr log content
func (m *readOnlyManager) getLogs(name string, logType string) (string, error) {
	service, err := m.get(name)
	if err != nil {
		return "", err
	}

	var logPath string
	switch logType {
	case LogTypeStdout:
		logPath = service.StdoutPath
	case LogTypeStderr:
		logPath = service.StderrPath
	default:
		return "", fmt.Errorf("invalid log type: %s (use 'stdout' or 'stderr')", logType)
	}

	if logPath == "" {
		return "", fmt.Errorf("no %s log path configured for service %s", logType, name)
	}

	return readLogTail(logPath)
}
