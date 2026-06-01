## 1. 顏色對應表與單元測試骨架

- [x] 1.1 在 `frontend/app/utils/ansiToHtml.ts` 定義 `SGR_COLOR_MAP` 常數，將 SGR codes 30–37、40–47、90–97、100–107 對應到 design 顏色 token 對應到既有 Tailwind surface palette 表格列出的 hex 值。驗證：執行 `cd frontend && npx vitest run app/utils/__tests__/ansiToHtml.test.ts`，新增的「SGR code to style mapping 例表」對照測試（每個 SGR code 一個 case，斷言對應的 style fragment 完全符合 spec 表格）全部通過。
- [x] 1.2 在 `frontend/app/utils/__tests__/ansiToHtml.test.ts` 為「Render ANSI SGR colors in service log output」需求中的純文字 scenario 寫一個失敗測試（輸入 `"hello world\n"`，期望輸出與輸入一致且不含 `<span>`），執行 `npx vitest run` 確認測試 fail（紅燈，因為 `ansiToHtml` 尚未實作回傳值）。

## 2. ansiToHtml 核心解析（前端，不引入第三方套件）

- [x] 2.1 在 `frontend/app/utils/ansiToHtml.ts` 實作 `escapeHtml(text: string): string`，依 spec「HTML-escape plain text and disallow non-whitelisted style attributes」需求把 `&` `<` `>` `"` `'` 映射為對應 entity。驗證：新增測試覆蓋 spec 中「HTML special characters in plain text are escaped」與「Quotes inside plain text are escaped」兩個 scenario，執行 `npx vitest run app/utils/__tests__/ansiToHtml.test.ts` 通過。
- [x] 2.2 實作 `ansiToHtml(input: string): string` 主解析函式（解析發生在前端，不在 Go 端），interface / data shape 依 design 規範 — 純函式、無 throw、輸入 string 輸出 HTML 字串；以線性掃描 + 狀態機處理 CSI（`ESC [`）序列；遇到末端 `m` 視為 SGR 並查表，未支援的參數或非 `m` 末端皆 strip。驗證：依 spec「Render ANSI SGR colors in service log output」的「Single foreground color span」「Bold combined with color」「Multiple parameters in one SGR sequence」三個 scenario 寫測試，先紅後綠，最後執行 `npx vitest run` 通過。
- [x] 2.3 在 `ansiToHtml` 內擴充處理「Strip unsupported and malformed escape sequences」需求要求的 OSC（`ESC ]` 直到 `BEL` 或 `ESC \`）、DCS / SOS / PM / APC、未終止 CSI、unknown final byte，全部 strip。failure modes 依 design 規範：不丟例外、不在 UI 報錯、剩餘純文字繼續輸出（對應「解析失敗與不支援的 sequence 一律 strip」決策）。驗證：對 spec「stripping behavior cases」例表中 6 個 input 各寫一個測試斷言輸出，執行 `npx vitest run` 通過。
- [x] 2.4 確保 `ansiToHtml` 在輸入結尾仍有未關閉的 span 時自動補上 `</span>`，並對 reset code `0` 與空參數 `\x1b[m` 視為 reset（依「不引入第三方 ANSI 解析套件」決策的手寫狀態機）。驗證：新增三個測試 case — `"\x1b[31mhi"` 結尾自動關 span、`"\x1b[m"` 視為 reset、reset 後的純文字不被包在 span 內。執行 `npx vitest run` 通過。
- [x] 2.5 為「v-html 與 XSS 防護」決策補上邊界測試：style 屬性絕不從 input 拼接、attacker payload 即使在被 SGR 包裹的範圍內也只能透過 escapeHtml 出現。驗證：依 spec scenario「Attacker payload inside an SGR-wrapped span is escaped」寫測試，斷言輸出中只存在 spec 列舉的四個白名單 style 屬性（`color`、`background-color`、`font-weight`、`text-decoration`），無 `onmouseover` 等其他屬性洩漏。執行 `npx vitest run` 通過。

## 3. ServiceLogs 整合

- [x] 3.1 在 `frontend/app/components/ServiceLogs.vue` 的 `<script setup>` 從 `~/utils/ansiToHtml` 匯入 `ansiToHtml`，新增 `renderedLogs` computed：當 `logs.value` 為 null/empty 回傳空字串，否則回傳 `ansiToHtml(logs.value)`。驗證：對 component 寫單元測試確認 `renderedLogs` 對 null 輸出 `''`、對 `"\x1b[32mOK\x1b[0m"` 輸出 spec「Logs containing ANSI colors render as colored spans」期望的 HTML，執行 `npx vitest run app/components/__tests__/ServiceLogs.test.ts` 通過。
- [x] 3.2 將 `ServiceLogs.vue` 模板中 log 內容區塊由 `<pre v-else class="text-gray-300 whitespace-pre-wrap break-all">{{ logs }}</pre>` 改為 `<pre v-else class="text-gray-300 whitespace-pre-wrap break-all" v-html="renderedLogs" />`，保留 `font-mono text-sm` 等容器類別。依「Mount rendered output in ServiceLogs view」需求，loading、error、empty 三個分支不動。驗證：新增 component 測試覆蓋 spec 該需求的四個 scenario（colored spans、empty placeholder、loading state、error state），執行 `npx vitest run app/components/__tests__/ServiceLogs.test.ts` 通過。
- [x] 3.3 跑全套自動化以滿足 design 列舉的 acceptance criteria 並守住 scope boundaries（僅 `frontend/app/utils/ansiToHtml.ts`、其測試、`ServiceLogs.vue`、其測試四檔；不動 Go code 與 Wails binding）。在專案根目錄依序執行 `make lint`（含 frontend eslint）與 `make test`（含 vitest、Go test、TypeScript typecheck）並全部通過；同時用 `git diff --name-only` 確認本 change 只動到上述四個檔案。驗證：兩個 make target 各自 exit 0，且 `git diff --name-only` 輸出僅包含 in-scope 檔案。

## 4. 手動視覺驗證

- [ ] 4.1 請使用者執行手動視覺驗證：建立或挑選一個 user service 讓其 stdout 寫入帶 ANSI 顏色的內容（例如指令 `printf '\x1b[31mERROR\x1b[0m booted\n' >> ~/Library/Logs/<label>/stdout.log`），在 LaunchPal Logs tab 打開該 service，確認 `ERROR` 文字顯示為紅色（hex `#e06c75`），且畫面上不出現 `[31m`、`[0m` 字面文字；同時切換 stdout/stderr、Refresh、Auto-scroll 按鈕，確認其行為與本 change 之前一致。驗證：使用者口頭回報視覺結果符合上述描述即視為通過；若不符則回到 3.1 / 3.2 修復。
