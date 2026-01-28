package launchctl

import "errors"

// Service represents a LaunchAgent service
type Service struct {
	Name        string            `json:"name"`
	Label       string            `json:"label"`
	Status      string            `json:"status"` // "running", "stopped", "unknown"
	PID         int               `json:"pid,omitempty"`
	Path        string            `json:"path"`
	Program     string            `json:"program,omitempty"`
	Arguments   []string          `json:"arguments,omitempty"`
	RunAtLoad   bool              `json:"runAtLoad"`
	KeepAlive   bool              `json:"keepAlive"`
	Schedule    *ScheduleConfig   `json:"schedule,omitempty"`
	Environment map[string]string `json:"environment,omitempty"`
	StdoutPath  string            `json:"stdoutPath,omitempty"`
	StderrPath  string            `json:"stderrPath,omitempty"`
	WorkingDir   string            `json:"workingDirectory,omitempty"`
	Type         string            `json:"type"`         // "user", "system", "apple-system"
	ReadOnly     bool              `json:"readOnly"`     // true for system services
	PlistFormat  string            `json:"plistFormat"`  // "xml", "binary", "unknown"
}

// ScheduleConfig represents StartCalendarInterval
type ScheduleConfig struct {
	Minute  *int `json:"minute,omitempty"`
	Hour    *int `json:"hour,omitempty"`
	Day     *int `json:"day,omitempty"`
	Weekday *int `json:"weekday,omitempty"`
	Month   *int `json:"month,omitempty"`
}

// ServiceConfig is used for creating/updating services
type ServiceConfig struct {
	Label       string            `json:"label"`
	Program     string            `json:"program,omitempty"`
	Arguments   []string          `json:"arguments,omitempty"`
	RunAtLoad   bool              `json:"runAtLoad"`
	KeepAlive   bool              `json:"keepAlive"`
	Schedule    *ScheduleConfig   `json:"schedule,omitempty"`
	Environment map[string]string `json:"environment,omitempty"`
	StdoutPath  string            `json:"stdoutPath,omitempty"`
	StderrPath  string            `json:"stderrPath,omitempty"`
	WorkingDir  string            `json:"workingDirectory,omitempty"`
}

// ErrReadOnlyManager is returned when attempting write operations on read-only managers
var ErrReadOnlyManager = errors.New("this manager is read-only")

// detectPlistFormat detects whether a plist file is XML or binary format
func detectPlistFormat(data []byte) string {
	if len(data) == 0 {
		return "unknown"
	}
	// Binary plist starts with "bplist"
	if len(data) >= 6 && string(data[0:6]) == "bplist" {
		return "binary"
	}
	// XML plist typically starts with "<?xml" or whitespace followed by "<?xml"
	for i := 0; i < len(data) && i < 100; i++ {
		if data[i] == '<' {
			if len(data) > i+5 && string(data[i:i+5]) == "<?xml" {
				return "xml"
			}
			break
		}
	}
	// If we see printable ASCII, likely XML
	for i := 0; i < len(data) && i < 100; i++ {
		if data[i] < 32 && data[i] != '\n' && data[i] != '\r' && data[i] != '\t' {
			return "binary"
		}
	}
	return "xml"
}
