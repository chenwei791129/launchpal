## ADDED Requirements

### Requirement: Status multi-select dropdown filter

The system SHALL display a "Status" multi-select dropdown in the filter bar on each service list page (User, System, Apple System).
The dropdown SHALL provide four options: Running, Loaded, Unloaded, Unknown.
The dropdown SHALL default to "All" with no options selected, displaying all services regardless of status.
When one or more options are selected, the system SHALL display only services whose status matches any selected option (OR logic).
The status mapping SHALL be:
- "Running" matches `status === 'running'`
- "Loaded" matches `status === 'loaded'`
- "Unloaded" matches `status === 'stopped'`
- "Unknown" matches `status === 'unknown'`

#### Scenario: No status filter selected

- **WHEN** no status options are selected in the dropdown
- **THEN** all services are displayed regardless of their status

##### Example: empty selection means All, not zero matches

- **GIVEN** four services with statuses `running`, `loaded`, `stopped`, and `unknown`
- **WHEN** the Status dropdown has no option selected
- **THEN** all four services are displayed

#### Scenario: Single status selected

- **WHEN** the user selects "Running" in the Status dropdown
- **THEN** only services with `status === 'running'` are displayed

#### Scenario: Multiple statuses selected with OR logic

- **WHEN** the user selects both "Running" and "Loaded" in the Status dropdown
- **THEN** services with `status === 'running'` OR `status === 'loaded'` are displayed

#### Scenario: Unloaded maps to stopped

- **WHEN** the user selects "Unloaded" in the Status dropdown
- **THEN** only services with `status === 'stopped'` are displayed

### Requirement: Type multi-select dropdown filter

The system SHALL display a "Type" multi-select dropdown in the filter bar on each service list page (User, System, Apple System).
The dropdown SHALL provide four options: Scheduled, KeepAlive, RunAtLoad, None.
The dropdown SHALL default to "All" with no options selected, displaying all services regardless of type.
When one or more options are selected, the system SHALL display only services matching any selected option (OR logic).
A service SHALL match "Scheduled" when `service.schedule` is defined (non-null, non-undefined).
A service SHALL match "KeepAlive" when `service.keepAlive?.enabled === true`.
A service SHALL match "RunAtLoad" when `service.runAtLoad === true`.
A service SHALL match "None" when `service.schedule` is not defined AND `service.keepAlive?.enabled` is not true AND `service.runAtLoad` is not true.
A single service MAY match more than one of "Scheduled", "KeepAlive", and "RunAtLoad" simultaneously.
The option set SHALL cover the same three launch-policy badges `ServiceRow.vue` renders in the Type column (Scheduled / KeepAlive / RunAtLoad), so a filter selection never contradicts the badge the user can see. Unlike the badge, which picks a single label by precedence, the filter matches every applicable option independently.

#### Scenario: No type filter selected

- **WHEN** no type options are selected in the dropdown
- **THEN** all services are displayed regardless of their schedule type

##### Example: empty selection means All, not zero matches

- **GIVEN** four services: one scheduled, one with `keepAlive.enabled === true`, one with `runAtLoad === true`, and one with none of those
- **WHEN** the Type dropdown has no option selected
- **THEN** all four services are displayed

#### Scenario: Scheduled filter selected

- **WHEN** the user selects "Scheduled" in the Type dropdown
- **THEN** only services with a defined `schedule` property are displayed

#### Scenario: KeepAlive filter selected

- **WHEN** the user selects "KeepAlive" in the Type dropdown
- **THEN** only services with `keepAlive.enabled === true` are displayed

##### Example: KeepAlive service is not classified as None

- **GIVEN** a service with no `schedule`, `keepAlive.enabled === true`, and `runAtLoad === false`
- **WHEN** the user selects "KeepAlive" in the Type dropdown
- **THEN** the service is displayed
- **AND WHEN** the user selects "None" instead
- **THEN** the service is NOT displayed

#### Scenario: RunAtLoad filter selected

- **WHEN** the user selects "RunAtLoad" in the Type dropdown
- **THEN** only services with `runAtLoad === true` are displayed

