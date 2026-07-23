# system-daemon-write-ops Specification

## Purpose

Defines the helper's write surface for `/Library/LaunchDaemons`: bootstrap/bootout/kickstart and atomic plist write/delete, all gated by strict path validation (must live under `/Library/LaunchDaemons/`) and label validation (reverse-DNS character class only). Pre-write backups are placed under the launching user's `~/.launchpal/backups/<label>/` and chowned back so the unprivileged GUI can read them.

## Requirements

### Requirement: Bootstrap a system daemon

When Admin Mode is Enabled and the client calls `Bootstrap(plistPath)`, the helper SHALL execute `launchctl bootstrap system <plistPath>` and return success if the exit code is `0`. On non-zero exit, the helper SHALL return an `launchctl_failed` error containing the stderr.

The `plistPath` MUST reside under `/Library/LaunchDaemons/`. Paths outside this directory SHALL be rejected with `invalid_params`.

#### Scenario: Successful bootstrap

- **WHEN** a valid plist under `/Library/LaunchDaemons/` is bootstrapped
- **THEN** the helper returns ok and the service appears in `launchctl print system`

#### Scenario: Path outside allowed directory

- **WHEN** `Bootstrap("/tmp/malicious.plist")` is called
- **THEN** the helper returns `{"error": {"code": "invalid_params", "message": "path must be under /Library/LaunchDaemons"}}`


<!-- @trace
source: session-privileged-helper
updated: 2026-04-22
code:
  - frontend/app/pages/services/[name].vue
  - frontend/app/components/ServiceLogs.vue
  - internal/launchctl/readonly.go
  - internal/launchctl/system.go
  - internal/privhelper/peer_darwin.go
  - frontend/wailsjs/go/main/App.d.ts
  - launchpal-privhelper
  - Makefile
  - cmd/launchpal-privhelper/main.go
  - go.mod
  - internal/launchctl/user.go
  - frontend/app/components/ServiceRow.vue
  - app.go
  - frontend/wailsjs/go/models.ts
  - internal/launchctl/plist_encode.go
  - frontend/app/components/BackupDiffDialog.vue
  - frontend/app/components/CreateServiceModal.vue
  - frontend/app/types/wails.d.ts
  - internal/privhelper/nofollow_other.go
  - internal/privhelper/protocol.go
  - internal/privhelper/server.go
  - frontend/app/composables/useAdminMode.ts
  - frontend/app/pages/system.vue
  - frontend/app/pages/settings.vue
  - internal/privhelper/nofollow_darwin.go
  - frontend/wailsjs/go/main/App.js
  - README.md
  - internal/privhelper/peer_other.go
  - internal/privhelper/handlers.go
  - admin_mode.go
  - internal/privhelper/client.go
tests:
  - internal/privhelper/handlers_test.go
  - internal/privhelper/server_test.go
  - internal/launchctl/plist_encode_test.go
  - internal/privhelper/protocol_test.go
  - internal/privhelper/client_test.go
  - cmd/launchpal-privhelper/helper_test.go
  - app_test.go
  - admin_mode_test.go
  - admin_mode_testhelpers_test.go
  - internal/launchctl/system_test.go
-->

---
### Requirement: Bootout a system daemon

The helper SHALL expose `Bootout(label)` that executes `launchctl bootout system/<label>`. The label SHALL match the pattern of a valid reverse-DNS identifier; arbitrary shell metacharacters SHALL be rejected.

#### Scenario: Bootout a running service

- **WHEN** `Bootout("com.example.daemon")` is called for a bootstrapped service
- **THEN** the helper returns ok and the service is no longer in `launchctl print system`

#### Scenario: Label contains shell metacharacter

- **WHEN** `Bootout("com.example; rm -rf /")` is called
- **THEN** the helper returns `invalid_params` without invoking launchctl


