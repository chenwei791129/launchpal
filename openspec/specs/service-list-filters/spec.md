# service-list-filters Specification

## Purpose

TBD - created by archiving change 'service-list-filters'. Update Purpose after archive.

## Requirements

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


<!-- @trace
source: service-list-filters
updated: 2026-08-26
code:
  - frontend/app/utils/launchPolicy.ts
  - internal/launchctl/plist_encode.go
  - frontend/app/utils/settingsValidation.ts
  - frontend/app/composables/useNextOccurrences.ts
  - CHANGELOG.md
  - .agents/skills/spectra-audit/SKILL.md
  - internal/launchctl/readonly.go
  - admin_mode.go
  - cmd/launchpal-privhelper/procinfo_darwin.go
  - frontend/package.json.md5
  - frontend/wailsjs/go/main/App.js
  - Makefile
  - cmd/launchpal-privhelper/main.go
  - frontend/app/pages/index.vue
  - frontend/pnpm-workspace.yaml
  - internal/privhelper/logpath.go
  - frontend/app/components/StatusConfidenceIcon.vue
  - frontend/app/components/ServiceRow.vue
  - internal/privhelper/handlers.go
  - internal/settings/settings.go
  - .github/workflows/build.yml
  - internal/launchctl/status_detect.go
  - frontend/app/components/ReadOnlyServiceList.vue
  - internal/privhelper/install.go
  - internal/launchctl/nofollow_other.go
  - internal/launchctl/nofollow_darwin.go
  - frontend/app/components/LaunchPolicyForm.vue
  - frontend/app/composables/useBackupDiff.ts
  - frontend/app/components/ServiceLogs.vue
  - internal/plistutil/plistutil.go
  - internal/launchctl/system.go
  - .agents/skills/spectra-discuss/SKILL.md
  - frontend/app/utils/formatters.ts
  - frontend/app/components/CreateServiceModal.vue
  - internal/launchctl/keepalive.go
  - internal/privhelper/peer_other.go
  - .github/workflows/release-please.yml
  - internal/launchctl/user.go
  - CLAUDE.md
  - frontend/app/composables/useServiceListFilters.ts
  - internal/launchctl/apple_system.go
  - frontend/app/pages/settings.vue
  - frontend/app/composables/useSettings.ts
  - frontend/app/utils/logPaths.ts
  - .agents/skills/spectra-debug/SKILL.md
  - frontend/wailsjs/go/models.ts
  - frontend/app/components/ServiceSummary.vue
  - internal/privhelper/logpath_other.go
  - internal/privhelper/integrity.go
  - internal/privhelper/nofollow_darwin.go
  - go.mod
  - frontend/app/utils/serviceFilters.ts
  - .agents/skills/spectra-archive/SKILL.md
  - cmd/launchpal-privhelper/procinfo_other.go
  - .agents/skills/spectra-apply/SKILL.md
  - frontend/wailsjs/go/main/App.d.ts
  - frontend/app/components/LogStorageSection.vue
  - frontend/app/pages/services/[name].vue
  - frontend/app/utils/serviceValidation.ts
  - frontend/app/components/ServiceFilterDropdown.vue
  - frontend/app/types/wails.d.ts
  - .agents/skills/spectra-commit/SKILL.md
  - internal/privhelper/logpath_darwin.go
  - frontend/app/components/ScheduleForm.vue
  - internal/privhelper/peer_darwin.go
  - .agents/skills/spectra-ingest/SKILL.md
  - frontend/package.json
  - internal/launchctl/manager.go
  - internal/privhelper/client.go
  - frontend/app/utils/ansiToHtml.ts
  - .agents/skills/spectra-propose/SKILL.md
  - main.go
  - frontend/app/components/ServiceFilterBar.vue
  - frontend/app/composables/useAdminMode.ts
  - frontend/app/utils/serviceToConfig.ts
  - .agents/skills/spectra-drift/SKILL.md
  - .spectra.yaml
  - frontend/nuxt.config.ts
  - go.sum
  - internal/backup/backup.go
  - frontend/app/components/BackupDiffDialog.vue
  - frontend/eslint.config.mjs
  - internal/privhelper/protocol.go
  - frontend/vitest.setup.ts
  - frontend/vitest.config.ts
  - frontend/app/pages/system.vue
  - internal/privhelper/server.go
  - app.go
  - .agents/skills/spectra-ask/SKILL.md
  - frontend/app/components/StatusBar.vue
  - AGENTS.md
  - internal/launchctl/types.go
  - frontend/app/components/InlineBanner.vue
  - internal/privhelper/nofollow_other.go
  - README.md
