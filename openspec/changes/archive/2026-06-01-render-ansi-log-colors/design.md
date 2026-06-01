## Context

LaunchPal 的 `ServiceLogs.vue` 元件目前以 Vue 文字插值 `<pre>{{ logs }}</pre>` 渲染從 Wails binding `GetLogs` / `GetSystemLogs` 取回的 log 字串。Vue 的 `{{ }}` 自動 escape HTML，但會把 ANSI escape sequence（如 `\x1b[36m`）視為一般字元，導致畫面顯示 `[36m`、`[0m` 等亂碼。

ANSI 顏色 escape sequence 在 launchd daemon log 中很常見：Homebrew 啟動的 service、開發者自寫的 Node/Go/Python CLI service、shell wrapper 用 `\033[` 都會輸出顏色。目前所有 log 相關 spec（core-service-management、clear-service-logs、log-path-customization、delete-log-files-on-service-removal）皆未涵蓋顏色渲染，屬於 spec gap。

整個渲染流程目前完全在前端：後端把整個檔案 string 透過 IPC 傳到前端，前端寫入 `logs` ref。因此 ANSI 處理最自然的位置是前端的一個轉換 utility，不需要動 Go 端、不需要新增 Wails binding。

## Goals / Non-Goals

**Goals:**

- 在 `ServiceLogs.vue` 的 log content `<pre>` 區塊正確渲染 ANSI SGR sequences 為對應的 HTML 樣式片段。
- 支援的 SGR 子集：reset、bold、underline、前景色 30-37、亮前景 90-97、背景色 40-47、亮背景 100-107，共 35 個 code。
- 不支援的 SGR 與所有非 SGR 控制序列一律 strip，輸出剩餘純文字。
- 純文字片段一律 HTML escape；產生的 HTML 只允許白名單 inline style 屬性，杜絕 XSS。
- 提供獨立可測的 utility `ansiToHtml(input: string): string`，在 vitest 下覆蓋 spec 列舉的所有 scenario。

**Non-Goals:**

- 不支援 256-color（`38;5;N`）與 24-bit truecolor（`38;2;R;G;B`）— 遇到時 strip 整段。
- 不模擬 terminal：cursor movement、screen clear、bell、`\b`、`\r` 視覺處理等一律 strip 或保留為原字元，不執行任何螢幕狀態邏輯。
- 不新增 Raw / Rendered toggle 與使用者偏好。
- 不引入第三方 npm 套件做 ANSI 解析。
- 不修改 backend log 抓取路徑（`GetLogs`、`GetSystemLogs`、`internal/launchctl` 任何檔案）。
- 不變更 Clear Logs、Auto-scroll、Refresh、log type 切換等其他 ServiceLogs 控制行為。

## Decisions

### 解析發生在前端，不在 Go 端

**選擇**：在 `frontend/app/utils/ansiToHtml.ts` 做轉換，`ServiceLogs.vue` 在拿到 `logs.value` 後呼叫它。

**理由**：log 抓取已經是一次性整檔字串傳輸，加在前端不需要新增 Wails binding、不污染 launchctl 抽象、保留現有 `internal/launchctl` 與 `internal/privhelper` 的 read-only 邊界。後端純文字輸出也讓「複製貼上」原始 log 仍可正常運作（如果未來開發者要 dump）。

**已考慮的替代方案**：在 `internal/launchctl` 的 `GetLogs` / `GetSystemLogs` 內把 ANSI 轉為 ANSI-stripped 字串或 JSON 結構。被否決因為這破壞了「後端傳純檔內容」的語意，且 IPC payload 會變複雜（必須回傳結構化 token 陣列）。

### 不引入第三方 ANSI 解析套件

**選擇**：自行手寫 SGR parser，純 TypeScript regex + 狀態機，~80 行內。

**理由**：支援的 SGR 子集很小（35 個 code）；引入 `anser` 或 `ansi_up` 會增加 ~5-10KB bundle 並把 XSS 邊界轉嫁給第三方。手寫實作讓 HTML escape 與 style 白名單完全可控。

**已考慮的替代方案**：`anser`（active maintain，輸出 sanitized HTML）。被否決因為 LaunchPal 是 unsigned macOS GUI、不希望追加 runtime dependency；該套件支援 256-color 與 truecolor，超出本 change 的範圍。

### 顏色 token 對應到既有 Tailwind surface palette

**選擇**：定義常數 `SGR_COLOR_MAP`，把 ANSI 顏色 code 映射到具體的 hex 值（針對 LaunchPal 深色背景 `bg-surface-500` 挑選對比足夠的色）。