<!-- @trace
source: session-privileged-helper
updated: 2026-04-22
code:
  - frontend/app/pages/services/[name].vue
  - frontend/app/components/ServiceLogs.vue
  - internal/launchctl/readonly.go
  - internal/launchctl/system.go
  - internal/privhelper/peer_darwin.go
  - frontend/wailsjs/go/main/App.d.ts
  - launchpal-privhelper
  - Makefile
  - cmd/launchpal-privhelper/main.go
  - go.mod
  - internal/launchctl/user.go
  - frontend/app/components/ServiceRow.vue
  - app.go
  - frontend/wailsjs/go/models.ts
  - internal/launchctl/plist_encode.go
  - frontend/app/components/BackupDiffDialog.vue
  - frontend/app/components/CreateServiceModal.vue
  - frontend/app/types/wails.d.ts
  - internal/privhelper/nofollow_other.go
  - internal/privhelper/protocol.go
  - internal/privhelper/server.go
  - frontend/app/composables/useAdminMode.ts
  - frontend/app/pages/system.vue
  - frontend/app/pages/settings.vue
  - internal/privhelper/nofollow_darwin.go
  - frontend/wailsjs/go/main/App.js
  - README.md
  - internal/privhelper/peer_other.go
  - internal/privhelper/handlers.go
  - admin_mode.go
  - internal/privhelper/client.go
tests:
  - internal/privhelper/handlers_test.go
  - internal/privhelper/server_test.go
  - internal/launchctl/plist_encode_test.go
  - internal/privhelper/protocol_test.go
  - internal/privhelper/client_test.go
  - cmd/launchpal-privhelper/helper_test.go
  - app_test.go
  - admin_mode_test.go
  - admin_mode_testhelpers_test.go
  - internal/launchctl/system_test.go
-->

---
### Requirement: Kickstart a system daemon

The helper SHALL expose `Kickstart(label)` that executes `launchctl kickstart -k system/<label>`. The `-k` flag kills any existing instance and starts a fresh one.

#### Scenario: Kickstart an existing service

- **WHEN** `Kickstart("com.example.daemon")` is called
- **THEN** the helper returns ok and the service's PID changes to a newly spawned process


<!-- @trace
source: session-privileged-helper
updated: 2026-04-22
code:
  - frontend/app/pages/services/[name].vue
  - frontend/app/components/ServiceLogs.vue
  - internal/launchctl/readonly.go
  - internal/launchctl/system.go
  - internal/privhelper/peer_darwin.go
  - frontend/wailsjs/go/main/App.d.ts
  - launchpal-privhelper
  - Makefile
  - cmd/launchpal-privhelper/main.go
  - go.mod
  - internal/launchctl/user.go
  - frontend/app/components/ServiceRow.vue
  - app.go
  - frontend/wailsjs/go/models.ts
  - internal/launchctl/plist_encode.go
  - frontend/app/components/BackupDiffDialog.vue
  - frontend/app/components/CreateServiceModal.vue
  - frontend/app/types/wails.d.ts
  - internal/privhelper/nofollow_other.go
  - internal/privhelper/protocol.go
  - internal/privhelper/server.go
  - frontend/app/composables/useAdminMode.ts
  - frontend/app/pages/system.vue
  - frontend/app/pages/settings.vue
  - internal/privhelper/nofollow_darwin.go
  - frontend/wailsjs/go/main/App.js
  - README.md
  - internal/privhelper/peer_other.go
  - internal/privhelper/handlers.go
  - admin_mode.go
  - internal/privhelper/client.go
tests:
  - internal/privhelper/handlers_test.go
  - internal/privhelper/server_test.go
  - internal/launchctl/plist_encode_test.go
  - internal/privhelper/protocol_test.go
  - internal/privhelper/client_test.go
  - cmd/launchpal-privhelper/helper_test.go
  - app_test.go
  - admin_mode_test.go
  - admin_mode_testhelpers_test.go
  - internal/launchctl/system_test.go
-->

---
### Requirement: Atomic plist write with backup

The helper SHALL expose `WritePlist(plistPath, base64Data)` that:
1. Validates `plistPath` starts with `/Library/LaunchDaemons/`
2. If a file already exists at `plistPath`, reads it and invokes backup via the shared backup mechanism using `--user-home` for path resolution
3. Writes the new content to a temp file in the same directory, then renames atomically to `plistPath`
4. Sets ownership to `root:wheel` and mode `0644`
5. After successful write, chowns the newly created backup file(s) to the launching user's UID/GID so the user-side LaunchPal can read them

#### Scenario: Write new plist

