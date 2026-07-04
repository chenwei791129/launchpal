## Why

Service 詳細頁的 Logs tab 目前只在進入頁面、切換 stdout/stderr、或手動點 Refresh 時載入日誌，沒有任何自動刷新機制。觀察執行中服務的輸出（類似 `tail -f` 的場景）必須反覆手動點擊 Refresh，體驗繁瑣。此前的討論已確認這是未實作的功能而非 bug。

## What Changes

- Logs tab 新增一個 **Auto-refresh** 開關（與既有 Auto-scroll checkbox 並列且互相獨立）：啟用時每 2 秒透過既有的 `loadLogs()` 路徑重新載入當前 stream 的日誌；停用、元件 unmount、或服務間導航（`serviceName` prop 變更而元件未 remount）時停止輪詢，後者同時將開關重置為關閉。
- 純前端實作：復用既有 `GetLogs` / `GetSystemLogs` bindings 整段重讀 tail（後端已封頂 1MB），不新增任何 IPC、不引入 fsnotify push、不做 offset 增量 tail。
- 輪詢期間若前一次載入尚未完成（in-flight），該次 tick 跳過，不堆疊請求。
- 依賴 change `fix-log-error-classification` 的狀態分類：當載入結果為非 `ok`（`no-path` / `not-found` / rejection / 無 bindings 的 development fallback）時，Auto-refresh 自動關閉（開關回到未勾選狀態），避免對結構性失敗做無意義的重複讀取或反覆報錯；此行為對輪詢、手動 Refresh、切換 stream 三種觸發來源一致。
- Auto-refresh 啟用且 Auto-scroll 同時勾選時，每次刷新後維持既有的捲動到底行為；Auto-scroll 未勾選時保留使用者目前的捲動位置。
- 開關狀態為元件層級（session 內、非持久化），預設關閉。

## Capabilities

### New Capabilities

- `log-auto-refresh`: Logs tab 的自動刷新輪詢 — 開關、2 秒週期、in-flight 跳過、unmount 停止、非 ok 結果自動關閉。

### Modified Capabilities

(none)

## Impact

- Affected specs: `log-auto-refresh`（新增）
- Affected code:
  - Modified: frontend/app/components/ServiceLogs.vue（Auto-refresh 開關、輪詢計時器與生命週期、非 ok 自動關閉）
  - Modified: frontend/app/components/__tests__/ServiceLogs.test.ts（fake timers 輪詢測試）
- 無 Go 後端變更、無 Wails bindings 變更、無 settings 變更。
- **前置依賴**：change `fix-log-error-classification`（LogsResult 狀態分類）必須先實作完成 — 自動關閉條件建立在 `status` 欄位上。
