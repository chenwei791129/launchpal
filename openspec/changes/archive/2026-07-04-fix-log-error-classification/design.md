## Context

Logs tab 的載入鏈為：`ServiceLogs.vue` 的 `loadLogs` → Wails binding（`App.GetLogs` / `App.GetSystemLogs`）→ `Manager.GetLogs`（`UserManager` / `SystemManager` / `AppleSystemManager`，後兩者共用 `readOnlyManager.getLogs`）→ `resolveLogPath` + `readLogTail`。

現況有兩個缺陷：

1. Wails v2 將 Go error 以字串 reject Promise，前端 `e instanceof Error ? e.message : 'Failed to load logs'` 永遠走 fallback，後端錯誤訊息全數遺失。
2. 「未配置 log path」與「log 檔不存在」目前以 error 回傳（`resolveLogPath` 與 `readLogTail`），前端因此把服務的正常狀態渲染成紅字錯誤。apple-system domain 尤其嚴重：實測 411 個服務僅 3–6 個配置了 log path，配置者又多指向 `/dev/console` / `/dev/null` / 不存在的檔案，導致該頁面幾乎所有服務都顯示「Failed to load logs」。

限制條件：Wails v2 的 error 通道只能傳字串，無法攜帶結構化欄位；因此「狀態分類」必須放在成功回傳值裡，error 通道只留給真正的失敗。

## Goals / Non-Goals

**Goals**

- 結構性狀態（no-path / not-found）以結構化回傳值表達，前端渲染為帶原因的 placeholder。
- 真正的失敗（權限不足、I/O 錯誤、無效 logType、服務不存在）維持 error 通道，且後端訊息原文透傳到前端紅字錯誤分支。
- 三種 service type（user / system / apple-system）行為一致。

**Non-Goals**

- 不做 Logs 自動刷新（tail -f）— 為後續獨立 change，該 change 將依賴本次的狀態分類決定輪詢停止條件。
- 不改 `GetLogClearStatus` / `ClearLogs` 的介面與行為。
- 不引入 macOS 統一日誌（`log show` / os_log）作為 apple-system 服務的替代 log 來源。
- 不做 i18n；placeholder 文案為英文硬字串。

## Decisions

### Decision 1: 狀態分類走結構化回傳 LogsResult

（而非前端比對錯誤字串。）

`Manager.GetLogs` 回傳型別改為 `(LogsResult, error)`：

```go
// internal/launchctl/types.go
type LogsResult struct {
    Content string `json:"content"` // log tail content; meaningful only when Status == "ok"
    Status  string `json:"status"`  // "ok" | "no-path" | "not-found"
    Path    string `json:"path"`    // resolved log path; empty when Status == "no-path"
}
```

- 為什麼不用「前端比對錯誤字串」：Wails 只透傳字串，比對 `no stdout log path configured` 之類的訊息會把 UI 分流耦合在錯誤文案上，後端改寫文案即靜默破壞分流。
- 為什麼不用「復用 `GetLogClearStatus`」（它已有 `logPath` / `exists` 欄位）：`loadLogs` 與 `loadLogClearStatus` 是兩個平行請求，存在讀取間隙的競態；且 apple-system 目前刻意跳過該查詢以維持頁面效能，復用會把效能取捨與錯誤分類綁死。
- 為什麼不用 Go sentinel error + binding 層轉換：分類邏輯會被拆散在 manager 與 binding 兩層；直接在 manager 層回傳 `LogsResult` 讓三個實作共用同一份分類語意。

### Decision 2: error 通道僅保留真正的失敗

- `no-path`：`selectLogPath` 為空 → `LogsResult{Status: "no-path"}`，不再走 `resolveLogPath` 的 error。
- `not-found`：開檔得到 `os.IsNotExist` → `LogsResult{Status: "not-found", Path: <resolved path>}`。
- 權限不足（`os.IsPermission`）、其他 I/O 錯誤（含路徑指向目錄的讀取失敗）、無效 logType、`Get(name)` 找不到服務 → 維持 error 回傳，訊息措辭沿用現有格式（`permission denied reading log file: <path>` 等）；permission-denied 訊息在三個 domain 都包含解析後路徑（由共用的 `readLogTail` 產生）。
- 空檔案：`Status: "ok"`、`Content: ""` — 前端沿用既有「No logs available」placeholder，與 Clear Logs 後的重載行為相容。
- `Path` 欄位語意為「實際用於開檔的路徑」：user domain 經 `expandTilde` 展開（維持現狀），system / apple-system domain 沿用 plist 中的字面路徑不做展開（維持 `readonly.go` 現狀 — 系統 daemon 以 root 執行，`~` 在該情境本無意義）。分類語意（`ok` / `no-path` / `not-found` 的判斷規則）三個 domain 一致，路徑展開行為維持各自現狀。

### Decision 3: 前端以 Status 分流與 rejection 原文透傳

`loadLogs` 的錯誤處理修正為同時接住字串與 Error 物件：

```ts
error.value = typeof e === 'string' ? e : e instanceof Error ? e.message : 'Failed to load logs'
```

placeholder 分流（皆為非紅字的 placeholder 樣式，文案為英文）：

