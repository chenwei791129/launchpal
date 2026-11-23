# Privileged helper protocol

Admin Mode uses newline-delimited JSON over a Unix domain socket. The helper restricts socket permissions to `0600` and verifies the connecting UID before processing requests. Requests are handled serially within each connection, and replies are sent in request order. The request scanner is bounded to 4 MiB per line.

The [protocol types](../internal/privhelper/protocol.go) are the authoritative method, parameter, and error-code definitions. Requests carry an integer `id`, a `method`, and optional `params`. A reply echoes the `id` and carries either `result` or `error`; malformed JSON receives `invalid_request` with a null ID. A successful `DeleteLogPaths` reply can include individual path failures in its result.

The helper validates service and log paths independently of the GUI. Log operations reject symlinks in the path below the allowed root on macOS; truncation never creates an absent file. The session ends when the primary client disconnects, on explicit shutdown, or when the watchdog or idle timeout stops the helper. See the [engineering guide](../.claude/CLAUDE.md#admin-mode-session-scoped-privileged-helper) for the full lifecycle and write boundaries.
