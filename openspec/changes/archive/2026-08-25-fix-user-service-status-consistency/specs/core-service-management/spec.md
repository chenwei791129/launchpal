## ADDED Requirements

### Requirement: Consistent user service runtime status

For user LaunchAgents, `UserManager.List` and `UserManager.Get` SHALL derive `Status` and `PID` exclusively from launchd information for the service label. Both operations SHALL apply the same classification when the launchd job state remains unchanged. The user-service status path SHALL NOT infer `StatusRunning` or a PID from a process-name or command-line substring scan.

For a non-empty service label, the classification SHALL be:

- launchd reports a positive PID: `StatusRunning` with that PID.
- launchd reports the job as loaded without a positive PID: `StatusLoaded` with PID `0`.
- launchd reports that the label is not loaded: `StatusStopped` with PID `0`.

For an empty service label, both operations SHALL return `StatusUnknown` with PID `0`.

#### Scenario: Running job is consistent between list and detail

- **GIVEN** a user LaunchAgent with a non-empty label for which launchd reports PID `4321`
- **WHEN** the unchanged service is returned by `UserManager.List` and `UserManager.Get`
- **THEN** both results have `StatusRunning` and PID `4321`

#### Scenario: Loaded wrapper command is not attributed an unrelated PID

- **GIVEN** a loaded user LaunchAgent whose program is `open`, launchd reports no PID for its label, and an unrelated process command line contains the substring `open`
- **WHEN** the unchanged service is returned by `UserManager.List` and `UserManager.Get`
- **THEN** both results have `StatusLoaded` and PID `0`
- **THEN** neither result uses the unrelated process PID

#### Scenario: Unloaded job is consistent between list and detail

- **GIVEN** a user LaunchAgent with a non-empty label that launchd reports as not loaded
- **WHEN** the unchanged service is returned by `UserManager.List` and `UserManager.Get`
- **THEN** both results have `StatusStopped` and PID `0`

#### Scenario: Empty label has an unknown status

- **GIVEN** a user LaunchAgent plist whose `Label` value is empty
- **WHEN** the service is returned by `UserManager.List` and `UserManager.Get`
- **THEN** both results have `StatusUnknown` and PID `0`

##### Example: launchd state classification

| Label | launchd state | launchd PID | Expected status | Expected PID |
| ----- | ------------- | ----------- | --------------- | ------------ |
| `com.example.running` | loaded and running | `4321` | `running` | `4321` |
| `com.example.loaded` | loaded, not running | none | `loaded` | `0` |
| `com.example.stopped` | not loaded | none | `stopped` | `0` |
| empty | not queried | none | `unknown` | `0` |
