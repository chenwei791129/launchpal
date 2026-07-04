## MODIFIED Requirements

### Requirement: Read service logs

The system SHALL read log files based on the service's configured `StandardOutPath` or `StandardErrorPath`.
The system SHALL expand `~` in log paths to the user's home directory.
GetLogs SHALL return a structured `LogsResult` value containing `Content` (log tail content), `Status` (one of `"ok"`, `"no-path"`, `"not-found"`), and `Path` (the resolved log path, empty when no path is configured).
When no log path is configured for the requested type, GetLogs SHALL return `LogsResult` with `Status: "no-path"` and a nil error.
When a log path is configured but the file does not exist, GetLogs SHALL return `LogsResult` with `Status: "not-found"`, the resolved path in `Path`, and a nil error.
When the log file exists and is readable, GetLogs SHALL return `LogsResult` with `Status: "ok"`, the tail content in `Content`, and the resolved path in `Path`; an empty file yields `Status: "ok"` with empty `Content`.
The system SHALL return an error for invalid log type (not "stdout" or "stderr").
The system SHALL return an error when the service does not exist.
The system SHALL return an error containing "permission denied" and the resolved path when the log file exists but is not readable.
The system SHALL return an error (not a `Status` value) for any other read failure, such as a configured path that points to a directory.
The classification of the not-found state SHALL be derived from the file-open result (such as `os.IsNotExist`), never from matching error message text.

#### Scenario: Read stdout log

- **WHEN** GetLogs is called with logType="stdout" for a service with StandardOutPath configured and the file exists
- **THEN** the system returns LogsResult with Status "ok" and the log file content in Content

#### Scenario: No log path configured

- **WHEN** GetLogs is called with logType="stdout" for a service with no StandardOutPath
- **THEN** the system returns LogsResult with Status "no-path" and a nil error

#### Scenario: Log file does not exist

- **WHEN** GetLogs is called for a service whose configured log path points to a nonexistent file
- **THEN** the system returns LogsResult with Status "not-found", the resolved path in Path, and a nil error

#### Scenario: Log file not readable

- **WHEN** GetLogs is called for a service whose configured log file exists but is not readable by the current process
- **THEN** the system returns an error whose message contains "permission denied" and the resolved path

#### Scenario: Empty log file

- **WHEN** GetLogs is called for a service whose configured log file exists and is 0 bytes
- **THEN** the system returns LogsResult with Status "ok" and empty Content

#### Scenario: Invalid log type

- **WHEN** GetLogs is called with logType="debug"
- **THEN** the system returns an error indicating invalid log type

#### Scenario: Log path points to a directory

- **WHEN** GetLogs is called for a service whose configured log path points to a directory
- **THEN** the system returns an error (not a LogsResult status)

##### Example: status classification for a user service

| StandardOutPath value        | File state          | Status      | Error                          |
| ---------------------------- | ------------------- | ----------- | ------------------------------ |
| (key absent)                 | —                   | `no-path`   | nil                            |
| `~/Library/Logs/foo/out.log` | does not exist      | `not-found` | nil                            |
| `~/Library/Logs/foo/out.log` | exists, 0 bytes     | `ok`        | nil                            |
| `~/Library/Logs/foo/out.log` | exists, readable    | `ok`        | nil                            |
| `~/Library/Logs/foo/out.log` | exists, mode 000    | —           | contains "permission denied"   |

### Requirement: Clear service logs

The system SHALL expose a `ClearLogs(name, logType)` operation on the Manager interface that truncates the configured log file for the given service to 0 bytes.
The operation SHALL accept `logType` values `stdout` or `stderr` only; any other value SHALL return an error containing the phrase "invalid log type".
The operation SHALL return an error containing the phrase "no log path" when the requested log type has no path configured in the plist.
The operation SHALL return an error containing the phrase "log file does not exist" when the configured log path does not exist on disk; it SHALL NOT create the file.
The operation SHALL truncate by opening the file with write-only, truncating, no-follow semantics (equivalent to `O_WRONLY | O_TRUNC | O_NOFOLLOW`); it SHALL NOT call `os.Remove` on the file.
The operation SHALL preserve the existing file inode, owner, group, and mode.
The user-domain implementation SHALL truncate directly without consulting any privileged helper.

#### Scenario: Truncate user stdout log

- **WHEN** `ClearLogs("com.example", "stdout")` is called for a user service whose stdout log file exists
- **THEN** the file size becomes 0, the inode and mode are unchanged, and a subsequent `GetLogs` returns LogsResult with Status "ok" and empty Content

#### Scenario: Invalid log type rejected

- **WHEN** `ClearLogs("com.example", "trace")` is called
- **THEN** the system returns an error indicating invalid log type and the file is not modified

#### Scenario: Missing log file is an error, not a no-op

- **WHEN** `ClearLogs("com.example", "stderr")` is called for a service whose stderr log path is configured but the file has not yet been created
- **THEN** the system returns an error indicating the log file does not exist and no file is created

#### Scenario: Symlink at log path is rejected

- **WHEN** the configured log path is a symbolic link
- **THEN** the operation returns an error and the link target is not truncated
