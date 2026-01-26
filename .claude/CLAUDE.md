# LaunchPal

macOS LaunchAgent 圖形化管理工具。

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
│   │   ├── types.go       # Service, ServiceConfig 等型別
│   │   ├── manager.go     # Manager interface
│   │   └── user.go        # UserManager 實作（~/Library/LaunchAgents）
│   └── backup/            # 備份管理
│       └── backup.go      # BackupManager 實作
├── frontend/              # Nuxt 4 前端專案
│   ├── app/
│   │   ├── pages/         # 頁面（index, settings, services/[name]）
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
make clean       # 清除建置產物
```

## 備份機制

- 備份目錄：`~/.launchpal/backups/<service-name>/`
- 備份格式：
  - `<timestamp>.plist` - plist 備份檔
  - `<timestamp>.meta.json` - 原始路徑等 metadata
- 自動備份時機：Update、Delete 前
- 保留數量：最近 10 個

## 狀態檢測邏輯

1. 先用 `launchctl list <label>` 取得服務資訊
2. 從輸出中解析 PID（若有則為 running）
3. 若無 PID，用 `pgrep -f <program>` 作為 fallback
4. 跳過常見 shell（bash, sh, zsh）的 pgrep 檢測以避免誤判

## 已知限制

- 僅管理用戶級別服務（~/Library/LaunchAgents）
- 無法停止以 root 運行的服務（需 sudo 權限）
- 同名服務若同時存在於 LaunchAgents 和 LaunchDaemons 可能造成狀態顯示不一致

## Worktree 設定

使用 `.worktrees/` 目錄（已加入 .gitignore）