tests:
  - frontend/app/pages/services/__tests__/edit-program-arguments-validation.test.ts
  - frontend/app/components/__tests__/ServiceLogs.test.ts
  - internal/privhelper/client_test.go
  - admin_mode_testhelpers_test.go
  - cmd/launchpal-privhelper/helper_test.go
  - internal/privhelper/protocol_test.go
  - frontend/app/pages/__tests__/serviceListFilterBar.test.ts
  - frontend/app/components/__tests__/CreateServiceModal.test.ts
  - frontend/app/composables/__tests__/useServiceListFilters.test.ts
  - frontend/app/composables/__tests__/useBackupDiff.test.ts
  - frontend/app/utils/__tests__/serviceFilters.test.ts
  - frontend/app/components/__tests__/CloneUserService.test.ts
  - internal/launchctl/types_test.go
  - internal/settings/settings_test.go
  - frontend/app/pages/services/__tests__/edit-env-masking.test.ts
  - internal/privhelper/handlers_test.go
  - frontend/app/components/__tests__/ServiceRow.test.ts
  - internal/privhelper/install_test.go
  - frontend/app/pages/services/__tests__/edit-launch-policy.test.ts
  - resolve_helper_test.go
  - internal/launchctl/readonly_test.go
  - internal/privhelper/server_test.go
  - internal/privhelper/integrity_test.go
  - frontend/app/utils/__tests__/serviceToConfig.test.ts
  - frontend/app/components/__tests__/LaunchPolicyForm.test.ts
  - internal/backup/backup_test.go
  - internal/launchctl/user_test.go
  - internal/launchctl/apple_system_test.go
  - admin_mode_test.go
  - frontend/app/components/__tests__/ServiceSummary.test.ts
  - internal/launchctl/keepalive_test.go
  - internal/plistutil/plistutil_test.go
  - frontend/app/utils/__tests__/ansiToHtml.test.ts
  - frontend/app/components/__tests__/ServiceFilterBar.test.ts
  - frontend/app/components/__tests__/BackupDiffDialog.test.ts
  - frontend/app/composables/__tests__/useSettings.test.ts
  - app_test.go
  - frontend/app/utils/__tests__/launchPolicy.test.ts
  - frontend/app/utils/__tests__/formatters.test.ts
  - frontend/app/pages/__tests__/settings.test.ts
  - internal/launchctl/system_test.go
  - internal/launchctl/status_detect_test.go
  - internal/launchctl/plist_encode_test.go
  - frontend/app/composables/__tests__/useNextOccurrences.test.ts
  - internal/plistutil/testhelpers_test.go
-->

---
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


