## Why

環境變數的 value 目前在所有頁面（Summary、Edit、Create）以明碼顯示。若使用者存放 API key、密碼等敏感資訊，會有洩漏風險。需要預設遮蔽 value，並提供手動揭露的切換機制。

## What Changes

- 環境變數的 value 預設以遮蔽字元（`••••••••`）顯示
- 每個環境變數旁新增一個眼睛圖示按鈕，點擊後 toggle 明碼/遮蔽
- 適用於三個位置：ServiceSummary（唯讀）、服務編輯表單、服務建立 Modal

## Non-Goals

- 不加密儲存環境變數值（plist 中仍為明碼，僅 UI 層遮蔽）
- 不提供「全部顯示/隱藏」的全域按鈕（逐一切換即可）

## Capabilities

### New Capabilities

（無）

### Modified Capabilities

- `env-vars-ui`: 新增環境變數值遮蔽顯示與 toggle 揭露的 requirement

## Impact

- 影響的程式碼：
  - `frontend/app/components/ServiceSummary.vue`（唯讀摘要頁的環境變數顯示）
  - `frontend/app/pages/services/[name].vue`（編輯表單的環境變數 input）
  - `frontend/app/components/CreateServiceModal.vue`（建立表單的環境變數 input）
