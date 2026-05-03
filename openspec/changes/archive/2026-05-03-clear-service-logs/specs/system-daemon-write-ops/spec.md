## ADDED Requirements

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

### Requirement: Re-validate authorization at clear time

The system SHALL NOT cache or trust any prior `GetLogClearStatus` result when executing `ClearLogs`. Each call SHALL perform its own OpenFile attempt and Admin Mode check at the time of execution.

#### Scenario: Admin Mode disables between status query and clear

- **WHEN** `GetLogClearStatus` returns `userWritable=false` for a system log, the user confirms in the UI, and Admin Mode auto-disables (idle timeout) before the `ClearSystemLogs` call lands
- **THEN** `ClearLogs` re-checks the helper connection, finds it absent, and returns an error indicating Admin Mode is required
