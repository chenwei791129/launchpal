## Context

三個服務清單頁面原本只有文字搜尋。服務數量龐大時（Apple System Services 有 400+ 個）難以定位特定狀態或類型的服務，因此加入 Status / Type 兩個 multi-select dropdown filter。

本 change 建立於 2026-04-09，實際實作於 2026-08-25，中間相隔 138 天、127 個 commit。這段落差造成兩處 artifacts 與程式碼脫節，都在實作期間才被發現，是本設計文件必須記錄的關鍵背景：

1. **Type filter 少了 KeepAlive**：`15b5f89 feat: structured KeepAlive config with launch-policy radio`（2026-06-01）引入第三種 launch policy。原 spec 的 Type 只有 Scheduled / RunAtLoad / None，且把 None 定義為「無 schedule 且 `runAtLoad === false`」。因為 Keep Alive 政策不寫獨立的 `RunAtLoad`（launchd 由 KeepAlive 隱含），KeepAlive 服務的 `runAtLoad === false`，會被誤判為 None —— 與畫面上實際掛著的 `KeepAlive` badge 直接矛盾。
2. **System 頁面不在共用元件內**：`0c543d7` 曾把 System 與 Apple System 收斂進 `ReadOnlyServiceList.vue`，那是本 change 建立時的狀態，所以原 tasks 寫「整合至 `ReadOnlyServiceList.vue`」時確實涵蓋兩者。但 `3c4ed37 feat(admin-mode)`（2026-04-23）為了加入 Admin Mode banner、New Service 按鈕與帶「Also delete log files」的刪除對話框，把 `system.vue` 拆回獨立頁面。現在只剩 `apple-system.vue` 委派給 `ReadOnlyServiceList`，因此「整合至 ReadOnlyServiceList」只覆蓋三個頁面中的兩個。

第二點的性質與第一點不同：spec 的**文字內容**沒錯，錯的是它對**程式碼結構**的假設。這類 drift 讀 artifacts 是看不出來的，必須回頭確認哪些頁面實際 consume 該元件。

## Goals / Non-Goals

**Goals:**

- 三個清單頁面（User / System / Apple System）都提供 Status 與 Type multi-select dropdown filter
- 同一 dropdown 內選項為 OR，跨 dropdown 與既有 Search 之間為 AND
- Type 選項與 `ServiceRow.vue` Type 欄位實際渲染的 launch-policy badge 對齊，filter 結果不得與使用者眼前看到的 badge 矛盾
- 過濾邏輯集中一處，三個頁面不得各自實作，以免行為分歧
- 以測試釘住「三個清單頁面都掛上 filter bar」這個集合，避免日後新增或拆分頁面時再度遺漏

**Non-Goals:**

- 不持久化 filter 選取狀態（不寫入 settings、不寫入 URL query、不跨頁面保留）
- 不引入 i18n；所有 filter 文案維持英文，符合專案既有 UI 語言規則
- 不改動 `ServiceRow.vue` 的 badge 顯示邏輯與其優先序
- 不新增後端 binding；過濾完全在前端既有的 `Service[]` 上進行
- 不調整 Status / Type 以外的過濾維度（例如依 log 路徑、依 domain 過濾）

## Decisions

### 共用 Filter Bar 元件與 filterServices 過濾邏輯分離

分成三層：UI 放在 `ServiceFilterBar.vue`（內含可重用的 `ServiceFilterDropdown.vue`），過濾 predicate 放在 `app/utils/serviceFilters.ts`，而**環繞 predicate 的 reactive 接線**放在 `app/composables/useServiceListFilters.ts`。

理由：三個清單頁面的**外殼差異很大**——`index.vue` 有 New Service 與刪除對話框，`system.vue` 另有 Admin Mode banner 與帶 log checkbox 的刪除對話框，`ReadOnlyServiceList.vue` 有權限警告 banner。要把三者塞回單一容器元件並不划算（`3c4ed37` 正是為此把 `system.vue` 拆出去的）。但**過濾行為必須完全一致**。把 predicate 抽成純函式，讓三個頁面各自保有外殼、共用同一份邏輯，是唯一同時滿足這兩個條件的切法。

替代方案：讓三個頁面都改回共用容器元件並以 slot 客製 —— 會把 `3c4ed37` 刻意拆開的東西重新綁回去，且 slot 數量會多到失去可讀性。

**第三層（composable）的必要性**：初版只抽出 predicate，結果三個頁面仍各自複製兩個 ref 加兩個 computed 的接線。純函式只保證「過濾結果一致」，不保證「接線一致」—— 新增第三個 filter 維度時仍要記得改三處，且沒有編譯器把關。`useServiceListFilters(services, searchQuery)` 補上這個缺口。

