## MODIFIED Requirements

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
