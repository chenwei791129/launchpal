# scheduled-service Specification

## Purpose

User services can express their run schedule as either a fixed interval (`StartInterval`) or one-or-more `StartCalendarInterval` entries. `ScheduleConfig.Interval` and `ScheduleConfig.Schedules []CalendarEntry` are mutually exclusive at write time, and the parser round-trips both forms so existing plists keep working unchanged.

## Requirements

### Requirement: Backend StartInterval support

The system SHALL support `StartInterval` in `ScheduleConfig` as an optional integer field representing seconds between executions. When `ScheduleConfig.Interval` is set, `writePlist` SHALL write a `StartInterval` key to the plist. When reading a plist with `StartInterval`, `parseSchedule` SHALL populate `ScheduleConfig.Interval`.

The `ScheduleConfig` struct SHALL use a `Schedules` field of type `[]CalendarEntry` to represent calendar intervals, replacing the previous top-level `Minute`, `Hour`, `Day`, `Weekday`, `Month` fields. The `HasMultiple` field SHALL be removed.

`CalendarEntry` SHALL be a struct with optional fields: `Minute`, `Hour`, `Day`, `Weekday`, `Month` (all `*int`).

#### Scenario: Create service with StartInterval

- **WHEN** a service is created with `ScheduleConfig.Interval` set to 3600
- **THEN** the generated plist SHALL contain `<key>StartInterval</key><integer>3600</integer>`

#### Scenario: Read existing service with StartInterval

- **WHEN** a plist contains `StartInterval` with value 1800
- **THEN** the parsed `Service.Schedule.Interval` SHALL be a pointer to 1800

#### Scenario: StartInterval and CalendarInterval are mutually derived

- **WHEN** `ScheduleConfig.Interval` is set
- **THEN** `writePlist` SHALL write `StartInterval` and SHALL NOT write `StartCalendarInterval`
- **WHEN** `ScheduleConfig.Interval` is nil and `Schedules` is non-empty
- **THEN** `writePlist` SHALL write `StartCalendarInterval` and SHALL NOT write `StartInterval`

#### Scenario: Write single calendar entry

- **WHEN** `ScheduleConfig.Schedules` contains exactly one `CalendarEntry`
- **THEN** `writePlist` SHALL write `StartCalendarInterval` as a single dict

#### Scenario: Write multiple calendar entries

- **WHEN** `ScheduleConfig.Schedules` contains more than one `CalendarEntry`
- **THEN** `writePlist` SHALL write `StartCalendarInterval` as an array of dicts

#### Scenario: Read plist with single calendar interval dict

- **WHEN** a plist contains `StartCalendarInterval` as a single dict with `{Hour: 9, Minute: 0}`
- **THEN** `parseSchedule` SHALL return `ScheduleConfig.Schedules` with one `CalendarEntry{Hour: &9, Minute: &0}`

#### Scenario: Read plist with multiple calendar interval array

- **WHEN** a plist contains `StartCalendarInterval` as an array of 3 dicts
- **THEN** `parseSchedule` SHALL return `ScheduleConfig.Schedules` with 3 `CalendarEntry` elements, each populated from the corresponding dict


<!-- @trace
source: cron-range-scheduling
updated: 2026-04-11
code:
  - frontend/app/pages/system.vue
  - frontend/app/pages/settings.vue
  - frontend/app/components/StatusBar.vue
  - app.go
  - internal/backup/backup.go
  - frontend/wailsjs/go/main/App.d.ts
  - frontend/app/components/ServiceSummary.vue
  - README.md
  - CHANGELOG.md
  - frontend/app/types/wails.d.ts
  - internal/launchctl/readonly.go
  - internal/launchctl/user.go
  - frontend/app/pages/services/[name].vue
  - go.mod
  - frontend/app/components/ReadOnlyServiceList.vue
  - frontend/wailsjs/go/main/App.js
  - internal/launchctl/apple_system.go
  - internal/launchctl/system.go
  - frontend/app/composables/useAppVersion.ts
  - frontend/wailsjs/go/models.ts
  - frontend/app/components/CreateServiceModal.vue
  - internal/launchctl/types.go
  - frontend/app/pages/apple-system.vue
  - frontend/app/composables/useNextOccurrences.ts
  - frontend/app/components/ScheduleForm.vue
