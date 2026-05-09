# app-settings Specification

## Purpose

LaunchPal persists user-editable preferences as a JSON document at `~/.launchpal/settings.json`. The `internal/settings` package provides `Default()`, `Load()`, `Save()`, and `Validate()` with atomic writes (temp file + fsync + rename), missing-file fallback to in-memory defaults, and corrupt-JSON recovery. Validation rules for `systemLogDir` import the shared allowlist constant from `internal/privhelper`, preventing drift between the user-facing validator and the privileged helper that consumes the path. The Wails bindings `GetSettings()` and `UpdateSettings(s)` expose the API to the frontend; `UpdateSettings` validates before writing, so invalid Settings never reach disk.

## Requirements

### Requirement: Settings file location and format

The system SHALL persist application settings as a JSON file at `~/.launchpal/settings.json`.
The system SHALL encode the file using UTF-8 with two-space indentation and a trailing newline.
The system SHALL place the settings file in the same root directory (`~/.launchpal/`) as the existing backup store.
The system SHALL NOT depend on any third-party JSON or YAML library; it SHALL use the Go standard library `encoding/json` exclusively.

#### Scenario: Settings file written to canonical path

- **WHEN** the application saves settings for the first time
- **THEN** the file `~/.launchpal/settings.json` exists, contains a valid JSON object, ends with a newline, and uses two-space indentation


<!-- @trace
source: customize-log-paths
updated: 2026-05-09
code:
  - frontend/app/components/CreateServiceModal.vue
  - app.go
  - frontend/app/types/wails.d.ts
  - frontend/wailsjs/go/main/App.js
  - frontend/app/composables/useSettings.ts
  - frontend/wailsjs/go/models.ts
  - frontend/app/utils/settingsValidation.ts
  - frontend/wailsjs/go/main/App.d.ts
  - internal/settings/settings.go
  - internal/privhelper/handlers.go
  - frontend/app/pages/settings.vue
  - frontend/app/components/LogStorageSection.vue
  - frontend/app/utils/logPaths.ts
tests:
  - internal/privhelper/handlers_test.go
  - app_test.go
  - frontend/app/composables/__tests__/useSettings.test.ts
  - frontend/app/pages/__tests__/settings.test.ts
  - frontend/app/components/__tests__/CreateServiceModal.test.ts
  - internal/settings/settings_test.go
-->

### Requirement: Default settings values

The system SHALL define built-in default values for every settings field:
- `userLogDir`: `~/Library/Logs`
- `systemLogDir`: `/Library/Logs`

The system SHALL expose a `Default()` constructor that returns a Settings value populated with the defaults above.
The system SHALL NOT write the defaults file at application startup; defaults SHALL only be returned in memory until the user explicitly saves or modifies a value.

#### Scenario: Defaults returned without filesystem side-effects

- **WHEN** `Default()` is invoked on a system where `~/.launchpal/settings.json` does not exist
- **THEN** a Settings value is returned with `userLogDir = "~/Library/Logs"` and `systemLogDir = "/Library/Logs"`, and no file is created on disk


<!-- @trace
source: customize-log-paths
updated: 2026-05-09
code:
  - frontend/app/components/CreateServiceModal.vue
  - app.go
  - frontend/app/types/wails.d.ts
  - frontend/wailsjs/go/main/App.js
  - frontend/app/composables/useSettings.ts
  - frontend/wailsjs/go/models.ts
  - frontend/app/utils/settingsValidation.ts
  - frontend/wailsjs/go/main/App.d.ts
  - internal/settings/settings.go
  - internal/privhelper/handlers.go
  - frontend/app/pages/settings.vue
  - frontend/app/components/LogStorageSection.vue
  - frontend/app/utils/logPaths.ts
tests:
  - internal/privhelper/handlers_test.go
  - app_test.go
  - frontend/app/composables/__tests__/useSettings.test.ts
  - frontend/app/pages/__tests__/settings.test.ts
  - frontend/app/components/__tests__/CreateServiceModal.test.ts
  - internal/settings/settings_test.go
-->

### Requirement: Load settings from disk