- **WHEN** `WritePlist("/Library/LaunchDaemons/com.example.plist", "<base64>")` is called for a non-existent file
- **THEN** the file is created with mode 0644, owner root:wheel, and the content matches the decoded base64

#### Scenario: Overwrite existing plist triggers backup

- **WHEN** the same path is overwritten
- **THEN** the previous content is backed up under the user's `~/.launchpal/backups/<label>/` with metadata, and the backup files are owned by the user

#### Scenario: Path outside allowed directory

- **WHEN** `WritePlist("/etc/passwd", ...)` is called
- **THEN** the helper returns `invalid_params` and does not write


<!-- @trace
source: session-privileged-helper
updated: 2026-04-22
code:
  - frontend/app/pages/services/[name].vue
  - frontend/app/components/ServiceLogs.vue
  - internal/launchctl/readonly.go
  - internal/launchctl/system.go
  - internal/privhelper/peer_darwin.go
  - frontend/wailsjs/go/main/App.d.ts
  - launchpal-privhelper
  - Makefile
  - cmd/launchpal-privhelper/main.go
  - go.mod
  - internal/launchctl/user.go
  - frontend/app/components/ServiceRow.vue
  - app.go
  - frontend/wailsjs/go/models.ts
  - internal/launchctl/plist_encode.go
  - frontend/app/components/BackupDiffDialog.vue
  - frontend/app/components/CreateServiceModal.vue
  - frontend/app/types/wails.d.ts
  - internal/privhelper/nofollow_other.go
  - internal/privhelper/protocol.go
  - internal/privhelper/server.go
  - frontend/app/composables/useAdminMode.ts
  - frontend/app/pages/system.vue
  - frontend/app/pages/settings.vue
  - internal/privhelper/nofollow_darwin.go
  - frontend/wailsjs/go/main/App.js
  - README.md
  - internal/privhelper/peer_other.go
  - internal/privhelper/handlers.go
  - admin_mode.go
  - internal/privhelper/client.go
tests:
  - internal/privhelper/handlers_test.go
  - internal/privhelper/server_test.go
  - internal/launchctl/plist_encode_test.go
  - internal/privhelper/protocol_test.go
  - internal/privhelper/client_test.go
  - cmd/launchpal-privhelper/helper_test.go
  - app_test.go
  - admin_mode_test.go
  - admin_mode_testhelpers_test.go
  - internal/launchctl/system_test.go
-->

---
### Requirement: Delete plist with backup

The helper SHALL expose `DeletePlist(plistPath)` that:
1. Validates `plistPath` starts with `/Library/LaunchDaemons/`
2. Backs up the current content before deletion (same user-owned backup rules as WritePlist)
3. Removes the file

#### Scenario: Delete existing plist

- **WHEN** `DeletePlist("/Library/LaunchDaemons/com.example.plist")` is called
- **THEN** the file is backed up then removed

#### Scenario: Delete non-existent plist

- **WHEN** the target file does not exist
- **THEN** the helper returns `{"error": {"code": "not_found", "message": "..."}}`


<!-- @trace
source: session-privileged-helper
updated: 2026-04-22
code:
  - frontend/app/pages/services/[name].vue
  - frontend/app/components/ServiceLogs.vue
  - internal/launchctl/readonly.go
  - internal/launchctl/system.go
  - internal/privhelper/peer_darwin.go
  - frontend/wailsjs/go/main/App.d.ts
  - launchpal-privhelper
  - Makefile
  - cmd/launchpal-privhelper/main.go
  - go.mod
  - internal/launchctl/user.go
  - frontend/app/components/ServiceRow.vue
  - app.go
  - frontend/wailsjs/go/models.ts
  - internal/launchctl/plist_encode.go
  - frontend/app/components/BackupDiffDialog.vue
  - frontend/app/components/CreateServiceModal.vue
  - frontend/app/types/wails.d.ts
  - internal/privhelper/nofollow_other.go
  - internal/privhelper/protocol.go
  - internal/privhelper/server.go
  - frontend/app/composables/useAdminMode.ts
  - frontend/app/pages/system.vue
  - frontend/app/pages/settings.vue
  - internal/privhelper/nofollow_darwin.go
  - frontend/wailsjs/go/main/App.js
  - README.md
  - internal/privhelper/peer_other.go
  - internal/privhelper/handlers.go
  - admin_mode.go
  - internal/privhelper/client.go
