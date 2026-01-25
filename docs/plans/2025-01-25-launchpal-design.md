# LaunchPal 設計文件

> macOS LaunchAgent 圖形化管理工具

## 專案概述

**目標**：提供一個直觀的圖形化介面，讓進階用戶能夠輕鬆管理 macOS LaunchAgents，無需記憶 launchctl 命令。

**目標用戶**：進階用戶（想要更方便地控制 Mac 上的背景服務，但不熟悉命令列）

## 技術架構

```
┌─────────────────────────────────────────┐
│           Wails v2 應用程式              │
├─────────────────┬───────────────────────┤
│   前端 (Web UI)  │     後端 (Go)          │
│   - Nuxt 4      │   - launchctl 封裝     │
│   - 深色主題     │   - plist 讀寫         │
│   - 參考 Podman │   - 檔案系統操作        │
│     Desktop     │   - 備份管理            │
└─────────────────┴───────────────────────┘
```

**技術棧**：
- **後端**：Go 1.21+ with Wails v2
- **前端**：Nuxt 4 + 深色主題（參考 Podman Desktop 風格）
- **資料格式**：plist (XML)、JSON (API)
- **儲存**：`~/.launchpal/` 用於設定與備份
- **分發**：macOS `.app` 單一執行檔

**功能範圍**：
- 第一階段：僅管理 `~/Library/LaunchAgents` 目錄
- 架構設計保留擴充至系統級別的彈性

## 核心功能

1. **服務列表** — 列出所有 LaunchAgents，顯示執行狀態
2. **服務控制** — 啟動/停止/重啟服務
3. **詳細資訊** — 查看服務設定與狀態
4. **編輯功能** — 表單式編輯器 + 原始碼切換
5. **新增/刪除** — 建立新服務、刪除服務
6. **日誌檢視** — 查看 stdout/stderr 輸出
7. **備份還原** — 修改前自動備份，可還原

## 介面設計

### 主介面佈局（參考 Podman Desktop）

```
┌─────────────────────────────────────────────────────────────┐
│ [🔴🟡🟢]                                                     │
├────┬────────────────────────────────────────────────────────┤
│    │  Services                    [🗑 清理] [+ 新增服務]     │
│ 🚀 │────────────────────────────────────────────────────────│
│    │  🔍 Search services...                                 │
│ ⚙️ │────────────────────────────────────────────────────────│
│    │  ☐  STATUS  NAME                    SCHEDULE    ACTIONS│
│    │────────────────────────────────────────────────────────│
│    │  ☐  🟢     com.user.myagent         RunAtLoad   ▶ 🗑 ⋮ │
│    │            Running · PID 1234                          │
│    │  ☐  🔴     com.user.backup          Daily 09:00 ▶ 🗑 ⋮ │
│    │            Stopped                                     │
│    │  ☐  🟢     com.user.sync            Every hour  ▶ 🗑 ⋮ │
│    │            Running · PID 5678                          │
├────┴────────────────────────────────────────────────────────┤
│  3 services · 2 running                        v0.1.0  🔔   │
└─────────────────────────────────────────────────────────────┘
```

### 詳情面板（點擊服務後）

```
┌─────────────────────────────────────────────────────────────┐
│  Services > Service Details                    [⏹][🗑][🔄]  │
├─────────────────────────────────────────────────────────────┤
│  🟢 com.user.myagent                                        │
│     Running · PID 1234                                      │
├─────────────────────────────────────────────────────────────┤
│  [Summary]  [Logs]  [Inspect]  [Edit]                       │
├─────────────────────────────────────────────────────────────┤
│  Program      /usr/local/bin/myagent                        │
│  Arguments    --daemon --port=8080                          │
│  Working Dir  /Users/jeff/projects                          │
│  Run At Load  Yes                                           │
│  Keep Alive   No                                            │
│  Schedule     -                                             │
└─────────────────────────────────────────────────────────────┘
```

### Menu Bar 圖示

```
┌─────────────────────────────────────────────────────┐
│  macOS Menu Bar                          [🚀]      │
│                                           ↓ 點擊   │
│                              ┌────────────────────┐│
│                              │ ● 3 個服務執行中   ││
│                              │ ─────────────────  ││
│                              │ 開啟主視窗...      ││
│                              │ ─────────────────  ││
│                              │ 結束 LaunchPal     ││
│                              └────────────────────┘│
└─────────────────────────────────────────────────────┘
```

### 設計特點

- 左側：極簡圖示導航（服務列表、設定）
- 主區域：表格式列表，可多選操作
- 每列：狀態燈號、名稱+副標題、排程、快速操作
- 底部：狀態列顯示統計與版本
- 詳情：分頁式（Summary / Logs / Inspect / Edit）
- 配色：深色底 + 紫色強調色

## 表單式編輯器

```
┌─ 基本設定 ─────────────────────────────────────────┐
│ 服務名稱 (Label)     [com.user.myagent          ] │
│ 程式路徑 (Program)   [/usr/local/bin/app    ] [📁] │
│ 執行參數 (Arguments) [--daemon, --port=8080     ] │
│ 工作目錄             [/Users/jeff/projects  ] [📁] │
└────────────────────────────────────────────────────┘

┌─ 啟動條件 ─────────────────────────────────────────┐
│ ☑ 登入時自動啟動 (RunAtLoad)                       │
│ ☐ 保持執行 (KeepAlive)                            │
│ ☐ 排程執行 (StartCalendarInterval)                │
│     ┌─ 排程設定 ─────────────────────┐            │
│     │ 每 [日 ▼] 的 [09]:[00] 執行    │            │
│     └────────────────────────────────┘            │
└────────────────────────────────────────────────────┘

┌─ 環境設定 ─────────────────────────────────────────┐
│ 環境變數 (EnvironmentVariables)                    │
│   [PATH    ] = [/usr/local/bin        ] [✕]       │
│   [+ 新增變數]                                     │
│ 日誌輸出路徑 (StandardOutPath)  [/tmp/out.log ] [📁]│
│ 錯誤輸出路徑 (StandardErrorPath)[/tmp/err.log ] [📁]│
└────────────────────────────────────────────────────┘

          [切換原始碼檢視]    [儲存] [取消]
```

