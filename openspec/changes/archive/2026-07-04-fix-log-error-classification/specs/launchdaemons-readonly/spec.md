## MODIFIED Requirements

### Requirement: Read system service logs

SystemManager.GetLogs and AppleSystemManager.GetLogs SHALL read log files based on the service's configured StandardOutPath or StandardErrorPath.
Both SHALL return a structured `LogsResult` value containing `Content` (log tail content), `Status` (one of `"ok"`, `"no-path"`, `"not-found"`), and `Path`, using the same status classification rules as the user domain.
`Path` SHALL be the path actually used to open the file: the system and apple-system domains SHALL use the configured plist path verbatim without tilde expansion (unlike the user domain), and `Path` SHALL be empty when no path is configured.
When no log path is configured for the requested type, both SHALL return `LogsResult` with `Status: "no-path"` and a nil error.
When a log path is configured but the file does not exist, both SHALL return `LogsResult` with `Status: "not-found"`, the resolved path in `Path`, and a nil error.
When the log file exists and is readable, both SHALL return `LogsResult` with `Status: "ok"` and the tail content in `Content`.
Both SHALL return an error for invalid log type (not "stdout" or "stderr").
Both SHALL return an error containing "permission denied" and the log path when the log file exists but is not readable.
Both SHALL return an error (not a `Status` value) for any other read failure, such as a configured path that points to a directory.
The classification of the not-found state SHALL be derived from the file-open result (such as `os.IsNotExist`), never from matching error message text.

#### Scenario: Read log with no path configured

- **WHEN** GetLogs is called for a system service with no StandardOutPath
- **THEN** the system returns LogsResult with Status "no-path" and a nil error

#### Scenario: Log file does not exist

- **WHEN** GetLogs is called and the configured log file does not exist
- **THEN** the system returns LogsResult with Status "not-found", the resolved path in Path, and a nil error

#### Scenario: Log file permission denied

- **WHEN** GetLogs is called and the configured log file exists but is not readable
- **THEN** the system returns an error whose message contains "permission denied" and the log path

#### Scenario: Log path points to a directory

- **WHEN** GetLogs is called and the configured log path points to a directory
- **THEN** the system returns an error (not a LogsResult status)

#### Scenario: Readable log file

- **WHEN** GetLogs is called and the configured log file exists and is readable
- **THEN** the system returns LogsResult with Status "ok" and the log tail content in Content
