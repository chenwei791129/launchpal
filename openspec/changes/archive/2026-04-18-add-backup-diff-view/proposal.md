## Why

目前 Settings → Backup History 只能直接 Restore，用戶在回復之前看不到備份與當前 plist 的差異，必須自己另外開啟檔案或靠記憶判斷是否真的要覆蓋。這使得 Restore 動作風險偏高，容易誤覆蓋重要修改。提供 diff 預覽可以讓用戶在 Restore 前做出明確、有依據的決定。

## What Changes

- 於 Settings 頁 Backup History 列表中，每一列在 Restore 按鈕左側新增 **Diff** icon 按鈕。
- 滑鼠 hover 於 Diff 按鈕時顯示明顯的文字 tooltip 說明按鈕功能（非僅依賴瀏覽器原生 `title` 屬性，因該原生提示延遲 1 秒且樣式不可控）。
- 點擊 Diff 按鈕開啟 Modal，以 **side-by-side diff（雙欄並列）** 方式呈現差異，左欄為 current（當前 plist），右欄為 backup（備份內容），左右同一列對應同一邏輯行：
  - 僅存在於右欄的行（Restore 後會新增）：右欄格底色綠、文字綠，左欄該列為空白 placeholder
  - 僅存在於左欄的行（Restore 後會移除）：左欄格底色紅、文字紅，右欄該列為空白 placeholder
  - 左右皆存在但內容不同的行：左欄紅、右欄綠
  - 未變動的行：左右皆顯示原文，無特別配色
- 當 plist 為 binary 格式時，先用 `plutil` 轉為 XML 後再 diff（沿用 Apple System Services 現有作法）。
- 當 service 的當前 plist 不存在（例如服務已被刪除）時，Diff 仍可開啟，左欄整欄為空白 placeholder、右欄為整份 backup 內容（皆標示為新增），代表「Restore 會新建整個檔案」。
- Diff 計算在前端執行，透過 npm `diff` 套件產生變更集合，再由前端渲染為左右對齊的兩欄；後端僅負責提供「當前 plist 內容（必要時 XML 化）」與既有的備份內容 API。
- Modal 內容左右獨立可縱向捲動、整體可橫向捲動以容納長行；提供 Cancel 與 Restore 按鈕，Restore 按鈕沿用現有確認對話框流程。

## Non-Goals

- **不** 提供 unified diff（單欄交錯）切換；僅支援 side-by-side。
- **不** 支援 backup 之間互相比較；僅比對 backup 與當前版本。
- **不** 提供編輯 diff 或部分套用（partial apply / patch）的功能。
- **不** 在 Service Summary 頁面加入 diff 按鈕；本次只更動 Settings 頁。
- **不** 處理 User Services 以外類型（System、Apple System）的 backup diff，因為那些服務為唯讀、不會產生 backup。

## Capabilities

### New Capabilities

- `backup-diff-preview`: 於 Settings 頁提供備份與當前 plist 的 unified diff 預覽，含 binary plist XML 化處理與當前版本不存在時的整檔 `+` 顯示行為。

### Modified Capabilities

(none)

## Impact

- **Affected specs**: 新增 `specs/backup-diff-preview/spec.md`
- **Affected code**:
  - `app.go`：新增取得當前 plist 內容（已 XML 化）的 IPC 方法（若不存在）
  - `internal/backup/backup.go`：可能新增 XML 化後的 backup 內容取得方法，或由 `app.go` 層處理轉換
  - `frontend/app/pages/settings.vue`：新增 Diff 按鈕、Diff Modal、呼叫 diff API
  - `frontend/package.json`：新增 `diff` 依賴
  - 可能新增 `frontend/app/components/BackupDiffDialog.vue`（若拆元件）
- **Dependencies**: 新增前端 npm 套件 `diff`（MIT）
