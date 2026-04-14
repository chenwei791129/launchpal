## 1. 修復 Summary tab 捲動

- [x] 1.1 在 `frontend/app/pages/services/[name].vue` 第 143 行，將 `<ServiceSummary>` 包裹在 `<div class="h-full overflow-auto">` 中（或直接在 `<ServiceSummary>` 外層加上相同的 wrapper），使 Summary tab content is scrollable，與 Edit tab（第 146 行）和 Inspect tab（第 251 行）的捲動行為一致
