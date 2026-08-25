## ADDED Requirements

### Requirement: List user services

The system SHALL list all `.plist` files in `~/Library/LaunchAgents` and return a `Service` struct for each.
The system SHALL return an empty list (not an error) when the directory does not exist.
The system SHALL skip non-plist files and directories.
The system SHALL skip services whose plist cannot be parsed, continuing with the remaining services.

#### Scenario: Directory with mixed plist files

- **WHEN** `~/Library/LaunchAgents` contains `com.user.app.plist`, `com.user.backup.plist`, `.DS_Store`, and a subdirectory
- **THEN** the system returns exactly 2 services corresponding to the two `.plist` files

#### Scenario: Empty or nonexistent directory

- **WHEN** `~/Library/LaunchAgents` does not exist
- **THEN** the system returns an empty list with no error

### Requirement: Get service details from plist

The system SHALL read a single service by name, loading its plist from `~/Library/LaunchAgents/<name>.plist`.
The system SHALL parse the following fields from the plist: Label, Program, ProgramArguments, RunAtLoad, KeepAlive, StartCalendarInterval, StartInterval, EnvironmentVariables, StandardOutPath, StandardErrorPath, WorkingDirectory.
The system SHALL set `Type` to `"user"` and `ReadOnly` to `false`.
The system SHALL detect plist format (xml or binary) and populate `PlistFormat`.
The system SHALL handle `KeepAlive` as either a boolean value or a dictionary (dictionary treated as `true`).

#### Scenario: Valid XML plist with all fields

- **WHEN** a plist file exists with Label, Program, RunAtLoad, KeepAlive (bool), EnvironmentVariables, StandardOutPath, StandardErrorPath, and WorkingDirectory
- **THEN** the returned Service struct contains all parsed fields, Type is `"user"`, ReadOnly is `false`, and PlistFormat is `"xml"`

#### Scenario: Plist with KeepAlive as dictionary

- **WHEN** a plist contains `KeepAlive` as a dictionary (e.g., `{SuccessfulExit: false}`)
- **THEN** the Service's KeepAlive field is `true`

#### Scenario: Nonexistent plist file

- **WHEN** the requested service name has no corresponding plist file
- **THEN** the system returns an error indicating the file could not be read


<!-- @trace
source: advanced-keepalive-options
updated: 2026-06-01
code:
  - frontend/app/types/wails.d.ts
  - frontend/app/components/LaunchPolicyForm.vue
  - frontend/app/pages/services/[name].vue
  - internal/launchctl/keepalive.go
  - frontend/app/components/ServiceSummary.vue
  - frontend/app/components/ServiceRow.vue
  - internal/launchctl/types.go
  - frontend/app/components/CreateServiceModal.vue
  - frontend/app/utils/serviceToConfig.ts
  - internal/launchctl/plist_encode.go
  - internal/launchctl/user.go
  - internal/launchctl/readonly.go
  - frontend/app/utils/launchPolicy.ts
  - frontend/wailsjs/go/models.ts
tests:
  - frontend/app/components/__tests__/CreateServiceModal.test.ts
  - frontend/app/utils/__tests__/serviceToConfig.test.ts
  - internal/launchctl/plist_encode_test.go
  - frontend/app/components/__tests__/LaunchPolicyForm.test.ts
  - internal/launchctl/keepalive_test.go
  - frontend/app/pages/services/__tests__/edit-launch-policy.test.ts
  - internal/launchctl/user_test.go
  - frontend/app/components/__tests__/CloneUserService.test.ts
  - frontend/app/components/__tests__/ServiceSummary.test.ts
  - frontend/app/components/__tests__/ServiceRow.test.ts
  - frontend/app/utils/__tests__/launchPolicy.test.ts
-->

### Requirement: Parse schedule configuration

The system SHALL parse `StartCalendarInterval` (single dict or array of dicts) and `StartInterval` (integer seconds) from plist data into a `ScheduleConfig`.
The system SHALL return `nil` when neither field is present.
The system SHALL prefer `StartInterval` when both are present (check `StartInterval > 0` first).
For arrays of `StartCalendarInterval`, the system SHALL use the first entry and set `HasMultiple` to `true` when there are more than one.

#### Scenario: Single StartCalendarInterval

- **WHEN** plist contains `StartCalendarInterval` with `{Hour: 9, Minute: 0}`
- **THEN** ScheduleConfig has Hour=9, Minute=0, HasMultiple=false

#### Scenario: Multiple StartCalendarInterval entries

- **WHEN** plist contains `StartCalendarInterval` as an array of 3 entries
- **THEN** ScheduleConfig uses the first entry's values and HasMultiple=true

#### Scenario: StartInterval

- **WHEN** plist contains `StartInterval` with value 300
- **THEN** ScheduleConfig has Interval=300

#### Scenario: No schedule

- **WHEN** plist contains neither StartCalendarInterval nor StartInterval
- **THEN** Schedule is nil

### Requirement: Validate schedule configuration

The system SHALL reject a `StartInterval` value less than 10 seconds with an error.
The system SHALL accept a nil schedule without error.

#### Scenario: Interval too small

- **WHEN** a ServiceConfig has Schedule.Interval = 5
- **THEN** the system returns an error stating "StartInterval must be at least 10 seconds"

#### Scenario: Nil schedule

- **WHEN** a ServiceConfig has nil Schedule
- **THEN** validation passes with no error

### Requirement: Write plist from ServiceConfig

The system SHALL serialize a `ServiceConfig` into XML plist format and write it to the specified path with `0644` permissions.
The system SHALL only include fields that are set (non-zero/non-empty).
The system SHALL expand `~` in StdoutPath and StderrPath to the user's home directory before writing.
The system SHALL write `StartCalendarInterval` as an empty dict when Schedule is set with no Interval and no calendar fields (launchd interprets this as "every minute").

#### Scenario: Minimal service config

- **WHEN** a ServiceConfig has only Label and Program
- **THEN** the written plist contains only Label and Program keys

#### Scenario: Config with schedule (StartInterval)

- **WHEN** a ServiceConfig has Schedule with Interval=60
- **THEN** the written plist contains `StartInterval` key with value 60

#### Scenario: Config with calendar schedule

- **WHEN** a ServiceConfig has Schedule with Hour=9 and Minute=30
- **THEN** the written plist contains `StartCalendarInterval` with `{Hour: 9, Minute: 30}`

#### Scenario: Empty calendar schedule

- **WHEN** a ServiceConfig has Schedule set but with no Interval and no calendar fields
- **THEN** the written plist contains `StartCalendarInterval` as an empty dict


<!-- @trace
source: advanced-keepalive-options
updated: 2026-06-01
code:
  - frontend/app/types/wails.d.ts
  - frontend/app/components/LaunchPolicyForm.vue
  - frontend/app/pages/services/[name].vue
  - internal/launchctl/keepalive.go
  - frontend/app/components/ServiceSummary.vue
  - frontend/app/components/ServiceRow.vue
  - internal/launchctl/types.go
  - frontend/app/components/CreateServiceModal.vue
  - frontend/app/utils/serviceToConfig.ts
  - internal/launchctl/plist_encode.go
  - internal/launchctl/user.go
  - internal/launchctl/readonly.go
  - frontend/app/utils/launchPolicy.ts
  - frontend/wailsjs/go/models.ts
