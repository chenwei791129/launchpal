package launchctl

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"unicode/utf8"
)

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

// ScheduleConfig represents StartCalendarInterval or StartInterval
type ScheduleConfig struct {
	Minute      *int `json:"minute,omitempty"`
	Hour        *int `json:"hour,omitempty"`
	Day         *int `json:"day,omitempty"`
	Weekday     *int `json:"weekday,omitempty"`
	Month       *int `json:"month,omitempty"`
	Interval    *int `json:"interval,omitempty"`
	HasMultiple bool `json:"hasMultiple,omitempty"`
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

// Service type constants
const (
	ServiceTypeUser        = "user"
	ServiceTypeSystem      = "system"
	ServiceTypeAppleSystem = "apple-system"
)

// Service status constants
const (
	StatusRunning = "running"
	StatusStopped = "stopped"
	StatusLoaded  = "loaded"
	StatusUnknown = "unknown"
)

// Log type constants
const (
	LogTypeStdout = "stdout"
	LogTypeStderr = "stderr"
)

// ErrReadOnlyManager is returned when attempting write operations on read-only managers
var ErrReadOnlyManager = errors.New("this manager is read-only")

// parseKeepAlive converts a plist KeepAlive value (bool or dict) to a bool
func parseKeepAlive(v any) bool {
	switch v := v.(type) {
	case bool:
		return v
	case map[string]any:
		return true
	}
	return false
}

// maxLogSize is the maximum number of bytes to read from the tail of a log file (1MB)
const maxLogSize = 1024 * 1024

// readLogTail reads up to the last maxLogSize bytes of a file.
// If the file is smaller than maxLogSize, it reads the entire file.
func readLogTail(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("log file not found: %s", path)
		}
		if os.IsPermission(err) {
			return "", fmt.Errorf("permission denied reading log file: %s", path)
		}
		return "", fmt.Errorf("failed to read log file: %w", err)
	}
	defer func() { _ = f.Close() }()

	info, err := f.Stat()
	if err != nil {
		return "", fmt.Errorf("failed to stat log file: %w", err)
	}

	size := info.Size()
	if size <= maxLogSize {
		data, err := io.ReadAll(f)
		if err != nil {
			return "", fmt.Errorf("failed to read log file: %w", err)
		}
		return string(data), nil
	}

	// Seek to the last maxLogSize bytes
	if _, err := f.Seek(-maxLogSize, io.SeekEnd); err != nil {
		return "", fmt.Errorf("failed to seek log file: %w", err)
	}

	data, err := io.ReadAll(f)
	if err != nil {
		return "", fmt.Errorf("failed to read log file: %w", err)
	}

	// Skip to the first newline to avoid a partial first line.
	// Since '\n' (0x0A) cannot appear inside a multi-byte UTF-8 sequence,
	// the byte after '\n' is always a valid UTF-8 character boundary.
	if idx := bytes.IndexByte(data, '\n'); idx >= 0 {
		data = data[idx+1:]
	} else {
		// No newline found (single line > maxLogSize): skip any leading
		// incomplete UTF-8 bytes to avoid garbled characters.
		for len(data) > 0 && !utf8.RuneStart(data[0]) {
			data = data[1:]
		}
	}

	return string(data), nil
}

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
