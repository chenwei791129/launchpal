## Context

LaunchPal 是 macOS LaunchAgent 圖形化管理工具（Go + Wails v2 + Nuxt 4）。目前 New Service 僅支援 `RunAtLoad` 類型。後端已部分支援 `StartCalendarInterval`（讀取、寫入），但缺少 `StartInterval` 支援，且前端完全沒有排程 UI。

現有相關程式碼：
- `ServiceConfig.Schedule *ScheduleConfig` — 僅有 CalendarInterval 欄位
- `plistData.StartCalendarInterval` — 已能解析
- `writePlist()` — 已能寫入 `StartCalendarInterval`
- `CreateServiceModal.vue` — 無排程欄位

## Goals / Non-Goals

**Goals:**

- 在 New Service 表單中加入排程類型選擇
- 支援 `StartCalendarInterval`（日曆排程）和 `StartInterval`（固定間隔）
- 允許 `RunAtLoad` 和排程同時啟用
- 在 ServiceSummary 顯示排程資訊
- 在服務編輯頁支援排程修改

**Non-Goals:**

- 不支援 `StartCalendarInterval` 的陣列格式（多組排程）
- 不顯示「下次執行時間」預測（可做為後續功能）
- 不加入 `EnvironmentVariables` UI（已拆為 Issue #5）

## Decisions

### 排程型別結構設計

擴充 `ScheduleConfig` 加入 `Interval` 欄位，而非拆分為兩個獨立型別。

```go
type ScheduleConfig struct {
    Minute   *int `json:"minute,omitempty"`
    Hour     *int `json:"hour,omitempty"`
    Day      *int `json:"day,omitempty"`
    Weekday  *int `json:"weekday,omitempty"`
    Month    *int `json:"month,omitempty"`
    Interval *int `json:"interval,omitempty"`
}
```

**理由**：保持 API 表面簡單，前端透過 `interval` 是否有值來判斷排程子類型。若 `interval` 有值則為 StartInterval，否則為 StartCalendarInterval。兩者互斥由前端 UI 控制。

**替代方案**：新增 `ScheduleType` 欄位明確標示類型。不採用，因為可從欄位值推斷，額外欄位是冗餘。

### 前端排程 UI 設計

在 `CreateServiceModal.vue` 加入排程區塊：

1. 新增 "Enable Schedule" checkbox
2. 啟用後顯示子類型切換：Calendar Interval / Fixed Interval
3. Calendar Interval：顯示 Minute、Hour、Day、Weekday、Month 五個可選下拉欄位
4. Fixed Interval：顯示一個數字輸入框（單位：秒）

**理由**：checkbox 而非 radio 選擇器，因為 `RunAtLoad` 和排程可共存。子類型用 tab 或 radio 切換即可。

### plistData 結構擴充

在 `plistData` 加入 `StartInterval` 欄位：

```go
type plistData struct {
    // ...existing fields...
    StartInterval int `plist:"StartInterval"`
}
```

`parseSchedule` 需同時檢查 `StartCalendarInterval` 和 `StartInterval`，將結果統一填入 `ScheduleConfig`。

### 編輯頁排程支援

服務詳情頁 `services/[name].vue` 的編輯模式需複用與 CreateServiceModal 相同的排程 UI 元件。考慮將排程表單抽取為獨立元件 `ScheduleForm.vue` 以避免重複。

## Risks / Trade-offs

- **[CalendarInterval 全空驗證]** → CalendarInterval 五個欄位全不填等同「每分鐘執行」。前端需提示使用者至少填一個欄位，或明確顯示「every minute」的含義。
- **[StartInterval 最小值]** → launchd 的 StartInterval 最小有效值為 10 秒。前端需設定合理的 min 值。
- **[既有排程服務的解析]** → 部分現有服務可能同時使用 `StartCalendarInterval` 陣列格式，`parseSchedule` 目前已處理此情況但僅取第一個。Summary 顯示時需註明此限制。
