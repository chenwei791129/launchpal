# privileged-helper-lifecycle Specification

## Purpose

Defines how the `launchpal-privhelper` binary is built, packaged inside `LaunchPal.app/Contents/MacOS/`, located at runtime, launched with administrator privileges via osascript, and torn down on idle timeout, parent exit, or explicit Disable so no privileged process survives LaunchPal's own process.

## Requirements

### Requirement: Helper binary packaged in app bundle

The `launchpal-privhelper` binary SHALL be compiled and placed at `LaunchPal.app/Contents/MacOS/launchpal-privhelper` during the build. LaunchPal SHALL locate the helper at runtime using `os.Executable()` to find the main binary directory, then joining `launchpal-privhelper`.

#### Scenario: Helper binary resolved at runtime

- **WHEN** LaunchPal enables Admin Mode
- **THEN** the helper path is computed as `<main-binary-dir>/launchpal-privhelper` and exists at that path

#### Scenario: Helper binary missing

- **WHEN** the helper binary is not found at the expected path
- **THEN** Admin Mode enablement fails with an error identifying the missing binary path


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
### Requirement: Helper refuses to run without required conditions

The helper SHALL exit immediately with a non-zero code when any of the following hold:

- The effective UID is not `0` (root)
- The `--socket` argument is missing or empty
- The `--parent-pid` argument is missing or does not resolve to a running process
- The `--launching-uid` argument is missing

#### Scenario: Non-root invocation

- **WHEN** helper is invoked with effective UID != 0
- **THEN** it exits with a non-zero code and prints an error to stderr

#### Scenario: Missing socket argument

- **WHEN** helper is invoked without `--socket`
- **THEN** it exits with a non-zero code


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
### Requirement: Helper launched via osascript with background execution

LaunchPal SHALL launch the helper by executing `osascript -e 'do shell script "<helper-path> --socket <path> --parent-pid <pid> --launching-uid <uid> &> /dev/null &" with administrator privileges'`. The `&` trailing operator ensures the helper continues running after `do shell script` returns.

#### Scenario: osascript authorization granted

- **WHEN** the user authorizes the osascript password/Touch ID prompt
- **THEN** `do shell script` returns successfully and the helper process starts as root

#### Scenario: osascript authorization cancelled

- **WHEN** the user cancels or fails the osascript prompt
- **THEN** LaunchPal receives an error indicating authorization was cancelled, and Admin Mode returns to Disabled without further action


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
### Requirement: Socket handshake with retry

After launching the helper, LaunchPal SHALL repeatedly attempt to connect to the socket path with exponential backoff until either: (a) connection succeeds and a `Ping` RPC returns a successful response, or (b) 10 seconds elapse, at which point Admin Mode enablement fails.

#### Scenario: Handshake succeeds within timeout

- **WHEN** LaunchPal connects to the socket and Ping returns ok within 10 seconds
- **THEN** Admin Mode transitions to Enabled

#### Scenario: Handshake times out

- **WHEN** LaunchPal cannot connect or Ping fails within 10 seconds
- **THEN** Admin Mode returns to Disabled with an error containing "helper handshake failed"


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
### Requirement: Parent PID watchdog

The helper SHALL record the parent LaunchPal process's start time at launch and SHALL spawn a background goroutine that checks the parent every second. The parent SHALL be considered alive only when a process with the parent PID exists AND its start time matches the recorded value. If the PID no longer exists, or exists but reports a different start time (PID reuse), the helper SHALL treat the parent as gone, remove the socket file, and exit within 2 seconds. On platforms where the start time cannot be obtained, the helper SHALL fall back to a PID-existence check.

#### Scenario: Parent exits normally

- **WHEN** LaunchPal exits without sending Shutdown
- **THEN** the helper detects the parent is gone, removes the socket, and exits within 2 seconds

#### Scenario: Parent PID reused by another process

- **WHEN** LaunchPal exits and its PID is subsequently claimed by an unrelated live process
- **THEN** the helper observes a mismatched parent start time, treats the parent as gone, and self-exits

#### Scenario: Parent is killed

- **WHEN** LaunchPal is killed via SIGKILL
- **THEN** the helper detects this within 1-2 seconds and self-exits


<!-- @trace
source: admin-mode-lifecycle-hardening
updated: 2026-07-23
code:
  - internal/privhelper/logpath_darwin.go
  - .github/workflows/build.yml
  - frontend/app/composables/useAdminMode.ts
  - internal/privhelper/integrity.go
  - cmd/launchpal-privhelper/procinfo_other.go
  - internal/launchctl/system.go
  - internal/privhelper/install.go
  - internal/privhelper/client.go
  - internal/privhelper/server.go
  - main.go
  - internal/launchctl/readonly.go
  - README.md
  - internal/launchctl/user.go
  - frontend/app/pages/settings.vue
  - cmd/launchpal-privhelper/main.go
  - admin_mode.go
  - cmd/launchpal-privhelper/procinfo_darwin.go
  - internal/privhelper/logpath_other.go
  - internal/privhelper/handlers.go
  - internal/privhelper/logpath.go
  - Makefile
tests:
  - internal/privhelper/server_test.go
  - internal/privhelper/handlers_test.go
  - resolve_helper_test.go
  - internal/launchctl/system_test.go
  - internal/privhelper/integrity_test.go
  - internal/launchctl/user_test.go
  - internal/privhelper/install_test.go
  - internal/privhelper/client_test.go
  - cmd/launchpal-privhelper/helper_test.go
  - admin_mode_test.go