<!-- @trace
source: service-list-filters
updated: 2026-08-26
code:
  - frontend/app/utils/launchPolicy.ts
  - internal/launchctl/plist_encode.go
  - frontend/app/utils/settingsValidation.ts
  - frontend/app/composables/useNextOccurrences.ts
  - CHANGELOG.md
  - .agents/skills/spectra-audit/SKILL.md
  - internal/launchctl/readonly.go
  - admin_mode.go
  - cmd/launchpal-privhelper/procinfo_darwin.go
  - frontend/package.json.md5
  - frontend/wailsjs/go/main/App.js
  - Makefile
  - cmd/launchpal-privhelper/main.go
  - frontend/app/pages/index.vue
  - frontend/pnpm-workspace.yaml
  - internal/privhelper/logpath.go
  - frontend/app/components/StatusConfidenceIcon.vue
  - frontend/app/components/ServiceRow.vue
  - internal/privhelper/handlers.go
  - internal/settings/settings.go
  - .github/workflows/build.yml
  - internal/launchctl/status_detect.go
  - frontend/app/components/ReadOnlyServiceList.vue
  - internal/privhelper/install.go
  - internal/launchctl/nofollow_other.go
  - internal/launchctl/nofollow_darwin.go
  - frontend/app/components/LaunchPolicyForm.vue
  - frontend/app/composables/useBackupDiff.ts
  - frontend/app/components/ServiceLogs.vue
  - internal/plistutil/plistutil.go
  - internal/launchctl/system.go
  - .agents/skills/spectra-discuss/SKILL.md
  - frontend/app/utils/formatters.ts
  - frontend/app/components/CreateServiceModal.vue
  - internal/launchctl/keepalive.go
  - internal/privhelper/peer_other.go
  - .github/workflows/release-please.yml
  - internal/launchctl/user.go
  - CLAUDE.md
  - frontend/app/composables/useServiceListFilters.ts
  - internal/launchctl/apple_system.go
  - frontend/app/pages/settings.vue
  - frontend/app/composables/useSettings.ts
  - frontend/app/utils/logPaths.ts
  - .agents/skills/spectra-debug/SKILL.md
  - frontend/wailsjs/go/models.ts
  - frontend/app/components/ServiceSummary.vue
  - internal/privhelper/logpath_other.go
  - internal/privhelper/integrity.go
  - internal/privhelper/nofollow_darwin.go
  - go.mod
  - frontend/app/utils/serviceFilters.ts
  - .agents/skills/spectra-archive/SKILL.md
  - cmd/launchpal-privhelper/procinfo_other.go
  - .agents/skills/spectra-apply/SKILL.md
  - frontend/wailsjs/go/main/App.d.ts
  - frontend/app/components/LogStorageSection.vue
  - frontend/app/pages/services/[name].vue
  - frontend/app/utils/serviceValidation.ts
  - frontend/app/components/ServiceFilterDropdown.vue
  - frontend/app/types/wails.d.ts
  - .agents/skills/spectra-commit/SKILL.md
  - internal/privhelper/logpath_darwin.go
  - frontend/app/components/ScheduleForm.vue
  - internal/privhelper/peer_darwin.go
  - .agents/skills/spectra-ingest/SKILL.md
  - frontend/package.json
  - internal/launchctl/manager.go
  - internal/privhelper/client.go
  - frontend/app/utils/ansiToHtml.ts
  - .agents/skills/spectra-propose/SKILL.md
  - main.go
  - frontend/app/components/ServiceFilterBar.vue
  - frontend/app/composables/useAdminMode.ts
  - frontend/app/utils/serviceToConfig.ts
  - .agents/skills/spectra-drift/SKILL.md
  - .spectra.yaml
  - frontend/nuxt.config.ts
  - go.sum
  - internal/backup/backup.go
  - frontend/app/components/BackupDiffDialog.vue
  - frontend/eslint.config.mjs
  - internal/privhelper/protocol.go
  - frontend/vitest.setup.ts
  - frontend/vitest.config.ts
  - frontend/app/pages/system.vue
  - internal/privhelper/server.go
  - app.go
  - .agents/skills/spectra-ask/SKILL.md
  - frontend/app/components/StatusBar.vue
  - AGENTS.md
  - internal/launchctl/types.go
  - frontend/app/components/InlineBanner.vue
  - internal/privhelper/nofollow_other.go
  - README.md
