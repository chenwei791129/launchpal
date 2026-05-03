# privileged-helper-rpc Specification

## Purpose

TBD - created by archiving change 'session-privileged-helper'. Update Purpose after archive.

## Requirements

### Requirement: Newline-delimited JSON RPC transport

Communication between LaunchPal and the helper SHALL use newline-delimited JSON messages over a Unix domain socket. Each message SHALL be a single line terminated by `\n`. Requests SHALL include `id` (monotonic integer), `method`, and optional `params`. Responses SHALL include the matching `id` and exactly one of `result` or `error`.

#### Scenario: Request-response correlation

- **WHEN** a client sends `{"id": 5, "method": "Ping"}`
- **THEN** the helper replies with `{"id": 5, "result": {"pong": true}}` on a single line terminated by `\n`

#### Scenario: Malformed JSON

- **WHEN** the helper receives a line that is not valid JSON
- **THEN** it returns `{"id": null, "error": {"code": "invalid_request", "message": "..."}}`


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
### Requirement: Socket path and permissions

The socket path SHALL be `$TMPDIR/launchpal-<launching-uid>-<16-hex-random>.sock`. The helper SHALL create the socket with mode `0600`. On shutdown or exit, the helper SHALL remove the socket file.

#### Scenario: Socket permissions after creation

- **WHEN** the helper creates the socket
- **THEN** the file mode is `0600` and the owner is `root`

#### Scenario: Socket path is per-user private

- **WHEN** the socket path is generated
- **THEN** it resides under `$TMPDIR` (which is per-user on macOS) with a random 16-hex-character suffix


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
### Requirement: Peer UID verification

For every new client connection, the helper SHALL query the peer UID via `LOCAL_PEERCRED` and reject the connection if it does not match the `--launching-uid` provided at startup. Rejected connections SHALL be closed immediately without reading any input.

#### Scenario: Matching peer UID

- **WHEN** a client whose UID equals `--launching-uid` connects
- **THEN** the connection is accepted and RPC processing begins

#### Scenario: Mismatched peer UID

- **WHEN** a client with a different UID connects
- **THEN** the helper closes the connection without processing any RPC


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
### Requirement: Supported RPC methods

The helper SHALL implement the following methods: `Ping`, `ListSystemDaemons`, `GetSystemDaemon`, `Bootstrap`, `Bootout`, `Kickstart`, `WritePlist`, `DeletePlist`, `EnsureLogAccess`, `TruncateLog`, `Shutdown`. Unknown methods SHALL return `{"error": {"code": "unknown_method", "message": "..."}}`.

#### Scenario: Unknown method

- **WHEN** a client sends `{"id": 1, "method": "DoesNotExist"}`
- **THEN** the helper returns `{"id": 1, "error": {"code": "unknown_method", "message": "..."}}`

#### Scenario: TruncateLog is reachable

- **WHEN** a client sends a valid `{"id": N, "method": "TruncateLog", "params": {"path": "<allowed path>"}}`
- **THEN** the helper dispatches to the TruncateLog handler and returns either `OKResult` or an error response, but never `unknown_method`


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
### Requirement: Serial request processing

The helper SHALL process RPC requests serially on a single connection. A new request SHALL NOT begin processing until the previous response has been written.

#### Scenario: Sequential processing

- **WHEN** a client pipelines three requests on one connection
- **THEN** the helper processes and responds to them in order


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
### Requirement: Error code taxonomy

RPC error responses SHALL use codes from a fixed set: `invalid_request`, `unknown_method`, `invalid_params`, `permission_denied`, `not_found`, `launchctl_failed`, `io_error`, `internal_error`. Each error response SHALL include a human-readable `message`.

#### Scenario: launchctl failure

- **WHEN** `Bootstrap` is called and `launchctl bootstrap system <path>` exits non-zero
- **THEN** the RPC returns `{"error": {"code": "launchctl_failed", "message": "<stderr>"}}`

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
### Requirement: TruncateLog RPC method

The helper SHALL implement a `TruncateLog(path)` RPC that truncates an existing log file to 0 bytes while preserving its inode, owner, group, and mode.

The handler SHALL:
1. Validate `path` against the same allowlist used by `EnsureLogAccess` (`/var/log/`, `/private/var/log/`, `/Library/Logs/`, `/tmp/`, `/private/tmp/`).
2. Reject paths that are not absolute.
3. Reject paths that, after `filepath.Clean`, do not start with one of the allowed prefixes.
4. Reject paths whose immediate parent equals an allowlist root (e.g. `/var/log/system.log` is rejected because it would target a system log root directly).
5. Verify the file exists; if not, return `not_found`.
6. Open the file with `O_WRONLY | O_TRUNC | syscallNoFollow` and immediately close it on success.
7. Return `OKResult{OK: true}` on success.

The handler SHALL NOT create a missing file. The handler SHALL NOT change owner, group, or mode. The handler SHALL NOT follow symbolic links at any step; an `ELOOP` from the OpenFile call SHALL be surfaced as `io_error`.

The error code mapping SHALL be:
- Path validation failure → `invalid_params`
- File does not exist → `not_found`
- OpenFile failure with `EACCES`, `ELOOP`, or other I/O errors → `io_error`

#### Scenario: Truncate root-owned log

- **WHEN** `TruncateLog("/var/log/com.example.log")` is called for an existing root:wheel 0644 file
- **THEN** the helper opens it with O_TRUNC, the file size becomes 0, and the result is `{ok: true}`. Owner and mode are unchanged.

#### Scenario: Path outside allowlist

- **WHEN** `TruncateLog("/etc/passwd")` is called
- **THEN** the helper returns `{error: {code: "invalid_params", message: "log path not under an allowed prefix"}}` and the file is not opened

#### Scenario: Path immediate-parent equals allowlist root

- **WHEN** `TruncateLog("/var/log/system.log")` is called
- **THEN** the helper returns `{error: {code: "invalid_params", message: "log path must live in a sub-directory of /var/log"}}` and the file is not opened

#### Scenario: Missing file is not_found

- **WHEN** `TruncateLog("/var/log/myapp/never-created.log")` is called and the file does not exist
- **THEN** the helper returns `{error: {code: "not_found", message: "<path>"}}` and no file is created on disk

#### Scenario: Symlink at log path

- **WHEN** the path resolves to a symbolic link
- **THEN** the helper returns `{error: {code: "io_error", ...}}` and the link target is not truncated

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