tests:
  - internal/privhelper/handlers_test.go
  - internal/privhelper/server_test.go
  - internal/launchctl/plist_encode_test.go
  - internal/privhelper/protocol_test.go
  - internal/privhelper/client_test.go
  - cmd/launchpal-privhelper/helper_test.go
  - app_test.go
  - admin_mode_test.go
  - admin_mode_testhelpers_test.go
  - internal/launchctl/system_test.go
-->

---
### Requirement: List and get system daemons via helper

The helper SHALL expose `ListSystemDaemons()` returning an array of `{ Label, Status, PID }` by invoking `launchctl print system` as root and parsing its output. The helper SHALL expose `GetSystemDaemon(label)` returning the same structure for a single label via `launchctl print system/<label>`.

#### Scenario: List returns authoritative status

- **WHEN** `ListSystemDaemons()` is called while Admin Mode is Enabled
- **THEN** the response contains every bootstrapped system daemon with its current Status and PID as reported by root-privileged launchctl

<!-- @trace
source: session-privileged-helper
updated: 2026-04-22
code:
  - frontend/app/pages/services/[name].vue
  - frontend/app/components/ServiceLogs.vue
  - internal/launchctl/readonly.go
  - internal/launchctl/system.go
  - internal/privhelper/peer_darwin.go
  - frontend/wailsjs/go/main/App.d.ts
  - launchpal-privhelper
  - Makefile
  - cmd/launchpal-privhelper/main.go
  - go.mod
  - internal/launchctl/user.go
  - frontend/app/components/ServiceRow.vue
  - app.go
  - frontend/wailsjs/go/models.ts
  - internal/launchctl/plist_encode.go
  - frontend/app/components/BackupDiffDialog.vue
  - frontend/app/components/CreateServiceModal.vue
  - frontend/app/types/wails.d.ts
  - internal/privhelper/nofollow_other.go
  - internal/privhelper/protocol.go
  - internal/privhelper/server.go
  - frontend/app/composables/useAdminMode.ts
  - frontend/app/pages/system.vue
  - frontend/app/pages/settings.vue
  - internal/privhelper/nofollow_darwin.go
  - frontend/wailsjs/go/main/App.js
  - README.md
  - internal/privhelper/peer_other.go
  - internal/privhelper/handlers.go
  - admin_mode.go
  - internal/privhelper/client.go
tests:
  - internal/privhelper/handlers_test.go
  - internal/privhelper/server_test.go
  - internal/launchctl/plist_encode_test.go
  - internal/privhelper/protocol_test.go
  - internal/privhelper/client_test.go
  - cmd/launchpal-privhelper/helper_test.go
  - app_test.go
  - admin_mode_test.go
  - admin_mode_testhelpers_test.go
  - internal/launchctl/system_test.go
-->

---
### Requirement: Truncate system daemon log with permission dispatch

The system SHALL expose `ClearLogs(name, logType)` on `SystemManager` that truncates a system daemon's stdout or stderr log file to 0 bytes.

The implementation SHALL dispatch on per-file write permission, in this order:
1. Resolve the log path from the daemon's plist; if unset, return an error indicating no log path is configured.
2. Verify the file exists; if not, return an error indicating the log file does not exist (the operation SHALL NOT create the file).
3. Attempt `os.OpenFile(path, O_WRONLY|O_TRUNC|O_NOFOLLOW, 0)` directly.
   - On success, close the file and return success. Admin Mode is not consulted.
   - On `EACCES`, fall through to step 4.
   - On any other error (e.g. `ELOOP` from a symlink, `ENOENT` for a path that disappeared between steps 2 and 3), return that error.
4. If Admin Mode is enabled, route through the privileged helper's `TruncateLog` RPC.
5. If Admin Mode is disabled, return `ErrReadOnlyManager`.

The system SHALL NOT consult `os.Stat` mode bits or compare uid/gid as a substitute for the OpenFile attempt; permission decisions SHALL come from the kernel response to the actual open call to avoid TOCTOU races.

The system SHALL NOT escalate to the helper when the OpenFile error is anything other than `EACCES`. In particular, `ENOENT`, `ELOOP`, and `EISDIR` SHALL surface to the caller.

