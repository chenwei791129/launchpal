### Requirement: Heuristic status detection for system daemons

SystemManager and AppleSystemManager SHALL detect a service's runtime status and PID without requiring elevated privileges, using the following algorithm:

1. Read `UserName` from the service plist. If absent, treat it as `root`.
2. Determine the program path: use `Program` if non-empty, otherwise use `ProgramArguments[0]`.
3. If the program path is empty, return `StatusUnknown` with PID `0` and confidence `unverified`.
4. If the program path matches one of the shell programs listed in the "Shell skip list" requirement, return `StatusLoaded` with PID `0` and confidence `verified`.
5. Execute `pgrep -u <UserName> -f <program>` to obtain candidate PIDs.
6. For each candidate PID, execute `ps -o ppid= -p <pid>` and retain only PIDs whose parent PID equals `1` (launchd).
7. Return the following based on filtered candidate count:
   - Exactly one PID: `StatusRunning`, that PID, confidence `verified`
   - Zero PIDs: `StatusStopped`, PID `0`, confidence `verified`
   - More than one PID: `StatusRunning`, the first PID, confidence `unverified`

#### Scenario: Single matching PID with launchd parent

- **WHEN** a daemon's program has exactly one running process owned by `UserName` with parent PID 1
- **THEN** the Service reports `StatusRunning`, the matching PID, and `StatusConfidence = "verified"`

#### Scenario: No matching process

- **WHEN** `pgrep -u <UserName> -f <program>` returns no output
- **THEN** the Service reports `StatusStopped`, PID `0`, and `StatusConfidence = "verified"`

#### Scenario: Multiple candidates with launchd parent

- **WHEN** two or more processes match and both have parent PID 1
- **THEN** the Service reports `StatusRunning` with the first PID and `StatusConfidence = "unverified"`

#### Scenario: User-launched process with non-launchd parent is filtered out

- **WHEN** `pgrep` finds a process owned by the target user but its parent PID is a shell PID (not 1)
- **THEN** the detection discards that PID and the Service reports `StatusStopped` with confidence `verified`

#### Scenario: UserName defaults to root when absent

- **WHEN** the service plist has no `UserName` key
- **THEN** the detection uses `root` as the user filter for `pgrep -u`

#### Scenario: Non-root UserName is honored

- **WHEN** the service plist has `UserName = _www`
- **THEN** the detection uses `_www` as the user filter and does not match root-owned processes

#### Scenario: Empty program path

- **WHEN** the service plist has neither `Program` nor `ProgramArguments`
- **THEN** the Service reports `StatusUnknown`, PID `0`, and `StatusConfidence = "unverified"`


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

### Requirement: Shell skip list

The detection logic SHALL skip `pgrep` matching when the program path equals one of: `/bin/bash`, `/bin/sh`, `/bin/zsh`, `/usr/bin/bash`, `/usr/bin/sh`, `/usr/bin/zsh`. For these programs the service SHALL be reported as `StatusLoaded`, PID `0`, confidence `verified`.

#### Scenario: Daemon invokes /bin/bash as Program

- **WHEN** the plist has `Program = /bin/bash`
- **THEN** the detection skips pgrep and reports `StatusLoaded` with PID `0` and confidence `verified`


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
5. Execute `pgrep -u <UserName> -f <program>` to obtain candidate PIDs.
6. For each candidate PID, execute `ps -o ppid= -p <pid>` and retain only PIDs whose parent PID equals `1` (launchd).
7. Return the following based on filtered candidate count:
   - Exactly one PID: `StatusRunning`, that PID, confidence `verified`
   - Zero PIDs: `StatusStopped`, PID `0`, confidence `verified`
   - More than one PID: `StatusRunning`, the first PID, confidence `unverified`

#### Scenario: Single matching PID with launchd parent

- **WHEN** a daemon's program has exactly one running process owned by `UserName` with parent PID 1
- **THEN** the Service reports `StatusRunning`, the matching PID, and `StatusConfidence = "verified"`

#### Scenario: No matching process

- **WHEN** `pgrep -u <UserName> -f <program>` returns no output
- **THEN** the Service reports `StatusStopped`, PID `0`, and `StatusConfidence = "verified"`

#### Scenario: Multiple candidates with launchd parent

- **WHEN** two or more processes match and both have parent PID 1
- **THEN** the Service reports `StatusRunning` with the first PID and `StatusConfidence = "unverified"`

#### Scenario: User-launched process with non-launchd parent is filtered out

- **WHEN** `pgrep` finds a process owned by the target user but its parent PID is a shell PID (not 1)
- **THEN** the detection discards that PID and the Service reports `StatusStopped` with confidence `verified`

#### Scenario: UserName defaults to root when absent

- **WHEN** the service plist has no `UserName` key
- **THEN** the detection uses `root` as the user filter for `pgrep -u`

#### Scenario: Non-root UserName is honored

- **WHEN** the service plist has `UserName = _www`
- **THEN** the detection uses `_www` as the user filter and does not match root-owned processes

#### Scenario: Empty program path

- **WHEN** the service plist has neither `Program` nor `ProgramArguments`
- **THEN** the Service reports `StatusUnknown`, PID `0`, and `StatusConfidence = "unverified"`

---
### Requirement: Shell skip list

The detection logic SHALL skip `pgrep` matching when the program path equals one of: `/bin/bash`, `/bin/sh`, `/bin/zsh`, `/usr/bin/bash`, `/usr/bin/sh`, `/usr/bin/zsh`. For these programs the service SHALL be reported as `StatusLoaded`, PID `0`, confidence `verified`.

#### Scenario: Daemon invokes /bin/bash as Program

- **WHEN** the plist has `Program = /bin/bash`
- **THEN** the detection skips pgrep and reports `StatusLoaded` with PID `0` and confidence `verified`

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