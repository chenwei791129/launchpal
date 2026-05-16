## MODIFIED Requirements

### Requirement: Read-only managers reject write operations

SystemManager SHALL return `ErrReadOnlyManager` for all write operations when Admin Mode is disabled: Start, Stop, Restart, Create, Update, Delete.
When Admin Mode is enabled, SystemManager SHALL delegate write operations to the privileged helper via RPC instead of returning `ErrReadOnlyManager`.
AppleSystemManager SHALL return `ErrReadOnlyManager` for all write operations unconditionally.

SystemManager SHALL additionally expose `DeleteWithOptions(name string, opts DeleteServiceOptions) error` for callers that need the optional log-cleanup flow. This method SHALL NOT be part of the `Manager` interface.

#### Scenario: SystemManager write operations without Admin Mode

- **WHEN** Start, Stop, Restart, Create, Update, or Delete is called on SystemManager and Admin Mode is disabled
- **THEN** each call returns `ErrReadOnlyManager`

#### Scenario: AppleSystemManager write operations

- **WHEN** Start, Stop, Restart, Create, Update, or Delete is called on AppleSystemManager
- **THEN** each call returns `ErrReadOnlyManager`

#### Scenario: SystemManager DeleteWithOptions without log cleanup

- **WHEN** `DeleteWithOptions` is called with `DeleteLogs: false`
- **THEN** the behavior is identical to `Delete`: bootout + DeletePlist RPC, no log files touched