tests:
  - internal/launchctl/user_test.go
  - internal/launchctl/apple_system_test.go
  - frontend/app/composables/__tests__/useNextOccurrences.test.ts
  - internal/launchctl/system_test.go
  - app_test.go
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

#### Scenario: Display single CalendarInterval schedule

- **WHEN** a service has one `CalendarEntry` in `Schedules`
- **THEN** ServiceSummary SHALL display the schedule fields (e.g., "Every day at 09:00")

#### Scenario: Display multiple CalendarInterval schedules

- **WHEN** a service has multiple `CalendarEntry` items in `Schedules`
- **THEN** ServiceSummary SHALL display a summary count and description (e.g., "9 schedules: hour 9-17, at minute 00")
- **THEN** ServiceSummary SHALL NOT display the previous "Multiple schedules defined; only the first is shown" warning

#### Scenario: Display StartInterval schedule

- **WHEN** a service has `StartInterval` configured
- **THEN** ServiceSummary SHALL display the interval (e.g., "Every 3600 seconds")

#### Scenario: No schedule configured

- **WHEN** a service has no schedule
- **THEN** ServiceSummary SHALL NOT display a schedule section


<!-- @trace
source: cron-range-scheduling
updated: 2026-04-11
code:
  - frontend/app/pages/system.vue
  - frontend/app/pages/settings.vue
  - frontend/app/components/StatusBar.vue
  - app.go
  - internal/backup/backup.go
  - frontend/wailsjs/go/main/App.d.ts
  - frontend/app/components/ServiceSummary.vue
  - README.md
  - CHANGELOG.md
  - frontend/app/types/wails.d.ts
  - internal/launchctl/readonly.go
  - internal/launchctl/user.go
  - frontend/app/pages/services/[name].vue
  - go.mod
  - frontend/app/components/ReadOnlyServiceList.vue
  - frontend/wailsjs/go/main/App.js
  - internal/launchctl/apple_system.go
  - internal/launchctl/system.go
  - frontend/app/composables/useAppVersion.ts
  - frontend/wailsjs/go/models.ts
  - frontend/app/components/CreateServiceModal.vue
  - internal/launchctl/types.go
  - frontend/app/pages/apple-system.vue
  - frontend/app/composables/useNextOccurrences.ts
  - frontend/app/components/ScheduleForm.vue
tests:
  - internal/launchctl/user_test.go
  - internal/launchctl/apple_system_test.go
  - frontend/app/composables/__tests__/useNextOccurrences.test.ts
  - internal/launchctl/system_test.go
  - app_test.go
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

---
### Requirement: WakeSystem toggle in schedule form

The `ScheduleForm` component SHALL include a "Wake System" toggle (checkbox). The toggle SHALL only be visible when the schedule is enabled. The toggle value SHALL be emitted as part of the schedule configuration to the parent component.

#### Scenario: WakeSystem toggle appears when schedule is enabled

- **WHEN** the user enables the schedule in `ScheduleForm`
- **THEN** a "Wake System" checkbox SHALL be displayed below the schedule type options

#### Scenario: WakeSystem toggle hidden when schedule is disabled

- **WHEN** the schedule is not enabled in `ScheduleForm`
- **THEN** the "Wake System" checkbox SHALL NOT be displayed

#### Scenario: WakeSystem state is emitted with schedule config

- **WHEN** the user enables the "Wake System" toggle
- **THEN** the emitted `update:wakeSystem` event SHALL carry the value `true`

#### Scenario: WakeSystem toggle initializes from existing service

- **WHEN** `ScheduleForm` receives a service with `WakeSystem` set to `true` and a schedule configured
- **THEN** the "Wake System" toggle SHALL be checked

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