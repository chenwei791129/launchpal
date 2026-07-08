## Context

啟用 Admin Mode 時,`admin_mode.go` 的 `resolveHelperPath` 僅以 `os.Executable()` 定位 bundle 內同層的 `launchpal-privhelper`,並只做一次 `os.Stat` 存在性檢查,隨即由 `internal/privhelper/client.go` 的 `LaunchHelper` 以 osascript `do shell script ... with administrator privileges` 將該路徑以 root 執行。過程中沒有任何擁有權、寫入權限或完整性驗證,`Ping` 握手也只回 `{pong:true}`,無法辨識 helper 是否被替換。

App 未簽章,`Makefile` 的 `install-helper` 只是 `cp` + `chmod 0755`。在 admin group 可寫的 `/Applications` 或使用者自有的 `~/Applications` 安裝下,bundle 內 helper 為使用者可寫。同 UID、非 root 攻擊者可先覆寫 bundle helper,再等待使用者下次正當啟用 Admin Mode 取得 root(plant-and-wait)。使用者無 Apple Developer Program,故 `SMAppService` / notarization / OS 強制驗簽皆不可用。

實測目標路徑父層權限:`/Library` 為 `root:wheel drwxr-xr-x`,`/Library/Application Support` 為 `root:admin drwxr-xr-x`(group 無寫入)—— 兩層皆僅 root 可寫入,適合作為非 root 攻擊者無法偽造的信任錨點。

## Goals / Non-Goals

**Goals:**

- 讓穩態(首次啟用之後、且未經 App 更新的期間)下,以 root 執行的 helper 一律來自一份非 root 攻擊者無法竄改的受保護複本,終結穩態 plant-and-wait。
- 在信任 bundle 複本(首次安裝、更新後重新佈署)之前,以 build 時釘選的 SHA-256 擋掉最低階的無腦覆寫。
- 對使用者可見行為零變更:`brew install` 不需 sudo,Admin Mode 授權次數與提示不變。

**Non-Goals:**

- 不實作 Option A(自簽 code-signing 憑證 + `SecStaticCodeCheckValidity` + cgo 綁 Security.framework)。成本(CI 私鑰管理、cgo)較高,列為未來防禦縱深。
- 不追求關閉 bootstrap window(見 Risks):首次啟用與每次 App 更新後的重新佈署,本質上必須先執行一次 bundle 複本。
- 不處理 notarization / Gatekeeper —— 仍不可行,cask 移除 quarantine 的既有做法不變。
- 不改動 Admin Mode 狀態機的既有狀態或 `admin-mode` capability 的既有需求(避免與已 park 的 admin-mode-lifecycle-hardening 變更衝突);完整性失敗以新 capability 的需求描述,錯誤碼透過既有 `failFromRequesting` 管道呈現。

## Decisions

### 受保護路徑選定

受保護複本安裝於 `/Library/Application Support/LaunchPal/launchpal-privhelper`。父目錄 `/Library/Application Support/LaunchPal` 由 helper 以 root 建立為 `root:wheel`、mode `0755`;binary 本身 `root:wheel`、mode `0755`(owner rwx、group/other r-x,無 group/other 寫入)。因兩層祖先目錄皆僅 root 可寫,非 root 攻擊者無法建立、替換或寫入此路徑。

替代方案:`/usr/local/...`(Homebrew 前綴,常為使用者/admin 可寫,不安全);bundle 內就地 `chmod go-w`(bundle 本身使用者可寫,無效)。均否決。

### helper 自我安裝

以 root 執行的 helper 在啟動(`cmd/launchpal-privhelper/main.go`)時執行 idempotent 的自我安裝:僅在自身執行路徑不等於受保護路徑常數(即從 bundle 啟動)時觸發;若受保護複本不存在或其 SHA-256 與自身不同,則複製到受保護路徑,設定擁有者 `root:wheel` 與 mode `0755`,建立缺少的父目錄。

複製來源為 helper 自身的執行映像(啟動時以 `os.Executable()` 開啟並加 `O_NOFOLLOW`),不重新讀取 bundle 路徑 —— 避免在「GUI 完成 pin 檢查」與「helper 讀取來源」之間出現第二個替換窗口(緩解 code-review F3 的 TOCTOU;寫入端 chmod/chown 亦對最終路徑元件加 `O_NOFOLLOW`)。此 TOCTOU 無法在無驗簽下完全消除,因為一旦被替換的 helper 已以 root 執行,持久化與否都已失守 —— 其邊界與下述 bootstrap window 相同,如實記於 Risks。

選擇 helper 端自我安裝而非新增 GUI 觸發的 RPC:特權檔案寫入留在 root 行程內、不新增 RPC 攻擊面、天然 idempotent。替代方案(新增 `InstallProtectedHelper` RPC)否決:增加特權介面且與啟動時機競爭。

### 解析優先序決策

`resolveHelperPath`(GUI,非特權)維持 `func() (string, error)` 簽章不變(不新增回傳值,避免破壞既有 `admin_mode_test.go` 對 `helperPath func()` 欄位的注入賦值 —— 修正 code-review F5),回傳「要啟動的路徑」或錯誤。信任錨點僅來自「root 擁有」,**不綁定 bundle 雜湊**。

