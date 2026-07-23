## ADDED Requirements

### Requirement: Routing name path-traversal confinement

The user, system, and read-only service managers SHALL reject any routing `name` (and `CreateService`'s `config.Label`) that is not a single path component before it is joined into a plist path. A value containing a path separator or a NUL byte, or equal to `.` or `..`, SHALL be rejected with a validation error and SHALL NOT reach any file operation. A single component that merely contains `..` as a substring (e.g. `com.example..worker`) is NOT rejected: with no path separator it cannot traverse out of the base directory, and it is a legal launchd label that must remain manageable. This confines all name-derived operations (get, read plist, read logs, create, update, delete, clear logs, and the system-domain start/stop/restart) to the intended base directory (`~/Library/LaunchAgents` for user services, `/Library/LaunchDaemons` for system daemons, or the read-only system directories), since `filepath.Join` alone does not confine `..` to the base directory. This is a GUI-side defense in depth; for system-domain writes the privileged helper independently re-validates the path, but the manager SHALL NOT rely on that alone.

#### Scenario: Traversal name is rejected (user and system domains)

- **WHEN** a binding is invoked with `name` set to `../../etc/passwd` (or `config.Label` containing `..`/`/`), in either the user domain or the system domain
- **THEN** the manager returns a validation error and performs no file read, write, or delete outside the base directory

#### Scenario: Normal service name is accepted

- **WHEN** `name` is a plain label such as `com.example.foo`
- **THEN** the operation proceeds normally
