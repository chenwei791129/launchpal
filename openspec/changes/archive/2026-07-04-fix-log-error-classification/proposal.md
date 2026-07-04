## Why

Service 詳細頁的 Logs tab 在後端 `GetLogs` / `GetSystemLogs` 失敗時，一律顯示籠統的紅字「Failed to load logs」。實測 apple-system 服務時，411 個服務中僅 3–6 個有配置 `StandardOutPath` / `StandardErrorPath`，其餘全部落入「no log path configured」錯誤；而僅有的幾個配置也指向 `/dev/console`、`/dev/null` 或不存在的檔案。使用者無法分辨失敗原因是權限不足、路徑未配置、還是檔案不存在。

問題有兩層：

1. **前端吃掉真實錯誤訊息**：`ServiceLogs.vue` 的 `loadLogs` 以 `e instanceof Error ? e.message : 'Failed to load logs'` 取錯誤訊息，但 Wails v2 將 Go error 以「字串」reject Promise，`instanceof Error` 永遠為 false，後端訊息（如 `no stdout log path configured for service X`）被 fallback 字串蓋掉。
2. **結構性狀態被當成錯誤**：「未配置 log path」與「log 檔尚不存在」是服務的正常狀態（絕大多數 apple-system 服務如此），不應以紅字錯誤呈現，而應落入既有的「No logs available」placeholder 分支，並附上具體原因。

## What Changes

- **BREAKING（內部 API）**：`Manager.GetLogs` 的回傳型別由 `(string, error)` 改為 `(LogsResult, error)`。`LogsResult` 為新的結構化回傳型別，包含 log 內容、狀態分類（`ok` / `no-path` / `not-found`）與解析後的 log 路徑。三個 Manager 實作（`UserManager`、`SystemManager`、`AppleSystemManager`）與 Wails bindings（`App.GetLogs`、`App.GetSystemLogs`）同步更新。此為單一 repo 內前後端同步修改，無外部消費者。
- 結構性狀態（未配置路徑、檔案不存在）不再以 Go error 回傳，改以 `LogsResult.Status` 表達；真正的失敗（權限不足、I/O 錯誤、無效 logType、服務不存在）仍以 error 回傳。
- 前端 `ServiceLogs.vue` 依 `LogsResult.Status` 分流：`ok` 顯示內容（空內容維持既有 placeholder）；`no-path` / `not-found` 顯示帶原因的 placeholder（非紅字錯誤）；Promise rejection 則將後端訊息原文顯示於紅字錯誤分支（同時修正字串 rejection 被 `instanceof Error` 判斷丟棄的缺陷）。
- 前端型別定義 `wails.d.ts` 新增 `LogsResult` 介面並更新兩個 binding 的回傳型別。

## Capabilities

### New Capabilities

- `log-load-feedback`: Logs tab 載入結果的前端呈現分類 — 結構化狀態對應 placeholder 文案、rejection 對應紅字錯誤且訊息原文透傳。

### Modified Capabilities

- `core-service-management`: 「Read service logs」requirement — user domain 的 `GetLogs` 改為回傳結構化 `LogsResult`；未配置路徑與檔案不存在由 error 改為狀態分類。「Clear service logs」requirement — 「subsequent `GetLogs` returns an empty string」場景措辭同步更新為新契約（`Status: "ok"` 且 `Content` 為空）。
- `launchdaemons-readonly`: 「Read system service logs」requirement — system / apple-system domain 的 `GetLogs` 同步改為結構化回傳與狀態分類。
- `ansi-log-rendering`: 「Mount rendered output in ServiceLogs view」requirement — `renderedLogs` 的來源由 `logs.value` 字串改為 `LogsResult` 的 content；「branches SHALL remain unchanged」的措辭更新為允許新增 `no-path` / `not-found` placeholder 分支。

## Impact

- Affected specs: `log-load-feedback`（新增）、`core-service-management`（修改）、`launchdaemons-readonly`（修改）、`ansi-log-rendering`（修改）
- Affected code:
  - Modified: internal/launchctl/types.go（新增 `LogsResult` 型別、`readLogTail` 錯誤分類配合調整）
  - Modified: internal/launchctl/manager.go（`Manager` 介面 `GetLogs` 簽名）
  - Modified: internal/launchctl/user.go（`UserManager.GetLogs`）
  - Modified: internal/launchctl/readonly.go（`readOnlyManager.getLogs`，SystemManager 與 AppleSystemManager 共用）
  - Modified: internal/launchctl/system.go（`SystemManager.GetLogs` wrapper 簽名，含 `var _ Manager` 介面斷言）
  - Modified: internal/launchctl/apple_system.go（`AppleSystemManager.GetLogs` wrapper 簽名，含 `var _ Manager` 介面斷言）
  - Modified: app.go（`GetLogs` / `GetSystemLogs` bindings 回傳型別）
  - Modified: frontend/app/types/wails.d.ts（`LogsResult` 介面與 binding 簽名）
  - Modified: frontend/wailsjs/go/main/App.d.ts（產生的 binding 簽名，`wails generate module` 重新產生或手動同步）
  - Modified: frontend/wailsjs/go/main/App.js（同上）
  - Modified: frontend/wailsjs/go/models.ts（產生的 `LogsResult` model，同上）
  - Modified: frontend/app/components/ServiceLogs.vue（`loadLogs` 分流、placeholder 文案、rejection 訊息透傳）
  - Modified: internal/launchctl/user_test.go（GetLogs 測試更新為狀態斷言）
  - Modified: internal/launchctl/system_test.go（`TestSystemManager_GetLogs` 更新為 `LogsResult` 斷言）
  - Modified: internal/launchctl/apple_system_test.go（`TestAppleSystemManager_GetLogs` 更新為 `LogsResult` 斷言）
  - Modified: frontend/app/components/__tests__/ServiceLogs.test.ts（狀態分流與 rejection 訊息測試）
- 不影響 `GetLogClearStatus` / `ClearLogs` 路徑；Clear Logs 後重新載入沿用新的 `loadLogs` 分流。