它必須是 **factory（每次呼叫回傳全新的 ref）**，而非 `useAdminMode` / `useSettings` 那種 module-level singleton：那兩者刻意是全域 session 狀態，但 filter 選取是每頁各自的選擇，跨頁共用會讓兩個清單面互相干擾。這點有專門的測試釘住。

**`ServiceFilterDropdown.vue` 的抽出**：Status 與 Type 兩個 dropdown 原本是兩塊近乎全等的樣板，差別只在選項表、testid 前綴、標籤與 handler。任何 dropdown 的調整（ARIA、鍵盤操作、hover 樣式）都得手動套用兩次且保持一致。抽成單一元件後由 prop 帶入差異；父層仍持有選取狀態與「哪個選單開啟」，維持同時只有一個選單開啟。

### Type 選項對齊 launch-policy badge（含 KeepAlive）

Type 選項為 Scheduled / KeepAlive / RunAtLoad / None，前三項對應 `ServiceRow.vue` 在 Type 欄位渲染的三種 badge。

關鍵差異：**badge 依優先序（`schedule` > `keepAlive` > `runAtLoad`）只顯示一個標籤，filter 則匹配所有適用的選項**。同時有 schedule 與 runAtLoad 的服務，badge 顯示 `Scheduled`，但選 Scheduled 或 RunAtLoad 都應該找得到它。兩者刻意不同：badge 要的是單一簡潔標示，filter 要的是不漏掉。

`None` 是前三者的否定，而非獨立欄位。必須排除 `keepAlive.enabled`，否則 KeepAlive 服務（`runAtLoad === false`）會落入 None。

替代方案：讓 filter 也照 badge 優先序只匹配單一分類 —— 會讓「選 RunAtLoad 卻找不到有排程且 runAtLoad 的服務」，違反使用者對 filter 的預期。

### Status 選項以原始 status 字串為 value

option 的 `value` 直接用 `Service.status` 的原始字串（`running` / `loaded` / `stopped` / `unknown`），`label` 才是使用者看到的文案。

理由：使用者看到的「Unloaded」對應的是 `status === 'stopped'`。把這個對應放進選項表的一列，predicate 就只是 `statusFilter.includes(service.status)`，不需要任何分支。若改用自訂 value 再於 predicate 內轉換，這個對應會散落成程式碼分支，日後新增狀態容易漏改。

`StatusFilterValue` 直接取自 `Service['status']`，`TypeFilterValue` 則是獨立的 union。兩者不只用於選項表，也貫穿 `ServiceFilters`、三個 predicate、composable 的 ref 與 `ServiceFilterBar` 的 props/emits —— 拼錯的字面值（`'stoped'`、`'sheduled'`）在編譯期就會被擋下（TS2820 並提示正確拼法），而不是靜默匹配到零筆。`ServiceFilterDropdown` 以 `generic="T extends string"` 保持對兩種 union 皆可重用。

### 空選取代表 All，構成 cross-filter AND logic

空陣列代表該 filter「未啟用」（顯示全部），而非「匹配零筆」。

理由：這正是讓 Status、Type、Search 三者能以 AND 串接、又各自可獨立略過的關鍵。`filterServices` 在三者皆未啟用時直接回傳原陣列，避免 400+ 服務的無謂複製。

### 三個清單頁面各自整合（index.vue / system.vue / ReadOnlyServiceList.vue）

三個檔案分別引入 `ServiceFilterBar`，並各自呼叫 `useServiceListFilters(services, searchQuery)` 取得 `statusFilter` / `typeFilter` / `hasActiveFilter` / `filteredServices`。頁面本身不再宣告任何 filter state —— 這一點由集合測試反向斷言（頁面原始碼不得出現 `const statusFilter = ref`）。

理由：如 Context 第 2 點，這三者是獨立檔案而非單一元件的三個使用者。任何清單 UI 的新增都必須三處都接。

空狀態一併處理：filter 啟用時清單為空，代表「filter 排除了全部」而非「使用者還沒有服務」，因此顯示 "No services match the selected filters"，且 `index.vue` 必須隱藏「Create your first service」按鈕。

### 以回歸測試釘住三個清單頁面的共用集合

新增 `app/pages/__tests__/serviceListFilterBar.test.ts`，把三個清單「面」當成一組來驗證：都掛了 filter bar、位置在 header 與 table header 之間、都走共用 `filterServices`、都有區隔的空狀態；另外單獨驗證 `apple-system.vue` 仍委派給 `ReadOnlyServiceList`。

理由：這正是 Context 第 2 點那類疏漏的唯一防線。per-component 測試看不見「某個頁面脫離了共用集合」——每個元件自己都測得過，缺的是集合層級的斷言。若 `apple-system.vue` 哪天不再用 `ReadOnlyServiceList`，針對該元件的斷言會靜默失去覆蓋，所以那條委派關係也要一併釘住。

## Implementation Contract

**Behavior**

