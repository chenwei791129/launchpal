## MODIFIED Requirements

### Requirement: Get service details from plist

The system SHALL read a single service by name, loading its plist from `~/Library/LaunchAgents/<name>.plist`.
The system SHALL parse the following fields from the plist: Label, Program, ProgramArguments, RunAtLoad, KeepAlive, ThrottleInterval, StartCalendarInterval, StartInterval, EnvironmentVariables, StandardOutPath, StandardErrorPath, WorkingDirectory.
The system SHALL set `Type` to `"user"` and `ReadOnly` to `false`.
The system SHALL detect plist format (xml or binary) and populate `PlistFormat`.
The system SHALL parse `KeepAlive` into a structured `KeepAliveConfig`: a boolean value SHALL produce `{Enabled: <value>, Mode: "boolean"}`, and a dictionary SHALL produce `{Enabled: true, Mode: "dictionary"}` with each recognized sub-key (`SuccessfulExit`, `Crashed`, `AfterInitialDemand`, `NetworkState`, `PathState`, `OtherJobEnabled`) parsed into the structure. A missing `KeepAlive` key SHALL produce `{Enabled: false}`.
The system SHALL NOT discard the `NetworkState`, `PathState`, or `OtherJobEnabled` sub-keys during parsing, even though they are not editable in the UI.
The system SHALL parse `ThrottleInterval` as an integer number of seconds into a nullable field, leaving it unset when the key is absent.

#### Scenario: Valid XML plist with all fields

- **WHEN** a plist file exists with Label, Program, RunAtLoad, KeepAlive (bool), EnvironmentVariables, StandardOutPath, StandardErrorPath, and WorkingDirectory
- **THEN** the returned Service struct contains all parsed fields, Type is `"user"`, ReadOnly is `false`, and PlistFormat is `"xml"`

#### Scenario: Plist with KeepAlive as dictionary

- **WHEN** a plist contains `KeepAlive` as a dictionary
- **THEN** the Service's KeepAlive field has `Enabled = true`, `Mode = "dictionary"`, and the recognized sub-keys are preserved in the structure

##### Example: KeepAlive dictionary round-trips every sub-key

- **GIVEN** a plist with `KeepAlive = {SuccessfulExit: false, Crashed: true, AfterInitialDemand: false, NetworkState: true, PathState: {"/tmp/flag": true}, OtherJobEnabled: {"com.other.job": true}}`
- **WHEN** the service is read
- **THEN** the KeepAliveConfig has `Mode = "dictionary"`, `SuccessfulExit = false`, `Crashed = true`, `AfterInitialDemand = false`, `NetworkState = true`, `PathState["/tmp/flag"] = true`, and `OtherJobEnabled["com.other.job"] = true`

#### Scenario: Plist with ThrottleInterval

- **WHEN** a plist contains `ThrottleInterval` with value 30
- **THEN** the Service's ThrottleInterval field equals 30

#### Scenario: Nonexistent plist file

- **WHEN** the requested service name has no corresponding plist file
- **THEN** the system returns an error indicating the file could not be read

### Requirement: Write plist from ServiceConfig

The system SHALL serialize a `ServiceConfig` into XML plist format and write it to the specified path with `0644` permissions.
The system SHALL only include fields that are set (non-zero/non-empty).
The system SHALL expand `~` in StdoutPath and StderrPath to the user's home directory before writing.
The system SHALL write `StartCalendarInterval` as an empty dict when Schedule is set with no Interval and no calendar fields (launchd interprets this as "every minute").
The system SHALL write `KeepAlive` based on the `KeepAliveConfig`: omit the key when `Enabled` is false; write `KeepAlive = true` when `Mode` is `"boolean"`; write `KeepAlive` as a dictionary when `Mode` is `"dictionary"`, including only the non-null boolean sub-keys and non-empty map sub-keys (`SuccessfulExit`, `Crashed`, `AfterInitialDemand`, `NetworkState`, `PathState`, `OtherJobEnabled`). When `Mode` is `"dictionary"` but no such sub-key is present, the system SHALL write the boolean form `KeepAlive = true` and SHALL NOT write an empty `KeepAlive` dictionary.
The system SHALL write `ThrottleInterval` only when it is explicitly set, and SHALL omit it otherwise.

