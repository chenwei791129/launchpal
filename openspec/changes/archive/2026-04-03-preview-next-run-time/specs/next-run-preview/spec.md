## ADDED Requirements

### Requirement: Calculate next occurrences from CalendarInterval

The system SHALL provide a frontend composable `useNextOccurrences` that computes the next N execution times from a `CalendarInterval` schedule configuration. The calculation SHALL start from the current time plus one minute and iterate minute-by-minute, matching each candidate against the CalendarInterval fields (month, day, weekday, hour, minute). A field with value `undefined` or `null` SHALL match any value for that field. The composable SHALL return an array of `Date` objects sorted chronologically.

#### Scenario: Specific hour and minute

- **WHEN** the CalendarInterval has hour=14 and minute=30, with day, weekday, and month unset
- **THEN** the system SHALL return dates where each occurrence is at 14:30 on consecutive days

#### Scenario: Specific weekday

- **WHEN** the CalendarInterval has weekday=1 (Monday), hour=9, minute=0
- **THEN** the system SHALL return dates that are all Mondays at 09:00

#### Scenario: All fields unset

- **WHEN** all CalendarInterval fields are unset (wildcard)
- **THEN** the system SHALL return dates at every minute starting from now+1

#### Scenario: Specific month and day

- **WHEN** the CalendarInterval has month=12, day=25, hour=0, minute=0
- **THEN** the system SHALL return dates on December 25th at 00:00 in consecutive years

### Requirement: Preview next runs in ScheduleForm

The ScheduleForm component SHALL display a preview of the next 3 execution times when the schedule type is "Calendar Interval" and the cron expression is valid. The preview SHALL update reactively as the user modifies the cron expression. The preview SHALL display each occurrence formatted as `M/D (Weekday) HH:mm` with the current timezone indicated. The preview SHALL NOT appear when the schedule type is "Fixed Interval".

#### Scenario: Valid calendar expression shows preview

- **WHEN** user selects Calendar Interval and enters a valid cron expression `0 8 * * *`
- **THEN** the form SHALL display a preview block showing the next 3 occurrences at 08:00 on consecutive days, formatted as `M/D (Weekday) HH:mm`

#### Scenario: Invalid expression hides preview

- **WHEN** user enters an invalid cron expression (e.g., `60 25 * * *`)
- **THEN** the form SHALL NOT display the next-runs preview block

#### Scenario: Fixed Interval has no preview

- **WHEN** user selects Fixed Interval schedule type
- **THEN** the form SHALL NOT display the next-runs preview block

#### Scenario: Expression changes update preview reactively

- **WHEN** user changes the cron expression from `0 8 * * *` to `30 14 * * 1`
- **THEN** the preview SHALL update to show the next 3 occurrences at 14:30 on Mondays

### Requirement: Display next runs in ServiceSummary

The ServiceSummary component SHALL display the next 3 execution times for services with a `CalendarInterval` schedule. The display SHALL appear below the existing schedule description. The display SHALL NOT appear for services with only a `StartInterval` schedule or no schedule.

#### Scenario: Service with CalendarInterval shows next runs

- **WHEN** a service has a CalendarInterval schedule with hour=3 and minute=0
- **THEN** ServiceSummary SHALL display the next 3 occurrences at 03:00 on consecutive days, below the schedule description

#### Scenario: Service with StartInterval has no next-run display

- **WHEN** a service has only a StartInterval schedule
- **THEN** ServiceSummary SHALL NOT display next-run times

#### Scenario: Service with no schedule has no next-run display

- **WHEN** a service has no schedule configured
- **THEN** ServiceSummary SHALL NOT display next-run times
