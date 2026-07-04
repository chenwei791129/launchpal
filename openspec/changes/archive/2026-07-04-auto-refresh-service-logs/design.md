## Context

`ServiceLogs.vue` 的 `loadLogs()` 目前只由三個時機觸發：`onMounted`、`watch(logType)`、手動 Refresh 按鈕。前置 change `fix-log-error-classification` 完成後，`loadLogs` 會依 `LogsResult.status`（runtime 小寫鍵）分流渲染，並在 rejection 時顯示後端訊息。本 change 在該基礎上加入輪詢。

後端 `readLogTail` 每次最多讀取檔案尾端 1MB；對三種 service type 的讀取路徑皆無需特權。元件現已有計時器清理慣例（`clearSuccessTimeout` 於 `onBeforeUnmount` 清理）。

## Goals / Non-Goals

**Goals**

- Logs tab 提供 tail -f 式的自動刷新，開關明確、行為可預期。
- 不對結構性失敗（no-path / not-found）或持續錯誤做無意義輪詢。
- 不新增任何後端介面或跨層 seam。

**Non-Goals**

- 不做後端 push（fsnotify + Wails events）— 輪詢已滿足需求，push 需管理 watcher 生命週期（切 tab、切 stream、切 service 都要 rewatch），深度不成比例。
- 不做 offset 增量 tail — 整段重讀對 Clear Logs 的 truncate 天然免疫，且 1MB 上限的成本可忽略。
- 不將刷新間隔做成使用者設定，也不持久化開關狀態（YAGNI；先以固定 2 秒與 session 內狀態驗證需求）。
- 不做視窗失焦 / 分頁不可見時的暫停（列為未來優化）。
- 不因檔案稍後出現而自動恢復輪詢 — 非 ok 結果自動關閉後，由使用者手動 Refresh 或重新勾選。

## Decisions

### Decision 1: 前端 setInterval 輪詢，固定 2 秒

啟用 Auto-refresh 時以 `setInterval` 每 2000ms 呼叫既有 `loadLogs()`；停用、切換元件、unmount 時 `clearInterval`。為什麼不用遞迴 `setTimeout`：`loadLogs` 本身有 `loading` 旗標可作 in-flight 防護（見 Decision 2），固定週期的 interval 較簡單且行為可預期；為什麼是 2 秒：對 1MB tail 的 IPC 與 `ansiToHtml` 重算成本可忽略，又足夠接近即時。

### Decision 2: in-flight tick 跳過

interval callback 先檢查 `loading.value`，為 true 時直接 return（跳過該次 tick），不排隊、不併發。這避免慢速讀取（例如接近 1MB 的大檔）時請求堆疊，也讓手動 Refresh 與輪詢共用同一套載入狀態。

### Decision 3: 非 ok 結果自動關閉開關

任一次載入（不論由輪詢、手動 Refresh、或切換 logType 觸發）結束後，若結果為 `status !== "ok"`（讀取 runtime 小寫鍵）、promise rejection、或走到「無 Wails bindings」的 development fallback（無 `LogsResult` 可判斷，輪詢同樣無意義），且 Auto-refresh 當時為啟用狀態，則自動將開關設為關閉並停止輪詢。判斷邏輯放在所有觸發來源共用的載入後路徑（`loadLogs` 的收尾），不得只掛在輪詢 callback 上。選擇「關閉開關」而非「保持勾選但暫停」：checkbox 狀態即輪詢狀態，所見即所得，可直接以 UI 斷言測試；「勾選但暫停」需要額外的暫停指示 UI，複雜度不值得。副作用：log 檔稍後才出現的服務不會自動恢復跟隨（已列入 Non-Goals）。

### Decision 4: 開關與 Auto-scroll 正交

Auto-refresh 只控制「何時重新載入」；Auto-scroll 維持既有語意「載入後是否捲到底」。兩者獨立勾選：跟隨模式 = 兩者皆開；「內容更新但停在原處看歷史」= 只開 Auto-refresh。不合併為單一 Follow 開關，保留後者場景。

### Decision 5: 切換 logType 時保留開關狀態

`watch(logType)` 既有的 `loadLogs()` 觸發不變；Auto-refresh 若為啟用狀態則對新 stream 繼續輪詢。若新 stream 首次載入即非 ok，依 Decision 3 自動關閉。

