## Why

macOS launchd 提供 `WakeSystem` 屬性，可在排程時間到達時將休眠中的 Mac 喚醒執行任務。這對凌晨備份、定時同步等場景至關重要。目前 LaunchPal 不支援讀取或設定此屬性，使用者無法透過 GUI 管理此功能。

## What Changes

- 後端 `Service` 和 `ServiceConfig` 新增 `WakeSystem` 布林欄位
- `plistData` 新增 `WakeSystem` 欄位，支援讀取現有 plist 中的設定
- `writePlist` 寫入時，當 `WakeSystem` 為 true 時輸出對應 key
- 三種服務管理器（User、System、Apple System）皆支援解析 `WakeSystem`
- 排程表單（`ScheduleForm.vue`）新增 WakeSystem toggle 開關
- 服務摘要（`ServiceSummary.vue`）顯示 WakeSystem 狀態

## Capabilities

### New Capabilities

- `wake-system`: 支援 launchd plist 的 `WakeSystem` 屬性讀寫與 UI 管理

### Modified Capabilities

- `scheduled-service`: 排程表單新增 WakeSystem toggle，排程相關 UI 整合此選項

## Impact

- Affected specs: `wake-system`（新增）、`scheduled-service`（修改）
- Affected code:
  - `internal/launchctl/types.go` — Service、ServiceConfig struct
  - `internal/launchctl/user.go` — plistData、writePlist、parseSchedule 周邊
  - `internal/launchctl/system.go` — 解析 WakeSystem
  - `internal/launchctl/apple_system.go` — 解析 WakeSystem
  - `frontend/app/components/ScheduleForm.vue` — 新增 toggle
  - `frontend/app/components/ServiceSummary.vue` — 顯示狀態
  - `frontend/app/types/wails.d.ts` — TypeScript 型別
  - `frontend/wailsjs/go/models.ts` — Wails 自動生成型別
