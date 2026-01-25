package launchctl

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
	WorkingDir  string            `json:"workingDirectory,omitempty"`
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
