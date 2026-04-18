## Context

Settings 頁的 Backup History 目前只提供 Restore 按鈕（`frontend/app/pages/settings.vue:111-116`），點擊後直接彈出確認對話框並覆寫當前 plist。備份機制由 `internal/backup/backup.go` 的 `BackupManager` 管理，檔案存於 `~/.launchpal/backups/<service>/<timestamp>.plist`，搭配 `<timestamp>.meta.json` 紀錄原始路徑。

現有後端 IPC：

- `GetBackupContent(serviceName, backupID) (string, error)`：回傳備份檔**原始 bytes** 字串（未做 binary→XML 轉換）
- `GetPlist(name) (string, error)`：透過 `UserManager` 取得當前 plist；User services 寫入的通常是 XML，但理論上用戶可能手動放入 binary plist
- `internal/launchctl/readonly.go:107` 已有 `plutil -convert xml1 -o -` 的 binary→XML 轉換範式
- `internal/launchctl/types.go:159` 的 `detectPlistFormat` 可偵測 plist 格式

前端目前未使用任何 diff 套件，`frontend/package.json` 依賴樹相對精簡。

## Goals / Non-Goals

**Goals:**

- 用戶在點 Restore 前，可先檢視「Restore 後會發生的變更」。
- Diff 呈現為 unified diff，綠加紅減配色；即使 plist 為 binary，也能呈現可讀的文字 diff。
- 當前 plist 不存在時，Diff 仍可開啟，內容整段為 `+`。
- 新增依賴控制在最小範圍（只新增前端 `diff` 套件，後端不新增）。

**Non-Goals:**

- 不提供 side-by-side diff。
- 不支援備份之間互比。
- 不在其他頁面（Service Summary、System Services）加入 diff 功能。
- 不支援跨格式 diff 排版優化（例如只針對 XML 做語意 diff）——統一以文字行為單位。

## Decisions

### Diff 計算放在前端

使用 npm `diff` 套件（MIT license，單一 JS 檔，體積小）於 Vue 元件內呼叫 `Diff.diffLines(current, backup)` 取得 `Change[]` 陣列（每個 change 有 `added`/`removed`/`value`/`count`），再由前端轉換為左右兩欄對齊的列結構。

**理由**：

- 後端零新增依賴；`diff` 套件是前端社群標準方案。
- Diff 是純顯示邏輯，沒有安全或效能敏感資料；在前端執行不需跨 IPC 傳大量結構化資料。
- 顏色渲染與列對齊排版屬前端職責，diff 計算放前端避免多一層傳輸格式設計。
- 使用 `diffLines` 而非 `createTwoFilesPatch`（後者產 unified diff 字串），因為 side-by-side 需要結構化的 change 陣列以進行左右對齊運算。

**Alternatives considered**：

- 後端用 Go 的 `sergi/go-diff`：多引入一個 Go 依賴，且要設計 diff 結構 IPC 回傳格式，冗餘。
- 直接顯示兩份全文讓用戶肉眼比對：等於沒做。

### 以 `current → backup` 為 diff 方向

Diff 語意 = 「按下 Restore 後會發生的變更」。

- `+` 行 = backup 有、current 沒有 → Restore 後會新增
- `-` 行 = current 有、backup 沒有 → Restore 後會移除

**理由**：讓 diff 直接對應用戶即將執行的動作，免去方向換算。符合 `git diff HEAD..target` 的心智模型（target 是要切過去的那個）。

**Alternatives considered**：反向 `backup → current`（顯示「當前相對於備份改了什麼」）——語意需要用戶自行反推 Restore 結果，認知負擔較高。

### 後端新增「XML 化後的 plist 內容」讀取 API

修改策略擇一（以（a）為主）：

- **(a) 修改 `GetBackupContent` 行為**：在讀取後若 `detectPlistFormat` 判為 binary，透過 `plutil -convert xml1 -o -` 轉為 XML 再回傳。
- **(b) 新增 `GetBackupContentXML`**：保留 `GetBackupContent` 原行為，新增專供 diff 使用的 XML 化版本。

**決策**：採 (a)。目前 `GetBackupContent` 只被 diff 場景（新）使用，未來若需要原始 bytes 可再分流；保留 API 表面最小。若實作時發現其他前端既有使用路徑依賴原始 bytes，回退為 (b)。

同樣地，`GetPlist`（User service 當前版本）若偵測為 binary 也要同樣轉 XML。若當前 plist 檔不存在（service 已被刪除、檔案 missing），回傳空字串 + 無錯誤，前端依此判斷為「全新增」情境。

**Alternatives considered**：在前端用 WebAssembly `plutil` port——過度工程。

### Diff Modal 佈局與觸發流程

