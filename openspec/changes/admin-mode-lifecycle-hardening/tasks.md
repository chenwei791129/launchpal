<!-- 各任務描述陳述「交付的行為/契約」與「驗證方式」；檔案路徑僅為定位脈絡。技術符號、路徑、指令保留原文。 -->

## 1. Helper self-terminates on client disconnect（斷線即自我終結，涵蓋所有連線結束路徑；Stop() 關閉現有連線）

- [x] [P] 1.1 在 `internal/privhelper/server_test.go` 新增測試涵蓋 spec「Helper self-terminates on client disconnect」：分別模擬 client 連線以 EOF、讀取錯誤、以及寫入（`encoder.Encode`）錯誤結束，斷言伺服器 `Stop()`、移除 socket 檔、`acceptLoop` 結束、後續連線嘗試失敗。驗證：`go test ./internal/privhelper -run` 對應測試綠燈。
- [x] 1.2 於 `internal/privhelper/server.go` 的 `acceptLoop` 在 `handleConn(conn)` 返回後（若 `!isStopped()`）觸發 `Stop()`，使所有連線結束返回路徑（EOF／讀取錯誤／寫入錯誤）都會拆除（設計：斷線即自我終結涵蓋所有連線結束路徑）。行為：Disable/GUI 結束/崩潰/斷線後 helper 於數秒內消失。驗證：1.1 測試通過。
- [x] 1.3 於 `internal/privhelper/server.go` 讓 Stop() 關閉現有已接受連線，使 idle/父程序觸發的停止能解除 `handleConn` 於 `scanner.Scan()` 的阻塞並終結程序；在 `server_test.go` 補測「GUI 仍連線時 `Stop()` 會使 `handleConn` 返回、`connWG` 完成」。驗證：對應 `go test` 綠燈。

## 2. Idle timeout（idle timeout 由 30 分鐘縮短為 5 分鐘（backstop，可調整））

- [x] [P] 2.1 在 `cmd/launchpal-privhelper/helper_test.go`（或 `internal/privhelper/server_test.go` 既有 idle 測試）新增測試涵蓋 spec「Idle timeout」：以可覆寫的 `idleTimeout` 斷言在 GUI 仍連線的情況下到期會關閉連線並終結程序（依賴 1.3）、移除 socket；期間有 RPC 活動則不終結（activity 重置計時）。驗證：對應 `go test` 綠燈。
- [x] 2.2 將 `cmd/launchpal-privhelper/main.go` 的 `idleTimeout` 預設常數由 30 分鐘改為 5 分鐘（單一可調常數）。行為：閒置 5 分鐘後 helper 自我終結、Admin Mode 回到 Disabled，下一次系統操作需重新授權。驗證：2.1 測試通過。

## 3. Parent PID watchdog（以「PID + 啟動時間」判定，以 closure 保留既有介面）

- [x] [P] 3.1 新增 `cmd/launchpal-privhelper/procinfo_darwin.go`：以 `kinfo_proc`（`sysctl` `KERN_PROC_PID`，經 `golang.org/x/sys/unix`）讀取指定 PID 的程序啟動時間。契約：回傳可比對的啟動時間值，PID 不存在時回傳錯誤。驗證：`go build` 通過並被 3.3 測試涵蓋。
- [x] 3.2 新增 `cmd/launchpal-privhelper/procinfo_other.go`（非 darwin 退化實作，維持既有 PID-存在判定），確保跨平台 `go build` 不中斷。驗證：`go vet ./...` / 建置通過。
- [x] 3.3 在 `cmd/launchpal-privhelper/helper_test.go` 新增測試涵蓋 spec「Parent PID watchdog」：存活判定 closure 在「PID 存在但啟動時間不符（PID 重用）」時回報父程序已結束。驗證：對應 `go test` 綠燈。
- [x] 3.4 於 `cmd/launchpal-privhelper/main.go` 啟動時記錄父程序啟動時間，並把該啟動時間**捕捉進傳給 `StartParentWatchdog` 的 `alive` closure**（darwin 比對 PID + 啟動時間，非 darwin 退化）；**不更動** `StartParentWatchdog` 與 `alive func(int) bool` 簽章（設計：watchdog 以 PID 加啟動時間判定並以 closure 保留介面）。行為：LaunchPal PID 被回收給他程序時 helper 仍自我終結。驗證：3.3 測試通過且 `go build` 不需改動呼叫端簽章。