tests:
  - frontend/app/pages/services/__tests__/edit-program-arguments-validation.test.ts
  - frontend/app/components/__tests__/ServiceLogs.test.ts
  - internal/privhelper/client_test.go
  - admin_mode_testhelpers_test.go
  - cmd/launchpal-privhelper/helper_test.go
  - internal/privhelper/protocol_test.go
  - frontend/app/pages/__tests__/serviceListFilterBar.test.ts
  - frontend/app/components/__tests__/CreateServiceModal.test.ts
  - frontend/app/composables/__tests__/useServiceListFilters.test.ts
  - frontend/app/composables/__tests__/useBackupDiff.test.ts
  - frontend/app/utils/__tests__/serviceFilters.test.ts
  - frontend/app/components/__tests__/CloneUserService.test.ts
  - internal/launchctl/types_test.go
  - internal/settings/settings_test.go
  - frontend/app/pages/services/__tests__/edit-env-masking.test.ts
  - internal/privhelper/handlers_test.go
  - frontend/app/components/__tests__/ServiceRow.test.ts
  - internal/privhelper/install_test.go
  - frontend/app/pages/services/__tests__/edit-launch-policy.test.ts
  - resolve_helper_test.go
  - internal/launchctl/readonly_test.go
  - internal/privhelper/server_test.go
  - internal/privhelper/integrity_test.go
  - frontend/app/utils/__tests__/serviceToConfig.test.ts
  - frontend/app/components/__tests__/LaunchPolicyForm.test.ts
  - internal/backup/backup_test.go
  - internal/launchctl/user_test.go
  - internal/launchctl/apple_system_test.go
  - admin_mode_test.go
  - frontend/app/components/__tests__/ServiceSummary.test.ts
  - internal/launchctl/keepalive_test.go
  - internal/plistutil/plistutil_test.go
  - frontend/app/utils/__tests__/ansiToHtml.test.ts
  - frontend/app/components/__tests__/ServiceFilterBar.test.ts
  - frontend/app/components/__tests__/BackupDiffDialog.test.ts
  - frontend/app/composables/__tests__/useSettings.test.ts
  - app_test.go
  - frontend/app/utils/__tests__/launchPolicy.test.ts
  - frontend/app/utils/__tests__/formatters.test.ts
  - frontend/app/pages/__tests__/settings.test.ts
  - internal/launchctl/system_test.go
  - internal/launchctl/status_detect_test.go
  - internal/launchctl/plist_encode_test.go
  - frontend/app/composables/__tests__/useNextOccurrences.test.ts
  - internal/plistutil/testhelpers_test.go
-->

---
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


