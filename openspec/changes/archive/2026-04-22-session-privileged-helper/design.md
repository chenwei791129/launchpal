## Context

LaunchPal 的 `SystemManager` 目前強制唯讀（`internal/launchctl/system.go` 所有寫入方法回 `ErrReadOnlyManager`），主因是 `/Library/LaunchDaemons` 的寫入、以及 `launchctl bootstrap/bootout/kickstart system/...` 操作都需要 root 權限，而未簽章的 LaunchPal 無法走 `SMAppService` 註冊 privileged helper daemon。

經 discuss 階段評估過四種路徑：
- Y（每次 osascript 提權）：UX 差，批次操作體驗不佳
- X（未簽章持久 root daemon）：本機安全性弱、留痕、違背最小權限
- 整 app 以 root 重啟：攻擊面過大（Chromium + 前端都在 root）、檔案權限汙染
- **Z（session-scoped helper）**：每 session 一次密碼、無殘留、小攻擊面 — 勝出

Z 的核心是：LaunchPal 在 user 身份啟動，使用者按下「Enable Admin Mode」時，透過 `osascript ... with administrator privileges` 執行 `launchpal-privhelper --socket <path>` 為背景行程（root 身份）。helper 建立 Unix socket 於 `$TMPDIR` 的隨機路徑，LaunchPal 連上後透過 JSON-over-newline RPC 指揮 helper 執行寫入操作。LaunchPal 退出時 helper 偵測 parent pipe 斷或 watchdog 失聯即 self-exit。

`$TMPDIR` 在 macOS 是 per-user 的 `/var/folders/.../T/`，對其他 user 不可見；加上 socket 0600 + 隨機檔名 + `SO_PEERCRED` 驗連線端 UID，即可在未簽章前提下守住「只有啟動它的 user 的這個 LaunchPal 實例能呼叫」。

## Goals / Non-Goals

**Goals:**

- 使用者在一個 LaunchPal session 內只輸入一次授權，即可自由管理 `/Library/LaunchDaemons` 下的服務
- LaunchPal 關閉後機器上完全不殘留 root 行程、daemon plist 或 socket
- 同機其他 user 或其他 process 無法呼叫 helper
- 授權 / 連線失敗時 UI 清楚呈現狀態，不靜默退回唯讀
- helper binary 小、職責單一，易於 audit
- 前端對 Admin Mode 狀態變化反應即時，且能在 Admin Mode 關閉時優雅 degrade 回唯讀

**Non-Goals:**

- 不跨 session 快取授權（不引入 keychain、helper plist、sudo token 常駐）
- 不支援 `/System/Library/LaunchDaemons` 的寫入
- 不實作 osascript per-operation 提權作為替代方案
- 不處理 app 崩潰時 helper 可能存活的邊角（由 parent watchdog 兜底）
- 不實作進度回報串流（寫入操作視為同步短命令）

## Decisions

### Helper 啟動通道（osascript + 背景行程）

在 Go 中透過 `exec.Command("osascript", "-e", script)` 執行：

```applescript
do shell script "/path/to/launchpal-privhelper --socket /var/folders/.../launchpal-<uid>-<rand>.sock --parent-pid <ppid> &> /dev/null &" with administrator privileges
```

關鍵點：
- `&` 讓 helper 背景執行，`do shell script` 立即回
- `--parent-pid` 傳入 LaunchPal PID 供 helper watchdog
- `&> /dev/null` 避免 helper 的 stdout/stderr 被 osascript 收集造成阻塞
- script 執行完成意味「helper 已 fork 完成」，不代表「helper 已 ready」。LaunchPal 要對 socket 路徑做 retry connect（例如 500ms × 10 次）直到成功或超時

**Alternatives considered:**
- `AuthorizationExecuteWithPrivileges`：已被 Apple deprecated 且 macOS 13+ 可能被移除，不可靠
- `sudo -A` + askpass helper：askpass UI 要自寫，體驗不如原生 osascript 對話框
- launchd 臨時 job：需 root 已經在手才能 bootstrap，chicken-and-egg

### Helper binary 位置與打包

helper 編譯為 `launchpal-privhelper` 並放進 app bundle：`LaunchPal.app/Contents/MacOS/launchpal-privhelper`（與主 binary 同目錄，便於相對路徑找到）。