令 `pin` = build 時注入的釘選值(可能為空,開發建置);`protected` = 通過驗證的受保護複本(一般檔案、非 symlink、擁有者 UID 0、`mode & 022 == 0`),否則視為不存在:

1. 若 `protected` 有效:
   - **僅當** `pin` 非空、可讀到 bundle helper、`bundleHash == pin`、且 `bundleHash != protectedHash` 時 → 判定為合法 App 更新,回傳 bundle 路徑(交由 helper 重新佈署)。
   - 其餘所有情況(`pin` 為空 / bundle 缺失或不可讀 / bundle 不符 pin / bundle 與受保護複本相符)→ 回傳受保護複本路徑。
2. 若 `protected` 無效(首次安裝):
   - `pin` 非空且(bundle 缺失或 `bundleHash != pin`)→ 回傳完整性錯誤 `helper_integrity_failed`,不啟動。
   - 否則(pin 相符,或 pin 為空的開發建置)→ 回傳 bundle 路徑(交由 helper 自我安裝)。

關鍵不變式:**只要有效的受保護複本存在,唯一改用 bundle 的情況是「非空 pin 證明 bundle 為合法新版」**。空 pin、缺 bundle、被竄改的 bundle 都不會使受保護複本被繞過或停用。這同時封住「攻擊者清空主程式的 pin 以強迫回退到惡意 bundle」的降級路徑(修正 code-review F1)與「刪除/破壞 bundle 使有效受保護複本無法啟用」的 Admin Mode DoS(修正 code-review F4)。

替代方案(把受保護複本的啟動綁定在「其雜湊必須等於 bundle 雜湊」)否決:那會讓清空 pin 或竄改 bundle 反而繞過 root 錨點,並讓刪除 bundle 造成 DoS —— 即 code-review F1/F4 的根因。

### 雜湊釘選注入

pin 注入必須在主程式**連結之前**發生。目前 `install-helper` 在 `wails build` 之後才執行(且只做 cp),因此無法用它注入 pin(修正 code-review F2)。正確順序為:先 `build-helper` 產生 helper → 計算其 SHA-256 → 執行 `wails build` 時以 `-ldflags -X main.<pinVar>=<sha256>` 連結進主程式 → 再把 helper 複製進 bundle。

需同時調整兩處實際出貨路徑:

- `Makefile` 的 `build` / `build-debug`:重排為「build-helper → 計算 SHA → wails build 帶 pin ldflags → 複製 helper 進 bundle」。
- `.github/workflows/build.yml` 的 "Build application" 步驟:出貨 DMG 由 CI 建置,其 `go tool wails build -ldflags "-X main.version=$VERSION"` 需在同一 ldflags 追加 `-X main.<pinVar>=<sha256>`,且 helper 須在該步驟前建置並計算 SHA。既有的版本注入位於 CI 的 wails build ldflags(非 Makefile),計畫不再誤植為「沿用 Makefile 既有機制」。

誠實界定:pin 存於使用者可寫的主程式二進位,能寫 helper 的攻擊者通常也能 patch 主程式改掉 pin —— 故 pin 為防禦縱深,只擋首次安裝/重新佈署窗口的無腦覆寫,不是錨點。錨點是 root 擁有的受保護複本,其可用性不依賴 pin。

### 完整性失敗拒絕

當釘選比對失敗或受保護複本擁有者/權限驗證失敗時,`Enable` 不執行 osascript、不啟動 helper,經 `failFromRequesting` 回到 `Disabled` 並帶明確錯誤碼(如 `helper_integrity_failed`),與既有 `helper_binary_not_found` / `helper_handshake_failed` 並列。錯誤訊息不洩漏內部路徑細節之外的資訊。

## Implementation Contract

**行為**:

- 首次啟用 Admin Mode(受保護複本尚不存在):GUI 以非空 pin 驗證 bundle helper 雜湊 → osascript 授權(既有那一次)→ 啟動 bundle helper(root)→ helper 自我安裝受保護複本。使用者無額外提示。
- 後續啟用(有效受保護複本存在):GUI 直接以 osascript 啟動受保護複本 —— 不論 bundle 是否存在、是否被竄改、pin 是否為空,受保護複本一律優先且不被繞過。
- App 更新後首次啟用(有效受保護複本存在,但非空 pin 證明 bundle 為與受保護複本不同的合法新版):GUI 啟動 bundle helper → helper 以新版覆寫受保護複本。
- 竄改/降級偵測:受保護複本非 root 擁有 / 可被 group/other 寫入 / 為 symlink → 視為無效;首次安裝時 bundle 不符非空 pin → 拒絕啟動並回到 Disabled 帶 `helper_integrity_failed`。

**介面 / 資料形狀**:

