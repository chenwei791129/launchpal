## ADDED Requirements

### Requirement: Environment variables key-value list in create service form

The create service modal SHALL include an "Environment Variables" section that allows users to configure key-value pairs for the service's `EnvironmentVariables` plist property.

The section SHALL be placed between the checkbox options (Run at Load / Keep Alive) and the Schedule section in the form layout.

#### Scenario: Empty state with no environment variables

- **WHEN** user opens the create service modal
- **THEN** the Environment Variables section SHALL display with no entries and an "Add" button

#### Scenario: Add a new environment variable entry

- **WHEN** user clicks the "Add" button in the Environment Variables section
- **THEN** a new row SHALL appear with two text input fields (key and value) and a delete button

#### Scenario: Add multiple environment variable entries

- **WHEN** user clicks the "Add" button multiple times
- **THEN** each click SHALL append a new key-value row to the list

#### Scenario: Remove an environment variable entry

- **WHEN** user clicks the delete button on an environment variable row
- **THEN** that row SHALL be removed from the list

### Requirement: Environment variables included in service creation payload

The form submission SHALL convert the key-value entries into a `Record<string, string>` and include it as the `environment` field of the `ServiceConfig` passed to the `CreateService` API.

#### Scenario: Submit form with environment variables

- **WHEN** user fills in environment variable entries with key "API_KEY" and value "abc123", and key "DB_URL" and value "postgres://localhost/db"
- **AND** user submits the form
- **THEN** the `ServiceConfig.environment` field SHALL contain `{"API_KEY": "abc123", "DB_URL": "postgres://localhost/db"}`

#### Scenario: Submit form with no environment variables

- **WHEN** user does not add any environment variable entries
- **AND** user submits the form
- **THEN** the `ServiceConfig.environment` field SHALL be undefined or an empty object

#### Scenario: Submit form with empty key entries filtered out

- **WHEN** user has environment variable rows where the key field is empty
- **AND** user submits the form
- **THEN** rows with empty keys SHALL be excluded from the `ServiceConfig.environment` field

### Requirement: Form reset clears environment variables

After successful service creation, the form state SHALL be reset including all environment variable entries.

#### Scenario: Form reset after successful creation

- **WHEN** user successfully creates a service with environment variables
- **THEN** all environment variable entries SHALL be cleared from the form