tests:
  - frontend/app/components/__tests__/CreateServiceModal.test.ts
  - frontend/app/utils/__tests__/serviceToConfig.test.ts
  - internal/launchctl/plist_encode_test.go
  - frontend/app/components/__tests__/LaunchPolicyForm.test.ts
  - internal/launchctl/keepalive_test.go
  - frontend/app/pages/services/__tests__/edit-launch-policy.test.ts
  - internal/launchctl/user_test.go
  - frontend/app/components/__tests__/CloneUserService.test.ts
  - frontend/app/components/__tests__/ServiceSummary.test.ts
  - frontend/app/components/__tests__/ServiceRow.test.ts
  - frontend/app/utils/__tests__/launchPolicy.test.ts
-->

### Requirement: CRUD operations for user services

The system SHALL create a new service by writing a plist to `~/Library/LaunchAgents/<label>.plist`.
The system SHALL reject creation when the label is empty.
The system SHALL reject creation when a service with the same label already exists.
The system SHALL reject creation when both `Program` is empty AND `Arguments` is empty, returning an error indicating that the service must specify either Program or at least one argument in Arguments.
The system SHALL reject update when both `Program` is empty AND `Arguments` is empty, returning the same error.
The system SHALL ensure the LaunchAgents directory exists before creating.
The system SHALL ensure log directories exist for configured stdout/stderr paths.
The system SHALL update an existing service by stopping it first, then writing the updated plist.
The system SHALL delete a service by stopping it first, then removing the plist file.

#### Scenario: Create a new service

- **WHEN** Create is called with a valid ServiceConfig (label="com.user.test", program="/usr/bin/test")
- **THEN** a plist file is created at `~/Library/LaunchAgents/com.user.test.plist`

#### Scenario: Create with duplicate label

- **WHEN** Create is called with a label that already has a plist file
- **THEN** the system returns an error indicating the service already exists

#### Scenario: Create with empty label

- **WHEN** Create is called with an empty label
- **THEN** the system returns an error indicating "service label is required"

#### Scenario: Create with only Arguments and no Program

- **WHEN** Create is called with a ServiceConfig where Program is empty but Arguments is a non-empty array (e.g., arguments=["/usr/bin/open", "/Applications/Synology Drive Client.app"])
- **THEN** the plist is written successfully and contains `ProgramArguments` but no `Program` key

#### Scenario: Create with neither Program nor Arguments

- **WHEN** Create is called with a ServiceConfig where Program is empty AND Arguments is empty or nil
- **THEN** the system returns an error indicating that the service must specify either Program or at least one argument in Arguments
- **AND** no plist file is written

#### Scenario: Update with neither Program nor Arguments

- **WHEN** Update is called for an existing service with a ServiceConfig where Program is empty AND Arguments is empty or nil
- **THEN** the system returns an error indicating that the service must specify either Program or at least one argument in Arguments
- **AND** the existing plist file is not modified

#### Scenario: Delete an existing service

- **WHEN** Delete is called for an existing service
- **THEN** the plist file is removed from disk

#### Scenario: Delete nonexistent service

- **WHEN** Delete is called for a service that does not exist
- **THEN** the system returns an error indicating "service not found"

### Requirement: Read raw plist content

The system SHALL return the raw file content of `~/Library/LaunchAgents/<name>.plist` as a string.

#### Scenario: Read existing plist

- **WHEN** GetPlist is called for an existing service
- **THEN** the raw XML content of the plist file is returned

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


<!-- @trace
source: clear-service-logs
updated: 2026-05-03
code:
  - internal/launchctl/system.go
  - internal/launchctl/manager.go
  - frontend/wailsjs/go/models.ts
  - internal/privhelper/client.go
  - frontend/wailsjs/go/main/App.js
  - internal/launchctl/user.go
  - frontend/vitest.setup.ts
  - frontend/app/types/wails.d.ts
  - internal/launchctl/apple_system.go
  - internal/launchctl/nofollow_other.go
  - README.md
  - app.go
  - frontend/package.json.md5
  - frontend/app/components/ServiceLogs.vue
  - internal/launchctl/nofollow_darwin.go
  - internal/launchctl/readonly.go
  - .github/workflows/build.yml
  - frontend/app/pages/services/[name].vue
  - internal/launchctl/types.go
  - internal/privhelper/protocol.go
  - frontend/wailsjs/go/main/App.d.ts
  - internal/privhelper/handlers.go
  - CHANGELOG.md
tests:
  - internal/launchctl/apple_system_test.go
  - internal/launchctl/system_test.go
  - internal/launchctl/types_test.go
  - internal/privhelper/handlers_test.go
  - internal/privhelper/client_test.go
  - internal/privhelper/protocol_test.go
  - app_test.go
  - internal/launchctl/user_test.go
  - frontend/app/components/__tests__/ServiceLogs.test.ts
-->


<!-- @trace
source: fix-log-error-classification
updated: 2026-07-04
code:
  - frontend/app/types/wails.d.ts
  - frontend/wailsjs/go/models.ts
  - internal/launchctl/system.go
  - internal/launchctl/types.go
  - frontend/wailsjs/go/main/App.d.ts
  - internal/launchctl/user.go
  - app.go
  - internal/launchctl/readonly.go
  - frontend/app/components/ServiceLogs.vue
  - internal/launchctl/apple_system.go
  - internal/launchctl/manager.go
tests:
  - internal/launchctl/apple_system_test.go
  - internal/launchctl/system_test.go
  - frontend/app/components/__tests__/ServiceLogs.test.ts
  - internal/launchctl/user_test.go
-->

### Requirement: Backup creation

The system SHALL create a backup by copying the plist file to `~/.launchpal/backups/<service-name>/<timestamp>.plist`.
The system SHALL generate the backup ID using the format `YYYYMMDD-HHMMSS`.
The system SHALL write a metadata file `<timestamp>.meta.json` containing the original plist path.
The system SHALL automatically prune backups after creation, keeping only the 10 most recent.

#### Scenario: Create a backup

- **WHEN** Create is called with serviceName="com.user.app" and a valid plist path
- **THEN** a `.plist` file and `.meta.json` file are created in `~/.launchpal/backups/com.user.app/`

#### Scenario: Auto-prune after creation

- **WHEN** a service already has 12 backups and a new backup is created
- **THEN** only the 10 most recent backups remain (2 oldest are deleted along with their metadata)

### Requirement: List backups

The system SHALL list all backups for a specific service, sorted by timestamp descending (newest first).
The system SHALL list all backups across all services when ListAll is called, also sorted newest first.
The system SHALL return an empty list (not an error) when no backups exist.
The system SHALL skip files that do not match the `YYYYMMDD-HHMMSS.plist` naming convention.

#### Scenario: List backups for a service

- **WHEN** List is called for a service with 3 backups
- **THEN** 3 Backup structs are returned, ordered newest first

#### Scenario: List all backups

