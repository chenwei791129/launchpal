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
	// One `ps` snapshot and one UserName→uid cache shared across every
	// heuristic detection call in this List, collapsing O(services) subprocess
	// forks to exactly one. If the fetch fails, `table` is nil and each
	// DetectSystemServiceStatus call will lazily retry — the 411-daemon retry
	// cascade is inefficient in the rare ps-broken state but still produces
	// the correct Stopped/Unverified verdict for every service.
	table, _ := readProcessTableFn()
	uidCache := make(map[string]int)

	services := []Service{}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".plist") {
			continue
		}

		name := strings.TrimSuffix(entry.Name(), ".plist")
		service, err := m.getWithStatus(name, statusMap, table, uidCache)
		if err != nil {
			continue
		}
		services = append(services, *service)
	}

	return services, nil
}

// get returns a single service by name (queries status individually)
func (m *readOnlyManager) get(name string) (*Service, error) {
	return m.getWithStatus(name, nil, nil, nil)
}

// getWithStatus returns a service, using pre-fetched statusMap, process table,
// and uidCache when provided. All three may be nil for single-service queries,
// in which case DetectSystemServiceStatus fetches what it needs on demand.
func (m *readOnlyManager) getWithStatus(name string, statusMap map[string]serviceStatus, table ProcessTable, uidCache map[string]int) (*Service, error) {
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
	service.Status, service.PID, service.StatusConfidence = DetectSystemServiceStatus(pd, table, uidCache)

	return service, nil
}

// getPlistContent returns the plist normalized to XML (matching the user
// domain's UserManager.GetPlistContent shape). Used by the backup-diff
// preview so the left column — the "current on-disk plist" — renders the
// same way for both domains. A missing file returns an empty Content
// without error so the diff view can draw the backup as pure additions.
func (m *readOnlyManager) getPlistContent(name string) (*plistutil.Content, error) {
	plistPath := filepath.Join(m.basePath, name+".plist")
	content, err := plistutil.NormalizeFromPath(plistPath)
	if err != nil {
		if os.IsNotExist(err) {
			return &plistutil.Content{}, nil
		}
		return nil, fmt.Errorf("failed to read plist file: %w", err)
	}
	return content, nil
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

// getLogClearStatus returns the resolved log path, its existence, and
// whether the current process can write it. Shared by SystemManager and
// AppleSystemManager — apple-system surfaces the same status struct so the
// frontend can display matching tooltips even though the Clear control is
// hidden in that domain.
func (m *readOnlyManager) getLogClearStatus(name string, logType string) (LogClearStatus, error) {
	service, err := m.get(name)
	if err != nil {
		return LogClearStatus{}, err
	}
	logPath := selectLogPath(service, logType)
	if logPath == "" {
		return LogClearStatus{}, nil
	}
	return logClearStatusFor(logPath), nil
}

// getLogs returns stdout or stderr log content
func (m *readOnlyManager) getLogs(name string, logType string) (string, error) {
	service, err := m.get(name)
	if err != nil {
		return "", err
	}
	logPath, err := resolveLogPath(service, name, logType)
	if err != nil {
		return "", err
	}
	return readLogTail(logPath)
}
