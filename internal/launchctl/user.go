package launchctl

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"howett.net/plist"

	"launchpal/internal/plistutil"
)

func guiDomain() string {
	return fmt.Sprintf("gui/%d", os.Getuid())
}

// serviceTarget builds a launchctl service-target of the form
// `<domain>/<label>`. An empty label is rejected: `launchctl bootout gui/<uid>/`
// is interpreted as the whole gui/<uid> domain-target, which unloads every
// user LaunchAgent and collapses the desktop session (requiring re-login).
func serviceTarget(domain, label string) (string, error) {
	if label == "" {
		return "", fmt.Errorf("service label is empty; refusing to build launchctl target that would unload the whole %s domain", domain)
	}
	return domain + "/" + label, nil
}

// UserManager manages user LaunchAgents in ~/Library/LaunchAgents
type UserManager struct {
	launchAgentsPath string
}

// NewUserManager creates a new UserManager
func NewUserManager() *UserManager {
	home, err := os.UserHomeDir()
	if err != nil {
		home = os.Getenv("HOME")
	}
	return &UserManager{
		launchAgentsPath: filepath.Join(home, "Library", "LaunchAgents"),
	}
}

// getLaunchAgentsPath returns the path to the LaunchAgents directory
func (m *UserManager) getLaunchAgentsPath() string {
	return m.launchAgentsPath
}

// plistData represents the structure of a LaunchAgent plist file
type plistData struct {
	Label                 string            `plist:"Label"`
	Program               string            `plist:"Program"`
	ProgramArguments      []string          `plist:"ProgramArguments"`
	UserName              string            `plist:"UserName"`
	RunAtLoad             bool              `plist:"RunAtLoad"`
	KeepAlive             interface{}       `plist:"KeepAlive"`
	StartCalendarInterval interface{}       `plist:"StartCalendarInterval"`
	StartInterval         int               `plist:"StartInterval"`
	EnvironmentVariables  map[string]string `plist:"EnvironmentVariables"`
	StandardOutPath       string            `plist:"StandardOutPath"`
	StandardErrorPath     string            `plist:"StandardErrorPath"`
	WorkingDirectory      string            `plist:"WorkingDirectory"`
	WakeSystem            bool              `plist:"WakeSystem"`
}

