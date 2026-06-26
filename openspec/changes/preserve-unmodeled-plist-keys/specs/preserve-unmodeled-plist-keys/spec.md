## ADDED Requirements

### Requirement: Preserve unmodeled plist keys on service Update

When updating an existing service in either the user domain (`~/Library/LaunchAgents`) or the system domain (`/Library/LaunchDaemons`), the system SHALL read the existing on-disk plist and preserve every key that LaunchPal does not model, writing those keys back verbatim. The set of preserved keys includes, but is not limited to, `ProcessType`, `Nice`, `UserName`, `GroupName`, `SoftResourceLimits`, `HardResourceLimits`, `MachServices`, `Sockets`, `LimitLoadToSessionType`, `ExitTimeOut`, and `AbandonProcessGroup`. The system SHALL preserve the original value and type of each unmodeled key, including nested dictionaries and arrays. This read-merge-write behavior SHALL apply only to Update; Create SHALL continue to write a plist solely from the supplied configuration.

#### Scenario: User-domain Update preserves an unmodeled scalar key

- **WHEN** a user service plist contains `ProcessType=Background` and `Nice=5`, and Update is called with a configuration that only changes `Program`
- **THEN** the written plist still contains `ProcessType=Background` and `Nice=5`
- **AND** the written plist reflects the updated `Program`

#### Scenario: Unmodeled nested structures round-trip unchanged

- **WHEN** a service plist contains a `Sockets` dictionary and a `MachServices` dictionary, and Update is called
- **THEN** the written plist contains the `Sockets` and `MachServices` dictionaries with the same structure and values as before

#### Scenario: System-domain Update preserves an unmodeled key

- **WHEN** Admin Mode is enabled and a system daemon plist contains `ProcessType=Standard`, and Update is called with a configuration that changes a modeled field
- **THEN** the bytes written via the privileged helper still contain `ProcessType=Standard`

### Requirement: Modeled keys remain form-authoritative on Update

The system SHALL treat every modeled key as authoritative from the supplied configuration on Update. The system SHALL overwrite a modeled key's existing on-disk value with the configuration-derived value, and SHALL remove a modeled key from the written plist when the configuration does not set it — even if that key was present in the existing plist. The merge SHALL remove modeled keys based on the complete set of keys LaunchPal models, not only the keys emitted by the current configuration, so that toggling between mutually exclusive modeled keys (for example `StartInterval` and `StartCalendarInterval`) never leaves a stale key behind.

#### Scenario: Clearing a modeled key removes it from disk

- **WHEN** a service plist contains `RunAtLoad=true` and an unmodeled `ExitTimeOut=30`, and Update is called with a configuration whose launch policy is On Demand (no `RunAtLoad`)
- **THEN** the written plist does NOT contain `RunAtLoad`
- **AND** the written plist still contains `ExitTimeOut=30`

#### Scenario: Switching schedule type leaves no stale key

- **WHEN** a service plist contains `StartInterval=60`, and Update is called with a configuration that sets a calendar schedule instead
- **THEN** the written plist contains `StartCalendarInterval`
- **AND** the written plist does NOT contain `StartInterval`

### Requirement: Graceful degradation when the existing plist is unavailable

When the existing plist cannot be read or cannot be parsed (for example a system daemon plist that is unreadable without Full Disk Access, or a corrupt file), the system SHALL skip the merge and write the plist solely from the supplied configuration. The system SHALL NOT fail the Update operation solely because the existing plist could not be read or parsed. On the system domain the reload (Bootout) step SHALL still run using the routing name as the label, so the updated plist takes effect rather than leaving launchd on its stale in-memory definition.

#### Scenario: Unreadable existing plist degrades to a fresh write

- **WHEN** Update is called and the existing plist cannot be read or parsed
- **THEN** the written plist equals the output produced from the supplied configuration alone
- **AND** the Update operation returns success
- **AND** on the system domain the daemon is still booted out via the routing name before the new plist is written

### Requirement: Single source of truth for the modeled key set

The system SHALL define the set of modeled plist keys in exactly one place, and the merge logic SHALL derive its removal set from that single definition. The modeled key set SHALL include every key the configuration-to-plist encoder can emit.

#### Scenario: Modeled key set covers every encoder-emitted key

- **WHEN** the configuration-to-plist encoder is run across a set of configurations that together exercise every modeled field, including one configuration with `StartInterval` and one with `StartCalendarInterval` (the two are mutually exclusive within a single configuration)
- **THEN** every key emitted across those plists is a member of the modeled key set