#### Scenario: Minimal service config

- **WHEN** a ServiceConfig has only Label and Program
- **THEN** the written plist contains only Label and Program keys

#### Scenario: Config with boolean KeepAlive

- **WHEN** a ServiceConfig has KeepAlive with `Enabled = true` and `Mode = "boolean"`
- **THEN** the written plist contains `KeepAlive` with value `true`

#### Scenario: Config with dictionary KeepAlive

- **WHEN** a ServiceConfig has KeepAlive with `Enabled = true`, `Mode = "dictionary"`, and `SuccessfulExit = false`
- **THEN** the written plist contains `KeepAlive` as a dictionary with `SuccessfulExit` set to `false`

#### Scenario: Dictionary KeepAlive preserves non-editable sub-keys

- **WHEN** a ServiceConfig has KeepAlive with `Mode = "dictionary"` and a non-empty `PathState` map carried over from a read
- **THEN** the written plist's `KeepAlive` dictionary contains the same `PathState` entries

#### Scenario: Dictionary KeepAlive with no sub-keys writes boolean true

- **WHEN** a ServiceConfig has KeepAlive with `Enabled = true`, `Mode = "dictionary"`, and no boolean sub-key set and no `NetworkState`/`PathState`/`OtherJobEnabled`
- **THEN** the written plist contains `KeepAlive` with the boolean value `true`
- **AND** the written plist does NOT contain an empty `KeepAlive` dictionary

#### Scenario: Config with ThrottleInterval

- **WHEN** a ServiceConfig has ThrottleInterval set to 10
- **THEN** the written plist contains `ThrottleInterval` with value 10

#### Scenario: Config without ThrottleInterval

- **WHEN** a ServiceConfig has ThrottleInterval unset
- **THEN** the written plist does not contain a `ThrottleInterval` key

#### Scenario: Config with schedule (StartInterval)

- **WHEN** a ServiceConfig has Schedule with Interval=60
- **THEN** the written plist contains `StartInterval` key with value 60

#### Scenario: Config with calendar schedule

- **WHEN** a ServiceConfig has Schedule with Hour=9 and Minute=30
- **THEN** the written plist contains `StartCalendarInterval` with `{Hour: 9, Minute: 30}`

#### Scenario: Empty calendar schedule

- **WHEN** a ServiceConfig has Schedule set but with no Interval and no calendar fields
- **THEN** the written plist contains `StartCalendarInterval` as an empty dict

### Requirement: Clone a user service

The system SHALL provide a clone action on every user service detail view that creates a new user service whose configuration is derived from the source service.

The clone action SHALL be available only for services whose `Type` is `"user"`. The system SHALL NOT expose the clone action on system service or apple-system service detail views.

When the clone action is triggered, the system SHALL open the existing service creation form pre-filled with the source service's `Program`, `ProgramArguments`, `WorkingDirectory`, `KeepAlive` (including its advanced sub-keys), `ThrottleInterval`, `EnvironmentVariables`, `Schedule`, and `WakeSystem` values. The system SHALL leave the `Label` input empty. The system SHALL set the launch-policy selection so that no `RunAtLoad` is written on submission unless KeepAlive is preserved: when the source's launch policy is `Keep Alive`, the clone SHALL preserve `Keep Alive`; otherwise the clone SHALL default to `On Demand` regardless of the source's `RunAtLoad` value.

The system SHALL require the user to provide a new `Label` before submission. The system SHALL submit the cloned configuration through the existing user-service creation path (`CreateService`), so log paths are re-composed from the new label and the user's configured log directory.

When the submitted label conflicts with an existing user service, the system SHALL surface the backend's `service <label> already exists` error inline in the creation form and SHALL NOT close the form, SHALL NOT reset the user-entered fields, and SHALL NOT create any file.

When the clone succeeds, the system SHALL navigate to the new service's detail view at `/services/<new-label>?type=user`.

#### Scenario: Clone action visibility by service type

