## 1. Frontend: locale fix in formatters.ts

- [x] 1.1 Add a unit test under `frontend/app/utils/__tests__/formatters.test.ts` asserting `formatTimestamp` produces `YYYY-MM-DD HH:mm:ss` (no Asian-style year/month markers, no Chinese characters in output)
- [x] 1.2 Replace `'zh-TW'` with `'en-CA'` in `frontend/app/utils/formatters.ts` so the formatter test from 1.1 passes
- [x] 1.3 Run `cd frontend && pnpm vitest run app/utils/__tests__/formatters.test.ts` to confirm the new test passes; spot-check Settings → Backup History UI shows ISO-style timestamps after `make dev`

## 2. Frontend: replace alert() with inline error state

- [x] 2.1 [P] In `frontend/app/pages/system.vue`, introduce an `actionError` ref (mirror `settings.vue`'s `restoreError` pattern) and replace `alert()` calls at lines 247 and 255 with assignments to that ref; render the message in a dark-themed inline error region consistent with the rest of the page
- [x] 2.2 [P] In `frontend/app/components/ReadOnlyServiceList.vue`, replace the `alert()` at line 177 with an emitted error event or a local error ref consumed by the parent page; ensure the visual style matches `apple-system.vue` / existing read-only list affordances
- [x] 2.3 Verify `grep -rn "alert(" frontend/app` returns no remaining occurrences inside templates or `<script setup>` blocks (other than intentional comment references, if any)
- [x] 2.4 Run `cd frontend && pnpm vitest run` and `cd frontend && pnpm typecheck` to confirm no regressions

## 3. Go backend: remove dead privhelper RPC surface

- [x] 3.1 In `internal/privhelper/protocol.go`, delete `MethodListSystemDaemons` and `MethodGetSystemDaemon` constants, the `DaemonInfo` struct, `GetSystemDaemonParams`, and `ListSystemDaemonsResult`; remove their entries from the `AllMethods` table
- [x] 3.2 In `internal/privhelper/client.go`, delete `Client.ListSystemDaemons` and `Client.GetSystemDaemon` methods
- [x] 3.3 In `internal/privhelper/handlers.go`, remove the `MethodListSystemDaemons` / `MethodGetSystemDaemon` dispatch cases, the `launchctlLister.ListDaemons` / `GetDaemon` methods (and the corresponding entries on the `Lister` interface), the `parsePrintSystem` and `parsePrintService` parser functions, and any now-orphaned helper code surfaced by the deletions
- [x] 3.4 Update `internal/privhelper/protocol_test.go`, `internal/privhelper/client_test.go`, and `internal/privhelper/handlers_test.go` to drop tests targeting the removed APIs; ensure remaining tests still cover `Ping`, `Bootstrap`, `Bootout`, `Kickstart`, `WritePlist`, `DeletePlist`, `EnsureLogAccess`, `TruncateLog`
- [x] 3.5 Run `go build ./...`, `go test ./internal/privhelper/...`, and `make lint` to confirm no callers were missed and no dead imports remain
- [x] 3.6 Search the repo for residual references with `grep -rn "ListSystemDaemons\|GetSystemDaemon\|DaemonInfo\|parsePrintSystem\|parsePrintService" --include="*.go"`; result must be empty

## 4. Go backend: remove UserManager.Stop pgrep+kill fallback

- [x] 4.1 Add or extend a test in `internal/launchctl/user_test.go` that documents the new contract: after `Stop`, the implementation MUST NOT invoke `pgrep` / `kill` outside of `launchctl bootout` (use a fake `exec` shim or assert on observed command list)
- [x] 4.2 In `internal/launchctl/user.go`, delete lines 343–354 (the `pgrep -f <program>` + `kill <pid>` block) so `Stop` ends after the `launchctl bootout` invocation; do NOT add a `launchctl kill SIGTERM` replacement in this change
- [x] 4.3 Run `go test ./internal/launchctl/...` and `make lint` to confirm the contract test passes and no callers depended on the removed behavior

## 5. Docs: fill in Purpose for 16 specs

For each spec, derive a 1–3 sentence Purpose from the existing Requirements/Scenarios; do NOT change requirements or scenarios. Spec files MUST stay in English.

- [x] 5.1 [P] Build/release specs — fill Purpose for `openspec/specs/build-version-injection/spec.md`, `openspec/specs/homebrew-auto-release/spec.md`, `openspec/specs/homebrew-cask-formula/spec.md`
- [x] 5.2 [P] Helper-lifecycle specs — fill Purpose for `openspec/specs/admin-mode/spec.md`, `openspec/specs/privileged-helper-lifecycle/spec.md`, `openspec/specs/privileged-helper-rpc/spec.md`
- [x] 5.3 [P] Service-feature specs — fill Purpose for `openspec/specs/kickstart-service/spec.md`, `openspec/specs/scheduled-service/spec.md`, `openspec/specs/cron-range-expansion/spec.md`, `openspec/specs/shell-arguments-parsing/spec.md`, `openspec/specs/clear-service-logs/spec.md`
- [x] 5.4 [P] UI/preview specs — fill Purpose for `openspec/specs/backup-diff-preview/spec.md`, `openspec/specs/env-vars-ui/spec.md`, `openspec/specs/next-run-preview/spec.md`, `openspec/specs/reveal-in-finder/spec.md`, `openspec/specs/wake-system/spec.md`
- [x] 5.5 [P] System-daemon write ops — fill Purpose for `openspec/specs/system-daemon-write-ops/spec.md`
- [x] 5.6 Verify `grep -rln "TBD - created by archiving" openspec/specs/` returns nothing

## 6. Docs: README — Run Now (Kickstart) feature bullet

- [x] 6.1 Add a "Run Now (Kickstart) — start a service immediately with one click" bullet to the Features section of `README.md`, placed near the existing scheduling / Clear Logs entries to preserve the current ordering

## 7. Final verification

- [x] 7.1 Run `make test` (Go tests + frontend vitest + TypeScript typecheck) and confirm green
- [x] 7.2 Run `make lint` (golangci-lint + eslint) and confirm green
- [x] 7.3 Run `spectra validate simplification-pass-2026-05` and confirm no errors
- [x] 7.4 Run `make build` to confirm the production app still bundles cleanly with the helper binary