- **WHEN** ListAll is called and two services each have 2 backups
- **THEN** 4 Backup structs are returned, globally sorted newest first

#### Scenario: No backups exist

- **WHEN** List is called for a service with no backups directory
- **THEN** an empty list is returned with no error

### Requirement: Get backup content

The system SHALL return the file content of a specific backup identified by service name and backup ID.
The system SHALL return an error when the backup does not exist.

#### Scenario: Get existing backup content

- **WHEN** GetContent is called with a valid service name and backup ID
- **THEN** the plist content of the backup file is returned as a string

#### Scenario: Get nonexistent backup

- **WHEN** GetContent is called with an invalid backup ID
- **THEN** the system returns an error indicating "backup not found"

### Requirement: Restore backup

The system SHALL restore a backup by copying the backup plist file to the specified target path.
The system SHALL return an error when the backup does not exist.

#### Scenario: Restore a backup

- **WHEN** Restore is called with a valid backup ID and target path
- **THEN** the backup file content is copied to the target path

#### Scenario: Restore nonexistent backup

- **WHEN** Restore is called with an invalid backup ID
- **THEN** the system returns an error indicating "backup not found"

### Requirement: Start user service via bootstrap

The system SHALL start a user service by executing `launchctl bootstrap gui/<uid> <plistPath>` where `<uid>` is the current user's UID obtained via `os.Getuid()`.
The system SHALL return an error when the plist file does not exist.
The system SHALL return an error with the launchctl output when bootstrap fails (e.g., service already loaded).

#### Scenario: Start a stopped service

- **WHEN** Start is called for an existing service that is not currently loaded
- **THEN** the system executes `launchctl bootstrap gui/<uid> <plistPath>` and returns no error

#### Scenario: Start an already loaded service

- **WHEN** Start is called for a service that is already loaded
- **THEN** the system returns an error containing the launchctl error output

#### Scenario: Start a nonexistent service

- **WHEN** Start is called for a service whose plist file does not exist
- **THEN** the system returns an error indicating "service not found"

### Requirement: Stop user service via bootout

The system SHALL stop a user service by executing `launchctl bootout gui/<uid>/<label>` where `<uid>` is the current user's UID and `<label>` is the service's Label from the plist.
The system SHALL ignore errors from bootout (the service may not be loaded).
The system SHALL attempt to kill the process via pgrep/kill as a fallback if the service program is still running after bootout.
The system SHALL skip pgrep fallback for common shell programs (`/bin/bash`, `/bin/sh`, `/bin/zsh` and their `/usr/bin` variants).

#### Scenario: Stop a running service

- **WHEN** Stop is called for a loaded service with label "com.user.app"
- **THEN** the system executes `launchctl bootout gui/<uid>/com.user.app`

#### Scenario: Stop a service that is not loaded

- **WHEN** Stop is called for a service that is not currently loaded
- **THEN** the system ignores the bootout error and returns no error

#### Scenario: Stop with fallback kill

- **WHEN** Stop is called and the service process is still running after bootout
- **THEN** the system uses pgrep to find and kill the process

### Requirement: GUI domain helper

The system SHALL provide a helper function that returns the `gui/<uid>` domain string using the current process's UID.
The system SHALL format the domain as `gui/<uid>` where `<uid>` is a decimal integer.

#### Scenario: Get GUI domain for current user

- **WHEN** the helper function is called by a process running as UID 501
- **THEN** it returns "gui/501"

### Requirement: Detect plist format

The system SHALL detect whether a plist file is in XML or binary format.
The system SHALL identify binary format when the file starts with the `bplist` magic bytes.
The system SHALL identify XML format when the file starts with `<?xml`.
The system SHALL return `"unknown"` for empty files.

#### Scenario: Binary plist

- **WHEN** file content starts with `bplist00`
- **THEN** detectPlistFormat returns `"binary"`

#### Scenario: XML plist

- **WHEN** file content starts with `<?xml version="1.0"`
- **THEN** detectPlistFormat returns `"xml"`

#### Scenario: Empty file

- **WHEN** file content is empty
- **THEN** detectPlistFormat returns `"unknown"`

## Requirements

### Requirement: Summary tab content is scrollable

The Summary tab panel SHALL be vertically scrollable when content exceeds the visible area, consistent with the Edit and Inspect tabs.

#### Scenario: Summary tab with overflowing content

- **WHEN** a service has enough detail (environment variables, schedule, paths, logs) that the Summary tab content exceeds the viewport height
- **THEN** a vertical scrollbar SHALL appear allowing the user to scroll to see all content

#### Scenario: Summary tab with minimal content

- **WHEN** a service has minimal detail that fits within the viewport
- **THEN** no scrollbar SHALL appear and all content SHALL be visible without scrolling

<!-- @trace
source: summary-tab-scrollable
updated: 2026-04-14
code:
  - frontend/app/pages/services/[name].vue
  - internal/launchctl/user.go
  - frontend/app/components/CreateServiceModal.vue
  - frontend/vitest.setup.ts
  - internal/launchctl/types.go
  - frontend/wailsjs/go/models.ts
  - frontend/package.json.md5
  - CHANGELOG.md
  - frontend/app/components/ScheduleForm.vue
  - frontend/app/types/wails.d.ts
  - README.md
  - frontend/package.json
  - frontend/vitest.config.ts
  - frontend/app/components/ServiceSummary.vue
  - frontend/app/composables/useNextOccurrences.ts
tests:
  - frontend/app/components/__tests__/ServiceSummary.test.ts
  - frontend/app/pages/services/__tests__/edit-env-masking.test.ts
  - frontend/app/components/__tests__/CreateServiceModal.test.ts
  - frontend/app/composables/__tests__/useNextOccurrences.test.ts
  - internal/launchctl/user_test.go
-->

---
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


<!-- @trace
source: fix-log-error-classification
updated: 2026-07-04
code:
  - frontend/app/types/wails.d.ts
  - frontend/wailsjs/go/models.ts
  - internal/launchctl/system.go
  - internal/launchctl/types.go
  - frontend/wailsjs/go/main/App.d.ts
  - internal/launchctl/user.go
  - app.go
  - internal/launchctl/readonly.go
  - frontend/app/components/ServiceLogs.vue
  - internal/launchctl/apple_system.go
  - internal/launchctl/manager.go
tests:
  - internal/launchctl/apple_system_test.go
  - internal/launchctl/system_test.go
  - frontend/app/components/__tests__/ServiceLogs.test.ts
  - internal/launchctl/user_test.go
-->

---
### Requirement: Query log clear authorization status

The system SHALL expose a `GetLogClearStatus(name, logType)` operation that returns four pieces of information for the given service and log type:
1. `logPath`: the resolved log file path (after `~` expansion for user services), or empty string when no path is configured.
2. `exists`: whether `logPath` exists on disk; SHALL be false when `logPath` is empty.
3. `userWritable`: whether the current process can open `logPath` for writing without following symlinks. SHALL be false when `logPath` is empty or `exists` is false.
4. `size`: the size of `logPath` in bytes as reported by `Stat()` on the open file descriptor; SHALL be `0` when `logPath` is empty, the file does not exist, or the file is not openable for read/write without following symlinks.