// List returns all services in ~/Library/LaunchAgents
func (m *UserManager) List() ([]Service, error) {
	path := m.getLaunchAgentsPath()

	// Create directory if it doesn't exist
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return []Service{}, nil
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read LaunchAgents directory: %w", err)
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

// Get returns a single service by name
func (m *UserManager) Get(name string) (*Service, error) {
	return m.getWithStatus(name, nil)
}

// getWithStatus returns a service, using pre-fetched statusMap if provided
func (m *UserManager) getWithStatus(name string, statusMap map[string]serviceStatus) (*Service, error) {
	plistPath := filepath.Join(m.getLaunchAgentsPath(), name+".plist")

	data, err := os.ReadFile(plistPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read plist file: %w", err)
	}

	var pd plistData
	_, err = plist.Unmarshal(data, &pd)
	if err != nil {
		return nil, fmt.Errorf("failed to parse plist: %w", err)
	}

	service := &Service{
		Name:             name,
		Label:            pd.Label,
		Path:             plistPath,
		Program:          pd.Program,
		Arguments:        pd.ProgramArguments,
		RunAtLoad:        pd.RunAtLoad,
		Environment:      pd.EnvironmentVariables,
		StdoutPath:       pd.StandardOutPath,
		StderrPath:       pd.StandardErrorPath,
		WorkingDir:       pd.WorkingDirectory,
		Type:             ServiceTypeUser,
		ReadOnly:         false,
		PlistFormat:      plistutil.DetectFormat(data),
		StatusConfidence: ConfidenceVerified,
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

// parseSchedule converts plist calendar interval or start interval to ScheduleConfig
func parseSchedule(calendarInterval interface{}, startInterval int) *ScheduleConfig {
	if calendarInterval == nil && startInterval == 0 {
		return nil
	}

	if startInterval > 0 {
		return &ScheduleConfig{Interval: &startInterval}
	}

	extractInt := func(v interface{}) *int {
		if v == nil {
			return nil
		}
		switch n := v.(type) {
		case int:
			return &n
		case int64:
			i := int(n)
			return &i
		case uint64:
			i := int(n)
			return &i
		case float64:
			i := int(n)
			return &i
		}
		return nil
	}

	dictToEntry := func(d map[string]interface{}) CalendarEntry {
		return CalendarEntry{
			Minute:  extractInt(d["Minute"]),
			Hour:    extractInt(d["Hour"]),
			Day:     extractInt(d["Day"]),
			Weekday: extractInt(d["Weekday"]),
			Month:   extractInt(d["Month"]),
		}
	}

	switch v := calendarInterval.(type) {
	case map[string]interface{}:
		return &ScheduleConfig{
			Schedules: []CalendarEntry{dictToEntry(v)},
		}
	case []interface{}:
		entries := make([]CalendarEntry, 0, len(v))
		for _, item := range v {
			if d, ok := item.(map[string]interface{}); ok {
				entries = append(entries, dictToEntry(d))
			}
		}
		if len(entries) > 0 {
			return &ScheduleConfig{Schedules: entries}
		}
	}
	return nil
}

// serviceStatus holds pre-fetched status info from batch launchctl list
type serviceStatus struct {
	status string
	pid    int
}

// getBatchServiceStatus runs `launchctl list` once and returns a map of label -> status/pid
func getBatchServiceStatus() map[string]serviceStatus {
	cmd := exec.Command("launchctl", "list")
	output, err := cmd.Output()
	if err != nil {
		return nil
	}

	result := make(map[string]serviceStatus)
	lines := strings.Split(string(output), "\n")
	for _, line := range lines[1:] { // skip header
		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}
		label := fields[2]
		pid, _ := strconv.Atoi(fields[0])
		if pid > 0 {
			result[label] = serviceStatus{status: StatusRunning, pid: pid}
		} else {
			result[label] = serviceStatus{status: StatusLoaded, pid: 0}
		}
	}
	return result
}

// getServiceStatus checks if a service is running via launchctl list
func getServiceStatus(label string) (string, int) {
	if label == "" {
		return StatusUnknown, 0
	}

	cmd := exec.Command("launchctl", "list", label)
	output, err := cmd.Output()
	if err != nil {
		// Service is not loaded
		return StatusStopped, 0
	}

	// Parse output to find PID and Program
	// Format: "PID" = <number> or similar
	lines := strings.Split(string(output), "\n")
	var program string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "\"PID\"") || strings.HasPrefix(line, "PID") {
			// Try to extract the number
			parts := strings.Split(line, "=")
			if len(parts) >= 2 {
				pidStr := strings.TrimSpace(parts[1])
				pidStr = strings.Trim(pidStr, ";\"")
				if pid, err := strconv.Atoi(pidStr); err == nil && pid > 0 {
					return StatusRunning, pid
				}
			}
		}
		// Extract Program path for fallback pgrep check
		if strings.Contains(line, "\"Program\"") {
			parts := strings.Split(line, "=")
			if len(parts) >= 2 {
				program = strings.TrimSpace(parts[1])
				program = strings.Trim(program, ";\"")
			}
		}
	}

	// Fallback: use pgrep to check if process is running.
	// Skip common shells as they would match unrelated processes;
	// commonShells lives in status_detect.go and is shared across managers.
	if program != "" && !commonShells[program] {
		pgrepCmd := exec.Command("pgrep", "-f", program)
		if pgrepOutput, err := pgrepCmd.Output(); err == nil {
			pidStr := strings.TrimSpace(strings.Split(string(pgrepOutput), "\n")[0])
			if pid, err := strconv.Atoi(pidStr); err == nil && pid > 0 {
				return StatusRunning, pid
			}
		}
	}

	// Service is loaded but not running
	return StatusLoaded, 0
}

// Start loads and starts a service
func (m *UserManager) Start(name string) error {
	plistPath := filepath.Join(m.getLaunchAgentsPath(), name+".plist")

	if _, err := os.Stat(plistPath); os.IsNotExist(err) {
		return fmt.Errorf("service %s not found", name)
	}

	cmd := exec.Command("launchctl", "bootstrap", guiDomain(), plistPath)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to start service: %s", string(output))
	}

	return nil
}

// Stop stops and unloads a service
func (m *UserManager) Stop(name string) error {
	service, err := m.Get(name)
	if err != nil {
		return err
	}

	target, err := serviceTarget(guiDomain(), service.Label)
	if err != nil {
		// A plist without a Label cannot be loaded by launchd, so there is
		// nothing to stop. (See serviceTarget for why we never issue the
		// malformed launchctl call.)
		return nil
	}

	cmd := exec.Command("launchctl", "bootout", target)
	_, _ = cmd.CombinedOutput() // Ignore error, service may not be loaded

	return nil
}

