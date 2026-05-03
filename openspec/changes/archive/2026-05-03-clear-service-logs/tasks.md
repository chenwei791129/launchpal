## 1. Protocol 與型別擴充

- [x] 1.1 [P] 在 `internal/privhelper/protocol.go` 新增 `MethodTruncateLog` 常數、`TruncateLogParams` 結構與 README/godoc 說明，未涉及行為（呼應 design「新增 helper RPC `TruncateLog` 而非重用 WritePlist」）
- [x] 1.2 [P] 在 `internal/launchctl/types.go` 新增 `LogClearStatus` 結構（`LogPath`、`Exists`、`UserWritable` 三欄位），含 JSON tag（呼應 design「`LogClearStatus` 結構而非單純 bool」）
- [x] 1.3 [P] 在 `internal/launchctl/manager.go` 的 `Manager` interface 新增 `ClearLogs(name, logType string) error` 與 `GetLogClearStatus(name, logType string) (LogClearStatus, error)` 方法簽章，使所有實作器需補齊

## 2. privhelper：TruncateLog handler（TDD）

- [x] 2.1 在 `internal/privhelper/handlers_test.go` 為 TruncateLog RPC method 新增單元測試：白名單命中、白名單未命中、parent 等於 allowlist root、檔案不存在、symlink 拒絕、root-owned 檔案截斷後 mode/owner 不變（呼應 spec Requirement「TruncateLog RPC method」）
- [x] 2.2 在 `internal/privhelper/handlers.go` 實作 `truncateLog` handler，重用 `validateLogPath`、`syscallNoFollow`、`O_WRONLY|O_TRUNC`，並把 dispatch 加進 `Handle` switch（呼應 spec Requirement「Supported RPC methods」與 design「新增 helper RPC `TruncateLog` 而非重用 WritePlist」）
- [x] 2.3 [P] 在 `internal/privhelper/client.go` 新增 `TruncateLog(ctx context.Context, path string) error` typed wrapper，沿用既有的 unmarshal / error mapping 慣例
- [x] 2.4 [P] 在 `internal/privhelper/protocol_test.go` / `client_test.go` 補上 round-trip 序列化測試確保 TruncateLog params/response 可正確編解碼

## 3. launchctl 共用：log 權限查詢（TDD）

- [x] 3.1 在 `internal/launchctl/types.go` 或鄰近檔新增共用 helper：`canWriteLogFile(path string) bool`（透過 `os.OpenFile(path, O_WRONLY|O_NOFOLLOW, 0)` 試開立即關閉）與對應單元測試，覆蓋可寫、唯讀、不存在、symlink 四案例
- [x] 3.2 確保所有 log 操作的 `~` 展開行為一致（呼應 spec Requirement「Read service logs」MODIFIED 中的 Tilde expansion shared across log operations）

## 4. UserManager 實作（TDD）

- [x] 4.1 在 `internal/launchctl/user_test.go` 為 `UserManager.ClearLogs` 寫測試：成功 truncate、無 logType、無設定 path、檔案不存在、symlink 拒絕（呼應 spec Requirement「Clear service logs」）
- [x] 4.2 在 `internal/launchctl/user.go` 實作 `UserManager.ClearLogs`，使用 `O_WRONLY|O_TRUNC|O_NOFOLLOW`，不刪除檔案、不更動 mode/owner（呼應 design「截斷 log 檔案而非僅清空畫面顯示」）
- [x] 4.3 在 `internal/launchctl/user_test.go` 為 `UserManager.GetLogClearStatus` 寫測試：path 存在且可寫、path 存在但 read-only、path 設定但檔案不存在、未設 path、service 不存在（呼應 spec Requirement「Query log clear authorization status」）
- [x] 4.4 在 `internal/launchctl/user.go` 實作 `UserManager.GetLogClearStatus`，整理出 helper（與 SystemManager 共用）以避免重複程式碼

## 5. SystemManager 與 readonly 實作（TDD）

- [x] 5.1 在 `internal/launchctl/system_test.go` 為 `SystemManager.ClearLogs` 寫測試覆蓋 dispatch 路徑：使用者可寫直接 truncate（不呼叫 helper）、EACCES 且 Admin Mode 啟用走 helper、EACCES 且 Admin Mode 關閉回 ErrReadOnlyManager、ELOOP 不 escalate（呼應 spec Requirement「Truncate system daemon log with permission dispatch」與 design「system 服務 per-file 寫入權限分流」）
- [x] 5.2 在 `internal/launchctl/system.go` 實作 `SystemManager.ClearLogs`：先嘗試直接 OpenFile，根據 errno 決定是否 fallback 至 `m.client().TruncateLog(...)`，並在 fallback 前重新呼叫 `m.client()` 以避免 stale client（呼應 design「後端在實際 truncate 時必再驗權限」與 spec Requirement「Re-validate authorization at clear time」）
- [x] 5.3 在 `internal/launchctl/readonly.go` 為 `readOnlyManager` 補上共用 `getLogClearStatus`（system 與 apple-system 共用），並讓 `apple_system.go` 暴露對應方法
- [x] 5.4 在 `internal/launchctl/apple_system.go` 加上 `ClearLogs` stub，回傳明確錯誤（`ErrReadOnlyManager` 或文字含 "apple-system" / "read-only"），對應測試在 `internal/launchctl/apple_system_test.go`（呼應 spec Requirement「Truncate system daemon log with permission dispatch」中的 apple-system 拒絕情境）

