# kickstart-service Specification

## Purpose

User services expose a Run Now action that immediately executes the loaded plist by invoking `launchctl kickstart -k gui/{UID}/{label}`. The action is wired through the `App.KickstartService` Wails binding, surfaced as a button on the user service detail page, and prompts for confirmation when the service is already running so users understand the existing process will be terminated.

## ADDED Requirements

### Requirement: Kickstart a user service

The system SHALL execute `launchctl kickstart -k gui/{UID}/{label}` to immediately run a user service.
The system SHALL obtain the current user's UID via `os.Getuid()`.
The system SHALL return an error if the plist file does not exist.

#### Scenario: Service is loaded and not running

- **WHEN** the user triggers kickstart on a service that is loaded but not running
- **THEN** the system executes `launchctl kickstart -k gui/{UID}/{label}` and returns no error

#### Scenario: Service is running

- **WHEN** the user triggers kickstart on a service that is currently running
- **THEN** the system executes `launchctl kickstart -k gui/{UID}/{label}`, which terminates the existing process and starts a new one

#### Scenario: Service is not loaded

- **WHEN** the user triggers kickstart on a service that is not loaded (stopped)
- **THEN** the system SHALL first execute `launchctl bootstrap gui/{UID} {plistPath}` to load the service
- **THEN** the system SHALL execute `launchctl kickstart -k gui/{UID}/{label}` to run it

#### Scenario: Plist file does not exist

- **WHEN** the user triggers kickstart on a service whose plist file does not exist
- **THEN** the system SHALL return an error indicating the service was not found


<!-- @trace
source: run-now-kickstart
updated: 2026-04-04
code:
  - app.go
  - frontend/app/components/ReadOnlyServiceList.vue
  - frontend/app/composables/useAppVersion.ts
  - frontend/app/pages/apple-system.vue
  - internal/launchctl/system.go
  - frontend/app/types/wails.d.ts
  - internal/launchctl/readonly.go
  - internal/launchctl/user.go
  - frontend/app/components/StatusBar.vue
  - internal/launchctl/apple_system.go
  - frontend/wailsjs/go/main/App.js
  - frontend/app/pages/system.vue
  - frontend/app/components/CreateServiceModal.vue
  - internal/launchctl/types.go
  - frontend/wailsjs/go/main/App.d.ts
  - frontend/app/pages/services/[name].vue
  - frontend/app/pages/settings.vue
  - frontend/wailsjs/go/models.ts
  - README.md
  - frontend/app/composables/useNextOccurrences.ts
  - CHANGELOG.md
  - frontend/app/components/ScheduleForm.vue
  - frontend/app/components/ServiceSummary.vue
  - internal/backup/backup.go
tests:
  - app_test.go
  - internal/launchctl/apple_system_test.go
  - internal/launchctl/user_test.go
  - internal/launchctl/system_test.go
  - frontend/app/composables/__tests__/useNextOccurrences.test.ts
-->

### Requirement: Kickstart backend binding

The system SHALL expose a `KickstartService(name string) error` method on the `App` struct as a Wails binding.
The method SHALL delegate to `UserManager.Kickstart`.

#### Scenario: Frontend calls KickstartService

- **WHEN** the frontend calls `window.go.main.App.KickstartService(name)`
- **THEN** the system executes the kickstart operation and returns the result


<!-- @trace
source: run-now-kickstart
updated: 2026-04-04
code:
  - app.go
  - frontend/app/components/ReadOnlyServiceList.vue
  - frontend/app/composables/useAppVersion.ts
  - frontend/app/pages/apple-system.vue
  - internal/launchctl/system.go
  - frontend/app/types/wails.d.ts
  - internal/launchctl/readonly.go
  - internal/launchctl/user.go
  - frontend/app/components/StatusBar.vue
  - internal/launchctl/apple_system.go
  - frontend/wailsjs/go/main/App.js
  - frontend/app/pages/system.vue
  - frontend/app/components/CreateServiceModal.vue
  - internal/launchctl/types.go
  - frontend/wailsjs/go/main/App.d.ts
  - frontend/app/pages/services/[name].vue
  - frontend/app/pages/settings.vue
  - frontend/wailsjs/go/models.ts
  - README.md
  - frontend/app/composables/useNextOccurrences.ts
  - CHANGELOG.md
  - frontend/app/components/ScheduleForm.vue
  - frontend/app/components/ServiceSummary.vue
  - internal/backup/backup.go