The `apple-system` manager SHALL NOT expose `ClearLogs` and SHALL reject any attempt to clear logs with `ErrReadOnlyManager` when called via the Wails layer.

#### Scenario: User-writable log truncated directly

- **WHEN** the active user has write permission on `/usr/local/var/log/myapp.out.log` (mode 0664, group `admin`) and Admin Mode is disabled
- **THEN** `SystemManager.ClearLogs` truncates the file directly without invoking the privileged helper

#### Scenario: Root-owned log requires Admin Mode

- **WHEN** the active log path is `/var/log/com.example.log` owned by `root:wheel` mode 0644 and Admin Mode is enabled
- **THEN** the OpenFile attempt fails with `EACCES`, the system routes to the helper's `TruncateLog`, and the file is truncated to 0 bytes with owner and mode unchanged

#### Scenario: Root-owned log without Admin Mode is rejected

- **WHEN** the active log path is root-owned and Admin Mode is disabled
- **THEN** the operation returns `ErrReadOnlyManager` and no file is modified

#### Scenario: Symlink at log path returns ELOOP, not escalated

- **WHEN** the configured log path is a symbolic link
- **THEN** the OpenFile attempt fails with `ELOOP`, the operation returns the error, and the helper is NOT contacted

#### Scenario: Apple system service rejects clear

- **WHEN** the Wails layer calls `ClearSystemLogs` with `serviceType = "apple-system"`
- **THEN** the call returns an error indicating apple-system services are read-only and the helper is NOT contacted


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

---
### Requirement: Re-validate authorization at clear time

The system SHALL NOT cache or trust any prior `GetLogClearStatus` result when executing `ClearLogs`. Each call SHALL perform its own OpenFile attempt and Admin Mode check at the time of execution.

#### Scenario: Admin Mode disables between status query and clear

- **WHEN** `GetLogClearStatus` returns `userWritable=false` for a system log, the user confirms in the UI, and Admin Mode auto-disables (idle timeout) before the `ClearSystemLogs` call lands
- **THEN** `ClearLogs` re-checks the helper connection, finds it absent, and returns an error indicating Admin Mode is required

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

---
### Requirement: Helper invokes launchctl by absolute path

The helper SHALL invoke `launchctl` by its absolute path `/bin/launchctl` for the bootstrap, bootout, and kickstart operations, rather than by the bare name `launchctl` resolved through `$PATH`. This makes the resolved binary independent of any inherited environment.

#### Scenario: launchctl resolved by absolute path

- **WHEN** the helper performs a bootstrap, bootout, or kickstart
- **THEN** it executes `/bin/launchctl`, regardless of the `$PATH` value inherited by the helper process


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
### Requirement: System daemon schedule validation parity

System daemon create and update SHALL apply the same schedule range validation as user services: `StartInterval` SHALL be at least 10, and calendar-entry fields SHALL be within range (minute 0-59, hour 0-23, day 1-31, weekday 0-6, month 1-12). In addition, the system-domain create and update path SHALL reject a schedule whose calendar-entry count exceeds the 50-entry cap (matching the cron range-expansion limit). Both checks SHALL be performed in the create/update path (which returns an error and writes no plist on failure), not in the error-less plist encoder, so the enforcement holds for every caller of the system create/update binding rather than only in the frontend form.

#### Scenario: Out-of-range system daemon schedule is rejected

- **WHEN** a system daemon create or update is called with an out-of-range calendar field or a `StartInterval` below 10
- **THEN** it returns a validation error and does not write the plist

#### Scenario: Over-cap system daemon schedule is rejected in the create/update path

- **WHEN** a system daemon create or update is called with more than 50 calendar entries
- **THEN** the create/update path returns a validation error and writes no plist, independently of the frontend

##### Example: system-domain schedule validation

| Input                                   | Result            |
| --------------------------------------- | ----------------- |
| StartInterval = 9                       | rejected          |
| StartInterval = 10                      | accepted          |
| calendar Hour = 24                      | rejected          |
| calendar Hour = 23                      | accepted          |
| expansion producing 51 calendar entries | rejected          |
| expansion producing 50 calendar entries | accepted          |

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