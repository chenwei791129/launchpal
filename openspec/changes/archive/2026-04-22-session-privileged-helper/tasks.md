## 1. 共用型別與協議基礎

- [x] 1.1 建立 `internal/privhelper/` package，定義 RPC request/response struct（`id`、`method`、`params`、`result`、`error`）及 error code 常數，落實「Newline-delimited JSON RPC transport」與「Error code taxonomy」— 對應決策「RPC 協議（newline-delimited JSON）」
- [x] 1.2 定義 `MethodName` 常數集合涵蓋所有「Supported RPC methods」：Ping / ListSystemDaemons / GetSystemDaemon / Bootstrap / Bootout / Kickstart / WritePlist / DeletePlist / Shutdown
- [x] 1.3 [P] 寫 `privhelper/protocol_test.go`：JSON 序列化 / 反序列化 round-trip、malformed JSON 處理

## 2. Helper binary 骨架（TDD）

- [x] 2.1 建立 `cmd/launchpal-privhelper/main.go` 入口，parse `--socket`、`--parent-pid`、`--launching-uid` 參數
- [x] 2.2 寫 `cmd/launchpal-privhelper/main_test.go`（或 `helper_test.go`）覆蓋「Helper refuses to run without required conditions」：非 root、缺 socket、缺 parent-pid、缺 launching-uid 各情境
- [x] 2.3 實作參數驗證與 early-exit 邏輯通過 2.2 測試
- [x] 2.4 落實「Helper binary packaged in app bundle」決策「Helper binary 位置與打包」：修改 `Makefile` 新增 build target 編譯 helper 至 `build/bin/launchpal-privhelper`，並於 app bundling 步驟複製到 `Contents/MacOS/launchpal-privhelper`

## 3. Socket server 與 peer 驗證

- [x] 3.1 寫 `privhelper/server_test.go` 覆蓋「Socket path and permissions」：socket 建立後 mode=0600、shutdown 後檔案被移除
- [x] 3.2 實作 `privhelper/server.go`：socket listen、`chmod 0600`、accept loop；對應決策「Socket 路徑與權限」
- [x] 3.3 寫測試覆蓋「Peer UID verification」：matching UID 接受、mismatched UID 拒絕並不讀 input
- [x] 3.4 實作 `LOCAL_PEERCRED` 查詢（via `golang.org/x/sys/unix` 或 cgo）並整合進 accept loop
- [x] 3.5 寫測試覆蓋「Serial request processing」：pipelined 請求按序處理
- [x] 3.6 實作 per-connection serial dispatcher

## 4. Parent watchdog 與 lifecycle

- [x] 4.1 寫測試覆蓋「Parent PID watchdog」：模擬 parent 不存在時 helper 2 秒內退出（可用測試 helper 行程 + short poll interval）— 對應決策「Parent watchdog」
- [x] 4.2 實作 1 秒 poll `syscall.Kill(pid, 0)` 的 watchdog goroutine
- [x] 4.3 寫測試覆蓋「Idle timeout」：30 分鐘無流量退出（測試時以注入 short timeout 驗證）
- [x] 4.4 實作 idle timer（每次收到 RPC 重置）
- [x] 4.5 寫測試覆蓋「Graceful shutdown via RPC」：收到 Shutdown 後回 ack、關閉 listener、清 socket、exit 0
- [x] 4.6 實作 Shutdown method handler

## 5. Write-op RPC methods（TDD）

- [x] 5.1 寫 `privhelper/handlers_test.go` 覆蓋「Bootstrap a system daemon」：路徑驗證（必須在 /Library/LaunchDaemons/ 下）、launchctl 成功 / 失敗情境（以 fake exec 注入）
- [x] 5.2 實作 `Bootstrap` handler
- [x] 5.3 寫測試覆蓋「Bootout a system daemon」與「Kickstart a system daemon」：label 格式驗證（拒絕 shell metachar）、launchctl 成功 / 失敗
- [x] 5.4 實作 `Bootout` 與 `Kickstart` handler
- [x] 5.5 寫測試覆蓋「Atomic plist write with backup」：路徑驗證、atomic rename、檔案 mode 0644、root:wheel ownership、覆蓋時觸發備份且 backup 檔 chown 回 launching user
- [x] 5.6 實作 `WritePlist` handler，處理決策「BackupManager 的 root 寫入」：接 `--user-home` 參數、備份檔 chown 回 user uid/gid
- [x] 5.7 寫測試覆蓋「Delete plist with backup」：deletion with prior backup、non-existent file 回 not_found
- [x] 5.8 實作 `DeletePlist` handler
- [x] 5.9 寫測試覆蓋「List and get system daemons via helper」：parse `launchctl print system` 輸出
- [x] 5.10 實作 `ListSystemDaemons` 與 `GetSystemDaemon` handler

## 6. Client 端啟動與 handshake

