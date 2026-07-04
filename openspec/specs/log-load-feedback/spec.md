## ADDED Requirements

### Requirement: Classify log load results in the Logs tab

The `ServiceLogs.vue` component SHALL branch its rendering on the `Status` field of the `LogsResult` returned by `GetLogs` / `GetSystemLogs`, for all three service types (user, system, apple-system).
The runtime object delivered by Wails carries the lowercase JSON keys `content` / `status` / `path` (per the Go struct's json tags); this spec refers to the Go field names `Content` / `Status` / `Path` for readability, and the component SHALL read the lowercase keys.
When no Wails binding is available (development fallback), the component SHALL render the existing "No logs available for {logType}" placeholder.
When `Status` is `"ok"` and `Content` is non-empty, the component SHALL render the content through the existing ANSI rendering pipeline.
When `Status` is `"ok"` and `Content` is empty, the component SHALL render the existing "No logs available for {logType}" placeholder.
When `Status` is `"no-path"`, the component SHALL render a placeholder stating "No {logType} log path configured for this service", styled as the existing neutral placeholder (not as the red error branch).
When `Status` is `"not-found"`, the component SHALL render a placeholder stating "Log file does not exist yet" with the resolved `Path` shown as secondary text, styled as the existing neutral placeholder (not as the red error branch).
All placeholder and error strings SHALL be written in English.

#### Scenario: No log path configured renders neutral placeholder

- **WHEN** GetLogs resolves with Status "no-path" for logType "stdout"
- **THEN** the component renders the placeholder text "No stdout log path configured for this service" and does not render the red error branch

#### Scenario: Missing log file renders neutral placeholder with path

- **WHEN** GetLogs resolves with Status "not-found" and Path "/var/log/foo/out.log"
- **THEN** the component renders the placeholder text "Log file does not exist yet" with "/var/log/foo/out.log" as secondary text, and does not render the red error branch

#### Scenario: Empty content keeps the existing placeholder

- **WHEN** GetLogs resolves with Status "ok" and empty Content
- **THEN** the component renders the existing "No logs available for stdout" placeholder

#### Scenario: Non-empty content renders through the ANSI pipeline

- **WHEN** GetLogs resolves with Status "ok" and Content "hello\n"
- **THEN** the component renders the log content in the existing `<pre>` element

#### Scenario: Development fallback without bindings

- **WHEN** the component loads logs and no Wails binding is available on `window.go`
- **THEN** the component renders the existing "No logs available for {logType}" placeholder and does not render the red error branch


<!-- @trace
source: fix-log-error-classification
updated: 2026-07-04
code:
  - frontend/app/types/wails.d.ts
  - frontend/wailsjs/go/models.ts
  - internal/launchctl/system.go
  - internal/launchctl/types.go
  - frontend/wailsjs/go/main/App.d.ts
  - internal/launchctl/user.go
  - app.go
  - internal/launchctl/readonly.go
  - frontend/app/components/ServiceLogs.vue
  - internal/launchctl/apple_system.go
  - internal/launchctl/manager.go
tests:
  - internal/launchctl/apple_system_test.go
  - internal/launchctl/system_test.go
  - frontend/app/components/__tests__/ServiceLogs.test.ts
  - internal/launchctl/user_test.go
-->

### Requirement: Surface backend error messages verbatim on load failure

When the `GetLogs` / `GetSystemLogs` promise rejects, the `ServiceLogs.vue` component SHALL display the rejection payload in the red error branch: a string rejection SHALL be displayed as-is, an `Error` rejection SHALL display its `message`, and only a rejection that is neither SHALL fall back to the generic text "Failed to load logs".

#### Scenario: String rejection from Wails is shown verbatim

- **WHEN** GetLogs rejects with the string "permission denied reading log file: /var/log/foo/out.log"
- **THEN** the error branch displays exactly that string instead of "Failed to load logs"

#### Scenario: Error object rejection shows its message

- **WHEN** GetLogs rejects with an Error whose message is "boom"
- **THEN** the error branch displays "boom"

#### Scenario: Switching state clears stale feedback

- **WHEN** a load for stdout fails with an error and the user switches to stderr whose load resolves with Status "ok" and non-empty Content
- **THEN** the error branch is cleared and the stderr content is rendered

## Requirements

<!-- @trace
source: fix-log-error-classification
updated: 2026-07-04
code:
  - frontend/app/types/wails.d.ts
  - frontend/wailsjs/go/models.ts
  - internal/launchctl/system.go
  - internal/launchctl/types.go
  - frontend/wailsjs/go/main/App.d.ts
  - internal/launchctl/user.go
  - app.go
  - internal/launchctl/readonly.go
  - frontend/app/components/ServiceLogs.vue
  - internal/launchctl/apple_system.go
  - internal/launchctl/manager.go
tests:
  - internal/launchctl/apple_system_test.go
  - internal/launchctl/system_test.go
  - frontend/app/components/__tests__/ServiceLogs.test.ts
  - internal/launchctl/user_test.go
-->

### Requirement: Classify log load results in the Logs tab

The `ServiceLogs.vue` component SHALL branch its rendering on the `Status` field of the `LogsResult` returned by `GetLogs` / `GetSystemLogs`, for all three service types (user, system, apple-system).
The runtime object delivered by Wails carries the lowercase JSON keys `content` / `status` / `path` (per the Go struct's json tags); this spec refers to the Go field names `Content` / `Status` / `Path` for readability, and the component SHALL read the lowercase keys.
When no Wails binding is available (development fallback), the component SHALL render the existing "No logs available for {logType}" placeholder.
When `Status` is `"ok"` and `Content` is non-empty, the component SHALL render the content through the existing ANSI rendering pipeline.
When `Status` is `"ok"` and `Content` is empty, the component SHALL render the existing "No logs available for {logType}" placeholder.
When `Status` is `"no-path"`, the component SHALL render a placeholder stating "No {logType} log path configured for this service", styled as the existing neutral placeholder (not as the red error branch).
When `Status` is `"not-found"`, the component SHALL render a placeholder stating "Log file does not exist yet" with the resolved `Path` shown as secondary text, styled as the existing neutral placeholder (not as the red error branch).
All placeholder and error strings SHALL be written in English.

#### Scenario: No log path configured renders neutral placeholder

- **WHEN** GetLogs resolves with Status "no-path" for logType "stdout"
- **THEN** the component renders the placeholder text "No stdout log path configured for this service" and does not render the red error branch

#### Scenario: Missing log file renders neutral placeholder with path

- **WHEN** GetLogs resolves with Status "not-found" and Path "/var/log/foo/out.log"
- **THEN** the component renders the placeholder text "Log file does not exist yet" with "/var/log/foo/out.log" as secondary text, and does not render the red error branch

#### Scenario: Empty content keeps the existing placeholder

- **WHEN** GetLogs resolves with Status "ok" and empty Content
- **THEN** the component renders the existing "No logs available for stdout" placeholder

#### Scenario: Non-empty content renders through the ANSI pipeline

- **WHEN** GetLogs resolves with Status "ok" and Content "hello\n"
- **THEN** the component renders the log content in the existing `<pre>` element

#### Scenario: Development fallback without bindings

- **WHEN** the component loads logs and no Wails binding is available on `window.go`
- **THEN** the component renders the existing "No logs available for {logType}" placeholder and does not render the red error branch

---
### Requirement: Surface backend error messages verbatim on load failure

When the `GetLogs` / `GetSystemLogs` promise rejects, the `ServiceLogs.vue` component SHALL display the rejection payload in the red error branch: a string rejection SHALL be displayed as-is, an `Error` rejection SHALL display its `message`, and only a rejection that is neither SHALL fall back to the generic text "Failed to load logs".

#### Scenario: String rejection from Wails is shown verbatim

- **WHEN** GetLogs rejects with the string "permission denied reading log file: /var/log/foo/out.log"
- **THEN** the error branch displays exactly that string instead of "Failed to load logs"

#### Scenario: Error object rejection shows its message

- **WHEN** GetLogs rejects with an Error whose message is "boom"
- **THEN** the error branch displays "boom"

#### Scenario: Switching state clears stale feedback

- **WHEN** a load for stdout fails with an error and the user switches to stderr whose load resolves with Status "ok" and non-empty Content
- **THEN** the error branch is cleared and the stderr content is rendered

---
### Requirement: Discard superseded concurrent log load results

The `ServiceLogs.vue` component funnels every log read through a single shared loader that can be invoked concurrently by distinct triggers (initial mount, stdout/stderr switch, manual Refresh, Auto-refresh polling, and the post-Clear reload). The component SHALL apply the result of a load — assigning the rendered `LogsResult`, setting the error branch, clearing the loading indicator, and driving the Auto-refresh auto-disable check — only when that load is the most recently issued one. When a newer load has been issued while an older load was still in flight, the older (superseded) load SHALL discard its outcome without mutating any shared rendering state, on both the resolve and the reject paths.

A single load that does not overlap any other load SHALL apply its result exactly as specified by the "Classify log load results in the Logs tab" and "Surface backend error messages verbatim on load failure" requirements; this requirement changes behavior only when two or more loads overlap.

The parallel Clear-button status query, which fires from the same triggers (initial mount, stdout/stderr switch, and the post-Clear reload), SHALL apply its result — the Clear control's enabled state and tooltip — only when it is the most recently issued status query; a superseded status query SHALL discard its outcome so the Clear control always describes the stream currently being viewed.

#### Scenario: Out-of-order stream switch keeps the newest stream's content

- **WHEN** a load for stdout is in flight, the user switches to stderr which starts a second load, and the stderr load resolves with Status "ok" before the stdout load resolves
- **THEN** the component renders the stderr content, and when the stdout load later resolves its result is discarded and does not replace the stderr content

#### Scenario: Superseded load does not drive the Auto-refresh auto-disable

- **WHEN** Auto-refresh is enabled, a stdout load is superseded by a newer stderr load that resolves with Status "ok", and the superseded stdout load later resolves with a non-ok Status
- **THEN** Auto-refresh remains enabled because the auto-disable decision is driven only by the newest load's outcome

#### Scenario: Superseded load rejection does not populate the error branch

- **WHEN** a load is superseded by a newer load and the superseded load later rejects with an error
- **THEN** the red error branch is not populated by the superseded rejection, and only the newest load's outcome controls the error branch

#### Scenario: Superseded Clear-button status query does not desync the Clear control

- **WHEN** the Clear-button status query for one stream is in flight, the user switches streams which starts a newer status query that resolves first, and the older status query resolves afterward
- **THEN** the Clear button's enabled state and tooltip reflect the newest stream's status, and the superseded status query's result is discarded

<!-- @trace
source: fix-log-load-request-sequencing
updated: 2026-07-05
code:
  - frontend/app/components/ServiceLogs.vue
tests:
  - frontend/app/components/__tests__/ServiceLogs.test.ts
-->