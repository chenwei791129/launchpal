## ADDED Requirements

### Requirement: WakeSystem field in Service and ServiceConfig

The `Service` struct SHALL include a `WakeSystem` boolean field (`json:"wakeSystem"`). The `ServiceConfig` struct SHALL include a `WakeSystem` boolean field (`json:"wakeSystem"`). The `plistData` struct SHALL include a `WakeSystem` boolean field (`plist:"WakeSystem"`).

#### Scenario: Service struct exposes WakeSystem

- **WHEN** a plist contains `<key>WakeSystem</key><true/>`
- **THEN** the parsed `Service.WakeSystem` SHALL be `true`

#### Scenario: Service without WakeSystem defaults to false

- **WHEN** a plist does not contain a `WakeSystem` key
- **THEN** the parsed `Service.WakeSystem` SHALL be `false`

---

### Requirement: WakeSystem plist read support across all managers

All three service managers (UserManager, SystemManager, AppleSystemManager) SHALL parse the `WakeSystem` key from plist files and populate the `Service.WakeSystem` field.

#### Scenario: UserManager reads WakeSystem from user agent plist

- **WHEN** a user agent plist at `~/Library/LaunchAgents/` contains `<key>WakeSystem</key><true/>`
- **THEN** `UserManager.Get()` SHALL return a `Service` with `WakeSystem` set to `true`

#### Scenario: SystemManager reads WakeSystem from system daemon plist

- **WHEN** a system daemon plist at `/Library/LaunchDaemons/` contains `<key>WakeSystem</key><true/>`
- **THEN** `SystemManager.Get()` SHALL return a `Service` with `WakeSystem` set to `true`

#### Scenario: AppleSystemManager reads WakeSystem from Apple system daemon plist

- **WHEN** an Apple system daemon plist at `/System/Library/LaunchDaemons/` contains `<key>WakeSystem</key><true/>`
- **THEN** `AppleSystemManager.Get()` SHALL return a `Service` with `WakeSystem` set to `true`

---

### Requirement: WakeSystem plist write support

When `ServiceConfig.WakeSystem` is `true`, `writePlist` SHALL write `<key>WakeSystem</key><true/>` to the plist output. When `ServiceConfig.WakeSystem` is `false`, `writePlist` SHALL NOT write a `WakeSystem` key.

#### Scenario: Create service with WakeSystem enabled

- **WHEN** a service is created with `ServiceConfig.WakeSystem` set to `true`
- **THEN** the generated plist SHALL contain `<key>WakeSystem</key><true/>`

#### Scenario: Create service with WakeSystem disabled

- **WHEN** a service is created with `ServiceConfig.WakeSystem` set to `false`
- **THEN** the generated plist SHALL NOT contain a `WakeSystem` key

#### Scenario: Update service to enable WakeSystem

- **WHEN** an existing service without WakeSystem is updated with `ServiceConfig.WakeSystem` set to `true`
- **THEN** the updated plist SHALL contain `<key>WakeSystem</key><true/>`

---

### Requirement: WakeSystem display in ServiceSummary

The `ServiceSummary` component SHALL display the WakeSystem status for all service types.

#### Scenario: Display WakeSystem enabled

- **WHEN** a service has `WakeSystem` set to `true`
- **THEN** ServiceSummary SHALL display "Wake System" with value "Yes"

#### Scenario: Display WakeSystem disabled or absent

- **WHEN** a service has `WakeSystem` set to `false`
- **THEN** ServiceSummary SHALL display "Wake System" with value "No"
