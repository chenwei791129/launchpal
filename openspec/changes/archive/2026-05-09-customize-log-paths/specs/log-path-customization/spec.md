## ADDED Requirements

### Requirement: Settings page exposes log directory controls

The Settings page SHALL render a "Log Storage" section immediately after the existing "Backup Storage" section.
The Log Storage section SHALL contain two independent controls labelled "User Log Directory" and "System Log Directory".
Each control SHALL display the current value sourced from `GetSettings`, an editable text input pre-populated with that value, a Save button, and a Reset to Default button.
The visual layout SHALL match the existing Backup Storage section style (icon, title, supporting paragraph, monospace path field).
All user-facing strings SHALL be in English (no Chinese characters in the rendered UI), in line with the project's UI language convention.

#### Scenario: Initial render of Log Storage section

- **WHEN** the user opens the Settings page on a fresh install with no `~/.launchpal/settings.json` present
- **THEN** the Log Storage section is rendered after the Backup Storage section, the User Log Directory input shows `~/Library/Logs`, and the System Log Directory input shows `/Library/Logs`

#### Scenario: Reset restores the default

- **WHEN** the user has previously saved `userLogDir = "/tmp/userlogs"` and then clicks the User Log Directory's "Reset to Default" button
- **THEN** the User Log Directory input value updates to `~/Library/Logs` and a subsequent Save call writes that default to disk

### Requirement: Save action validates and persists

The Save button next to each log directory input SHALL invoke `UpdateSettings` with the full Settings object.
The page SHALL display a non-blocking inline error beside the failing field when `UpdateSettings` returns a validation error, surfacing the backend error message verbatim.
The page SHALL display a transient success indicator next to the saved field upon a successful save and SHALL refresh its locally cached Settings via `GetSettings`.
The Save button SHALL be disabled while a save is in flight and SHALL re-enable on completion.
The page SHALL NOT issue `UpdateSettings` calls in response to keystrokes; saves SHALL only occur on explicit Save button activation.

#### Scenario: Invalid System Log Directory shows inline error

- **WHEN** the user types `/etc/launchpal/logs` into the System Log Directory input and clicks Save
- **THEN** the inline error beneath the System Log Directory input shows the validation message returned by `UpdateSettings`, the input value remains as the user typed, and the on-disk settings file is unchanged

#### Scenario: Successful save updates cached settings

- **WHEN** the user types `/Library/Logs/launchpal` into the System Log Directory input and clicks Save
- **THEN** `UpdateSettings` is invoked exactly once with the new value, a success indicator is shown, and a subsequent open of the New Service modal sees the new default

### Requirement: New Service modal sources defaults from settings

The CreateServiceModal SHALL load the current Settings via `GetSettings` (or a shared `useSettings` composable) when it mounts.
The modal SHALL compute its preview log paths as `<userLogDir>/<label>/stdout.log` and `<userLogDir>/<label>/stderr.log` for `serviceType === "user"`.
The modal SHALL compute its preview log paths as `<systemLogDir>/<label>/stdout.log` and `<systemLogDir>/<label>/stderr.log` for `serviceType === "system"`.
The modal SHALL submit the resolved paths as `stdoutPath` and `stderrPath` on the ServiceConfig sent to `CreateService` / `CreateSystemService`.
The modal SHALL NOT expose a UI control for editing the directory portion per service in this change; only the auto-generated path is shown (read-only preview, as today).
The modal SHALL re-read settings each time it is opened, so that changes saved on the Settings page take effect on the next New Service interaction without requiring an application restart.

##### Example: path composition

| serviceType | settings.userLogDir | settings.systemLogDir | form.label    | resulting stdoutPath                       |
| ----------- | ------------------- | --------------------- | ------------- | ------------------------------------------ |
| user        | `~/Library/Logs`    | `/Library/Logs`       | `com.user.x`  | `~/Library/Logs/com.user.x/stdout.log`     |
| user        | `/tmp/u`            | `/Library/Logs`       | `com.user.x`  | `/tmp/u/com.user.x/stdout.log`             |
| system      | `~/Library/Logs`    | `/Library/Logs`       | `com.sys.x`   | `/Library/Logs/com.sys.x/stdout.log`       |
| system      | `~/Library/Logs`    | `/var/log/lp`         | `com.sys.x`   | `/var/log/lp/com.sys.x/stdout.log`         |

#### Scenario: Modal reflects updated settings without restart

- **WHEN** the user changes `userLogDir` to `/tmp/userlogs` via Settings page, saves it, and then opens the New Service modal for a user service with label `com.user.demo`
- **THEN** the modal's stdout preview shows `/tmp/userlogs/com.user.demo/stdout.log` and the stderr preview shows `/tmp/userlogs/com.user.demo/stderr.log`

#### Scenario: System service uses systemLogDir

- **WHEN** the user opens the New Service modal in system mode with the default settings and types label `com.sys.demo`
- **THEN** the stdout preview shows `/Library/Logs/com.sys.demo/stdout.log` and the stderr preview shows `/Library/Logs/com.sys.demo/stderr.log`

### Requirement: Existing services are not migrated

The system SHALL NOT scan, edit, or rewrite the `StandardOutPath` or `StandardErrorPath` values of any pre-existing plist when Settings change.
The system SHALL NOT issue any backup write or privileged-helper RPC as a side effect of saving Settings.

#### Scenario: Saving settings does not alter existing plists

- **WHEN** a user has 5 existing user services and 2 existing system services and the user changes both log directory settings
- **THEN** the on-disk content of every pre-existing plist file is byte-identical before and after the Settings save, and no privileged-helper invocation occurs

### Requirement: Helper allowlist drift is prevented

The system SHALL define the system-domain log path allowlist (`/var/log/`, `/private/var/log/`, `/Library/Logs/`, `/tmp/`, `/private/tmp/`) in exactly one Go package.
The settings validator and the privileged-helper handler for `EnsureLogAccess` SHALL both consume this single shared constant; neither SHALL declare its own copy.
The exported symbol SHALL be referenced by name (not duplicated as a literal slice) in both call sites.

#### Scenario: Adding a prefix to the allowlist

- **WHEN** a developer adds a new prefix entry to the shared allowlist constant
- **THEN** both the settings validator and the privileged-helper handler accept the new prefix without any other source-code change in either of those two call sites
