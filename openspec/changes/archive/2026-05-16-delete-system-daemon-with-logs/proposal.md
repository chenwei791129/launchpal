## Why

刪除 system daemon 後，其 `StandardOutPath` / `StandardErrorPath` 指向的 log 目錄與檔案會原封不動留在磁碟上，造成系統雜訊。使用者在完成 `session-privileged-helper` 的 code review 後提出需求：希望在刪除 daemon 時，可以選擇一併清除對應的 log 檔案與空目錄，讓清理流程更完整。

## What Changes

- **Delete 確認對話框新增 checkbox**：`frontend/app/pages/system.vue` 的 delete dialog 中增加「Also delete log files」checkbox（預設不勾選）。
- **`DeleteSystemService` API 新增 options 參數**：`App.DeleteSystemService(name string, options DeleteServiceOptions)` 接受 `DeleteServiceOptions{DeleteLogs bool}`，取代原本只有 `name` 的簽名。
- **新增 `DeleteLogPaths` privhelper RPC**：helper 接受 `[]string` paths，以 `validateLogPath`（allowlist 驗證 + `O_NOFOLLOW` symlink 防護）刪除各 log 檔，並在 parent 目錄變空後一併刪除該目錄。
- **`SystemManager.Delete` 串接新流程**：若 `options.DeleteLogs` 為 `true`，在 `DeletePlist` 成功後，解析 plist 的 `StandardOutPath` / `StandardErrorPath`，呼叫 `DeleteLogPaths` RPC。
- **TypeScript 型別同步**：`frontend/app/types/wails.d.ts` 新增 `DeleteServiceOptions` 型別，並更新 `DeleteSystemService` 函式簽名。

## Non-Goals

- User services（`~/Library/LaunchAgents`）的 log 清除：User 有完整的 home 目錄寫入權，log 通常在 user 自己管理的路徑；此 change 僅涵蓋 system domain，user services 的 log 清除列為 future extension。
- Log 檔案備份：log 體積通常很大且備份成本高；本 change 刪除的 log 不加入備份流程。這是刻意的 trade-off — plist 備份已由 `DeletePlist` handler 在 helper 端完成，log 屬於執行期產物，使用者勾選後視同知情同意不可復原。
- Apple System Services（`/System/Library/LaunchDaemons`）：SIP 保護下一般無法刪除相關系統 log，不在本 change 範圍。
- Log 內容預覽或確認：UI 只顯示 checkbox，不預覽將被刪除的路徑清單（避免 UX 過重）。

## Capabilities

### New Capabilities

- `delete-log-files-on-service-removal`: 刪除 system daemon 時，以 privhelper 安全刪除 plist 內 `StandardOutPath` / `StandardErrorPath` 指向的 log 檔案及清空的 parent 目錄，整合至 Delete 確認 UI 的可選 checkbox 流程。

### Modified Capabilities

- `launchdaemons-readonly`: `SystemManager.Delete` 新增 `options` 參數，Admin Mode 啟用後的刪除流程可選擇性地呼叫 `DeleteLogPaths` RPC；spec 需新增此行為的需求描述。

## Impact

- 受影響的 specs：`launchdaemons-readonly`（修改）、`delete-log-files-on-service-removal`（新增）
- 受影響的程式碼：
  - `internal/privhelper/protocol.go` — 新增 `DeleteLogPaths` method 常數與 Request/Response 型別
  - `internal/privhelper/handlers.go` — 新增 `handleDeleteLogPaths` handler，實作 allowlist 驗證、`O_NOFOLLOW` 刪檔、空目錄清除
  - `internal/privhelper/client.go` — 新增 `DeleteLogPaths(paths []string) error` client wrapper
  - `internal/launchctl/system.go` — `SystemManager.Delete` 新增 `options DeleteServiceOptions` 參數，串接 `DeleteLogPaths`
  - `internal/launchctl/manager.go` — `Manager` interface `Delete` 簽名更新（或新增 `DeleteServiceOptions` 至 `user.go`）
  - `app.go` — `DeleteSystemService` 新增 `options DeleteServiceOptions` 參數
  - `frontend/app/pages/system.vue` — delete dialog 新增 checkbox
  - `frontend/app/types/wails.d.ts` — 新增 `DeleteServiceOptions` 型別，更新 `DeleteSystemService` 簽名
