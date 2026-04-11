## Why

目前 LaunchPal 的 Calendar Interval 僅支援單一值（如 `0 9 * * *`），無法方便地表達「每天 9:00 到 17:00 每小時執行」這類常見需求。使用者必須手動編輯 plist 才能建立多筆 `StartCalendarInterval`。

## What Changes

- 擴展 cron 語法解析器，在現有 5 欄位（minute hour day month weekday）中支援**範圍 `a-b`** 和**列舉 `a,b,c`** 語法
- 多欄位使用範圍/列舉時，自動做**笛卡爾積展開**，產生多筆 `StartCalendarInterval` dict
- 後端 `ScheduleConfig` 從單筆改為支援**多筆** calendar interval 的讀寫
- 前端預覽改為**摘要 + 可展開完整列表**，讓使用者確認展開結果
- 移除 `HasMultiple` workaround，多筆 interval 成為 first-class 功能

## Non-Goals

- 不支援步進語法 `*/n`（`StartInterval` 已覆蓋此需求）
- 不支援混合語法（如 `1-3,7`），僅支援純範圍或純列舉
- 不重新設計排程 UI 佈局，保持在現有 Calendar Interval tab 內

## Capabilities

### New Capabilities

- `cron-range-expansion`: Cron 語法的範圍 `a-b` 與列舉 `a,b,c` 解析、笛卡爾積展開、以及展開結果預覽

### Modified Capabilities

- `scheduled-service`: `ScheduleConfig` 從單筆擴展為多筆 calendar interval 的讀寫，移除 `HasMultiple` flag

## Impact

- 受影響的 specs：`cron-range-expansion`（新增）、`scheduled-service`（修改）
- 受影響的程式碼：
  - `internal/launchctl/types.go` — `ScheduleConfig` 結構調整
  - `internal/launchctl/user.go` — `parseSchedule` 讀取多筆、`writePlist` 寫入 array
  - `frontend/app/components/ScheduleForm.vue` — cron 解析器擴展、預覽 UI
  - `frontend/app/components/ServiceSummary.vue` — 多筆排程顯示
  - `frontend/app/types/wails.d.ts` — TypeScript 型別更新
  - `frontend/wailsjs/go/models.ts` — Wails 自動生成型別
  - `internal/launchctl/user_test.go` — 測試更新
