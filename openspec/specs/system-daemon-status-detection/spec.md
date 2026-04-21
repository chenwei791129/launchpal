### Requirement: Heuristic status detection for system daemons

SystemManager and AppleSystemManager SHALL detect a service's runtime status and PID without requiring elevated privileges, using the following algorithm:

1. Read `UserName` from the service plist. If absent, treat it as `root`.
2. Determine the program path: use `Program` if non-empty, otherwise use `ProgramArguments[0]`.
3. If the program path is empty, return `StatusUnknown` with PID `0` and confidence `unverified`.
4. If the program path matches one of the shell programs listed in the "Shell skip list" requirement, return `StatusLoaded` with PID `0` and confidence `verified`.
5. Resolve `UserName` to a numeric UID via the system user database.
6. Obtain a process table snapshot containing, for every running process, its `UID`, `PPID`, and full argument string (e.g. via a single `ps -axo uid=,pid=,ppid=,args=` invocation).
7. From the process table, select PIDs where all of the following hold: `UID` equals the resolved UID, `PPID` equals `1` (launchd), and the process's argument string contains the program path as a substring. Sort the selected PIDs in ascending numeric order.
8. Return the following based on the selected PID count:
   - Exactly one PID: `StatusRunning`, that PID, confidence `verified`
   - Zero PIDs: `StatusStopped`, PID `0`, confidence `verified`
   - More than one PID: `StatusRunning`, the first (lowest) PID, confidence `unverified`

If the UID lookup fails or the process table cannot be obtained, the detection SHALL return `StatusStopped` with PID `0` and confidence `unverified`. In the new algorithm candidate PIDs are only known after scanning the process table, so when either prerequisite is missing no candidates exist; Stopped with an unverified flag signals the uncertain state without fabricating a PID.

The `SystemManager.List` and `AppleSystemManager.List` operations SHALL obtain exactly one process table snapshot per call and share it across all per-service detection calls.

#### Scenario: Single matching PID with launchd parent

- **WHEN** a daemon's program has exactly one running process owned by `UserName` with parent PID 1 and an argument string containing the program path
- **THEN** the Service reports `StatusRunning`, the matching PID, and `StatusConfidence = "verified"`

#### Scenario: No matching process

- **WHEN** no entry in the process table matches the UID, PPID, and program path criteria
- **THEN** the Service reports `StatusStopped`, PID `0`, and `StatusConfidence = "verified"`

#### Scenario: Multiple candidates with launchd parent

- **WHEN** two or more processes match and all have parent PID 1
- **THEN** the Service reports `StatusRunning` with the lowest-numbered PID and `StatusConfidence = "unverified"`

#### Scenario: User-launched process with non-launchd parent is filtered out

- **WHEN** the process table contains a process owned by the target user whose argument string contains the program path but whose parent PID is not 1
- **THEN** the detection discards that PID and the Service reports `StatusStopped` with confidence `verified`

#### Scenario: UserName defaults to root when absent

- **WHEN** the service plist has no `UserName` key
- **THEN** the detection resolves the UID of `root` and filters the process table by that UID

#### Scenario: Non-root UserName is honored

- **WHEN** the service plist has `UserName = _www`
- **THEN** the detection resolves the UID of `_www` and does not match root-owned processes

#### Scenario: Empty program path

- **WHEN** the service plist has neither `Program` nor `ProgramArguments`
- **THEN** the Service reports `StatusUnknown`, PID `0`, and `StatusConfidence = "unverified"`

#### Scenario: Process table snapshot shared across a single List call

- **WHEN** `SystemManager.List` or `AppleSystemManager.List` is invoked
- **THEN** exactly one process table snapshot is obtained for that call, and every per-service detection reads from the same snapshot

#### Scenario: Process table fetch failure

- **WHEN** the process table cannot be obtained (e.g., the `ps` invocation returns an error)
- **THEN** the Service reports `StatusStopped` with PID `0` and `StatusConfidence = "unverified"` (not confident Stopped)

#### Scenario: UID lookup failure

- **WHEN** resolving `UserName` to a numeric UID fails (e.g., the user database has no such entry)
- **THEN** the Service reports `StatusStopped` with PID `0` and `StatusConfidence = "unverified"`


<!-- @trace
source: system-daemon-status-detection
updated: 2026-04-21
code:
  - internal/launchctl/user.go
  - README.md
  - internal/launchctl/readonly.go
  - frontend/app/pages/services/[name].vue
  - frontend/wailsjs/go/models.ts
  - internal/launchctl/status_detect.go
  - frontend/app/types/wails.d.ts
  - internal/launchctl/types.go
  - frontend/app/components/ServiceRow.vue
  - frontend/app/components/StatusConfidenceIcon.vue
tests:
  - internal/launchctl/readonly_test.go
  - internal/launchctl/status_detect_test.go
-->


<!-- @trace
source: improve-system-service-list-performance
updated: 2026-04-22
code:
  - internal/launchctl/readonly.go
  - internal/launchctl/status_detect.go
tests:
  - internal/launchctl/readonly_test.go
  - internal/launchctl/status_detect_test.go
