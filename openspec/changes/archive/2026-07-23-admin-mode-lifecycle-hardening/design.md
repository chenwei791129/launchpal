## Context

Admin Mode 透過 `osascript ... with administrator privileges` 啟動一個以 root 執行的 `launchpal-privhelper`，並以 0600 Unix socket + `LOCAL_PEERCRED`（同 UID 比對）作為唯一驗證。GUI 以非特權使用者身分執行，因此**無法**對 root helper 送 SIGKILL（`kill(2)` → `EPERM`）——任何「由 GUI 強制終結 helper」的方案都不可行，除非再次提權。

現況（安全稽核 run-1 + 計畫 code-review 對照程式碼確認）：

- `LaunchHelper`（`internal/privhelper/client.go`）以 `&` 背景化啟動 helper 且不保留其 PID；client 端只持有一條長連線（`a.client`，被所有系統操作重用）。
- `LaunchHelper` 在 `Connect` 成功但 `Ping` 失敗時，已在內部呼叫 `client.Close()` 並回傳 `(nil, err)`——因此 `admin_mode.go` 在握手失敗後**沒有可用的 client 可送 Shutdown**；任何 best-effort Shutdown 必須在 `LaunchHelper` 內、於連線尚存時送出。
- `Server.Stop()`（`internal/privhelper/server.go`）只關閉 listener、移除 socket、送出 stop 訊號，**不關閉已接受的連線**；因此當 GUI 長連線仍開啟時，idle 觸發的 `Stop()` 無法讓 `handleConn`（阻塞於 `scanner.Scan()`）解除阻塞，helper 程序不會退出。
- `handleConn` 會從迴圈內多個路徑返回：`isStopped()`、`encoder.Encode()` 寫入錯誤、shutdown 請求，以及迴圈結束後的 EOF/scan 錯誤路徑。若只在 EOF 路徑加拆除，寫入錯誤造成的斷線會被漏掉。
- 伺服器在 client 斷線後仍保留 accept 迴圈，socket 只在 `Stop()` 移除。
- idle timeout 30 分鐘、且每次 RPC 都會重置——攻擊者重連並 ping 可無限延長。
- `parentAlive` 僅 `syscall.Kill(pid,0)`，對 PID 重用盲目。`StartParentWatchdog(parentPID int, interval, alive func(int) bool, onExit func())` 的注入判定函式**只收 PID**。
- `Disable` 在狀態非 `Enabled` 時直接返回，Requesting 期間的取消會被丟棄。
- helper 連線中斷經 `Client.OnDisconnect` → `handleHelperCrash`，一律 `setState(Disabled, "helper_crashed")`。

本變更僅處理生命週期收斂；不涉及 helper 二進位完整性/簽章（Finding 1，另案），也不涉及檔案系統/輸入驗證（Change 2）。

## Goals / Non-Goals

**Goals:**

- 讓「以 root 監聽於 socket」的窗口嚴格對齊使用者實際授權的 Admin Mode 期間。
- 修復以 helper 端自我終結為主，不依賴 GUI 對 root 程序送訊號。
- 消除 PID 重用導致 helper 活過父程序的誤判，且不更動既有 watchdog 介面。
- 讓 Requesting 期間的 Disable 意圖被尊重，且行為對多次點擊具決定性。
- 讓 by-design 的 idle 自我終結不被誤呈現為崩潰錯誤。

**Non-Goals:**

- helper 二進位的簽章/完整性驗證（Finding 1，另案；使用者無 Apple Developer Program）。
- 檔案系統 symlink / 路徑 / 輸入驗證強化（Change 2：#4/#5/#6/#7/#8）。
- 改變 osascript 提權機制或 socket 驗證模型本身。
- 為 helper 引入跨 session 憑證快取。

## Decisions

### 斷線即自我終結涵蓋所有連線結束路徑