The system SHALL provide a `Load()` operation that reads `~/.launchpal/settings.json` and returns a Settings value.
The system SHALL return the result of `Default()` with no error when the file does not exist.
The system SHALL return the result of `Default()` with no error when the file exists but contains invalid JSON; the implementation SHALL log a warning describing the parse error so the failure is observable in application logs.
The system SHALL fill any field that is absent from the JSON with its default value (forward-compatible decoding).
The system SHALL return an error only for filesystem errors other than "file does not exist" (for example, permission denied on the parent directory).

#### Scenario: Missing file returns defaults

- **WHEN** `Load()` is invoked and `~/.launchpal/settings.json` does not exist
- **THEN** the returned Settings equals `Default()` and no error is returned

#### Scenario: Corrupt JSON falls back to defaults

- **WHEN** `Load()` is invoked and `~/.launchpal/settings.json` contains the bytes `{not json`
- **THEN** the returned Settings equals `Default()`, no error is returned, and a warning is written to the application log

#### Scenario: Partial JSON merges with defaults

- **WHEN** `Load()` is invoked and `~/.launchpal/settings.json` contains `{"userLogDir": "/tmp/mylogs"}`
- **THEN** the returned Settings has `userLogDir = "/tmp/mylogs"` and `systemLogDir = "/Library/Logs"`


<!-- @trace
source: customize-log-paths
updated: 2026-05-09
code:
  - frontend/app/components/CreateServiceModal.vue
  - app.go
  - frontend/app/types/wails.d.ts
  - frontend/wailsjs/go/main/App.js
  - frontend/app/composables/useSettings.ts
  - frontend/wailsjs/go/models.ts
  - frontend/app/utils/settingsValidation.ts
  - frontend/wailsjs/go/main/App.d.ts
  - internal/settings/settings.go
  - internal/privhelper/handlers.go
  - frontend/app/pages/settings.vue
  - frontend/app/components/LogStorageSection.vue
  - frontend/app/utils/logPaths.ts
tests:
  - internal/privhelper/handlers_test.go
  - app_test.go
  - frontend/app/composables/__tests__/useSettings.test.ts
  - frontend/app/pages/__tests__/settings.test.ts
  - frontend/app/components/__tests__/CreateServiceModal.test.ts
  - internal/settings/settings_test.go
-->

### Requirement: Atomic save to disk

The system SHALL provide a `Save(s Settings)` operation that persists the given Settings to `~/.launchpal/settings.json`.
The system SHALL ensure the parent directory `~/.launchpal/` exists with mode `0755` before writing, creating it if necessary.
The system SHALL write atomically: the implementation SHALL marshal the Settings to a temporary file in the same directory, fsync the temporary file, and then rename it over the destination path.
The system SHALL set the destination file mode to `0644`.
The system SHALL NOT call `Save` automatically on application startup; it SHALL only run when the application explicitly requests a write.
The system SHALL validate the Settings via the validation contract (see "Validate settings before save") before writing; if validation fails, no temporary file SHALL remain on disk and no rename SHALL occur.

#### Scenario: Atomic replace on existing file

- **WHEN** `Save` is invoked while `~/.launchpal/settings.json` already exists with previous content
- **THEN** the file ends up containing the new Settings, the old content is no longer present, and at no intermediate point is the destination path observable as zero bytes or a partially written JSON document

#### Scenario: Parent directory created when missing

- **WHEN** `Save` is invoked on a fresh system where `~/.launchpal/` does not yet exist
- **THEN** the directory is created with mode `0755` and the settings file is written successfully


<!-- @trace
source: customize-log-paths
updated: 2026-05-09
code:
  - frontend/app/components/CreateServiceModal.vue
  - app.go
  - frontend/app/types/wails.d.ts
  - frontend/wailsjs/go/main/App.js
  - frontend/app/composables/useSettings.ts
  - frontend/wailsjs/go/models.ts
  - frontend/app/utils/settingsValidation.ts
  - frontend/wailsjs/go/main/App.d.ts
  - internal/settings/settings.go
  - internal/privhelper/handlers.go
  - frontend/app/pages/settings.vue
  - frontend/app/components/LogStorageSection.vue
  - frontend/app/utils/logPaths.ts
tests:
  - internal/privhelper/handlers_test.go
  - app_test.go
  - frontend/app/composables/__tests__/useSettings.test.ts
  - frontend/app/pages/__tests__/settings.test.ts
  - frontend/app/components/__tests__/CreateServiceModal.test.ts
  - internal/settings/settings_test.go
-->

### Requirement: Validate settings before save

