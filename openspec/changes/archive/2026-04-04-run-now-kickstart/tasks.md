## 1. 後端實作

- [x] [P] 1.1 在 `internal/launchctl/user.go` 新增 `Kickstart(name string) error` 方法，實作「Kickstart a user service」需求：檢查 plist 存在、取得 UID (`os.Getuid()`)、判斷服務是否已 loaded（若未 loaded 先執行 `launchctl bootstrap gui/{UID} {plistPath}`）、執行 `launchctl kickstart -k gui/{UID}/{label}`
- [x] [P] 1.2 在 `app.go` 新增 `KickstartService(name string) error` 方法作為 Wails binding，實作「Kickstart backend binding」需求，委派至 `UserManager.Kickstart`

## 2. 前端實作

- [x] 2.1 在 `frontend/app/pages/services/[name].vue` 的 header 按鈕列新增「Run Now」按鈕，實作「Run Now button in service detail page」需求：僅對非 readOnly 的 user service 顯示
- [x] 2.2 在 `frontend/app/pages/services/[name].vue` 實作「Confirmation dialog when service is running」需求：當 service status 為 running 時跳出確認對話框提醒現有 process 會被終止，非 running 狀態直接執行；確認後呼叫 `KickstartService` 並刷新狀態

## 3. 測試

- [x] 3.1 在 `internal/launchctl/user_test.go` 新增 Kickstart 方法的單元測試，涵蓋 plist 不存在的錯誤場景