<!-- @trace
source: service-list-filters
updated: 2026-08-26
code:
  - frontend/app/utils/launchPolicy.ts
  - internal/launchctl/plist_encode.go
  - frontend/app/utils/settingsValidation.ts
  - frontend/app/composables/useNextOccurrences.ts
  - CHANGELOG.md
  - .agents/skills/spectra-audit/SKILL.md
  - internal/launchctl/readonly.go
  - admin_mode.go
  - cmd/launchpal-privhelper/procinfo_darwin.go
  - frontend/package.json.md5
  - frontend/wailsjs/go/main/App.js
  - Makefile
  - cmd/launchpal-privhelper/main.go
  - frontend/app/pages/index.vue
  - frontend/pnpm-workspace.yaml
  - internal/privhelper/logpath.go
  - frontend/app/components/StatusConfidenceIcon.vue
  - frontend/app/components/ServiceRow.vue
  - internal/privhelper/handlers.go
  - internal/settings/settings.go
  - .github/workflows/build.yml
  - internal/launchctl/status_detect.go
  - frontend/app/components/ReadOnlyServiceList.vue
  - internal/privhelper/install.go
  - internal/launchctl/nofollow_other.go
  - internal/launchctl/nofollow_darwin.go
  - frontend/app/components/LaunchPolicyForm.vue
  - frontend/app/composables/useBackupDiff.ts
  - frontend/app/components/ServiceLogs.vue
  - internal/plistutil/plistutil.go
  - internal/launchctl/system.go
  - .agents/skills/spectra-discuss/SKILL.md
  - frontend/app/utils/formatters.ts
  - frontend/app/components/CreateServiceModal.vue
  - internal/launchctl/keepalive.go
  - internal/privhelper/peer_other.go
  - .github/workflows/release-please.yml
  - internal/launchctl/user.go
  - CLAUDE.md
  - frontend/app/composables/useServiceListFilters.ts
  - internal/launchctl/apple_system.go
  - frontend/app/pages/settings.vue
  - frontend/app/composables/useSettings.ts
  - frontend/app/utils/logPaths.ts
  - .agents/skills/spectra-debug/SKILL.md
  - frontend/wailsjs/go/models.ts
  - frontend/app/components/ServiceSummary.vue
  - internal/privhelper/logpath_other.go
  - internal/privhelper/integrity.go
  - internal/privhelper/nofollow_darwin.go
  - go.mod
  - frontend/app/utils/serviceFilters.ts
  - .agents/skills/spectra-archive/SKILL.md
  - cmd/launchpal-privhelper/procinfo_other.go
  - .agents/skills/spectra-apply/SKILL.md
  - frontend/wailsjs/go/main/App.d.ts
  - frontend/app/components/LogStorageSection.vue
  - frontend/app/pages/services/[name].vue
  - frontend/app/utils/serviceValidation.ts
  - frontend/app/components/ServiceFilterDropdown.vue
  - frontend/app/types/wails.d.ts
  - .agents/skills/spectra-commit/SKILL.md
  - internal/privhelper/logpath_darwin.go
  - frontend/app/components/ScheduleForm.vue
  - internal/privhelper/peer_darwin.go
  - .agents/skills/spectra-ingest/SKILL.md
  - frontend/package.json
  - internal/launchctl/manager.go
  - internal/privhelper/client.go
  - frontend/app/utils/ansiToHtml.ts
  - .agents/skills/spectra-propose/SKILL.md
  - main.go
  - frontend/app/components/ServiceFilterBar.vue
  - frontend/app/composables/useAdminMode.ts
  - frontend/app/utils/serviceToConfig.ts
  - .agents/skills/spectra-drift/SKILL.md
  - .spectra.yaml
  - frontend/nuxt.config.ts
  - go.sum
  - internal/backup/backup.go
  - frontend/app/components/BackupDiffDialog.vue
  - frontend/eslint.config.mjs
  - internal/privhelper/protocol.go
  - frontend/vitest.setup.ts
  - frontend/vitest.config.ts
  - frontend/app/pages/system.vue
  - internal/privhelper/server.go
  - app.go
  - .agents/skills/spectra-ask/SKILL.md
  - frontend/app/components/StatusBar.vue
  - AGENTS.md
  - internal/launchctl/types.go
  - frontend/app/components/InlineBanner.vue
  - internal/privhelper/nofollow_other.go
  - README.md
tests:
  - frontend/app/pages/services/__tests__/edit-program-arguments-validation.test.ts
  - frontend/app/components/__tests__/ServiceLogs.test.ts
  - internal/privhelper/client_test.go
  - admin_mode_testhelpers_test.go
  - cmd/launchpal-privhelper/helper_test.go
  - internal/privhelper/protocol_test.go
  - frontend/app/pages/__tests__/serviceListFilterBar.test.ts
  - frontend/app/components/__tests__/CreateServiceModal.test.ts
  - frontend/app/composables/__tests__/useServiceListFilters.test.ts
  - frontend/app/composables/__tests__/useBackupDiff.test.ts
  - frontend/app/utils/__tests__/serviceFilters.test.ts
  - frontend/app/components/__tests__/CloneUserService.test.ts
  - internal/launchctl/types_test.go
  - internal/settings/settings_test.go
  - frontend/app/pages/services/__tests__/edit-env-masking.test.ts
  - internal/privhelper/handlers_test.go
  - frontend/app/components/__tests__/ServiceRow.test.ts
  - internal/privhelper/install_test.go
  - frontend/app/pages/services/__tests__/edit-launch-policy.test.ts
  - resolve_helper_test.go
  - internal/launchctl/readonly_test.go
  - internal/privhelper/server_test.go
  - internal/privhelper/integrity_test.go
  - frontend/app/utils/__tests__/serviceToConfig.test.ts
  - frontend/app/components/__tests__/LaunchPolicyForm.test.ts
  - internal/backup/backup_test.go
  - internal/launchctl/user_test.go
  - internal/launchctl/apple_system_test.go
  - admin_mode_test.go
  - frontend/app/components/__tests__/ServiceSummary.test.ts
  - internal/launchctl/keepalive_test.go
  - internal/plistutil/plistutil_test.go
  - frontend/app/utils/__tests__/ansiToHtml.test.ts
  - frontend/app/components/__tests__/ServiceFilterBar.test.ts
  - frontend/app/components/__tests__/BackupDiffDialog.test.ts
  - frontend/app/composables/__tests__/useSettings.test.ts
  - app_test.go
  - frontend/app/utils/__tests__/launchPolicy.test.ts
  - frontend/app/utils/__tests__/formatters.test.ts
  - frontend/app/pages/__tests__/settings.test.ts
  - internal/launchctl/system_test.go
  - internal/launchctl/status_detect_test.go
  - internal/launchctl/plist_encode_test.go
  - frontend/app/composables/__tests__/useNextOccurrences.test.ts
  - internal/plistutil/testhelpers_test.go
