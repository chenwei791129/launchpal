## ADDED Requirements

### Requirement: Newline-delimited JSON RPC transport

Communication between LaunchPal and the helper SHALL use newline-delimited JSON messages over a Unix domain socket. Each message SHALL be a single line terminated by `\n`. Requests SHALL include `id` (monotonic integer), `method`, and optional `params`. Responses SHALL include the matching `id` and exactly one of `result` or `error`.

#### Scenario: Request-response correlation

- **WHEN** a client sends `{"id": 5, "method": "Ping"}`
- **THEN** the helper replies with `{"id": 5, "result": {"pong": true}}` on a single line terminated by `\n`

#### Scenario: Malformed JSON

- **WHEN** the helper receives a line that is not valid JSON
- **THEN** it returns `{"id": null, "error": {"code": "invalid_request", "message": "..."}}`

### Requirement: Socket path and permissions

The socket path SHALL be `$TMPDIR/launchpal-<launching-uid>-<16-hex-random>.sock`. The helper SHALL create the socket with mode `0600`. On shutdown or exit, the helper SHALL remove the socket file.

#### Scenario: Socket permissions after creation

- **WHEN** the helper creates the socket
- **THEN** the file mode is `0600` and the owner is `root`

#### Scenario: Socket path is per-user private

- **WHEN** the socket path is generated
- **THEN** it resides under `$TMPDIR` (which is per-user on macOS) with a random 16-hex-character suffix

### Requirement: Peer UID verification

For every new client connection, the helper SHALL query the peer UID via `LOCAL_PEERCRED` and reject the connection if it does not match the `--launching-uid` provided at startup. Rejected connections SHALL be closed immediately without reading any input.

#### Scenario: Matching peer UID

- **WHEN** a client whose UID equals `--launching-uid` connects
- **THEN** the connection is accepted and RPC processing begins

#### Scenario: Mismatched peer UID

- **WHEN** a client with a different UID connects
- **THEN** the helper closes the connection without processing any RPC

### Requirement: Supported RPC methods

The helper SHALL implement the following methods: `Ping`, `ListSystemDaemons`, `GetSystemDaemon`, `Bootstrap`, `Bootout`, `Kickstart`, `WritePlist`, `DeletePlist`, `Shutdown`. Unknown methods SHALL return `{"error": {"code": "unknown_method", "message": "..."}}`.

#### Scenario: Unknown method

- **WHEN** a client sends `{"id": 1, "method": "DoesNotExist"}`
- **THEN** the helper returns `{"id": 1, "error": {"code": "unknown_method", "message": "..."}}`

### Requirement: Serial request processing

The helper SHALL process RPC requests serially on a single connection. A new request SHALL NOT begin processing until the previous response has been written.

#### Scenario: Sequential processing

- **WHEN** a client pipelines three requests on one connection
- **THEN** the helper processes and responds to them in order

### Requirement: Error code taxonomy

RPC error responses SHALL use codes from a fixed set: `invalid_request`, `unknown_method`, `invalid_params`, `permission_denied`, `not_found`, `launchctl_failed`, `io_error`, `internal_error`. Each error response SHALL include a human-readable `message`.

#### Scenario: launchctl failure

- **WHEN** `Bootstrap` is called and `launchctl bootstrap system <path>` exits non-zero
- **THEN** the RPC returns `{"error": {"code": "launchctl_failed", "message": "<stderr>"}}`