// Restart stops and starts a service
func (m *UserManager) Restart(name string) error {
	// Try to stop first, ignore errors if not running
	_ = m.Stop(name)

	return m.Start(name)
}

// Create creates a new service with the given config
func (m *UserManager) Create(config *ServiceConfig) error {
	if config.Label == "" {
		return fmt.Errorf("service label is required")
	}

	plistPath := filepath.Join(m.getLaunchAgentsPath(), config.Label+".plist")

	// Check if service already exists
	if _, err := os.Stat(plistPath); err == nil {
		return fmt.Errorf("service %s already exists", config.Label)
	}

	// Ensure LaunchAgents directory exists
	if err := os.MkdirAll(m.getLaunchAgentsPath(), 0755); err != nil {
		return fmt.Errorf("failed to create LaunchAgents directory: %w", err)
	}

	// Ensure log directories exist
	if config.StdoutPath != "" {
		logDir := filepath.Dir(expandTilde(config.StdoutPath))
		if err := os.MkdirAll(logDir, 0755); err != nil {
			return fmt.Errorf("failed to create log directory: %w", err)
		}
	}
	if config.StderrPath != "" {
		logDir := filepath.Dir(expandTilde(config.StderrPath))
		if err := os.MkdirAll(logDir, 0755); err != nil {
			return fmt.Errorf("failed to create log directory: %w", err)
		}
	}

	if err := validateSchedule(config.Schedule); err != nil {
		return err
	}

	return m.writePlist(plistPath, config)
}

// Update updates an existing service
func (m *UserManager) Update(name string, config *ServiceConfig) error {
	plistPath := filepath.Join(m.getLaunchAgentsPath(), name+".plist")

	// Check if service exists
	if _, err := os.Stat(plistPath); os.IsNotExist(err) {
		return fmt.Errorf("service %s not found", name)
	}

	// Stop the service first
	_ = m.Stop(name)

	if err := validateSchedule(config.Schedule); err != nil {
		return err
	}

	// Write updated plist
	if err := m.writePlist(plistPath, config); err != nil {
		return err
	}

	return nil
}

// validateSchedule checks that a ScheduleConfig has valid values
func validateSchedule(s *ScheduleConfig) error {
	if s == nil {
		return nil
	}
	if s.Interval != nil && *s.Interval < 10 {
		return fmt.Errorf("StartInterval must be at least 10 seconds, got %d", *s.Interval)
	}
	for i, e := range s.Schedules {
		if err := validateCalendarEntry(i, e); err != nil {
			return err
		}
	}
	return nil
}

// validateCalendarEntry checks that a CalendarEntry has valid field ranges
func validateCalendarEntry(index int, e CalendarEntry) error {
	checks := []struct {
		field string
		val   *int
		min   int
		max   int
	}{
		{"Minute", e.Minute, 0, 59},
		{"Hour", e.Hour, 0, 23},
		{"Day", e.Day, 1, 31},
		{"Weekday", e.Weekday, 0, 6},
		{"Month", e.Month, 1, 12},
	}
	for _, c := range checks {
		if c.val != nil && (*c.val < c.min || *c.val > c.max) {
			return fmt.Errorf("schedule entry %d: %s must be %d-%d, got %d", index, c.field, c.min, c.max, *c.val)
		}
	}
	return nil
}

// Delete removes a service
func (m *UserManager) Delete(name string) error {
	plistPath := filepath.Join(m.getLaunchAgentsPath(), name+".plist")

	// Check if service exists
	if _, err := os.Stat(plistPath); os.IsNotExist(err) {
		return fmt.Errorf("service %s not found", name)
	}

	// Stop the service first
	_ = m.Stop(name)

	// Remove the plist file
	if err := os.Remove(plistPath); err != nil {
		return fmt.Errorf("failed to delete service: %w", err)
	}

	return nil
}

// GetPlist returns the raw plist content
func (m *UserManager) GetPlist(name string) (string, error) {
	plistPath := filepath.Join(m.getLaunchAgentsPath(), name+".plist")

	data, err := os.ReadFile(plistPath)
	if err != nil {
		return "", fmt.Errorf("failed to read plist file: %w", err)
	}

	return string(data), nil
}

