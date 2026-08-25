## 1. 共用 Filter Bar 元件

對應設計決策：共用 Filter Bar 元件與 filterServices 過濾邏輯分離；Status 選項以原始 status 字串為 value；Type 選項對齊 launch-policy badge（含 KeepAlive）。

- [x] [P] 1.1 建立 `frontend/app/components/ServiceFilterBar.vue`，包含 Status multi-select dropdown 和 Type multi-select dropdown，props 接收 `statusFilter: string[]` 和 `typeFilter: string[]`，emit `update:statusFilter` 和 `update:typeFilter` 事件。Status 選項為 Running / Loaded / Unloaded / Unknown；Type 選項為 Scheduled / KeepAlive / RunAtLoad / None。預設顯示 "All"，選取後顯示已選項目數量或名稱。

## 2. 整合 Filter 至 User Services 頁面

對應設計決策：三個清單頁面各自整合（index.vue / system.vue / ReadOnlyServiceList.vue）；空選取代表 All，構成 cross-filter AND logic。

- [x] 2.1 在 `frontend/app/pages/index.vue` 引入 `ServiceFilterBar`，放置在 header 與 table header 之間（filter bar placement and shared component）。新增 `statusFilter` 和 `typeFilter` reactive state。

- [x] 2.2 修改 `index.vue` 的 `filteredServices` computed，加入 status multi-select dropdown filter 邏輯：當 `statusFilter` 非空時，僅顯示 status 在選取清單中的服務（Running→running, Loaded→loaded, Unloaded→stopped, Unknown→unknown）。

- [x] 2.3 在 `index.vue` 的 `filteredServices` computed 加入 type multi-select dropdown filter 邏輯：Scheduled 匹配 `service.schedule` 已定義、KeepAlive 匹配 `service.keepAlive?.enabled === true`、RunAtLoad 匹配 `service.runAtLoad === true`、None 匹配三者皆不成立。

- [x] 2.4 確認 `index.vue` 中 Status、Type、Search 三者為 cross-filter AND logic：服務必須同時滿足所有啟用的 filter 才顯示。

## 3. 整合 Filter 至 ReadOnlyServiceList 元件（Apple System）

對應設計決策：三個清單頁面各自整合（index.vue / system.vue / ReadOnlyServiceList.vue）。

- [x] 3.1 在 `frontend/app/components/ReadOnlyServiceList.vue` 引入 `ServiceFilterBar`，放置在 header 與 table header 之間（filter bar placement and shared component）。新增 `statusFilter` 和 `typeFilter` reactive state。

- [x] 3.2 修改 `ReadOnlyServiceList.vue` 的 `filteredServices` computed，加入 status multi-select dropdown filter 與 type multi-select dropdown filter 邏輯，行為與 `index.vue` 一致。

- [x] 3.3 確認 `ReadOnlyServiceList.vue` 中 Status、Type、Search 三者為 cross-filter AND logic。

## 4. 整合 Filter 至 System Services 頁面

對應設計決策：三個清單頁面各自整合（index.vue / system.vue / ReadOnlyServiceList.vue）；以回歸測試釘住三個清單頁面的共用集合。

`system.vue` 在 `3c4ed37 feat(admin-mode)`（2026-04-23，本 change 建立之後）已從 `ReadOnlyServiceList` 拆出為獨立頁面，因此第 3 節並未涵蓋它。

- [x] 4.1 在 `frontend/app/pages/system.vue` 引入 `ServiceFilterBar`，放置在 header 與 table header 之間。新增 `statusFilter` / `typeFilter` state，並改用共用的 `filterServices`，行為與另外兩個頁面一致。

- [x] 4.2 新增回歸測試，確認三個清單頁面（index / system / ReadOnlyServiceList）都掛上 filter bar 且都走共用 `filterServices`，避免日後再有頁面脫離共用集合而無人察覺。

## 5. 重構與強化（/simplify 與 /code-review 之後）

對應設計決策：共用 Filter Bar 元件與 filterServices 過濾邏輯分離；Status 選項以原始 status 字串為 value；三個清單頁面各自整合（index.vue / system.vue / ReadOnlyServiceList.vue）；以回歸測試釘住三個清單頁面的共用集合。

- [x] 5.1 新增 `serviceFilters.ts` 的 `hasActiveFilter`，取代三個頁面與 filter bar 中重複四次的相同判斷；`filterServices` 的 early return 也改用它，使「無 filter」只有一份定義。

- [x] 5.2 新增 `app/composables/useServiceListFilters.ts`（factory，每次呼叫回傳全新 ref），收攏三個頁面重複的 `statusFilter` / `typeFilter` ref 與 `hasActiveFilter` / `filteredServices` computed；補上狀態獨立性的測試。

- [x] 5.3 抽出 `ServiceFilterDropdown.vue`，消除 Status 與 Type 兩塊近乎全等的樣板；testid 由 prop 前綴組成，既有 13 則元件測試在測試檔未改動的情況下須全數通過。

- [x] 5.4 將集合測試的位置斷言從原始碼註解 `<!-- Table header -->` 改為錨定真實 markup `data-testid="service-list-table-header"`，並把「走共用邏輯」的斷言改為 `useServiceListFilters(...)` 加上「頁面不得自行宣告 `const statusFilter = ref`」的反向斷言。

- [x] 5.5 將 `StatusFilterValue` / `TypeFilterValue` union 貫穿 `ServiceFilters`、三個 predicate、composable 的 ref 與 `ServiceFilterBar` 的 props/emits；`ServiceFilterDropdown` 改用 `generic="T extends string"` 保持可重用。以拼錯字面值探針驗證 typecheck 會報 TS2820。

- [x] 5.6 為 `ServiceFilterDropdown` 補上 ARIA 語意（`aria-haspopup` / `aria-expanded` / `aria-controls` / `role="listbox"` / `aria-multiselectable` / `role="option"` / `aria-selected`，勾選圖示 `aria-hidden`）與所有 button 的 `type="button"`，並新增 4 則測試釘住。

- [x] 5.7 為 `ServiceFilterBar.test.ts` 補上 `afterEach` 卸載，避免 `attachTo: document.body` 掛載的 document listener 與 DOM 節點跨測試累積（比照 `ServiceLogs.test.ts` 既有慣例）。
