## 1. 環境變數 key-value 列表 UI

- [x] 1.1 在 `CreateServiceModal.vue` 的 `form` reactive 物件中新增 `envVars` 陣列（型別 `Array<{key: string, value: string}>`），初始為空陣列
- [x] 1.2 在 Checkboxes 與 ScheduleForm 之間新增「Environment Variables」區塊，包含：標題、動態 key-value 行列表、以及底部的「Add」按鈕。實作 environment variables key-value list in create service form 規格
- [x] 1.3 每一行包含 key 輸入框（placeholder: `KEY`）、value 輸入框（placeholder: `Value`）、以及刪除按鈕（移除該行）。實作 add/remove environment variable entry 規格

## 2. 表單提交與重置

- [x] 2.1 修改 `handleSubmit` 函數，將 `envVars` 陣列過濾空 key 後轉為 `Record<string, string>`，寫入 `config.environment`。實作 environment variables included in service creation payload 規格（含 empty key entries filtered out）
- [x] 2.2 在表單重置區塊中加入 `envVars` 的清空邏輯。實作 form reset clears environment variables 規格