## 6. Wails layer

- [x] 6.1 [P] 在 `app.go` 新增 `ClearLogs(name, logType)`、`ClearSystemLogs(name, serviceType, logType)`、`GetLogClearStatus(name, serviceType, logType)` 三支 binding，apple-system 直接回 error，system 走 systemManager（呼應 design「新增三支 Wails binding 而非擴增現有 GetLogs / GetSystemLogs」）
- [x] 6.2 [P] 在 `app_test.go`（若存在）或新增測試覆蓋 dispatch：user / system / apple-system 三條路徑與 invalid serviceType 的錯誤訊息

## 7. Frontend：型別與 binding

- [x] 7.1 在 `frontend/app/types/wails.d.ts` 新增 `LogClearStatus` 介面與三支 binding 的 type signature
- [x] 7.2 確認 `frontend/wailsjs/go/main/App.d.ts` 由 wails generate 重新產出（在 dev 流程裡執行 `wails generate module` 或 `make dev` 即可），並 commit

## 8. Frontend：ServiceLogs.vue 控制與狀態

- [x] 8.1 在 `frontend/app/components/ServiceLogs.vue` 新增 reactive ref `logClearStatus`，於 mount 與 `logType` watch 內呼叫 `GetLogClearStatus`（user 服務）或 `GetLogClearStatus` 經 system binding 取狀態，並處理 in-flight loading 旗標（呼應 spec Requirement「Status query is lazy and bounded」）
- [x] 8.2 在 `ServiceLogs.vue` 新增 prop `serviceType`、`canWrite`（或 `adminEnabled`），由 `frontend/app/pages/services/[name].vue` 傳入，避免元件直接 import composable
- [x] 8.3 在 `ServiceLogs.vue` 新增 computed `clearControlState` 計算 `(visible, enabled, tooltipReason)`，完整對應 spec Requirement「Button availability matrix」中的 5 種 state（含 No log path configured / Log file does not exist / Enable Admin Mode to clear 三種 tooltip）（呼應 design「按鈕可見性 / 可用性矩陣」）
- [x] 8.4 在 `ServiceLogs.vue` controls row 新增 Clear Logs 按鈕（圖示風格與 Refresh 一致），綁定 `clearControlState`（呼應 spec Requirement「Clear Logs button in ServiceLogs view」）

## 9. Frontend：確認 modal 與動作

- [x] 9.1 在 `ServiceLogs.vue` 新增 Teleport-based 確認 dialog，沿用 `[name].vue:320-348` 的 surface 樣式但使用紅色按鈕；標題 `Clear Logs`、文案包含目前 `logType` 與 service name 並警示 0 byte truncation 不可恢復（呼應 design「確認 modal 沿用 Run Now 風格」與 spec Requirement「Confirmation dialog before clearing」）
- [x] 9.2 實作 confirm handler：呼叫對應 binding、處理錯誤（包含 helper 中斷、apple-system 誤呼叫等）、成功時 reload `loadLogs()` 並顯示 1 秒以上的 transient success 指示（呼應 spec Requirement「Successful clear shows feedback and reloads logs」與「Clear failure surfaces a recoverable error」）
- [x] 9.3 [P] 在 `frontend/app/components/__tests__/` 新增 `ServiceLogs.test.ts`，至少覆蓋：apple-system 隱藏按鈕、各種 disabled tooltip、confirm 取消、confirm 後呼叫正確 binding

## 10. 文件與 issue 同步

- [x] 10.1 [P] 更新 `.claude/CLAUDE.md`：在 Service Types 區塊註記 user / system 的 Clear Logs 行為與權限分流；在 Admin Mode 區塊補上 `TruncateLog` RPC
- [x] 10.2 [P] 在 `README.md` Features 列表新增「Clear stdout / stderr logs from the Logs tab」
- [x] 10.3 在 GitHub issue #18 留言摘要實作內容（提供按鈕位置、權限分流、apple-system 限制、不可恢復警告），但**不關閉 issue**，依專案 release 流程處理

## 11. 驗證

- [x] 11.1 跑 `make test`（Go + frontend vitest + TS typecheck）並修掉所有 failure
- [x] 11.2 跑 `make lint`（golangci-lint + eslint）並修掉所有 warning
- [x] 11.3 手動驗證情境：user 服務 stdout/stderr 各一次清除、Homebrew daemon 不啟 Admin Mode 即可清、root-owned daemon 必須啟 Admin Mode、apple-system 完全看不到按鈕、Admin Mode idle 過期時的錯誤訊息
