## 1. CI Workflow 修改

- [x] [P] 1.1 修改 `.github/workflows/build.yml`：新增安裝 `create-dmg` 步驟（`brew install create-dmg`），將 "Package application" 步驟從 `tar -czf launchpal.app.gz launchpal.app` 改為執行 `create-dmg` 產生 DMG image（含 Applications folder shortcut），並將 Upload artifact 的路徑改為 `.dmg` — 涵蓋 DMG image generation、Applications folder shortcut、CI workflow produces DMG
- [x] [P] 1.2 修改 `.github/workflows/release-please.yml`：將 `gh release upload` 的檔案從 `launchpal.app.gz` 改為 `LaunchPal.dmg`，同步更新 Download App Artifact 的 artifact name — 涵蓋 Release workflow uploads DMG

## 2. 本地開發工具

- [x] 2.1 在 `Makefile` 新增 `dmg` target：先執行 `wails build`，再執行 `create-dmg` 產生 DMG 到 `build/bin/` 目錄。若 `create-dmg` 未安裝則失敗並顯示錯誤訊息 — 涵蓋 Local DMG build target、create-dmg not installed locally
