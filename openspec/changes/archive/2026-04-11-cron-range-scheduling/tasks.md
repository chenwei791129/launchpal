## 1. 後端 ScheduleConfig 結構調整

- [x] 1.1 [P] 在 `internal/launchctl/types.go` 中新增 `CalendarEntry` struct（含 `Minute`, `Hour`, `Day`, `Weekday`, `Month` 各為 `*int`），修改 `ScheduleConfig` 將單筆欄位替換為 `Schedules []CalendarEntry`，移除 `HasMultiple` 欄位（對應 design 決定「ScheduleConfig 結構調整」）

## 2. 後端讀寫多筆 Calendar Interval

- [x] 2.1 修改 `internal/launchctl/user.go` 的 `parseSchedule` 函數，實作 Backend StartInterval support：讀取 plist 中的 `StartCalendarInterval`，無論是 dict 或 array 格式，統一解析為 `[]CalendarEntry` 填入 `ScheduleConfig.Schedules`
- [x] 2.2 修改 `internal/launchctl/user.go` 的 `writePlist` 函數：當 `Schedules` 為單筆時寫入 dict，多筆時寫入 array of dicts（對應 design 決定「Plist 寫入策略」）
- [x] 2.3 修改 `internal/launchctl/system.go` 和 `internal/launchctl/apple_system.go` 的 `parseSchedule` 呼叫，適配新的 `ScheduleConfig` 結構
- [x] 2.4 更新 `internal/launchctl/user_test.go` 測試：涵蓋單筆 dict 讀寫、多筆 array 讀寫、`StartInterval` 互斥寫入

## 3. 前端 Cron 解析器擴展（Cron field range syntax、Cron field enumeration syntax、Cartesian product expansion）

- [x] 3.1 [P] 實作 Cron field range syntax：重構 `frontend/app/components/ScheduleForm.vue` 中的 `parseCron()` 函數，每個欄位支援 `*`、單一值 `N`、範圍 `a-b`，回傳 `number[]`（對應 design 決定「Cron 解析器擴展策略」）
- [x] 3.2 [P] 實作 Cron field enumeration syntax：擴展 `parseCron()` 支援列舉 `a,b,c` 語法，含去重邏輯
- [x] 3.3 [P] 實作 Cartesian product expansion across multiple fields：將五個欄位的 `number[]` 做笛卡爾積展開，產生 `CalendarEntry[]`，並加入 Expansion count limit 檢查（上限 50 筆）

## 4. 前端預覽 UI（Expansion preview with summary and expandable list）

- [x] 4.1 實作 Expansion preview with summary and expandable list：修改 `frontend/app/components/ScheduleForm.vue` 的預覽區域，單筆時顯示現有文字描述，多筆時顯示摘要行（筆數 + 語意描述），點擊可展開/收合完整列表（對應 design 決定「展開結果預覽 UI」）

## 5. 前端資料流適配

- [x] 5.1 更新 `frontend/app/types/wails.d.ts` 中的 `ScheduleConfig` 型別：新增 `CalendarEntry` interface 和 `schedules` 欄位，移除單筆欄位和 `hasMultiple`
- [x] 5.2 更新 `frontend/app/components/ScheduleForm.vue` 的 `watch` 和 `emit` 邏輯：將解析結果轉為 `ScheduleConfig.schedules` 陣列送出；從 `modelValue` 初始化時，將 `schedules` 還原為 cron 表達式
- [x] 5.3 更新 `frontend/app/components/CreateServiceModal.vue`，適配新的 `ScheduleConfig` 結構

## 6. 顯示層更新（Schedule information display in ServiceSummary）

- [x] 6.1 [P] 實作 Schedule information display in ServiceSummary：更新 `frontend/app/components/ServiceSummary.vue`，多筆 schedules 時顯示摘要描述，移除 `hasMultiple` 警告文字
- [x] 6.2 [P] 更新 `frontend/app/components/ServiceRow.vue`：適配 `schedules` 陣列的顯示邏輯
- [x] 6.3 [P] 更新 `frontend/app/pages/services/[name].vue`（Schedule editing in service detail page）：確保編輯頁面正確載入和儲存多筆 schedules

## 7. 型別同步與驗證

- [x] 7.1 執行 Wails 型別生成，確認 `frontend/wailsjs/go/models.ts` 正確反映新的 `CalendarEntry` 和 `ScheduleConfig` 結構
- [x] 7.2 執行 `go run` 驗證後端編譯通過，執行 `go test ./internal/launchctl/...` 確認測試通過
