## ADDED Requirements

### Requirement: Clear Logs button in ServiceLogs view

The system SHALL render a "Clear Logs" control in the `ServiceLogs.vue` controls row, adjacent to the existing Refresh button.
The control SHALL act on the currently selected log type only (`stdout` or `stderr`), never on both at once.
The control SHALL be hidden entirely for `apple-system` services.
The control SHALL be visible for `user` and `system` services and SHALL be enabled or disabled per the visibility-and-availability matrix below.

#### Scenario: Render on user service Logs tab

- **WHEN** the Logs tab is opened for a `user` service
- **THEN** the Clear Logs control is rendered, enabled, and operates on the active `logType`

#### Scenario: Hidden on apple-system service

- **WHEN** the Logs tab is opened for an `apple-system` service
- **THEN** the Clear Logs control is not rendered at all (no disabled-state placeholder)

### Requirement: Button availability matrix

The system SHALL compute the Clear Logs control state from `(serviceType, userWritable, adminMode, exists, hasLogPath)` according to the matrix below. `userWritable` and `exists` SHALL be derived from `GetLogClearStatus` for the active service and `logType`. The control SHALL never be enabled when `hasLogPath` is false or `exists` is false, regardless of other inputs.

| serviceType  | hasLogPath | exists | userWritable | adminMode | Control state                                      |
|--------------|------------|--------|--------------|-----------|----------------------------------------------------|
| user         | true       | true   | —            | —         | enabled                                            |
| user         | false      | —      | —            | —         | disabled, tooltip "No log path configured"         |
| user         | true       | false  | —            | —         | disabled, tooltip "Log file does not exist"        |
| system       | true       | true   | true         | —         | enabled                                            |
| system       | true       | true   | false        | enabled   | enabled                                            |
| system       | true       | true   | false        | disabled  | disabled, tooltip "Enable Admin Mode to clear"     |
| system       | false      | —      | —            | —         | disabled, tooltip "No log path configured"         |
| system       | true       | false  | —            | —         | disabled, tooltip "Log file does not exist"        |
| apple-system | —          | —      | —            | —         | not rendered                                       |

#### Scenario: System service writable by user without Admin Mode

- **WHEN** the Logs tab shows a `system` service whose active log file is user-writable and Admin Mode is disabled
- **THEN** the Clear Logs control is enabled

##### Example: Homebrew daemon log

- **GIVEN** a system service with `StandardOutPath = /usr/local/var/log/myapp.out.log` whose file mode is `0664` and group `admin`
- **WHEN** the active user is a member of the `admin` group and Admin Mode is disabled
- **THEN** the Clear Logs control is enabled and clearing succeeds without prompting for Admin Mode

#### Scenario: System service not user-writable, Admin Mode disabled

- **WHEN** the Logs tab shows a `system` service whose active log file is not user-writable and Admin Mode is disabled
- **THEN** the Clear Logs control is disabled and exposes the tooltip "Enable Admin Mode to clear"

#### Scenario: Active log type has no configured path

- **WHEN** the active `logType` corresponds to a missing `StandardOutPath` or `StandardErrorPath` in the plist
- **THEN** the Clear Logs control is disabled and exposes the tooltip "No log path configured"

#### Scenario: Active log file does not yet exist on disk

- **WHEN** the active log path is configured but the file has not been created yet
- **THEN** the Clear Logs control is disabled and exposes the tooltip "Log file does not exist"

### Requirement: Confirmation dialog before clearing

The system SHALL display a Teleport-based confirmation dialog before performing the clear operation.
The dialog SHALL match the visual style of the existing Run Now confirmation: dark surface, rounded corners, Cancel and primary action buttons.
The primary action button SHALL use red coloring to convey destructive intent.
The dialog title SHALL be "Clear Logs".
The dialog body SHALL state which `logType` and which service name will be affected and warn that the file will be truncated to 0 bytes and existing entries cannot be recovered.
The dialog SHALL NOT auto-confirm; the action SHALL only proceed when the user clicks the primary button.

#### Scenario: User cancels confirmation

- **WHEN** the user clicks the Cancel button or clicks the modal backdrop
- **THEN** the dialog closes and the log file is not modified

#### Scenario: User confirms clear

- **WHEN** the user clicks the primary "Clear Logs" button
- **THEN** the appropriate Wails binding is invoked, the dialog closes, and on success the log content is reloaded showing 0 bytes

### Requirement: Status query is lazy and bounded

The system SHALL query `GetLogClearStatus` only when the Logs tab is mounted or when `logType` changes; never as part of `ListServices` or initial service detail load.
While a status query is in flight, the system SHALL render the Clear Logs control as disabled, regardless of the previously cached status.

#### Scenario: Switching log type triggers re-query

- **WHEN** the user toggles `logType` from `stdout` to `stderr`
- **THEN** a new `GetLogClearStatus` call is made for `stderr` and the control is re-evaluated against the matrix

#### Scenario: List view does not query log status

- **WHEN** the services list page (`/`, `/system`, `/apple-system`) loads
- **THEN** no `GetLogClearStatus` calls are made for any service in the list

### Requirement: Successful clear shows feedback and reloads logs

On successful clear, the system SHALL reload the log content via the existing GetLogs / GetSystemLogs path so the displayed buffer reflects the now-empty file.
The system SHALL show a transient success indicator near the control (such as a brief checkmark or status text) lasting at least 1 second.

#### Scenario: Successful clear updates display

- **WHEN** the clear operation returns successfully
- **THEN** the log pane re-renders showing "No logs available for {logType}" or an empty pre block, and a transient success indicator is shown

### Requirement: Clear failure surfaces a recoverable error

On a failed clear, the system SHALL surface the backend error message in an inline error region (red text) within `ServiceLogs.vue` without dismissing the Logs tab or losing the existing log buffer.
When the failure is due to Admin Mode being unexpectedly disabled mid-operation, the error message SHALL include guidance to re-enable Admin Mode.

#### Scenario: Helper not connected mid-clear

- **WHEN** Admin Mode auto-disabled (idle timeout) between status query and confirmation
- **THEN** the failed clear surfaces an error message including the phrase "Enable Admin Mode" and the existing displayed log content remains visible
