## Why

目前 LaunchPal 的發布格式為 `.app.gz`（tar 壓縮），使用者需要解壓後手動移動到 Applications 資料夾。這不符合 macOS 應用的標準安裝體驗——大部分 macOS 應用都使用 `.dmg` 格式，開啟後拖曳到 Applications 即可完成安裝。

## What Changes

- 將打包格式從 `.app.gz` 改為 `.dmg`
- DMG 內包含 `LaunchPal.app` 與指向 `/Applications` 的 symlink，提供標準的拖曳安裝體驗
- 使用 `create-dmg` 開源工具產生 DMG 映像檔
- 修改 CI/CD workflow（build.yml、release-please.yml）以產生並上傳 `.dmg`
- 在 Makefile 新增 `dmg` target 方便本地測試

## Non-Goals

- **Code Signing / Notarization**：目前不包含 Apple 開發者簽名與公證流程，這是獨立的後續工作
- **多格式發布**：不同時提供 `.app.gz` 和 `.dmg`，僅發布 `.dmg`
- **自訂 DMG 背景圖**：初期使用 `create-dmg` 預設樣式，後續可視需要美化

## Capabilities

### New Capabilities

- `dmg-packaging`: 使用 `create-dmg` 將 macOS 應用打包為 `.dmg` 格式，提供標準拖曳安裝體驗

### Modified Capabilities

（無）

## Impact

- 受影響的檔案：
  - `.github/workflows/build.yml` — 打包步驟從 `tar` 改為 `create-dmg`
  - `.github/workflows/release-please.yml` — 上傳的 asset 從 `.app.gz` 改為 `.dmg`
  - `Makefile` — 新增 `dmg` target
- 受影響的依賴：新增 `create-dmg` 工具（CI 環境需安裝）
- 對使用者的影響：Release 下載物從 `.app.gz` 變為 `.dmg`，安裝體驗改善
