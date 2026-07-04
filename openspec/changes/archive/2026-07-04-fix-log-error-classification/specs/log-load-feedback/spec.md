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
