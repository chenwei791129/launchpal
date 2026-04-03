## Why

使用者設定 `StartCalendarInterval` 排程後，無法直觀知道服務下次何時執行。目前 ServiceSummary 只顯示排程描述（如「Every day at 03:00」），但不顯示具體的下次執行日期時間。在前端純計算下次執行時間，幫助使用者驗證排程設定是否正確。

## What Changes

- 新增前端 composable `useNextOccurrences`，根據 `CalendarInterval` 欄位計算下次 N 次執行時間
- 在 `ScheduleForm.vue` 的 Calendar Interval 模式下，即時預覽下次 3 次執行時間
- 在 `ServiceSummary.vue` 的排程區塊中，顯示下次 3 次執行時間
- 僅限 `CalendarInterval` 排程；`StartInterval` 不提供預覽（因 launchd 的實際觸發時間受系統狀態影響，且無法取得 load time）

## Non-Goals

- **不支援 `StartInterval` 預覽**：macOS launchctl 不暴露服務的 load time 或 last run time，`now + interval * n` 的估算不準確，容易誤導使用者
- **不在後端計算**：CalendarInterval 的下次執行時間可純前端計算，不需要新增 Wails binding
- **不支援多 CalendarInterval 的合併預覽**：目前 `ScheduleConfig` 只存單一 CalendarInterval（`hasMultiple` flag 標示多排程），暫不改變此結構

## Capabilities

### New Capabilities

- `next-run-preview`: 根據 CalendarInterval 排程設定，在前端計算並顯示下次執行時間的預覽功能

### Modified Capabilities

（無）

## Impact

- 受影響的 spec：`scheduled-service`（現有排程顯示行為不變，新增預覽為獨立能力）
- 受影響的程式碼：
  - `frontend/app/composables/` — 新增 `useNextOccurrences` composable
  - `frontend/app/components/ScheduleForm.vue` — 新增即時預覽區塊
  - `frontend/app/components/ServiceSummary.vue` — 新增下次執行時間顯示
