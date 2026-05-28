## Context

LaunchPal 目前管理 user services（`~/Library/LaunchAgents`）時，建立新 service 一律走 `CreateServiceModal.vue` 從空白表單輸入。當使用者想建立與既有 service 相似的新 service（例如把 `com.example.ticker` 複製成 `com.example.ticker-staging`，只改少數欄位），必須把 Program、Arguments、WorkingDirectory、EnvironmentVariables、Schedule 等所有欄位逐一抄回 modal，過程繁瑣且容易抄錯。

GitHub issue #17 要求加入一個 Copy/Clone 動作，讓使用者一鍵複製 user service 設定到 New Service 流程。Issue 字面描述為「name input dialog」，但在 discuss 階段確認改採「預填的完整 CreateServiceModal」方案 — 既可避免 name-only 對話框後使用者還要再進 Edit tab 改其他欄位的兩步式流程，也保留同一介面內微調 label 之外欄位的彈性（Edit tab 目前不支援改 label）。

Stakeholders：LaunchPal 終端使用者（管理多組相似 launchd unit 的開發者／DevOps）。受影響檔案僅限 frontend；backend / Wails binding / `.plist` schema 不變。

## Goals / Non-Goals

**Goals:**

- 使用者在 user service detail 頁可以一鍵啟動「以當前 service 為樣板建立新 service」的流程。
- Clone 出來的 service 預設不會在登入時自動執行（`RunAtLoad = false`），避免使用者意外得到兩個同時 run-at-load 的 service。
- Clone 流程允許使用者在送出前微調任何欄位（不只是 label），不需要先建立再進 Edit tab。
- 重用既有 `CreateServiceModal` 的所有驗證與 log path 自動組裝邏輯；不複製貼上 modal 的 UI 程式碼。
- Clone 完成後自動導向新 service 的 detail 頁，讓使用者立即看到結果。

**Non-Goals:**

- 不在後端新增 `CloneService` Wails binding 或 `UserManager.Clone` method。
- 不支援 System service / Apple System service 的 clone。
- 不支援跨 LaunchAgents directory 的 clone（例如把 user service clone 為 system daemon）。
- 不引入 name-only 對話框路徑；單一 prefilled modal 已涵蓋 issue 描述的全部需求。
- 不對 label 做前端預檢（不在輸入時查詢是否存在），重複 label 由後端 `UserManager.Create` 回報。
- 不修改 `CreateServiceModal` 在非 prefill 模式下的既有行為（label 空白、`runAtLoad` 預設 true、所有欄位空白）。

## Decisions

### Reuse the existing `CreateServiceModal` via an optional `prefill` prop

`CreateServiceModal` 既有的內部狀態（`form`、`argumentsText`、`envVars`、`schedule`、`wakeSystem`）與 reset 邏輯都已成熟。Clone 流程加入一個 optional `prefill?: ServiceConfig` prop：

- 當 `prefill` 為 `undefined`（既有路徑，從 list 頁的 "New Service" 按鈕開啟）→ 行為完全不變，所有欄位空白、`runAtLoad` 預設 `true`。
- 當 `prefill` 有值 → modal 開啟時把 `prefill` 的所有欄位填入內部 state，但兩個欄位有特殊處理：
  - `form.label` 強制為空字串（使用者必須另外輸入新 label）
  - `form.runAtLoad` 強制為 `false`（避免登入時雙重執行）

**Alternatives considered**:

- 另寫一個獨立的 `CloneServiceModal` 元件：被否決。會大量重複 `CreateServiceModal` 的表單、驗證、log path 組裝邏輯，後續任何 modal 改動都得同步兩份。
- 在 `CreateServiceModal` 加入 `mode: 'create' | 'clone'` enum prop：被否決。`mode` 只是「有沒有 prefill」的代理，optional prop 本身已足夠表達意圖且更貼近 Vue 慣例（presence-based behavior）。

### Copy button 放在 detail 頁 header action 區，僅 user service 顯示

`services/[name].vue` 的 header 已是現有的 action 集中地（Start/Stop/Restart/Run Now）。新增的 Copy button 條件式渲染在 `serviceType === 'user'`，圖示用「兩張紙堆疊」的標準 copy icon，配色採用既有 `bg-surface-200 hover:bg-surface-100`（與 Restart 同調，中性、非破壞性動作）。