- 三個清單頁面（User `/`、System `/system`、Apple System `/apple-system`）在 page header 與 table header row 之間顯示 filter bar，內含 `Status:` 與 `Type:` 兩個 dropdown trigger。
- 未選任何項目時 trigger 顯示 `All`；選一項顯示該項 label；選多項顯示 `N selected`。有任一 filter 啟用時，額外出現 `Clear all`。
- 同一 dropdown 內多選為 OR；Status、Type 與既有搜尋框三者為 AND。
- 選取後選單保持開啟（可連續多選）；點擊元件外部或按 Escape 關閉；同時只有一個選單開啟。
- filter 啟用且結果為空時，顯示 `No services match the selected filters`；`index.vue` 在此情況不得顯示「Create your first service」按鈕。
- filter 狀態不持久化，重新載入或切換頁面即回到未選取。

**Interface / data shape**

`app/utils/serviceFilters.ts` 匯出：

- `type StatusFilterValue = Service['status']`、`type TypeFilterValue = 'scheduled' | 'keepAlive' | 'runAtLoad' | 'none'`、`interface FilterOption<T extends string> { value: T; label: string }`
- `STATUS_FILTER_OPTIONS: readonly FilterOption<StatusFilterValue>[]` — 依序 `running`/Running、`loaded`/Loaded、`stopped`/Unloaded、`unknown`/Unknown
- `TYPE_FILTER_OPTIONS: readonly FilterOption<TypeFilterValue>[]` — 依序 `scheduled`/Scheduled、`keepAlive`/KeepAlive、`runAtLoad`/RunAtLoad、`none`/None
- `hasActiveFilter(statusFilter: readonly StatusFilterValue[], typeFilter: readonly TypeFilterValue[]): boolean` — 「是否有 dropdown 正在收窄清單」的唯一定義（**不含**文字搜尋）
- `matchesStatusFilter(service: Service, statusFilter: readonly StatusFilterValue[]): boolean`
- `matchesTypeFilter(service: Service, typeFilter: readonly TypeFilterValue[]): boolean`
- `filterServices(services: Service[], filters: ServiceFilters): Service[]`，其中 `ServiceFilters = { searchQuery: string; statusFilter: readonly StatusFilterValue[]; typeFilter: readonly TypeFilterValue[] }`

`app/composables/useServiceListFilters.ts` 匯出 `useServiceListFilters(services: Ref<Service[]>, searchQuery: Ref<string>)`，回傳 `{ statusFilter, typeFilter, hasActiveFilter, filteredServices }`。**每次呼叫回傳全新的 ref**（factory，非 singleton）。

`ServiceFilterBar.vue` 的 props 為 `statusFilter: StatusFilterValue[]`、`typeFilter: TypeFilterValue[]`，emit `update:statusFilter` / `update:typeFilter`（帶對應的 union 陣列），支援 `v-model:status-filter` / `v-model:type-filter`。元件**必須 emit 新陣列，不得就地修改傳入的 prop**。

`ServiceFilterDropdown.vue` 以 `generic="T extends string"` 宣告，props 為 `{ label: string; testid: string; options: readonly FilterOption<T>[]; selected: readonly T[]; open: boolean }`，emit `toggleMenu` / `select(value: T)`。testid 依 `testid` prop 組成：`${testid}-filter-trigger`、`${testid}-filter-menu`、`${testid}-option-${value}`。

Type 判定：`scheduled` = `service.schedule` 非 null/undefined；`keepAlive` = `service.keepAlive?.enabled === true`；`runAtLoad` = `service.runAtLoad === true`；`none` = 三者皆不成立。

**無障礙**：trigger 帶 `aria-haspopup="listbox"`、`aria-expanded`（隨開闔更新）、`aria-controls="${testid}-filter-menu"`；選單為 `role="listbox"` + `aria-multiselectable="true"`；選項為 `role="option"` + `aria-selected`；勾選圖示為 `aria-hidden="true"`。所有 button 皆須標 `type="button"`。

**Failure modes**

- 空選取（`[]`）代表未啟用、顯示全部，而非匹配零筆。
- `keepAlive` 物件缺失（`undefined`）視為非 KeepAlive，並可落入 `none`——不得拋錯。
- 未知的 filter 選項字串一律不匹配（`default: return false`），不拋錯。
- 搜尋字串在比對前 trim 並轉小寫；純空白等同未輸入。
- 三個 filter 皆未啟用時 `filterServices` 直接回傳原陣列（不複製）。

**Acceptance criteria**

