# env-vars-ui Specification

## Purpose

The Create Service modal and the service Edit form expose an Environment Variables section so users can manage a service's `EnvironmentVariables` plist dictionary as a key-value list. Users can add, remove, and toggle visibility on individual entries; the section sits between the run-options checkboxes and the Schedule section so it is discoverable without scrolling.

## Requirements

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


<!-- @trace
source: add-env-vars-ui
updated: 2026-03-30
code:
  - Makefile
  - frontend/app/components/CreateServiceModal.vue
  - frontend/wailsjs/go/models.ts
  - frontend/app/pages/services/[name].vue
-->

---
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


<!-- @trace
source: add-env-vars-ui
updated: 2026-03-30
code:
  - Makefile
  - frontend/app/components/CreateServiceModal.vue
  - frontend/wailsjs/go/models.ts
  - frontend/app/pages/services/[name].vue
-->

---
### Requirement: Form reset clears environment variables

After successful service creation, the form state SHALL be reset including all environment variable entries.

#### Scenario: Form reset after successful creation

- **WHEN** user successfully creates a service with environment variables
- **THEN** all environment variable entries SHALL be cleared from the form

<!-- @trace
source: add-env-vars-ui
updated: 2026-03-30
code:
  - Makefile
  - frontend/app/components/CreateServiceModal.vue
  - frontend/wailsjs/go/models.ts
  - frontend/app/pages/services/[name].vue
-->

---
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

<!-- @trace
source: mask-env-var-values
updated: 2026-04-11
code:
  - internal/launchctl/user.go
  - frontend/app/pages/services/[name].vue
  - frontend/wailsjs/go/models.ts
  - frontend/vitest.config.ts
  - frontend/vitest.setup.ts
  - frontend/app/components/ScheduleForm.vue
  - internal/launchctl/types.go
  - CHANGELOG.md
  - frontend/package.json.md5
  - frontend/app/components/ServiceSummary.vue
  - frontend/app/composables/useNextOccurrences.ts
  - frontend/package.json
  - frontend/app/types/wails.d.ts
  - README.md
  - frontend/app/components/CreateServiceModal.vue
tests:
  - frontend/app/components/__tests__/ServiceSummary.test.ts
  - internal/launchctl/user_test.go
  - frontend/app/components/__tests__/CreateServiceModal.test.ts
  - frontend/app/pages/services/__tests__/edit-env-masking.test.ts
  - frontend/app/composables/__tests__/useNextOccurrences.test.ts
-->