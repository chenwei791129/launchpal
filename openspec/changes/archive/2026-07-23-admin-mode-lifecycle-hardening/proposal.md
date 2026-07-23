## Why

安全稽核（run-1）發現 Admin Mode 的 root helper 生命週期依賴「合作式關閉」，在數個路徑下會讓一個以 root 執行、監聽於 0600 socket 的 helper 存活於使用者以為 Admin Mode 已關閉的期間。因為 socket 的唯一驗證是同 UID 的 `LOCAL_PEERCRED`，任何同 UID 程序都能在該窗口重連並驅動 `WritePlist`+`Bootstrap` 取得持久 root。GUI 為非特權程序，無法對 root helper 送出 SIGKILL（`kill(2)` 回傳 `EPERM`），因此修復必須以 helper 端自我終結為主。此變更收斂該生命週期，使「以 root 監聽」的窗口嚴格對齊使用者實際授權的期間。

## What Changes

- **helper 端「client 斷線即結束 session」**：本 helper 為單一 client 設計；`acceptLoop` 在 `handleConn` 返回後（不論因 EOF、讀取錯誤或寫入錯誤）即呼叫 `Stop()` 自我終結，涵蓋所有連線結束路徑而非只有 EOF 末端路徑。這是關閉孤兒窗口的主要機制。
- **`Stop()` 關閉現有連線**：`Stop()` 目前只關 listener，導致 GUI 仍連線時 idle/父程序觸發的停止無法終結程序。改為同時關閉已接受的連線，讓 `handleConn` 解除阻塞、程序真正退出。
- **縮短 idle timeout**：把無 RPC 流量的自動結束時間從 30 分鐘縮短為較短值（設計預設 5 分鐘），作為連線保持開啟但無活動時的 backstop（配合 `Stop()` 關閉連線才能真正終結）。活躍使用者不受影響（每次 RPC 都會重置計時），只有閒置者會較早需要重新授權。
- **Enable 失敗後 best-effort Shutdown（於 client 端）**：因為握手失敗時 `admin_mode.go` 取得的是 nil client（連線已在 `LaunchHelper` 內關閉），此 best-effort `Shutdown` 改置於 `internal/privhelper/client.go` 的 `LaunchHelper`：`Connect` 成功、`Ping` 失敗時於關閉連線前送出；`Connect` 失敗（無連線）時交由 idle/watchdog backstop。
- **parent watchdog 加入身分比對（不改介面）**：`parentAlive` 目前僅用 `syscall.Kill(pid,0)`，對 PID 重用盲目。改為在啟動時記錄父程序啟動時間（macOS 透過 `kinfo_proc`），並把啟動時間捕捉進傳給 `StartParentWatchdog` 的 `alive` closure 中比對；**不更動** `StartParentWatchdog` 與 `alive func(int) bool` 簽章。
- **非預期斷線改中性訊息**：因 idle 自我終結現為 by-design 且 GUI 無法與崩潰區分，Enabled 期間連線中斷改以中性 `admin_session_ended` 狀態呈現，取代紅色 `helper_crashed`。
- **Requesting 期間的 Disable 不再被丟棄**：`Disable` 目前在狀態非 `Enabled` 時直接返回，使 osascript 提示期間的 Disable 成為 no-op，隨後成功授權會違反使用者最後意圖地進入 `Enabled`。改為記錄 pending-disable 意圖（釋放鎖後才拆除以免凍結狀態查詢），握手成功後若該意圖存在則拆除 helper 並回到 `Disabled`；意圖僅在一次真正開始請求的 Enable 起始時清除。拆除序列抽出共用 `teardownClient` 輔助供三處使用。

## Capabilities

### New Capabilities

(none)

### Modified Capabilities

- `privileged-helper-lifecycle`: 新增「Helper self-terminates on client disconnect」需求（涵蓋所有連線結束路徑，且 `Stop()` 關閉現有連線）；「Idle timeout」需求的門檻由 30 分鐘縮短為設計指定的較短值並於閒置時關閉連線以真正終結；「Parent PID watchdog」需求改為以「PID + 啟動時間」判定父程序存活（以 closure 保留既有介面），避免 PID 重用誤判。
- `admin-mode`: 「Disable shuts down helper gracefully」需求擴充為涵蓋「連線已建立但 Ping 失敗」時於 client 端送出 best-effort Shutdown；「Admin Mode status states」需求新增對 Requesting 期間 Disable 的 pending-disable 處理與明確的 reset 語意；「Helper crash detection and recovery」需求改為對 Enabled 期間斷線呈現中性 `admin_session_ended` 狀態而非崩潰錯誤。

## Impact

- Affected specs: `privileged-helper-lifecycle`, `admin-mode`
- Affected code:
  - Modified:
    - internal/privhelper/server.go
    - cmd/launchpal-privhelper/main.go
    - admin_mode.go
    - internal/privhelper/client.go
    - frontend/app/composables/useAdminMode.ts
    - frontend/app/pages/settings.vue
  - New:
    - cmd/launchpal-privhelper/procinfo_darwin.go
    - cmd/launchpal-privhelper/procinfo_other.go
  - Removed: (none)
- Tests affected: internal/privhelper/server_test.go, admin_mode_test.go, cmd/launchpal-privhelper/helper_test.go
- Behavior change surfaced to users: idle-timeout shortening means an Admin Mode session left idle auto-disables sooner and the next system-daemon operation re-prompts for authorization. No change for active users.
