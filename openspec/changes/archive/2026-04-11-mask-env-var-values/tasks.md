## 1. ServiceSummary 遮蔽顯示

- [x] [P] 1.1 在 `ServiceSummary.vue` 中新增 `revealedEnvKeys` ref（`Set<string>` 型別）追蹤已揭露的環境變數 key，將 `{{ key }}={{ value }}` 改為預設顯示 `{{ key }}=••••••••`，當 key 存在於 `revealedEnvKeys` 時顯示明碼。每行旁邊新增 inline SVG 眼睛圖示按鈕（`w-4 h-4`，與現有按鈕風格一致），點擊時 toggle `revealedEnvKeys` 中的 key。（Environment variable values are masked by default）

## 2. 編輯表單遮蔽顯示

- [x] [P] 2.1 在 `services/[name].vue` 的 Edit tab 中，將環境變數 value 的 `<input type="text">` 改為 `type="password"` 預設遮蔽，新增 `editEnvVisibility` reactive `Set<number>`（以 index 追蹤），每個 input 旁新增 inline SVG 眼睛圖示按鈕，點擊時 toggle 該 index 的 input type 為 `text`/`password`。（Environment variable values are masked by default）

## 3. 建立表單遮蔽顯示

- [x] [P] 3.1 在 `CreateServiceModal.vue` 中，將環境變數 value 的 `<input type="text">` 改為 `type="password"` 預設遮蔽，新增 `envVisibility` reactive `Set<number>`（以 index 追蹤），每個 input 旁新增 inline SVG 眼睛圖示按鈕，點擊時 toggle 該 index 的 input type 為 `text`/`password`。（Environment variable values are masked by default）