#### Scenario: Service matches both Scheduled and RunAtLoad

- **WHEN** the user selects "Scheduled" in the Type dropdown
- **AND** a service has both `schedule` defined and `runAtLoad === true`
- **THEN** the service is displayed

#### Scenario: None filter excludes scheduled, keepAlive and runAtLoad services

- **WHEN** the user selects "None" in the Type dropdown
- **THEN** only services with no schedule, `keepAlive.enabled` not true, and `runAtLoad` not true are displayed

### Requirement: Cross-filter AND logic

The system SHALL apply AND logic between the Status dropdown, Type dropdown, and the existing text search.
A service SHALL be displayed only when it matches ALL active filters simultaneously.

#### Scenario: Status and Type combined

- **WHEN** the user selects "Running" in Status and "Scheduled" in Type
- **THEN** only services that are both running AND have a schedule are displayed

#### Scenario: Filters combined with text search

- **WHEN** the user selects "Running" in Status, "Scheduled" in Type, and types "com.apple" in search
- **THEN** only services matching all three conditions (running AND scheduled AND name contains "com.apple") are displayed

#### Scenario: No results from combined filters

- **WHEN** the active filter combination matches zero services
- **THEN** the system SHALL display an empty state message

##### Example: filtered-to-empty reads differently from having no services

- **GIVEN** a list containing only `running` services that all have a schedule
- **WHEN** the user selects "Unknown" in Status and "None" in Type
- **THEN** zero services are displayed
- **AND** the empty state reads "No services match the selected filters", not "No services found"
- **AND** on the User services page the "Create your first service" button is NOT shown, because the list is empty due to the filters rather than because the user has no services

### Requirement: Filter bar placement and shared component

The system SHALL render the filter bar between the page header (containing title and search) and the table header row (STATUS / NAME / TYPE / ACTIONS).
The filter bar component SHALL be reusable across all three service list surfaces: `pages/index.vue` (User services), `pages/system.vue` (System services), and `components/ReadOnlyServiceList.vue` (Apple System services).
Note: `system.vue` is a standalone page, not a `ReadOnlyServiceList` consumer — it was split back out when Admin Mode added its banner, New Service button, and delete-with-logs dialog. Only `apple-system.vue` delegates to `ReadOnlyServiceList`. Each of the three surfaces therefore needs the filter bar wired up individually.

#### Scenario: Filter bar visible on User services page

- **WHEN** the user navigates to the User services page
- **THEN** the filter bar with Status and Type dropdowns is displayed between header and table header

##### Example: implemented in pages/index.vue

- **GIVEN** the User services page is implemented by `frontend/app/pages/index.vue`
- **WHEN** the page renders
- **THEN** `ServiceFilterBar` appears after the closing `</header>` and before the table header row
- **AND** the page filters through the shared `filterServices` helper rather than its own predicate

#### Scenario: Filter bar visible on System services page

- **WHEN** the user navigates to the System services page
- **THEN** the filter bar with Status and Type dropdowns is displayed between header and table header

##### Example: implemented in pages/system.vue, NOT via ReadOnlyServiceList

- **GIVEN** the System services page is implemented by `frontend/app/pages/system.vue` as a standalone page
- **AND** it does NOT delegate to `ReadOnlyServiceList.vue` (Admin Mode split it back out)
- **WHEN** the page renders
- **THEN** `ServiceFilterBar` appears after the closing `</header>` and before the table header row
- **AND** the page filters through the shared `filterServices` helper rather than its own predicate

#### Scenario: Filter bar visible on Apple System services page

- **WHEN** the user navigates to the Apple System services page
- **THEN** the filter bar with Status and Type dropdowns is displayed between header and table header

##### Example: implemented via ReadOnlyServiceList

- **GIVEN** `frontend/app/pages/apple-system.vue` delegates to `frontend/app/components/ReadOnlyServiceList.vue`
- **WHEN** the page renders
- **THEN** `ServiceFilterBar` appears in `ReadOnlyServiceList.vue` after the closing `</header>` and before the table header row
- **AND** the component filters through the shared `filterServices` helper rather than its own predicate
