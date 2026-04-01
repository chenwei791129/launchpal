## Why

`docs/plans/` 下有三份設計文件記錄了 LaunchPal 的核心架構與 LaunchDaemons 唯讀支援，但這些功能從未正式轉為 spec。同時，核心模組（backup、UserManager.Get()、SystemManager/AppleSystemManager 的讀取操作）的測試覆蓋率極低（backup ~10%、app ~5%）。將 plan 轉為 spec 並補上測試，能確保功能規格有跡可查、既有行為有測試保護。

## What Changes

- 新增 `core-service-management` spec：從 `2025-01-25-launchpal-design.md` 擷取核心服務管理的行為規格（列表、控制、CRUD、plist 讀寫、日誌、備份）
- 新增 `launchdaemons-readonly` spec：從 `2026-01-28-launchdaemons-readonly-support-design.md` 擷取唯讀模式的行為規格（System/AppleSystem 服務的讀取、寫入拒絕、binary plist 轉換）
- 補充 `internal/backup/backup_test.go`：Create、List、GetContent、Restore、pruneBackups
- 補充 `internal/launchctl/user_test.go`：Get() 方法的 plist 解析測試
- 補充 `internal/launchctl/system_test.go`：Get()、GetPlist() 測試
- 補充 `internal/launchctl/apple_system_test.go`：Get()、GetPlist() 測試

## Non-Goals

- **不測試需要 launchctl 的方法**：Start、Stop、Restart 涉及系統指令執行，需重構 command executor 才能測試，不在此次範圍
- **不轉換 implementation.md**：`2025-01-25-launchpal-implementation.md` 是實作步驟紀錄，不適合轉為 spec
- **不重構現有程式碼**：僅新增 spec 和測試，不修改既有實作

## Capabilities

### New Capabilities

- `core-service-management`: 核心服務管理行為規格 — 涵蓋 User LaunchAgent 的列表、狀態檢測、啟停控制、CRUD、plist 讀寫、日誌讀取、備份與還原
- `launchdaemons-readonly`: LaunchDaemons 唯讀模式規格 — 涵蓋 System/AppleSystem 服務的讀取操作、寫入拒絕行為、binary plist 自動轉換

### Modified Capabilities

(none)

## Impact

- Affected specs: `core-service-management`（新增）、`launchdaemons-readonly`（新增）
- Affected code:
  - `internal/backup/backup_test.go` — 大幅擴充測試
  - `internal/launchctl/user_test.go` — 新增 Get() 測試
  - `internal/launchctl/system_test.go` — 新增 Get()、GetPlist() 測試
  - `internal/launchctl/apple_system_test.go` — 新增 Get()、GetPlist() 測試