- **WHEN** the user opens a detail view at `/services/<label>?type=user`
- **THEN** the header action area renders a Copy button next to the existing Start/Stop/Restart/Run Now buttons
- **AND** when the user opens a detail view at `/services/<label>?type=system` or `/services/<label>?type=apple-system`, no Copy button is rendered

#### Scenario: Pre-filled creation form on clone

- **WHEN** the user clicks the Copy button on a user service whose configuration is fully populated
- **THEN** the creation form opens with all of `Program`, `ProgramArguments`, `WorkingDirectory`, `KeepAlive`, `ThrottleInterval`, `EnvironmentVariables`, `Schedule`, `WakeSystem` set to the source service's values
- **AND** the `Label` input is empty
- **AND** the launch-policy radio is not set to `Run at Load`: a source with `Keep Alive` stays on `Keep Alive`, and any other source defaults to `On Demand`

##### Example: Cloning `com.example.ticker`

- **GIVEN** source service `com.example.ticker` has Program=`/usr/bin/foo`, ProgramArguments=`["--port=8080"]`, EnvironmentVariables=`{LOG_LEVEL: "debug"}`, launch policy `Keep Alive` (KeepAlive boolean), Schedule=`StartInterval(60)`
- **WHEN** the user clicks Copy
- **THEN** the form opens with Program=`/usr/bin/foo`, Arguments text=`--port=8080`, EnvironmentVariables row `LOG_LEVEL=debug`, the launch-policy radio on `Keep Alive`, Schedule=`StartInterval(60)`
- **AND** the Label input is empty

#### Scenario: Successful clone creates new service and navigates

- **GIVEN** the user has the creation form open with a prefilled clone of `com.example.ticker`
- **WHEN** the user enters `com.example.ticker-staging` as the label and submits
- **THEN** the system writes `~/Library/LaunchAgents/com.example.ticker-staging.plist` containing the cloned configuration
- **AND** because the clone's launch policy is `Keep Alive`, the written plist contains a `KeepAlive` key and does NOT contain a standalone `RunAtLoad` key (launchd implies it)
- **AND** the browser navigates to `/services/com.example.ticker-staging?type=user`
- **AND** the source service `com.example.ticker` and its plist file remain unchanged

#### Scenario: Duplicate label is rejected inline

- **GIVEN** a user service `com.example.ticker-staging` already exists
- **WHEN** the user submits a clone with the same label `com.example.ticker-staging`
- **THEN** the form remains open with all entered fields preserved
- **AND** an inline error message `service com.example.ticker-staging already exists` is shown
- **AND** no plist file is created or modified
- **AND** no navigation occurs

#### Scenario: User selects Run at Load before submitting

- **GIVEN** the user has the creation form open with a prefilled clone defaulting to `On Demand`
- **WHEN** the user selects the `Run at Load` launch-policy radio and submits with a new label
- **THEN** the resulting plist contains `RunAtLoad = true`
- **AND** the default `On Demand` selection is only the initial state, not a submission-time constraint

## ADDED Requirements

### Requirement: Launch policy selection in service creation form

The service creation form and the service edit form SHALL present launch behavior as a single mutually-exclusive radio group named "Launch Policy" with exactly three options: `On Demand`, `Run at Load`, and `Keep Alive`. The forms SHALL NOT present `Run at Load` and `Keep Alive` as two independent checkboxes.
On submission, the system SHALL map the selected option as follows: `On Demand` writes neither `RunAtLoad` nor `KeepAlive`; `Run at Load` writes `RunAtLoad = true` and no `KeepAlive`; `Keep Alive` writes a `KeepAlive` value and SHALL NOT additionally write `RunAtLoad`, because launchd implies `RunAtLoad` from `KeepAlive`.
When loading an existing service into either form, the system SHALL map the plist back to a radio selection by KeepAlive precedence: when the parsed `KeepAlive` is enabled the selection SHALL be `Keep Alive` regardless of `RunAtLoad`; otherwise when `RunAtLoad` is true the selection SHALL be `Run at Load`; otherwise the selection SHALL be `On Demand`. A legacy service carrying both `RunAtLoad = true` and a `KeepAlive` value SHALL therefore load as `Keep Alive`, and on the next save SHALL NOT emit a standalone `RunAtLoad` key.
All launch-policy labels and helper text SHALL be in English.

