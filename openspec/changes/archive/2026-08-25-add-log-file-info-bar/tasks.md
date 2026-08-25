## 1. Backend: extend LogClearStatus with Size

- [x] 1.1 Add `Size int64` field to `LogClearStatus` struct in `internal/launchctl/types.go` with json tag `size`, and update the doc comment to describe how `Size` is populated for the four states (success / ENOENT / other-error / empty path) — extends the **Query log clear authorization status** requirement
- [x] 1.2 Write failing test cases in `internal/launchctl/types_test.go` covering `logClearStatusFor`: file exists with size N → `Size == N`; ENOENT → `Size == 0`; permission-denied (mode 0) → `Size == 0`; empty path → `Size == 0`
- [x] 1.3 Modify `logClearStatusFor` in `internal/launchctl/types.go` (shared by `readonly.go` and `user.go`) to call `f.Stat()` on the already-open file descriptor before closing it and populate `Size` from `Stat().Size()`; do not issue a separate `os.Stat` call
- [x] 1.4 Run `go test ./internal/launchctl/...` to confirm 1.2 cases now pass

## 2. Backend: update existing manager tests

- [x] 2.1 [P] Update `internal/launchctl/user_test.go` GetLogClearStatus expectations to assert `Size` is reported correctly for at least one user-service success case (existing log file with known content) — extends **Query log clear authorization status**
- [x] 2.2 [P] Update `internal/launchctl/system_test.go` GetLogClearStatus expectations to assert `Size` for system-service success and missing-file cases
- [x] 2.3 [P] Update `internal/launchctl/apple_system_test.go` GetLogClearStatus expectations to assert `Size` is populated for Apple system services (read-only domain still produces meaningful Size)

## 3. Frontend types & helpers

- [x] 3.1 [P] Add `size: number` to the `LogClearStatus` interface in `frontend/app/types/wails.d.ts` to match the backend contract — supports **Display log file metadata in Logs tab**
- [x] 3.2 [P] In `frontend/app/components/ServiceLogs.vue` `<script setup>`, add a `formatLogSize(bytes: number): string` helper that returns `0 B` / `512 B` / `1.0 KB` / `2.4 MB` / `1.1 GB` per the boundary table in the spec (1024-base, integer for B, one decimal for KB/MB/GB) — supports **Display log file metadata in Logs tab**
- [x] 3.3 [P] In the same `<script setup>`, add a `truncatePathMiddle(path: string, maxLen: number): string` helper that preserves the basename suffix and inserts `…` in the middle when path length exceeds `maxLen` — supports **Display log file metadata in Logs tab**

## 4. Frontend: render info row in ServiceLogs.vue

- [x] 4.1 Write failing tests in `frontend/app/components/__tests__/ServiceLogs.test.ts` asserting the info row renders the resolved path and formatted size for a user service whose log exists — covers the **Display log file metadata in Logs tab** happy path
- [x] 4.2 Write failing test asserting the info row renders for service-type `apple-system` (must not be skipped) — covers **Display log file metadata in Logs tab** apple-system parity
- [x] 4.3 Write failing test asserting the info row renders `No stdout path configured` when `logPath` is empty, and that the size segment is not rendered — covers **Display log file metadata in Logs tab** empty-path branch
- [x] 4.4 Write failing test asserting the info row renders the resolved path with size `—` when `exists` is false — covers **Display log file metadata in Logs tab** missing-file branch
- [x] 4.5 Write failing test asserting that toggling between stdout and stderr triggers both `loadLogs` and `loadLogClearStatus` so the info row updates with the new path and size — covers **Display log file metadata in Logs tab** refresh behavior
- [x] 4.6 Add the info row markup in `frontend/app/components/ServiceLogs.vue` between the existing controls row and the transient feedback row, wired to `logClearStatus` reactively; bind a tooltip (`title` attribute) on the path element with the unabbreviated path
- [x] 4.7 Remove the apple-system early return in `loadLogClearStatus` (currently around lines 328–333) so the status query runs for all three service types; `clearControlState` already short-circuits `visible: false` for apple-system, so the Clear button stays hidden
- [x] 4.8 Write a failing test asserting that an Auto-refresh poll tick refreshes the info row (both `loadLogs` and `loadLogClearStatus` run per tick), then extend `startPolling`'s tick in `frontend/app/components/ServiceLogs.vue` to call `loadLogClearStatus()` alongside `loadLogs()` — covers **Display log file metadata in Logs tab** auto-refresh parity (the info row adds no timer of its own; it rides the existing one)
- [x] 4.9 Verify all tests in 4.1–4.5 and 4.8 now pass

## 5. Verification

- [x] 5.1 Run `make test` and confirm Go tests, Vitest, and TypeScript typecheck all pass
- [x] 5.2 Run `make lint` and confirm `golangci-lint` and ESLint produce no new findings
- [x] 5.3 Run `make dev` and manually verify the Logs tab on a user service, a system service, and an Apple system service: path appears, size formats correctly, tooltip shows full path on hover, missing-file shows `—`, no-path shows `No stdout path configured`

