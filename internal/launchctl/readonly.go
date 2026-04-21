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

	statusMap := getBatchServiceStatus()
	// One `ps` invocation shared across every heuristic detection call in this
	// List, instead of O(candidates * services) forks.
	ppidTable, _ := readAllPPIDsFn()

	services := []Service{}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".plist") {
			continue
		}

		name := strings.TrimSuffix(entry.Name(), ".plist")
		service, err := m.getWithStatus(name, statusMap, ppidTable)
		if err != nil {
			continue
		}
		services = append(services, *service)
	}

	return services, nil
}

// get returns a single service by name (queries status individually)
func (m *readOnlyManager) get(name string) (*Service, error) {
	return m.getWithStatus(name, nil, nil)
}

// getWithStatus returns a service, using pre-fetched statusMap and ppidTable
// if provided (both may be nil for single-service queries).
func (m *readOnlyManager) getWithStatus(name string, statusMap map[string]serviceStatus, ppidTable map[int]int) (*Service, error) {
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

	// Resolve status/PID/confidence. For system-domain managers, `launchctl
	// list` (which powers statusMap) only sees gui/<uid> services, so misses
	// are the common case — fall through to heuristic detection instead of
	// defaulting to Stopped.
	if statusMap != nil {
		if s, ok := statusMap[pd.Label]; ok {
			service.Status = s.status
			service.PID = s.pid
			service.StatusConfidence = ConfidenceVerified
			return service, nil
		}
	}
	service.Status, service.PID, service.StatusConfidence = DetectSystemServiceStatus(pd, ppidTable)

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