## Go 後端架構

### 目錄結構

```
launchpal/
├── main.go
├── app.go               # Wails 應用程式綁定
├── internal/
│   ├── launchctl/       # launchctl 命令封裝
│   │   ├── service.go   # 服務介面定義
│   │   └── user.go      # 用戶級別實作
│   ├── plist/           # plist 讀寫解析
│   ├── backup/          # 備份管理
│   └── logger/          # 日誌讀取
├── frontend/            # Nuxt 4 專案
│   ├── nuxt.config.ts
│   ├── pages/
│   ├── components/
│   └── ...
├── build/               # Wails 建置設定
└── wails.json
```

### 核心抽象

```go
type ServiceManager interface {
    List() ([]Service, error)
    Get(name string) (*Service, error)
    Start(name string) error
    Stop(name string) error
    Restart(name string) error
    Create(config *ServiceConfig) error
    Update(name string, config *ServiceConfig) error
    Delete(name string) error
    GetLogs(name string, logType string) (string, error)
}

type Service struct {
    Name       string            `json:"name"`
    Label      string            `json:"label"`
    Status     string            `json:"status"`
    PID        int               `json:"pid,omitempty"`
    Path       string            `json:"path"`
    Program    string            `json:"program"`
    Arguments  []string          `json:"arguments,omitempty"`
    RunAtLoad  bool              `json:"runAtLoad"`
    KeepAlive  bool              `json:"keepAlive"`
    Schedule   *ScheduleConfig   `json:"schedule,omitempty"`
}
```

## API 設計

| 方法 | 路徑 | 說明 |
|------|------|------|
| GET | `/api/services` | 取得所有服務列表與狀態 |
| GET | `/api/services/:name` | 取得單一服務詳細資訊 |
| POST | `/api/services/:name/start` | 啟動服務 |
| POST | `/api/services/:name/stop` | 停止服務 |
| POST | `/api/services/:name/restart` | 重啟服務 |
| PUT | `/api/services/:name` | 更新服務設定（plist） |
| POST | `/api/services` | 建立新服務 |
| DELETE | `/api/services/:name` | 刪除服務 |
| GET | `/api/services/:name/logs` | 取得服務日誌 |
| GET | `/api/services/:name/plist` | 取得原始 plist 內容 |
| GET | `/api/backups` | 取得備份列表 |
| POST | `/api/backups/:id/restore` | 還原指定備份 |

> 註：使用 Wails 時，這些會透過 Go bindings 直接呼叫，非 HTTP API。

## 日誌檢視

```
┌─ 日誌 ──────────────────────────────────────────────┐
│ 來源: [stdout ▼]  [🔄 重新整理]  [⬇ 自動捲動: 開]   │
├─────────────────────────────────────────────────────┤
│ 2024-01-25 10:23:01  Server started on port 8080   │
│ 2024-01-25 10:23:05  Connected to database         │
│ 2024-01-25 10:24:12  Received request GET /api     │
│ 2024-01-25 10:24:12  Response 200 OK (23ms)        │
└─────────────────────────────────────────────────────┘
```

**實作方式**：讀取 plist 中定義的 `StandardOutPath` / `StandardErrorPath`。

## 備份機制

- **自動備份時機**：每次透過編輯器修改 plist 前
- **備份儲存位置**：`~/.launchpal/backups/<service-name>/<timestamp>.plist`
- **保留策略**：每個服務保留最近 10 個版本

```
┌─ 備份記錄 ───────────────────────────────────────────┐
│ com.user.myagent                                    │
├─────────────────────────────────────────────────────┤
│ 📄 2024-01-25 10:30:00  [檢視] [還原]              │
│ 📄 2024-01-24 15:20:00  [檢視] [還原]              │
│ 📄 2024-01-23 09:10:00  [檢視] [還原]              │
└─────────────────────────────────────────────────────┘
```

## 錯誤處理

| 情境 | 處理方式 |
|------|----------|
| launchctl 命令失敗 | 顯示友善錯誤訊息 + 原始錯誤輸出 |
| plist 格式錯誤 | 儲存前驗證，提示具體錯誤位置 |
| 檔案權限不足 | 提示用戶確認檔案權限 |
| 服務不存在 | 自動重新整理列表，提示服務已被移除 |
| 日誌檔案不存在 | 顯示「尚無日誌輸出」提示 |

## 安全考量

- 僅允許操作 `~/Library/LaunchAgents` 目錄（第一階段）
- 刪除服務前要求二次確認
- 備份機制確保可還原誤操作
- 不儲存或傳輸任何敏感資訊（純本地應用）
- plist 驗證：儲存前檢查必要欄位

## 開發順序

1. **基礎架構** — Wails 專案骨架、Nuxt 專案初始化
2. **服務列表** — launchctl 封裝、列出服務、顯示狀態
3. **服務控制** — 啟動/停止/重啟功能
4. **詳細資訊** — plist 解析、顯示服務設定
5. **編輯功能** — 表單編輯器、plist 寫入、原始碼切換
6. **新增/刪除** — 建立新服務、刪除服務
7. **備份功能** — 自動備份、還原機制
8. **日誌檢視** — 日誌讀取與即時更新
9. **Menu Bar** — 狀態列圖示與快速選單
