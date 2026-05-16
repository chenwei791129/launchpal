## Requirements

### Requirement: DeleteLogPaths RPC method

The privileged helper SHALL expose a `DeleteLogPaths` RPC method that accepts a list of file paths and deletes each one.

For each path in the list the helper SHALL:
1. Validate the path against the log allowlist (`/var/log/`, `/private/var/log/`, `/Library/Logs/`, `/tmp/`, `/private/tmp/`) and require the path to be at least one subdirectory level deep inside the allowlist root.
2. Use `os.Lstat` (not `os.Stat`) to inspect the path without following symlinks; if the target is a symlink the helper SHALL reject it and record an error for that path.
3. If the target is a regular file, delete it with `os.Remove`.
4. After deleting the file, attempt to delete the parent directory with `os.Remove`; if the parent directory is non-empty (`syscall.ENOTEMPTY`) the helper SHALL silently ignore the error.
5. Collect all per-path errors and return them in the response; a partial success (some paths deleted, some failed) is a valid response state.

The helper SHALL NOT delete directories that are not empty, SHALL NOT recurse into subdirectories, and SHALL NOT follow symlinks at any point.

#### Scenario: Delete a single log file in an allowed path

- **WHEN** `DeleteLogPaths` is called with `["/var/log/myservice/out.log"]` and the file exists
- **THEN** the file is deleted, and if the parent directory `/var/log/myservice/` is now empty it is also deleted

#### Scenario: Parent directory has other files

- **WHEN** `DeleteLogPaths` is called and after deleting the log file the parent directory still contains other files
- **THEN** the file is deleted and the parent directory is left intact; no error is returned

#### Scenario: Path outside allowlist is rejected

- **WHEN** `DeleteLogPaths` is called with a path such as `/etc/passwd`
- **THEN** the helper records an error for that path and does not delete the file; other valid paths in the same request are still processed

#### Scenario: Path is a symlink

- **WHEN** `DeleteLogPaths` is called with a path that resolves (via `lstat`) to a symlink
- **THEN** the helper records an error for that path and does not follow or delete the symlink target

#### Scenario: Log file does not exist

- **WHEN** `DeleteLogPaths` is called with a path to a file that does not exist
- **THEN** the helper records an `os.ErrNotExist` error for that path and continues processing remaining paths

#### Scenario: Partial failure

- **WHEN** `DeleteLogPaths` is called with three paths and one path is outside the allowlist
- **THEN** the two valid paths are deleted and the response contains one error entry for the rejected path


<!-- @trace
source: delete-system-daemon-with-logs
updated: 2026-05-16
code:
  - internal/launchctl/system.go
  - frontend/wailsjs/go/main/App.js
  - frontend/app/types/wails.d.ts
  - internal/privhelper/client.go
  - internal/launchctl/types.go
  - internal/privhelper/protocol.go
  - frontend/wailsjs/go/main/App.d.ts
  - internal/privhelper/handlers.go
  - frontend/wailsjs/go/models.ts
  - frontend/app/pages/system.vue
  - app.go
tests:
  - app_test.go
  - internal/launchctl/system_test.go
  - internal/privhelper/client_test.go
  - internal/privhelper/protocol_test.go
  - internal/privhelper/handlers_test.go
-->

---
### Requirement: DeleteLogPaths protocol coverage

`DeleteLogPaths` SHALL be registered in `AllMethods` and covered by `TestAllMethods_Coverage`.

#### Scenario: AllMethods coverage test

- **WHEN** the `TestAllMethods_Coverage` test runs
- **THEN** `MethodDeleteLogPaths` appears in `AllMethods` and the test passes


<!-- @trace
source: delete-system-daemon-with-logs
updated: 2026-05-16
code:
  - internal/launchctl/system.go
  - frontend/wailsjs/go/main/App.js
  - frontend/app/types/wails.d.ts
  - internal/privhelper/client.go
  - internal/launchctl/types.go
  - internal/privhelper/protocol.go
  - frontend/wailsjs/go/main/App.d.ts
  - internal/privhelper/handlers.go
  - frontend/wailsjs/go/models.ts
  - frontend/app/pages/system.vue
  - app.go