helper 為單一 client 設計。在 `acceptLoop` 中，`s.handleConn(conn)` 返回後（不論其因 EOF、讀取錯誤、或 `encoder.Encode()` 寫入錯誤而返回），若伺服器尚未在停止中（`!isStopped()`），即觸發 `Stop()`。把拆除掛在「`handleConn` 返回」這一單點，可涵蓋所有連線結束路徑，避免只掛在 EOF 末端路徑而漏掉寫入錯誤造成的斷線。因為 GUI 全程持有一條長連線，正常多步操作期間連線不會關閉——此機制只在 Disable、GUI 結束、GUI 崩潰、或真實斷線時觸發，因此不影響正常使用。這是唯一不需 GUI 對 root 程序送訊號即可可靠拆除的方案。

替代方案：GUI 保留 helper PID 後 SIGKILL——已否決，GUI 非特權會得到 `EPERM`。只在 `handleConn` 的 EOF 末端路徑加拆除——已否決，會漏掉寫入錯誤返回路徑。

### Stop() 關閉現有已接受連線

`Server.Stop()` 目前只關 listener，不關已接受連線。由於 GUI 的長連線在閒置時仍保持開啟，idle watchdog 呼叫 `Stop()` 時 `handleConn` 仍阻塞於 `scanner.Scan()`，程序不會退出——idle backstop 形同虛設。因此 `Stop()` 需追蹤並關閉目前的已接受連線（關閉後 `scanner.Scan()` 回傳、`handleConn` 返回、`connWG` 完成、程序退出）。此變更同時讓「idle 到期」「父程序消失」都能終結一個仍連線著的 helper。

### idle timeout 由 30 分鐘縮短為 5 分鐘（backstop，可調整）

在連線保持開啟但無 RPC 流量的情況（GUI 仍在、Admin Mode 仍 Enabled、使用者離開）下，idle timeout 是限制「站立中 root socket」窗口的 backstop（配合上一決策，idle-stop 會關閉連線以真正終結程序）。設計採 5 分鐘：活躍使用者不受影響（每次 RPC 重置計時），只有閒置者會較早需要重新授權。此值是本變更唯一有感的行為改變，實作為單一常數以便日後調整。

替代方案：維持 30 分鐘——已否決，孤兒窗口過長；設為 0/停用 idle——已否決，失去離開情境的 backstop。

### watchdog 以 PID 加啟動時間判定並以 closure 保留介面

啟動時透過 macOS `kinfo_proc`（`sysctl` `KERN_PROC_PID`，經 `golang.org/x/sys/unix`）讀取父程序啟動時間並記錄。**不更動** `StartParentWatchdog` 或其注入的 `alive func(int) bool` 簽章；改為把記錄下的啟動時間**捕捉進傳入的 `alive` closure**（例如 `alive: func(pid int) bool { return pidExistsWithStart(pid, recordedStart) }`）。watchdog 每次輪詢比對「PID 仍存在且啟動時間一致」；PID 存在但啟動時間不同（PID 已被回收）時視為父程序已結束並自我終結。非 darwin 平台提供退化實作（維持既有 `Kill(pid,0)` 行為即可，因本產品僅於 macOS 發佈）。

替代方案：更動 `alive`/`StartParentWatchdog` 簽章——已否決，會造成不必要的介面 churn 與呼叫端改動，closure 捕捉即可達成。pipe-EOF（父程序持 pipe 一端，helper 於 EOF 退出）——已否決，`with administrator privileges` 的提權 trampoline 不保證繼承任意父程序 fd，跨提權邊界傳遞 pipe 不可靠。

### best-effort Shutdown 置於 LaunchHelper

因為 `admin_mode.go` 在握手失敗後拿到的是 `(nil, err)`（連線已於 `LaunchHelper` 內關閉），best-effort Shutdown 必須在 `LaunchHelper` 內、於連線尚存時送出：當 `Connect` 成功但 `Ping` 失敗時，先以短逾時送 best-effort `Shutdown`、再 `Close()`、才回傳 error。當 `Connect` 本身失敗（從未建立連線）時，沒有通道可送 Shutdown，交由 helper 自身的 idle timeout 與 parent watchdog 兜底。此決策同時對齊 proposal Impact 已列出的 `internal/privhelper/client.go`（原本無對應任務）。

### 非預期斷線改中性 admin_session_ended 狀態