System / Apple System detail 頁完全不顯示此 button，因為：
- Issue AC 明文限制為 user services。
- System daemon 通常與 Homebrew 等安裝路徑綁定，clone 到 `~/Library/LaunchAgents` 不會自動把對應的執行檔搬過來，使用者得到一個指向不存在 program 的 user service。

**Alternatives considered**:

- 放進 ServiceSummary 區塊內：被否決。Copy 與其他 action button 性質一致，分散會讓使用者掃描動作時失去焦點。
- 包進新的「...」overflow menu：被否決。目前 header 只有 4 個 action（Stop/Restart/Run Now + 即將加入的 Copy = 4 顆），還沒到需要 overflow 的密度；為一顆按鈕額外開一條 menu 設計過度。

### `RunAtLoad` 在 clone 時強制為 `false`，無 escape hatch

來源 service 即使 `runAtLoad: true`，clone 出來的 modal 也會把 checkbox 設為 unchecked。使用者若真的希望新 service 也 run-at-load，可以在 modal 內自己勾回去再送出 — 強制 default 為 false 只是預設值，不是 hard constraint。

理由：clone 的典型使用情境是「同一台機器上跑兩個變體」，兩個都 run-at-load 幾乎都不是使用者的意圖（埠號衝突、log 互蓋、resource 競爭）。把預設改為 false 把「容易踩雷」變成「需要主動選擇」。

### Modal `created` event 帶上新 label，由 detail 頁負責導頁

Modal 既有的 `created` event 目前沒有 payload，因為 list 頁只需要重新拉服務列表。Clone 流程需要知道新建 service 的 label 才能導頁，因此：

- 把 `created` event 的 signature 改為 `created(label: string)`，所有 emit 點都帶上 `form.label` 的值。
- `pages/index.vue` 與 `pages/system.vue` 的 listener 多接一個參數但忽略它（既有重新拉列表邏輯不變）— 向後相容。
- `pages/services/[name].vue` 在 listener 內呼叫 `navigateTo('/services/<label>?type=user')`。

**Alternatives considered**:

- 在 modal 內直接 navigate：被否決。讓子元件決定導航是耦合，且 modal 在 list 頁的既有用法不需要導航。
- 透過 Wails event 機制：殺雞用牛刀，前端 emit 已足夠。

### 重複 label 不做前端預檢

提交時若後端 `UserManager.Create` 回 `service X already exists` 錯誤，modal 把錯誤訊息塞進既有 `error.value` 顯示區，不關閉表單、不重置欄位。使用者改 label 後可直接重試。

理由：(a) 後端是 source of truth，前端預檢與後端可能不一致；(b) 重複命名是 edge case，不值得為了即時回饋多打一次 Wails 呼叫；(c) 既有的 error 顯示 UX 已成熟。

## Implementation Contract

**Observable behavior**

1. 在 `/services/<label>?type=user` 頁面，header 的 action 區出現一顆「Copy」button，位於 Run Now 右側。
2. 按下 Copy → `CreateServiceModal` 開啟，標題保留「New Service」（不需特別區分 clone 模式），所有 form 欄位以當前 service 的設定填入，唯獨 `Service Label` 輸入框為空、`Run at Load` checkbox 為 unchecked。
3. 使用者輸入新 label（例：`com.example.ticker-staging`）後送出。若 label 與既有 service 衝突，modal 顯示 inline 紅字 "service com.example.ticker-staging already exists"，欄位內容保留。
4. 成功送出後 modal 關閉，瀏覽路徑切換到 `/services/com.example.ticker-staging?type=user`，新的 detail 頁載入完整資料。
5. 新建立的 plist 檔案位於 `~/Library/LaunchAgents/com.example.ticker-staging.plist`，內含與來源相同的 Program / ProgramArguments / EnvironmentVariables / Schedule / WorkingDirectory / KeepAlive / StandardOut|ErrPath（log path 依新 label 由 `composeLogPaths` 重新組裝），但 `RunAtLoad = false`。
6. 在 `/services/<label>?type=system` 或 `?type=apple-system` 頁面，Copy button **不出現**。

**Interface / data shape**

- `CreateServiceModal` 新 prop：`prefill?: ServiceConfig | null`（optional，預設 `undefined`）。
- `CreateServiceModal` 的 `created` event signature：由 `() => void` 改為 `(label: string) => void`。所有現有 emit 點需更新。
- `services/[name].vue` 新狀態：`showCloneModal: Ref<boolean>`、`cloneSource: Ref<ServiceConfig | null>`。
- 不新增任何 Wails binding，不修改 `ServiceConfig` 型別定義。

