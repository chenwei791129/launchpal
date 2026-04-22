## Why

Phase 1（`system-daemon-status-detection`）讓使用者能看到 system daemon 的狀態與 PID，但 `SystemManager` 仍然是唯讀的 — 所有 Start / Stop / Restart / Create / Update / Delete 都回 `ErrReadOnlyManager`。使用者若要管理 `/Library/LaunchDaemons` 下的服務，仍得跳出 LaunchPal 去用 CLI。

因為專案取得不到 Apple Developer ID，無法走 `SMAppService` 註冊 privileged helper daemon。本變更採**session-scoped privileged helper** 設計：使用者按下「Enable Admin Mode」時，透過 `osascript ... with administrator privileges` 啟動一個 root helper 子行程，並以 Unix socket（位於 `$TMPDIR`、隨機路徑、0600 權限）作為 IPC 通道。LaunchPal 退出時 helper 自動結束，機器上不留任何常駐 root 行程或 `/Library/LaunchDaemons/` 檔案。

這讓使用者**每個 LaunchPal session 只需輸入一次密碼 / Touch ID**，之後所有 system daemon 寫入操作都不需再提權，同時避免「整個 Wails app 以 root 執行」造成的巨大攻擊面。

## What Changes

- 新增 Go binary `launchpal-privhelper`：純 CLI、無 UI、以 root 執行、透過 stdin/stdout 或 socket 接收 RPC
- LaunchPal app 啟動時預設 Admin Mode 關閉；新增「Enable Admin Mode」UI 控制項（Settings 頁）
- 啟用 Admin Mode 的流程：
  1. 在 `$TMPDIR/launchpal-<uid>-<rand>.sock` 預備 socket 路徑
  2. 透過 `osascript` 以 admin 權限執行 `launchpal-privhelper --socket <path>` 為背景行程
  3. helper 建立 socket 並 `chmod 0600`、用 `SO_PEERCRED` 驗連線端 UID 等於啟動 UID
  4. LaunchPal 連上 socket，發送 handshake RPC 確認 helper 就緒
- RPC 協議（JSON over newline-delimited messages）：`ListSystemDaemons` / `GetSystemDaemon` / `Bootstrap` / `Bootout` / `Kickstart` / `WritePlist` / `DeletePlist` / `Shutdown`
- `SystemManager` 新增 privileged-aware 模式：當 Admin Mode 啟用且 helper 就緒時，寫入操作走 helper；否則回既有 `ErrReadOnlyManager`
- `/System/Library/LaunchDaemons` 仍維持唯讀（不應修改 Apple 系統服務），僅 `/Library/LaunchDaemons` 在 Admin Mode 下可寫
- Helper 寫入 plist 前呼叫既有 `BackupManager` 做備份（沿用 Phase 1 之前既有行為）— 需讓 BackupManager 支援 system path 或於 helper 端實作等價備份
- LaunchPal 退出或使用者明確 Disable Admin Mode 時，送 Shutdown RPC 讓 helper 自行結束；helper 也 watchdog parent PID，parent 死了就 self-exit
- 前端 Settings 頁顯示 Admin Mode 狀態（OFF / Requesting / ON / Error）、授權入口、停用按鈕
- System 列表與詳情頁：Admin Mode ON 時顯示 Start / Stop / Restart / Edit / Delete / Create 按鈕；OFF 時顯示鎖頭圖示與「Enable Admin Mode to manage」提示

## Non-Goals

- 不取得 Apple Developer ID、不走 `SMAppService` / `SMJobBless`
- 不在機器上留下任何常駐 root daemon 或 LaunchDaemon plist 檔
- 不做持久授權快取（跨 session 免密碼）；每次啟動 LaunchPal 要重新授權
- 不開放對 `/System/Library/LaunchDaemons` 的寫入（Apple 系統服務仍唯讀）
- 不實作 Phase 2a 的 per-operation osascript 提權（相斥的設計）
- 不提供「整個 app 以 root 重啟」模式

## Capabilities

### New Capabilities

- `privileged-helper-lifecycle`: helper 子行程的啟動、authorization、socket handshake、parent watchdog、關閉流程
- `privileged-helper-rpc`: JSON-over-socket RPC 協議與認證（SO_PEERCRED）
- `admin-mode`: LaunchPal 的 Admin Mode 狀態管理、UI、錯誤處理
- `system-daemon-write-ops`: 透過 helper 執行 Bootstrap / Bootout / Kickstart / WritePlist / DeletePlist 等 system daemon 寫入操作

### Modified Capabilities

- `launchdaemons-readonly`: `SystemManager` 的「讀寫」語意由「一律唯讀」改為「預設唯讀、Admin Mode 下可寫（僅 `/Library/LaunchDaemons`）」；`AppleSystemManager` 保持純唯讀不變

## Impact

- Affected specs:
  - 新增 `openspec/specs/privileged-helper-lifecycle/spec.md`
  - 新增 `openspec/specs/privileged-helper-rpc/spec.md`
  - 新增 `openspec/specs/admin-mode/spec.md`
  - 新增 `openspec/specs/system-daemon-write-ops/spec.md`
  - 修改 `openspec/specs/launchdaemons-readonly/spec.md`
- Affected code:
  - 新增 `cmd/launchpal-privhelper/main.go`：helper binary 入口
  - 新增 `internal/privhelper/`：helper 端與 client 端共用的 RPC 協議、socket server / client
  - 修改 `internal/launchctl/system.go`：依 Admin Mode 狀態決定走原 `ErrReadOnlyManager` 或透過 helper RPC
  - 修改 `internal/launchctl/manager.go`：`Manager` interface 的語意文件更新（或新增 `AdminCapable` 子 interface）
  - 修改 `internal/backup/backup.go`：讓備份邏輯可被 helper 以 root 身份呼叫（或在 helper 端重實作最小集）
  - 修改 `app.go`：新增 `EnableAdminMode() / DisableAdminMode() / GetAdminModeStatus()` Wails bindings，並讓既有 write-op bindings 在 system type 下走 helper
  - 新增 `frontend/app/pages/settings.vue` 的 Admin Mode 區塊
  - 修改 `frontend/app/pages/system.vue`、`frontend/app/pages/services/[name].vue`：Admin Mode 條件渲染
  - 修改 `Makefile`：新增 helper binary 的 build target 與打包進 app bundle
  - 修改 `build/darwin/`：helper binary 的 bundle 位置（`Contents/MacOS/launchpal-privhelper`）
  - 修改 `wails.json`：若影響 build hook
- 相依工具：需 `osascript`、`launchctl`、Unix domain socket 支援（皆為 macOS 內建）
- 前置依賴：`system-daemon-status-detection` change 需已 archived 或至少已實作 `StatusConfidence` 與偵測邏輯
