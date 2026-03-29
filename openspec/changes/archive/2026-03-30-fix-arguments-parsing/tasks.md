## 1. 建立共用 Utility

- [x] 1.1 [P] 建立 `frontend/app/utils/shell-args.ts`，實作 `parseShellArgs(input: string): string[]` 函數，支援 parse quoted arguments from text input（單引號與雙引號），未引號 token 以空白分割，空輸入回傳空陣列
- [x] 1.2 [P] 在同一檔案實作 `serializeShellArgs(args: string[]): string` 函數，實作 serialize argument array to text with quoting — 含空格的 argument 用單引號包裹，無空格的保持原樣

## 2. Consistent parsing across create and edit flows

- [x] 2.1 [P] 修改 `frontend/app/pages/services/[name].vue`：將 `editArgumentsText.value.split(/\s+/).filter(Boolean)` 替換為 `parseShellArgs(editArgumentsText.value)`，確保 consistent parsing across create and edit flows
- [x] 2.2 [P] 修改 `frontend/app/components/CreateServiceModal.vue`：將 `argumentsText.value.split(/\s+/).filter(Boolean)` 替換為 `parseShellArgs(argumentsText.value)`

## 3. 替換序列化邏輯

- [x] 3.1 [P] 修改 `frontend/app/pages/services/[name].vue` 的 `populateEditForm()`：將 `service.value.arguments?.join(' ')` 替換為 `serializeShellArgs(service.value.arguments ?? [])`
- [x] 3.2 [P] 修改 `frontend/app/components/ServiceSummary.vue`：將 `service.arguments?.join(' ')` 替換為 `serializeShellArgs(service.arguments ?? [])`

## 4. 測試驗證

- [x] 4.1 建立 `frontend/app/utils/__tests__/shell-args.test.ts`，覆蓋 specs 中所有 scenario：簡單無引號、單引號含空格、雙引號含空格、混合引號、空輸入、round-trip 一致性
- [x] 4.2 執行測試確認全部通過
