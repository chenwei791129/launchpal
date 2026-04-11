## ADDED Requirements

### Requirement: Environment variable values are masked by default

All environment variable values SHALL be displayed as masked characters (`••••••••`) by default across all UI surfaces: ServiceSummary (read-only), service edit form, and create service modal.

Each environment variable row SHALL include a toggle button with an eye icon to reveal or hide the value.

#### Scenario: Value masked by default in ServiceSummary

- **WHEN** user views a service with environment variables in the Summary tab
- **THEN** each environment variable SHALL display as `KEY=••••••••` instead of showing the actual value

#### Scenario: Value masked by default in edit form

- **WHEN** user opens the Edit tab for a service with environment variables
- **THEN** each environment variable value input SHALL use `type="password"` to mask the value

#### Scenario: Value masked by default in create modal

- **WHEN** user opens the create service modal and adds environment variable entries
- **THEN** each environment variable value input SHALL use `type="password"` to mask the value

#### Scenario: Toggle value visibility in ServiceSummary

- **WHEN** user clicks the eye icon button next to a masked environment variable in the Summary tab
- **THEN** that variable's value SHALL be revealed as plain text
- **AND** the eye icon SHALL change to indicate the value is currently visible

#### Scenario: Toggle value back to masked in ServiceSummary

- **WHEN** user clicks the eye icon button next to a revealed environment variable in the Summary tab
- **THEN** that variable's value SHALL be masked again as `••••••••`
- **AND** the eye icon SHALL change to indicate the value is currently hidden

#### Scenario: Toggle value visibility in edit form

- **WHEN** user clicks the eye icon button next to an environment variable input in the Edit tab
- **THEN** that input's type SHALL change from `password` to `text` to reveal the value

#### Scenario: Toggle value visibility in create modal

- **WHEN** user clicks the eye icon button next to an environment variable input in the create service modal
- **THEN** that input's type SHALL change from `password` to `text` to reveal the value

#### Scenario: Each variable toggles independently

- **WHEN** user reveals the value of one environment variable
- **THEN** all other environment variables SHALL remain masked