The system SHALL provide a `Validate(s Settings)` operation that returns `nil` for valid Settings and a non-nil error for invalid Settings.
The system SHALL reject `userLogDir` values that are empty or contain any of the shell metacharacters `; & | $ \` newline carriage-return null-byte`.
The system SHALL accept `userLogDir` values that begin with `~/` (tilde-home) or `/` (absolute) and otherwise treat them as invalid; relative paths SHALL be rejected.
The system SHALL reject `systemLogDir` values that are empty or contain shell metacharacters as above.
The system SHALL require `systemLogDir` to be an absolute path whose canonical (after `filepath.Clean`) form has one of the prefixes `/var/log/`, `/private/var/log/`, `/Library/Logs/`, `/tmp/`, `/private/tmp/`, treating an exact match against the bare allowlist root as the prefix being satisfied (e.g. `/Library/Logs` is accepted because the resulting log file `<systemLogDir>/<label>/<stream>.log` always interpolates a label that adds depth).
The system SHALL source the allowed prefix list from the same shared constant used by the privileged helper, so any future change applies to both validators simultaneously.
The system SHALL surface validation failures as Go errors whose `Error()` text identifies the invalid field name and the rule that was violated.

##### Example: systemLogDir validation matrix

| Input                       | Result  | Rule violated                                   |
| --------------------------- | ------- | ----------------------------------------------- |
| `/Library/Logs/launchpal`   | accept  | (valid)                                         |
| `/var/log/myapp`            | accept  | (valid)                                         |
| `/Library/Logs`             | accept  | (valid; bare allowlist root is acceptable because label adds depth) |
| `/Library/Logs/`            | accept  | (valid; trailing slash normalizes via filepath.Clean) |
| `/etc/launchpal`            | reject  | prefix not in allowlist                         |
| `~/Library/Logs/myapp`      | reject  | systemLogDir must be absolute (no tilde)        |
| ``                          | reject  | empty                                           |
| `/var/log/$(rm -rf)`        | reject  | contains shell metacharacters                   |

##### Example: userLogDir validation matrix

| Input                       | Result  | Rule violated                                   |
| --------------------------- | ------- | ----------------------------------------------- |
| `~/Library/Logs`            | accept  | (valid)                                         |
| `/Users/foo/logs`           | accept  | (valid)                                         |
| `Library/Logs`              | reject  | must be tilde-home or absolute                  |
| ``                          | reject  | empty                                           |
| `~/logs;rm`                 | reject  | contains shell metacharacters                   |

#### Scenario: Invalid systemLogDir blocks save

- **WHEN** `Save` is invoked with `systemLogDir = "/etc/launchpal/logs"`
- **THEN** an error is returned identifying `systemLogDir` and "must start with one of: /var/log/, /private/var/log/, /Library/Logs/, /tmp/, /private/tmp/"; the on-disk settings file is unchanged

#### Scenario: Bare allowlist root accepted as default

- **WHEN** `Validate` is invoked with `systemLogDir = "/Library/Logs"` (the built-in default) and a valid `userLogDir`
- **THEN** the operation returns `nil`, because the consuming `<systemLogDir>/<label>/<stream>.log` composition guarantees at least one component of additional depth

#### Scenario: Tilde and absolute userLogDir accepted

- **WHEN** `Validate` is invoked with `userLogDir = "~/Library/Logs"` or `userLogDir = "/Users/jeff/logs"` and a valid `systemLogDir`
- **THEN** the operation returns `nil`


<!-- @trace
source: customize-log-paths
updated: 2026-05-09
code:
  - frontend/app/components/CreateServiceModal.vue
  - app.go
  - frontend/app/types/wails.d.ts
  - frontend/wailsjs/go/main/App.js
  - frontend/app/composables/useSettings.ts
  - frontend/wailsjs/go/models.ts
  - frontend/app/utils/settingsValidation.ts
  - frontend/wailsjs/go/main/App.d.ts
  - internal/settings/settings.go
  - internal/privhelper/handlers.go
  - frontend/app/pages/settings.vue
  - frontend/app/components/LogStorageSection.vue
  - frontend/app/utils/logPaths.ts
tests:
  - internal/privhelper/handlers_test.go
  - app_test.go
  - frontend/app/composables/__tests__/useSettings.test.ts
  - frontend/app/pages/__tests__/settings.test.ts
  - frontend/app/components/__tests__/CreateServiceModal.test.ts
  - internal/settings/settings_test.go
