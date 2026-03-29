## Why

後端的 `ServiceConfig` 已完整支援 `EnvironmentVariables` 的讀寫（`map[string]string`），但新增服務的 UI（`CreateServiceModal.vue`）沒有暴露這個欄位。使用者若需要為服務設定環境變數（如 API keys、資料庫 URL），必須手動編輯 plist 檔案。由於 launchd 不會載入 shell profiles，環境變數設定對許多服務來說是必要的。

來源：GitHub Issue #5。

## What Changes

- 在 `CreateServiceModal.vue` 的表單中新增「Environment Variables」區塊
- 提供動態 key-value 列表 UI，支援新增與刪除環境變數
- 提交時將 key-value pairs 轉為 `Record<string, string>` 傳入 `ServiceConfig.environment`
- 後端零改動——既有的 `writePlist` 已處理 `EnvironmentVariables` 寫入

## Non-Goals

- 不修改後端 Go 程式碼（已完整支援）
- 不在編輯服務頁面加入環境變數（可作為後續工作）
- 不提供環境變數的匯入/匯出功能
- 不驗證環境變數值的格式（使用者可自行填入任意字串）

## Capabilities

### New Capabilities

- `env-vars-ui`: 新增服務表單中的環境變數設定介面，提供動態 key-value 列表的新增、刪除操作

### Modified Capabilities

（無）

## Impact

- 受影響的程式碼：`frontend/app/components/CreateServiceModal.vue`
- 受影響的型別：無（`ServiceConfig` 已包含 `environment` 欄位）
- 後端 API：無改動（`CreateService` 已接受 `environment` 參數）
