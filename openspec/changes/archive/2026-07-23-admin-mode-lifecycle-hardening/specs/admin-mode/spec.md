## MODIFIED Requirements

### Requirement: Admin Mode status states

LaunchPal SHALL track Admin Mode using one of the following states: `Disabled`, `Requesting`, `Enabled`, `ShuttingDown`. State transitions SHALL occur only via the defined events: user-initiated Enable, user-initiated Disable, authorization result, helper crash, LaunchPal exit. If a user-initiated Disable occurs while the state is `Requesting` (the authorization prompt is in flight), LaunchPal SHALL record a pending-disable intent instead of ignoring the request. WHEN the in-flight Enable subsequently completes its handshake successfully, LaunchPal SHALL honor the recorded intent by tearing down the just-launched helper and transitioning to `Disabled` rather than `Enabled`.

The pending-disable intent SHALL be reset (cleared) only when a fresh Enable actually begins a request — i.e. when Enable transitions `Disabled` → `Requesting`. An Enable call that returns early as a no-op because the state is already `Requesting` or `Enabled` SHALL NOT clear the intent. This makes the multi-click outcome deterministic: while a single authorization prompt is in flight, a Disable click followed by further Enable clicks (which no-op) resolves to `Disabled`, because only a brand-new request cycle clears a pending disable.

#### Scenario: Rapid multi-click during one authorization prompt

- **WHEN** the state is `Requesting` and the user clicks Enable (no-op), Disable (records intent), then Enable again (no-op) before completing authorization, then authorizes successfully
- **THEN** the pending-disable intent remains set (the no-op Enables did not clear it) and the final state is `Disabled`

##### Example: intent resolution during one Requesting window

| Sequence during Requesting | disableRequested at handshake success | Final state |
| -------------------------- | ------------------------------------- | ----------- |
| (no Disable)               | false                                 | Enabled     |
| Disable                    | true                                  | Disabled    |
| Disable, Enable (no-op)    | true                                  | Disabled    |

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

#### Scenario: Disable requested during Requesting is honored

- **WHEN** user clicks Enable, then clicks Disable while the authorization prompt is showing, then completes authorization successfully
- **THEN** LaunchPal tears down the just-launched helper and the final state is `Disabled` with no surviving helper process or socket

### Requirement: Disable shuts down helper gracefully

`DisableAdminMode()` SHALL send a `Shutdown` RPC to the helper, wait up to 3 seconds for acknowledgment, then close the connection. After shutdown, state SHALL return to `Disabled`. Additionally, when an Enable attempt fails AFTER osascript authorization has already succeeded and AFTER a connection to the helper was established but the `Ping` handshake failed, LaunchPal SHALL send a best-effort `Shutdown` RPC over that established connection before closing it, rather than only closing the connection. When the connection itself was never established (the `Connect` step failed), there is no channel on which to send `Shutdown`; in that case the launched helper SHALL be left to its own idle self-termination and parent-PID watchdog as the backstop. If a best-effort Shutdown is attempted but cannot be delivered, the same backstops apply.

#### Scenario: Clean disable

- **WHEN** user clicks Disable
- **THEN** state goes Enabled → ShuttingDown → Disabled, and the socket file is removed

#### Scenario: Helper unresponsive on shutdown

- **WHEN** Shutdown RPC does not respond within 3 seconds
- **THEN** LaunchPal closes the connection anyway and returns to Disabled; the helper's disconnect/idle self-termination ensures eventual cleanup

#### Scenario: Enable fails after authorization with a connection established

- **WHEN** osascript authorization succeeds and a connection to the helper is established but the `Ping` handshake fails
- **THEN** LaunchPal sends a best-effort `Shutdown` over that connection, then closes it, and returns to Disabled with error `helper_handshake_failed`

#### Scenario: Enable fails before a connection is established

- **WHEN** osascript authorization succeeds but the client never connects (the `Connect` step fails)
- **THEN** LaunchPal returns to Disabled with error `helper_handshake_failed` without attempting a Shutdown, and the launched helper is reaped by its idle timeout and parent-PID watchdog

### Requirement: Helper crash detection and recovery

LaunchPal SHALL detect an ended helper connection by observing EOF or connection errors on the RPC connection while Admin Mode is `Enabled`. Because the helper now self-terminates on its idle timeout by design, and a clean self-termination is observed by the GUI as the same EOF/connection error as an actual crash, LaunchPal SHALL NOT surface every such disconnect as a red `helper_crashed` error. Upon detecting a connection end while `Enabled`, Admin Mode SHALL transition to `Disabled` with a neutral session-ended status (reason `admin_session_ended`, presented as an informational "Admin Mode session ended — re-enable to continue" message, not an error), and the frontend SHALL be notified via event or the next `GetAdminModeStatus()` poll. The write controls SHALL be hidden again as for any `Disabled` state.

#### Scenario: Helper connection ends while Enabled

- **WHEN** the helper exits (idle self-termination, clean teardown, or an actual crash) while Admin Mode is `Enabled`, and the next RPC or the connection watcher observes the closed connection
- **THEN** state becomes `Disabled` with the neutral `admin_session_ended` status, the UI shows an informational re-enable prompt rather than a crash error, and system-daemon write controls are hidden
