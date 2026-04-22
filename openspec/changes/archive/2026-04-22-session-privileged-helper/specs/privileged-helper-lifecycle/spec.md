## ADDED Requirements

### Requirement: Helper binary packaged in app bundle

The `launchpal-privhelper` binary SHALL be compiled and placed at `LaunchPal.app/Contents/MacOS/launchpal-privhelper` during the build. LaunchPal SHALL locate the helper at runtime using `os.Executable()` to find the main binary directory, then joining `launchpal-privhelper`.

#### Scenario: Helper binary resolved at runtime

- **WHEN** LaunchPal enables Admin Mode
- **THEN** the helper path is computed as `<main-binary-dir>/launchpal-privhelper` and exists at that path

#### Scenario: Helper binary missing

- **WHEN** the helper binary is not found at the expected path
- **THEN** Admin Mode enablement fails with an error identifying the missing binary path

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

### Requirement: Helper launched via osascript with background execution

LaunchPal SHALL launch the helper by executing `osascript -e 'do shell script "<helper-path> --socket <path> --parent-pid <pid> --launching-uid <uid> &> /dev/null &" with administrator privileges'`. The `&` trailing operator ensures the helper continues running after `do shell script` returns.

#### Scenario: osascript authorization granted

- **WHEN** the user authorizes the osascript password/Touch ID prompt
- **THEN** `do shell script` returns successfully and the helper process starts as root

#### Scenario: osascript authorization cancelled

- **WHEN** the user cancels or fails the osascript prompt
- **THEN** LaunchPal receives an error indicating authorization was cancelled, and Admin Mode returns to Disabled without further action

### Requirement: Socket handshake with retry

After launching the helper, LaunchPal SHALL repeatedly attempt to connect to the socket path with exponential backoff until either: (a) connection succeeds and a `Ping` RPC returns a successful response, or (b) 10 seconds elapse, at which point Admin Mode enablement fails.

#### Scenario: Handshake succeeds within timeout

- **WHEN** LaunchPal connects to the socket and Ping returns ok within 10 seconds
- **THEN** Admin Mode transitions to Enabled

#### Scenario: Handshake times out

- **WHEN** LaunchPal cannot connect or Ping fails within 10 seconds
- **THEN** Admin Mode returns to Disabled with an error containing "helper handshake failed"

### Requirement: Parent PID watchdog

The helper SHALL spawn a background goroutine that checks the parent LaunchPal process every second using `syscall.Kill(parentPID, 0)`. If the parent is no longer running, the helper SHALL remove the socket file and exit within 2 seconds.

#### Scenario: Parent exits normally

- **WHEN** LaunchPal exits without sending Shutdown
- **THEN** the helper detects the parent is gone, removes the socket, and exits within 2 seconds

#### Scenario: Parent is killed

- **WHEN** LaunchPal is killed via SIGKILL
- **THEN** the helper detects this within 1-2 seconds and self-exits

### Requirement: Graceful shutdown via RPC

The helper SHALL accept a `Shutdown` RPC method. Upon receiving it, the helper SHALL acknowledge the request, close the listener, remove the socket file, and exit with code `0`.

#### Scenario: Client requests shutdown

- **WHEN** LaunchPal sends Shutdown RPC
- **THEN** helper responds with ok, removes the socket, and exits cleanly

### Requirement: Idle timeout

The helper SHALL track the time of the last successful RPC. If no RPC is received for 30 minutes, the helper SHALL remove the socket and exit.

#### Scenario: Extended idle period

- **WHEN** no RPC traffic occurs for 30 minutes
- **THEN** helper cleans up and self-exits, and any subsequent client connection attempt fails
