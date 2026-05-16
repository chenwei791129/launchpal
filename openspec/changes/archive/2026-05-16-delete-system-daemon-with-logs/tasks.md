## 1. Protocol 定義（DeleteLogPaths RPC）

- [x] 1.1 在 `internal/privhelper/protocol.go` 新增 `MethodDeleteLogPaths = "DeleteLogPaths"` 常數，以及 `DeleteLogPathsRequest { Paths []string }` 與 `DeleteLogPathsResponse { Errors []string }` 型別（對應 DeleteLogPaths protocol coverage 需求）
- [x] 1.2 在 `internal/privhelper/handlers.go` 的 `AllMethods` slice 中新增 `MethodDeleteLogPaths`（對應 AllMethods coverage 需求）
- [x] 1.3 撰寫 `TestAllMethods_Coverage` 單元測試，確認 `MethodDeleteLogPaths` 已列入 `AllMethods`（TDD：先寫測試，再於 1.2 補實作使其通過；對應 DeleteLogPaths protocol coverage 需求）

## 2. Helper Handler 實作（TDD）

- [x] 2.1 在 `internal/privhelper/handlers_test.go` 撰寫 `handleDeleteLogPaths` 的單元測試，涵蓋以下場景：刪除單一 log 檔並清空 parent 目錄；parent 目錄有其他檔案時不刪 parent；路徑在 allowlist 外被拒絕；路徑為 symlink 被拒絕；log 檔不存在回傳 `ErrNotExist`；部分失敗的混合情境（對應 DeleteLogPaths RPC method 所有 Scenario）
- [x] 2.2 在 `internal/privhelper/handlers.go` 實作 `handleDeleteLogPaths`：以 `validateLogPath` 驗證每條路徑（路徑驗證策略（allowlist + lstat））— 確認在 allowlist 且至少一層子目錄深；用 `os.Lstat` 型別檢查拒絕 symlink（Symlink 防護）；呼叫 `os.Remove` 刪除檔案；依照 parent 目錄清除策略嘗試 `os.Remove(parent)` 並靜默忽略 `ENOTEMPTY`；依 DeleteLogPaths 部分失敗處理收集每條路徑的錯誤後一次回傳（對應 DeleteLogPaths RPC method 需求、DeleteLogPaths RPC 設計決策）
- [x] 2.3 在 `handleDeleteLogPaths` 中，將新 handler 注冊到 server dispatch map（switch/case 或 map[string]handler）

## 3. Client Wrapper

- [x] 3.1 [P] 在 `internal/privhelper/client.go` 新增 `DeleteLogPaths(paths []string) (warnings []string, err error)` client wrapper，序列化 `DeleteLogPathsRequest`、反序列化 `DeleteLogPathsResponse`、將 `Response.Errors` 作為 warnings 回傳（對應 DeleteLogPaths RPC method 需求）
- [x] 3.2 [P] 撰寫 client wrapper 的單元測試（使用 mock server 或 httptest-style transport），確認 request 正確序列化且 response 正確反序列化

## 4. DeleteServiceOptions 型別與 SystemManager 整合（TDD）

- [x] 4.1 在 `internal/launchctl/types.go`（或新增 `internal/launchctl/options.go`）定義 `DeleteServiceOptions struct { DeleteLogs bool }` 型別
- [x] 4.2 在 `internal/launchctl/system_test.go` 撰寫 `SystemManager.DeleteWithOptions` 的單元測試：`DeleteLogs: false` 行為與現有 `Delete` 一致；`DeleteLogs: true` 且 plist 有 `StandardOutPath`/`StandardErrorPath` 時呼叫 `DeleteLogPaths` RPC；plist 無 log 路徑時跳過 RPC；`DeleteLogPaths` 回傳 errors 時整體 delete 仍成功（回傳 warning）（對應 Optional log deletion on system daemon delete 所有 Scenario、Manager interface 相容性設計決策）
- [x] 4.3 在 `internal/launchctl/system.go` 實作 `SystemManager.DeleteWithOptions(name string, opts DeleteServiceOptions) error`（`Manager` interface 相容性：此方法不加入 interface，僅在 SystemManager 層暴露）：執行現有刪除流程（bootout + DeletePlist）；若 `DeleteLogs: true` 則從備份解析 plist 取得 `StandardOutPath`/`StandardErrorPath`；收集非空路徑後呼叫 `client.DeleteLogPaths`；依照 `DeleteLogPaths` 部分失敗處理原則，RPC 失敗時將 warnings 整合進回傳 error message 但整體仍為 success（對應 Optional log deletion on system daemon delete 需求）

- [x] 4.4 確認 `Manager` interface 的 `Delete(name string) error` 簽名不變（Read-only managers reject write operations 行為由現有 `SystemManager.Delete` 保持）；只有 `DeleteWithOptions` 是新增方法，不影響 interface 合規性；若現有 SystemManager write-reject 測試有影響則一併更新（對應 `launchdaemons-readonly` spec 中 Read-only managers reject write operations 需求）

## 5. App Binding 更新

- [x] 5.1 在 `app.go` 更新 `DeleteSystemService(name string, options DeleteServiceOptions) error`，接受 `DeleteServiceOptions` 並呼叫 `SystemManager.DeleteWithOptions`（對應 App binding for DeleteSystemService options 需求）
- [x] 5.2 撰寫 `TestDeleteSystemService_Options` 單元測試，確認 options 正確傳遞至 `SystemManager.DeleteWithOptions`（對應 App binding for DeleteSystemService options 需求 Scenario）

## 6. TypeScript 型別同步

- [x] 6.1 [P] 在 `frontend/app/types/wails.d.ts` 新增 `DeleteServiceOptions` interface（`{ deleteLogs: boolean }`），並更新 `DeleteSystemService(name: string, options: DeleteServiceOptions): Promise<void>` 函式簽名（對應 App binding for DeleteSystemService options 需求）

## 7. 前端 UI — Delete Dialog Checkbox

- [x] 7.1 在 `frontend/app/pages/system.vue` 的 `showDeleteDialog` 區塊新增一個 `ref<boolean>` 狀態 `deleteLogsChecked`，預設為 `false`（對應 Delete dialog log cleanup checkbox 需求、UI 預設行為設計決策）
- [x] 7.2 在 delete 確認對話框中加入 checkbox，label 為「Also delete log files」，輔助說明文字為「Log files will be permanently deleted and cannot be recovered.」（英文，符合 UI Language 規範；對應 Delete dialog log cleanup checkbox 需求所有 Scenario）
- [x] 7.3 修改 delete 確認觸發邏輯：呼叫 `DeleteSystemService(name, { deleteLogs: deleteLogsChecked.value })`，對話框關閉後重置 `deleteLogsChecked` 為 `false`（對應 User deletes service with log cleanup checked/unchecked Scenario）
- [x] 7.4 手動測試：啟用 Admin Mode → 確認 delete dialog 顯示 checkbox → 勾選後刪除服務 → 確認 log 目錄被清除；不勾選刪除 → 確認 log 目錄保留

## 8. 整合驗證

- [x] 8.1 執行 `make test` 確認所有 Go 單元測試通過，包含新增的 handler、client、SystemManager、App binding 測試
- [x] 8.2 更新 `.claude/CLAUDE.md` 的 Admin Mode 段落，新增 `DeleteLogPaths` 至 helper 能力清單
