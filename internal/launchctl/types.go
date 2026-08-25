// Package launchctl provides interfaces and implementations for managing
// macOS LaunchAgents and LaunchDaemons via the launchctl command.
package launchctl

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"unicode/utf8"
)

// Service represents a LaunchAgent service
type Service struct {
	Name             string            `json:"name"`
	Label            string            `json:"label"`
	Status           string            `json:"status"` // "running", "stopped", "unknown"
	PID              int               `json:"pid,omitempty"`
	Path             string            `json:"path"`
	Program          string            `json:"program,omitempty"`
	Arguments        []string          `json:"arguments,omitempty"`
	RunAtLoad        bool              `json:"runAtLoad"`
	KeepAlive        KeepAliveConfig   `json:"keepAlive"`
	ThrottleInterval *int              `json:"throttleInterval,omitempty"`
	Schedule         *ScheduleConfig   `json:"schedule,omitempty"`
	Environment      map[string]string `json:"environment,omitempty"`
	StdoutPath       string            `json:"stdoutPath,omitempty"`
	StderrPath       string            `json:"stderrPath,omitempty"`
	WakeSystem       bool              `json:"wakeSystem"`
	WorkingDir       string            `json:"workingDirectory,omitempty"`
	Type             string            `json:"type"`             // "user", "system", "apple-system"
	ReadOnly         bool              `json:"readOnly"`         // true for system services
	PlistFormat      string            `json:"plistFormat"`      // "xml", "binary", "unknown"
	StatusConfidence string            `json:"statusConfidence"` // "verified", "unverified"
}

// CalendarEntry represents a single StartCalendarInterval entry
type CalendarEntry struct {
	Minute  *int `json:"minute,omitempty"`
	Hour    *int `json:"hour,omitempty"`
	Day     *int `json:"day,omitempty"`
	Weekday *int `json:"weekday,omitempty"`
	Month   *int `json:"month,omitempty"`
}

// ScheduleConfig represents StartCalendarInterval or StartInterval
type ScheduleConfig struct {
	Schedules []CalendarEntry `json:"schedules,omitempty"`
	Interval  *int            `json:"interval,omitempty"`
}

