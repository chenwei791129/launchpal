## Problem

服務詳細頁的 Summary tab 內容超出可視區域時無法捲動，底部的資訊（如 Environment Variables）被截斷而無法看到。

## Root Cause

Tab 內容的外層容器 `services/[name].vue:115` 使用了 `overflow-hidden`。Edit tab（第 146 行）和 Inspect tab（第 251 行）各自加了 `h-full overflow-auto`，但 Summary tab（第 143 行）直接渲染 `<ServiceSummary>` 元件，沒有包裹捲動容器，也沒有在元件內設定 overflow。

## Proposed Solution

為 `<ServiceSummary>` 元件加上 `h-full overflow-auto` 的外層包裹，與 Edit、Inspect tab 保持一致的捲動行為。

## Non-Goals

- 不重構其他 tab 的 overflow 處理方式
- 不改變 ServiceSummary 的內容排版

## Success Criteria

- Summary tab 在內容超出可視區域時出現垂直捲軸
- 可以捲動到頁面底部看到所有資訊（包含 Environment Variables）
- Edit 和 Inspect tab 的捲動行為不受影響

## Impact

- 影響的程式碼：`frontend/app/pages/services/[name].vue`（第 143 行，Summary tab 渲染處）