-->

---
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

<!-- @trace
source: service-list-filters
updated: 2026-08-26
code:
  - frontend/app/utils/launchPolicy.ts
  - internal/launchctl/plist_encode.go
  - frontend/app/utils/settingsValidation.ts
  - frontend/app/composables/useNextOccurrences.ts
  - CHANGELOG.md
  - .agents/skills/spectra-audit/SKILL.md
  - internal/launchctl/readonly.go
  - admin_mode.go
  - cmd/launchpal-privhelper/procinfo_darwin.go
  - frontend/package.json.md5
  - frontend/wailsjs/go/main/App.js
  - Makefile
  - cmd/launchpal-privhelper/main.go
  - frontend/app/pages/index.vue
  - frontend/pnpm-workspace.yaml
  - internal/privhelper/logpath.go
  - frontend/app/components/StatusConfidenceIcon.vue
  - frontend/app/components/ServiceRow.vue
  - internal/privhelper/handlers.go
  - internal/settings/settings.go
  - .github/workflows/build.yml
  - internal/launchctl/status_detect.go
  - frontend/app/components/ReadOnlyServiceList.vue
  - internal/privhelper/install.go
  - internal/launchctl/nofollow_other.go
  - internal/launchctl/nofollow_darwin.go
  - frontend/app/components/LaunchPolicyForm.vue
  - frontend/app/composables/useBackupDiff.ts
  - frontend/app/components/ServiceLogs.vue
  - internal/plistutil/plistutil.go
  - internal/launchctl/system.go
  - .agents/skills/spectra-discuss/SKILL.md
  - frontend/app/utils/formatters.ts
  - frontend/app/components/CreateServiceModal.vue
  - internal/launchctl/keepalive.go
  - internal/privhelper/peer_other.go
  - .github/workflows/release-please.yml
  - internal/launchctl/user.go
  - CLAUDE.md
  - frontend/app/composables/useServiceListFilters.ts
  - internal/launchctl/apple_system.go
  - frontend/app/pages/settings.vue
  - frontend/app/composables/useSettings.ts
  - frontend/app/utils/logPaths.ts
  - .agents/skills/spectra-debug/SKILL.md
  - frontend/wailsjs/go/models.ts
  - frontend/app/components/ServiceSummary.vue
  - internal/privhelper/logpath_other.go
  - internal/privhelper/integrity.go
  - internal/privhelper/nofollow_darwin.go
  - go.mod
  - frontend/app/utils/serviceFilters.ts
  - .agents/skills/spectra-archive/SKILL.md
  - cmd/launchpal-privhelper/procinfo_other.go
  - .agents/skills/spectra-apply/SKILL.md
  - frontend/wailsjs/go/main/App.d.ts
  - frontend/app/components/LogStorageSection.vue
  - frontend/app/pages/services/[name].vue
  - frontend/app/utils/serviceValidation.ts
  - frontend/app/components/ServiceFilterDropdown.vue
  - frontend/app/types/wails.d.ts
  - .agents/skills/spectra-commit/SKILL.md
  - internal/privhelper/logpath_darwin.go
  - frontend/app/components/ScheduleForm.vue
  - internal/privhelper/peer_darwin.go
  - .agents/skills/spectra-ingest/SKILL.md
  - frontend/package.json
  - internal/launchctl/manager.go
  - internal/privhelper/client.go
  - frontend/app/utils/ansiToHtml.ts
  - .agents/skills/spectra-propose/SKILL.md
  - main.go
  - frontend/app/components/ServiceFilterBar.vue
  - frontend/app/composables/useAdminMode.ts
  - frontend/app/utils/serviceToConfig.ts
  - .agents/skills/spectra-drift/SKILL.md
  - .spectra.yaml
  - frontend/nuxt.config.ts
  - go.sum
  - internal/backup/backup.go
  - frontend/app/components/BackupDiffDialog.vue
  - frontend/eslint.config.mjs
  - internal/privhelper/protocol.go
  - frontend/vitest.setup.ts
  - frontend/vitest.config.ts
  - frontend/app/pages/system.vue
  - internal/privhelper/server.go
  - app.go
  - .agents/skills/spectra-ask/SKILL.md
  - frontend/app/components/StatusBar.vue
  - AGENTS.md
  - internal/launchctl/types.go
  - frontend/app/components/InlineBanner.vue
  - internal/privhelper/nofollow_other.go
  - README.md
