## ADDED Requirements

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

## MODIFIED Requirements

### Requirement: Supported RPC methods

The helper SHALL implement the following methods: `Ping`, `ListSystemDaemons`, `GetSystemDaemon`, `Bootstrap`, `Bootout`, `Kickstart`, `WritePlist`, `DeletePlist`, `EnsureLogAccess`, `TruncateLog`, `Shutdown`. Unknown methods SHALL return `{"error": {"code": "unknown_method", "message": "..."}}`.

#### Scenario: Unknown method

- **WHEN** a client sends `{"id": 1, "method": "DoesNotExist"}`
- **THEN** the helper returns `{"id": 1, "error": {"code": "unknown_method", "message": "..."}}`

#### Scenario: TruncateLog is reachable

- **WHEN** a client sends a valid `{"id": N, "method": "TruncateLog", "params": {"path": "<allowed path>"}}`
- **THEN** the helper dispatches to the TruncateLog handler and returns either `OKResult` or an error response, but never `unknown_method`
