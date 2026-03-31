## Context

LaunchPal 目前透過 GitHub Release 發布 DMG 檔。使用者需要手動下載、拖曳安裝、再處理 Gatekeeper 攔截（因為沒有 code signing）。目標是提供 `brew install --cask chenwei791129/apps/launchpal` 一鍵安裝體驗。

涉及兩個 repo：
- `chenwei791129/launchpal`（現有，產生 release）
- `chenwei791129/homebrew-apps`（新建，存放 cask formula）

## Goals / Non-Goals

**Goals:**

- 使用者可透過一行 Homebrew 指令安裝 LaunchPal
- 安裝後自動移除 quarantine，不需手動處理 Gatekeeper
- Release 時自動同步更新 cask formula 的版本號與 SHA256
- homebrew-apps repo 可供未來其他應用程式共用

**Non-Goals:**

- Code signing / notarization（不加入 Apple Developer Program）
- 提交至官方 homebrew-cask
- 建立 Homebrew formula（CLI 工具用），僅建立 cask（桌面應用用）

## Decisions

### 使用 GitHub Personal Access Token (Fine-grained) 進行跨 repo 寫入

release-please workflow 需要寫入 homebrew-apps repo 來更新 cask formula。`GITHUB_TOKEN` 僅限當前 repo 權限，無法跨 repo 操作。

**方案比較：**

| 方案 | 優點 | 缺點 |
|------|------|------|
| Fine-grained PAT | 可限定只授權 homebrew-apps repo 的 contents:write | 需手動建立、有過期時間（最長 1 年） |
| GitHub App | 更安全、不綁個人帳號 | 設定複雜、對個人專案 overkill |
| 手動更新 | 零設定 | 容易忘記、版本不同步 |

**決定**：使用 Fine-grained PAT，scope 限定為 `chenwei791129/homebrew-apps` repo 的 `contents:write` 權限。存為 launchpal repo 的 secret `HOMEBREW_TAP_TOKEN`。

### 使用 repository_dispatch 觸發 homebrew-apps 更新

**方案比較：**

| 方案 | 優點 | 缺點 |
|------|------|------|
| launchpal workflow 直接 push 到 homebrew-apps | 簡單、一個 workflow 搞定 | cask 模板邏輯散落在 launchpal repo |
| repository_dispatch 觸發 homebrew-apps 自己的 workflow | 關注點分離、cask 模板由 homebrew-apps 管理 | 多一個 workflow 要維護 |
| GitHub Actions reusable workflow | 可重用 | 跨 repo reusable workflow 設定複雜 |

**決定**：由 launchpal 的 release workflow 直接 push 更新到 homebrew-apps。原因：目前只有一個 app，關注點分離的好處不大；直接 push 更簡單且減少 moving parts。未來 app 數量增加時可再遷移到 repository_dispatch。

具體流程：
1. release-please 建立 release tag
2. build workflow 產生 DMG 並上傳至 release
3. 新增步驟：下載 DMG → 計算 SHA256 → 用 `sed` 更新 homebrew-apps 中的 `launchpal.rb` → commit & push

### Cask formula 使用 postflight 移除 quarantine

在 cask formula 中加入 `postflight` stanza，安裝後自動執行 `xattr -dr com.apple.quarantine`。這在自建 tap 中是合法的（官方 tap 禁止，但第三方 tap 無此限制）。

同時在 `caveats` 中說明此行為，讓使用者知道發生了什麼。

## Risks / Trade-offs

- **[PAT 過期]** → Fine-grained PAT 最長 1 年有效期，過期後 release 時自動更新會失敗。緩解：設定到期提醒；workflow 失敗時會有 GitHub notification。
- **[Quarantine 移除的信任問題]** → 部分使用者可能對自動移除 quarantine 感到不安。緩解：在 caveats 和 README 中清楚說明原因。
- **[SHA256 計算時機]** → 必須在 DMG 上傳至 release 後才能取得最終 SHA256。緩解：workflow 中先上傳 DMG，再從 release 下載計算 SHA256，確保一致性。