tests:
  - app_test.go
  - internal/launchctl/system_test.go
  - internal/privhelper/client_test.go
  - internal/privhelper/protocol_test.go
  - internal/privhelper/handlers_test.go
-->

---
### Requirement: Optional log deletion on system daemon delete

`SystemManager` SHALL expose a `DeleteWithOptions(name string, opts DeleteServiceOptions) error` method. `DeleteServiceOptions` SHALL contain a boolean field `DeleteLogs`.

When `DeleteWithOptions` is called with `DeleteLogs: true` the system SHALL:
1. Parse the live plist via `SystemManager.Get` BEFORE deletion to capture `StandardOutPath` and `StandardErrorPath`. If `Get` fails (typically Full Disk Access denied) the system SHALL return a `*LogDeletionWarning` after step 3 explaining that log paths could not be determined, instead of silently skipping the user's explicit request.
2. Execute the existing Delete flow (bootout + DeletePlist RPC).
3. Collect non-empty paths from step 1 into a slice.
4. If the slice is non-empty, call `DeleteLogPaths` RPC with the collected paths.
5. If `DeleteLogPaths` returns per-path errors (warnings) or a transport error, surface them as a `*LogDeletionWarning` typed error so callers can treat the overall delete as success.

When `DeleteWithOptions` is called with `DeleteLogs: false` the system SHALL behave identically to the existing `Delete` method.

#### Scenario: Delete with log cleanup enabled

- **WHEN** `DeleteWithOptions` is called with `DeleteLogs: true` for a service that has `StandardOutPath: /var/log/svc/out.log`
- **THEN** the plist is booted out and deleted, then `/var/log/svc/out.log` is deleted via `DeleteLogPaths`, and the empty parent `/var/log/svc/` is removed

#### Scenario: Delete with log cleanup disabled (default)

- **WHEN** `DeleteWithOptions` is called with `DeleteLogs: false`
- **THEN** only the plist is deleted; no `DeleteLogPaths` RPC is issued

#### Scenario: Plist has no log paths configured

- **WHEN** `DeleteWithOptions` is called with `DeleteLogs: true` for a service without `StandardOutPath` or `StandardErrorPath`
- **THEN** the plist is deleted normally and no `DeleteLogPaths` call is made

#### Scenario: Log cleanup fails but plist deletion succeeded

- **WHEN** `DeleteWithOptions` is called with `DeleteLogs: true` and the `DeleteLogPaths` RPC returns errors
- **THEN** the delete operation returns a `*LogDeletionWarning` typed error containing the log deletion entries; callers treat this as overall success and surface the warning

#### Scenario: Log path discovery fails under DeleteLogs=true

- **WHEN** `DeleteWithOptions` is called with `DeleteLogs: true` but `SystemManager.Get` cannot read the plist (e.g. Full Disk Access not granted)
- **THEN** the daemon plist is still deleted, no `DeleteLogPaths` RPC is issued, and the operation returns a `*LogDeletionWarning` containing a message explaining that log paths could not be determined


<!-- @trace
source: delete-system-daemon-with-logs
updated: 2026-05-16
code:
  - internal/launchctl/system.go
  - frontend/wailsjs/go/main/App.js
  - frontend/app/types/wails.d.ts
  - internal/privhelper/client.go
  - internal/launchctl/types.go
  - internal/privhelper/protocol.go
  - frontend/wailsjs/go/main/App.d.ts
  - internal/privhelper/handlers.go
  - frontend/wailsjs/go/models.ts
  - frontend/app/pages/system.vue
  - app.go
tests:
  - app_test.go
  - internal/launchctl/system_test.go
  - internal/privhelper/client_test.go
  - internal/privhelper/protocol_test.go
  - internal/privhelper/handlers_test.go
