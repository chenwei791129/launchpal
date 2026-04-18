## 1. 前置準備

- [x] 1.1 研讀 `internal/launchctl/readonly.go` 的 `plutil -convert xml1` 轉換流程與 `internal/launchctl/types.go` 的 `detectPlistFormat` helper，確認可複用的公開介面；若為 package-private，規劃搬到 `internal/launchctl` 共用位置或在 `internal/backup` 內實作對等 helper
- [x] 1.2 [P] 在 `frontend/` 執行 `pnpm add diff`（diff 9+ 內建 types，不需 `@types/diff`），驗證匯入 `Diff.diffLines` 可用

## 2. 後端：後端新增「XML 化後的 plist 內容」讀取 API

- [x] 2.1 實作 Binary plist normalized to XML before diffing：在 `internal/backup/backup.go` 的 `GetContent` 讀取檔案後，若偵測為 binary plist 則透過 `plutil -convert xml1 -o -` 轉為 XML；失敗時回退為原始 bytes 並以旗標或額外欄位告知前端格式轉換失敗
- [x] 2.2 在 `app.go` 新增或調整取得當前 plist 內容的 IPC（`GetPlist`），同樣做 binary→XML 轉換；若當前 plist 檔案不存在則回傳空字串搭配 `nil` error，對齊 Diff works when current plist is absent 的行為
- [x] 2.3 [P] 針對 `GetContent` 的 XML 化行為撰寫 `internal/backup/backup_test.go` 單元測試：XML plist 原樣輸出、binary plist 被轉為 XML、`plutil` 失敗時回退為原始 bytes
- [x] 2.4 [P] 針對 `GetPlist`（或其對應的 manager 方法）補上測試：當前檔不存在回傳空字串、binary 當前檔會被轉 XML

## 3. 前端：Diff 計算放在前端，以 current → backup 為 diff 方向

- [x] 3.1 新增 composable `frontend/app/composables/useBackupDiff.ts`，封裝呼叫 `Diff.diffLines(current, backup)` 取得 `Change[]`；以 `current → backup` 為 diff 方向，確保輸入順序明確固定
- [x] 3.2 在同一 composable 實作列對齊轉換器：將 `Change[]` 轉為 `{ left: Row[]; right: Row[] }`，其中 Row 含 `type: 'content' | 'placeholder'`、`text`、行號；context 左右同時 push 原文、removed 僅 left 有內容右側為 placeholder、added 反之
- [x] 3.3 實作 Large diff output is bounded 截斷邏輯：若任一欄 rows 超過 10,000 則截斷並回傳省略行數供 UI 顯示告知
- [x] 3.4 [P] 撰寫 `frontend/app/composables/__tests__/useBackupDiff.test.ts`：涵蓋無差異、純新增、純刪除、混合、current 為空字串、超大輸入截斷等案例

## 4. 前端：Side-by-side Diff Modal 元件（檔案組織；Diff Modal 佈局與觸發流程）

- [x] 4.1 實作 Diff modal shows side-by-side diff with current on left and backup on right：新增 `frontend/app/components/BackupDiffDialog.vue`，props：`visible: boolean`、`backup: Backup | null`；透過 Wails IPC 分別取得當前 plist 與 backup 內容，組入 `useBackupDiff`
- [x] 4.2 以 `Teleport` + fixed overlay 實作 Modal 容器，`max-w-5xl` 寬度、`max-h-[80vh]` 高度；Header 顯示 service 名稱、timestamp、格式警示 banner（若格式轉換失敗）、「no current version」indicator（若 current 為空）
- [x] 4.3 Body 以 `grid grid-cols-2` 實作左右兩欄：左欄綁定 `left` rows、右欄綁定 `right` rows；等寬字體、`whitespace-pre`、允許整個 body 水平捲動；placeholder row 使用 `bg-surface-600/40` 底色
- [x] 4.4 套用配色：removed（左欄 only）`bg-red-500/10 text-red-300`、added（右欄 only）`bg-green-500/10 text-green-300`、未變動 `text-gray-300`、行號 gutter `text-gray-600`
- [x] 4.5 實作 Diff modal supports Cancel and Restore actions：Footer 提供 `Cancel` 按鈕（emit `close`）與 `Restore` 按鈕（emit `restore` 附帶當前 backup）
- [x] 4.6 處理「Backup identical to current plist」情境：當兩欄無差異時顯示「No changes」訊息取代 diff grid
- [x] 4.7 [P] 撰寫 `frontend/app/components/__tests__/BackupDiffDialog.test.ts`：渲染兩欄、配色類別存在、empty-diff 訊息、truncation notice、current 不存在時 header 顯示對應 indicator

## 5. 前端：Settings 頁整合

- [x] 5.1 實作 Diff button next to Restore：於 `frontend/app/pages/settings.vue` Backup History 每列 Restore 按鈕**左側**新增 icon-only Diff 按鈕，附 `title="View diff"` tooltip；確認空列表時不渲染按鈕
- [x] 5.2 在 `settings.vue` 內加入 `showDiffDialog` / `backupToDiff` 狀態、引入 `BackupDiffDialog`；點擊 Diff 按鈕時帶入對應 backup 並開啟 Modal
- [x] 5.3 監聽 `BackupDiffDialog` 的 `restore` event：關閉 Diff Modal 後呼叫既有 `confirmRestore()`，讓 Restore 流程沿用現有確認對話框
- [x] 5.4 實作 Diff button shows a custom hover tooltip explaining its function（design：Diff 按鈕的 hover tooltip）：於 `settings.vue` Diff 按鈕外包 `<span class="relative group">`，旁邊放一個 absolute-positioned tooltip `<span>` 使用 `opacity-0 group-hover:opacity-100 transition-opacity pointer-events-none` 控制顯隱；顯示位置 `top-full right-0 mt-1`，`whitespace-nowrap`；文字內容為 "Preview diff against current plist"；保留 `title` 屬性作為 accessibility fallback

## 6. 驗證

- [x] 6.1 `make test` 跑過 Go 單元測試、`pnpm exec vitest run` 跑過前端測試（Go 4 packages pass, frontend 7 files / 61 tests pass）
- [x] 6.2 `make build-debug` 後開啟 app 手動驗證：XML plist backup 差異、binary plist backup 差異、已刪除服務的 backup（current 不存在）、identical backup、超過 10,000 行 plist 的截斷提示（用戶透過回饋循環確認 side-by-side 顯示正常；補強了 scrollbar 共用與 hover tooltip）
- [x] 6.3 更新 `README.md` 與 `.claude/CLAUDE.md` 若涉及備份相關說明（加入 Diff preview 功能描述）
