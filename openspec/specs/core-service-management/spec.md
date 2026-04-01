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

### Requirement: CRUD operations for user services

The system SHALL create a new service by writing a plist to `~/Library/LaunchAgents/<label>.plist`.
The system SHALL reject creation when the label is empty.
The system SHALL reject creation when a service with the same label already exists.
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
The system SHALL return an error for invalid log type (not "stdout" or "stderr").
The system SHALL return an error when no log path is configured for the requested type.
The system SHALL return an error when the log file does not exist.

#### Scenario: Read stdout log

- **WHEN** GetLogs is called with logType="stdout" for a service with StandardOutPath configured
- **THEN** the content of the stdout log file is returned

#### Scenario: No log path configured

- **WHEN** GetLogs is called with logType="stdout" for a service with no StandardOutPath
- **THEN** the system returns an error indicating no stdout log path is configured

#### Scenario: Invalid log type

- **WHEN** GetLogs is called with logType="debug"
- **THEN** the system returns an error indicating invalid log type

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
