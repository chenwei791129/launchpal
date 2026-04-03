## 1. 核心計算邏輯

- [x] 1.1 [P] 新增 `frontend/app/composables/useNextOccurrences.ts` composable，實作 "Calculate next occurrences from CalendarInterval" — 接受 `ScheduleConfig`（含 minute、hour、day、weekday、month 欄位）與 count 參數，從 now+1 分鐘開始逐分鐘掃描，回傳下次 N 個 `Date` 物件陣列。wildcard 欄位（undefined）匹配任意值。掃描上限 400 天。

## 2. 格式化工具

- [x] 2.1 [P] 在 `useNextOccurrences.ts` 中新增 `formatDateTime` 函式，將 `Date` 格式化為 `M/D (Weekday) HH:mm`，weekday 使用英文縮寫（Sun, Mon, Tue, Wed, Thu, Fri, Sat）。

## 3. ScheduleForm 預覽

- [x] 3.1 修改 `frontend/app/components/ScheduleForm.vue`，在 Calendar Interval 模式且 cron 表達式合法時，顯示 "Preview next runs in ScheduleForm" 預覽區塊：呼叫 `useNextOccurrences` 取得下次 3 次執行時間，使用 `formatDateTime` 格式化顯示，並標示當前時區（`Intl.DateTimeFormat().resolvedOptions().timeZone`）。cron 表達式無效或為 Fixed Interval 模式時不顯示。

## 4. ServiceSummary 顯示

- [x] 4.1 修改 `frontend/app/components/ServiceSummary.vue`，實作 "Display next runs in ServiceSummary" — 當服務有 CalendarInterval 排程（`schedule` 存在且 `interval` 為 undefined）時，在排程描述下方顯示下次 3 次執行時間。StartInterval 或無排程的服務不顯示。
