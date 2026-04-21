## Why

目前 `SystemManager` 與 `AppleSystemManager` 的狀態查詢依賴 `launchctl list`，而該命令只列出當前 user 的 GUI session 服務，完全查不到 `/Library/LaunchDaemons` 與 `/System/Library/LaunchDaemons` 下的 system domain 服務。結果是 UI 永遠把這些服務顯示為 `Stopped`、`PID = 0`，使用者無法判斷 system daemon 是否真的在運行，大幅降低 LaunchPal 的實用性。

因為 LaunchPal 未取得 Apple Developer ID，無法使用 `SMAppService` 建立 privileged helper，本變更採完全不需提權的啟發式偵測 — 結合 plist 的 `UserName` 篩選與 `ppid=1`（launchd parent）過濾，在絕大多數情境下準確。當偵測結果歧義（多個候選 PID）時，保守標記為 `StatusConfidence = "unverified"` 讓 UI 揭示不確定性，但**不提供任何提權校準通道**；真正需要管理 system daemon 的使用者會在 Phase 2（`session-privileged-helper`）取得 Admin Mode 後獲得 authoritative 狀態。

## What Changes

- 新增一套 **system-scope 狀態偵測邏輯**，依序：
  1. 讀取 plist 的 `UserName`（預設 `root`）與可執行檔路徑（`Program` 或 `ProgramArguments[0]`）
  2. 執行 `pgrep -u <UserName> -f <Program>` 取得候選 PID
  3. 用 `ps -o ppid= -p <pid>` 過濾出 `ppid == 1`（由 launchd 起）的 PID
  4. 一個候選 → `Running`；零個 → `Stopped`；多個 → `Running` 但信心度為 `unverified`
- `readOnlyManager.getWithStatus` 於批次查詢沒命中時，改呼叫新邏輯而非直接設為 `Stopped`
- `Service` struct 新增 `StatusConfidence` 欄位（`verified` / `unverified`），讓前端決定是否標示歧義
- 前端 System / Apple System 列表與詳情頁：`StatusConfidence == "unverified"` 時於 Status 欄位旁顯示 info icon + tooltip 告知「偵測到多個候選 PID，顯示的 PID 可能不是 launchd 管理的那個」，**不提供按鈕或進一步動作**
- 保留既有 `commonShells` (`bash`/`sh`/`zsh`) 誤判 skip 邏輯
- `SystemManager` / `AppleSystemManager` **仍維持唯讀**，本變更不觸及任何寫入操作
- **不引入任何提權通道、不呼叫 osascript、不新增 Verify 相關 Wails binding**

## Non-Goals

- 不支援 Start / Stop / Restart / Create / Update / Delete 等寫入操作（留待 Phase 2 `session-privileged-helper` 處理）
- 不建立持久 root helper daemon 或 Admin Mode
- 不改動 `UserManager` 既有狀態偵測（user domain 的 `launchctl list` 已足夠）
- 不透過 Developer ID 簽章 / `SMAppService` 路徑（已確認取得不到）
- **不提供任何單次提權校準（Verify）機制**：`StatusConfidence = "unverified"` 僅作為資訊性標記，無升級為 `verified` 的通道

## Capabilities

### New Capabilities

- `system-daemon-status-detection`: 以 `pgrep -u` + `ppid=1` 啟發式偵測 system domain daemon 狀態、PID 與信心度，完全不需提權

### Modified Capabilities

- `launchdaemons-readonly`: 狀態查詢行為由「查無 label 即 Stopped」改為「查無 label 時改走啟發式偵測」，並對外暴露 `StatusConfidence`

## Impact

- Affected specs:
  - 新增 `openspec/specs/system-daemon-status-detection/spec.md`
  - 修改 `openspec/specs/launchdaemons-readonly/spec.md`
- Affected code:
  - `internal/launchctl/readonly.go`：`getWithStatus` 於 statusMap 未命中時改走新邏輯
  - `internal/launchctl/user.go`：`getServiceStatus` 原有 pgrep fallback 的邏輯抽出供 system 共用（或新檔 `status_detect.go`）
  - `internal/launchctl/types.go`：`Service` 新增 `StatusConfidence` 欄位；新增 `ConfidenceVerified` / `ConfidenceUnverified` 常數
  - `internal/launchctl/system.go` / `apple_system.go`：確認走新偵測路徑
  - `frontend/app/types/`：`Service` TypeScript 型別新增 `statusConfidence`
  - `frontend/app/pages/system.vue`、`frontend/app/pages/apple-system.vue`、`frontend/app/pages/services/[name].vue`：unverified 狀態的 info icon + tooltip 顯示
  - `frontend/app/composables/`：若有 service 相關 composable 需傳遞新欄位
- 相依工具：需 `pgrep`、`ps`、`launchctl`（皆為 macOS 內建）
