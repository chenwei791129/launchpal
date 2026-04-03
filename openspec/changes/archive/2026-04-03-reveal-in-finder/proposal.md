## Why

LaunchPal 目前只能複製 plist 路徑。使用者需要手動開啟 Finder 並導覽到該路徑才能找到檔案，增加一個按鈕可以一鍵定位 plist 檔案位置。

## What Changes

- 後端新增 `RevealInFinder(path string)` 方法，執行 `open -R <path>` 在 Finder 中顯示指定檔案
- 前端 `ServiceSummary.vue` 的 Plist File 區塊，在 copy icon 右側新增「在 Finder 中顯示」按鈕
- 按鈕使用 `@click.stop` 防止觸發外層的複製行為
- 三種服務類型（User / System / Apple System）皆適用

## Capabilities

### New Capabilities

- `reveal-in-finder`: 提供「在 Finder 中顯示」功能，讓使用者可從服務詳情頁一鍵在 Finder 中定位 plist 檔案

### Modified Capabilities

（無）

## Impact

- 受影響程式碼：`app.go`（新增 Wails binding）、`frontend/app/components/ServiceSummary.vue`（新增按鈕）
