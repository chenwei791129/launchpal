# cron-range-expansion Specification

## Purpose

TBD - created by archiving change 'cron-range-scheduling'. Update Purpose after archive.

## Requirements

### Requirement: Cron field range syntax

The cron parser SHALL accept range syntax `a-b` in any of the five fields (minute, hour, day, month, weekday). When a field contains `a-b`, the parser SHALL expand it to all integers from `a` to `b` inclusive. The values `a` and `b` MUST be within the valid range for that field. `a` MUST be less than or equal to `b`.

#### Scenario: Hour range expansion

- **WHEN** the cron expression is `0 9-11 * * *`
- **THEN** the parser SHALL produce 3 calendar entries: `{minute: 0, hour: 9}`, `{minute: 0, hour: 10}`, `{minute: 0, hour: 11}`

#### Scenario: Minute range expansion

- **WHEN** the cron expression is `0-2 9 * * *`
- **THEN** the parser SHALL produce 3 calendar entries: `{minute: 0, hour: 9}`, `{minute: 1, hour: 9}`, `{minute: 2, hour: 9}`

#### Scenario: Invalid range bounds

- **WHEN** the cron expression contains `25-30` in the hour field
- **THEN** the parser SHALL return a validation error indicating the value is out of range (hour: 0-23)

#### Scenario: Reversed range

- **WHEN** the cron expression contains `17-9` in the hour field
- **THEN** the parser SHALL return a validation error indicating that the start value MUST be less than or equal to the end value


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
### Requirement: Cron field enumeration syntax

The cron parser SHALL accept enumeration syntax `a,b,c` in any of the five fields. When a field contains comma-separated values, the parser SHALL use each listed value. All values MUST be within the valid range for that field. Duplicate values SHALL be deduplicated.

#### Scenario: Hour enumeration

- **WHEN** the cron expression is `0 9,12,18 * * *`
- **THEN** the parser SHALL produce 3 calendar entries: `{minute: 0, hour: 9}`, `{minute: 0, hour: 12}`, `{minute: 0, hour: 18}`

#### Scenario: Weekday enumeration

- **WHEN** the cron expression is `0 9 * * 1,3,5`
- **THEN** the parser SHALL produce 3 calendar entries: `{minute: 0, hour: 9, weekday: 1}`, `{minute: 0, hour: 9, weekday: 3}`, `{minute: 0, hour: 9, weekday: 5}`

#### Scenario: Invalid enumeration value

- **WHEN** the cron expression contains `9,25` in the hour field
- **THEN** the parser SHALL return a validation error for the out-of-range value 25

#### Scenario: Duplicate values in enumeration

- **WHEN** the cron expression contains `9,9,12` in the hour field
- **THEN** the parser SHALL deduplicate and produce entries for hours 9 and 12 only


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
### Requirement: Cartesian product expansion across multiple fields

When multiple fields contain range or enumeration syntax, the parser SHALL compute the Cartesian product of all expanded values. Fields using `*` or a single value SHALL be treated as a single-element set for the product.

#### Scenario: Hour range with weekday enumeration

- **WHEN** the cron expression is `0 9-11 * * 1,3`
- **THEN** the parser SHALL produce 6 calendar entries: all combinations of hours {9, 10, 11} and weekdays {1, 3}, each with minute 0

#### Scenario: Single value fields in Cartesian product

- **WHEN** the cron expression is `30 9-10 * * *`
- **THEN** the parser SHALL produce 2 calendar entries: `{minute: 30, hour: 9}` and `{minute: 30, hour: 10}`

#### Scenario: Wildcard fields excluded from entries

- **WHEN** the cron expression is `0 9-10 * * *`
- **THEN** each calendar entry SHALL NOT include day, month, or weekday fields (wildcard fields are omitted)


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
### Requirement: Expansion count limit

The parser SHALL enforce a maximum expansion count of 50 calendar entries. When the Cartesian product exceeds 50 entries, the parser SHALL return a validation error.

#### Scenario: Expansion within limit

- **WHEN** the cron expression is `0 0-23 * * 1-5`
- **THEN** the parser SHALL reject the expression because the expansion produces 120 entries (24 hours × 5 weekdays), which exceeds the limit of 50

#### Scenario: Expansion at limit

- **WHEN** the cron expression is `0 9-17 * * 1-5`
- **THEN** the parser SHALL accept the expression because the expansion produces 45 entries (9 hours × 5 weekdays), which is within the limit of 50


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
### Requirement: Expansion preview with summary and expandable list

The ScheduleForm SHALL display a preview of the expanded calendar entries. The preview SHALL show a summary line indicating the total count and a human-readable description. The summary SHALL be clickable to expand/collapse the full list of entries.

#### Scenario: Summary display for range expression

- **WHEN** the cron expression is `0 9-17 * * *`
- **THEN** the preview SHALL display a summary like "9 schedules: hour 9-17, at minute 00"

#### Scenario: Expandable list toggle

- **WHEN** the user clicks on the summary line
- **THEN** the preview SHALL expand to show each individual entry (e.g., "09:00", "10:00", ..., "17:00")
- **WHEN** the user clicks again
- **THEN** the list SHALL collapse back to the summary

#### Scenario: Single entry preview unchanged

- **WHEN** the cron expression is `0 9 * * *`
- **THEN** the preview SHALL display the existing text description format without expand/collapse functionality

#### Scenario: Expansion limit exceeded preview

- **WHEN** the Cartesian product exceeds 50 entries
- **THEN** the preview SHALL display an error message indicating the limit is exceeded

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