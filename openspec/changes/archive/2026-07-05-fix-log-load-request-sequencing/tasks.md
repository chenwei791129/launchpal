## 1. 測試先行（TDD）

- [x] 1.1 在 frontend/app/components/__tests__/ServiceLogs.test.ts 新增 "Discard superseded concurrent log load results" 的測試（先寫、先失敗），使用既有的 fake-timer 慣例（`useAutoRefreshFakeTimers` / `settle` / `clickToggle` helper）並在 teardown 沿用 `vi.useRealTimers()`。以可控 Promise（各自持有 resolve/reject handle，例如 `GetLogs` 依 logType 回傳不同的手動控制 Promise）製造「舊 stdout 載入尚未 resolve 時切到 stderr，stderr 先 resolve、stdout 後 resolve」的亂序情境：斷言 (a) 畫面渲染 stderr 內容、stdout 稍後 resolve 不覆蓋；(b) Auto-refresh 啟用下、被取代的 stdout 載入即使以非 ok status resolve，也不觸發 auto-disable（checkbox 維持勾選）；(c) 另一案：被取代的載入稍後 reject 時，紅色 error 分支不被污染（`.text-red-400` 不出現該錯誤訊息）。依 design.md「Decision 1/2」。驗證：新測試於實作前失敗。

## 2. 實作

- [x] 2.1 修改 frontend/app/components/ServiceLogs.vue，實作 design.md 的 **Decision 1: Monotonic request-sequence token** — 在 `<script setup>` 內新增 module-scoped 單調 request-sequence token `loadSeq`（初始 0），於 `loadLogs()` 進入時同步 `const seq = ++loadSeq`（在任何 await 之前）。實作 **Decision 2: Gate shared-state mutation on the sequence check** — 在 binding settle 後，僅當 `seq === loadSeq` 時才 mutate 共用狀態：resolve 路徑的 `logs.value` 指派與 `loadOk` 驅動的 Auto-refresh auto-disable、catch 路徑的 `error.value` 指派、以及 `finally` 內的 `loading.value = false` 重置皆以此 sequence check 為前提；被取代（superseded）的載入靜默捨棄、不改任何共用狀態。不新增 props / emits / bindings。依 **Decision 3: Leave the nextTick/scroll micro-window unguarded**，`await nextTick()` 後的 `scrollToBottom()` 不另加守衛。驗證：1.1 測試全數轉綠。

## 2b. 延伸：Clear-button 狀態查詢的同類 race（code-review finding ③）

- [x] 2b.1 在 frontend/app/components/__tests__/ServiceLogs.test.ts 新增測試（先寫、先失敗）：以可控 Promise 讓 `GetLogClearStatus` 依 logType 回傳不同的手動控制 Promise，製造「舊 stream 的狀態查詢尚未 resolve 時切到另一 stream，新查詢先 resolve、舊查詢後 resolve」的亂序情境，斷言 Clear 按鈕的 enabled 狀態與 tooltip 對應目前所選 stream、被取代的舊查詢結果被捨棄。依 design.md「Decision 4」。驗證：新測試於實作前失敗。
- [x] 2b.2 修改 frontend/app/components/ServiceLogs.vue，新增 module-scoped `clearStatusSeq`（初始 0），於 `loadLogClearStatus()` 的 `GetLogClearStatus` await 前同步 `const seq = ++clearStatusSeq`；resolve 路徑的 `logClearStatus.value` 指派與 catch（silent-fail → `null`）路徑皆以 `seq === clearStatusSeq` 為前提，被取代的查詢靜默捨棄。計數器與 `loadSeq` 獨立。驗證：2b.1 測試轉綠。

## 3. 整體驗證

- [x] 3.1 執行 `make test` 與 `make lint`，全數通過且既有 ServiceLogs 測試（log-load-feedback 分流、ANSI 渲染、backend error passthrough、Clear Logs、log-auto-refresh 輪詢與 auto-disable）不回歸。驗證：兩命令 exit code 0。
- [ ] 3.2 請使用者以 `make dev` 啟動應用進行手動驗證：對一個持續寫 log 的 user service 勾選 Auto-refresh，於輪詢進行中快速切換 stdout/stderr，確認畫面內容始終對應目前所選 stream、不出現另一 stream 的殘影；並在 Auto-refresh 進行中對可清除的 log 執行 Clear Logs，確認成功後綠色 "Log cleared" 提示正常顯示、不被輪詢載入干擾。
