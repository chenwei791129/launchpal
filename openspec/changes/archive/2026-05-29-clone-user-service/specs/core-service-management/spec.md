## ADDED Requirements

### Requirement: Clone a user service

The system SHALL provide a clone action on every user service detail view that creates a new user service whose configuration is derived from the source service.

The clone action SHALL be available only for services whose `Type` is `"user"`. The system SHALL NOT expose the clone action on system service or apple-system service detail views.

When the clone action is triggered, the system SHALL open the existing service creation form pre-filled with the source service's `Program`, `ProgramArguments`, `WorkingDirectory`, `KeepAlive`, `EnvironmentVariables`, `Schedule`, and `WakeSystem` values. The system SHALL leave the `Label` input empty and SHALL set `RunAtLoad` to `false` regardless of the source service's `RunAtLoad` value.

The system SHALL require the user to provide a new `Label` before submission. The system SHALL submit the cloned configuration through the existing user-service creation path (`CreateService`), so log paths are re-composed from the new label and the user's configured log directory.

When the submitted label conflicts with an existing user service, the system SHALL surface the backend's `service <label> already exists` error inline in the creation form and SHALL NOT close the form, SHALL NOT reset the user-entered fields, and SHALL NOT create any file.

When the clone succeeds, the system SHALL navigate to the new service's detail view at `/services/<new-label>?type=user`.

#### Scenario: Clone action visibility by service type

- **WHEN** the user opens a detail view at `/services/<label>?type=user`
- **THEN** the header action area renders a Copy button next to the existing Start/Stop/Restart/Run Now buttons
- **AND** when the user opens a detail view at `/services/<label>?type=system` or `/services/<label>?type=apple-system`, no Copy button is rendered

#### Scenario: Pre-filled creation form on clone

- **WHEN** the user clicks the Copy button on a user service whose configuration is fully populated
- **THEN** the creation form opens with all of `Program`, `ProgramArguments`, `WorkingDirectory`, `KeepAlive`, `EnvironmentVariables`, `Schedule`, `WakeSystem` set to the source service's values
- **AND** the `Label` input is empty
- **AND** the `RunAtLoad` checkbox is unchecked even if the source service has `RunAtLoad = true`

##### Example: Cloning `com.example.ticker`

- **GIVEN** source service `com.example.ticker` has Program=`/usr/bin/foo`, ProgramArguments=`["--port=8080"]`, EnvironmentVariables=`{LOG_LEVEL: "debug"}`, RunAtLoad=`true`, KeepAlive=`true`, Schedule=`StartInterval(60)`
- **WHEN** the user clicks Copy
- **THEN** the form opens with Program=`/usr/bin/foo`, Arguments text=`--port=8080`, EnvironmentVariables row `LOG_LEVEL=debug`, KeepAlive=checked, Schedule=`StartInterval(60)`
- **AND** the Label input is empty
- **AND** the RunAtLoad checkbox is unchecked

#### Scenario: Successful clone creates new service and navigates

- **GIVEN** the user has the creation form open with a prefilled clone of `com.example.ticker`
- **WHEN** the user enters `com.example.ticker-staging` as the label and submits
- **THEN** the system writes `~/Library/LaunchAgents/com.example.ticker-staging.plist` containing the cloned configuration with `RunAtLoad = false`
- **AND** the browser navigates to `/services/com.example.ticker-staging?type=user`
- **AND** the source service `com.example.ticker` and its plist file remain unchanged

#### Scenario: Duplicate label is rejected inline

- **GIVEN** a user service `com.example.ticker-staging` already exists
- **WHEN** the user submits a clone with the same label `com.example.ticker-staging`
- **THEN** the form remains open with all entered fields preserved
- **AND** an inline error message `service com.example.ticker-staging already exists` is shown
- **AND** no plist file is created or modified
- **AND** no navigation occurs

#### Scenario: User overrides RunAtLoad before submitting

- **GIVEN** the user has the creation form open with a prefilled clone
- **WHEN** the user manually re-checks the `Run at Load` checkbox and submits with a new label
- **THEN** the resulting plist contains `RunAtLoad = true`
- **AND** the forced default to `false` is only the initial state of the checkbox, not a submission-time constraint