- 於 Backup History 每一列 Restore 按鈕**左側**新增 icon-only 按鈕（使用 compare / two-rectangles icon），附 `title="View diff"` tooltip。
- 點擊開啟新 Teleport Modal（與現有 Restore 對話框相同模式），比原本的 Restore 確認對話框更寬（例如 `max-w-5xl`，佔用視窗寬度的 80–90%）以容納雙欄。
- Modal 結構：
  - Header：service 名稱、backup timestamp、格式警示（若有）、「no current version」警示（若當前不存在）
  - Body：side-by-side diff，CSS `grid grid-cols-2`，左欄 = current、右欄 = backup
    - 每一列（row）對應一個邏輯行，左右同高同內容；不同步橫捲以免混亂
    - 欄內行號 gutter（可選）在每欄最左、等寬灰字
    - 等寬字體 `font-mono`、`whitespace-pre`、整個 body 允許水平捲動以容納長行
  - Footer：`Cancel` 按鈕 + `Restore` 按鈕（Restore 點擊後關閉 Diff Modal 並開啟現有 Restore 確認對話框）
- 列對齊演算法：遍歷 `Diff.diffLines` 回傳的 `Change[]`：
  - `!added && !removed`（context）：左右兩欄同時插入原文，對齊
  - `removed`：左欄插入原文並標紅，右欄插入同等數量的 placeholder 空行
  - `added`：右欄插入原文並標綠，左欄插入同等數量的 placeholder 空行
  - 連續 `removed` + `added` 成對時仍維持上述對齊，不做語意上的「modified line」配對（交給視覺對齊自然呈現）
- 配色：
  - 移除（左欄 only）：`bg-red-500/10 text-red-300`
  - 新增（右欄 only）：`bg-green-500/10 text-green-300`
  - Placeholder 空行：`bg-surface-600/40`（與內容格有區別的淡底色）
  - 未變動列：`text-gray-300`，無底色
  - 行號 gutter：`text-gray-600`

**理由**：沿用現有 Modal 模式一致性高；雙欄能讓用戶同時看到「原本是什麼」與「會變成什麼」，對於 plist 這種結構化內容特別直觀。放棄 line-pair 配對處理（把 removed+added 當同一列的 modification）是為了實作簡單且符合 `diff` 套件的語意。

### 檔案組織

Diff Modal 拆成獨立元件 `frontend/app/components/BackupDiffDialog.vue`，避免 `settings.vue` 過於肥大。Props 接 `backup` 與 `visible`，emit `close` 與 `restore`。

### Diff 按鈕的 hover tooltip

Icon-only 按鈕的可發現性偏低，瀏覽器原生 `title` 屬性雖然會顯示提示，但延遲約 1 秒、樣式不可控、且在部分 Chromium 版本會被 Wails webview 壓到視窗外不可見。決定改以 **純 CSS 自訂 tooltip** 實作：

- 實作方式：在 Diff 按鈕外包一層 `<span class="relative group">`，按鈕旁邊放一個 absolute-positioned 的 tooltip `<span class="absolute ... opacity-0 group-hover:opacity-100 pointer-events-none">`。以 `transition-opacity` 做淡入淡出。
- 顯示延遲：透過 `group-hover:delay-300` 或 `transition-delay` 做 ~200ms 緩衝，避免滑鼠掃過瞬間就閃現。
- 位置：tooltip 顯示於按鈕下方（`top-full mt-1`），並以 `whitespace-nowrap` 避免換行；水平位置以 `right-0` 靠右對齊，避免被切到 Modal 邊緣。
- 文字：簡短一句說明功能（例如「Preview diff against the current plist」）。
- 保留 `title` 屬性作為 accessibility fallback（螢幕閱讀器、鍵盤 focus hint）。

**理由**：不引入額外 tooltip 套件（如 `floating-ui`、`vue-tippy`），保持依賴樹精簡；純 CSS 方案對小型靜態 tooltip 足夠，不需要處理 viewport collision 等複雜情境。

**Alternatives considered**：
- 僅沿用 `title`：簡單但延遲長、樣式醜，與用戶「跳出文字提示」預期不符。
- 引入 `@floating-ui/vue`：功能完整但對單一按鈕 overkill。

## Risks / Trade-offs

- **巨大 plist 讓 Modal 卡頓** → 設定 Modal body 最大高度並 virtual-scroll 的成本過高；先限制 diff 結果若左右任一欄超過 10,000 列則截斷顯示，並顯示警告告知截斷資訊。實務上 plist 幾乎不會這麼大，是防禦性設計。
- **雙欄在窄視窗下空間不足** → Modal 設 `max-w-5xl`，視窗寬度若小於 Modal 寬度由瀏覽器自動縮至 viewport 寬度；body 內允許水平捲動以容納長 XML 行，不強制文字 wrap（wrap 會破壞左右對齊）。
- **`plutil` 執行失敗（檔案損毀）** → 回退為原始 bytes；若內容非 UTF-8 可讀，前端 diff 會顯示為亂碼，但至少按鈕可點、不會 crash。需在 Diff Modal header 顯示格式警示。
- **修改 `GetBackupContent` 語意可能影響未來其他調用者** → 目前只有 diff 使用，風險可控；若未來有「原始 bytes」需求再拆分 API。
- **前端新增 `diff` npm 依賴** → 社群活躍、MIT license、無 transitive 依賴，風險低。
- **Icon-only 按鈕的可發現性** → 必須加 `title` tooltip（例如「View diff」）；考量本頁用戶已理解 Restore，Diff 按鈕旁邊即解釋其用途。
