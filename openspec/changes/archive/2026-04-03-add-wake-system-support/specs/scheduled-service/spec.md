## ADDED Requirements

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