- `make test` exit 0、`make lint` exit 0。
- `app/utils/__tests__/serviceFilters.test.ts` 涵蓋選項表順序、Status OR、Unloaded→stopped 對應、四種 Type 判定、「KeepAlive 服務不得被歸類為 None」、`keepAlive` 缺失、以及 Status/Type/Search 三者 AND 組合與空結果。
- `app/utils/__tests__/serviceFilters.test.ts` 另涵蓋 `hasActiveFilter` 的三態（皆空為 false、任一非空為 true、不受文字搜尋影響）。
- `app/components/__tests__/ServiceFilterBar.test.ts` 涵蓋四個 Status 與四個 Type 選項的渲染、選單互斥開啟、多選與取消選取、trigger 文案三態（All / 單項 label / `N selected`）、不修改傳入 prop、Clear all，以及 ARIA 合約（`aria-haspopup` / `aria-expanded` 隨開闔更新 / `role="listbox"` / `aria-selected` / 所有 button 皆為 `type="button"`）。每個測試以 `attachTo: document.body` 掛載，**必須在 `afterEach` 卸載**——元件在 `onMounted` 註冊 document 層級的 listener，不卸載會讓 listener 與 DOM 節點跨測試累積（同 repo 的 `ServiceLogs.test.ts` 已有此慣例）。
- `app/composables/__tests__/useServiceListFilters.test.ts` 涵蓋初始狀態、filter 變動與 services 變動時的重新計算、搜尋與 dropdown 的 AND 組合，以及**每次呼叫回傳獨立狀態**（factory 而非 singleton）。
- `app/pages/__tests__/serviceListFilterBar.test.ts` 對三個清單面逐一斷言 filter bar 的存在與位置、共用 `useServiceListFilters` 的使用（並反向斷言頁面不得自行宣告 `const statusFilter = ref`）、區隔的空狀態，並斷言 `apple-system.vue` 仍委派 `ReadOnlyServiceList`。位置斷言錨定在 `data-testid="service-list-table-header"` 這個真實 markup 上，而非原始碼註解。此測試的有效性須經反向驗證：移除任一頁面的整合後該測試必須轉紅。
- 型別安全須可驗證：以 `matchesStatusFilter(s, ['stoped'])` 之類的拼錯字面值探針執行 `pnpm nuxi typecheck`，必須報 TS2820。

**Scope boundaries**

**In scope**：`ServiceFilterBar.vue`、`ServiceFilterDropdown.vue`、`serviceFilters.ts`、`useServiceListFilters.ts` 及其測試；`index.vue`、`system.vue`、`ReadOnlyServiceList.vue` 三處的 filter bar 掛載、改用共用 composable、`data-testid="service-list-table-header"` 錨點、空狀態文案；`README.md` 與 `.claude/CLAUDE.md` 的對應更新。

**Out of scope**：`ServiceRow.vue` 的 badge 邏輯與優先序；後端 `Service` 型別與任何 Wails binding；filter 狀態持久化；服務詳情頁；排序功能；i18n。

## Risks / Trade-offs

- **[Type filter 與 badge 再度脫節]** → 若日後新增第四種 launch policy（如新的 `ServiceRow` badge），Type 選項不會自動跟上。緩解：design 與 `.claude/CLAUDE.md` 均明記兩者必須對齊，且 `TYPE_FILTER_OPTIONS` 的註解點名這層對應關係；但目前沒有測試能自動偵測 badge 新增，這是已知殘留風險。
- **[新增第四個清單頁面時遺漏 filter bar]** → `serviceListFilterBar.test.ts` 只釘住目前這三個面，新頁面不會自動納入。緩解：`.claude/CLAUDE.md` 明記三個面是獨立檔案、清單 UI 需三處都接；新增頁面時須同步擴充該測試的 `LIST_SURFACES`。
- **[以原始碼字串比對做測試]** → `serviceListFilterBar.test.ts` 讀 `.vue` 原始碼比對字串，而非掛載頁面。取捨：三個頁面在 `onMounted` 會呼叫 Wails binding、用到 `NuxtLink` 與 `useAdminMode`，完整掛載成本高且脆弱（同 repo 的 `pages/services/__tests__/edit-launch-policy.test.ts` 已明文記錄避開這條路），且抽出邏輯到 host 元件並不能驗證「頁面樣板是否接上元件」這件真正會壞的事。斷言已全數錨定在真實 markup（`<ServiceFilterBar`、`data-testid="service-list-table-header"`）與程式碼識別字（`useServiceListFilters(...)`）上，不再依賴原始碼註解，但仍對大幅重構敏感。已透過反向驗證確認非空轉測試。
- **[大清單的過濾成本]** → Apple System 有 400+ 服務，每次輸入或選取都重跑一次 `filterServices`。實測可接受（純陣列走訪，無子行程），且三者皆未啟用時走 early return —— 該 early return 以共用的 `hasActiveFilter` 表達，不是第二份手動維護的「無 filter」定義。小型固定長度陣列（各 4 項）刻意不轉 `Set`，那屬於過早最佳化。若日後成為瓶頸再考慮 memoize。