// ServiceConfig is used for creating/updating services
type ServiceConfig struct {
	Label            string            `json:"label"`
	Program          string            `json:"program,omitempty"`
	Arguments        []string          `json:"arguments,omitempty"`
	RunAtLoad        bool              `json:"runAtLoad"`
	KeepAlive        KeepAliveConfig   `json:"keepAlive"`
	ThrottleInterval *int              `json:"throttleInterval,omitempty"`
	Schedule         *ScheduleConfig   `json:"schedule,omitempty"`
	Environment      map[string]string `json:"environment,omitempty"`
	WakeSystem       bool              `json:"wakeSystem"`
	StdoutPath       string            `json:"stdoutPath,omitempty"`
	StderrPath       string            `json:"stderrPath,omitempty"`
	WorkingDir       string            `json:"workingDirectory,omitempty"`
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

// Status confidence constants
const (
	ConfidenceVerified   = "verified"
	ConfidenceUnverified = "unverified"
)

// Log type constants
const (
	LogTypeStdout = "stdout"
	LogTypeStderr = "stderr"
)

// LogsResult status constants
const (
	LogStatusOK       = "ok"
	LogStatusNoPath   = "no-path"
	LogStatusNotFound = "not-found"
)

// LogsResult is the structured return value of Manager.GetLogs. Structural
// states — no log path configured, log file not created yet — travel in
// Status rather than the error channel: the Wails bridge can only carry
// strings through Promise rejections, so the frontend must never classify
// by matching error text. Content is meaningful only when Status is "ok".
// Path is the path actually used to open the file (tilde-expanded in the
// user domain, plist-literal in the system domains) and is empty when
// Status is "no-path".
type LogsResult struct {
	Content string `json:"content"`
	Status  string `json:"status"`
	Path    string `json:"path"`
}

// LogClearStatus describes whether a service's log file can be truncated, and
// carries the log file metadata the Logs tab's info row renders.
// LogPath is the resolved (post-tilde-expansion) path or "" when the service
// has no log path configured for the requested type. Exists is whether
// LogPath exists on disk; UserWritable is whether the current process can
// open LogPath for writing without following symlinks. Both Exists and
// UserWritable are false when LogPath is empty.
//
// Size is the file size in bytes, taken from Stat() on the same descriptor
// used for the UserWritable probe. It is 0 in every case where that Stat
// cannot report a meaningful size:
//
//   - empty LogPath (nothing was opened)
//   - the file does not exist (ENOENT)
//   - the open failed for any other reason (EACCES, ELOOP, EISDIR, …)
//   - the file exists and is genuinely empty
//
// A caller therefore cannot distinguish "empty file" from "size unknown" by
// Size alone; pair it with Exists, which is true only in the first case.
type LogClearStatus struct {
	LogPath      string `json:"logPath"`
	Exists       bool   `json:"exists"`
	UserWritable bool   `json:"userWritable"`
	Size         int64  `json:"size"`
}

// ErrReadOnlyManager is returned when attempting write operations on read-only managers
var ErrReadOnlyManager = errors.New("this manager is read-only")

// DeleteServiceOptions tunes SystemManager.DeleteWithOptions. DeleteLogs
// triggers a best-effort cleanup of the daemon's StandardOutPath /
// StandardErrorPath after the plist is removed. Defaults are conservative —
// a zero value preserves the legacy Delete behaviour.
type DeleteServiceOptions struct {
	DeleteLogs bool `json:"deleteLogs"`
}

// LogDeletionWarning signals that DeleteWithOptions removed the daemon's
// plist successfully but the helper's DeleteLogPaths RPC returned one or
// more per-path failures. Callers SHOULD treat this as overall success and
// surface the entries as a non-fatal warning rather than a hard error;
// detect with errors.As.
type LogDeletionWarning struct {
	Errors []string
}

// Error renders the warning as a single-line message; callers that want
// per-path detail should read Errors directly.
func (e *LogDeletionWarning) Error() string {
	if e == nil || len(e.Errors) == 0 {
		return "log deletion completed with warnings"
	}
	return "log deletion completed with warnings: " + strings.Join(e.Errors, "; ")
}

// selectLogPath returns the StandardOutPath or StandardErrorPath of service
// based on logType. Returns "" for an unrecognized logType so callers can
// treat "missing path" and "unknown type" uniformly via the empty-string
// check; the validation against {stdout, stderr} happens earlier in the
// public API surface.
func selectLogPath(service *Service, logType string) string {
	switch logType {
	case LogTypeStdout:
		return service.StdoutPath
	case LogTypeStderr:
		return service.StderrPath
	}
	return ""
}

// validateLogType returns the shared "invalid log type" error for values
// outside {stdout, stderr}, so the wording stays in one place.
func validateLogType(logType string) error {
	if logType != LogTypeStdout && logType != LogTypeStderr {
		return fmt.Errorf("invalid log type: %s (use 'stdout' or 'stderr')", logType)
	}
	return nil
}

// resolveLogPath validates logType and returns the configured path on
// service. Returns the same "invalid log type" and "no log path configured"
// errors used by every Clear caller, so error wording stays in one place.
func resolveLogPath(service *Service, serviceName, logType string) (string, error) {
	if err := validateLogType(logType); err != nil {
		return "", err
	}
	p := selectLogPath(service, logType)
	if p == "" {
		return "", fmt.Errorf("no %s log path configured for service %s", logType, serviceName)
	}
	return p, nil
}

// getServiceLogs is the Get-then-classify scaffolding shared by
// UserManager.GetLogs and readOnlyManager.getLogs (mirroring
// validateClearLogsArgs on the Clear path): an invalid logType stays on the
// error channel, a missing path maps to Status "no-path", a missing file
// maps to Status "not-found", and every other read failure (permission,
// directory, I/O) stays on the error channel. expand applies tilde expansion
// to the configured path (user domain); the system domains open the
// plist-literal path verbatim.
func getServiceLogs(get func(string) (*Service, error), name, logType string, expand bool) (LogsResult, error) {
	service, err := get(name)
	if err != nil {
		return LogsResult{}, err
	}
	if err := validateLogType(logType); err != nil {
		return LogsResult{}, err
	}
	logPath := selectLogPath(service, logType)
	if logPath == "" {
		return LogsResult{Status: LogStatusNoPath}, nil
	}
	if expand {
		logPath = expandTilde(logPath)
	}
	content, err := readLogTail(logPath)
	if err != nil {
		if os.IsNotExist(err) {
			return LogsResult{Status: LogStatusNotFound, Path: logPath}, nil
		}
		return LogsResult{}, err
	}
	return LogsResult{Content: content, Status: LogStatusOK, Path: logPath}, nil
}

// validateClearLogsArgs is the Get-then-resolveLogPath scaffolding shared by
// UserManager.ClearLogs and SystemManager.ClearLogs. The caller performs
// the OpenFile itself because each domain maps errnos differently.
func validateClearLogsArgs(get func(string) (*Service, error), name, logType string, expand bool) (string, error) {
	service, err := get(name)
	if err != nil {
		return "", err
	}
	logPath, err := resolveLogPath(service, name, logType)
	if err != nil {
		return "", err
	}
	if expand {
		logPath = expandTilde(logPath)
	}
	return logPath, nil
}

// canWriteLogFile reports whether the calling process can open path for
// writing without following symlinks. The check uses an actual OpenFile
// attempt (immediately closed) rather than a stat + mode-bit comparison,
// so ACLs, group membership, and SIP all flow through the kernel's
// authoritative answer instead of being re-implemented in user space.
//
// Returns false for missing files and symlinks (O_NOFOLLOW makes the latter
// fail with ELOOP).
func canWriteLogFile(path string) bool {
	if path == "" {
		return false
	}
	f, err := os.OpenFile(path, os.O_WRONLY|nofollowFlag, 0)
	if err != nil {
		return false
	}
	_ = f.Close()
	return true
}

// truncateLogFile opens path with O_WRONLY|O_TRUNC|O_NOFOLLOW and closes it
// immediately. Combining O_TRUNC with O_NOFOLLOW (instead of os.Truncate)
// is what makes a symlink planted at path fail with ELOOP rather than be
// dereferenced. Without O_CREATE, ENOENT surfaces unchanged so callers can
// distinguish "missing" from "permission denied" without an extra Lstat.
func truncateLogFile(path string) error {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_TRUNC|nofollowFlag, 0)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("log file does not exist: %s", path)
		}
		return fmt.Errorf("failed to truncate log file: %w", err)
	}
	return f.Close()
}