The operation SHALL determine `userWritable` by attempting `os.OpenFile(logPath, O_WRONLY|O_NOFOLLOW, 0)` and immediately closing the file if the open succeeds; it SHALL NOT rely on `os.Stat` plus mode-bit comparison.
The operation SHALL determine `size` by calling `Stat()` on the same file descriptor used for the `userWritable` probe before closing it; the operation SHALL NOT issue a separate `os.Stat` call.
The operation SHALL succeed (return a `LogClearStatus` value, not an error) even when `logPath` is empty, the file does not exist, or the file is not writable; the structure fields convey the state.
The operation SHALL only return an error for cases that prevent producing a meaningful status, such as the service not existing.

#### Scenario: Path exists and writable

- **WHEN** `GetLogClearStatus("com.example", "stdout")` is called and the log file exists with mode that allows the current user to write
- **THEN** the result is `{logPath: "<resolved path>", exists: true, userWritable: true, size: <byte count>}` with no error

##### Example: 2.4 MB log

- **GIVEN** the resolved stdout path is `/Users/alice/Library/Logs/com.example.stdout.log` with file size 2516582 bytes
- **WHEN** `GetLogClearStatus("com.example", "stdout")` is called
- **THEN** the result is `{logPath: "/Users/alice/Library/Logs/com.example.stdout.log", exists: true, userWritable: true, size: 2516582}`

#### Scenario: Path configured but file missing

- **WHEN** the log path is configured in the plist but no file exists at that location
- **THEN** the result is `{logPath: "<resolved path>", exists: false, userWritable: false, size: 0}` with no error

#### Scenario: No log path in plist

- **WHEN** the plist does not configure `StandardOutPath` for the requested log type
- **THEN** the result is `{logPath: "", exists: false, userWritable: false, size: 0}` with no error

#### Scenario: Service not found

- **WHEN** `GetLogClearStatus` is called for a service name that does not exist in the manager's domain
- **THEN** an error is returned and the result is not consulted

#### Scenario: Path exists but not writable by current user

- **WHEN** `GetLogClearStatus` is called for a system service whose log path exists but is owned by root with mode 0600
- **THEN** the result is `{logPath: "<resolved path>", exists: true, userWritable: false, size: 0}` with no error


<!-- @trace
source: add-log-file-info-bar
updated: 2026-08-25
code:
  - internal/launchctl/types.go
  - frontend/app/components/LogStorageSection.vue
  - .agents/skills/spectra-ingest/SKILL.md
  - admin_mode.go
  - frontend/app/components/ReadOnlyServiceList.vue
  - frontend/app/components/InlineBanner.vue
  - frontend/app/pages/settings.vue
  - .agents/skills/spectra-ask/SKILL.md
  - internal/launchctl/user.go
  - go.sum
  - internal/launchctl/nofollow_other.go
  - frontend/app/utils/formatters.ts
  - internal/privhelper/handlers.go
  - internal/launchctl/system.go
  - frontend/package.json
  - .agents/skills/spectra-audit/SKILL.md
  - internal/launchctl/apple_system.go
  - frontend/app/pages/system.vue
  - .github/workflows/release-please.yml
  - frontend/wailsjs/go/main/App.js
  - .agents/skills/spectra-apply/SKILL.md
  - internal/privhelper/protocol.go
  - .github/workflows/build.yml
  - internal/privhelper/install.go
  - app.go
  - internal/privhelper/server.go
  - frontend/wailsjs/go/models.ts
  - internal/launchctl/manager.go
  - .agents/skills/spectra-propose/SKILL.md
  - AGENTS.md
  - .agents/skills/spectra-archive/SKILL.md
  - internal/privhelper/logpath_darwin.go
  - .agents/skills/spectra-debug/SKILL.md
  - frontend/app/utils/launchPolicy.ts
  - frontend/app/utils/ansiToHtml.ts
  - frontend/vitest.setup.ts
  - README.md
  - frontend/wailsjs/go/main/App.d.ts
  - frontend/app/utils/logPaths.ts
  - CHANGELOG.md
  - internal/launchctl/nofollow_darwin.go
  - cmd/launchpal-privhelper/main.go
  - frontend/app/components/LaunchPolicyForm.vue
  - cmd/launchpal-privhelper/procinfo_other.go
  - internal/launchctl/plist_encode.go
  - frontend/app/composables/useAdminMode.ts
  - internal/privhelper/logpath_other.go
  - internal/privhelper/client.go
  - Makefile
  - frontend/app/utils/serviceToConfig.ts
  - main.go
  - frontend/app/components/ServiceSummary.vue
  - .agents/skills/spectra-commit/SKILL.md
  - frontend/pnpm-workspace.yaml
  - frontend/app/utils/serviceValidation.ts
  - cmd/launchpal-privhelper/procinfo_darwin.go
  - frontend/app/pages/services/[name].vue
  - internal/launchctl/readonly.go
  - frontend/app/composables/useSettings.ts
  - frontend/app/components/ServiceLogs.vue
  - frontend/package.json.md5
  - .agents/skills/spectra-drift/SKILL.md
  - internal/settings/settings.go
  - go.mod
  - internal/privhelper/integrity.go
  - frontend/app/components/CreateServiceModal.vue
  - internal/launchctl/keepalive.go
  - frontend/app/components/ServiceRow.vue
  - frontend/app/types/wails.d.ts
  - frontend/app/utils/settingsValidation.ts
  - .agents/skills/spectra-discuss/SKILL.md
  - internal/privhelper/logpath.go
tests:
  - internal/privhelper/integrity_test.go
  - internal/launchctl/system_test.go
  - internal/privhelper/server_test.go
  - cmd/launchpal-privhelper/helper_test.go
  - internal/privhelper/protocol_test.go
  - frontend/app/pages/services/__tests__/edit-program-arguments-validation.test.ts
  - frontend/app/components/__tests__/ServiceSummary.test.ts
  - internal/launchctl/types_test.go
  - frontend/app/components/__tests__/CloneUserService.test.ts
  - frontend/app/utils/__tests__/launchPolicy.test.ts
  - frontend/app/components/__tests__/ServiceLogs.test.ts
  - frontend/app/pages/services/__tests__/edit-launch-policy.test.ts
  - internal/privhelper/handlers_test.go
  - admin_mode_test.go
  - frontend/app/pages/__tests__/settings.test.ts
  - frontend/app/components/__tests__/LaunchPolicyForm.test.ts
  - frontend/app/utils/__tests__/ansiToHtml.test.ts
  - frontend/app/utils/__tests__/serviceToConfig.test.ts
  - internal/launchctl/keepalive_test.go
  - frontend/app/components/__tests__/CreateServiceModal.test.ts
  - app_test.go
  - frontend/app/utils/__tests__/formatters.test.ts
  - internal/launchctl/plist_encode_test.go
  - internal/privhelper/install_test.go
  - internal/settings/settings_test.go
  - frontend/app/composables/__tests__/useSettings.test.ts
  - internal/launchctl/user_test.go
  - internal/privhelper/client_test.go
  - internal/launchctl/apple_system_test.go
  - frontend/app/components/__tests__/ServiceRow.test.ts
  - resolve_helper_test.go
