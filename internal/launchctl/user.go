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
)

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
	if m.launchAgentsPath != "" {
		return m.launchAgentsPath
	}
	home, err := os.UserHomeDir()
	if err != nil {
		home = os.Getenv("HOME")
	}
	return filepath.Join(home, "Library", "LaunchAgents")
}

// plistData represents the structure of a LaunchAgent plist file
type plistData struct {
	Label                  string            `plist:"Label"`
	Program                string            `plist:"Program"`
	ProgramArguments       []string          `plist:"ProgramArguments"`
	RunAtLoad              bool              `plist:"RunAtLoad"`
	KeepAlive              interface{}       `plist:"KeepAlive"`
	StartCalendarInterval  interface{}       `plist:"StartCalendarInterval"`
	StartInterval          int              `plist:"StartInterval"`
	EnvironmentVariables   map[string]string `plist:"EnvironmentVariables"`
	StandardOutPath        string            `plist:"StandardOutPath"`
	StandardErrorPath      string            `plist:"StandardErrorPath"`
	WorkingDirectory       string            `plist:"WorkingDirectory"`
}

// calendarInterval represents StartCalendarInterval in a plist
type calendarInterval struct {
	Minute  int `plist:"Minute"`
	Hour    int `plist:"Hour"`
	Day     int `plist:"Day"`
	Weekday int `plist:"Weekday"`
	Month   int `plist:"Month"`
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

	var services []Service
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".plist") {
			continue
		}

		name := strings.TrimSuffix(entry.Name(), ".plist")
		service, err := m.Get(name)
		if err != nil {
			// Log but continue with other services
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
func (m *UserManager) Get(name string) (*Service, error) {
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
		Type:        "user",
		ReadOnly:    false,
		PlistFormat: detectPlistFormat(data),
	}

	// Handle KeepAlive which can be bool or dict
	switch v := pd.KeepAlive.(type) {
	case bool:
		service.KeepAlive = v
	case map[string]interface{}:
		// If it's a dict, consider it as "conditionally keepalive"
		service.KeepAlive = true
	}

	// Handle schedule (StartCalendarInterval or StartInterval)
	service.Schedule = m.parseSchedule(pd.StartCalendarInterval, pd.StartInterval)

	// Get service status
	status, pid := m.getServiceStatus(pd.Label)
	service.Status = status
	service.PID = pid

	return service, nil
}

// parseSchedule converts plist calendar interval or start interval to ScheduleConfig
func (m *UserManager) parseSchedule(calendarInterval interface{}, startInterval int) *ScheduleConfig {
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

	switch v := calendarInterval.(type) {
	case map[string]interface{}:
		return &ScheduleConfig{
			Minute:  extractInt(v["Minute"]),
			Hour:    extractInt(v["Hour"]),
			Day:     extractInt(v["Day"]),
			Weekday: extractInt(v["Weekday"]),
			Month:   extractInt(v["Month"]),
		}
	case []interface{}:
		// Multiple intervals - just use the first one for simplicity
		if len(v) > 0 {
			if first, ok := v[0].(map[string]interface{}); ok {
				return &ScheduleConfig{
					Minute:  extractInt(first["Minute"]),
					Hour:    extractInt(first["Hour"]),
					Day:     extractInt(first["Day"]),
					Weekday: extractInt(first["Weekday"]),
					Month:   extractInt(first["Month"]),
				}
			}
		}
	}
	return nil
}

// getServiceStatus checks if a service is running via launchctl list
func (m *UserManager) getServiceStatus(label string) (string, int) {
	if label == "" {
		return "unknown", 0
	}

	cmd := exec.Command("launchctl", "list", label)
	output, err := cmd.Output()
	if err != nil {
		// Service is not loaded
		return "stopped", 0
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
					return "running", pid
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

	// Fallback: use pgrep to check if process is running
	// Skip common shells as they would match unrelated processes
	commonShells := map[string]bool{
		"/bin/bash": true, "/bin/sh": true, "/bin/zsh": true,
		"/usr/bin/bash": true, "/usr/bin/sh": true, "/usr/bin/zsh": true,
	}
	if program != "" && !commonShells[program] {
		pgrepCmd := exec.Command("pgrep", "-f", program)
		if pgrepOutput, err := pgrepCmd.Output(); err == nil {
			pidStr := strings.TrimSpace(strings.Split(string(pgrepOutput), "\n")[0])
			if pid, err := strconv.Atoi(pidStr); err == nil && pid > 0 {
				return "running", pid
			}
		}
	}

	// Service is loaded but not running
	return "loaded", 0
}

// Start loads and starts a service
func (m *UserManager) Start(name string) error {
	plistPath := filepath.Join(m.getLaunchAgentsPath(), name+".plist")

	if _, err := os.Stat(plistPath); os.IsNotExist(err) {
		return fmt.Errorf("service %s not found", name)
	}

	cmd := exec.Command("launchctl", "load", plistPath)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to start service: %s", string(output))
	}

	return nil
}