idle 自我終結現在是 by-design 行為，而 GUI 端看到的都是相同的 EOF/連線錯誤，無法區分「idle 自我終結／乾淨拆除」與「真正崩潰」。因此 `handleHelperCrash` 對「Enabled 期間連線中斷」改為 `setState(Disabled, "admin_session_ended")`，前端以資訊性「Admin Mode session ended — re-enable to continue」呈現（非紅色錯誤）。寫入控制如同任何 `Disabled` 狀態隱藏。

替代方案：維持 `helper_crashed`——已否決，會把正常的 idle 逾時呈現為嚇人的崩潰；讓 helper 在乾淨退出前先送訊號以區分——已否決，連線關閉後 GUI 兩種情況都只看到 EOF，無可靠區分手段。

### pending-disable 鎖釋放後拆除並共用 teardownClient

新增 `disableRequested` 旗標（受既有 `a.mu` 保護）。`Disable` 在狀態為 `Requesting` 時設立該旗標並返回（而非無效 no-op）。`Enable` 握手成功後，於鎖內快照 client 與旗標、**釋放鎖後**才執行拆除（避免像既有 `Disable` 一樣在持鎖期間跑阻塞的 Shutdown RPC，否則並行的 `GetAdminModeStatus` 會被卡最多 3 秒）；若旗標為真則不進入 `Enabled`，改為拆除 helper 並回到 `Disabled`。

reset 語意需明確：`disableRequested` **僅在一次真正開始請求的 Enable（`Disabled` → `Requesting` 轉移）起始時清除**；因狀態已是 `Requesting`/`Enabled` 而提早 no-op 返回的 Enable **不清除**旗標。這使多次點擊具決定性（見 spec 的 example 表）。

拆除序列（短逾時 `Shutdown` → `Close` → 清除 client/state）目前散落於既有 `Disable`、握手失敗路徑、pending-disable 路徑三處。抽出單一 `teardownClient` 輔助函式供三處共用，避免日後修改（如調整逾時或加 `ClearAdminClient`）需三處同步而漂移。

替代方案：讓 Disable 阻塞等待 Enable 完成——已否決，osascript 提示為系統 modal，會造成 UI 卡住；在持鎖期間拆除——已否決，會凍結狀態查詢。

## Implementation Contract

**行為（使用者/呼叫者可觀察）：**

- Disable、GUI 結束、GUI 崩潰或連線中斷（含寫入錯誤造成的斷線）後，`launchpal-privhelper` 程序於數秒內消失，且 socket 檔被移除；此後任何連線嘗試失敗。
- Admin Mode 在 GUI 仍連線但閒置達 5 分鐘後，helper 關閉連線並自我終結、程序退出；下一次系統 daemon 操作需重新授權；有 RPC 活動則不受影響。
- Enabled 期間 helper 連線中斷（含 idle 自我終結）時，狀態回到 `Disabled` 並以中性 `admin_session_ended` 呈現（非崩潰錯誤）。
- osascript 成功、連線已建立但 Ping 失敗時，GUI 於該連線送出 best-effort `Shutdown` 再關閉，回到 `Disabled`（`helper_handshake_failed`）；連線從未建立時不送 Shutdown，交由 backstop。
- 使用者在 osascript 提示期間點 Disable、隨後完成授權時，最終狀態為 `Disabled` 且無存活 helper/socket；期間的 no-op Enable 不覆蓋此取消意圖。
- LaunchPal 父程序 PID 被回收給他程序時，helper 不再誤判父程序存活，仍會自我終結。

**介面/契約：**

- 伺服器：`acceptLoop` 於 `handleConn` 返回後（若非停止中）觸發 `Stop()`；`Stop()` 除既有語意外，另關閉目前已接受的連線。`Stop()` 的冪等性（`stopped` 守衛）維持。
- helper watchdog：新增父程序啟動時間擷取（darwin `kinfo_proc` + 非 darwin 退化）；**不更動** `StartParentWatchdog` 與 `alive func(int) bool` 簽章，改由 closure 捕捉啟動時間。
- 常數：idle timeout 預設值改為 5 分鐘（單一可調常數）。
- `client.go`：`LaunchHelper` 在 `Connect` 成功、`Ping` 失敗時，於關閉連線前送 best-effort `Shutdown`（短逾時）。
- `admin_mode`：新增 `disableRequested`（受 `a.mu` 保護）；新增共用 `teardownClient` 輔助；`Enable` 成功路徑於鎖內快照、釋放鎖後依旗標拆除或進入 `Enabled`；`Disable` 於 `Requesting` 期間設立旗標；`handleHelperCrash` 對 Enabled 期間斷線改用 `admin_session_ended`。Wails 綁定 `EnableAdminMode`/`DisableAdminMode`/`GetAdminModeStatus` 對外簽章不變。

