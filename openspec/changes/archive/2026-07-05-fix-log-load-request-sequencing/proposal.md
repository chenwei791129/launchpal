## Problem

The Logs tab's shared loader `loadLogs()` in `frontend/app/components/ServiceLogs.vue` has no request sequencing, so two invocations can be in flight at once and the later-resolving one wins `logs.value` / `error.value` regardless of which the user is actually viewing. Concrete symptoms:

- **Wrong stream content**: with Auto-refresh enabled, a 2-second poll can start a `GetLogs(stdout)` read; if the user clicks the stderr toggle before it resolves, the `logType` watcher starts a second concurrent `GetLogs(stderr)`. If the older stdout read resolves last, `logs.value` ends up holding stdout content while the UI shows the stderr tab, and the Auto-refresh auto-disable decision is driven by the stale stdout status.
- **Suppressed success indicator on Clear**: `confirmClear()` reloads via `loadLogs()` and then reads `error.value` to decide whether to flash the green "Log cleared" indicator. A concurrent poll load that transiently rejects during that window writes `error.value` and suppresses the indicator even though the clear + reload succeeded.

Both are timing-dependent and self-heal on the next poll tick, but they produce visibly wrong output while they last. The in-flight guard added for Auto-refresh only prevents poll-versus-poll overlap (`if (loading.value) return` inside the interval callback); it does not prevent a poll load from overlapping a stream switch, a manual Refresh, or a Clear-triggered reload.

## Root Cause

`loadLogs()` mutates shared reactive state (`logs.value`, `error.value`, and — via the `loadOk` computation — the Auto-refresh toggle and the success indicator) unconditionally when its awaited `GetLogs` / `GetSystemLogs` promise settles. There is no token identifying which load is the newest, so a stale (superseded) load's result is applied as if it were current. The `loading` flag is a single boolean and cannot distinguish "a newer load has started" from "this load is still running".

## Proposed Solution

Introduce a monotonic request-sequence token in `ServiceLogs.vue`, incremented at the start of every `loadLogs()` call. After the awaited binding settles (both the resolve and the reject paths), the load SHALL compare its captured token against the latest issued token and apply its result — assigning `logs.value`, setting `error.value`, clearing the `loading` flag, and running the Auto-refresh auto-disable check — only when it is still the newest load. A superseded load SHALL discard its result without mutating any shared state, leaving the newest load authoritative.

This keeps every happy-path (single, non-overlapping load) behavior byte-for-byte identical and only changes behavior when loads overlap: the stream/service/trigger the user last acted on wins, and stale responses can no longer overwrite content, flip the toggle, or pollute the success/error indicators.

## Non-Goals

- No request cancellation at the binding/backend layer — the superseded Go read still runs to completion; only its front-end result is discarded. Wails/`launchctl` reads are cheap (≤1MB tail) so abandoning the result is sufficient.
- No change to the 2-second polling cadence, the in-flight tick skip, the Auto-refresh auto-disable semantics, or any placeholder/error wording — those remain as specified by `log-auto-refresh` and `log-load-feedback`.
- No backend, Wails binding, or settings changes.
- No timeout/watchdog for a load that never resolves (a hung binding) — out of scope for this fix.

## Success Criteria

- When a poll-initiated `loadLogs` for stdout is in flight and the user switches to stderr, the pane renders the stderr result even if the stdout read resolves afterward; the stale stdout result is discarded and does not drive the Auto-refresh auto-disable check.
- When a stale (superseded) load rejects, it does not set `error.value`; only the newest load's outcome controls the error branch and the Clear success indicator.
- The parallel Clear-button status query (`loadLogClearStatus`) carries the same request-sequencing guard, so on a rapid stream switch an older status response can no longer leave the Clear button's enabled state/tooltip describing the wrong stream.
- A single non-overlapping load behaves exactly as today (content, placeholder, error, Auto-refresh auto-disable, and scroll behavior all unchanged).
- New vitest cases in `frontend/app/components/__tests__/ServiceLogs.test.ts` cover the out-of-order stream-switch race and the stale-rejection case; `make test` and `make lint` pass with no regression in existing ServiceLogs tests.

## Impact

- Affected specs: `log-load-feedback` (adds a requirement on discarding superseded concurrent load results)
- Affected code:
  - Modified: frontend/app/components/ServiceLogs.vue
  - Modified: frontend/app/components/__tests__/ServiceLogs.test.ts
- No Go backend changes, no Wails bindings changes, no settings changes.
