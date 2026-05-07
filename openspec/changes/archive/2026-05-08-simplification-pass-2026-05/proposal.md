## Why

`.local/2026-05-03-simplification-audit.md` 對 Go backend、Vue frontend 與 Markdown docs 進行了完整審查，發現整體 codebase 健康，但累積了少量「事實性錯誤」與「文件債」。本 change 一次性清掉其中第一層（正確性 / 規則違反）與第二層（低成本債務）共 6 項，避免進入下一個 feature cycle 時帶著這些問題前進。所有項目皆為實作面修正，不改變任何使用者可見功能或 spec 行為。

## What Changes

- 將 `frontend/app/utils/formatters.ts` 的 `toLocaleString('zh-TW', ...)` 改為 `'en-CA'`（ISO-style `YYYY-MM-DD HH:mm:ss`），符合專案 English-only UI 規則
- 刪除 `internal/privhelper` 中從未被 `app.go` 呼叫的死 RPC 表面：`MethodListSystemDaemons`、`MethodGetSystemDaemon`、`DaemonInfo`、`GetSystemDaemonParams`、`ListSystemDaemonsResult`、`Client.ListSystemDaemons`、`Client.GetSystemDaemon`、handler dispatch、`launchctlLister.ListDaemons` / `GetDaemon`、`parsePrintSystem`、`parsePrintService`，以及 `AllMethods` 表中對應 entries
- 移除 `internal/launchctl/user.go` `UserManager.Stop` 中的 pgrep + kill fallback（`user.go:343-354`），避免 `kill <pid>` 誤殺 argv 含 `service.Program` 字串的不相關行程；保留 `launchctl bootout` 為唯一停止路徑
- 補上 16 個 spec 的 `## Purpose` 佔位符（被 `/spectra-archive` 留下的 `"TBD - created by archiving change '...'."`）。受影響 specs：除 `dmg-packaging`、`core-service-management`、`launchdaemons-readonly`、`system-daemon-status-detection`、`system-daemon-write-ops` 已填寫外的 16 個
- 將 `frontend/app/pages/system.vue:247, 255` 與 `frontend/app/components/ReadOnlyServiceList.vue:177` 共 3 處的 `alert()` 改為 inline error state，視覺上與 `settings.vue` 的 `restoreError` ref 模式一致
- 在 `README.md` 補上 v1.9.0 漏記的 "Run Now (Kickstart)" feature bullet

## Non-Goals (optional)

- 不處理 audit report 中的「可以做」第 5、7、8、9 項（`EnvVarEditor` 抽出、privhelper E2E 整合測試、`[name].vue` 拆分、`ServiceList` 抽出）— 這些屬於 refactor 性質，留到後續獨立 change
- 不處理「可略過」清單中的純美容項目（`sort.Slice` → `slices.SortFunc`、`@trace` 區塊重複等）
- 不引入 i18n 框架 — 僅修正單一 locale 字串
- 不改變任何 capability 的 spec 行為，因此無 spec deltas

## Capabilities

### New Capabilities

(none)

### Modified Capabilities

(none — 本 change 為純粹的實作修正、死碼刪除與文件補正，不改變任何 spec 定義的行為)

## Impact

- Affected specs: 無 spec deltas（修改 `openspec/specs/*/spec.md` 的 Purpose 佔位符屬於文件修補，不改變 Requirements / Scenarios）
- Affected code:
  - Modified:
    - frontend/app/utils/formatters.ts
    - frontend/app/pages/system.vue
    - frontend/app/components/ReadOnlyServiceList.vue
    - internal/privhelper/protocol.go
    - internal/privhelper/client.go
    - internal/privhelper/handlers.go
    - internal/launchctl/user.go
    - README.md
    - openspec/specs/admin-mode/spec.md
    - openspec/specs/backup-diff-preview/spec.md
    - openspec/specs/build-version-injection/spec.md
    - openspec/specs/clear-service-logs/spec.md
    - openspec/specs/cron-range-expansion/spec.md
    - openspec/specs/env-vars-ui/spec.md
    - openspec/specs/homebrew-auto-release/spec.md
    - openspec/specs/homebrew-cask-formula/spec.md
    - openspec/specs/kickstart-service/spec.md
    - openspec/specs/next-run-preview/spec.md
    - openspec/specs/privileged-helper-lifecycle/spec.md
    - openspec/specs/privileged-helper-rpc/spec.md
    - openspec/specs/reveal-in-finder/spec.md
    - openspec/specs/scheduled-service/spec.md
    - openspec/specs/shell-arguments-parsing/spec.md
    - openspec/specs/wake-system/spec.md
  - New: (none)
  - Removed: (none — privhelper 死碼為程式碼片段刪除，非整檔刪除)
- Affected tests: 既有 `internal/privhelper/*_test.go` 中針對被刪除 RPC 的測試需同步移除；既有 `internal/launchctl/user_test.go` 對 `Stop` 行為的斷言需更新以反映 fallback 已移除