tests:
  - frontend/app/pages/services/__tests__/edit-program-arguments-validation.test.ts
  - frontend/app/components/__tests__/ServiceLogs.test.ts
  - internal/privhelper/client_test.go
  - admin_mode_testhelpers_test.go
  - cmd/launchpal-privhelper/helper_test.go
  - internal/privhelper/protocol_test.go
  - frontend/app/pages/__tests__/serviceListFilterBar.test.ts
  - frontend/app/components/__tests__/CreateServiceModal.test.ts
  - frontend/app/composables/__tests__/useServiceListFilters.test.ts
  - frontend/app/composables/__tests__/useBackupDiff.test.ts
  - frontend/app/utils/__tests__/serviceFilters.test.ts
  - frontend/app/components/__tests__/CloneUserService.test.ts
  - internal/launchctl/types_test.go
  - internal/settings/settings_test.go
  - frontend/app/pages/services/__tests__/edit-env-masking.test.ts
  - internal/privhelper/handlers_test.go
  - frontend/app/components/__tests__/ServiceRow.test.ts
  - internal/privhelper/install_test.go
  - frontend/app/pages/services/__tests__/edit-launch-policy.test.ts
  - resolve_helper_test.go
  - internal/launchctl/readonly_test.go
  - internal/privhelper/server_test.go
  - internal/privhelper/integrity_test.go
  - frontend/app/utils/__tests__/serviceToConfig.test.ts
  - frontend/app/components/__tests__/LaunchPolicyForm.test.ts
  - internal/backup/backup_test.go
  - internal/launchctl/user_test.go
  - internal/launchctl/apple_system_test.go
  - admin_mode_test.go
  - frontend/app/components/__tests__/ServiceSummary.test.ts
  - internal/launchctl/keepalive_test.go
  - internal/plistutil/plistutil_test.go
  - frontend/app/utils/__tests__/ansiToHtml.test.ts
  - frontend/app/components/__tests__/ServiceFilterBar.test.ts
  - frontend/app/components/__tests__/BackupDiffDialog.test.ts
  - frontend/app/composables/__tests__/useSettings.test.ts
  - app_test.go
  - frontend/app/utils/__tests__/launchPolicy.test.ts
  - frontend/app/utils/__tests__/formatters.test.ts
  - frontend/app/pages/__tests__/settings.test.ts
  - internal/launchctl/system_test.go
  - internal/launchctl/status_detect_test.go
  - internal/launchctl/plist_encode_test.go
  - frontend/app/composables/__tests__/useNextOccurrences.test.ts
  - internal/plistutil/testhelpers_test.go
-->