tests:
  - app_test.go
  - internal/launchctl/apple_system_test.go
  - internal/launchctl/user_test.go
  - internal/launchctl/system_test.go
  - frontend/app/composables/__tests__/useNextOccurrences.test.ts
-->

### Requirement: Run Now button in service detail page

The frontend SHALL display a "Run Now" button in the service detail page header action buttons area, alongside the existing Start/Stop and Restart buttons.
The button SHALL only appear for user services (not read-only services).

#### Scenario: User service detail page

- **WHEN** the user views a user service detail page
- **THEN** a "Run Now" button is visible in the header action buttons

#### Scenario: Read-only service detail page

- **WHEN** the user views a system or apple-system service detail page
- **THEN** no "Run Now" button is displayed


<!-- @trace
source: run-now-kickstart
updated: 2026-04-04
code:
  - app.go
  - frontend/app/components/ReadOnlyServiceList.vue
  - frontend/app/composables/useAppVersion.ts
  - frontend/app/pages/apple-system.vue
  - internal/launchctl/system.go
  - frontend/app/types/wails.d.ts
  - internal/launchctl/readonly.go
  - internal/launchctl/user.go
  - frontend/app/components/StatusBar.vue
  - internal/launchctl/apple_system.go
  - frontend/wailsjs/go/main/App.js
  - frontend/app/pages/system.vue
  - frontend/app/components/CreateServiceModal.vue
  - internal/launchctl/types.go
  - frontend/wailsjs/go/main/App.d.ts
  - frontend/app/pages/services/[name].vue
  - frontend/app/pages/settings.vue
  - frontend/wailsjs/go/models.ts
  - README.md
  - frontend/app/composables/useNextOccurrences.ts
  - CHANGELOG.md
  - frontend/app/components/ScheduleForm.vue
  - frontend/app/components/ServiceSummary.vue
  - internal/backup/backup.go
tests:
  - app_test.go
  - internal/launchctl/apple_system_test.go
  - internal/launchctl/user_test.go
  - internal/launchctl/system_test.go
  - frontend/app/composables/__tests__/useNextOccurrences.test.ts
-->

### Requirement: Confirmation dialog when service is running

The frontend SHALL display a confirmation dialog when the user clicks "Run Now" and the service status is "running".
The dialog SHALL inform the user that the current process will be terminated and restarted.
The frontend SHALL NOT display a confirmation dialog when the service status is not "running" (stopped, loaded, or unknown).

#### Scenario: Click Run Now while service is running

- **WHEN** the user clicks "Run Now" and the service status is "running"
- **THEN** a confirmation dialog appears with a message explaining that the existing process will be terminated
- **WHEN** the user confirms
- **THEN** the system calls `KickstartService` and refreshes the service status

#### Scenario: Click Run Now while service is running and user cancels

- **WHEN** the user clicks "Run Now" and the service status is "running"
- **THEN** a confirmation dialog appears
- **WHEN** the user cancels
- **THEN** no action is taken

#### Scenario: Click Run Now while service is not running

- **WHEN** the user clicks "Run Now" and the service status is "stopped", "loaded", or "unknown"
- **THEN** the system calls `KickstartService` directly without showing a confirmation dialog
- **THEN** the service status is refreshed after execution

## Requirements


