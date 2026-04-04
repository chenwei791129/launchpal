## Why

macOS `man launchctl` 已將 `load`/`unload` 歸類為 **LEGACY SUBCOMMANDS**，並明確建議改用 `bootstrap`/`bootout`。Legacy 指令存在一個關鍵缺陷：無論操作是否實際成功，`load`/`unload` 幾乎永遠回傳 exit code 0，導致 LaunchPal 無法正確偵測啟動或停止是否失敗。macOS Ventura (13.x) 已有使用者回報 `unload` 出現 I/O error，顯示 Apple 正在逐步淘汰 legacy 支援。本次僅遷移 `load`/`unload`，不涉及 `list` 指令。

## What Changes

- 將 `Start()` 中的 `launchctl load <path>` 替換為 `launchctl bootstrap gui/<uid> <path>`
- 將 `Stop()` 中的 `launchctl unload <path>` 替換為 `launchctl bootout gui/<uid>/<label>`
- 新增取得當前使用者 UID 的輔助邏輯（`os.Getuid()`）
- 調整 `Stop()` 的錯誤處理：`bootout` 在服務未載入時會回傳非零 exit code，需容忍此情況（目前 `unload` 的錯誤已被忽略）

## Non-Goals

- 不遷移 `launchctl list`（用於查詢服務狀態），留待後續 Phase 2 評估
- 不遷移 `launchctl start`/`stop`（legacy 的 label-based 啟停），目前未使用
- 不改變 System / Apple System 管理器的行為（它們為唯讀，不呼叫 load/unload）

## Capabilities

### New Capabilities

（無）

### Modified Capabilities

- `core-service-management`: Start/Stop 的底層 launchctl 指令從 legacy `load`/`unload` 遷移至 `bootstrap`/`bootout`，錯誤偵測能力提升

## Impact

- Affected specs: `core-service-management`（修改）
- Affected code:
  - `internal/launchctl/user.go` — `Start()`、`Stop()`、`Restart()` 方法
  - `internal/launchctl/user_test.go` — 相關測試可能需要調整
