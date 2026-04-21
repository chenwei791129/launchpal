# LaunchPal

macOS LaunchAgent 圖形化管理工具。

## UI 語言

**前端 UI 全部文字（含 tooltip、label、button、description、alert、error message）必須使用英文。** 即使討論或 spec 使用中文，實際寫入 Vue 模板或 TypeScript 字串字面量時一律翻成英文。

- **原因**：全 UI（"System Services"、"Read-only"、"Stop service" 等）都是英文，中文字串會破壞一致性；本專案尚未引入 i18n 框架，寫死中文等於永久鎖死該語系。
- **套用時機**：
  - 新增 `.vue` template 文字或 `<script setup>` 內的字串字面量（tooltip、`alert()`、`throw new Error()`、`label`）
  - 改既有字串時若發現中文，順手翻成英文
  - 從 spec/tasks（中文）實作 UI 時，翻譯而非直接貼上
- **例外**：`settings.vue` 等若未來加入 i18n 才可支援中文，屆時英文仍為預設 locale。

## 技術棧

- **後端**: Go + Wails v2
- **前端**: Nuxt 4 + Vue 3 + TailwindCSS
- **平台**: macOS only（launchctl 是 macOS 專屬）

## 目錄結構

```
├── app.go                 # Wails 應用程式綁定（前端呼叫的 API）
├── main.go                # 應用程式入口
├── Makefile               # 建置指令
├── internal/
│   ├── launchctl/         # launchctl 命令封裝
│   │   ├── types.go       # Service, ServiceConfig 等型別（含 StatusConfidence）
│   │   ├── manager.go     # Manager interface
│   │   ├── user.go        # UserManager 實作（~/Library/LaunchAgents）
│   │   ├── system.go      # SystemManager 實作（/Library/LaunchDaemons，唯讀）
│   │   ├── apple_system.go # AppleSystemManager 實作（/System/Library/LaunchDaemons，唯讀）
│   │   └── status_detect.go # System domain 啟發式狀態偵測（pgrep -u + ppid=1 過濾）
│   ├── backup/            # 備份管理
│   │   └── backup.go      # BackupManager 實作
│   └── plistutil/         # plist 格式偵測與 binary→XML 正規化（供 backup、launchctl 共用）
│       └── plistutil.go   # DetectFormat、NormalizeFromPath
├── frontend/              # Nuxt 4 前端專案
│   ├── app/
│   │   ├── pages/         # 頁面（index, system, apple-system, settings, services/[name]）
│   │   ├── components/    # Vue 元件
│   │   ├── composables/   # 組合式函數
│   │   └── types/         # TypeScript 型別
│   └── nuxt.config.ts
├── build/
│   └── darwin/            # macOS 建置設定
└── wails.json             # Wails 專案設定
```

## 開發指令

```bash
make setup       # 安裝依賴
make test        # 執行測試
make build       # 建置 production app
make build-debug # 建置含 devtools
make dev         # 建置並開啟 app
make dmg         # 建置並打包為 DMG
make clean       # 清除建置產物
```

## 備份機制

- 備份目錄：`~/.launchpal/backups/<service-name>/`
- 備份格式：
  - `<timestamp>.plist` - plist 備份檔
  - `<timestamp>.meta.json` - 原始路徑等 metadata
- 自動備份時機：Update、Delete 前
- 保留數量：最近 10 個
- Settings → Backup History 每一筆可透過 Diff 按鈕開啟 **Side-by-side diff 預覽**（左 current / 右 backup，紅/綠配色）再決定是否 Restore；binary plist 會由後端自動轉 XML，若 service 已刪除則左欄全為 placeholder、右欄整份為新增。Diff 行上限 10,000 列，超過顯示截斷提示。

## 服務類型

LaunchPal 支援三種類型的服務：

1. **User Services** (`~/Library/LaunchAgents`)
   - 完整的讀寫權限
   - 可以啟動、停止、建立、更新、刪除服務
   - 支援立即執行（Kickstart：`launchctl kickstart -k`）
   - 支援排程設定（StartCalendarInterval / StartInterval）
   - Cron 語法支援範圍 `a-b`、列舉 `a,b,c`，自動笛卡爾積展開為多筆 StartCalendarInterval（上限 50 筆）
   - 支援環境變數設定（EnvironmentVariables）