<!-- @trace
source: run-now-kickstart
updated: 2026-04-04
code:
  - app.go
  - frontend/app/components/ReadOnlyServiceList.vue
  - frontend/app/composables/useAppVersion.ts
  - frontend/app/pages/apple-system.vue
  - internal/launchctl/system.go
  - frontend/app/types/wails.d.ts
  - internal/launchctl/readonly.go
  - internal/launchctl/user.go
  - frontend/app/components/StatusBar.vue
  - internal/launchctl/apple_system.go
  - frontend/wailsjs/go/main/App.js
  - frontend/app/pages/system.vue
  - frontend/app/components/CreateServiceModal.vue
  - internal/launchctl/types.go
  - frontend/wailsjs/go/main/App.d.ts
  - frontend/app/pages/services/[name].vue
  - frontend/app/pages/settings.vue
  - frontend/wailsjs/go/models.ts
  - README.md
  - frontend/app/composables/useNextOccurrences.ts
  - CHANGELOG.md
  - frontend/app/components/ScheduleForm.vue
  - frontend/app/components/ServiceSummary.vue
  - internal/backup/backup.go
tests:
  - app_test.go
  - internal/launchctl/apple_system_test.go
  - internal/launchctl/user_test.go
  - internal/launchctl/system_test.go
  - frontend/app/composables/__tests__/useNextOccurrences.test.ts
-->

### Requirement: Kickstart a user service

The system SHALL execute `launchctl kickstart -k gui/{UID}/{label}` to immediately run a user service.
The system SHALL obtain the current user's UID via `os.Getuid()`.
The system SHALL return an error if the plist file does not exist.

#### Scenario: Service is loaded and not running

- **WHEN** the user triggers kickstart on a service that is loaded but not running
- **THEN** the system executes `launchctl kickstart -k gui/{UID}/{label}` and returns no error

#### Scenario: Service is running

- **WHEN** the user triggers kickstart on a service that is currently running
- **THEN** the system executes `launchctl kickstart -k gui/{UID}/{label}`, which terminates the existing process and starts a new one

#### Scenario: Service is not loaded

- **WHEN** the user triggers kickstart on a service that is not loaded (stopped)
- **THEN** the system SHALL first execute `launchctl bootstrap gui/{UID} {plistPath}` to load the service
- **THEN** the system SHALL execute `launchctl kickstart -k gui/{UID}/{label}` to run it

#### Scenario: Plist file does not exist

- **WHEN** the user triggers kickstart on a service whose plist file does not exist
- **THEN** the system SHALL return an error indicating the service was not found

---
### Requirement: Kickstart backend binding

The system SHALL expose a `KickstartService(name string) error` method on the `App` struct as a Wails binding.
The method SHALL delegate to `UserManager.Kickstart`.

#### Scenario: Frontend calls KickstartService

- **WHEN** the frontend calls `window.go.main.App.KickstartService(name)`
- **THEN** the system executes the kickstart operation and returns the result

---
### Requirement: Run Now button in service detail page

The frontend SHALL display a "Run Now" button in the service detail page header action buttons area, alongside the existing Start/Stop and Restart buttons.
The button SHALL only appear for user services (not read-only services).

#### Scenario: User service detail page

- **WHEN** the user views a user service detail page
- **THEN** a "Run Now" button is visible in the header action buttons

#### Scenario: Read-only service detail page

- **WHEN** the user views a system or apple-system service detail page
- **THEN** no "Run Now" button is displayed

---
### Requirement: Confirmation dialog when service is running

The frontend SHALL display a confirmation dialog when the user clicks "Run Now" and the service status is "running".
The dialog SHALL inform the user that the current process will be terminated and restarted.
The frontend SHALL NOT display a confirmation dialog when the service status is not "running" (stopped, loaded, or unknown).

#### Scenario: Click Run Now while service is running

- **WHEN** the user clicks "Run Now" and the service status is "running"
- **THEN** a confirmation dialog appears with a message explaining that the existing process will be terminated
- **WHEN** the user confirms
- **THEN** the system calls `KickstartService` and refreshes the service status

#### Scenario: Click Run Now while service is running and user cancels

- **WHEN** the user clicks "Run Now" and the service status is "running"
- **THEN** a confirmation dialog appears
- **WHEN** the user cancels
- **THEN** no action is taken

#### Scenario: Click Run Now while service is not running

- **WHEN** the user clicks "Run Now" and the service status is "stopped", "loaded", or "unknown"
- **THEN** the system calls `KickstartService` directly without showing a confirmation dialog
- **THEN** the service status is refreshed after execution