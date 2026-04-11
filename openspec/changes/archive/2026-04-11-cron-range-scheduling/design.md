## Context

LaunchPal 目前的 Calendar Interval 排程使用簡化的 cron 語法（`minute hour day month weekday`），每個欄位只接受 `*` 或單一數值。後端 `ScheduleConfig` 也只儲存單筆 calendar interval。

當 plist 中存在多筆 `StartCalendarInterval`（array 格式），後端能讀取但只取第一筆，並用 `HasMultiple` flag 在前端顯示警告。使用者無法透過 UI 建立或編輯多筆 calendar interval。

## Goals / Non-Goals

**Goals:**

- 讓使用者在現有 cron 輸入框中使用範圍 `a-b` 和列舉 `a,b,c` 語法
- 多欄位使用範圍/列舉時，自動做笛卡爾積展開為多筆 `StartCalendarInterval`
- 提供展開結果預覽（摘要 + 可展開列表）
- 後端完整支援多筆 calendar interval 的讀寫

**Non-Goals:**

- 不支援步進語法 `*/n`（`StartInterval` 已覆蓋定時重複的需求）
- 不支援混合語法如 `1-3,7`（範圍與列舉不混用於同一欄位）
- 不重新設計排程 UI 佈局

## Decisions

### Cron 解析器擴展策略

在現有 `parseCron()` 函數中擴展每個欄位的解析邏輯，支援三種格式：

1. `*` — 任意值（現有）
2. `N` — 單一數值（現有）
3. `N-M` — 範圍，展開為 N, N+1, ..., M
4. `N,M,O` — 列舉，直接使用列出的值

每個欄位解析後回傳 `number[]` 而非 `number | undefined`。所有欄位的結果做笛卡爾積，產生 `ParsedCron[]` 陣列。

**替代方案**：新增獨立的 range picker UI 元件。放棄此方案因為 cron 語法對開發者更熟悉且不需要額外 UI 空間。

### ScheduleConfig 結構調整

將 Go 後端的 `ScheduleConfig` 新增 `Schedules` 欄位（`[]CalendarEntry`），取代原本的單筆欄位：

```
type CalendarEntry struct {
    Minute  *int `json:"minute,omitempty"`
    Hour    *int `json:"hour,omitempty"`
    Day     *int `json:"day,omitempty"`
    Weekday *int `json:"weekday,omitempty"`
    Month   *int `json:"month,omitempty"`
}

type ScheduleConfig struct {
    Schedules []CalendarEntry `json:"schedules,omitempty"`
    Interval  *int            `json:"interval,omitempty"`
}
```

移除 `HasMultiple` flag 和原本的 `Minute/Hour/Day/Weekday/Month` 頂層欄位。

**替代方案**：保留單筆欄位並另加陣列欄位。放棄此方案因為會造成兩種表示法並存的混亂。

### 展開結果預覽 UI

預覽區分為兩層：
1. **摘要行**：顯示展開的筆數和語意描述（如「9 schedules: hour 9-17, at minute 00」）
2. **展開列表**：點擊摘要可展開/收合完整列表，每筆顯示為可讀的時間描述

當展開超過 50 筆時，顯示警告提示使用者確認。

### Plist 寫入策略

- 單筆時：寫入 `StartCalendarInterval` 為 dict（與現有行為一致）
- 多筆時：寫入 `StartCalendarInterval` 為 array of dicts
- 讀取時：統一解析為 `[]CalendarEntry`，不論 plist 中是 dict 或 array

## Risks / Trade-offs

- **笛卡爾積爆炸** → 前端加入上限檢查（50 筆警告），避免使用者無意中產生大量排程條目
- **向後相容** → 移除 `HasMultiple` 和單筆欄位是 breaking change，但因為前端和後端是同一應用程式打包，不影響外部 API 相容性
- **Wails 型別同步** → `ScheduleConfig` 結構變更後需要重新生成 `frontend/wailsjs/go/models.ts`，確保前後端型別一致