- `resolveHelperPath` 維持 `func() (string, error)` 簽章不變,回傳要啟動的路徑或 `helper_integrity_failed` 錯誤;不新增回傳值,故 `helperPath func()` 注入點與既有測試賦值不受影響。啟動來源(受保護 vs bundle)不需傳給 `LaunchHelper` —— 既有 `LaunchHelperOptions.HelperPath` 已是不透明路徑參數,故 `internal/privhelper/client.go` 無需改動(修正 code-review F6)。
- helper 端新增自我安裝進入點,於 `main` 啟動流程中在綁定 socket 前呼叫;僅在自身路徑 != 受保護路徑常數時執行;失敗僅記錄不阻斷正常服務(佈署失敗不應使當次 Admin Mode 不可用)。
- 釘選值為主程式套件層級 `string` 變數,build 時以 `-ldflags -X` 注入;未注入(開發建置)時 pin 為空:此時「有效受保護複本存在 → 一律啟動受保護複本、永不因空 pin 觸發回退 bundle」,故空 pin 不構成降級路徑。
- 受保護路徑常數為單一事實來源,GUI(解析)與 helper(安裝)共用。

**失敗模式**:

- 首次安裝時完整性驗證失敗(非空 pin 不符 bundle)→ `Disabled` + `helper_integrity_failed`,不執行 osascript。
- helper 自我安裝失敗 → 記錄於 helper 端;當次仍以已啟動的複本提供服務,受保護複本留待下次補齊(不使 Admin Mode 失效)。
- 有效受保護複本存在但 bundle 缺失/不可讀 → 啟動受保護複本(不視為錯誤、不 DoS)。

**驗收準則**:

- 單元測試:擁有權/權限驗證函式對(root 擁有 0755)、(非 root 擁有)、(group/other 可寫)、(symlink)、(不存在)各案例回傳正確判定。
- 單元測試:解析優先序 —— (a) 有效受保護複本存在 + 空 pin → 選受保護;(b) 有效受保護複本 + bundle 缺失 → 選受保護(不報錯);(c) 有效受保護複本 + 非空 pin 且 bundle 符 pin 且與受保護不同 → 選 bundle(重新佈署);(d) 無受保護複本 + 非空 pin 且 bundle 不符 → `helper_integrity_failed`;(e) 無受保護複本 + bundle 符 pin(或空 pin)→ 選 bundle。
- 單元測試:雜湊釘選比對在相符/不符/釘選為空三種情況的行為。
- 手動驗證(使用者執行):首次啟用後 `/Library/Application Support/LaunchPal/launchpal-privhelper` 為 `root:wheel 0755`;第二次啟用不再觸發自我安裝;刪除 bundle helper 後仍能啟用 Admin Mode;`brew install` 全程無 sudo 提示。

**Scope 邊界**:

- In scope:受保護路徑安裝、解析優先序(含解耦不變式)、擁有權/權限驗證、bundle 雜湊釘選與(連結前)注入、更新後重新佈署、完整性失敗拒絕。
- Out of scope:Option A 自簽驗簽;Admin Mode 狀態機既有狀態/需求變更;`privileged-helper-rpc` 的既有 RPC 驗證器;`internal/privhelper/client.go` 的 `LaunchHelper` 介面(無需改動);已 park 的兩個變更(admin-mode-lifecycle-hardening、filesystem-input-hardening)之範圍。

## Risks / Trade-offs

- [Bootstrap window:首次安裝與每次 App 更新後的重新佈署,仍必須先執行一次使用者可寫的 bundle 複本] → 以非空 pin 閘擋無腦覆寫;把窗口降到「使用者這輩子第一次啟用」與「每次更新後第一次啟用」,不再是每次啟用。無法在無付費帳號下完全關閉,如實記錄。
- [釘選值 / 驗證邏輯位於使用者可寫的主程式,理論上可被 patch] → 承認為防禦縱深;真正的錨點是 root 擁有的受保護複本,其可用性不依賴 pin,非 root 攻擊者無法偽造。清空 pin 不會繞過已存在的有效受保護複本(見解析優先序不變式)。
- [helper 自我安裝失敗導致穩態仍跑 bundle] → 自我安裝失敗僅記錄不阻斷;下次啟用重試佈署;期間若已有有效受保護複本則仍優先啟動它。
- [自我安裝 TOCTOU(code-review F3):bundle 在 GUI 檢查後、helper 讀取前被替換,可能把被竄改的映像持久化到受保護路徑] → helper 複製來源為自身執行映像(非重讀 bundle),消除第二個替換窗口;殘餘窗口與 bootstrap window 同界 —— 一旦被替換的 helper 已以 root 執行,持久化與否都已失守,無法在無驗簽下進一步關閉。

## Migration Plan

- 既有安裝首次升級到含本變更的版本後,第一次啟用 Admin Mode 會自動佈署受保護複本,無需使用者動作。
- 回退:移除本變更後,`resolveHelperPath` 退回僅解析 bundle;殘留的 `/Library/Application Support/LaunchPal/` 不影響舊版運作(舊版不讀該路徑),可留存或由使用者手動清除。

## Open Questions

- 無。受保護路徑、pin 注入時機(連結前、Makefile 與 CI build.yml 兩處)、解耦不變式與失敗語意皆已定;實作時沿用既有 `O_NOFOLLOW` helper,並比照 CI 現有 `main.version` 的 wails build ldflags 注入方式追加 pin 變數。