-->

---
### Requirement: Clone a user service

The system SHALL provide a clone action on every user service detail view that creates a new user service whose configuration is derived from the source service.

The clone action SHALL be available only for services whose `Type` is `"user"`. The system SHALL NOT expose the clone action on system service or apple-system service detail views.

When the clone action is triggered, the system SHALL open the existing service creation form pre-filled with the source service's `Program`, `ProgramArguments`, `WorkingDirectory`, `KeepAlive` (including its advanced sub-keys), `ThrottleInterval`, `EnvironmentVariables`, `Schedule`, and `WakeSystem` values. The system SHALL leave the `Label` input empty. The system SHALL set the launch-policy selection so that no `RunAtLoad` is written on submission unless KeepAlive is preserved: when the source's launch policy is `Keep Alive`, the clone SHALL preserve `Keep Alive`; otherwise the clone SHALL default to `On Demand` regardless of the source's `RunAtLoad` value.

The system SHALL require the user to provide a new `Label` before submission. The system SHALL submit the cloned configuration through the existing user-service creation path (`CreateService`), so log paths are re-composed from the new label and the user's configured log directory.

When the submitted label conflicts with an existing user service, the system SHALL surface the backend's `service <label> already exists` error inline in the creation form and SHALL NOT close the form, SHALL NOT reset the user-entered fields, and SHALL NOT create any file.

When the clone succeeds, the system SHALL navigate to the new service's detail view at `/services/<new-label>?type=user`.

#### Scenario: Clone action visibility by service type

- **WHEN** the user opens a detail view at `/services/<label>?type=user`
- **THEN** the header action area renders a Copy button next to the existing Start/Stop/Restart/Run Now buttons
- **AND** when the user opens a detail view at `/services/<label>?type=system` or `/services/<label>?type=apple-system`, no Copy button is rendered

#### Scenario: Pre-filled creation form on clone

- **WHEN** the user clicks the Copy button on a user service whose configuration is fully populated
- **THEN** the creation form opens with all of `Program`, `ProgramArguments`, `WorkingDirectory`, `KeepAlive`, `ThrottleInterval`, `EnvironmentVariables`, `Schedule`, `WakeSystem` set to the source service's values
- **AND** the `Label` input is empty
- **AND** the launch-policy radio is not set to `Run at Load`: a source with `Keep Alive` stays on `Keep Alive`, and any other source defaults to `On Demand`

##### Example: Cloning `com.example.ticker`

- **GIVEN** source service `com.example.ticker` has Program=`/usr/bin/foo`, ProgramArguments=`["--port=8080"]`, EnvironmentVariables=`{LOG_LEVEL: "debug"}`, launch policy `Keep Alive` (KeepAlive boolean), Schedule=`StartInterval(60)`
- **WHEN** the user clicks Copy
- **THEN** the form opens with Program=`/usr/bin/foo`, Arguments text=`--port=8080`, EnvironmentVariables row `LOG_LEVEL=debug`, the launch-policy radio on `Keep Alive`, Schedule=`StartInterval(60)`
- **AND** the Label input is empty

#### Scenario: Successful clone creates new service and navigates

- **GIVEN** the user has the creation form open with a prefilled clone of `com.example.ticker`
- **WHEN** the user enters `com.example.ticker-staging` as the label and submits
- **THEN** the system writes `~/Library/LaunchAgents/com.example.ticker-staging.plist` containing the cloned configuration
- **AND** because the clone's launch policy is `Keep Alive`, the written plist contains a `KeepAlive` key and does NOT contain a standalone `RunAtLoad` key (launchd implies it)
- **AND** the browser navigates to `/services/com.example.ticker-staging?type=user`
- **AND** the source service `com.example.ticker` and its plist file remain unchanged

#### Scenario: Duplicate label is rejected inline

- **GIVEN** a user service `com.example.ticker-staging` already exists
- **WHEN** the user submits a clone with the same label `com.example.ticker-staging`
- **THEN** the form remains open with all entered fields preserved
- **AND** an inline error message `service com.example.ticker-staging already exists` is shown
- **AND** no plist file is created or modified
- **AND** no navigation occurs

#### Scenario: User selects Run at Load before submitting

- **GIVEN** the user has the creation form open with a prefilled clone defaulting to `On Demand`
- **WHEN** the user selects the `Run at Load` launch-policy radio and submits with a new label
- **THEN** the resulting plist contains `RunAtLoad = true`
- **AND** the default `On Demand` selection is only the initial state, not a submission-time constraint


<!-- @trace
source: advanced-keepalive-options
updated: 2026-06-01
code:
  - frontend/app/types/wails.d.ts
  - frontend/app/components/LaunchPolicyForm.vue
  - frontend/app/pages/services/[name].vue
  - internal/launchctl/keepalive.go
  - frontend/app/components/ServiceSummary.vue
  - frontend/app/components/ServiceRow.vue
  - internal/launchctl/types.go
  - frontend/app/components/CreateServiceModal.vue
  - frontend/app/utils/serviceToConfig.ts
  - internal/launchctl/plist_encode.go
  - internal/launchctl/user.go
  - internal/launchctl/readonly.go
  - frontend/app/utils/launchPolicy.ts
  - frontend/wailsjs/go/models.ts
tests:
  - frontend/app/components/__tests__/CreateServiceModal.test.ts
  - frontend/app/utils/__tests__/serviceToConfig.test.ts
  - internal/launchctl/plist_encode_test.go
  - frontend/app/components/__tests__/LaunchPolicyForm.test.ts
  - internal/launchctl/keepalive_test.go
  - frontend/app/pages/services/__tests__/edit-launch-policy.test.ts
  - internal/launchctl/user_test.go
  - frontend/app/components/__tests__/CloneUserService.test.ts
  - frontend/app/components/__tests__/ServiceSummary.test.ts
  - frontend/app/components/__tests__/ServiceRow.test.ts
  - frontend/app/utils/__tests__/launchPolicy.test.ts
-->

---
### Requirement: Launch policy selection in service creation form

The service creation form and the service edit form SHALL present launch behavior as a single mutually-exclusive radio group named "Launch Policy" with exactly three options: `On Demand`, `Run at Load`, and `Keep Alive`. The forms SHALL NOT present `Run at Load` and `Keep Alive` as two independent checkboxes.
On submission, the system SHALL map the selected option as follows: `On Demand` writes neither `RunAtLoad` nor `KeepAlive`; `Run at Load` writes `RunAtLoad = true` and no `KeepAlive`; `Keep Alive` writes a `KeepAlive` value and SHALL NOT additionally write `RunAtLoad`, because launchd implies `RunAtLoad` from `KeepAlive`.
When loading an existing service into either form, the system SHALL map the plist back to a radio selection by KeepAlive precedence: when the parsed `KeepAlive` is enabled the selection SHALL be `Keep Alive` regardless of `RunAtLoad`; otherwise when `RunAtLoad` is true the selection SHALL be `Run at Load`; otherwise the selection SHALL be `On Demand`. A legacy service carrying both `RunAtLoad = true` and a `KeepAlive` value SHALL therefore load as `Keep Alive`, and on the next save SHALL NOT emit a standalone `RunAtLoad` key.
All launch-policy labels and helper text SHALL be in English.

