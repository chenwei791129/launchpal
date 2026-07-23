## Why

啟用 Admin Mode 時,LaunchPal 會透過 osascript 以 root 身分執行 `launchpal-privhelper`,而在此之前 `resolveHelperPath` 只做 `os.Stat` 的存在性檢查,完全沒有驗證這個二進位的擁有者、寫入權限或完整性。由於 App 未簽章、常見安裝位置(admin group 可寫的 `/Applications`、使用者自有的 `~/Applications`)下 bundle 為使用者可寫,同 UID、非 root 的惡意程式可以先覆寫 bundle 內的 helper,再靜待使用者下次基於正當理由啟用 Admin Mode —— 屆時被竄改的二進位便以 root 執行(plant-and-wait 提權)。這是 security audit run-1 的 Finding 1(HIGH),且與「Admin Mode 開啟時同 UID 可驅動 root RPC」的既知設計風險不同:它不需要攻擊者在 session 進行中在線,能直接把「使用者偶爾使用 Admin Mode」轉化為「攻擊者取得 root」。

因為使用者沒有 Apple Developer Program,`SMAppService`、notarization、以及 OS 強制的 code-signing requirement 都不可行。本變更採用不依賴付費帳號、且不改變使用者可見行為的信任錨點:把信任綁定在一個「非 root 攻擊者無法偽造」的成品 —— 一份 root 擁有、使用者不可寫的 helper 複本。

## What Changes

- 新增 `privileged-helper-integrity` capability,涵蓋:
  - **受保護路徑安裝(Option C,主信任錨點)**:當 helper 從 bundle 以 root 啟動時,自我複製到 `/Library/Application Support/LaunchPal/launchpal-privhelper`,並設為 `root:wheel`、移除 group/other 寫入權限(`0755`);父目錄同樣建立為 root 擁有 `0755`。
  - **解析優先受保護複本(與 pin/bundle 解耦)**:`resolveHelperPath` 在有效受保護複本存在時一律優先啟動它(驗證為一般檔案、非 symlink、擁有者 UID 0、非 group/other 可寫),不論 bundle 是否存在或被竄改、pin 是否為空。受保護複本的可用性不綁定 bundle 雜湊,避免「清空 pin 觸發降級」與「刪 bundle 造成 DoS」。
  - **bundle helper 雜湊釘選(Option B,防禦縱深)**:build 時把 helper 的 SHA-256 以 ldflags 於主程式**連結之前**注入;僅在需要啟動 bundle 複本的窗口(首次安裝、合法更新後重新佈署)以非空 pin 比對磁碟 bundle helper,不符則拒絕啟動並回報明確錯誤。
  - **App 更新後重新佈署**:僅當有效受保護複本存在、且非空 pin 證明 bundle 為與受保護複本不同的合法新版時,才改啟動 bundle 複本以重新安裝受保護複本;空 pin 或竄改的 bundle 都不會觸發回退。
  - **完整性失敗即拒絕**:首次安裝時非空 pin 不符 bundle,Enable 不啟動 helper,回到 Disabled 並帶明確錯誤碼。
- 修改 `privileged-helper-lifecycle` capability:helper 路徑解析語意由「取 bundle 內同層 launchpal-privhelper」改為「有效受保護複本存在時優先啟動它,僅在無有效受保護複本或非空 pin 證明合法更新時使用 bundle 複本」。
- 建置流程:重排 `Makefile` 的 `build` / `build-debug`(先 build-helper → 算 SHA → wails build 帶 pin ldflags → 複製進 bundle),並在 `.github/workflows/build.yml` 的 wails build 步驟追加 pin ldflags(既有版本注入位於 CI 而非 Makefile)。
- 使用者可見行為零變更:`brew install` 不需 sudo;受保護路徑安裝寄生在既有那一次 Admin Mode osascript 授權內,授權次數不變。

## Non-Goals

（本變更會建立 design.md,scope 邊界與已拒絕方案記於該處。）

## Capabilities

### New Capabilities

- `privileged-helper-integrity`: 在以 root 啟動特權 helper 前,對其做完整性與擁有權驗證,並以 root 擁有的受保護路徑複本作為信任錨點;涵蓋受保護路徑安裝、解析優先序、bundle 雜湊釘選、更新後重新佈署、完整性失敗拒絕。

### Modified Capabilities

- `privileged-helper-lifecycle`: helper 執行期路徑解析改為在有效受保護複本存在時優先啟動它,bundle 複本降為首次安裝與合法更新時的來源。

## Impact

- Affected specs:
  - New: `privileged-helper-integrity`
  - Modified: `privileged-helper-lifecycle`
- Affected code:
  - Modified:
    - admin_mode.go（resolveHelperPath 改為解耦的優先序決策,簽章維持 `func() (string, error)`;Enable 新增完整性失敗路徑）
    - cmd/launchpal-privhelper/main.go（helper 啟動時從自身執行映像自我安裝至受保護路徑）
    - main.go（新增 build 時注入的 pin 套件層級變數,比照既有 `main.version`）
    - Makefile（重排 build / build-debug:build-helper → 算 SHA → wails build 帶 pin ldflags → 複製進 bundle）
    - .github/workflows/build.yml（出貨 DMG 的 wails build 步驟追加 pin ldflags,helper 於該步驟前建置並計算 SHA）
    - admin_mode_test.go（既有 `helperPath` 注入賦值隨解析邏輯調整測試案例;簽章不變故不需改型別）
  - New:
    - internal/privhelper/install.go（從自身執行映像安裝至受保護路徑並設定 root 擁有權的共用邏輯）
    - internal/privhelper/integrity.go（SHA-256 計算、擁有者/權限驗證、受保護路徑常數的共用邏輯）
- 殘餘風險(記於 design.md):首次安裝與每次合法更新後的重新佈署仍會短暫執行/信任 bundle 複本(bootstrap window,含自我安裝 TOCTOU);Option B 的釘選值存於使用者可寫的主程式,故屬防禦縱深而非錨點,但清空 pin 不會繞過已存在的有效受保護複本;notarization/Gatekeeper 仍不可行,cask 移除 quarantine 的做法維持不變。