| Status      | 呈現                                                              |
| ----------- | ----------------------------------------------------------------- |
| `ok`（空）  | 既有 "No logs available for {logType}"                            |
| `no-path`   | "No {logType} log path configured for this service"               |
| `not-found` | "Log file does not exist yet" 並以次要文字顯示 `Path`             |
| rejection   | 既有紅字錯誤分支，內容為後端訊息原文                              |
| 無 bindings | 既有 "No logs available for {logType}"（development fallback）    |

注意：Wails 以 json tags 序列化，前端收到的 runtime 物件鍵名為小寫 `content` / `status` / `path`（`wails.d.ts` 與產生的 `models.ts` 同此形狀）；設計文件與 spec 中的 `Content` / `Status` / `Path` 為 Go 欄位名。

### Decision 4: readLogTail 的 not-found 以檔案開啟結果辨識

`readLogTail` 目前把 `os.IsNotExist` 包成 `log file not found: <path>` 錯誤字串。改為讓 `GetLogs` 各實作能以 `os.IsNotExist`（或等價的 sentinel）辨識並轉為 `Status: "not-found"`，避免以字串比對錯誤訊息分類。具體形式（sentinel error 或 `readLogTail` 拆出開檔步驟）由實作決定，但分類判斷不得依賴錯誤訊息文字。

## Implementation Contract

**可觀察行為**

1. user / system / apple-system 任一 domain 的服務，若請求的 stream 未配置 log path：Logs tab 顯示 placeholder「No {logType} log path configured for this service」，無紅字錯誤、無 console error。
2. 已配置 log path 但檔案不存在：placeholder「Log file does not exist yet」+ 路徑，無紅字錯誤。
3. log 檔存在但無讀取權限：紅字錯誤分支，內容包含後端訊息 `permission denied reading log file: <path>` 原文。
4. log 檔存在且可讀：內容照常渲染（含 ANSI 處理）；空檔案顯示既有「No logs available for {logType}」。
5. Refresh 按鈕與 stdout/stderr 切換在上述所有狀態間往返時，分流正確更新，不殘留前一狀態的錯誤或 placeholder。

**介面形狀**

- Go：`Manager.GetLogs(name, logType string) (LogsResult, error)`；`LogsResult{Content, Status, Path}`，`Status` 常數 `"ok"` / `"no-path"` / `"not-found"`。簽名變更點含 `manager.go` 介面、`user.go` / `readonly.go` 實作、`system.go` / `apple_system.go` 的 wrapper 方法（各自的 `var _ Manager` 斷言會在編譯期揭露遺漏）。
- TS：`wails.d.ts` 新增 `LogsResult` 介面（鍵名小寫 `content` / `status` / `path`）；`GetLogs` / `GetSystemLogs` 回傳 `Promise<LogsResult>`。checked-in 的產生檔 `frontend/wailsjs/go/main/App.d.ts`、`App.js`、`frontend/wailsjs/go/models.ts` 以 `wails generate module` 重新產生（或手動同步至相同形狀），不得留下 `Promise<string>` 舊簽名。

**驗收目標**

- Go 測試：`internal/launchctl/user_test.go` 覆蓋 user domain 的四種狀態（ok / no-path / not-found / permission-denied error）；system domain 針對 `readOnlyManager.getLogs` 覆蓋相同分類，既有 `TestSystemManager_GetLogs`（system_test.go）與 `TestAppleSystemManager_GetLogs`（apple_system_test.go）的字串斷言更新為 `LogsResult` 斷言。permission-denied 測試案例在 `os.Geteuid() == 0` 時 skip（root 不受 mode bits 限制，無法製造該情境）。
- 前端測試：`frontend/app/components/__tests__/ServiceLogs.test.ts` 覆蓋 Status 三分流的 placeholder 文案、字串 rejection 訊息透傳、以及空內容 placeholder 不回歸。
- `make test`（Go 測試 + vitest + typecheck）全綠。

**範圍邊界**

- In scope：上述 Go / TS 介面、三個 manager 實作（含 `system.go` / `apple_system.go` wrapper）、兩個 Wails bindings 與 checked-in 產生檔（`frontend/wailsjs/**`）、`ServiceLogs.vue` 分流與文案、對應測試、`ansi-log-rendering` 與 `core-service-management`「Clear service logs」既有 spec 措辭的同步更新。
- Out of scope：自動刷新、`ClearLogs` / `GetLogClearStatus` 介面、macOS 統一日誌整合、i18n。

## Risks / Trade-offs

- [`GetLogs` 簽名變更漣漪] 所有呼叫端（bindings、測試）需同步更新 → 以編譯錯誤驅動修改，單 repo 內一次完成；grep 確認無其他呼叫端。
- [前端 `logs` ref 語意改變（由 string|null 變為內容 + 狀態）] 模板分支條件需重排，可能影響既有 ANSI 渲染與 Clear Logs 重載 → 以既有 `ServiceLogs.test.ts` 全量回歸 + 新增分流測試護住。
- [`ansi-log-rendering` spec 的「Existing error state is unaffected」場景措辭] 本次不改該 spec 的渲染行為；error 分支的觸發條件收窄（僅真失敗），但分支本身的渲染方式不變，不構成該 capability 的 requirement 變更。

## Migration Plan

單 repo 前後端同步修改，一次 PR 完成，無資料遷移。實作順序：Go 型別與 manager → bindings → TS 型別 → 前端分流 → 測試。回滾即 revert 單一 commit。

## Open Questions

（無 — 方案已在 discuss 階段收斂，無待決事項。）
