## Problem

Arguments 欄位使用簡單的空白分割（`split(/\s+/)`）來將使用者輸入的字串轉換為 `ProgramArguments` 陣列。這導致包含空格的參數（如用引號包裹的 prompt 字串）被錯誤地拆散成多個獨立的 argument。

例如，輸入：
```
--print -p 'run daily backup and send report'
```
會被錯誤拆分為 `["--print", "-p", "'run", "daily", "backup", "and", "send", "report'"]`，而正確結果應該是 `["--print", "-p", "run daily backup and send report"]`。

## Root Cause

三個位置都使用了不支援引號的簡單處理：

1. **編輯頁面**（`frontend/app/pages/services/[name].vue:434`）：`split(/\s+/).filter(Boolean)` — 存檔時將字串拆為陣列
2. **建立 Modal**（`frontend/app/components/CreateServiceModal.vue:207`）：同上
3. **反向顯示**（`services/[name].vue:407` 及 `ServiceSummary.vue:36`）：`join(' ')` — 從陣列還原為字串時，含空格的 argument 會失去引號邊界，導致再次編輯時無法正確還原

## Proposed Solution

實作 shell-like 的引號解析與反序列化：

1. **解析（字串 → 陣列）**：支援單引號 `'...'` 和雙引號 `"..."` 包裹含空格的參數，引號本身不納入結果
2. **序列化（陣列 → 字串）**：含空格的 argument 自動加上單引號顯示，確保編輯時可正確還原
3. 將解析邏輯抽為共用的 composable 或 utility，供建立與編輯頁面共用

## Non-Goals

- 不處理跳脫字元（如 `\'`、`\"`）— 一般 launchd 使用情境不需要
- 不修改後端 — 後端已正確處理 `[]string` 到 plist `ProgramArguments` 的轉換

## Success Criteria

- 包含 `-p '含空格的 prompt'` 的 arguments 能正確存入 plist 為獨立的 `<string>` 元素
- 從 plist 讀回後，含空格的 argument 在 UI 中以引號包裹顯示
- 編輯並重新存檔後，plist 內容不變（round-trip 一致性）
- 建立新服務和編輯既有服務都使用相同的解析邏輯

## Impact

- Affected code:
  - `frontend/app/pages/services/[name].vue` — 編輯表單的解析與顯示
  - `frontend/app/components/CreateServiceModal.vue` — 建立表單的解析
  - `frontend/app/components/ServiceSummary.vue` — Summary 頁面的顯示
  - 新增 utility 函數（如 `frontend/app/utils/shell-args.ts`）
