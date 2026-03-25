# scheduled-service Specification

## Purpose

TBD - created by archiving change 'add-scheduled-service-type'. Update Purpose after archive.

## Requirements

### Requirement: Backend StartInterval support

The system SHALL support `StartInterval` in `ScheduleConfig` as an optional integer field representing seconds between executions. When `ScheduleConfig.Interval` is set, `writePlist` SHALL write a `StartInterval` key to the plist. When reading a plist with `StartInterval`, `parseSchedule` SHALL populate `ScheduleConfig.Interval`.

#### Scenario: Create service with StartInterval

- **WHEN** a service is created with `ScheduleConfig.Interval` set to 3600
- **THEN** the generated plist SHALL contain `<key>StartInterval</key><integer>3600</integer>`

#### Scenario: Read existing service with StartInterval

- **WHEN** a plist contains `StartInterval` with value 1800
- **THEN** the parsed `Service.Schedule.Interval` SHALL be a pointer to 1800

#### Scenario: StartInterval and CalendarInterval are mutually derived

- **WHEN** `ScheduleConfig.Interval` is set
- **THEN** `writePlist` SHALL write `StartInterval` and SHALL NOT write `StartCalendarInterval`
- **WHEN** `ScheduleConfig.Interval` is nil and calendar fields are set
- **THEN** `writePlist` SHALL write `StartCalendarInterval` and SHALL NOT write `StartInterval`


<!-- @trace
source: add-scheduled-service-type
updated: 2026-03-26
code:
  - internal/launchctl/apple_system.go
  - internal/launchctl/types.go
  - frontend/wailsjs/go/models.ts
  - internal/launchctl/user.go
  - frontend/app/types/wails.d.ts
  - CLAUDE.md
  - frontend/app/components/CreateServiceModal.vue
  - frontend/app/components/ScheduleForm.vue
  - internal/launchctl/system.go
  - README.md
  - frontend/app/pages/services/[name].vue
  - frontend/app/components/ServiceSummary.vue
  - frontend/app/components/ServiceRow.vue
tests:
  - internal/launchctl/user_test.go
-->

---
### Requirement: Schedule type selection in New Service UI

The New Service form SHALL include an "Enable Schedule" option. When enabled, the user SHALL be able to choose between "Calendar Interval" and "Fixed Interval" as the schedule sub-type.

#### Scenario: User enables schedule with Calendar Interval

- **WHEN** user enables schedule and selects "Calendar Interval"
- **THEN** the form SHALL display optional fields for Minute, Hour, Day, Weekday, and Month
- **THEN** the form SHALL allow the user to leave any field unset

#### Scenario: User enables schedule with Fixed Interval

- **WHEN** user enables schedule and selects "Fixed Interval"
- **THEN** the form SHALL display a numeric input for interval in seconds
- **THEN** the minimum accepted value SHALL be 10 seconds

#### Scenario: Schedule coexists with RunAtLoad

- **WHEN** user enables both "Run at Load" and a schedule
- **THEN** the generated plist SHALL contain both `RunAtLoad` and the schedule key


<!-- @trace
source: add-scheduled-service-type
updated: 2026-03-26
code:
  - internal/launchctl/apple_system.go
  - internal/launchctl/types.go
  - frontend/wailsjs/go/models.ts
  - internal/launchctl/user.go
  - frontend/app/types/wails.d.ts
  - CLAUDE.md
  - frontend/app/components/CreateServiceModal.vue
  - frontend/app/components/ScheduleForm.vue
  - internal/launchctl/system.go
  - README.md
  - frontend/app/pages/services/[name].vue
  - frontend/app/components/ServiceSummary.vue
  - frontend/app/components/ServiceRow.vue
tests:
  - internal/launchctl/user_test.go
-->

---
### Requirement: Schedule information display in ServiceSummary

The ServiceSummary component SHALL display schedule information when a service has a schedule configured.

#### Scenario: Display CalendarInterval schedule

- **WHEN** a service has `StartCalendarInterval` configured
- **THEN** ServiceSummary SHALL display the schedule fields (e.g., "Every day at 03:00")

#### Scenario: Display StartInterval schedule

- **WHEN** a service has `StartInterval` configured
- **THEN** ServiceSummary SHALL display the interval (e.g., "Every 3600 seconds")

#### Scenario: No schedule configured

- **WHEN** a service has no schedule
- **THEN** ServiceSummary SHALL NOT display a schedule section


<!-- @trace
source: add-scheduled-service-type
updated: 2026-03-26
code:
  - internal/launchctl/apple_system.go
  - internal/launchctl/types.go
  - frontend/wailsjs/go/models.ts
  - internal/launchctl/user.go
  - frontend/app/types/wails.d.ts
  - CLAUDE.md
  - frontend/app/components/CreateServiceModal.vue
  - frontend/app/components/ScheduleForm.vue
  - internal/launchctl/system.go
  - README.md
  - frontend/app/pages/services/[name].vue
  - frontend/app/components/ServiceSummary.vue
  - frontend/app/components/ServiceRow.vue
tests:
  - internal/launchctl/user_test.go
-->

---
### Requirement: Schedule editing in service detail page

The service detail page SHALL support editing schedule configuration for user services.

#### Scenario: Edit schedule on existing service

- **WHEN** user edits a service and modifies the schedule configuration
- **THEN** the updated plist SHALL reflect the new schedule settings

#### Scenario: Add schedule to existing non-scheduled service

- **WHEN** user edits a service that has no schedule and enables a schedule
- **THEN** the updated plist SHALL include the new schedule key

#### Scenario: Remove schedule from existing scheduled service

- **WHEN** user edits a scheduled service and disables the schedule
- **THEN** the updated plist SHALL NOT contain `StartInterval` or `StartCalendarInterval`


<!-- @trace
source: add-scheduled-service-type
updated: 2026-03-26
code:
  - internal/launchctl/apple_system.go
  - internal/launchctl/types.go
  - frontend/wailsjs/go/models.ts
  - internal/launchctl/user.go
  - frontend/app/types/wails.d.ts
  - CLAUDE.md
  - frontend/app/components/CreateServiceModal.vue
  - frontend/app/components/ScheduleForm.vue
  - internal/launchctl/system.go
  - README.md
  - frontend/app/pages/services/[name].vue
  - frontend/app/components/ServiceSummary.vue
  - frontend/app/components/ServiceRow.vue
tests:
  - internal/launchctl/user_test.go
-->

---
### Requirement: Calendar Interval empty field validation

The system SHALL warn users when all CalendarInterval fields are left empty, as this results in execution every minute.

#### Scenario: All CalendarInterval fields empty

- **WHEN** user selects Calendar Interval but leaves all fields empty
- **THEN** the form SHALL display a warning indicating the service will run every minute
- **THEN** the form SHALL still allow submission if the user confirms

<!-- @trace
source: add-scheduled-service-type
updated: 2026-03-26
code:
  - internal/launchctl/apple_system.go
  - internal/launchctl/types.go
  - frontend/wailsjs/go/models.ts
  - internal/launchctl/user.go
  - frontend/app/types/wails.d.ts
  - CLAUDE.md
  - frontend/app/components/CreateServiceModal.vue
  - frontend/app/components/ScheduleForm.vue
  - internal/launchctl/system.go
  - README.md
  - frontend/app/pages/services/[name].vue
  - frontend/app/components/ServiceSummary.vue
  - frontend/app/components/ServiceRow.vue
tests:
  - internal/launchctl/user_test.go
-->