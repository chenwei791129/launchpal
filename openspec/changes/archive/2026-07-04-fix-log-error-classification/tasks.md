## 1. Go 後端 — LogsResult 型別與狀態分類（TDD）

- [x] [P] 1.1 在 internal/launchctl/user_test.go 新增 / 改寫 UserManager.GetLogs 測試（先寫、先失敗）：依 design.md「Decision 1: 狀態分類走結構化回傳 LogsResult」與「Decision 2: error 通道僅保留真正的失敗」，斷言 "Read service logs" requirement 的四種結果 — 未配置路徑回傳 `LogsResult{Status: "no-path"}` 且 error 為 nil、路徑存在但檔案不存在回傳 `Status: "not-found"` 且 `Path` 為解析後路徑、空檔案回傳 `Status: "ok"` 且 `Content` 為空、mode 000 檔案回傳含 "permission denied" 與解析後路徑的 error（此案例於 `os.Geteuid() == 0` 時 skip，root 不受 mode bits 限制）、路徑指向目錄回傳 error（非 Status）。同時更新 "Clear service logs" requirement 的既有測試斷言：ClearLogs 後的 `GetLogs` 回傳 `Status: "ok"` 且 `Content` 為空（原斷言為空字串）。驗證：`go test ./internal/launchctl/ -run GetLogs` 於實作前失敗（編譯錯誤或斷言失敗）。
- [x] [P] 1.2 更新 internal/launchctl/system_test.go 的 `TestSystemManager_GetLogs` 與 internal/launchctl/apple_system_test.go 的 `TestAppleSystemManager_GetLogs`（先寫、先失敗）：字串斷言改為 `LogsResult` 斷言，並補齊 "Read system service logs" requirement 的 no-path / not-found / ok / permission-denied 分類案例 — 狀態分類規則與 user domain 一致，但 `Path` 為 plist 字面路徑（system domain 不做 tilde 展開）。驗證：`go test ./internal/launchctl/` 於實作前失敗。
- [x] 1.3 在 internal/launchctl/types.go 新增 `LogsResult` 結構（`Content` / `Status` / `Path`，含 json tags 與 `"ok"` / `"no-path"` / `"not-found"` 狀態常數），並依 design.md「Decision 4: readLogTail 的 not-found 以檔案開啟結果辨識」調整辨識形式（sentinel error 或拆出開檔步驟），使呼叫端以 `os.IsNotExist` 等值判斷分類、不比對錯誤訊息文字。驗證：`go vet ./internal/launchctl/` 通過，types_test.go 既有測試不回歸。
- [x] 1.4 修改 internal/launchctl/manager.go 的 `Manager.GetLogs` 簽名為 `(LogsResult, error)`，並同步實作 internal/launchctl/user.go 的 `UserManager.GetLogs`、internal/launchctl/readonly.go 的 `readOnlyManager.getLogs`（SystemManager 與 AppleSystemManager 共用），以及 internal/launchctl/system.go 與 internal/launchctl/apple_system.go 的 `GetLogs` wrapper 方法（兩檔各自的 `var _ Manager` 斷言須通過編譯）：結構性狀態走 `LogsResult.Status`，無效 logType / 服務不存在 / 權限不足 / 其他 I/O 錯誤維持 error 通道且措辭沿用現有格式。驗證：1.1 與 1.2 的測試轉綠，`go test ./internal/launchctl/` 全數通過。

## 2. Wails bindings 與前端型別

- [x] [P] 2.1 更新 app.go 的 `GetLogs` / `GetSystemLogs` bindings 回傳型別為 `(launchctl.LogsResult, error)`，行為為透傳 manager 回傳值。驗證：`go build ./...` 編譯通過，全 repo grep 確認無其他 `GetLogs` 呼叫端遺漏。
- [x] [P] 2.2 更新 frontend/app/types/wails.d.ts：新增 `LogsResult` 介面（小寫鍵 `content` / `status` / `path`），`GetLogs` / `GetSystemLogs` 回傳 `Promise<LogsResult>`；並同步 checked-in 的產生檔 frontend/wailsjs/go/main/App.d.ts、frontend/wailsjs/go/main/App.js、frontend/wailsjs/go/models.ts（以 `wails generate module` 重新產生，或手動更新至相同形狀），不得殘留 `Promise<string>` 舊簽名。驗證：`make test` 中的 TypeScript typecheck 通過（此時 ServiceLogs.vue 尚未改完會先報型別錯，於 3.x 完成後回綠），且 grep frontend/wailsjs 無 `GetLogs(...):Promise<string>`。

## 3. 前端 ServiceLogs 分流（TDD）

- [x] 3.1 在 frontend/app/components/__tests__/ServiceLogs.test.ts 新增測試（先寫、先失敗）：依 "Classify log load results in the Logs tab" requirement 斷言 — `no-path` 顯示 "No stdout log path configured for this service"、`not-found` 顯示 "Log file does not exist yet" 與 path 次要文字、`ok` 空內容顯示既有 "No logs available"、`ok` 非空內容經 ANSI pipeline 渲染、無 Wails bindings（development fallback）顯示既有 "No logs available" placeholder，且上述皆不渲染紅字錯誤分支。驗證：新測試於實作前失敗。
- [x] 3.2 在同一測試檔新增 "Surface backend error messages verbatim on load failure" requirement 的測試（先寫、先失敗）：字串 rejection 原文顯示、Error rejection 顯示 message、非字串非 Error 顯示 "Failed to load logs"、切換 logType 後前一狀態的錯誤 / placeholder 不殘留。驗證：新測試於實作前失敗。
- [x] 3.3 依 design.md「Decision 3: 前端以 Status 分流與 rejection 原文透傳」修改 frontend/app/components/ServiceLogs.vue：`loadLogs` 改為處理 `LogsResult`（讀取小寫鍵 `content` / `status` / `path`），模板依 Status 分流 placeholder（英文文案），錯誤處理改為 `typeof e === 'string'` 優先接住 Wails 字串 rejection；`renderedLogs` computed 依修改後的 "Mount rendered output in ServiceLogs view" requirement 改以 LogsResult 的 content 為輸入，`<pre>` 的既有 CSS classes 不變；Clear Logs 後的重載沿用同一分流。驗證：3.1 與 3.2 測試轉綠，既有 ServiceLogs ANSI 渲染測試不回歸。

## 4. 整體驗證

- [x] 4.1 執行 `make test`（Go 測試 + vitest + TypeScript typecheck）與 `make lint`，全數通過且無既有測試回歸。驗證：兩命令 exit code 0。
- [x] 4.2 請使用者以 `make dev` 啟動應用進行手動驗證：於 apple-system 頁面確認未配置 log path 的服務顯示 "No stdout log path configured for this service" placeholder（非紅字錯誤）；於 user 服務確認正常 log 仍照常渲染；製造一個不可讀的 log 檔確認紅字錯誤顯示後端 "permission denied reading log file" 原文訊息。
