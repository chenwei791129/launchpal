## 1. 後端：排程型別結構設計

- [x] [P] 1.1 擴充 `ScheduleConfig` 加入 `Interval *int` 欄位，plistData 結構擴充加入 `StartInterval` 欄位（排程型別結構設計）
- [x] [P] 1.2 更新 `writePlist` 支援 Backend StartInterval support — 當 `Interval` 有值時寫入 `StartInterval`，否則寫入 `StartCalendarInterval`（StartInterval and CalendarInterval are mutually derived）
- [x] [P] 1.3 更新 `parseSchedule` 讀取 `StartInterval` 並填入 `ScheduleConfig.Interval`（Read existing service with StartInterval）

## 2. 前端：TypeScript 型別更新

- [x] [P] 2.1 更新 `frontend/app/types/wails.d.ts` 中 `ScheduleConfig` 加入 `interval?: number` 欄位

## 3. 前端：排程表單元件

- [x] 3.1 建立 `ScheduleForm.vue` 共用元件，實作前端排程 UI 設計及 Schedule type selection in New Service UI（編輯頁排程支援）
- [x] 3.2 在 `ScheduleForm.vue` 實作 Calendar Interval empty field validation 警告提示

## 4. 前端：新增服務整合

- [x] 4.1 在 `CreateServiceModal.vue` 整合 `ScheduleForm.vue`，支援 Schedule coexists with RunAtLoad

## 5. 前端：服務顯示與編輯

- [x] [P] 5.1 在 `ServiceSummary.vue` 實作 Schedule information display in ServiceSummary（顯示 CalendarInterval 和 StartInterval）
- [x] 5.2 在 `services/[name].vue` 整合 `ScheduleForm.vue`，實作 Schedule editing in service detail page（Add/Edit/Remove schedule）

## 6. 測試

- [x] [P] 6.1 後端測試：`writePlist` 寫入 `StartInterval` 的測試（Create service with StartInterval）
- [x] [P] 6.2 後端測試：`parseSchedule` 讀取 `StartInterval` 的測試
- [x] [P] 6.3 後端測試：驗證 `StartInterval` 和 `StartCalendarInterval` 互斥寫入

## 7. 後續修正

- [x] 7.1 Calendar Interval UI 改為 cron 表達式輸入（`minute hour day month weekday` 格式），取代 5 個 dropdown
- [x] 7.2 修正 cron parse 錯誤時 emit undefined 導致 Enable Schedule 被取消勾選的 bug
- [x] 7.3 為 `loaded` 狀態（已載入等待觸發）加入藍色燈號，區分 stopped（灰色）
- [x] 7.4 修正 loaded 狀態的服務顯示 Start 按鈕的問題，改為顯示 Stop 按鈕（unload）
