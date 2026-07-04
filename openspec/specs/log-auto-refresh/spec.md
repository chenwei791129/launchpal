# log-auto-refresh Specification

## Purpose

TBD - created by archiving change 'auto-refresh-service-logs'. Update Purpose after archive.

## Requirements

### Requirement: Auto-refresh toggle in the Logs tab

The `ServiceLogs.vue` component SHALL render an "Auto-refresh" checkbox in the log controls row for all three service types (user, system, apple-system).
The checkbox SHALL default to unchecked and SHALL be independent of the existing Auto-scroll checkbox: either can be toggled without changing the other.
The toggle state SHALL be component-local and session-only; it SHALL NOT be persisted to settings.
The label text SHALL be written in English.

#### Scenario: Toggle is rendered unchecked by default

- **WHEN** the Logs tab is opened
- **THEN** an "Auto-refresh" checkbox is visible and unchecked, and no polling occurs

#### Scenario: Toggles are independent

- **WHEN** Auto-refresh is checked while Auto-scroll is unchecked
- **THEN** periodic reloads occur and the scroll position is not forced to the bottom after each reload

#### Scenario: Follow mode scrolls to bottom

- **WHEN** Auto-refresh and Auto-scroll are both checked and a polled reload resolves with new content
- **THEN** the log view scrolls to the bottom after the reload, matching the existing Auto-scroll behavior of manual loads


<!-- @trace
source: auto-refresh-service-logs
updated: 2026-07-04
code:
  - frontend/app/components/ServiceLogs.vue
tests:
  - frontend/app/components/__tests__/ServiceLogs.test.ts
-->

---
### Requirement: Periodic reload while Auto-refresh is enabled

While the Auto-refresh checkbox is checked, the component SHALL reload the current log stream through the existing `loadLogs` path every 2 seconds.
A polling tick SHALL be skipped (not queued, not run concurrently) when a previous load is still in flight.
Unchecking the checkbox SHALL stop the polling immediately.
Unmounting the component SHALL clear the polling timer so no further requests are issued.
A change of the `serviceName` prop SHALL stop polling and reset the Auto-refresh checkbox to unchecked: the detail page reuses the same route component across service-to-service navigation without remounting `ServiceLogs`, so unmount hooks alone MUST NOT be relied on to stop polling.
Switching between stdout and stderr SHALL keep the toggle state: if Auto-refresh is enabled, polling continues against the newly selected stream.

#### Scenario: Enabled toggle reloads periodically

- **WHEN** Auto-refresh is checked and 2 seconds elapse
- **THEN** the component issues one reload through the same path used by the manual Refresh button

#### Scenario: In-flight tick is skipped

- **WHEN** a polling tick fires while the previous load has not yet resolved
- **THEN** the tick performs no additional load call

#### Scenario: Unmount stops polling

- **WHEN** the component is unmounted while Auto-refresh is checked
- **THEN** the polling timer is cleared and no further load calls occur

#### Scenario: Service navigation stops polling

- **WHEN** Auto-refresh is checked and the `serviceName` prop changes because the user navigated from one service's detail page to another without the component remounting
- **THEN** the polling timer is cleared and the Auto-refresh checkbox becomes unchecked

#### Scenario: Stream switch keeps polling

- **WHEN** Auto-refresh is checked and the user switches from stdout to stderr whose loads resolve with Status "ok"
- **THEN** polling continues against stderr


<!-- @trace
source: auto-refresh-service-logs
updated: 2026-07-04
code:
  - frontend/app/components/ServiceLogs.vue
tests:
  - frontend/app/components/__tests__/ServiceLogs.test.ts
-->

---
### Requirement: Auto-refresh disables itself on a non-ok load outcome

After any load completes — triggered by polling, the manual Refresh button, or a stream switch — while Auto-refresh is enabled, the component SHALL set the Auto-refresh checkbox to unchecked and stop polling when the load resolved with a `LogsResult` whose `Status` is `"no-path"` or `"not-found"`, or when the load promise rejected.
The disable check SHALL live in the shared post-load path used by all load triggers, and SHALL read the lowercase runtime key `status` (the Wails runtime object carries lowercase json keys; `Status` above is the Go field name).
A load that completes without producing a `LogsResult` because no Wails binding is available (development fallback) SHALL likewise disable Auto-refresh and stop polling.
The rendered feedback for the non-ok outcome SHALL remain exactly as specified by the `log-load-feedback` capability; auto-disabling SHALL NOT add any additional error surface.
Auto-refresh SHALL NOT re-enable itself; resuming requires the user to check the checkbox again.

#### Scenario: Structural status disables polling

- **WHEN** Auto-refresh is checked and a polled load resolves with Status "not-found"
- **THEN** the checkbox becomes unchecked, polling stops, and the "Log file does not exist yet" placeholder is rendered

#### Scenario: Rejection disables polling

- **WHEN** Auto-refresh is checked and a polled load rejects with a backend error string
- **THEN** the checkbox becomes unchecked, polling stops, and the error branch shows the backend message

#### Scenario: Manual Refresh failure also disables polling

- **WHEN** Auto-refresh is checked and the user clicks the manual Refresh button whose load rejects
- **THEN** the checkbox becomes unchecked and polling stops

#### Scenario: Stream switch to a stream without a log path disables polling

- **WHEN** Auto-refresh is checked and the user switches to a stream whose load resolves with Status "no-path"
- **THEN** the checkbox becomes unchecked, polling stops, and the "no log path configured" placeholder is rendered

#### Scenario: Development fallback disables polling

- **WHEN** Auto-refresh is checked and a load completes via the development fallback because no Wails binding is available
- **THEN** the checkbox becomes unchecked and polling stops

#### Scenario: No automatic resume

- **WHEN** Auto-refresh was auto-disabled and the underlying log file later becomes available
- **THEN** no polling occurs until the user checks the Auto-refresh checkbox again

<!-- @trace
source: auto-refresh-service-logs
updated: 2026-07-04
code:
  - frontend/app/components/ServiceLogs.vue
tests:
  - frontend/app/components/__tests__/ServiceLogs.test.ts
-->