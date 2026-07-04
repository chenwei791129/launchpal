## 1. 前置確認

- [x] 1.1 確認前置 change fix-log-error-classification 已實作完成：frontend/app/components/ServiceLogs.vue 的 `loadLogs` 已依 `LogsResult` 的小寫 `status` 鍵分流、rejection 訊息原文顯示。驗證：直接檢查程式碼 — grep frontend/app/components/ServiceLogs.vue 出現 `no-path` / `not-found` 狀態分支、grep frontend/app/types/wails.d.ts 的 `GetLogs` 回傳 `Promise<LogsResult>`，且 `make test` 全綠（該 change 目前 parked 於 `spectra list --parked`，不會出現在 `spectra list`；若上述 grep 落空即表示前置未完成，本 change 不可開工）。

## 2. 測試先行（TDD）

- [x] 2.1 在 frontend/app/components/__tests__/ServiceLogs.test.ts 以 `vi.useFakeTimers` 新增 "Auto-refresh toggle in the Logs tab" requirement 的測試（先寫、先失敗）：預設未勾選且不輪詢、與 Auto-scroll 互相獨立（依 design.md「Decision 4: 開關與 Auto-scroll 正交」，只開 Auto-refresh 時刷新後不強制捲底、兩者同開時每次輪詢刷新後捲到底）。使用 fake timers 的測試組必須在 teardown 呼叫 `vi.useRealTimers()`（既有 `afterEach` 只有 `vi.restoreAllMocks()`，否則 fake timers 外洩會使 `confirmClear` 的真實 `setTimeout(2000)` 測試停擺）。驗證：新測試於實作前失敗。
- [x] 2.2 同檔新增 "Periodic reload while Auto-refresh is enabled" requirement 的測試（先寫、先失敗）：勾選後每 2000ms 觸發一次 `loadLogs` 路徑（依 design.md「Decision 1: 前端 setInterval 輪詢，固定 2 秒」）、in-flight tick 跳過不併發（依「Decision 2: in-flight tick 跳過」）、取消勾選即停止、unmount 清理 interval 不再發出請求、`serviceName` prop 變更（服務間導航、元件未 remount）時開關重置為關閉且輪詢停止、切換 stdout/stderr 保留開關並對新 stream 續輪詢（依「Decision 5: 切換 logType 時保留開關狀態」）。驗證：新測試於實作前失敗。
- [x] 2.3 同檔新增 "Auto-refresh disables itself on a non-ok load outcome" requirement 的測試（先寫、先失敗）：依 design.md「Decision 3: 非 ok 結果自動關閉開關」，分別對 `status: "no-path"`、`status: "not-found"`、promise rejection、無 bindings 的 development fallback 四種結果斷言 checkbox 變為未勾選、輪詢停止、且畫面回饋維持 log-load-feedback 的既有分流（placeholder / 錯誤訊息），不新增額外錯誤 UI；觸發來源除輪詢 tick 外，另涵蓋手動 Refresh 失敗與切換 stream 至 no-path stream 兩種路徑；並斷言不會自動恢復輪詢。驗證：新測試於實作前失敗。

## 3. 實作

- [x] 3.1 修改 frontend/app/components/ServiceLogs.vue：新增 "Auto-refresh" checkbox（英文文案，與 Auto-scroll 並列、狀態獨立、元件層級不持久化）、`setInterval` 2000ms 輪詢與 `clearInterval` 生命週期（停用 / unmount 時清理，沿用既有 `clearSuccessTimeout` 的清理慣例）、新增 `watch(() => props.serviceName)` 於服務間導航（元件未 remount）時關閉開關並清理 interval、interval callback 以 `loading.value` 防護 in-flight 跳過、自動關閉判斷放在 `loadLogs` 收尾的共用路徑（讀取小寫 `status` 鍵；非 `ok`、rejection、或 development fallback 且開關啟用時關閉開關並停止輪詢），三種觸發來源（輪詢 / 手動 Refresh / 切 stream）行為一致。驗證：2.1–2.3 測試全數轉綠。

## 4. 整體驗證

- [x] 4.1 執行 `make test` 與 `make lint`，全數通過且既有 ServiceLogs 測試（分流、ANSI 渲染、Clear Logs）不回歸。驗證：兩命令 exit code 0。
- [x] 4.2 請使用者以 `make dev` 啟動應用進行手動驗證：對一個持續寫 log 的 user service 勾選 Auto-refresh，確認新行約每 2 秒自動出現；搭配 Auto-scroll 開/關各驗證一次捲動行為；於 apple-system 服務勾選 Auto-refresh，確認因 no-path 立即自動關閉且顯示 placeholder；離開頁面後確認無殘留輪詢（Console/網路無週期性活動）。