#### Scenario: Three mutually-exclusive options

- **WHEN** the user opens the service creation form
- **THEN** a "Launch Policy" radio group is rendered with the options `On Demand`, `Run at Load`, and `Keep Alive`, exactly one of which is selectable at a time

#### Scenario: Launch policy maps to plist keys

- **WHEN** the user selects a launch policy and submits an otherwise minimal config
- **THEN** the written plist contains `RunAtLoad` and `KeepAlive` keys exactly as specified in the table below

##### Example: launch policy to plist keys

| Selected policy | RunAtLoad written | KeepAlive written |
| --------------- | ----------------- | ----------------- |
| On Demand       | no                | no                |
| Run at Load     | yes (`true`)      | no                |
| Keep Alive      | no                | yes               |

#### Scenario: Legacy plist with both RunAtLoad and KeepAlive loads as Keep Alive

- **GIVEN** an existing service whose plist contains both `RunAtLoad = true` and `KeepAlive = {SuccessfulExit: false}`
- **WHEN** the user opens it in the edit form
- **THEN** the launch-policy radio is set to `Keep Alive`
- **AND** when the user saves without further changes, the written plist contains the `KeepAlive` dictionary and does NOT contain a standalone `RunAtLoad` key


<!-- @trace
source: advanced-keepalive-options
updated: 2026-06-01
code:
  - frontend/app/types/wails.d.ts
  - frontend/app/components/LaunchPolicyForm.vue
  - frontend/app/pages/services/[name].vue
  - internal/launchctl/keepalive.go
  - frontend/app/components/ServiceSummary.vue
  - frontend/app/components/ServiceRow.vue
  - internal/launchctl/types.go
  - frontend/app/components/CreateServiceModal.vue
  - frontend/app/utils/serviceToConfig.ts
  - internal/launchctl/plist_encode.go
  - internal/launchctl/user.go
  - internal/launchctl/readonly.go
  - frontend/app/utils/launchPolicy.ts
  - frontend/wailsjs/go/models.ts
tests:
  - frontend/app/components/__tests__/CreateServiceModal.test.ts
  - frontend/app/utils/__tests__/serviceToConfig.test.ts
  - internal/launchctl/plist_encode_test.go
  - frontend/app/components/__tests__/LaunchPolicyForm.test.ts
  - internal/launchctl/keepalive_test.go
  - frontend/app/pages/services/__tests__/edit-launch-policy.test.ts
  - internal/launchctl/user_test.go
  - frontend/app/components/__tests__/CloneUserService.test.ts
  - frontend/app/components/__tests__/ServiceSummary.test.ts
  - frontend/app/components/__tests__/ServiceRow.test.ts
  - frontend/app/utils/__tests__/launchPolicy.test.ts
-->

---
### Requirement: KeepAlive advanced options

When the launch-policy selection is `Keep Alive`, the service creation form SHALL reveal an advanced options section. The section SHALL allow the user to choose between a boolean `Keep Alive` and a dictionary form, and SHALL provide editable controls for `SuccessfulExit`, `Crashed`, and `AfterInitialDemand`. The form SHALL provide an editable integer control for `ThrottleInterval` in the same advanced section.
The advanced section SHALL display informational text stating that multiple dictionary conditions are combined with OR semantics and that `Keep Alive` implies `Run at Load`.
The form SHALL NOT render editable controls for `NetworkState`, `PathState`, or `OtherJobEnabled`; the system SHALL preserve any such values read from an existing plist and write them back unchanged.
When the dictionary form is selected but no effective sub-key is set (no editable boolean set and no preserved `NetworkState`/`PathState`/`OtherJobEnabled`), the system SHALL write `KeepAlive = true` (boolean form) rather than an empty dictionary. The system SHALL NOT write an empty `KeepAlive` dictionary.
When the launch-policy selection is not `Keep Alive`, the advanced options section SHALL be hidden. All advanced-option labels and helper text SHALL be in English.

#### Scenario: Advanced section visibility follows launch policy

- **WHEN** the user selects the `Keep Alive` launch policy
- **THEN** the advanced KeepAlive options section is shown
- **AND** when the user selects `On Demand` or `Run at Load`, the advanced section is hidden

#### Scenario: Editing dictionary sub-keys produces a dictionary KeepAlive

- **GIVEN** the user has selected `Keep Alive` and switched to the dictionary form
- **WHEN** the user sets `SuccessfulExit` to false and submits
- **THEN** the written plist contains `KeepAlive` as a dictionary with `SuccessfulExit` set to `false`

#### Scenario: ThrottleInterval edited in advanced section

- **WHEN** the user enters `15` in the ThrottleInterval control and submits
- **THEN** the written plist contains `ThrottleInterval` with value 15

#### Scenario: Non-editable sub-keys are preserved through edit

- **GIVEN** an existing service whose `KeepAlive` dictionary contains `PathState = {"/tmp/flag": true}`
- **WHEN** the user opens it for editing, changes only `SuccessfulExit`, and submits
- **THEN** the written plist's `KeepAlive` dictionary still contains `PathState = {"/tmp/flag": true}`

#### Scenario: Dictionary form with no effective sub-key downgrades to boolean

- **GIVEN** the user has selected `Keep Alive` and the dictionary form
- **WHEN** the user clears every editable sub-key (and no preserved `NetworkState`/`PathState`/`OtherJobEnabled` remains) and submits
- **THEN** the written plist contains `KeepAlive` with the boolean value `true` and does NOT contain an empty `KeepAlive` dictionary

<!-- @trace
source: advanced-keepalive-options
updated: 2026-06-01
code:
  - frontend/app/types/wails.d.ts
  - frontend/app/components/LaunchPolicyForm.vue
  - frontend/app/pages/services/[name].vue
  - internal/launchctl/keepalive.go
  - frontend/app/components/ServiceSummary.vue
  - frontend/app/components/ServiceRow.vue
  - internal/launchctl/types.go
  - frontend/app/components/CreateServiceModal.vue
  - frontend/app/utils/serviceToConfig.ts
  - internal/launchctl/plist_encode.go
  - internal/launchctl/user.go
  - internal/launchctl/readonly.go
  - frontend/app/utils/launchPolicy.ts
  - frontend/wailsjs/go/models.ts
tests:
  - frontend/app/components/__tests__/CreateServiceModal.test.ts
  - frontend/app/utils/__tests__/serviceToConfig.test.ts
  - internal/launchctl/plist_encode_test.go
  - frontend/app/components/__tests__/LaunchPolicyForm.test.ts
  - internal/launchctl/keepalive_test.go
  - frontend/app/pages/services/__tests__/edit-launch-policy.test.ts
  - internal/launchctl/user_test.go
  - frontend/app/components/__tests__/CloneUserService.test.ts
  - frontend/app/components/__tests__/ServiceSummary.test.ts
  - frontend/app/components/__tests__/ServiceRow.test.ts
  - frontend/app/utils/__tests__/launchPolicy.test.ts
