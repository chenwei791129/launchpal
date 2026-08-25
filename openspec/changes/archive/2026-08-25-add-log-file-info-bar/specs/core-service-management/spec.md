## MODIFIED Requirements

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

## ADDED Requirements

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