-->

### Requirement: Shell skip list

The detection logic SHALL skip process-table matching when the program path equals one of: `/bin/bash`, `/bin/sh`, `/bin/zsh`, `/usr/bin/bash`, `/usr/bin/sh`, `/usr/bin/zsh`. For these programs the service SHALL be reported as `StatusLoaded`, PID `0`, confidence `verified`.

#### Scenario: Daemon invokes /bin/bash as Program

- **WHEN** the plist has `Program = /bin/bash`
- **THEN** the detection skips process-table matching and reports `StatusLoaded` with PID `0` and confidence `verified`


<!-- @trace
source: system-daemon-status-detection
updated: 2026-04-21
code:
  - internal/launchctl/user.go
  - README.md
  - internal/launchctl/readonly.go
  - frontend/app/pages/services/[name].vue
  - frontend/wailsjs/go/models.ts
  - internal/launchctl/status_detect.go
  - frontend/app/types/wails.d.ts
  - internal/launchctl/types.go
  - frontend/app/components/ServiceRow.vue
  - frontend/app/components/StatusConfidenceIcon.vue
tests:
  - internal/launchctl/readonly_test.go
  - internal/launchctl/status_detect_test.go
-->


<!-- @trace
source: improve-system-service-list-performance
updated: 2026-04-22
code:
  - internal/launchctl/readonly.go
  - internal/launchctl/status_detect.go
tests:
  - internal/launchctl/readonly_test.go
  - internal/launchctl/status_detect_test.go
-->

### Requirement: StatusConfidence field on Service

The `Service` struct SHALL include a `StatusConfidence` field of type string with values `verified` or `unverified`. UserManager SHALL set this field to `verified` for all its services. SystemManager and AppleSystemManager SHALL set this field according to the heuristic detection outcome.

#### Scenario: User service confidence

- **WHEN** a service is returned by UserManager.List or UserManager.Get
- **THEN** `StatusConfidence = "verified"`

#### Scenario: System service ambiguous detection

- **WHEN** SystemManager detection yields more than one matching PID with parent PID 1
- **THEN** the returned Service has `StatusConfidence = "unverified"`


<!-- @trace
source: system-daemon-status-detection
updated: 2026-04-21
code:
  - internal/launchctl/user.go
  - README.md
  - internal/launchctl/readonly.go
  - frontend/app/pages/services/[name].vue
  - frontend/wailsjs/go/models.ts
  - internal/launchctl/status_detect.go
  - frontend/app/types/wails.d.ts
  - internal/launchctl/types.go
  - frontend/app/components/ServiceRow.vue
  - frontend/app/components/StatusConfidenceIcon.vue
tests:
  - internal/launchctl/readonly_test.go
  - internal/launchctl/status_detect_test.go
-->

### Requirement: Status detection replaces Stopped fallback

When the batch `launchctl list` map does not contain a label for a system or apple-system service, the manager SHALL invoke heuristic status detection rather than defaulting to `StatusStopped`.

#### Scenario: Batch map miss triggers heuristic

- **WHEN** `getBatchServiceStatus()` does not contain the service label and the manager is SystemManager or AppleSystemManager
- **THEN** the manager calls heuristic detection to populate Status, PID, and StatusConfidence

#### Scenario: Batch map hit takes precedence

- **WHEN** `getBatchServiceStatus()` contains the service label with a PID
- **THEN** the manager uses the batch result and sets `StatusConfidence = "verified"` without invoking heuristic detection

## Requirements

<!-- @trace
source: system-daemon-status-detection
updated: 2026-04-21
code:
  - internal/launchctl/user.go
  - README.md
  - internal/launchctl/readonly.go
  - frontend/app/pages/services/[name].vue
  - frontend/wailsjs/go/models.ts
  - internal/launchctl/status_detect.go
  - frontend/app/types/wails.d.ts
  - internal/launchctl/types.go
  - frontend/app/components/ServiceRow.vue
  - frontend/app/components/StatusConfidenceIcon.vue
tests:
  - internal/launchctl/readonly_test.go
  - internal/launchctl/status_detect_test.go
-->

### Requirement: Heuristic status detection for system daemons

SystemManager and AppleSystemManager SHALL detect a service's runtime status and PID without requiring elevated privileges, using the following algorithm:

1. Read `UserName` from the service plist. If absent, treat it as `root`.
2. Determine the program path: use `Program` if non-empty, otherwise use `ProgramArguments[0]`.
3. If the program path is empty, return `StatusUnknown` with PID `0` and confidence `unverified`.
4. If the program path matches one of the shell programs listed in the "Shell skip list" requirement, return `StatusLoaded` with PID `0` and confidence `verified`.
5. Resolve `UserName` to a numeric UID via the system user database.
6. Obtain a process table snapshot containing, for every running process, its `UID`, `PPID`, and full argument string (e.g. via a single `ps -axo uid=,pid=,ppid=,args=` invocation).
7. From the process table, select PIDs where all of the following hold: `UID` equals the resolved UID, `PPID` equals `1` (launchd), and the process's argument string contains the program path as a substring. Sort the selected PIDs in ascending numeric order.
8. Return the following based on the selected PID count:
   - Exactly one PID: `StatusRunning`, that PID, confidence `verified`
   - Zero PIDs: `StatusStopped`, PID `0`, confidence `verified`
   - More than one PID: `StatusRunning`, the first (lowest) PID, confidence `unverified`