2. **System Services** (`/Library/LaunchDaemons`)
   - 唯讀模式
   - 可以查看服務資訊、狀態、plist 內容和 logs
   - 第三方系統級服務

3. **Apple System Services** (`/System/Library/LaunchDaemons`)
   - 唯讀模式
   - 可以查看服務資訊、狀態、plist 內容和 logs
   - macOS 系統內建服務
   - 許多服務使用 binary plist 格式，會自動轉換為 XML 顯示

## 狀態檢測邏輯

### User domain (`UserManager`)

1. 先用 `launchctl list <label>` 取得服務資訊
2. 從輸出中解析 PID（若有則為 running）
3. 若無 PID，用 `pgrep -f <program>` 作為 fallback
4. 跳過常見 shell（bash, sh, zsh）的 pgrep 檢測以避免誤判
5. `StatusConfidence` 永遠為 `verified`（`launchctl list` 在 user domain 為 authoritative）

### System domain (`SystemManager` / `AppleSystemManager`)

`launchctl list` 在 user context 下只看得到 `gui/<uid>` domain 的服務，完全查不到 `/Library/LaunchDaemons` 與 `/System/Library/LaunchDaemons` 的 system daemon，故改用 `status_detect.go` 的啟發式偵測：

1. 從 plist 取得 `UserName`（預設 `root`）與 `program`（`Program` 優先，否則 `ProgramArguments[0]`）
2. `program` 為空 → `unknown` / PID 0 / `unverified`
3. `program` 落在 `commonShells` → `loaded` / PID 0 / `verified`
4. 執行 `pgrep -u <UserName> -f <program>` 取得候選 PID
5. 用 `ps -o ppid= -p <pid>` 過濾保留 `ppid == 1`（由 launchd 起）的 PID
6. 1 個 → `running` / PID / `verified`；0 個 → `stopped` / 0 / `verified`；多個 → `running` / 首個 / `unverified`

### `StatusConfidence` 欄位

`Service` 結構新增 `StatusConfidence string`（`verified` / `unverified`）。前端在 `unverified` 時於 Status 旁顯示 info icon + tooltip 提示「可能不是實際對應的 PID」，純資訊性、無任何動作按鈕。

## Plist 格式處理

- 自動偵測 plist 格式（XML 或 binary）
- Binary plist 會使用 `plutil` 自動轉換為 XML 格式顯示
- 在 Summary 頁面顯示原始格式類型

## Commit Message 規範

本專案使用 [release-please](https://github.com/googleapis/release-please) 自動管理版本號與發布（見 `.github/workflows/release-please.yml`）。`feat` 類型的 commit 會觸發 minor 版本升級，`fix` 會觸發 patch 升級。

- 若變更**未涉及主程式邏輯**（如文件、CI、設定檔、測試、重構），應使用 `chore`、`docs`、`ci`、`test`、`refactor` 等類型，避免不必要的版本更新
- 僅在**實際新增或變更使用者可感知功能**時才使用 `feat`
- 僅在**修復實際 bug** 時才使用 `fix`
- 當用戶要求 commit 當前改動時，應**依功能別分開 commit**，不要將不同功能的變更混在同一個 commit

## Homebrew 發布

- Homebrew Tap：`chenwei791129/homebrew-apps`（獨立 repo，可供多個 app 共用）
- 安裝指令：`brew install --cask chenwei791129/apps/launchpal`
- Cask formula 位於 `homebrew-apps` repo 的 `Casks/launchpal.rb`
- 因未簽名，使用 `postflight` 自動移除 quarantine attribute
- Release 時 `release-please.yml` 的 `update-homebrew` job 自動更新 formula（版本號 + SHA256）
- 跨 repo 寫入使用 `HOMEBREW_TAP_TOKEN`（Fine-grained PAT，scope 限 homebrew-apps）

## 已知限制

- 僅能修改用戶級別服務（~/Library/LaunchAgents）
- 系統服務（/Library/LaunchDaemons、/System/Library/LaunchDaemons）為唯讀模式
- 無法停止以 root 運行的服務（需 sudo 權限）
- 部分系統服務可能需要 Full Disk Access 權限才能查看

## Worktree 設定

使用 `.worktrees/` 目錄（已加入 .gitignore）