- [x] 6.1 寫 `privhelper/client_test.go` 覆蓋「Socket handshake with retry」：1 秒內 ready → 成功；10 秒超時 → 失敗
- [x] 6.2 實作 `privhelper/client.go` 的 `Connect` 函式（exponential backoff retry）
- [x] 6.3 實作 `privhelper.LaunchHelper(helperPath, launchingUID, parentPID) (*Client, error)`：組 osascript 指令、exec、回 client — 對應決策「Helper 啟動通道（osascript + 背景行程）」
- [x] 6.4 實作 osascript 執行與授權取消偵測（區分 `ErrAuthorizationCanceled` 與其他錯誤，對應「Helper launched via osascript with background execution」scenario）
- [x] 6.5 實作 Client 端的 request/response correlation：map id → channel、background reader goroutine

## 7. SystemManager 雙模式

- [x] 7.1 寫 `internal/launchctl/system_test.go` 新案例覆蓋「Read-only managers reject write operations」的新語意：Admin Mode Disabled 時回 `ErrReadOnlyManager`、Enabled 時 delegate 到 helper client
- [x] 7.2 修改 `internal/launchctl/system.go` 的 `SystemManager`，新增 `SetAdminClient(c *privhelper.Client)` / `ClearAdminClient()`；write 方法依 adminClient 決定路徑 — 對應決策「`SystemManager` 雙模式改造」
- [x] 7.3 確認 `AppleSystemManager` 行為不變（「AppleSystemManager 不變」決策 + scenario「AppleSystemManager write operations always rejected」）
- [x] 7.4 跑既有 `system_test.go`、`apple_system_test.go` 無回歸

## 8. Admin Mode 狀態管理（app.go）

- [x] 8.1 在 `app.go` 新增 `AdminModeStatus` struct 與內部狀態欄位（遵循「Admin Mode 狀態機」決策）
- [x] 8.2 實作「Admin Mode status states」轉移邏輯：Disabled → Requesting → Enabled / Disabled，包含「Successful enablement path」、「User cancels authorization」、「Handshake failure」各 scenario
- [x] 8.3 實作 Wails binding `EnableAdminMode() error`、`DisableAdminMode() error`、`GetAdminModeStatus() AdminModeStatus` — 落實「Wails bindings for Admin Mode」
- [x] 8.4 實作「Disable shuts down helper gracefully」：Shutdown RPC + 3 秒 timeout，逾時仍進 Disabled
- [x] 8.5 實作「Helper crash detection and recovery」：client EOF 觸發狀態轉 Disabled 並記錄 `helper_crashed` error；對應決策「錯誤處理與回復」
- [x] 8.6 狀態變更透過 Wails event 通知前端（若 Wails 支援）或由前端 poll

## 9. 前端 Admin Mode UI

- [x] 9.1 [P] 在 `frontend/app/pages/settings.vue` 新增 Admin Mode 區塊，實作「Admin Mode UI in Settings page」：狀態顯示、Enable / Disable 按鈕、最近錯誤、說明文字
- [x] 9.2 [P] 新增 composable `useAdminMode()` 包裝 `EnableAdminMode` / `DisableAdminMode` / `GetAdminModeStatus` 與狀態反應
- [x] 9.3 修改 `frontend/app/pages/system.vue` 與 `frontend/app/pages/services/[name].vue`：實作「Write controls conditionally rendered」— 依 Admin Mode 狀態顯示 / 隱藏 Start/Stop/Restart/Edit/Delete/Create，Disabled 時以鎖頭 + tooltip 呈現
- [x] 9.4 在寫入操作失敗時（helper error），前端依 error code 顯示對應訊息

## 10. 整合測試與審核

- [ ] 10.1 手動 E2E：Enable → 密碼 / Touch ID → System 列表出現寫入按鈕 → Start / Stop 既有 daemon → 觀察狀態更新 → Disable → 按鈕消失（需實機驗證）
- [ ] 10.2 手動測試：LaunchPal 強制 quit（kill -9）後確認 helper 2 秒內自我退出、socket 檔被清除（需實機驗證）
- [ ] 10.3 手動測試：osascript 密碼取消 → Settings 顯示已取消、無 helper 殘留（需實機驗證）
- [x] 10.4 Audit：helper binary 對 label 的 shell metachar 防護（Bootout / Kickstart）；plist 路徑 prefix 檢查無 `..` 或 symlink escape；argument array vs shell 拼接策略
- [x] 10.5 Audit：socket 路徑隨機熵足夠（16 hex = 64 bit）、`$TMPDIR` 假設在目標 macOS 版本成立
- [x] 10.6 Audit：helper 的 chown backup 檔對 symlink attack 安全（`lchown` vs `chown`）
- [x] 10.7 執行 `make test` 確認所有 Go 測試通過
- [ ] 10.8 執行 `make build` 驗證 production bundle 含 helper binary 於正確位置（需完整 frontend build pipeline，留待 release 前驗證）
- [ ] 10.9 執行 `make dmg` 驗證 DMG 打包含 helper、brew cask postflight 不衝突（需完整 release pipeline，留待 release 前驗證）
- [x] 10.10 更新 `.claude/CLAUDE.md`：新增 Admin Mode 架構、helper 位置、已知限制（每 session 要授權）
- [x] 10.11 更新 `README.md`：新增 Admin Mode feature 說明、截圖、首次使用流程