## 4. Best-effort Shutdown on handshake failure（置於 LaunchHelper／client.go）

- [x] [P] 4.1 在 `internal/privhelper/client_test.go` 新增測試涵蓋 spec「Disable shuts down helper gracefully」：`LaunchHelper` 在 `Connect` 成功、`Ping` 失敗的路徑上，於關閉連線前送出 best-effort `Shutdown`；在 `Connect` 失敗（無連線）時不送 `Shutdown`。驗證：對應 `go test` 綠燈。
- [x] 4.2 於 `internal/privhelper/client.go` 的 `LaunchHelper`：`Ping` 失敗時先以短逾時送 best-effort `Shutdown` 再 `Close()` 才回傳 error；`Connect` 失敗時維持現狀（無連線可送，交由 backstop）。行為：健康但握手時序失敗的 helper 會被要求退出（設計：best-effort Shutdown 置於 LaunchHelper）。驗證：4.1 測試通過。

## 5. Admin Mode 拆除、pending-disable 與中性斷線訊息（admin_mode.go）

- [x] [P] 5.1 於 `admin_mode.go` 抽出共用 `teardownClient` 輔助（短逾時 `Shutdown` → `Close` → 清除 client/state），並讓既有 `Disable` 改用它。契約：三處拆除邏輯集中一處。驗證：既有 `admin_mode_test.go` Disable 相關測試仍綠燈（`go test`）。
- [x] 5.2 在 `admin_mode_test.go` 新增測試涵蓋 spec「Admin Mode status states」：Requesting 期間 `Disable` 後握手成功 → 最終 `Disabled` 且無存活 helper；且 Requesting 期間的 no-op `Enable` 不清除 pending-disable 意圖（多次點擊 example 表對應行為）。驗證：對應 `go test` 綠燈。
- [x] 5.3 於 `admin_mode.go` 新增 `disableRequested`（受 `a.mu` 保護）：`Disable` 在 `Requesting` 時設立旗標而非 no-op；`Enable` 握手成功後於鎖內快照 client 與旗標、**釋放鎖後**再依旗標以 `teardownClient` 拆除或進入 `Enabled`（避免持鎖跑阻塞 RPC 凍結 `GetAdminModeStatus`）；旗標僅在一次 `Disabled`→`Requesting` 的新 Enable 起始時清除，no-op Enable 不清除（設計：pending-disable 鎖釋放後拆除並共用 teardownClient）。驗證：5.2 測試通過。
- [x] [P] 5.4 在 `admin_mode_test.go` 新增測試涵蓋 spec「Helper crash detection and recovery」：Enabled 期間 helper 連線中斷時，狀態回到 `Disabled` 並帶中性 `admin_session_ended`（非 `helper_crashed`）。驗證：對應 `go test` 綠燈。
- [x] 5.5 於 `admin_mode.go` 的 `handleHelperCrash` 將 Enabled 期間斷線改為 `setState(Disabled, "admin_session_ended")`；於前端（`useAdminMode`/Settings 顯示）將該 reason 呈現為資訊性「Admin Mode session ended — re-enable to continue」而非紅色錯誤（設計：非預期斷線改中性 admin_session_ended 狀態）。驗證：5.4 測試通過 + 前端字串審閱（英文 UI）。

## 6. 驗證與文件

- [x] 6.1 執行 `make test`（Go 測試 + 前端）與 `make lint`，確認全綠且無新違規。驗證：指令輸出無失敗。
- [x] 6.2 更新 `.claude/CLAUDE.md` 的 Admin Mode／privileged-helper 生命週期說明（斷線即結束涵蓋所有路徑、`Stop()` 關閉現有連線、idle 5 分鐘、watchdog 以啟動時間判定且介面不變、best-effort Shutdown 於 client 端、`admin_session_ended` 中性訊息、pending-disable），使文件與行為一致。驗證：內容審閱對照本 change 的 design 契約。
- [ ] 6.3 請使用者手動驗證：啟用 Admin Mode 後關閉 App，確認 `launchpal-privhelper` 程序與 socket 於數秒內消失；啟用後 GUI 仍開著閒置 5 分鐘，確認 helper 程序退出且 Admin Mode 回到 Disabled 並顯示中性訊息。驗證：使用者回報觀察結果。
