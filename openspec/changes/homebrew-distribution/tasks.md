## 1. 建立 homebrew-apps repo

- [x] 1.1 在 GitHub 建立 `chenwei791129/homebrew-apps` repo，包含 `Casks/` 目錄和 `README.md`
- [x] 1.2 [P] 建立 cask formula provides standard Homebrew installation：撰寫 `Casks/launchpal.rb`，包含 version、url、sha256、name、homepage 等欄位，cask formula 使用 postflight 移除 quarantine（postflight removes quarantine attribute，`xattr -dr com.apple.quarantine`），caveats inform user about quarantine removal（說明未簽名與自動移除 quarantine），以及 cask supports uninstallation 的 uninstall stanza（quit + delete）

## 2. 跨 repo 認證設定

- [x] 2.1 使用 GitHub Personal Access Token (Fine-grained) 進行跨 repo 寫入：建立 `HOMEBREW_TAP_TOKEN`，實現 cross-repo authentication uses fine-grained PAT，scope 限定 `chenwei791129/homebrew-apps` 的 `contents:write` 權限
- [x] 2.2 [P] 將 `HOMEBREW_TAP_TOKEN` 加入 `chenwei791129/launchpal` repo 的 Actions secrets

## 3. Release workflow 自動更新 cask formula

- [x] 3.1 修改 `.github/workflows/release-please.yml`，實現 release workflow updates cask formula automatically：在 upload-release-asset job 之後新增 `update-homebrew` job（決定不使用 repository_dispatch 觸發 homebrew-apps 更新，而是直接 push），使用 `HOMEBREW_TAP_TOKEN` 認證，下載 release 的 DMG、計算 SHA256、更新 `homebrew-apps` repo 中 `Casks/launchpal.rb` 的 version 和 sha256，commit & push 到 homebrew-apps main branch
- [x] 3.2 確保 cask update step 失敗不影響其他 release 步驟（DMG 上傳至 GitHub Release 已先完成）

## 4. README 更新

- [x] 4.1 [P] README documents Homebrew installation：更新 `README.md`，加入 Homebrew 安裝說明，包含完整指令 `brew install --cask chenwei791129/apps/launchpal` 和兩步式 `brew tap chenwei791129/apps` + `brew install --cask launchpal`

## 5. 驗證

- [x] 5.1 本地測試 cask formula 安裝：`brew install --cask chenwei791129/apps/launchpal`，確認 app 安裝至 `/Applications/` 且無 Gatekeeper 攔截
- [x] 5.2 本地測試 cask 解除安裝：`brew uninstall --cask launchpal`，確認 app 被移除
- [x] 5.3 模擬 release 流程，確認 workflow 能正確更新 homebrew-apps 的 cask formula
