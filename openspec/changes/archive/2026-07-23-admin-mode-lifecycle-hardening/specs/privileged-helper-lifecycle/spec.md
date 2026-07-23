## ADDED Requirements

### Requirement: Helper self-terminates on client disconnect

The helper serves a single client connection. WHEN that client connection ends for any reason — the connection handler returning on EOF, on a read error, or on a write/encode error — the helper SHALL remove the socket file and exit, ending the accept loop, rather than continuing to listen until the idle timeout or parent watchdog fires. This teardown SHALL be triggered on every connection-handler exit path that indicates the connection ended (not only the post-scan EOF path), and SHALL NOT be triggered when the server is already stopping for another reason. This is the primary teardown mechanism, because the unprivileged GUI cannot signal the root helper directly.

#### Scenario: Client disconnects (EOF)

- **WHEN** the LaunchPal client connection to the helper closes cleanly (Disable, GUI exit, GUI crash, or a transient drop) and the connection handler returns on EOF
- **THEN** the helper removes the socket and exits within a few seconds, and any subsequent connection attempt to the socket fails

#### Scenario: Disconnect surfaced via a failed write

- **WHEN** the connection dies while the helper is writing a response, so the connection handler returns from a failed write/encode rather than from the EOF path
- **THEN** the helper still removes the socket and exits within a few seconds — the failed-write return path is not exempt from teardown

## MODIFIED Requirements

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

### Requirement: Idle timeout

The helper SHALL track the time of the last successful RPC. If no RPC is received for 5 minutes, the helper SHALL remove the socket and exit. Any successful RPC resets the idle timer, so an actively used session is unaffected; the timeout bounds only the window during which an idle-but-still-connected session keeps a root socket alive. Because the GUI holds a single long-lived connection that stays open while idle, the idle-driven stop SHALL close the active accepted connection (not only the listener) so that the connection handler unblocks and the helper process actually exits; closing the listener alone would leave a connected-but-idle helper running.

#### Scenario: Extended idle period with the GUI still connected

- **WHEN** no RPC traffic occurs for 5 minutes while the GUI connection is still open
- **THEN** the helper closes the active connection, cleans up, and self-exits (the process terminates, not merely the listener), and any subsequent client connection attempt fails

#### Scenario: Activity resets the idle timer

- **WHEN** RPCs continue to arrive at intervals shorter than 5 minutes
- **THEN** the helper does not self-exit on idle and remains available