啟動時 LaunchPal 用 `os.Executable()` 取主 binary 路徑，再 `filepath.Join(dir, "launchpal-privhelper")` 組出 helper 路徑傳給 osascript。

**為何放在 bundle 內**：使用者透過 DMG / brew cask 安裝時 helper 自動隨附，無額外安裝步驟。

**Alternatives considered:**
- 放在 `Contents/Library/PrivilegedHelperTools/`：Apple 的 SMJobBless 傳統位置，但 Z 路線不走該路徑，放這裡反而誤導；放 `Contents/MacOS/` 更合語義
- 在 build 時下載：增加 supply chain 風險

### RPC 協議（newline-delimited JSON）

每條訊息一行 JSON，格式：

```json
{"id": 1, "method": "Bootstrap", "params": {"plistPath": "/Library/LaunchDaemons/com.example.plist"}}
{"id": 1, "result": {"ok": true}}
{"id": 2, "method": "WritePlist", "params": {"plistPath": "...", "data": "<base64>"}}
{"id": 2, "error": {"code": "permission_denied", "message": "..."}}
```

方法集：
- `Ping` — 連線 handshake 與 keepalive
- `ListSystemDaemons` — 回 `[]ServiceInfo`（label + status + pid）
- `GetSystemDaemon(label)` — 回單一 `ServiceInfo`
- `Bootstrap(plistPath)` — `launchctl bootstrap system <plistPath>`
- `Bootout(label)` — `launchctl bootout system/<label>`
- `Kickstart(label)` — `launchctl kickstart -k system/<label>`
- `WritePlist(plistPath, data)` — atomic write + 備份
- `DeletePlist(plistPath)` — 備份 + rm
- `Shutdown` — 命令 helper 優雅退出

**為何選 newline JSON 不選 gRPC / protobuf**：單一 binary、單一進程、對稱語言、少量 method — protobuf 的工具鏈成本遠大於收益。newline JSON 用 `bufio.Scanner` 一行搞定。

**為何 one-at-a-time 不做並行**：helper 執行的都是 launchctl / file IO，本質 serializable；並行控制是複雜度來源；LaunchPal 的 UI 操作也不需要同時送多個請求。

### Socket 路徑與權限

- 路徑：`$TMPDIR/launchpal-<uid>-<16-hex-random>.sock`
- 建立順序：LaunchPal 產生路徑字串 → 傳給 osascript → helper 建立 socket → helper `chmod 0600` → helper 接受連線 → LaunchPal connect retry
- `SO_PEERCRED`（macOS 為 `LOCAL_PEERCRED` via `getsockopt`）驗連線端 UID 等於 helper 啟動時的 UID（從 `--launching-uid` 參數傳入，或 helper 自 query 環境變數 `SUDO_UID` / `USER`）
- helper 關閉時 `os.Remove(socketPath)` 清除

**為何用 `$TMPDIR` 而非 `/tmp`**：macOS 的 `$TMPDIR` 是 `/var/folders/<hash>/.../T/`，**對其他 user 不可讀**，天然 per-user 隔離。`/tmp` 全域可見。

**Alternatives considered:**
- `/var/run/` 或 `/Library/Application Support/`：需 root 建立、跨 session 留痕，違反 Z 原則
- 只用 socket 檔權限不驗 `SO_PEERCRED`：檔案權限已足夠擋大多情況，但 double check 成本極低

### Parent watchdog

helper 啟動時收到 `--parent-pid <pid>` 參數，背景 goroutine 每秒檢查：

```go
if err := syscall.Kill(parentPID, 0); err != nil { /* parent 死了 */ }
```

parent 死了就清 socket、退出。同時 helper 對 socket 的 accept loop 設 idle timeout（例如 10 分鐘無流量即視為孤兒，主動退）。

**為何不用 pipe EOF**：osascript 的 `do shell script` 不能方便保留 pipe；watchdog 更可靠。

### `SystemManager` 雙模式改造

保留 `SystemManager` 名稱，新增內部狀態 `adminClient *privhelper.Client`：

- `nil` → 舊行為，write ops 回 `ErrReadOnlyManager`
- 非 `nil` → write ops 透過 client 呼叫 helper RPC

切換由 `app.go` 的 `EnableAdminMode` / `DisableAdminMode` 控制，對外呼叫 `systemManager.SetAdminClient(client)`。