-->

---
### Requirement: Graceful shutdown via RPC

The helper SHALL accept a `Shutdown` RPC method. Upon receiving it, the helper SHALL acknowledge the request, close the listener, remove the socket file, and exit with code `0`.

#### Scenario: Client requests shutdown

- **WHEN** LaunchPal sends Shutdown RPC
- **THEN** helper responds with ok, removes the socket, and exits cleanly


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
### Requirement: Idle timeout

The helper SHALL track the time of the last successful RPC. If no RPC is received for 5 minutes, the helper SHALL remove the socket and exit. Any successful RPC resets the idle timer, so an actively used session is unaffected; the timeout bounds only the window during which an idle-but-still-connected session keeps a root socket alive. Because the GUI holds a single long-lived connection that stays open while idle, the idle-driven stop SHALL close the active accepted connection (not only the listener) so that the connection handler unblocks and the helper process actually exits; closing the listener alone would leave a connected-but-idle helper running.

#### Scenario: Extended idle period with the GUI still connected

- **WHEN** no RPC traffic occurs for 5 minutes while the GUI connection is still open
- **THEN** the helper closes the active connection, cleans up, and self-exits (the process terminates, not merely the listener), and any subsequent client connection attempt fails

#### Scenario: Activity resets the idle timer

- **WHEN** RPCs continue to arrive at intervals shorter than 5 minutes
- **THEN** the helper does not self-exit on idle and remains available


<!-- @trace
source: admin-mode-lifecycle-hardening
updated: 2026-07-23
code:
  - internal/privhelper/logpath_darwin.go
  - .github/workflows/build.yml
  - frontend/app/composables/useAdminMode.ts
  - internal/privhelper/integrity.go
  - cmd/launchpal-privhelper/procinfo_other.go
  - internal/launchctl/system.go
  - internal/privhelper/install.go
  - internal/privhelper/client.go
  - internal/privhelper/server.go
  - main.go
  - internal/launchctl/readonly.go
  - README.md
  - internal/launchctl/user.go
  - frontend/app/pages/settings.vue
  - cmd/launchpal-privhelper/main.go
  - admin_mode.go
  - cmd/launchpal-privhelper/procinfo_darwin.go
  - internal/privhelper/logpath_other.go
  - internal/privhelper/handlers.go
  - internal/privhelper/logpath.go
  - Makefile
tests:
  - internal/privhelper/server_test.go
  - internal/privhelper/handlers_test.go
  - resolve_helper_test.go
  - internal/launchctl/system_test.go
  - internal/privhelper/integrity_test.go
  - internal/launchctl/user_test.go
  - internal/privhelper/install_test.go
  - internal/privhelper/client_test.go
  - cmd/launchpal-privhelper/helper_test.go
  - admin_mode_test.go
-->

---
### Requirement: Helper self-terminates on client disconnect

The helper serves a single client connection. WHEN that client connection ends for any reason — the connection handler returning on EOF, on a read error, or on a write/encode error — the helper SHALL remove the socket file and exit, ending the accept loop, rather than continuing to listen until the idle timeout or parent watchdog fires. This teardown SHALL be triggered on every connection-handler exit path that indicates the connection ended (not only the post-scan EOF path), and SHALL NOT be triggered when the server is already stopping for another reason. This is the primary teardown mechanism, because the unprivileged GUI cannot signal the root helper directly.

#### Scenario: Client disconnects (EOF)

- **WHEN** the LaunchPal client connection to the helper closes cleanly (Disable, GUI exit, GUI crash, or a transient drop) and the connection handler returns on EOF
- **THEN** the helper removes the socket and exits within a few seconds, and any subsequent connection attempt to the socket fails

#### Scenario: Disconnect surfaced via a failed write

- **WHEN** the connection dies while the helper is writing a response, so the connection handler returns from a failed write/encode rather than from the EOF path
- **THEN** the helper still removes the socket and exits within a few seconds — the failed-write return path is not exempt from teardown

<!-- @trace
source: admin-mode-lifecycle-hardening
updated: 2026-07-23
code:
  - internal/privhelper/logpath_darwin.go
  - .github/workflows/build.yml
  - frontend/app/composables/useAdminMode.ts
  - internal/privhelper/integrity.go
  - cmd/launchpal-privhelper/procinfo_other.go
  - internal/launchctl/system.go
  - internal/privhelper/install.go
  - internal/privhelper/client.go
  - internal/privhelper/server.go
  - main.go
  - internal/launchctl/readonly.go
  - README.md
  - internal/launchctl/user.go
  - frontend/app/pages/settings.vue
  - cmd/launchpal-privhelper/main.go
  - admin_mode.go
  - cmd/launchpal-privhelper/procinfo_darwin.go
  - internal/privhelper/logpath_other.go
  - internal/privhelper/handlers.go
  - internal/privhelper/logpath.go
  - Makefile
tests:
  - internal/privhelper/server_test.go
  - internal/privhelper/handlers_test.go
  - resolve_helper_test.go
  - internal/launchctl/system_test.go
  - internal/privhelper/integrity_test.go
  - internal/launchctl/user_test.go
  - internal/privhelper/install_test.go
  - internal/privhelper/client_test.go
  - cmd/launchpal-privhelper/helper_test.go
  - admin_mode_test.go
-->