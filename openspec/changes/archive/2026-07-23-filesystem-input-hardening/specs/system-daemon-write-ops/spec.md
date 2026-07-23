## ADDED Requirements

### Requirement: Helper invokes launchctl by absolute path

The helper SHALL invoke `launchctl` by its absolute path `/bin/launchctl` for the bootstrap, bootout, and kickstart operations, rather than by the bare name `launchctl` resolved through `$PATH`. This makes the resolved binary independent of any inherited environment.

#### Scenario: launchctl resolved by absolute path

- **WHEN** the helper performs a bootstrap, bootout, or kickstart
- **THEN** it executes `/bin/launchctl`, regardless of the `$PATH` value inherited by the helper process

### Requirement: System daemon schedule validation parity

System daemon create and update SHALL apply the same schedule range validation as user services: `StartInterval` SHALL be at least 10, and calendar-entry fields SHALL be within range (minute 0-59, hour 0-23, day 1-31, weekday 0-6, month 1-12). In addition, the system-domain create and update path SHALL reject a schedule whose calendar-entry count exceeds the 50-entry cap (matching the cron range-expansion limit). Both checks SHALL be performed in the create/update path (which returns an error and writes no plist on failure), not in the error-less plist encoder, so the enforcement holds for every caller of the system create/update binding rather than only in the frontend form.

#### Scenario: Out-of-range system daemon schedule is rejected

- **WHEN** a system daemon create or update is called with an out-of-range calendar field or a `StartInterval` below 10
- **THEN** it returns a validation error and does not write the plist

#### Scenario: Over-cap system daemon schedule is rejected in the create/update path

- **WHEN** a system daemon create or update is called with more than 50 calendar entries
- **THEN** the create/update path returns a validation error and writes no plist, independently of the frontend

##### Example: system-domain schedule validation

| Input                                   | Result            |
| --------------------------------------- | ----------------- |
| StartInterval = 9                       | rejected          |
| StartInterval = 10                      | accepted          |
| calendar Hour = 24                      | rejected          |
| calendar Hour = 23                      | accepted          |
| expansion producing 51 calendar entries | rejected          |
| expansion producing 50 calendar entries | accepted          |