**為何不開新 struct type**：UI 不需要感知「哪個 manager」；切換 admin mode 時若換 manager type，前端拿到的物件 identity 會斷，徒增複雜度。

**Alternatives considered:**
- `PrivilegedSystemManager` 新類型繼承 `SystemManager`：Go 沒繼承，用 composition 會讓 interface 方法分派混亂
- 在 Manager interface 加 `IsWritable()`：跨 type 洩漏 admin mode 概念

### AppleSystemManager 不變

`/System/Library/LaunchDaemons` 屬於 Apple 系統服務，不應由使用者工具修改（SIP 也會擋）。`AppleSystemManager` 即使 Admin Mode 開啟也**繼續回 `ErrReadOnlyManager`**。UI 在 apple-system 列表不顯示寫入按鈕。

### BackupManager 的 root 寫入

現有 BackupManager 寫到 `~/.launchpal/backups/`（user 家目錄）。helper 以 root 跑時，寫到 `~/.launchpal/` 會被解讀成 `/var/root/.launchpal/`，跟 user 的備份分家。

解法：helper 啟動時收到 `--user-home <path>` 參數，backup 路徑顯式用該 path，不依賴 `$HOME`。檔案建立後 `chown` 回 user uid:gid，使用者從 LaunchPal（user 身份）才讀得到 backup。

**Alternatives considered:**
- 兩份 backup（user 一份、root 一份）：混亂、diff 不一致
- backup 由 LaunchPal client 端做，helper 只負責原子寫：client 要先讀舊 plist（無權限），做不到

### Admin Mode 狀態機

```
Disabled ──Enable──▶ Requesting ──success──▶ Enabled
              │           │
              │           └──cancel/error──▶ Disabled
              │                             (with error detail)
              └──────────────────────────────────────▶
Enabled ──Disable──▶ Shutting Down ──▶ Disabled
Enabled ──helper crash──▶ Disabled (with error detail)
```

前端 Settings 頁顯示當前狀態與 error detail；System 列表與詳情頁依 `AdminModeStatus` 決定是否顯示寫入按鈕。

### 錯誤處理與回復

- osascript 授權取消 → `ErrAuthorizationCanceled`，Settings 顯示「已取消」、不視為錯誤
- helper 啟動但連不上 socket（超時）→ kill helper（若能找到）、Admin Mode 回 Disabled、提示使用者重試
- helper 中途 crash → client detect EOF，`adminClient` 置 nil、狀態回 Disabled、顯示錯誤
- LaunchPal 不正常退出 → helper watchdog 在 1 秒內 self-exit

## Risks / Trade-offs

- **[未簽章 root helper 的審查疑慮]** → 程式碼量小（預期 < 500 LOC）、邏輯單純；可在 README 放審查連結並提供 build reproducibility；未來取得 Developer ID 後可升級 SMAppService 取代
- **[osascript AppleEvents 沙箱 / TCC 限制]** → LaunchPal 首次呼叫 osascript 可能觸發「允許 LaunchPal 控制 System Events」授權視窗；需在 Settings 頁文件化此現象並提供說明連結
- **[helper 啟動失敗但 osascript 沒報錯（例如 binary 不存在）]** → LaunchPal connect 會超時；此時主動 kill helper 在此路徑上無意義（根本沒活），僅需 timeout + 錯誤訊息帶出 `launchpal-privhelper not found` 等診斷資訊
- **[parent watchdog 可能漏殺 helper（極罕見 race）]** → 次要緩解：accept loop idle timeout、SIGTERM handler；最壞情況下有個 orphan root process 直到使用者重開機或 `kill`，不會被其他 app 連上（socket 會 stale 但 accept loop 仍只認對的 UID）
- **[RPC 協議未來擴充]** → JSON 是 forward-compatible，新欄位可加；method 不足時走 minor version bump + handshake negotiation
- **[與既有 `ErrReadOnlyManager` 相容性]** → `SystemManager` 在 Admin Mode 關閉時**仍回原錯誤**，前端與測試無需改；只有 Admin Mode 開啟的分支是全新路徑
- **[Helper binary 被使用者拿去單獨執行的風險]** → helper 需要 `--socket` 參數且會驗 `SO_PEERCRED`；單獨跑沒有 client 連上，無實際危害；仍要在 helper 啟動時顯式拒絕無參數執行、拒絕非 root UID 執行（`if os.Geteuid() != 0 { exit }`）