-->

### Requirement: GetSettings Wails binding

The system SHALL expose a `GetSettings()` method on the Wails App struct that returns the current Settings value.
The method SHALL invoke `Load()` and return its result.
The method SHALL return a Settings value to the frontend even when the underlying load encountered a non-fatal recovery (missing file, corrupt JSON), populated with the Default values.
The method SHALL only return an error to the frontend for unrecoverable filesystem errors as defined by `Load()`.

#### Scenario: GetSettings on first run

- **WHEN** the frontend invokes `GetSettings` and `~/.launchpal/settings.json` does not exist
- **THEN** the returned object has `userLogDir = "~/Library/Logs"` and `systemLogDir = "/Library/Logs"` and no error is propagated


<!-- @trace
source: customize-log-paths
updated: 2026-05-09
code:
  - frontend/app/components/CreateServiceModal.vue
  - app.go
  - frontend/app/types/wails.d.ts
  - frontend/wailsjs/go/main/App.js
  - frontend/app/composables/useSettings.ts
  - frontend/wailsjs/go/models.ts
  - frontend/app/utils/settingsValidation.ts
  - frontend/wailsjs/go/main/App.d.ts
  - internal/settings/settings.go
  - internal/privhelper/handlers.go
  - frontend/app/pages/settings.vue
  - frontend/app/components/LogStorageSection.vue
  - frontend/app/utils/logPaths.ts
tests:
  - internal/privhelper/handlers_test.go
  - app_test.go
  - frontend/app/composables/__tests__/useSettings.test.ts
  - frontend/app/pages/__tests__/settings.test.ts
  - frontend/app/components/__tests__/CreateServiceModal.test.ts
  - internal/settings/settings_test.go
-->

### Requirement: UpdateSettings Wails binding

The system SHALL expose an `UpdateSettings(s Settings)` method on the Wails App struct.
The method SHALL invoke `Validate(s)`; on failure it SHALL return the validation error to the frontend without writing to disk.
The method SHALL invoke `Save(s)` after successful validation and return any save error to the frontend.
The method SHALL return `nil` on success.
The method SHALL NOT broadcast a Wails event after a successful save; the frontend reload of dependent state is the caller's responsibility.

#### Scenario: Validation failure is propagated

- **WHEN** the frontend invokes `UpdateSettings` with `systemLogDir = "/etc/foo"`
- **THEN** the returned error contains the validation message identifying `systemLogDir`, the on-disk settings file is unchanged, and no Wails event is emitted

#### Scenario: Successful save

- **WHEN** the frontend invokes `UpdateSettings` with valid Settings
- **THEN** `~/.launchpal/settings.json` reflects the new values and the method returns `nil`

## Requirements


<!-- @trace
source: customize-log-paths
updated: 2026-05-09
code:
  - frontend/app/components/CreateServiceModal.vue
  - app.go
  - frontend/app/types/wails.d.ts
  - frontend/wailsjs/go/main/App.js
  - frontend/app/composables/useSettings.ts
  - frontend/wailsjs/go/models.ts
  - frontend/app/utils/settingsValidation.ts
  - frontend/wailsjs/go/main/App.d.ts
  - internal/settings/settings.go
  - internal/privhelper/handlers.go
  - frontend/app/pages/settings.vue
  - frontend/app/components/LogStorageSection.vue
  - frontend/app/utils/logPaths.ts
tests:
  - internal/privhelper/handlers_test.go
  - app_test.go
  - frontend/app/composables/__tests__/useSettings.test.ts
  - frontend/app/pages/__tests__/settings.test.ts
  - frontend/app/components/__tests__/CreateServiceModal.test.ts
  - internal/settings/settings_test.go
-->

### Requirement: Settings file location and format

The system SHALL persist application settings as a JSON file at `~/.launchpal/settings.json`.
The system SHALL encode the file using UTF-8 with two-space indentation and a trailing newline.
The system SHALL place the settings file in the same root directory (`~/.launchpal/`) as the existing backup store.
The system SHALL NOT depend on any third-party JSON or YAML library; it SHALL use the Go standard library `encoding/json` exclusively.

#### Scenario: Settings file written to canonical path

