## ADDED Requirements

### Requirement: Three manager types with distinct access levels

The system SHALL provide three Manager implementations:
- `UserManager` for `~/Library/LaunchAgents` with full read-write access
- `SystemManager` for `/Library/LaunchDaemons` with read-only access
- `AppleSystemManager` for `/System/Library/LaunchDaemons` with read-only access

All three SHALL implement the same `Manager` interface.

#### Scenario: Manager interface compliance

- **WHEN** SystemManager and AppleSystemManager are instantiated
- **THEN** both satisfy the Manager interface at compile time (verified by `var _ Manager = (*SystemManager)(nil)`)

### Requirement: Read-only managers reject write operations

SystemManager and AppleSystemManager SHALL return `ErrReadOnlyManager` for all write operations: Start, Stop, Restart, Create, Update, Delete.

#### Scenario: SystemManager write operations

- **WHEN** Start, Stop, Restart, Create, Update, or Delete is called on SystemManager
- **THEN** each call returns `ErrReadOnlyManager`

#### Scenario: AppleSystemManager write operations

- **WHEN** Start, Stop, Restart, Create, Update, or Delete is called on AppleSystemManager
- **THEN** each call returns `ErrReadOnlyManager`

### Requirement: List system services

SystemManager SHALL list all `.plist` files in `/Library/LaunchDaemons`.
AppleSystemManager SHALL list all `.plist` files in `/System/Library/LaunchDaemons`.
Both SHALL return an empty list (not an error) when the directory does not exist or is not readable due to permissions.
Both SHALL skip services whose plist cannot be read or parsed.

#### Scenario: List with permission denied

- **WHEN** the manager's directory is not readable
- **THEN** the system returns an empty list with no error

#### Scenario: List with mixed readable and unreadable plists

- **WHEN** the directory contains 5 plist files but 2 cannot be read
- **THEN** the system returns 3 services (the readable ones)

### Requirement: Get system service details

SystemManager.Get SHALL read and parse a plist from `/Library/LaunchDaemons/<name>.plist`.
AppleSystemManager.Get SHALL read and parse a plist from `/System/Library/LaunchDaemons/<name>.plist`.
Both SHALL set the `Type` field to their respective type (`"system"` or `"apple-system"`).
Both SHALL set `ReadOnly` to `true`.
Both SHALL detect plist format (xml or binary) and populate `PlistFormat`.
Both SHALL populate `Status`, `PID`, and `StatusConfidence` according to the `system-daemon-status-detection` capability.
Both SHALL return a permission denied error when the file is not readable.

#### Scenario: Get a system service

- **WHEN** Get is called with a valid service name in `/Library/LaunchDaemons`
- **THEN** the returned Service has Type=`"system"`, ReadOnly=`true`, all parsed plist fields populated, and Status/PID/StatusConfidence populated per heuristic detection

#### Scenario: Get an Apple system service

- **WHEN** Get is called with a valid service name in `/System/Library/LaunchDaemons`
- **THEN** the returned Service has Type=`"apple-system"`, ReadOnly=`true`, all parsed plist fields populated, and Status/PID/StatusConfidence populated per heuristic detection

#### Scenario: Permission denied on Get

- **WHEN** Get is called for a plist file the user cannot read
- **THEN** the system returns an error containing "permission denied"


<!-- @trace
source: system-daemon-status-detection
updated: 2026-04-21
code:
  - internal/launchctl/user.go
  - README.md
  - internal/launchctl/readonly.go
  - frontend/app/pages/services/[name].vue
  - frontend/wailsjs/go/models.ts
  - internal/launchctl/status_detect.go
  - frontend/app/types/wails.d.ts
  - internal/launchctl/types.go
  - frontend/app/components/ServiceRow.vue
  - frontend/app/components/StatusConfidenceIcon.vue
tests:
  - internal/launchctl/readonly_test.go
  - internal/launchctl/status_detect_test.go
-->

### Requirement: Get raw plist content with format conversion

SystemManager.GetPlist and AppleSystemManager.GetPlist SHALL use `plutil -convert xml1 -o -` to convert the plist to XML format before returning.
This ensures binary plists are returned as readable XML.

#### Scenario: Get binary plist content

- **WHEN** GetPlist is called for a service with a binary plist
- **THEN** the returned content is valid XML plist (converted by plutil)

#### Scenario: Get XML plist content

- **WHEN** GetPlist is called for a service with an XML plist
- **THEN** the returned content is the XML plist (passed through plutil unchanged)

### Requirement: Read system service logs

SystemManager.GetLogs and AppleSystemManager.GetLogs SHALL read log files based on the service's configured StandardOutPath or StandardErrorPath.
Both SHALL return an error for invalid log type (not "stdout" or "stderr").
Both SHALL return an error when no log path is configured.
Both SHALL return specific errors for missing files and permission denied.

#### Scenario: Read log with no path configured

- **WHEN** GetLogs is called for a system service with no StandardOutPath
- **THEN** the system returns an error indicating no log path is configured

#### Scenario: Log file not found

- **WHEN** GetLogs is called and the configured log file does not exist
- **THEN** the system returns an error indicating "log file not found"

#### Scenario: Log file permission denied

- **WHEN** GetLogs is called and the configured log file is not readable
- **THEN** the system returns an error indicating "permission denied"

### Requirement: Service type and read-only fields

The Service struct SHALL include a `Type` field with values `"user"`, `"system"`, or `"apple-system"`.
The Service struct SHALL include a `ReadOnly` field set to `false` for user services and `true` for system and apple-system services.

#### Scenario: User service fields

- **WHEN** a service is loaded from UserManager
- **THEN** Type is `"user"` and ReadOnly is `false`

#### Scenario: System service fields

- **WHEN** a service is loaded from SystemManager
- **THEN** Type is `"system"` and ReadOnly is `true`

#### Scenario: Apple system service fields

- **WHEN** a service is loaded from AppleSystemManager
- **THEN** Type is `"apple-system"` and ReadOnly is `true`

## Requirements
