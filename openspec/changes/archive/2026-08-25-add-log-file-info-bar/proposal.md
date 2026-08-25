## Why

服務日誌檢視頁面目前只顯示 log 檔案內容，使用者無法在不離開應用程式的情況下知道：(1) launchctl 解析後的實際 log 路徑（特別是 `~` 展開後）、(2) 檔案大小。當 log 檔案沒有產生（例如服務從未啟動）、檔案異常龐大、或使用者想用其他工具開啟時，這兩個資訊是必要的診斷依據。

## What Changes

- 後端 `LogClearStatus` 結構新增 `Size int64` 欄位，由 `logClearStatusFor` 在已開啟的 fd 上 `Stat()` 取得，missing/permission-denied 一律回傳 0
- 前端 `ServiceLogs.vue` 在 stdout/stderr 切換列下方加入一條常駐資訊列，顯示「目前 log type · 解析後路徑（中段省略 + tooltip 顯示完整路徑）· 格式化檔案大小」
- 移除 `ServiceLogs.vue` 對 apple-system 的 `loadLogClearStatus` early return，使三種服務類型（user、system、apple-system）一致顯示資訊列
- 資訊列更新時機：與 logs 一同刷新（mount、切換 logType、按 Refresh），不做主動 polling
- 檔案大小格式化：`B / KB / MB / GB`，1 位小數（例如 `512 B`、`2.4 MB`、`1.1 GB`）；`exists=false` 時顯示 `—`
- 路徑為空（plist 未配置 StandardOutPath/StandardErrorPath）時，整列改顯示 `No <stdout|stderr> path configured`

## Non-Goals

- **不做主動 polling 或 file watcher 推播**：size 只在使用者按 Refresh 或切換 stdout/stderr 時更新；自動更新會引入額外計時器或 FS event 訂閱，目前需求未到那個複雜度
- **不新增路徑點擊互動**：不會把資訊列的路徑做成可點擊（在 Finder 開啟、複製到剪貼簿）；既有 reveal-in-finder capability 已涵蓋類似需求，重複實作會造成 UI 不一致
- **不改名 `LogClearStatus`**：雖然這個結構從「Clear 用」延伸到「log 元資料」載體有命名漂移，但目前沒有第三個用途，重命名留到真有需要時再做以避免無謂的 churn
- **不新增獨立的 `GetLogInfo` Wails binding**：複用 `GetLogClearStatus` 省一次 round-trip，前端已經呼叫該方法
- **不引入新的 i18n 框架**：資訊列文字（`No stdout path configured`、size 單位）依專案規範以英文 hard-code

## Capabilities

### New Capabilities

(none)

### Modified Capabilities

- `core-service-management`: "Read service logs" requirement 擴展為同時提供 log 檔元資料（解析後路徑與檔案大小），供 Logs 分頁的資訊列顯示

## Impact

- Affected specs: `core-service-management`
- Affected code:
  - Modified: `internal/launchctl/types.go`
  - Modified: `internal/launchctl/readonly.go`
  - Modified: `frontend/app/components/ServiceLogs.vue`
  - Modified: `frontend/app/types/wails.d.ts`
  - Modified: `frontend/app/components/__tests__/ServiceLogs.test.ts`
  - Modified: `internal/launchctl/user_test.go`
  - Modified: `internal/launchctl/system_test.go`
  - Modified: `internal/launchctl/apple_system_test.go`