**失敗模式：**

- best-effort Shutdown 送不到或無連線可送：靜默略過，由「斷線即結束」、idle backstop、parent watchdog 兜底。
- 非 darwin 平台無法取得啟動時間：退化為既有 `Kill(pid,0)` 行為，不阻擋建置。
- 暫時性連線中斷誤觸自我終結：使用者以重新授權即可恢復，與既有斷線回到 Disabled 的體驗一致（現改為中性訊息）。

**驗收標準：**

- 新增/更新的 Go 測試涵蓋：client 斷線（EOF／讀取錯誤／寫入錯誤）後伺服器 `Stop()` 並移除 socket、後續連線失敗；`Stop()` 會關閉現有連線使 `handleConn` 解除阻塞；idle timeout 使用新常數，且在 GUI 仍連線的情況下到期會真正終結程序；watchdog closure 在「PID 相同但啟動時間不同」時判定父程序已死；`LaunchHelper` 在 Ping 失敗路徑會於關閉前送 Shutdown；Requesting 期間 Disable 後成功握手最終落在 Disabled 且無存活 helper，no-op Enable 不清除意圖；Enabled 期間斷線以 `admin_session_ended` 呈現。
- `make test` 綠燈；`make lint` 無新違規。
- 使用者驗證（手動）：啟用 Admin Mode 後關閉 App，確認 helper 程序與 socket 於數秒內消失；啟用後 GUI 仍開著但閒置 5 分鐘，確認 helper 程序退出且 Admin Mode 回到 Disabled 並顯示中性訊息。

**範圍邊界：**

- 僅限本 change 列出的生命週期行為與其測試。
- 不修改 helper 的 RPC 方法集、路徑/label 驗證、log handler、或任何檔案系統寫入邏輯（屬 Change 2 / Finding 1）。

## Risks / Trade-offs

- [縮短 idle timeout 造成更頻繁的重新授權] → 僅影響閒置使用者；活躍操作會重置計時；值集中於單一常數，可依回饋調整。
- [「斷線即結束」若 GUI 連線因暫時性問題中斷會誤殺 helper] → 對本地 Unix socket 而言暫時性中斷極罕見；且斷線本來就會回到 Disabled，改為中性訊息後體驗更佳（重新授權即可）。
- [`Stop()` 關閉連線與 `handleConn` 併發] → 以既有 `stopped` 守衛與 `connWG` 協調；測試涵蓋 idle-stop 與斷線兩條路徑不重複拆除。
- [best-effort Shutdown 置於 client.go 增加該檔改動面] → 已在 proposal Impact 列出並補上對應任務；改動集中於 `LaunchHelper` 單一函式。
- [kinfo_proc 讀取啟動時間的平台差異] → darwin 實作 + 非 darwin 退化實作；本產品僅於 macOS 發佈，退化路徑僅為建置相容性。
- [pending-disable 與既有狀態機互動產生新競態] → 以既有 mutex 保護 `disableRequested`，拆除在釋放鎖後執行避免凍結狀態查詢，並以測試涵蓋 Requesting→Disable→握手成功與多次點擊序列。
- [中性化斷線訊息會遮蔽真正崩潰] → 兩種情況的使用者動作都是「重新啟用」，訊息提示 re-enable；接受此權衡以避免把正常 idle 逾時呈現為崩潰。

## Migration Plan

無資料遷移。純執行期行為變更，隨版本更新即生效。回退策略：還原本 change 的程式碼變更即可回到舊生命週期行為（無持久化狀態需清理）。

## Open Questions

- idle timeout 的最終值（設計採 5 分鐘）是否符合使用者對「離開後多久要求重新授權」的偏好——可於實作後依實際體驗調整此單一常數。
