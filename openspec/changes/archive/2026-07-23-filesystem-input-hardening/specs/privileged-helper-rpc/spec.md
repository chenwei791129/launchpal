## ADDED Requirements

### Requirement: Symlink-safe resolution of log-path arguments

The helper's log-path operations — `EnsureLogAccess` (which creates and chmods the parent directory), `TruncateLog`, `DeleteLogPaths`, and the pre-write backup path handling in `backupExisting` — SHALL resolve or open EVERY component of the target path in a symlink-safe manner, not only the final component. Because the log-path allowlist includes world-writable directories (`/tmp/`, `/private/tmp/`) where a same-UID process can plant a symlink at an intermediate directory, lexical validation (`filepath.Clean` + prefix match) together with a leaf-only `O_NOFOLLOW`/`Lstat` guard SHALL NOT be relied upon. The helper SHALL open each path component with `O_NOFOLLOW` (openat-style traversal, e.g. opening the parent directory with `O_DIRECTORY|O_NOFOLLOW` and operating on the leaf via `*at` calls with `AT_SYMLINK_NOFOLLOW`) so that a symlink at any component causes the operation to fail or to remain confined to an allowlisted real path. Where `EnsureLogAccess` must create missing intermediate directories (today via `os.MkdirAll`), the components SHALL be created symlink-safely as well (e.g. `Mkdirat` relative to the verified parent fd, per component), so directory creation does not reintroduce the intermediate-symlink escape. The lexical `validateLogPath` check MAY remain as a fast pre-filter but is not the enforcement boundary.

#### Scenario: Intermediate-directory symlink does not escape the allowlist

- **WHEN** a log-path argument routes through a symlinked intermediate directory that points outside the allowlist (for example an allowlisted `/tmp/<link>/file` where `<link>` is a symlink to `/etc`)
- **THEN** the helper does not chmod, truncate, or delete the real target outside the allowlist — the operation is rejected rather than following the symlink

#### Scenario: Legitimate nested log path still works

- **WHEN** the log path is a genuine (non-symlinked) file at least one sub-directory deep under an allowlisted prefix
- **THEN** the operation proceeds normally
