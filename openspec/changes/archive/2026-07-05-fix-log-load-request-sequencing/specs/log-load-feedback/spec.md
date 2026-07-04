## ADDED Requirements

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
