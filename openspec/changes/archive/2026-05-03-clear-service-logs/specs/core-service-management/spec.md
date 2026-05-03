## ADDED Requirements

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
- **THEN** the file size becomes 0, the inode and mode are unchanged, and a subsequent `GetLogs` returns an empty string

#### Scenario: Invalid log type rejected

- **WHEN** `ClearLogs("com.example", "trace")` is called
- **THEN** the system returns an error indicating invalid log type and the file is not modified

#### Scenario: Missing log file is an error, not a no-op

- **WHEN** `ClearLogs("com.example", "stderr")` is called for a service whose stderr log path is configured but the file has not yet been created
- **THEN** the system returns an error indicating the log file does not exist and no file is created

#### Scenario: Symlink at log path is rejected

- **WHEN** the configured log path is a symbolic link
- **THEN** the operation returns an error and the link target is not truncated

### Requirement: Query log clear authorization status

The system SHALL expose a `GetLogClearStatus(name, logType)` operation that returns three pieces of information for the given service and log type:
1. `logPath`: the resolved log file path (after `~` expansion for user services), or empty string when no path is configured.
2. `exists`: whether `logPath` exists on disk; SHALL be false when `logPath` is empty.
3. `userWritable`: whether the current process can open `logPath` for writing without following symlinks. SHALL be false when `logPath` is empty or `exists` is false.

The operation SHALL determine `userWritable` by attempting `os.OpenFile(logPath, O_WRONLY|O_NOFOLLOW, 0)` and immediately closing the file if the open succeeds; it SHALL NOT rely on `os.Stat` plus mode-bit comparison.
The operation SHALL succeed (return a `LogClearStatus` value, not an error) even when `logPath` is empty, the file does not exist, or the file is not writable; the structure fields convey the state.
The operation SHALL only return an error for cases that prevent producing a meaningful status, such as the service not existing.

#### Scenario: Path exists and writable

- **WHEN** `GetLogClearStatus("com.example", "stdout")` is called and the log file exists with mode that allows the current user to write
- **THEN** the result is `{logPath: "<resolved path>", exists: true, userWritable: true}` with no error

#### Scenario: Path configured but file missing

- **WHEN** the log path is configured in the plist but no file exists at that location
- **THEN** the result is `{logPath: "<resolved path>", exists: false, userWritable: false}` with no error

#### Scenario: No log path in plist

- **WHEN** the plist does not configure `StandardOutPath` for the requested log type
- **THEN** the result is `{logPath: "", exists: false, userWritable: false}` with no error

#### Scenario: Service not found

- **WHEN** `GetLogClearStatus` is called for a service name that does not exist in the manager's domain
- **THEN** an error is returned and the result is not consulted

## MODIFIED Requirements

### Requirement: Read service logs

The system SHALL read log files based on the service's configured `StandardOutPath` or `StandardErrorPath`.
The system SHALL expand `~` in log paths to the user's home directory.
The system SHALL return an error for invalid log type (not "stdout" or "stderr").
The system SHALL return an error when no log path is configured for the requested type.
The system SHALL return an error when the log file does not exist.
The system SHALL apply the same `~` expansion and `logType` validation rules to `ClearLogs` and `GetLogClearStatus`.

#### Scenario: Read stdout log

- **WHEN** GetLogs is called with logType="stdout" for a service with StandardOutPath configured
- **THEN** the content of the stdout log file is returned

#### Scenario: No log path configured

- **WHEN** GetLogs is called with logType="stdout" for a service with no StandardOutPath
- **THEN** the system returns an error indicating no stdout log path is configured

#### Scenario: Invalid log type

- **WHEN** GetLogs is called with logType="debug"
- **THEN** the system returns an error indicating invalid log type

#### Scenario: Tilde expansion shared across log operations

- **WHEN** any of `GetLogs`, `ClearLogs`, or `GetLogClearStatus` is called with a log path beginning with `~/`
- **THEN** the path is resolved against the current user's home directory before any filesystem operation