| ANSI | role | hex |
| --- | --- | --- |
| 30 / 90 | black / bright black | `#5c6370` / `#828896` |
| 31 / 91 | red / bright red | `#e06c75` / `#ff7b85` |
| 32 / 92 | green / bright green | `#98c379` / `#b5e890` |
| 33 / 93 | yellow / bright yellow | `#e5c07b` / `#ffd97d` |
| 34 / 94 | blue / bright blue | `#61afef` / `#82c8ff` |
| 35 / 95 | magenta / bright magenta | `#c678dd` / `#e08af0` |
| 36 / 96 | cyan / bright cyan | `#56b6c2` / `#73d1de` |
| 37 / 97 | white / bright white | `#abb2bf` / `#ffffff` |

**理由**：raw web `red` (#ff0000) 在 `bg-surface-500` 上過於刺眼且飽和度過高；OneDark 風格的調色盤已經是 LaunchPal 整體 UI 的視覺基礎，沿用可保持風格一致。

**已考慮的替代方案**：直接用 CSS named colors。被否決因為對比與飽和度不一致，紫色與藍色在深色背景上幾乎不可見。

### 解析失敗與不支援的 sequence 一律 strip

**選擇**：遇到 malformed escape（例如 `\x1b[` 後面沒接 `m`）、超出白名單的 SGR code（例如 `38;5;N` 256-color）、或非 SGR 的 CSI（例如 `\x1b[2J` 清螢幕）一律 strip 整段 escape sequence 與其參數，純文字部分繼續輸出。不丟出例外、不在 UI 顯示警告。

**理由**：log 渲染是 best-effort 視覺輔助；若因為一段 malformed escape 把整個 log 變空白會比顯示原始亂碼更糟。靜默 strip 行為與 terminal 的容錯行為一致。

### v-html 與 XSS 防護

**選擇**：`ansiToHtml` 輸出固定型態的 HTML 字串：`<span style="...">...</span>` 與 plain text 的串接，所有純文字片段先經過 `escapeHtml`（替換 `&` `<` `>` `"` `'`）。`style` 屬性只生成白名單四個屬性（`color`、`background-color`、`font-weight`、`text-decoration`），值來自固定的 `SGR_COLOR_MAP` 常數，絕不從 user input 拼接。

**理由**：使用 `v-html` 等於把 XSS 邊界拉到 `ansiToHtml` 函式本身。讓函式輸出的 HTML 完全來自內部常數 + escape 後文字，外部 attacker 即使在 log 內塞 `<script>` 或 `"` 也只會看到字面字元。

**已考慮的替代方案**：把字串拆成 token 陣列再用 `v-for` + `<span :style="...">`。被否決因為 token 陣列在 Vue reactivity 下大檔案會更慢；單次 `v-html` 賦值比建立數千個 vnode 划算。

## Implementation Contract

### Behavior

當 user 開啟任一 service 的 Logs tab，`ServiceLogs.vue` 從 `window.go.main.App.GetLogs` 或 `GetSystemLogs` 取得 log 字串並顯示時：

- 包含 ANSI SGR escape sequences 的部分 SHALL 渲染為對應的 HTML 樣式（顏色、bold、underline）。
- 不支援的 SGR code 與所有非 SGR 控制序列 SHALL 從輸出中移除，剩餘純文字 SHALL 正常顯示。
- 純文字內含的 HTML 特殊字元（`<`、`>`、`&`、`"`、`'`）SHALL 以 entity escape，杜絕 XSS。
- 排版屬性（`whitespace-pre-wrap`、`break-all`、`font-mono text-sm`、`text-gray-300`）SHALL 保留。
- 空字串或 null 不影響既有「No logs available for {logType}」 placeholder 行為。

### Interface / Data Shape

新增 utility：

```ts
// frontend/app/utils/ansiToHtml.ts
export function ansiToHtml(input: string): string
```

- 輸入：任意字串（log 內容）。
- 輸出：HTML 字串，可直接綁定到 `v-html`。
- 輸出 grammar：純文字片段（HTML escape 過）與 `<span style="...">...</span>` 的串接，`</span>` 在遇到 reset `\x1b[0m` 或函式結尾時關閉。
- 函式為純函式，無副作用，無 throw。

`ServiceLogs.vue` 模板改動只發生在 log 內容區塊：原本的

```html
<pre v-else class="text-gray-300 whitespace-pre-wrap break-all">{{ logs }}</pre>
```

改為

```html
<pre v-else class="text-gray-300 whitespace-pre-wrap break-all" v-html="renderedLogs" />
```

其中 `renderedLogs` 是 `computed(() => logs.value ? ansiToHtml(logs.value) : '')`。

### Failure Modes

- malformed escape（未閉合的 `\x1b[`）→ strip 該段、繼續解析後續。
- 不支援的 SGR code（例如 `38;5;33`、`5` blink、`7` reverse-video）→ strip 整段 SGR escape（含所有分號分隔參數）。
- 非 SGR CSI（例如 `\x1b[2J` clear screen、`\x1b[H` cursor home）→ strip。
- OSC / DCS / 其他非 CSI escape（`\x1b]...`、`\x1bP...`）→ strip 整段直到該 sequence 的終止字元（`\x07` 或 `\x1b\\`），若終止字元缺失則 strip 至字串結尾。
- 輸入為空字串 → 輸出空字串，元件顯示「No logs available」 placeholder。

### Acceptance Criteria

- `frontend/app/utils/__tests__/ansiToHtml.test.ts` 新增 unit test，至少包含以下 case 並全部通過：
  - 純文字（無 escape）→ 輸出與輸入一致，但 HTML 特殊字元已 escape。
  - 單一前景色 `\x1b[31mhello\x1b[0m` → `<span style="color:#e06c75">hello</span>`。
  - 巢狀 `\x1b[1m\x1b[31mhi\x1b[0m` → 包含 bold 與 red 的 style。
  - Unterminated `\x1b[31mhi` → 結尾自動關閉 `</span>`，內容為 `hi`。
  - Malformed `\x1b[zzz` → strip 該段。
  - 256-color `\x1b[38;5;33mhi\x1b[0m` → strip SGR、純文字 `hi`。
  - 非 SGR `\x1b[2Jhi` → strip `\x1b[2J`、純文字 `hi`。
  - XSS payload `<script>alert(1)</script>` → 字面 escape 為 `&lt;script&gt;alert(1)&lt;/script&gt;`，無 `<script>` tag。
- `frontend/app/components/__tests__/ServiceLogs.test.ts` 新增 case：mock `GetLogs` 回傳含 `\x1b[31mERROR\x1b[0m` 的字串，斷言 `<pre>` 內出現 `<span style="color:#e06c75">ERROR</span>` 且無 `[31m` 字面文字。
- 執行 `make test` 全部通過（含既有 vitest、Go test、TypeScript typecheck）。
- 執行 `make lint` 全部通過。
- 手動驗證：起一個 user service 把 stdout 寫入帶 ANSI 色碼的字串（例如 `printf '\x1b[31mhello\x1b[0m\n' >> ~/Library/Logs/test/stdout.log`），在 LaunchPal Logs tab 開啟該 service 應看到紅色 `hello`，且 `[31m`、`[0m` 字面文字不出現。

### Scope Boundaries

**In scope**：
- `frontend/app/utils/ansiToHtml.ts`（新檔）
- `frontend/app/utils/__tests__/ansiToHtml.test.ts`（新檔）
- `frontend/app/components/ServiceLogs.vue` 的 `<pre>` 區塊與新增 `renderedLogs` computed
- `frontend/app/components/__tests__/ServiceLogs.test.ts` 新增 ANSI 渲染相關 case

**Out of scope**：
- 任何 `internal/`、`app.go`、`admin_mode.go` 的 Go code
- 任何 Wails binding 介面變更
- 256-color 與 truecolor 解析
- 非 SGR 控制序列的模擬（cursor、clear、bell、`\b`、`\r`）
- ServiceLogs 內的其他 UI 行為（Clear、Refresh、Auto-scroll、log type 切換、Teleport dialog）
- 使用者偏好（toggle、設定持久化）
- README / .claude/CLAUDE.md 以外的文件更新

## Risks / Trade-offs

- **Risk**：`v-html` 是 XSS 通道，若 `ansiToHtml` 有漏網之魚會直接 inject。
  → **Mitigation**：所有純文字片段透過 `escapeHtml` 處理；`style` 屬性只 emit 白名單四屬性、值來自編譯期常數；新增 unit test 專門針對 `<script>`、屬性 escape、單雙引號注入等 XSS payload 斷言輸出安全。
- **Risk**：大型 log（MB 級）執行 regex 解析可能延遲渲染。
  → **Mitigation**：目前 `GetLogs` 已限制讀取最近 1MB 範圍，且解析以線性掃描 + 狀態機實作（O(n)），無回溯。若未來實測有問題再評估 chunk 化或 worker 解析；本 change 不預先優化。
- **Risk**：手寫 parser 漏掉 edge case（例如 `\x1b[0;31m` 多重參數）。
  → **Mitigation**：unit test 覆蓋 spec 中所有 scenario，包括多參數、空參數（`\x1b[m` 等同 reset）、不支援參數混雜支援參數（`\x1b[1;38;5;33m` 應整段 strip）。spec 中以 SBE 表格列出每個輸入對應的輸出。
- **Trade-off**：放棄 256-color / truecolor 支援使部分現代 CLI 工具（如 `chalk` 預設 truecolor）的顏色全部消失。
  → 接受。基本 16-color 已能涵蓋 80% 場景；之後若有需要再以新 change 擴充 palette。