-->

---
### Requirement: Routing name path-traversal confinement

The user, system, and read-only service managers SHALL reject any routing `name` (and `CreateService`'s `config.Label`) that is not a single path component before it is joined into a plist path. A value containing a path separator or a NUL byte, or equal to `.` or `..`, SHALL be rejected with a validation error and SHALL NOT reach any file operation. A single component that merely contains `..` as a substring (e.g. `com.example..worker`) is NOT rejected: with no path separator it cannot traverse out of the base directory, and it is a legal launchd label that must remain manageable. This confines all name-derived operations (get, read plist, read logs, create, update, delete, clear logs, and the system-domain start/stop/restart) to the intended base directory (`~/Library/LaunchAgents` for user services, `/Library/LaunchDaemons` for system daemons, or the read-only system directories), since `filepath.Join` alone does not confine `..` to the base directory. This is a GUI-side defense in depth; for system-domain writes the privileged helper independently re-validates the path, but the manager SHALL NOT rely on that alone.

#### Scenario: Traversal name is rejected (user and system domains)

- **WHEN** a binding is invoked with `name` set to `../../etc/passwd` (or `config.Label` containing `..`/`/`), in either the user domain or the system domain
- **THEN** the manager returns a validation error and performs no file read, write, or delete outside the base directory

#### Scenario: Normal service name is accepted

- **WHEN** `name` is a plain label such as `com.example.foo`
- **THEN** the operation proceeds normally

<!-- @trace
source: filesystem-input-hardening
updated: 2026-07-23
code:
  - internal/launchctl/user.go
  - .github/workflows/build.yml
  - internal/privhelper/client.go
  - README.md
  - frontend/app/composables/useAdminMode.ts
  - internal/privhelper/handlers.go
  - main.go
  - frontend/app/pages/settings.vue
  - internal/privhelper/integrity.go
  - internal/privhelper/server.go
  - internal/privhelper/logpath.go
  - internal/launchctl/readonly.go
  - internal/privhelper/install.go
  - cmd/launchpal-privhelper/procinfo_other.go
  - Makefile
  - cmd/launchpal-privhelper/procinfo_darwin.go
  - internal/launchctl/system.go
  - internal/privhelper/logpath_other.go
  - cmd/launchpal-privhelper/main.go
  - admin_mode.go
  - internal/privhelper/logpath_darwin.go
tests:
  - internal/privhelper/integrity_test.go
  - internal/privhelper/client_test.go
  - internal/privhelper/handlers_test.go
  - cmd/launchpal-privhelper/helper_test.go
  - admin_mode_test.go
  - internal/privhelper/server_test.go
  - resolve_helper_test.go
  - internal/privhelper/install_test.go
  - internal/launchctl/system_test.go
  - internal/launchctl/user_test.go
-->

---
### Requirement: Consistent user service runtime status

For user LaunchAgents, `UserManager.List` and `UserManager.Get` SHALL derive `Status` and `PID` exclusively from launchd information for the service label. Both operations SHALL apply the same classification when the launchd job state remains unchanged. The user-service status path SHALL NOT infer `StatusRunning` or a PID from a process-name or command-line substring scan.

For a non-empty service label, the classification SHALL be:

- launchd reports a positive PID: `StatusRunning` with that PID.
- launchd reports the job as loaded without a positive PID: `StatusLoaded` with PID `0`.
- launchd reports that the label is not loaded: `StatusStopped` with PID `0`.

For an empty service label, both operations SHALL return `StatusUnknown` with PID `0`.

#### Scenario: Running job is consistent between list and detail

- **GIVEN** a user LaunchAgent with a non-empty label for which launchd reports PID `4321`
- **WHEN** the unchanged service is returned by `UserManager.List` and `UserManager.Get`
- **THEN** both results have `StatusRunning` and PID `4321`

#### Scenario: Loaded wrapper command is not attributed an unrelated PID

- **GIVEN** a loaded user LaunchAgent whose program is `open`, launchd reports no PID for its label, and an unrelated process command line contains the substring `open`
- **WHEN** the unchanged service is returned by `UserManager.List` and `UserManager.Get`
- **THEN** both results have `StatusLoaded` and PID `0`
- **THEN** neither result uses the unrelated process PID

#### Scenario: Unloaded job is consistent between list and detail

- **GIVEN** a user LaunchAgent with a non-empty label that launchd reports as not loaded
- **WHEN** the unchanged service is returned by `UserManager.List` and `UserManager.Get`
- **THEN** both results have `StatusStopped` and PID `0`

#### Scenario: Empty label has an unknown status

- **GIVEN** a user LaunchAgent plist whose `Label` value is empty
- **WHEN** the service is returned by `UserManager.List` and `UserManager.Get`
- **THEN** both results have `StatusUnknown` and PID `0`

##### Example: launchd state classification

| Label | launchd state | launchd PID | Expected status | Expected PID |
| ----- | ------------- | ----------- | --------------- | ------------ |
| `com.example.running` | loaded and running | `4321` | `running` | `4321` |
| `com.example.loaded` | loaded, not running | none | `loaded` | `0` |
| `com.example.stopped` | not loaded | none | `stopped` | `0` |
| empty | not queried | none | `unknown` | `0` |

<!-- @trace
source: fix-user-service-status-consistency
updated: 2026-08-25
code:
  - go.mod
  - frontend/pnpm-workspace.yaml
  - internal/launchctl/user.go
  - .github/workflows/build.yml
  - go.sum
  - frontend/package.json
  - frontend/package.json.md5
tests:
  - internal/launchctl/user_test.go
-->

---
### Requirement: Display log file metadata in Logs tab

The Logs tab in the service detail view SHALL display a persistent info row directly below the stdout/stderr toggle, showing for the currently selected log type: the resolved log file path and the human-readable file size.

The info row SHALL render the path with middle truncation when its rendered width exceeds the available space, while preserving the file basename suffix; the full unabbreviated path SHALL be available via a hover tooltip on the path element.

The info row SHALL format the size using base-1024 units (`B`, `KB`, `MB`, `GB`) with one decimal place for `KB`/`MB`/`GB` and integer precision for `B`. When the log file does not exist (`exists: false`) the size SHALL render as the literal string `—` (em dash).

When the resolved log path is empty (no `StandardOutPath` or `StandardErrorPath` configured for the selected log type), the info row SHALL display the literal string `No stdout path configured` or `No stderr path configured` and SHALL omit the size field.

The info row SHALL be visible for all three service categories: user services, system services, and Apple system services.

The info row SHALL refresh whenever the log content refreshes — on component mount, on stdout/stderr toggle, when the user clicks the existing Refresh control, and on each Auto-refresh poll tick while the Auto-refresh toggle is enabled. The info row SHALL NOT introduce a timer, file-system watcher, or any other active polling mechanism of its own; it SHALL update only alongside a log-content refresh, so the displayed path and size always describe the same load as the content shown.

#### Scenario: User service with existing stdout log

- **WHEN** the Logs tab is opened for a user service whose stdout log exists at `/Users/alice/Library/Logs/com.example.stdout.log` with size 2516582 bytes
- **THEN** the info row reads `stdout · /Users/alice/Library/Logs/com.example.stdout.log · 2.4 MB` (path may be middle-truncated to fit the available width; full path appears in the tooltip)

#### Scenario: Apple system service info row visible

- **WHEN** the Logs tab is opened for an Apple system service (read-only)
- **THEN** the info row is rendered with the resolved path and size, identical in layout to user and system services

#### Scenario: Toggle between stdout and stderr

- **WHEN** the user clicks the stderr toggle while viewing a service whose stderr path differs from its stdout path
- **THEN** both the log content and the info row update to reflect the stderr path and stderr file size

#### Scenario: No stdout path configured

- **WHEN** the Logs tab is opened for a service whose plist does not configure `StandardOutPath`
- **THEN** the info row displays `No stdout path configured` and the size segment is not rendered

#### Scenario: Log path configured but file not yet created

- **WHEN** the Logs tab is opened for a service whose log path is configured but the file has not yet been created
- **THEN** the info row displays the resolved path and the size segment renders as `—`

##### Example: size formatting boundary cases

| size (bytes) | rendered |
| ------------ | -------- |
| 0            | 0 B      |
| 512          | 512 B    |
| 1024         | 1.0 KB   |
| 2516582      | 2.4 MB   |
| 1181116006   | 1.1 GB   |

#### Scenario: No active polling of its own between refreshes

- **WHEN** the Logs tab remains open for several seconds without user interaction, Auto-refresh is disabled, and the log file grows on disk
- **THEN** the info row continues to display the size measured at the most recent refresh; the size only updates when the user clicks Refresh, toggles stdout/stderr, or remounts the component

#### Scenario: Auto-refresh keeps the info row in step with the content

- **WHEN** the user enables the Auto-refresh toggle for a service whose log file is growing on disk
- **THEN** each poll tick refreshes both the log content and the info row, so the displayed size advances with the content rather than remaining pinned to the size measured at the first load

<!-- @trace
source: add-log-file-info-bar
updated: 2026-08-25
code:
  - internal/launchctl/types.go
  - frontend/app/components/LogStorageSection.vue
  - .agents/skills/spectra-ingest/SKILL.md
  - admin_mode.go
  - frontend/app/components/ReadOnlyServiceList.vue
  - frontend/app/components/InlineBanner.vue
  - frontend/app/pages/settings.vue
  - .agents/skills/spectra-ask/SKILL.md
  - internal/launchctl/user.go
  - go.sum
  - internal/launchctl/nofollow_other.go
  - frontend/app/utils/formatters.ts
  - internal/privhelper/handlers.go
  - internal/launchctl/system.go
  - frontend/package.json
  - .agents/skills/spectra-audit/SKILL.md
  - internal/launchctl/apple_system.go
  - frontend/app/pages/system.vue
  - .github/workflows/release-please.yml
  - frontend/wailsjs/go/main/App.js
  - .agents/skills/spectra-apply/SKILL.md
  - internal/privhelper/protocol.go
  - .github/workflows/build.yml
  - internal/privhelper/install.go
  - app.go
  - internal/privhelper/server.go
  - frontend/wailsjs/go/models.ts
  - internal/launchctl/manager.go
  - .agents/skills/spectra-propose/SKILL.md
  - AGENTS.md
  - .agents/skills/spectra-archive/SKILL.md
  - internal/privhelper/logpath_darwin.go
  - .agents/skills/spectra-debug/SKILL.md
  - frontend/app/utils/launchPolicy.ts
  - frontend/app/utils/ansiToHtml.ts
  - frontend/vitest.setup.ts
  - README.md
  - frontend/wailsjs/go/main/App.d.ts
  - frontend/app/utils/logPaths.ts
  - CHANGELOG.md
  - internal/launchctl/nofollow_darwin.go
  - cmd/launchpal-privhelper/main.go
  - frontend/app/components/LaunchPolicyForm.vue
  - cmd/launchpal-privhelper/procinfo_other.go
  - internal/launchctl/plist_encode.go
  - frontend/app/composables/useAdminMode.ts
  - internal/privhelper/logpath_other.go
  - internal/privhelper/client.go
  - Makefile
  - frontend/app/utils/serviceToConfig.ts
  - main.go
  - frontend/app/components/ServiceSummary.vue
  - .agents/skills/spectra-commit/SKILL.md
  - frontend/pnpm-workspace.yaml
  - frontend/app/utils/serviceValidation.ts
  - cmd/launchpal-privhelper/procinfo_darwin.go
  - frontend/app/pages/services/[name].vue
  - internal/launchctl/readonly.go
  - frontend/app/composables/useSettings.ts
  - frontend/app/components/ServiceLogs.vue
  - frontend/package.json.md5
  - .agents/skills/spectra-drift/SKILL.md
  - internal/settings/settings.go
  - go.mod
  - internal/privhelper/integrity.go
  - frontend/app/components/CreateServiceModal.vue
  - internal/launchctl/keepalive.go
  - frontend/app/components/ServiceRow.vue
  - frontend/app/types/wails.d.ts
  - frontend/app/utils/settingsValidation.ts
  - .agents/skills/spectra-discuss/SKILL.md
  - internal/privhelper/logpath.go
tests:
  - internal/privhelper/integrity_test.go
  - internal/launchctl/system_test.go
  - internal/privhelper/server_test.go
  - cmd/launchpal-privhelper/helper_test.go
  - internal/privhelper/protocol_test.go
  - frontend/app/pages/services/__tests__/edit-program-arguments-validation.test.ts
  - frontend/app/components/__tests__/ServiceSummary.test.ts
  - internal/launchctl/types_test.go
  - frontend/app/components/__tests__/CloneUserService.test.ts
  - frontend/app/utils/__tests__/launchPolicy.test.ts
  - frontend/app/components/__tests__/ServiceLogs.test.ts
  - frontend/app/pages/services/__tests__/edit-launch-policy.test.ts
  - internal/privhelper/handlers_test.go
  - admin_mode_test.go
  - frontend/app/pages/__tests__/settings.test.ts
  - frontend/app/components/__tests__/LaunchPolicyForm.test.ts
  - frontend/app/utils/__tests__/ansiToHtml.test.ts
  - frontend/app/utils/__tests__/serviceToConfig.test.ts
  - internal/launchctl/keepalive_test.go
  - frontend/app/components/__tests__/CreateServiceModal.test.ts
  - app_test.go
  - frontend/app/utils/__tests__/formatters.test.ts
  - internal/launchctl/plist_encode_test.go
  - internal/privhelper/install_test.go
  - internal/settings/settings_test.go
  - frontend/app/composables/__tests__/useSettings.test.ts
  - internal/launchctl/user_test.go
  - internal/privhelper/client_test.go
  - internal/launchctl/apple_system_test.go
  - frontend/app/components/__tests__/ServiceRow.test.ts
  - resolve_helper_test.go
-->