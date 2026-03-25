## Why

目前「新增服務」功能只支援 `RunAtLoad` 類型。需要建立排程服務的使用者必須手動編輯 plist，這違背了 GUI 管理工具的初衷。加入 Scheduled 類型支援能讓 LaunchPal 成為更完整的 macOS 服務管理工具。（對應 GitHub Issue #4）

## What Changes

- 後端新增 `StartInterval`（固定間隔秒數）支援，目前僅支援 `StartCalendarInterval`
- 前端 New Service 表單加入服務類型選擇（RunAtLoad / Scheduled）
- 選擇 Scheduled 時顯示排程子類型切換（CalendarInterval / Interval）及對應設定欄位
- `RunAtLoad` 和 Schedule 允許同時啟用
- CalendarInterval 和 Interval 在 UI 上為二選一
- 服務詳情頁（Summary）顯示排程資訊
- 服務編輯頁同步支援排程設定

## Capabilities

### New Capabilities

- `scheduled-service`: 排程服務的建立、編輯、顯示功能，涵蓋 `StartCalendarInterval` 和 `StartInterval` 兩種排程機制

### Modified Capabilities

（無）

## Impact

- 受影響的後端程式碼：
  - `internal/launchctl/types.go` — 擴充 `ScheduleConfig` 型別，加入 `StartInterval`
  - `internal/launchctl/user.go` — `plistData` 結構、`writePlist`、`parseSchedule` 需支援 `StartInterval`
- 受影響的前端程式碼：
  - `frontend/app/components/CreateServiceModal.vue` — 加入排程 UI
  - `frontend/app/components/ServiceSummary.vue` — 顯示排程資訊
  - `frontend/app/pages/services/[name].vue` — 編輯頁支援排程
  - `frontend/app/types/wails.d.ts` — 更新 TypeScript 型別