**Failure modes**

- 重複 label → 後端回 `service <label> already exists`，前端塞進 modal 既有 `error.value`，不關閉表單。
- 來源 service `service.value` 為 null（尚未載入完成或載入失敗）→ Copy button `disabled`，避免送出空 prefill。
- log path 不存在的 parent directory → `UserManager.Create` 既有邏輯會 `MkdirAll`，clone 不需特別處理。
- 使用者中途關閉 modal（按 X 或 Cancel）→ 既有 reset 邏輯把表單清空；下次正常開啟 modal（非 prefill）仍會看到空白表單。

**Acceptance criteria**

- 自動化測試：
  - `frontend/app/components/__tests__/CreateServiceModal.test.ts` 新增測試：傳入 `prefill` 後，所有非 `label`/`runAtLoad` 欄位被正確填入；`label` 為空、`runAtLoad` 為 false；提交時 `ServiceConfig` 含新 label 與 `runAtLoad: false`。
  - 新檔 `frontend/app/components/__tests__/CloneUserService.test.ts`（或併入 detail 頁的 test）：模擬點擊 Copy button → modal opens with prefill；模擬 `@created('new-label')` → `navigateTo` 被呼叫且路徑為 `/services/new-label?type=user`。
  - 既有 `CreateServiceModal.test.ts` 在 `prefill` 未提供時的所有既有測試 → 全數通過（向後相容驗證）。
- 手動驗證（必須在 user 端跑一次）：建立一個 user service `com.example.copy-source`，按 Copy → 命名為 `com.example.copy-dest` 送出 → 確認 `~/Library/LaunchAgents/com.example.copy-dest.plist` 已建立、`RunAtLoad` 為 false、其他設定與來源相同、頁面導向新 service detail。
- Lint / typecheck：`make lint`、`make test` 全綠。

**Scope boundaries**

In scope：
- `frontend/app/pages/services/[name].vue`：新增 Copy button、clone state、`@created` listener 含導頁。
- `frontend/app/components/CreateServiceModal.vue`：新 prop、prefill 初始化邏輯、`created` event payload。
- 既有 `created` event listener（list 頁、system 頁）：接收新 payload 但忽略它。
- 對應單元測試。
- `openspec/specs/core-service-management/spec.md` delta：新增 "Cloning a user service" requirement。

Out of scope：
- 後端 `internal/launchctl/user.go`、`app.go`：完全不動。
- System / Apple System service 的 clone UI。
- New Service modal 在非 prefill 模式的任何行為改動（label 預設、`runAtLoad` 預設等）。
- 任何 `.plist` schema 或 `ServiceConfig` 結構調整。
- 從 detail 頁的其他位置（例如 ServiceSummary）啟動 clone — 只在 header action 列。

## Risks / Trade-offs

- [使用者忘記 `runAtLoad` 已被強制 false] → 把 checkbox 視覺擺在 prefill 完成的 modal 上方顯眼處（已是現有位置），使用者送出前一定會看到；不再額外加 banner，避免訊息過載。
- [`created` event signature 變更為 breaking change] → 改動點集中在三個 listener（`pages/index.vue`、`pages/system.vue`、新加的 `pages/services/[name].vue`），且新 payload 只是 optional 多帶；既有 listener 不接收參數仍可運作（TypeScript 會允許忽略 emit args）。
- [使用者把 user service clone 後立刻改 program 路徑指到不存在的 binary] → 與 New Service 既有風險完全等價，沒有新增 attack surface；後端啟動失敗時 `launchctl list` 既有 status 顯示會反映。
- [重用 modal 導致 prefill 模式下 reset 行為可能誤觸 — 例：modal 關閉後再開非 prefill 流程，殘留欄位] → 在 modal 內：watch `prefill` 變化時重新初始化；watch `isOpen` 由 false→true 時，依當下 `prefill` 重新填或清空，確保乾淨狀態。

## Migration Plan

無資料遷移、無 schema 變更。部署即生效。回滾策略：revert PR 即可，不需要清理任何已寫入的檔案（既有透過 clone 建立的 plist 與一般 Create 等價）。

## Open Questions

- Copy button 的 icon 與 tooltip 文案：規劃用 Heroicons 的 `document-duplicate` outline 風格、tooltip "Copy this service" — 若 UX 已有別的偏好可在 apply 階段調整，不影響 design 決策。