// GetPlistContent returns the current plist content normalized to XML. Binary
// plists are converted via plutil. If the plist file does not exist an empty
// Content is returned with no error so callers (e.g. the diff preview) can
// represent the service as having no current version.
func (m *UserManager) GetPlistContent(name string) (*plistutil.Content, error) {
	plistPath := filepath.Join(m.getLaunchAgentsPath(), name+".plist")

	content, err := plistutil.NormalizeFromPath(plistPath)
	if err != nil {
		if os.IsNotExist(err) {
			return &plistutil.Content{}, nil
		}
		return nil, fmt.Errorf("failed to read plist file: %w", err)
	}
	return content, nil
}

// GetLogs returns stdout or stderr log content
func (m *UserManager) GetLogs(name string, logType string) (string, error) {
	service, err := m.Get(name)
	if err != nil {
		return "", err
	}
	logPath, err := resolveLogPath(service, name, logType)
	if err != nil {
		return "", err
	}
	return readLogTail(expandTilde(logPath))
}

// ClearLogs truncates the configured stdout or stderr log file for the
// service to 0 bytes. The file's inode, owner, and mode are preserved; the
// file is not deleted. Errno mapping comes from the OpenFile call itself
// rather than a separate Lstat — the kernel's answer is the only one
// without a TOCTOU window.
func (m *UserManager) ClearLogs(name string, logType string) error {
	logPath, err := validateClearLogsArgs(m.Get, name, logType, true)
	if err != nil {
		return err
	}
	return truncateLogFile(logPath)
}

// GetLogClearStatus returns the resolved log path, its existence, and
// whether the current process can write it without following symlinks.
// Errors are only returned for missing services; other states are encoded
// in the returned LogClearStatus.
func (m *UserManager) GetLogClearStatus(name string, logType string) (LogClearStatus, error) {
	service, err := m.Get(name)
	if err != nil {
		return LogClearStatus{}, err
	}
	logPath := selectLogPath(service, logType)
	if logPath == "" {
		return LogClearStatus{}, nil
	}
	return logClearStatusFor(expandTilde(logPath)), nil
}

// expandTilde expands ~ to the user's home directory
func expandTilde(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			home = os.Getenv("HOME")
		}
		return filepath.Join(home, path[2:])
	}
	return path
}

// writePlist writes a ServiceConfig to a plist file.
func (m *UserManager) writePlist(path string, config *ServiceConfig) error {
	pd := BuildPlistDict(config, true)

	var buf bytes.Buffer
	encoder := plist.NewEncoder(&buf)
	encoder.Indent("\t")
	if err := encoder.Encode(pd); err != nil {
		return fmt.Errorf("failed to encode plist: %w", err)
	}

	if err := os.WriteFile(path, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write plist file: %w", err)
	}

	return nil
}

// Kickstart immediately runs a user service using launchctl kickstart -k.
// If the service is not loaded, it bootstraps it first.
func (m *UserManager) Kickstart(name string) error {
	plistPath := filepath.Join(m.getLaunchAgentsPath(), name+".plist")

	if _, err := os.Stat(plistPath); os.IsNotExist(err) {
		return fmt.Errorf("service %s not found", name)
	}

	// Read plist to get the label
	data, err := os.ReadFile(plistPath)
	if err != nil {
		return fmt.Errorf("failed to read plist file: %w", err)
	}

	var pd plistData
	if _, err := plist.Unmarshal(data, &pd); err != nil {
		return fmt.Errorf("failed to parse plist: %w", err)
	}

	domain := guiDomain()
	target, err := serviceTarget(domain, pd.Label)
	if err != nil {
		return fmt.Errorf("cannot kickstart %s: %w", name, err)
	}

	// Check if the service is loaded by querying launchctl list
	checkCmd := exec.Command("launchctl", "list", pd.Label)
	if err := checkCmd.Run(); err != nil {
		// Service is not loaded — bootstrap it first
		bootstrapCmd := exec.Command("launchctl", "bootstrap", domain, plistPath)
		if output, err := bootstrapCmd.CombinedOutput(); err != nil {
			return fmt.Errorf("failed to bootstrap service: %s", string(output))
		}
	}

	// Kickstart with -k to terminate existing process and restart
	kickCmd := exec.Command("launchctl", "kickstart", "-k", target)
	if output, err := kickCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to kickstart service: %s", string(output))
	}

	return nil
}

// Ensure UserManager implements Manager interface
var _ Manager = (*UserManager)(nil)
