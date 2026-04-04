## Why

目前 LaunchPal 僅支援啟動（Start）和停止（Stop）服務，缺少「立即執行一次」的能力。對於排程服務或需要臨時測試的場景，用戶必須等待排程觸發或手動停止再啟動，操作不直覺。macOS 的 `launchctl kickstart` 指令提供了精確的「立即執行」語意，適合作為此功能的基礎。

## What Changes

- 後端新增 `Kickstart` 方法於 `UserManager`，使用 `launchctl kickstart -k gui/{UID}/{label}` 實作
- 不修改 `Manager` interface（kickstart 僅適用於 user services）
- 若服務尚未 loaded，自動先執行 bootstrap 再 kickstart
- 前端在服務詳情頁 header 按鈕列新增「Run Now」按鈕
- 當服務處於 running 狀態時，點擊「Run Now」跳出確認對話框，提醒用戶此操作會終止現有 process 後重新執行
- 服務非 running 狀態時直接執行，不顯示確認

## Non-Goals

- 不在 ServiceRow 列表頁加 Run Now 按鈕（空間有限，屬進階操作）
- 不對 system/apple-system 服務提供 kickstart（唯讀模式）
- 不提供 kickstart 不帶 `-k` 的選項（語意不夠明確，可能導致重複 process）

## Capabilities

### New Capabilities

- `kickstart-service`: 立即執行（kickstart）user service 的能力，包含前端確認流程

### Modified Capabilities

（無）

## Impact

- 受影響程式碼：
  - `internal/launchctl/user.go` — 新增 `Kickstart` 方法
  - `app.go` — 新增 `KickstartService` binding
  - `frontend/app/pages/services/[name].vue` — 新增 Run Now 按鈕與確認對話框
- 受影響 specs：新增 `kickstart-service` spec