// Stop stops and unloads a service
func (m *UserManager) Stop(name string) error {
	plistPath := filepath.Join(m.getLaunchAgentsPath(), name+".plist")

	if _, err := os.Stat(plistPath); os.IsNotExist(err) {
		return fmt.Errorf("service %s not found", name)
	}

	// Get service info to find label and program
	service, err := m.Get(name)
	if err != nil {
		return err
	}

	// Try launchctl unload first
	cmd := exec.Command("launchctl", "unload", plistPath)
	cmd.CombinedOutput() // Ignore error, may fail for root-owned processes

	// Check if process is still running and kill it
	if service.Program != "" {
		pgrepCmd := exec.Command("pgrep", "-f", service.Program)
		if output, err := pgrepCmd.Output(); err == nil {
			pidStr := strings.TrimSpace(strings.Split(string(output), "\n")[0])
			if pid, err := strconv.Atoi(pidStr); err == nil && pid > 0 {
				// Try to kill the process
				killCmd := exec.Command("kill", pidStr)
				killCmd.Run()
			}
		}
	}

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

// GetLogs returns stdout or stderr log content
func (m *UserManager) GetLogs(name string, logType string) (string, error) {
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

	// Expand ~ in path
	if strings.HasPrefix(logPath, "~") {
		home, _ := os.UserHomeDir()
		logPath = filepath.Join(home, logPath[1:])
	}

	data, err := os.ReadFile(logPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("log file not found: %s", logPath)
		}
		return "", fmt.Errorf("failed to read log file: %w", err)
	}

	return string(data), nil
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

// writePlist writes a ServiceConfig to a plist file
func (m *UserManager) writePlist(path string, config *ServiceConfig) error {
	pd := map[string]interface{}{
		"Label": config.Label,
	}

	// Set program or program arguments
	if config.Program != "" {
		pd["Program"] = config.Program
	}
	if len(config.Arguments) > 0 {
		pd["ProgramArguments"] = config.Arguments
	}

	// Set other options
	if config.RunAtLoad {
		pd["RunAtLoad"] = true
	}
	if config.KeepAlive {
		pd["KeepAlive"] = true
	}
	if config.WorkingDir != "" {
		pd["WorkingDirectory"] = config.WorkingDir
	}
	if config.StdoutPath != "" {
		pd["StandardOutPath"] = expandTilde(config.StdoutPath)
	}
	if config.StderrPath != "" {
		pd["StandardErrorPath"] = expandTilde(config.StderrPath)
	}
	if len(config.Environment) > 0 {
		pd["EnvironmentVariables"] = config.Environment
	}

	// Handle schedule
	if config.Schedule != nil {
		if config.Schedule.Interval != nil {
			pd["StartInterval"] = *config.Schedule.Interval
		} else {
			calInterval := make(map[string]int)
			if config.Schedule.Minute != nil {
				calInterval["Minute"] = *config.Schedule.Minute
			}
			if config.Schedule.Hour != nil {
				calInterval["Hour"] = *config.Schedule.Hour
			}
			if config.Schedule.Day != nil {
				calInterval["Day"] = *config.Schedule.Day
			}
			if config.Schedule.Weekday != nil {
				calInterval["Weekday"] = *config.Schedule.Weekday
			}
			if config.Schedule.Month != nil {
				calInterval["Month"] = *config.Schedule.Month
			}
			// Always write StartCalendarInterval when schedule is set;
			// empty map means "every minute" in launchd semantics
			pd["StartCalendarInterval"] = calInterval
		}
	}

	// Marshal to XML plist format
	var buf bytes.Buffer
	encoder := plist.NewEncoder(&buf)
	encoder.Indent("\t")
	if err := encoder.Encode(pd); err != nil {
		return fmt.Errorf("failed to encode plist: %w", err)
	}

	// Write to file
	if err := os.WriteFile(path, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write plist file: %w", err)
	}

	return nil
}

// Ensure UserManager implements Manager interface
var _ Manager = (*UserManager)(nil)