If the UID lookup fails or the process table cannot be obtained, the detection SHALL return `StatusStopped` with PID `0` and confidence `unverified`. In the new algorithm candidate PIDs are only known after scanning the process table, so when either prerequisite is missing no candidates exist; Stopped with an unverified flag signals the uncertain state without fabricating a PID.

The `SystemManager.List` and `AppleSystemManager.List` operations SHALL obtain exactly one process table snapshot per call and share it across all per-service detection calls.

#### Scenario: Single matching PID with launchd parent

- **WHEN** a daemon's program has exactly one running process owned by `UserName` with parent PID 1 and an argument string containing the program path
- **THEN** the Service reports `StatusRunning`, the matching PID, and `StatusConfidence = "verified"`

#### Scenario: No matching process

- **WHEN** no entry in the process table matches the UID, PPID, and program path criteria
- **THEN** the Service reports `StatusStopped`, PID `0`, and `StatusConfidence = "verified"`

#### Scenario: Multiple candidates with launchd parent

- **WHEN** two or more processes match and all have parent PID 1
- **THEN** the Service reports `StatusRunning` with the lowest-numbered PID and `StatusConfidence = "unverified"`

#### Scenario: User-launched process with non-launchd parent is filtered out

- **WHEN** the process table contains a process owned by the target user whose argument string contains the program path but whose parent PID is not 1
- **THEN** the detection discards that PID and the Service reports `StatusStopped` with confidence `verified`

#### Scenario: UserName defaults to root when absent

- **WHEN** the service plist has no `UserName` key
- **THEN** the detection resolves the UID of `root` and filters the process table by that UID

#### Scenario: Non-root UserName is honored

- **WHEN** the service plist has `UserName = _www`
- **THEN** the detection resolves the UID of `_www` and does not match root-owned processes

#### Scenario: Empty program path

- **WHEN** the service plist has neither `Program` nor `ProgramArguments`
- **THEN** the Service reports `StatusUnknown`, PID `0`, and `StatusConfidence = "unverified"`

#### Scenario: Process table snapshot shared across a single List call

- **WHEN** `SystemManager.List` or `AppleSystemManager.List` is invoked
- **THEN** exactly one process table snapshot is obtained for that call, and every per-service detection reads from the same snapshot

#### Scenario: Process table fetch failure

- **WHEN** the process table cannot be obtained (e.g., the `ps` invocation returns an error)
- **THEN** the Service reports `StatusStopped` with PID `0` and `StatusConfidence = "unverified"` (not confident Stopped)

#### Scenario: UID lookup failure

- **WHEN** resolving `UserName` to a numeric UID fails (e.g., the user database has no such entry)
- **THEN** the Service reports `StatusStopped` with PID `0` and `StatusConfidence = "unverified"`

---
### Requirement: Shell skip list

The detection logic SHALL skip process-table matching when the program path equals one of: `/bin/bash`, `/bin/sh`, `/bin/zsh`, `/usr/bin/bash`, `/usr/bin/sh`, `/usr/bin/zsh`. For these programs the service SHALL be reported as `StatusLoaded`, PID `0`, confidence `verified`.

#### Scenario: Daemon invokes /bin/bash as Program

- **WHEN** the plist has `Program = /bin/bash`
- **THEN** the detection skips process-table matching and reports `StatusLoaded` with PID `0` and confidence `verified`

---
### Requirement: StatusConfidence field on Service

The `Service` struct SHALL include a `StatusConfidence` field of type string with values `verified` or `unverified`. UserManager SHALL set this field to `verified` for all its services. SystemManager and AppleSystemManager SHALL set this field according to the heuristic detection outcome.

#### Scenario: User service confidence

- **WHEN** a service is returned by UserManager.List or UserManager.Get
- **THEN** `StatusConfidence = "verified"`

#### Scenario: System service ambiguous detection

- **WHEN** SystemManager detection yields more than one matching PID with parent PID 1
- **THEN** the returned Service has `StatusConfidence = "unverified"`

---
### Requirement: Status detection replaces Stopped fallback

When the batch `launchctl list` map does not contain a label for a system or apple-system service, the manager SHALL invoke heuristic status detection rather than defaulting to `StatusStopped`.

#### Scenario: Batch map miss triggers heuristic

- **WHEN** `getBatchServiceStatus()` does not contain the service label and the manager is SystemManager or AppleSystemManager
- **THEN** the manager calls heuristic detection to populate Status, PID, and StatusConfidence

#### Scenario: Batch map hit takes precedence

- **WHEN** `getBatchServiceStatus()` contains the service label with a PID
- **THEN** the manager uses the batch result and sets `StatusConfidence = "verified"` without invoking heuristic detection