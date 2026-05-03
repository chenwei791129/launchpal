## Why

GitHub issue #18 指出使用者目前無法從 LaunchPal 內清除某個服務累積的 stdout/stderr log 檔案，必須跳到 Terminal 手動 truncate 才行。長時間執行的 daemon log 容易膨脹到數百 MB 甚至吃光磁碟，這項功能要直接讓使用者在服務 detail 頁面的 Logs tab 一鍵清空當前 log，並在 system 服務上根據檔案實際寫入權限決定是否需要 Admin Mode。

## What Changes

- 在 `ServiceLogs.vue` 的控制列新增「Clear Logs」按鈕，作用於目前選中的 log type（stdout 或 stderr 二選一，不同時清兩個）。
- 點擊按鈕後彈出與 Run Now 同款的 Teleport 確認 modal（紅色強調），確認後才實際截斷檔案。
- 新增 Wails binding：
  - `ClearLogs(name, logType) error` — user 服務截斷 log 檔。
  - `ClearSystemLogs(name, serviceType, logType) error` — system 服務截斷 log 檔；apple-system 一律拒絕。
  - `GetLogClearStatus(name, serviceType, logType) (LogClearStatus, error)` — 提供前端決定按鈕可用性的資訊（log path、檔案是否存在、使用者是否可寫）。
- 後端對 system 服務做 per-file 寫入權限分流：使用者對 log 檔有 W_OK 時直接 truncate；不可寫時走 privileged helper 的新 RPC `TruncateLog`。
- privileged helper 新增 `TruncateLog` RPC method，沿用既有 `validateLogPath` allowlist 與 `O_NOFOLLOW` 安全 pattern。
- 按鈕可見性與可用性：
  - user 服務：永遠顯示且可用。
  - system 服務：永遠顯示；可用條件為「log 存在且使用者可寫」或「log 存在且 Admin Mode 已啟用」；不滿足時 disabled 並透過 tooltip 說明原因（`No log path configured` / `Log file does not exist` / `Enable Admin Mode to clear`）。
  - apple-system 服務：完全隱藏按鈕。
- 後端在實際截斷時必須再驗一次權限，避免前端 UI 與實際執行之間的 TOCTOU race。

## Non-Goals

- 不提供「同時清 stdout + stderr」的整合按鈕；維持與 Refresh 一致，按鈕只作用於目前 tab。
- 不提供 log rotation / archive 功能（清掉就是清掉，不另存副本）。
- 不為 apple-system 服務開放任何寫操作，包含理論上 user-writable 的 log 檔；維持「Apple 服務一律不碰」的整體原則。
- 不為 user 服務做 per-file 權限檢查；user 服務的 log 預設位於使用者可寫位置，現行 `GetLogs` 也是這個假設。
- 不在 `ListServices` 階段預取 log 寫入權限；只在 detail 頁開啟或 `logType` 切換時 lazily 查詢，避免影響清單載入效能。

## Capabilities

### New Capabilities

- `clear-service-logs`: 一鍵清空特定服務 log 檔案（stdout 或 stderr）的能力，含按鈕可見性矩陣、權限分流策略、確認 modal 行為。

### Modified Capabilities

- `core-service-management`: 在 Manager 介面新增 `ClearLogs(name, logType)` 與 `GetLogClearStatus(name, logType)` 行為要求。
- `system-daemon-write-ops`: 新增「截斷 system daemon log」的寫入操作以及 per-file 寫入權限分流的判斷邏輯。
- `privileged-helper-rpc`: 新增 `TruncateLog` RPC method 規格（允許路徑、O_NOFOLLOW、錯誤碼）。

## Impact

- Affected specs: `clear-service-logs`（新）、`core-service-management`、`system-daemon-write-ops`、`privileged-helper-rpc`
- Affected code:
  - New:
    - openspec/specs/clear-service-logs/spec.md
  - Modified:
    - app.go
    - internal/launchctl/manager.go
    - internal/launchctl/user.go
    - internal/launchctl/system.go
    - internal/launchctl/apple_system.go
    - internal/launchctl/readonly.go
    - internal/launchctl/types.go
    - internal/privhelper/protocol.go
    - internal/privhelper/handlers.go
    - internal/privhelper/client.go
    - frontend/app/components/ServiceLogs.vue
    - frontend/app/pages/services/[name].vue
    - frontend/app/types/wails.d.ts
    - .claude/CLAUDE.md
  - Removed: (none)