// logClearStatusFor builds a LogClearStatus from a resolved (post-tilde)
// path using a single OpenFile attempt. Stat-then-open would race; the
// kernel's response to one OpenFile is the only authoritative answer.
//
//   - success → exists=true, writable=true, size from Stat() on that fd
//   - ENOENT → exists=false, writable=false, size=0
//   - any other error (EACCES, ELOOP, EISDIR, …) → exists=true, writable=false,
//     size=0 (no descriptor was obtained, so no size can be measured)
//
// Size comes from the already-open descriptor rather than a second os.Stat so
// the reported size describes the same file the writability probe answered
// for; a separate path-based Stat would reopen the same race this avoids.
func logClearStatusFor(resolvedPath string) LogClearStatus {
	status := LogClearStatus{LogPath: resolvedPath}
	if resolvedPath == "" {
		return status
	}
	f, err := os.OpenFile(resolvedPath, os.O_WRONLY|nofollowFlag, 0)
	if err == nil {
		// A failing Stat leaves Size at 0, matching the other unknown-size
		// states; the writability answer the open already gave stands.
		if info, statErr := f.Stat(); statErr == nil {
			status.Size = info.Size()
		}
		_ = f.Close()
		status.Exists = true
		status.UserWritable = true
		return status
	}
	if os.IsNotExist(err) {
		return status
	}
	status.Exists = true
	return status
}

// maxLogSize is the maximum number of bytes to read from the tail of a log file (1MB)
const maxLogSize = 1024 * 1024

// readLogTail reads up to the last maxLogSize bytes of a file.
// If the file is smaller than maxLogSize, it reads the entire file.
// A nonexistent path returns the os.Open error unwrapped so callers can
// classify the not-found state with os.IsNotExist instead of matching
// error message text.
func readLogTail(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", err
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
