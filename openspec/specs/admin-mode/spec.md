# admin-mode Specification

## Purpose

TBD - created by archiving change 'session-privileged-helper'. Update Purpose after archive.

## Requirements

### Requirement: Admin Mode status states

LaunchPal SHALL track Admin Mode using one of the following states: `Disabled`, `Requesting`, `Enabled`, `ShuttingDown`. State transitions SHALL occur only via the defined events: user-initiated Enable, user-initiated Disable, authorization result, helper crash, LaunchPal exit.

#### Scenario: Initial state

- **WHEN** LaunchPal starts
- **THEN** Admin Mode is `Disabled`

#### Scenario: Successful enablement path

- **WHEN** user clicks Enable → osascript authorized → helper handshake succeeds
- **THEN** state transitions Disabled → Requesting → Enabled

#### Scenario: User cancels authorization

- **WHEN** user clicks Enable → cancels osascript prompt
- **THEN** state transitions Disabled → Requesting → Disabled with error `authorization_cancelled`

#### Scenario: Handshake failure

- **WHEN** helper is launched but handshake times out
- **THEN** state transitions Disabled → Requesting → Disabled with error `helper_handshake_failed`


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
### Requirement: Wails bindings for Admin Mode

`app.go` SHALL expose Wails bindings `EnableAdminMode() error`, `DisableAdminMode() error`, and `GetAdminModeStatus() AdminModeStatus`. The `AdminModeStatus` struct SHALL include `State string` and `Error *string`.

#### Scenario: GetAdminModeStatus returns current state

- **WHEN** the frontend calls `GetAdminModeStatus()` while Admin Mode is Enabled
- **THEN** the response is `{ State: "enabled", Error: null }`

#### Scenario: EnableAdminMode while already enabled is a no-op

- **WHEN** `EnableAdminMode()` is called and state is already `Enabled`
- **THEN** the call returns successfully without spawning a new helper


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
### Requirement: Disable shuts down helper gracefully

`DisableAdminMode()` SHALL send a `Shutdown` RPC to the helper, wait up to 3 seconds for acknowledgment, then close the connection. After shutdown, state SHALL return to `Disabled`.

#### Scenario: Clean disable

- **WHEN** user clicks Disable
- **THEN** state goes Enabled → ShuttingDown → Disabled, and the socket file is removed

#### Scenario: Helper unresponsive on shutdown

- **WHEN** Shutdown RPC does not respond within 3 seconds
- **THEN** LaunchPal closes the connection anyway and returns to Disabled; parent watchdog on the helper side ensures eventual cleanup


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
### Requirement: Helper crash detection and recovery

LaunchPal SHALL detect helper crashes by observing EOF or connection errors on the RPC connection. Upon detection, Admin Mode SHALL transition to `Disabled` with error `helper_crashed`, and the frontend SHALL be notified via event or next `GetAdminModeStatus()` poll.

#### Scenario: Helper crashes mid-operation

- **WHEN** the helper exits unexpectedly while Admin Mode is Enabled
- **THEN** the next RPC call returns a connection error and state becomes `Disabled` with `helper_crashed`


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
### Requirement: Admin Mode UI in Settings page

The Settings page SHALL contain an Admin Mode section displaying: current state, last error (if any), Enable button (visible when Disabled), Disable button (visible when Enabled), and explanatory text about what Admin Mode grants. The section SHALL update reactively when state changes.

#### Scenario: Settings reflects Enabled state

- **WHEN** Admin Mode becomes Enabled
- **THEN** the Settings Admin Mode section shows "Enabled" with a Disable button

#### Scenario: Settings shows cancel reason

- **WHEN** user cancels osascript authorization
- **THEN** the Settings section shows the `authorization_cancelled` error message and an Enable button to retry


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
### Requirement: Write controls conditionally rendered

System service list (`pages/system.vue`) and detail (`pages/services/[name].vue`) SHALL display Start/Stop/Restart/Edit/Delete/Create controls only when Admin Mode is `Enabled`. When `Disabled`, these controls SHALL be hidden or shown as a lock icon with tooltip "Enable Admin Mode to manage".

#### Scenario: Write buttons hidden when Admin Mode Disabled

- **WHEN** the user views system service detail with Admin Mode Disabled
- **THEN** Start/Stop/Restart/Edit/Delete are not shown; a lock affordance is shown instead

#### Scenario: Write buttons visible when Enabled

- **WHEN** Admin Mode is Enabled
- **THEN** the same page shows functional Start/Stop/Restart/Edit/Delete controls

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