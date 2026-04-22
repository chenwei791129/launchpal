## MODIFIED Requirements

### Requirement: Read-only managers reject write operations

`AppleSystemManager` SHALL return `ErrReadOnlyManager` for all write operations (Start, Stop, Restart, Create, Update, Delete) regardless of any elevated mode.

`SystemManager` SHALL return `ErrReadOnlyManager` for write operations when Admin Mode is Disabled. When Admin Mode is Enabled, `SystemManager` write operations SHALL be delegated to the privileged helper via RPC (see `system-daemon-write-ops` capability).

#### Scenario: SystemManager write operations with Admin Mode Disabled

- **WHEN** Start, Stop, Restart, Create, Update, or Delete is called on SystemManager while Admin Mode is Disabled
- **THEN** each call returns `ErrReadOnlyManager`

#### Scenario: SystemManager write operations with Admin Mode Enabled

- **WHEN** Start is called on SystemManager while Admin Mode is Enabled
- **THEN** the call delegates to the helper's `Bootstrap` (or equivalent) RPC and returns the helper's result

#### Scenario: AppleSystemManager write operations always rejected

- **WHEN** Start, Stop, Restart, Create, Update, or Delete is called on AppleSystemManager regardless of Admin Mode
- **THEN** each call returns `ErrReadOnlyManager`