-->

---
### Requirement: Delete dialog log cleanup checkbox

The system daemon delete confirmation dialog SHALL include a checkbox labeled "Also delete log files".

- The checkbox SHALL default to unchecked.
- When checked and the user confirms deletion, the frontend SHALL call `DeleteSystemService` with `deleteLogs: true`.
- When unchecked or when `StandardOutPath`/`StandardErrorPath` are absent, the frontend SHALL call `DeleteSystemService` with `deleteLogs: false`.
- The dialog SHALL display helper text: "Log files will be permanently deleted and cannot be recovered."

#### Scenario: User deletes service with log cleanup checked

- **WHEN** the user opens the delete dialog, checks "Also delete log files", and confirms
- **THEN** `DeleteSystemService` is called with `{ deleteLogs: true }`

#### Scenario: User deletes service with log cleanup unchecked (default)

- **WHEN** the user opens the delete dialog without changing the checkbox and confirms
- **THEN** `DeleteSystemService` is called with `{ deleteLogs: false }`

#### Scenario: Service has no log paths — checkbox still visible but ineffective

- **WHEN** the user checks "Also delete log files" for a service with no `StandardOutPath` or `StandardErrorPath`
- **THEN** `DeleteSystemService` is called with `{ deleteLogs: true }` and the backend silently skips the `DeleteLogPaths` call


<!-- @trace
source: delete-system-daemon-with-logs
updated: 2026-05-16
code:
  - internal/launchctl/system.go
  - frontend/wailsjs/go/main/App.js
  - frontend/app/types/wails.d.ts
  - internal/privhelper/client.go
  - internal/launchctl/types.go
  - internal/privhelper/protocol.go
  - frontend/wailsjs/go/main/App.d.ts
  - internal/privhelper/handlers.go
  - frontend/wailsjs/go/models.ts
  - frontend/app/pages/system.vue
  - app.go
tests:
  - app_test.go
  - internal/launchctl/system_test.go
  - internal/privhelper/client_test.go
  - internal/privhelper/protocol_test.go
  - internal/privhelper/handlers_test.go
-->

---
### Requirement: App binding for DeleteSystemService options

`App.DeleteSystemService(name string, options DeleteServiceOptions) (string, error)` SHALL accept a `DeleteServiceOptions` struct with a `DeleteLogs bool` field, pass it through to `SystemManager.DeleteWithOptions`, and translate any `*LogDeletionWarning` into the first (warning) return value while clearing the error so the Wails Promise resolves rather than rejects.

#### Scenario: Wails binding receives options

- **WHEN** the frontend calls `DeleteSystemService("myservice", { deleteLogs: true })`
- **THEN** the App method calls `SystemManager.DeleteWithOptions("myservice", DeleteServiceOptions{DeleteLogs: true})`

#### Scenario: LogDeletionWarning is translated to a warning string

- **WHEN** `SystemManager.DeleteWithOptions` returns a `*LogDeletionWarning`
- **THEN** `App.DeleteSystemService` returns the warning's message as the first string return value and `nil` as the error, so the frontend Promise resolves with the warning text instead of rejecting

<!-- @trace
source: delete-system-daemon-with-logs
updated: 2026-05-16
code:
  - internal/launchctl/system.go
  - frontend/wailsjs/go/main/App.js
  - frontend/app/types/wails.d.ts
  - internal/privhelper/client.go
  - internal/launchctl/types.go
  - internal/privhelper/protocol.go
  - frontend/wailsjs/go/main/App.d.ts
  - internal/privhelper/handlers.go
  - frontend/wailsjs/go/models.ts
  - frontend/app/pages/system.vue
  - app.go
tests:
  - app_test.go
  - internal/launchctl/system_test.go
  - internal/privhelper/client_test.go
  - internal/privhelper/protocol_test.go
  - internal/privhelper/handlers_test.go
-->