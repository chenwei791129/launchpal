## Why

Logs 檢視頁面（`ServiceLogs.vue`）目前以 `<pre>{{ logs }}</pre>` 純文字插值方式渲染 log 內容，使得 daemon 程式寫入的 ANSI 顏色 escape sequence（例如 `\x1b[36m...\x1b[0m`）被當作字面文字顯示，畫面充斥 `[36m`、`[0m` 等亂碼字元。Homebrew 服務、開發者自寫的 Node/Go/Python CLI service 普遍輸出彩色 log，這個落差降低了 LaunchPal 作為 launchd 除錯工具的可讀性。

既有所有與 logs 相關的 spec（core-service-management、clear-service-logs、log-path-customization、delete-log-files-on-service-removal）皆未涵蓋 ANSI 顏色渲染，目前的亂碼行為並非刻意設計，而是 spec gap。

## What Changes

- 在前端新增 ANSI → HTML 轉換層 `frontend/app/utils/ansiToHtml.ts`，將輸入的 log 文字解析為純字元片段與 SGR 樣式片段的混合結構。
- `ServiceLogs.vue` 的 log 內容區塊改為呼叫 `ansiToHtml(logs)` 並透過 `v-html` 渲染轉換後的 HTML；保留 `whitespace-pre-wrap break-all` 字型與排版設定。
- 支援的 SGR 子集：reset（`0`）、bold（`1`）、underline（`4`）、前景色 30-37、亮前景 90-97、背景色 40-47、亮背景 100-107。其餘 SGR codes 與所有非 SGR escape sequence（cursor movement、screen clear、bell 等）一律 strip 掉，不渲染、不報錯。
- 顏色實際輸出時對應到 LaunchPal 既有 TailwindCSS surface palette 的對比色，而非 raw web 16-color，以確保深色背景上的可讀性。
- 安全性：轉換層輸出前對所有純文字片段進行 HTML escape；style 屬性只允許 `color`、`background-color`、`font-weight`、`text-decoration` 四個白名單屬性。
- 不新增任何 Wails binding；不動 `internal/launchctl`、`internal/privhelper`、`app.go`；不引入 npm runtime 相依。轉換層為單一前端 adapter，所有解析邏輯放在 `ansiToHtml.ts` 中並用 vitest 覆蓋。

## Non-Goals

- 不支援 256-color（`38;5;N`）與 truecolor（`38;2;R;G;B`）palette。落到此 escape 時，整段 SGR 一律 strip。理由：launchd daemon log 場景中極少出現，且需要決定更大的色彩 token 對應表，徒增複雜度。
- 不支援非 SGR 控制序列（cursor up/down、clear line、保存游標、`\b`、`\r` 視覺處理等）。Log 檢視是被動唯讀的 buffer dump，不是 terminal emulator。
- 不新增使用者可切換的 Raw / Rendered toggle 與偏好設定。預設一律 render；要看 raw 可複製貼到 terminal。
- 不修改 backend log 抓取路徑（`GetLogs` / `GetSystemLogs`）；轉換完全發生在前端拿到字串之後。
- 不變更 Clear Logs 控制、Auto-scroll、Refresh 按鈕的行為。
- 不引入 `ansi_up`、`anser` 等第三方 npm 套件。SGR 解析範圍小、安全邊界簡單，自行手寫實作以避免追加 runtime dependency；測試覆蓋率以 spec 中列舉的場景為準。

## Capabilities

### New Capabilities

- `ansi-log-rendering`: ServiceLogs 檢視在前端解析 stdout/stderr 內含的 ANSI SGR escape sequences 並渲染為對應顏色與字重的 HTML，無法識別或不支援的 escape sequence 一律 strip 後輸出純文字。

### Modified Capabilities

(none)

## Impact

- Affected specs: `ansi-log-rendering`（新增）
- Affected code:
  - New:
    - frontend/app/utils/ansiToHtml.ts
    - frontend/app/utils/__tests__/ansiToHtml.test.ts
  - Modified:
    - frontend/app/components/ServiceLogs.vue
    - frontend/app/components/__tests__/ServiceLogs.test.ts
  - Removed: (none)
