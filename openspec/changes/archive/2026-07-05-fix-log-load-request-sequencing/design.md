## Context

`ServiceLogs.vue` funnels every log read through one async function, `loadLogs()`. It is invoked from five triggers: `onMounted`, the `logType` (stdout/stderr) watcher, the manual Refresh button, the Auto-refresh poll `setInterval`, and `confirmClear()`'s post-clear reload. On settle it mutates shared reactive state: `logs.value` (rendered content / placeholder), `error.value` (red error branch), the `loading` flag, and — through the `loadOk` computation — the Auto-refresh toggle (`autoRefresh`) and, indirectly, the Clear success indicator (`confirmClear` reads `error.value` afterward).

The Auto-refresh work added an in-flight guard, but only inside the poll callback (`if (loading.value) return`), which prevents poll-versus-poll stacking and nothing else. Any two loads from *different* triggers can still overlap, and because `loadLogs` applies its result unconditionally on settle, the later-resolving promise wins regardless of which load the user actually cares about. A slow read that started earlier can therefore overwrite the content of a newer read.

Constraint: the fix must be purely front-end in `ServiceLogs.vue`; no backend, binding, or settings changes. Happy-path behavior (a single non-overlapping load) must stay identical, so existing `log-load-feedback` and `log-auto-refresh` scenarios keep passing unchanged.

## Goals / Non-Goals

**Goals**

- A superseded (older) in-flight load never overwrites the content, error, toggle, or success indicator owned by a newer load.
- The stream / service / trigger the user last acted on is always the one whose result is rendered.
- Zero behavior change when loads do not overlap.

**Non-Goals**

- No cancellation of the underlying Go read — the superseded read completes; only its front-end result is dropped.
- No change to polling cadence, in-flight tick skip, auto-disable semantics, or any wording.
- No timeout/watchdog for a load that never resolves.
- No backend / binding / settings changes.

## Decisions

### Decision 1: Monotonic request-sequence token

Add a module-scoped counter in `<script setup>` (e.g. `loadSeq`, initialized to 0). At the very top of `loadLogs()`, capture `const seq = ++loadSeq`. After the awaited binding settles, the load applies its result only while `seq === loadSeq` (it is still the newest load); otherwise it discards silently. Why a monotonic counter over a boolean or an `AbortController`: the boolean `loading` flag cannot distinguish "a newer load started" from "still running", and Wails bindings are plain promises with no abort signal, so a comparison token is the minimal correct mechanism. Why increment before the await (not after): the token must be claimed synchronously at call time so a second `loadLogs()` entered before the first awaits already owns a higher number.

### Decision 2: Gate shared-state mutation on the sequence check

The `seq === loadSeq` check SHALL gate: the assignment to `logs.value`, the `loadOk`-derived Auto-refresh auto-disable, the `error.value` assignment in the catch path, and the `loading.value = false` reset. A superseded load leaving `loading` false would clear the spinner the newest load still needs, so the reset SHALL run only for the newest load; the newest load's own `finally` is the single place that clears it. The success indicator in `confirmClear()` is fixed transitively: because a stale poll load can no longer write `error.value`, `confirmClear`'s `if (!error.value)` check reflects only its own reload's outcome.

### Decision 3: Leave the nextTick/scroll micro-window unguarded

After a successful load the current code does `await nextTick()` then `scrollToBottom()` when Auto-scroll is on. A newer load could be issued during that microtask gap. The stale-load check is placed at the primary points that mutate persistent shared state (content assignment, error, toggle, loading reset); a scroll performed by a just-superseded load is visually harmless (the newer load will scroll again on its own completion) and does not corrupt `logs.value`. Adding a second guard around the scroll is unnecessary complexity, so it is explicitly out of scope.

### Decision 4: Extend the same guard to the parallel Clear-button status query

`loadLogClearStatus()` drives the Clear control's enabled state and tooltip and fires from the same triggers as `loadLogs` (mount, the `logType` watcher, and `confirmClear`'s `Promise.all`). It has the identical last-resolver-wins hazard: on a rapid stdout↔stderr switch an older status query can resolve after the newer one and leave `logClearStatus.value` describing the wrong stream (e.g. the Clear button enabled/disabled for the stream no longer shown). It carries its own independent `clearStatusSeq` counter, claimed synchronously before the `GetLogClearStatus` await; the assignment on both the resolve and the catch (silent-fail → `null`) paths is gated on `seq === clearStatusSeq`. The counter is separate from `loadSeq` because the two queries are issued and settle independently — sharing one counter would let a log-content load spuriously supersede a status query and vice versa.

## Implementation Contract

**Observable behavior**

1. With Auto-refresh on, a poll-started stdout load in flight, and the user switching to stderr: the pane renders the stderr result; if the stdout read resolves after the stderr read, its result is discarded and does not appear, and it does not drive the Auto-refresh auto-disable check.
2. A superseded load that rejects does not set `error.value`; the red error branch reflects only the newest load's outcome. Consequently the Clear success indicator is suppressed only by the reload initiated inside `confirmClear`, never by an overlapping poll load.
3. A single, non-overlapping load renders content / placeholder / error, drives the Auto-refresh auto-disable, and scrolls exactly as before this change.

**Interface / data shape**

- Internal to `ServiceLogs.vue` only: two added module-scoped counters (`loadSeq` for `loadLogs`, `clearStatusSeq` for `loadLogClearStatus`) and per-call captured tokens. No new props, emits, bindings, or exported symbols.

**Failure modes**

- If a load never settles, its token is never compared and no result is applied — identical to today's behavior (the spinner stays until a newer load settles). This is unchanged and out of scope.

**Acceptance criteria**

- New vitest cases in `frontend/app/components/__tests__/ServiceLogs.test.ts`:
  - out-of-order resolution: an earlier stdout load resolving after a later stderr load does not overwrite the stderr content and does not disable Auto-refresh based on the stdout status;
  - stale rejection: a superseded load rejecting does not populate the error branch.
  - stale Clear-status query: an older `loadLogClearStatus` resolving after a newer stream switch does not flip the Clear button's enabled state/tooltip away from the newest stream.
- All existing ServiceLogs tests still pass; `make test` and `make lint` exit 0.

**Scope boundaries**

- In scope: request-sequencing guard inside `loadLogs()` and `loadLogClearStatus()`, and the accompanying tests.
- Out of scope: backend/binding/settings, request cancellation, hung-load timeout, polling cadence, auto-disable wording, and the nextTick/scroll micro-window (Decision 3).

## Risks / Trade-offs

- [A superseded load still consumes a backend read (no cancellation)] → Accepted: reads are cheap (≤1MB tail) and infrequent (2s poll); dropping the result is sufficient and far simpler than plumbing an abort path through Wails.
- [Token comparison after `await` is easy to get subtly wrong across resolve/catch/finally] → Mitigated by gating all three (resolve assignment, catch, finally reset) on the same `seq === loadSeq` check and covering the ordering with the two new fake-timer tests.

## Migration Plan

Single front-end component change, one commit; rollback is a straight revert. No data or schema migration. No dependency on other in-flight changes.

## Open Questions

(None — the token mechanism, the gated mutation set, and the scroll micro-window decision are all settled above.)
