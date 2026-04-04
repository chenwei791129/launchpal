## 1. GUI domain helper（UID 格式化輔助函數）

- [x] 1.1 在 `internal/launchctl/user.go` 新增 `guiDomain()` 函數，使用 `gui/<uid>` 作為 domain target，回傳 `fmt.Sprintf("gui/%d", os.Getuid())`，實作 GUI domain helper 規格
- [x] 1.2 在 `internal/launchctl/user_test.go` 新增 `TestGuiDomain` 測試，驗證回傳格式為 `gui/<uid>`

## 2. 遷移 Start（bootstrap 語法：傳入 plist 路徑）

- [x] 2.1 修改 `internal/launchctl/user.go` 的 `Start()` 方法，將 `exec.Command("launchctl", "load", plistPath)` 替換為 `exec.Command("launchctl", "bootstrap", guiDomain(), plistPath)`，實作 Start user service via bootstrap 規格

## 3. 遷移 Stop（bootout 語法：使用 service-target（label））

- [x] 3.1 修改 `internal/launchctl/user.go` 的 `Stop()` 方法，將 `exec.Command("launchctl", "unload", plistPath)` 替換為 `exec.Command("launchctl", "bootout", guiDomain()+"/"+service.Label)`，實作 Stop user service via bootout 規格

## 4. 驗證

- [x] 4.1 執行 `go test ./internal/launchctl/` 確認所有既有測試通過
- [x] [P] 4.2 手動測試：透過 LaunchPal UI 啟動一個 user service，確認 `launchctl list` 能看到該服務
- [x] [P] 4.3 手動測試：透過 LaunchPal UI 停止該服務，確認 `launchctl list` 不再包含該服務
