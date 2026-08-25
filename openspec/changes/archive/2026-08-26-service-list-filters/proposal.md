## Why

目前三個服務清單頁面（User / System / Apple System）僅提供文字搜尋，當服務數量龐大時（如 Apple System Services 有 400+ 個），用戶難以快速定位特定狀態或類型的服務。新增 filter 功能可以大幅提升瀏覽效率。

## What Changes

- 在三個服務清單頁面的 header 與 table header 之間新增 filter bar
- 提供 **Status** multi-select dropdown，選項：Running / Loaded / Unloaded / Unknown
- 提供 **Type** multi-select dropdown，選項：Scheduled / KeepAlive / RunAtLoad / None（對齊 `ServiceRow.vue` Type 欄位的三種 launch-policy badge）
- 同一 dropdown 內的選項使用 OR 邏輯（如選 Running + Loaded → 顯示 running 或 loaded）
- 跨 dropdown 使用 AND 邏輯（如 Status=Running + Type=Scheduled → 只顯示正在跑且有排程的服務）
- Filter 與現有的 Search 搜尋列同時作用（AND）
- 抽出共用 `ServiceFilterBar.vue` 元件、`serviceFilters.ts` 過濾邏輯與 `useServiceListFilters` composable（reactive 接線），供 `index.vue`、`system.vue`、`ReadOnlyServiceList.vue` 三個清單頁面共用
- filter 啟用而結果為空時，空狀態顯示 "No services match the selected filters"，與「使用者尚無服務」區隔；`index.vue` 在此情況隱藏「Create your first service」按鈕

## Capabilities

### New Capabilities

- `service-list-filters`: 服務清單頁面的 multi-select dropdown 篩選功能，包含 Status 與 Type 兩種篩選維度

### Modified Capabilities

（無）

## Impact

- 受影響程式碼：
  - `frontend/app/pages/index.vue` — 整合 filter bar 與 filteredServices 邏輯（User）
  - `frontend/app/pages/system.vue` — 整合 filter bar 與 filteredServices 邏輯（System；Admin Mode 之後已從 ReadOnlyServiceList 拆出為獨立頁面）
  - `frontend/app/components/ReadOnlyServiceList.vue` — 整合 filter bar 與 filteredServices 邏輯（Apple System）
  - 新增 `frontend/app/components/ServiceFilterBar.vue` — 共用 filter bar 元件
  - 新增 `frontend/app/components/ServiceFilterDropdown.vue` — 可重用的單一 multi-select dropdown（Status 與 Type 共用，含 ARIA 語意）
  - 新增 `frontend/app/utils/serviceFilters.ts` — 共用過濾 predicate 與型別，三個頁面共用同一份邏輯
  - 新增 `frontend/app/composables/useServiceListFilters.ts` — 共用 reactive 接線（factory，每頁狀態獨立）
  - 新增 `frontend/app/utils/__tests__/serviceFilters.test.ts` — 過濾 predicate 的單元測試
  - 新增 `frontend/app/components/__tests__/ServiceFilterBar.test.ts` — filter bar 元件的互動與 ARIA 測試
  - 新增 `frontend/app/composables/__tests__/useServiceListFilters.test.ts` — composable 的反應性與狀態獨立性測試
  - 新增 `frontend/app/pages/__tests__/serviceListFilterBar.test.ts` — 釘住三個清單頁面都掛上 filter bar 的回歸測試
  - `README.md` — Features 清單新增 filter 說明
  - `.claude/CLAUDE.md` — 新增「Service List Filters」章節，並註記三個清單面為獨立檔案
