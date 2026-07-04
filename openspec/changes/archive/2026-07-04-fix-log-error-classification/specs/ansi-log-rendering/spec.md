## MODIFIED Requirements

### Requirement: Mount rendered output in ServiceLogs view

The `ServiceLogs.vue` component SHALL render log content through a `v-html`-bound `<pre>` element whose contents come from a `renderedLogs` computed value defined as `ansiToHtml` applied to the log content string carried in the `LogsResult` returned by `GetLogs` / `GetSystemLogs` (or the empty string when no content is loaded). The CSS classes `text-gray-300`, `whitespace-pre-wrap`, `break-all`, `font-mono`, and `text-sm` SHALL remain on the `<pre>` element so that monospace, line-wrapping, and color base behavior are unchanged. The existing loading and error branches SHALL be preserved; the placeholder branch SHALL follow the `log-load-feedback` capability, which extends it with status-specific wording for the `no-path` and `not-found` states.

#### Scenario: Logs containing ANSI colors render as colored spans

- **WHEN** `GetLogs` resolves with Status "ok" and Content `"\x1b[32mOK\x1b[0m booted"` and the Logs tab is opened
- **THEN** the `<pre>` element contains `<span style="color:#98c379">OK</span>` followed by ` booted`, and the literal characters `[32m` and `[0m` do not appear in the rendered DOM

#### Scenario: Empty log preserves placeholder

- **WHEN** `GetLogs` resolves with Status "ok" and empty Content
- **THEN** the component renders the existing "No logs available for {logType}" placeholder branch and does not render the `<pre>` element

#### Scenario: Existing loading state is unaffected

- **WHEN** `loading` is true and no log content is loaded
- **THEN** the component renders the existing spinner branch ("Loading logs..."), not the `<pre>` element

#### Scenario: Existing error state is unaffected

- **WHEN** the `GetLogs` promise rejects and `error.value` is set
- **THEN** the component renders the existing red-text error branch and does not render the `<pre>` element
