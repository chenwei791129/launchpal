## Why

目前 LaunchPal 僅提供 GitHub Release 下載 DMG 的安裝方式。macOS 使用者普遍習慣透過 Homebrew 安裝應用程式，提供 `brew install --cask` 安裝方式可以降低安裝門檻、簡化更新流程。

由於目前不加入 Apple Developer Program（無 code signing / notarization），app 會被 macOS Gatekeeper 攔截。透過自建 Homebrew Tap 搭配 postflight 自動移除 quarantine attribute，可以讓使用者安裝後直接開啟，無需手動處理。

## What Changes

- 在 GitHub 建立 `chenwei791129/homebrew-apps` repo 作為 Homebrew Tap，供所有應用程式共用
- 建立 `Casks/launchpal.rb` cask formula，包含 `postflight` 自動執行 `xattr -dr com.apple.quarantine`
- 修改 `release-please.yml` workflow，在 release 時自動更新 homebrew-apps repo 的 cask formula（版本號 + SHA256）
- 更新 `README.md` 加入 Homebrew 安裝指令

## Non-Goals

- **加入 Apple Developer Program**：不做 code signing / notarization，使用 postflight xattr workaround
- **提交至官方 homebrew-cask**：官方要求 notarization，不符合現況
- **支援 Linux / Windows**：LaunchPal 是 macOS 專屬工具

## Capabilities

### New Capabilities

- `homebrew-cask-formula`: Homebrew Cask formula 定義，包含 postflight quarantine 移除、版本資訊、安裝路徑等設定
- `homebrew-auto-release`: Release 時自動更新 homebrew-apps repo 的 cask formula，確保版本號與 SHA256 同步

### Modified Capabilities

（無）

## Impact

- 新增外部 repo：`chenwei791129/homebrew-apps`（含 `Casks/launchpal.rb`）
- 受影響的檔案：
  - `.github/workflows/release-please.yml` — 新增自動更新 cask formula 的步驟
  - `README.md` — 新增 Homebrew 安裝說明
- 受影響的依賴：需要一個有權限寫入 homebrew-apps repo 的 GitHub token（PAT 或 GitHub App）