- **WHEN** the application saves settings for the first time
- **THEN** the file `~/.launchpal/settings.json` exists, contains a valid JSON object, ends with a newline, and uses two-space indentation

---
### Requirement: Default settings values

The system SHALL define built-in default values for every settings field:
- `userLogDir`: `~/Library/Logs`
- `systemLogDir`: `/Library/Logs`

The system SHALL expose a `Default()` constructor that returns a Settings value populated with the defaults above.
The system SHALL NOT write the defaults file at application startup; defaults SHALL only be returned in memory until the user explicitly saves or modifies a value.

#### Scenario: Defaults returned without filesystem side-effects

- **WHEN** `Default()` is invoked on a system where `~/.launchpal/settings.json` does not exist
- **THEN** a Settings value is returned with `userLogDir = "~/Library/Logs"` and `systemLogDir = "/Library/Logs"`, and no file is created on disk

---
### Requirement: Load settings from disk

The system SHALL provide a `Load()` operation that reads `~/.launchpal/settings.json` and returns a Settings value.
The system SHALL return the result of `Default()` with no error when the file does not exist.
The system SHALL return the result of `Default()` with no error when the file exists but contains invalid JSON; the implementation SHALL log a warning describing the parse error so the failure is observable in application logs.
The system SHALL fill any field that is absent from the JSON with its default value (forward-compatible decoding).
The system SHALL return an error only for filesystem errors other than "file does not exist" (for example, permission denied on the parent directory).

#### Scenario: Missing file returns defaults

- **WHEN** `Load()` is invoked and `~/.launchpal/settings.json` does not exist
- **THEN** the returned Settings equals `Default()` and no error is returned

#### Scenario: Corrupt JSON falls back to defaults

- **WHEN** `Load()` is invoked and `~/.launchpal/settings.json` contains the bytes `{not json`
- **THEN** the returned Settings equals `Default()`, no error is returned, and a warning is written to the application log

#### Scenario: Partial JSON merges with defaults

- **WHEN** `Load()` is invoked and `~/.launchpal/settings.json` contains `{"userLogDir": "/tmp/mylogs"}`
- **THEN** the returned Settings has `userLogDir = "/tmp/mylogs"` and `systemLogDir = "/Library/Logs"`

---
### Requirement: Atomic save to disk

The system SHALL provide a `Save(s Settings)` operation that persists the given Settings to `~/.launchpal/settings.json`.
The system SHALL ensure the parent directory `~/.launchpal/` exists with mode `0755` before writing, creating it if necessary.
The system SHALL write atomically: the implementation SHALL marshal the Settings to a temporary file in the same directory, fsync the temporary file, and then rename it over the destination path.
The system SHALL set the destination file mode to `0644`.
The system SHALL NOT call `Save` automatically on application startup; it SHALL only run when the application explicitly requests a write.
The system SHALL validate the Settings via the validation contract (see "Validate settings before save") before writing; if validation fails, no temporary file SHALL remain on disk and no rename SHALL occur.

#### Scenario: Atomic replace on existing file

- **WHEN** `Save` is invoked while `~/.launchpal/settings.json` already exists with previous content
- **THEN** the file ends up containing the new Settings, the old content is no longer present, and at no intermediate point is the destination path observable as zero bytes or a partially written JSON document

#### Scenario: Parent directory created when missing

- **WHEN** `Save` is invoked on a fresh system where `~/.launchpal/` does not yet exist
- **THEN** the directory is created with mode `0755` and the settings file is written successfully

---
### Requirement: Validate settings before save

The system SHALL provide a `Validate(s Settings)` operation that returns `nil` for valid Settings and a non-nil error for invalid Settings.
The system SHALL reject `userLogDir` values that are empty or contain any of the shell metacharacters `; & | $ \` newline carriage-return null-byte`.
The system SHALL accept `userLogDir` values that begin with `~/` (tilde-home) or `/` (absolute) and otherwise treat them as invalid; relative paths SHALL be rejected.
The system SHALL reject `systemLogDir` values that are empty or contain shell metacharacters as above.
The system SHALL require `systemLogDir` to be an absolute path whose canonical (after `filepath.Clean`) form has one of the prefixes `/var/log/`, `/private/var/log/`, `/Library/Logs/`, `/tmp/`, `/private/tmp/`, treating an exact match against the bare allowlist root as the prefix being satisfied (e.g. `/Library/Logs` is accepted because the resulting log file `<systemLogDir>/<label>/<stream>.log` always interpolates a label that adds depth).
The system SHALL source the allowed prefix list from the same shared constant used by the privileged helper, so any future change applies to both validators simultaneously.
The system SHALL surface validation failures as Go errors whose `Error()` text identifies the invalid field name and the rule that was violated.

