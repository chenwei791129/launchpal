## 1. 完整性驗證與雜湊共用邏輯

- [x] 1.1 [P] 於 internal/privhelper/integrity.go 新增共用函式:計算檔案 SHA-256,以及對指定路徑做「Ownership and permission verification of protected copy」—— 判定是否為一般檔案(非 symlink)、擁有者 UID 為 0、且 `mode & 022 == 0`;涵蓋「完整性失敗拒絕」所需的判定基礎。驗證:單元測試對 root 擁有 0755、非 root 擁有、group/other 可寫、symlink、不存在各案例回傳正確判定。
- [x] 1.2 [P] 定義受保護路徑常數 `/Library/Application Support/LaunchPal/launchpal-privhelper` 為單一事實來源,供 GUI 解析與 helper 安裝共用(對應 design「受保護路徑選定」)。驗證:GUI 與 helper 端引用同一常數,`go build ./...` 通過。

## 2. helper 自我安裝(root context)

- [x] 2.1 於 internal/privhelper/install.go 實作「Root-owned protected helper copy」:以 root 將 helper「自身執行映像」(以 `os.Executable()` 開啟並加 `O_NOFOLLOW`,非重讀 bundle 路徑,以緩解 F3 TOCTOU)複製到受保護路徑,設定 `root:wheel` mode `0755`,建立缺少的父目錄為 `root:wheel` `0755`,對最終寫入路徑元件以 O_NOFOLLOW 等價保護拒絕 symlink,且在受保護複本已與自身相符時為 idempotent(對應 design「helper 自我安裝」)。驗證:單元測試涵蓋首次安裝、已相符時不複製、symlink 目標拒絕三情境。
- [x] 2.2 在 cmd/launchpal-privhelper/main.go 啟動流程中、綁定 socket 前呼叫自我安裝,且僅在自身路徑 != 受保護路徑常數時執行;實作「Helper self-install does not block Admin Mode」—— 安裝失敗僅記錄不阻斷,當次仍以已啟動複本提供服務。驗證:helper 測試模擬安裝失敗時仍完成 socket 服務啟動。

## 3. GUI 解析優先序(與 pin/bundle 解耦)

- [x] 3.1 於 admin_mode.go 改寫 resolveHelperPath 實作「Launch prefers verified protected copy」,維持 `func() (string, error)` 簽章不變(不新增回傳值,故不影響 `helperPath func()` 注入點):有效受保護複本存在時一律回傳受保護路徑,唯一例外是「非空 pin、bundle 可讀、`bundleHash==pin` 且 `bundleHash!=protectedHash`」時回傳 bundle 路徑重新佈署;無有效受保護複本時走首次安裝路徑(對應 design「解析優先序決策」)。驗證:單元測試涵蓋 (a) 有效受保護+空 pin→受保護、(b) 有效受保護+bundle 缺失→受保護(不報錯,修正 F4 DoS)、(c) 竄改 bundle 不符 pin→受保護(修正 F1 降級)、(d) 合法更新→bundle、(e) 無受保護+bundle 不符 pin→錯誤。
- [x] 3.2 更新既有 admin_mode_test.go 中對 `helperPath` 欄位的注入賦值與相關測試案例,使其涵蓋解耦後的優先序結果;因簽章維持 `func() (string, error)` 故不需改欄位型別(修正 code-review F5)。驗證:`go test ./...` 全綠,既有 Enable 測試不因簽章變更而編譯失敗。

## 4. bundle 雜湊釘選與(連結前)注入

- [x] 4.1 [P] 在 main.go 新增套件層級釘選變數並實作「Bundle helper hash pinning」比對:僅在需啟動 bundle 複本的窗口以非空 pin 比對其 SHA-256,不符則不啟動;pin 為空時跳過 bundle 雜湊閘且絕不因此繞過已存在的有效受保護複本(對應 design「雜湊釘選注入」)。驗證:單元測試涵蓋相符、不符、pin 為空三情況,且驗證空 pin + 有效受保護複本仍選受保護。
- [x] 4.2 [P] 修正 pin 注入時機(code-review F2):重排 Makefile 的 build / build-debug 為「build-helper → 計算 helper SHA-256 → `wails build` 帶 `-ldflags -X main.<pinVar>=<sha256>` → 複製 helper 進 bundle」,使 pin 在主程式連結前注入。驗證:`make build` 後於 `.app` 主二進位以 `go tool nm` 或執行期讀取確認 pin 變數為非空且等於 bundle helper 的 SHA-256。
- [x] 4.3 [P] 於 .github/workflows/build.yml 的 "Build application" 步驟,在 `go tool wails build` 的 ldflags 追加 `-X main.<pinVar>=<sha256>`(兩個版本分支皆需),並於該步驟前建置 helper 並計算 SHA;既有 `main.version` 注入位於此 CI 步驟而非 Makefile。驗證:CI 建置產出的出貨二進位其 pin 變數為非空(以建置日誌或下載產物驗證)。

## 5. 完整性失敗拒絕與收尾

- [x] 5.1 在 admin_mode.go 的 Enable 實作「Integrity failure refuses helper launch」:無有效受保護複本且非空 pin 不符 bundle(或 bundle 缺失)時,不執行 osascript、不啟動 helper,經 failFromRequesting 回到 Disabled 帶錯誤碼 `helper_integrity_failed`,訊息不洩漏額外內部細節(對應 design「完整性失敗拒絕」)。驗證:單元測試確認該情況下 osascript 未被呼叫且狀態為 Disabled + helper_integrity_failed。
- [x] 5.2 對應「Helper binary packaged in app bundle」修改後語意,確認 Enable 呼叫端消費 resolveHelperPath 回傳的啟動路徑並傳入既有 `LaunchHelperOptions.HelperPath`,不改動 internal/privhelper/client.go 的 LaunchHelper 介面(修正 code-review F6)。驗證:admin_mode 測試確認 Enable 以解析所得路徑呼叫 launchFn,且 client.go 無 diff。
- [x] 5.3 更新 README.md 與 .claude/CLAUDE.md 說明 Admin Mode 首次啟用會將 helper 佈署至 root 擁有的受保護路徑、bundle 雜湊釘選為防禦縱深(且不綁定受保護複本可用性)、以及 bootstrap window 與自我安裝 TOCTOU 殘餘風險。驗證:內容審閱確認文件與 design.md 的 Risks 一致。
- [ ] 5.4 請使用者手動驗證:首次啟用 Admin Mode 後 `/Library/Application Support/LaunchPal/launchpal-privhelper` 為 `root:wheel 0755`;第二次啟用不再觸發自我安裝;刪除 bundle helper 後仍能啟用 Admin Mode;`brew install` 全程無 sudo 提示。驗證:使用者回報四項觀察結果符合。