切換 **service**（`serviceName` prop 變更）則相反 — 開關重置為關閉並停止輪詢：detail page（`pages/services/[name].vue`）在服務間導航時重用同一個 route component 且 `ServiceLogs` 無 `:key`，元件不會 remount，`onBeforeUnmount` 不會觸發，因此需要 `watch(() => props.serviceName)` 主動清理；且 Auto-refresh 是使用者對「當時那個服務」的選擇，不應靜默延續到另一個服務。

## Implementation Contract

**可觀察行為**

1. Logs tab 控制列出現 "Auto-refresh" checkbox（英文文案），預設未勾選，位於既有 Auto-scroll 附近且兩者可獨立勾選。
2. 勾選後，日誌內容每 2 秒自動更新（新寫入的行出現在畫面上，無需手動 Refresh）；取消勾選後停止更新。
3. 勾選期間離開頁面（元件 unmount）後不再有任何輪詢請求發出；在服務間導航（`serviceName` prop 變更、元件未 remount）時開關自動變為未勾選且輪詢停止。
4. 一次載入結果為 no-path / not-found / rejection / development fallback（無 bindings）時，checkbox 自動變為未勾選且輪詢停止；畫面顯示對應的 placeholder 或錯誤（沿用 log-load-feedback 的分流）。此關閉行為對輪詢、手動 Refresh、切換 stream 三種觸發來源一致。
5. 輪詢期間手動 Refresh 與切換 stdout/stderr 照常可用；in-flight 時的 tick 不會造成重複請求。
6. Auto-refresh + Auto-scroll 同開時，每次自動刷新後視圖捲到底；Auto-scroll 關閉時捲動位置不被刷新打斷。

**介面形狀**

- 純 `ServiceLogs.vue` 內部狀態：`autoRefresh: Ref<boolean>` + interval handle；無新 props、無新 emits、無新 bindings。

**驗收目標**

- `frontend/app/components/__tests__/ServiceLogs.test.ts` 以 vitest fake timers 覆蓋：勾選後 tick 觸發重載、取消勾選停止、unmount 清理 interval、`serviceName` 變更關閉開關並停止輪詢、in-flight tick 跳過、非 ok 結果（no-path / not-found / rejection / dev fallback）自動關閉、手動 Refresh 與切 stream 觸發的關閉、Auto-scroll 正交與 follow mode 捲底。使用 fake timers 的測試必須在 teardown 呼叫 `vi.useRealTimers()`（既有 `afterEach` 只有 `vi.restoreAllMocks()`，fake timers 會外洩至其他測試 — 例如 `confirmClear` 的真實 `setTimeout(2000)`）。
- `make test` 與 `make lint` 全綠。

**範圍邊界**

- In scope：`ServiceLogs.vue` 的開關 UI、輪詢生命週期、自動關閉邏輯、對應測試。
- Out of scope：後端與 bindings、設定頁、間隔可調、狀態持久化、visibility 暫停、自動恢復輪詢。

## Risks / Trade-offs

- [1MB tail 每 2 秒重算 `ansiToHtml` + `v-html` 重繪造成卡頓] → 內容字串未變時跳過重繪的優化留待實測；`renderedLogs` 是 computed，字串相同時 Vue 的 diff 成本已有限。若實測卡頓，追加「內容不變即跳過賦值」的 short-circuit。
- [自動關閉在瞬時錯誤（如單次 I/O 失敗）時過於激進] → 接受：手動重新勾選即可恢復；比「連續 N 次失敗才停」的計數狀態機簡單且可預期。
- [測試中 fake timers 與 async loadLogs 的交錯] → 使用 vitest 的 `vi.useFakeTimers` + `flushPromises` 慣例並於 teardown `vi.useRealTimers()`；既有測試檔已有 async 元件測試先例。
- [輪詢只刷新日誌，不刷新 `logClearStatus`，Clear 按鈕狀態在檔案於輪詢期間被外部輪替/刪除時會過時] → 接受：Clear 操作本身有後端驗證兜底（不存在的檔案回錯誤），低嚴重度；輪詢加掛 status 查詢會使每 tick 的 IPC 加倍，不值得。

## Migration Plan

單一前端元件變更，一次 commit 完成，回滾即 revert。**前置依賴：`fix-log-error-classification` 須先實作合入**（自動關閉條件讀取 `LogsResult.status`）；若順序顛倒，本 change 的非 ok 判斷無從實作。

## Open Questions

（無 — 間隔、預設值、自動關閉語意已在本設計中定案。）