#### Scenario: Three mutually-exclusive options

- **WHEN** the user opens the service creation form
- **THEN** a "Launch Policy" radio group is rendered with the options `On Demand`, `Run at Load`, and `Keep Alive`, exactly one of which is selectable at a time

#### Scenario: Launch policy maps to plist keys

- **WHEN** the user selects a launch policy and submits an otherwise minimal config
- **THEN** the written plist contains `RunAtLoad` and `KeepAlive` keys exactly as specified in the table below

##### Example: launch policy to plist keys

| Selected policy | RunAtLoad written | KeepAlive written |
| --------------- | ----------------- | ----------------- |
| On Demand       | no                | no                |
| Run at Load     | yes (`true`)      | no                |
| Keep Alive      | no                | yes               |

#### Scenario: Legacy plist with both RunAtLoad and KeepAlive loads as Keep Alive

- **GIVEN** an existing service whose plist contains both `RunAtLoad = true` and `KeepAlive = {SuccessfulExit: false}`
- **WHEN** the user opens it in the edit form
- **THEN** the launch-policy radio is set to `Keep Alive`
- **AND** when the user saves without further changes, the written plist contains the `KeepAlive` dictionary and does NOT contain a standalone `RunAtLoad` key

### Requirement: KeepAlive advanced options

When the launch-policy selection is `Keep Alive`, the service creation form SHALL reveal an advanced options section. The section SHALL allow the user to choose between a boolean `Keep Alive` and a dictionary form, and SHALL provide editable controls for `SuccessfulExit`, `Crashed`, and `AfterInitialDemand`. The form SHALL provide an editable integer control for `ThrottleInterval` in the same advanced section.
The advanced section SHALL display informational text stating that multiple dictionary conditions are combined with OR semantics and that `Keep Alive` implies `Run at Load`.
The form SHALL NOT render editable controls for `NetworkState`, `PathState`, or `OtherJobEnabled`; the system SHALL preserve any such values read from an existing plist and write them back unchanged.
When the dictionary form is selected but no effective sub-key is set (no editable boolean set and no preserved `NetworkState`/`PathState`/`OtherJobEnabled`), the system SHALL write `KeepAlive = true` (boolean form) rather than an empty dictionary. The system SHALL NOT write an empty `KeepAlive` dictionary.
When the launch-policy selection is not `Keep Alive`, the advanced options section SHALL be hidden. All advanced-option labels and helper text SHALL be in English.

#### Scenario: Advanced section visibility follows launch policy

- **WHEN** the user selects the `Keep Alive` launch policy
- **THEN** the advanced KeepAlive options section is shown
- **AND** when the user selects `On Demand` or `Run at Load`, the advanced section is hidden

#### Scenario: Editing dictionary sub-keys produces a dictionary KeepAlive

- **GIVEN** the user has selected `Keep Alive` and switched to the dictionary form
- **WHEN** the user sets `SuccessfulExit` to false and submits
- **THEN** the written plist contains `KeepAlive` as a dictionary with `SuccessfulExit` set to `false`

#### Scenario: ThrottleInterval edited in advanced section

- **WHEN** the user enters `15` in the ThrottleInterval control and submits
- **THEN** the written plist contains `ThrottleInterval` with value 15

#### Scenario: Non-editable sub-keys are preserved through edit

- **GIVEN** an existing service whose `KeepAlive` dictionary contains `PathState = {"/tmp/flag": true}`
- **WHEN** the user opens it for editing, changes only `SuccessfulExit`, and submits
- **THEN** the written plist's `KeepAlive` dictionary still contains `PathState = {"/tmp/flag": true}`

#### Scenario: Dictionary form with no effective sub-key downgrades to boolean

- **GIVEN** the user has selected `Keep Alive` and the dictionary form
- **WHEN** the user clears every editable sub-key (and no preserved `NetworkState`/`PathState`/`OtherJobEnabled` remains) and submits
- **THEN** the written plist contains `KeepAlive` with the boolean value `true` and does NOT contain an empty `KeepAlive` dictionary
