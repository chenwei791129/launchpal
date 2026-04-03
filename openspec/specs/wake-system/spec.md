# wake-system Specification

## Purpose

TBD - created by archiving change 'add-wake-system-support'. Update Purpose after archive.

## Requirements

### Requirement: WakeSystem field in Service and ServiceConfig

The `Service` struct SHALL include a `WakeSystem` boolean field (`json:"wakeSystem"`). The `ServiceConfig` struct SHALL include a `WakeSystem` boolean field (`json:"wakeSystem"`). The `plistData` struct SHALL include a `WakeSystem` boolean field (`plist:"WakeSystem"`).

#### Scenario: Service struct exposes WakeSystem

- **WHEN** a plist contains `<key>WakeSystem</key><true/>`
- **THEN** the parsed `Service.WakeSystem` SHALL be `true`

#### Scenario: Service without WakeSystem defaults to false

- **WHEN** a plist does not contain a `WakeSystem` key
- **THEN** the parsed `Service.WakeSystem` SHALL be `false`


<!-- @trace
source: add-wake-system-support
updated: 2026-04-03
code:
  - app.go
  - frontend/wailsjs/go/models.ts
  - frontend/app/components/ScheduleForm.vue
  - internal/launchctl/readonly.go
  - frontend/app/pages/settings.vue
  - frontend/app/pages/apple-system.vue
  - frontend/wailsjs/go/main/App.js
  - frontend/app/composables/useAppVersion.ts
  - internal/backup/backup.go
  - internal/launchctl/apple_system.go
  - frontend/app/components/ServiceSummary.vue
  - frontend/app/components/StatusBar.vue
  - internal/launchctl/system.go
  - frontend/wailsjs/go/main/App.d.ts
  - frontend/app/types/wails.d.ts
  - internal/launchctl/user.go
  - frontend/app/components/ReadOnlyServiceList.vue
  - frontend/app/pages/system.vue
  - internal/launchctl/types.go
  - frontend/app/pages/services/[name].vue
  - README.md
  - frontend/app/components/CreateServiceModal.vue
tests:
  - internal/launchctl/system_test.go
  - internal/launchctl/user_test.go
  - internal/launchctl/apple_system_test.go
  - app_test.go
-->

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


<!-- @trace
source: add-wake-system-support
updated: 2026-04-03
code:
  - app.go
  - frontend/wailsjs/go/models.ts
  - frontend/app/components/ScheduleForm.vue
  - internal/launchctl/readonly.go
  - frontend/app/pages/settings.vue
  - frontend/app/pages/apple-system.vue
  - frontend/wailsjs/go/main/App.js
  - frontend/app/composables/useAppVersion.ts
  - internal/backup/backup.go
  - internal/launchctl/apple_system.go
  - frontend/app/components/ServiceSummary.vue
  - frontend/app/components/StatusBar.vue
  - internal/launchctl/system.go
  - frontend/wailsjs/go/main/App.d.ts
  - frontend/app/types/wails.d.ts
  - internal/launchctl/user.go
  - frontend/app/components/ReadOnlyServiceList.vue
  - frontend/app/pages/system.vue
  - internal/launchctl/types.go
  - frontend/app/pages/services/[name].vue
  - README.md
  - frontend/app/components/CreateServiceModal.vue
tests:
  - internal/launchctl/system_test.go
  - internal/launchctl/user_test.go
  - internal/launchctl/apple_system_test.go
  - app_test.go
-->

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


<!-- @trace
source: add-wake-system-support
updated: 2026-04-03
code:
  - app.go
  - frontend/wailsjs/go/models.ts
  - frontend/app/components/ScheduleForm.vue
  - internal/launchctl/readonly.go
  - frontend/app/pages/settings.vue
  - frontend/app/pages/apple-system.vue
  - frontend/wailsjs/go/main/App.js
  - frontend/app/composables/useAppVersion.ts
  - internal/backup/backup.go
  - internal/launchctl/apple_system.go
  - frontend/app/components/ServiceSummary.vue
  - frontend/app/components/StatusBar.vue
  - internal/launchctl/system.go
  - frontend/wailsjs/go/main/App.d.ts
  - frontend/app/types/wails.d.ts
  - internal/launchctl/user.go
  - frontend/app/components/ReadOnlyServiceList.vue
  - frontend/app/pages/system.vue
  - internal/launchctl/types.go
  - frontend/app/pages/services/[name].vue
  - README.md
  - frontend/app/components/CreateServiceModal.vue
tests:
  - internal/launchctl/system_test.go
  - internal/launchctl/user_test.go
  - internal/launchctl/apple_system_test.go
  - app_test.go
-->

---
### Requirement: WakeSystem display in ServiceSummary

The `ServiceSummary` component SHALL display the WakeSystem status for all service types.

#### Scenario: Display WakeSystem enabled

- **WHEN** a service has `WakeSystem` set to `true`
- **THEN** ServiceSummary SHALL display "Wake System" with value "Yes"

#### Scenario: Display WakeSystem disabled or absent

- **WHEN** a service has `WakeSystem` set to `false`
- **THEN** ServiceSummary SHALL display "Wake System" with value "No"

<!-- @trace
source: add-wake-system-support
updated: 2026-04-03
code:
  - app.go
  - frontend/wailsjs/go/models.ts
  - frontend/app/components/ScheduleForm.vue
  - internal/launchctl/readonly.go
  - frontend/app/pages/settings.vue
  - frontend/app/pages/apple-system.vue
  - frontend/wailsjs/go/main/App.js
  - frontend/app/composables/useAppVersion.ts
  - internal/backup/backup.go
  - internal/launchctl/apple_system.go
  - frontend/app/components/ServiceSummary.vue
  - frontend/app/components/StatusBar.vue
  - internal/launchctl/system.go
  - frontend/wailsjs/go/main/App.d.ts
  - frontend/app/types/wails.d.ts
  - internal/launchctl/user.go
  - frontend/app/components/ReadOnlyServiceList.vue
  - frontend/app/pages/system.vue
  - internal/launchctl/types.go
  - frontend/app/pages/services/[name].vue
  - README.md
  - frontend/app/components/CreateServiceModal.vue
tests:
  - internal/launchctl/system_test.go
  - internal/launchctl/user_test.go
  - internal/launchctl/apple_system_test.go
  - app_test.go
-->