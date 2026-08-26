## Why

Go 1.27 已於 2026 年 8 月釋出，專案目前停在 go 1.26.0。單純的版本落後不急迫，但這次升級帶有一個必須同步處理的副作用：Go 1.27 起官方要求 macOS 13 Ventura 以上，linker 預設會把產出 binary 的 LC_BUILD_VERSION minos 標記寫成 13.0（本機實測確認，Go 1.26 為 12.0）。

這個標記不只是宣告，dyld 會在載入時強制檢查。而專案的 `build/darwin/Info.plist` 至今仍宣告 LSMinimumSystemVersion 為 10.13.0（Wails 模板預設值，從未更新過），`README.md` 則完全沒有系統需求章節。三個數字互不一致的狀態已經存在（現行 binary 為 12.0，宣告為 10.13.0），升級只是把落差從「排除 macOS 11」擴大到「排除 macOS 11 至 12」。

落差的後果是使用者體驗層面的：Launch Services 依 Info.plist 的 10.13.0 放行啟動，dyld 隨後依 minos 拒絕載入，使用者看到的是模糊的「應用程式無法開啟」而非明確的版本需求訊息。對這個未簽章的 app 特別有害——README 已有整節在教使用者處理 Gatekeeper 攔截，使用者最可能把這個失敗誤判為簽章問題而反覆嘗試清除 quarantine。

## What Changes

- **BREAKING**：最低系統需求正式提高至 macOS 13 Ventura。macOS 12 Monterey 及更早版本不再支援。
- go.mod 的 go directive 由 1.26.0 提升至 1.27.0。CI 透過 go-version-file 指向 go.mod，會自動取得對應 toolchain，無需修改 workflow。
- `Makefile` 明確釘死 wails 建置的 macOS deployment target 為 13.0。Go 1.27 的 13.0 預設只在內部連結時生效，而主 binary 因 Wails 的 darwin ObjC 後端走外部連結，minos 改由 Wails 硬編碼的 `-mmacosx-version-min=10.13`（clamp 後 11.0）決定；不釘死則宣告值會高於實際 minos。CI 走 `make build`，故一併涵蓋。
- `build/darwin/Info.plist` 與 `build/darwin/Info.dev.plist` 的 LSMinimumSystemVersion 由 10.13.0 校正為 13.0.0，與 Go 1.27 產出的 minos 對齊。
- `README.md` 新增系統需求章節，明確標示 macOS 13 Ventura 或以上。
- Homebrew cask 加入 macOS 版本相依宣告，讓不符合的系統在安裝階段就被擋下，而非安裝完成後才發現無法啟動。
- 新增回歸測試，將「Info.plist 宣告值」與「預期的最低版本」綁定為可驗證的斷言。這是本次改動中唯一具備長期價值的部分：10.13.0 之所以能停留數年無人察覺，正是因為沒有任何機制在它與實際產出脫節時發出警告。

## Non-Goals

- **不採用 linker 的 macOS 版本覆寫選項來維持 macOS 12 支援。** Go 1.27 新增的 linker 選項確實可以把 minos 覆寫回 12.0（已實測驗證），但它改變的只是版本標記，不改變 runtime 行為。Go 停止支援某個 macOS 版本意味著 runtime 從此可自由使用該版本以上的 libSystem API 而不視為 bug，踩到時的症狀是執行期失敗且不會有 upstream 修復。本專案的 launchpal-privhelper 以 root 權限執行，把未經支援的組態帶進 root 程序的代價，超過它換回的不確定使用者數量。
- **不暫緩升級以維持現有支援範圍。** minos 12.0 這條線已經把 macOS 11 使用者排除在外且無人回報問題，延後只是把同一個決定推遲到下一個 Go 版本。
- **不調整 CI workflow。** go-version-file 機制已經涵蓋 toolchain 連動。
- **不處理程式碼層面的 Go 1.27 適應性修改。** 相容性已在討論階段完整實測：建置、測試、golangci-lint、Wails CLI v2.15.0 在 go directive 設為 1.27.0 的狀態下全數通過，無需任何原始碼變更。

## Capabilities

### New Capabilities

- `macos-minimum-version`: 定義 LaunchPal 支援的最低 macOS 版本，以及該版本必須在哪些位置被一致地宣告（app bundle 的 Info.plist、README 使用者文件、Homebrew cask 相依條件），並要求這些宣告與 Go toolchain 實際產出的 binary 載入需求對齊。

### Modified Capabilities

- `homebrew-cask-formula`: cask formula 新增 macOS 版本相依宣告，使安裝階段即可攔截不符合最低系統需求的環境。

## Impact

- Affected specs: `macos-minimum-version`（新增）、`homebrew-cask-formula`（修改）
- Affected code:
  - Modified:
    - `go.mod`
    - `Makefile`
    - `build/darwin/Info.plist`
    - `build/darwin/Info.dev.plist`
    - `README.md`
    - `.claude/CLAUDE.md`
  - New:
    - `build_metadata_test.go`
  - Removed: （無）
- 跨 repo 影響：Homebrew cask formula 位於 chenwei791129/homebrew-apps，需另行手動提交，不在本 repo 的檔案異動範圍內。
- 使用者影響：macOS 12 Monterey 及更早版本的使用者無法再升級到後續版本。此變更需進入 CHANGELOG，因此相關 commit 使用 fix 前綴而非 chore。
