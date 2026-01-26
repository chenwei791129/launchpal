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
# 開發模式
wails dev

# 建置
wails build

# 建置（含 devtools 調試）
wails build -debug
```

## 重要注意事項

- 僅支援 macOS（launchctl 是 macOS 專屬命令）
- 目前僅管理用戶級別服務（~/Library/LaunchAgents）
- 備份儲存於 ~/.launchpal/backups/
- 前端使用 SSG 模式（ssr: false），靜態資源嵌入 Go 執行檔

## Worktree 設定

使用 `.worktrees/` 目錄（已加入 .gitignore）