##### Example: systemLogDir validation matrix

| Input                       | Result  | Rule violated                                   |
| --------------------------- | ------- | ----------------------------------------------- |
| `/Library/Logs/launchpal`   | accept  | (valid)                                         |
| `/var/log/myapp`            | accept  | (valid)                                         |
| `/Library/Logs`             | accept  | (valid; bare allowlist root is acceptable because label adds depth) |
| `/Library/Logs/`            | accept  | (valid; trailing slash normalizes via filepath.Clean) |
| `/etc/launchpal`            | reject  | prefix not in allowlist                         |
| `~/Library/Logs/myapp`      | reject  | systemLogDir must be absolute (no tilde)        |
| ``                          | reject  | empty                                           |
| `/var/log/$(rm -rf)`        | reject  | contains shell metacharacters                   |

##### Example: userLogDir validation matrix

| Input                       | Result  | Rule violated                                   |
| --------------------------- | ------- | ----------------------------------------------- |
| `~/Library/Logs`            | accept  | (valid)                                         |
| `/Users/foo/logs`           | accept  | (valid)                                         |
| `Library/Logs`              | reject  | must be tilde-home or absolute                  |
| ``                          | reject  | empty                                           |
| `~/logs;rm`                 | reject  | contains shell metacharacters                   |

#### Scenario: Invalid systemLogDir blocks save

- **WHEN** `Save` is invoked with `systemLogDir = "/etc/launchpal/logs"`
- **THEN** an error is returned identifying `systemLogDir` and "must start with one of: /var/log/, /private/var/log/, /Library/Logs/, /tmp/, /private/tmp/"; the on-disk settings file is unchanged

#### Scenario: Bare allowlist root accepted as default

- **WHEN** `Validate` is invoked with `systemLogDir = "/Library/Logs"` (the built-in default) and a valid `userLogDir`
- **THEN** the operation returns `nil`, because the consuming `<systemLogDir>/<label>/<stream>.log` composition guarantees at least one component of additional depth

#### Scenario: Tilde and absolute userLogDir accepted

- **WHEN** `Validate` is invoked with `userLogDir = "~/Library/Logs"` or `userLogDir = "/Users/jeff/logs"` and a valid `systemLogDir`
- **THEN** the operation returns `nil`

---
### Requirement: GetSettings Wails binding

The system SHALL expose a `GetSettings()` method on the Wails App struct that returns the current Settings value.
The method SHALL invoke `Load()` and return its result.
The method SHALL return a Settings value to the frontend even when the underlying load encountered a non-fatal recovery (missing file, corrupt JSON), populated with the Default values.
The method SHALL only return an error to the frontend for unrecoverable filesystem errors as defined by `Load()`.

#### Scenario: GetSettings on first run

- **WHEN** the frontend invokes `GetSettings` and `~/.launchpal/settings.json` does not exist
- **THEN** the returned object has `userLogDir = "~/Library/Logs"` and `systemLogDir = "/Library/Logs"` and no error is propagated

---
### Requirement: UpdateSettings Wails binding

The system SHALL expose an `UpdateSettings(s Settings)` method on the Wails App struct.
The method SHALL invoke `Validate(s)`; on failure it SHALL return the validation error to the frontend without writing to disk.
The method SHALL invoke `Save(s)` after successful validation and return any save error to the frontend.
The method SHALL return `nil` on success.
The method SHALL NOT broadcast a Wails event after a successful save; the frontend reload of dependent state is the caller's responsibility.

#### Scenario: Validation failure is propagated

- **WHEN** the frontend invokes `UpdateSettings` with `systemLogDir = "/etc/foo"`
- **THEN** the returned error contains the validation message identifying `systemLogDir`, the on-disk settings file is unchanged, and no Wails event is emitted

#### Scenario: Successful save

- **WHEN** the frontend invokes `UpdateSettings` with valid Settings
- **THEN** `~/.launchpal/settings.json` reflects the new values and the method returns `nil`