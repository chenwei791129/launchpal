## Context

專案目前的 go directive 為 1.26.0。Go 1.27 於 2026 年 8 月釋出，官方同時宣告要求 macOS 13 Ventura 以上，並且 linker 預設會把產出 binary 的 LC_BUILD_VERSION minos 標記寫為 13.0。

本機實測的建置矩陣（各組合皆實際產出 binary 並以 otool 檢視）：

| 建置組合 | 連結方式 | minos | sdk |
| --- | --- | --- | --- |
| Go 1.26.1，CGO 關閉或開啟 | 內部 | 12.0 | 12.0 |
| Go 1.27.0，CGO 關閉或開啟 | 內部 | 13.0 | 26.2 |
| Go 1.27.0 加上 linker 的 macOS 版本覆寫選項 | 內部 | 12.0 | 26.2 |
| Go 1.27.0 + `-tags desktop,production`，沿用 Wails 注入的 `-mmacosx-version-min=10.13` | 外部 | **11.0** | 26.2 |
| Go 1.27.0 + `-tags desktop,production`，`-mmacosx-version-min=13.0` | 外部 | 13.0 | 26.2 |
| Go 1.27.0 + `-linkmode=external`，不帶任何 version-min flag | 外部 | **15.0**（建置機 SDK） | 26.2 |

**Go 的預設值只在內部連結時成立。** 上表前三列都是純 `go build` 的內部連結，`LC_BUILD_VERSION` 由 Go 自己寫入，CGO flag 完全被忽略。但實際的 app 走的是 `go tool wails build`：`desktop,production` tags 會把 Wails 的 darwin Objective-C 後端編進來，強制外部連結，此時 minos 改由 clang/ld 依 `-mmacosx-version-min` 決定，Go 的預設值不參與。

Wails v2.15.0 在 `pkg/commands/build/base.go:313` 與 `:352` 對 darwin 建置硬性注入 `-mmacosx-version-min=10.13` 到 `CGO_CFLAGS` / `CGO_LDFLAGS`（arm64 下被 clamp 到 11.0），因此主 binary 與 Go 版本無關地停在 11.0。`launchpal-privhelper` 由 Makefile 以純 `go build` 產出、無 ObjC 也無 tags，走內部連結，才拿到 Go 誠實的 13.0。這正是現況「主 binary 11.0、helper 12.0」兩個數字不同的真正成因——它不是 Go 版本差異造成的。

升級 Wails 無法迴避此事：v2.15.0 已是 v2 最新版，硬編碼就在其中；v3（beta）雖已把該值移出 CLI、改由專案自己的 `build/darwin/Taskfile.yml` 持有，但 scaffold 的預設值是 12.0（`internal/commands/build_assets/darwin/Taskfile.yml` 並有測試釘住），仍需專案明確設定。最後一列同時說明「乾脆不設」也不可行——外部連結下 clang 會退回建置機的 SDK 版本，使 minos 隨建置機而異。

專案現況存在三個互不一致的版本數字：`build/darwin/Info.plist` 宣告 LSMinimumSystemVersion 為 10.13.0（Wails 模板預設，從未修改）；現有 app bundle 主 binary 的 minos 為 11.0；以 Go 1.26.1 建置的 launchpal-privhelper 的 minos 為 12.0。`README.md` 沒有任何系統需求章節。

不一致的執行後果是分工造成的：Launch Services 讀 Info.plist 決定是否允許啟動，dyld 讀 LC_BUILD_VERSION 決定是否允許載入。宣告值偏低時，前者放行、後者攔截，使用者得到的是模糊的啟動失敗而非明確的版本訊息。

相容性已於討論階段完整實測：在 go directive 暫時設為 1.27.0 的狀態下，全部套件的 go build、go test（含 Go 1.27 新預設啟用的 stdversion vet 檢查）、golangci-lint 皆通過且零 issue，Wails CLI v2.15.0 亦可由 Go 1.27 正常編譯。無任何原始碼需要修改。

## Goals / Non-Goals

**Goals:**

- 將 go directive 提升至 1.27.0，並確保 CI 自動取得對應 toolchain。
- 讓最低 macOS 版本在所有對外宣告位置（app bundle、README、Homebrew cask）取得一致，且與 Go toolchain 實際產出的載入需求對齊。
- 建立自動化機制，使宣告值日後再度脫節時能被測試攔截，而非再次沉默數年。

**Non-Goals:**

- 不使用 linker 的 macOS 版本覆寫選項維持 macOS 12 支援（理由見決策一）。
- 不暫緩升級。minos 12.0 這條線已排除 macOS 11 使用者且無人回報問題，延後只是把同一決定推遲到下個 Go 版本。
- 不修改 CI workflow。go-version-file 指向 go.mod 的機制已涵蓋 toolchain 連動。
- 不進行任何原始碼適應性修改，相容性已實測確認無需變更。
- 不驗證實際產出 binary 的 minos 值（理由與補償措施見決策三與風險段落）。

## Decisions

### 採用 macOS 13 為新基準，不以 linker 選項覆寫版本標記

Go 1.27 新增的 linker macOS 版本選項可將 minos 覆寫回 12.0，實測確認有效。此方案被否決。

覆寫改變的只是版本標記，不改變 runtime 行為。Go 停止支援某個 macOS 版本的實質意義是：runtime 自此可自由使用該版本以上才有的 libSystem API，且不再視為 bug。一旦踩到，症狀為執行期的 symbol 找不到或未定義行為，且不會有 upstream 修復。

對本專案代價尤其高：launchpal-privhelper 以 root 權限執行。把 Go 官方不支援的組態帶進 root 程序所增加的不確定性，超過它換回的使用者數量——而該數量本身也無資料支持。

替代方案「暫緩升級停留於 1.26」同樣被否決：現行 minos 12.0 早已排除 macOS 11 使用者，此決定的本質只會在下個 Go 版本重演。

### Info.plist 宣告值校正為 13.0.0 並與 Go toolchain 對齊

`build/darwin/Info.plist` 與 `build/darwin/Info.dev.plist` 的 LSMinimumSystemVersion 由 10.13.0 改為 13.0.0。

選 13.0.0 而非其他值的依據是：它必須等於 Go 1.27 linker 寫入 binary 的 minos，否則就重建了目前這個「Launch Services 放行、dyld 攔截」的落差。宣告值高於 minos 會無謂排除可執行的系統；低於 minos 則重現模糊失敗。兩者相等是唯一正確狀態。

這兩個檔案已納入版本控制，Wails 僅在檔案不存在時才生成模板，因此修改不會於後續建置被覆寫。

### 於 Makefile 明確釘死 Wails 建置的 macOS deployment target

`Makefile` 新增單一來源的 `MACOS_MIN_VERSION := 13.0`，經 `WAILS_CGO_ENV` 以 `CGO_CFLAGS` / `CGO_LDFLAGS` 傳入 `build` 與 `build-debug` 兩個 target 的 wails 建置。

此決策為實作階段的實測結果所迫：原本預期 Go 1.27 的 13.0 預設會自動涵蓋所有產出，實際上它只涵蓋內部連結的 `launchpal-privhelper`，主 binary 因外部連結而由 Wails 的 10.13（clamp 後 11.0）決定（成因見 Context）。若不處理，本次變更會把 Info.plist 校正為 13.0.0，卻讓主 binary 停在 11.0——宣告高於 minos，是 requirement 明文禁止的另一種不一致。

三個候選中選擇釘死而非其他兩者：

- **把 Info.plist 改成 11.0 以遷就實際 minos**（tasks 原先寫下的退路）：不可行。兩個 binary 的 minos 本就不同（11.0 與 13.0），不存在單一的「實際值」可遷就；且 helper 為 13.0，macOS 11/12 上 Admin Mode 仍會被 dyld 擋下，而主 binary 內的 Go 1.27 runtime 實際上也不支援這些系統。結果會完整重建本變更要消滅的「Launch Services 放行、載入後失敗」落差。
- **移除 Wails 注入的 flag**：不可行。外部連結下 clang 會改用建置機的 SDK 版本（實測為 15.0），minos 將隨建置機而異，比現況更糟。
- **釘死為 13.0**：採用。使兩個 binary 與宣告值三者相等。

需與決策一區分：決策一拒絕的是「為保住 macOS 12 而把版本標記往下壓，讓標記與 runtime 實際支援範圍脫節」。此處方向相反——是把一個既存的、由 Wails 預設造成的向下覆寫，校正回 Go runtime 真正要求的樓地板。兩者都服從同一條原則：版本標記必須誠實反映 runtime 的實際需求。

代價是專案從此自行持有一個建置參數，Wails 日後若改變預設不會自動反映。此代價由 `.claude/CLAUDE.md` 記錄的 Go 升級檢查步驟承接。

### 以 plist 解析的回歸測試鎖定最低版本宣告

新增 `build_metadata_test.go`（package main），讀取兩個 Info.plist 檔案，斷言 LSMinimumSystemVersion 等於預期的最低版本常數。

解析方式選用專案既有的 howett.net/plist 依賴，而非字串或正則比對。此選擇需要驗證，因為兩個檔案含有 Go template 指令（例如條件區塊）夾雜於 XML 元素之間，理論上可能破壞解析。已實測確認 plist parser 可正常處理並正確取出 LSMinimumSystemVersion 值，因此不需要退回較脆弱的文字比對。

此測試是本次改動中唯一具長期價值的部分。10.13.0 能停留數年無人察覺，根因正是沒有任何機制在宣告與實際產出脫節時發出警告；僅修正字串而不建立此測試，等同於接受同樣的漂移再次發生。

測試僅鎖定宣告值，不解析實際建置產物（理由見風險段落）。

### README 新增系統需求章節置於 Installation 之前

`README.md` 目前的章節順序為 Features、Screenshots、Installation、Admin Mode、Known Limitations。系統需求章節插入於 Installation 之前，因為使用者必須在執行安裝指令之前得知版本限制；置於其後將使章節僅對已失敗的使用者有意義。

內容需明確標示 macOS 13 Ventura 或以上，並說明此為 Go toolchain 的載入需求而非任意選擇，以免被誤讀為可繞過的建議值。

### Homebrew cask 以 depends_on 於安裝階段攔截

`chenwei791129/homebrew-apps` 的 cask formula 加入 macOS 版本相依宣告，使 brew install 在不符合的系統上直接拒絕，而非安裝完成後才在啟動時失敗。

此檔案位於另一個 repo，不在本 repo 的檔案異動範圍內，需以獨立的跨 repo 提交完成。此事實必須在 tasks 中明確標示，否則極易在本 repo 的變更合併後被遺漏，導致 cask 仍向不相容系統供應安裝。

### go.mod 升級與宣告校正分為兩個 commit

專案使用 release-please，commit 前綴決定發版行為。go directive 的提升對使用者不可見，使用 chore 前綴且不應觸發發版；最低系統需求提高則是縮減支援範圍的使用者可見變更，使用 fix 前綴以進入 CHANGELOG。

兩者仍屬同一個 change 且必須於同一批 release 出貨——若只出貨前者，產出的 binary 會帶著更大的宣告落差面對使用者。切分僅發生在 commit 層級，不切分為兩個 change，避免產生宣告不一致的中間狀態。

## Implementation Contract

**行為**：完成後，以本專案建置出的 app bundle 在 macOS 12 或更早的系統上會被 Launch Services 以明確的版本需求訊息拒絕啟動，而非通過啟動後由 dyld 以模糊錯誤攔截。macOS 13 以上的行為完全不變。透過 Homebrew 安裝的使用者在不相容系統上會於 brew install 階段收到版本不符的拒絕訊息。

**介面與資料形狀**：

- go.mod 的 go directive 值為 1.27.0。
- `Makefile` 以單一變數 `MACOS_MIN_VERSION` 持有 13.0，並將其以 `-mmacosx-version-min` 傳入 `build` 與 `build-debug` 的 wails 建置。
- `build/darwin/Info.plist` 與 `build/darwin/Info.dev.plist` 中 LSMinimumSystemVersion 對應的字串值為 13.0.0。
- `build_metadata_test.go` 位於 package main，以 howett.net/plist 將 plist 解碼為 map，取出 LSMinimumSystemVersion 鍵值並與最低版本常數比較。常數定義於測試檔內，單一來源。
- Homebrew cask 的相依宣告指向 Ventura。

**失敗模式**：

- 回歸測試在任一 Info.plist 的宣告值不等於常數時失敗，錯誤訊息需同時包含檔案路徑、實際值與期望值，使失敗原因無需閱讀測試碼即可理解。
- 檔案不存在或 plist 解析失敗時測試同樣失敗，不得以跳過或靜默通過處理——沉默正是此測試要防止的失效模式。

**驗收條件**：

- go build 全套件通過。
- go test 全套件通過，且新測試在人為將宣告值改為其他值時確實失敗（需實際驗證此反向情形，僅確認測試通過不足以證明它有作用）。
- golangci-lint 零 issue。
- make build 成功產出 app bundle，且以 otool 檢視主 binary 與 launchpal-privhelper 的 LC_BUILD_VERSION，兩者 minos 皆為 13.0，與宣告值一致。
- README 的系統需求章節位於 Installation 章節之前。

**範圍邊界**：

- 範圍內：go.mod 的 go directive、`Makefile` 的 macOS deployment target 釘定、兩個 Info.plist 的宣告值、README 系統需求章節、回歸測試、`.claude/CLAUDE.md` 的對應說明、Homebrew cask 的相依宣告（跨 repo）。
- 範圍外：CI workflow（`make build` 已涵蓋，無須另改）、任何 Go 原始碼的邏輯修改、Wails 或其他相依套件的版本升級、Info.plist 中 LSMinimumSystemVersion 以外的任何鍵、實際建置產物的自動化 minos 驗證。

## Risks / Trade-offs

- **macOS 12 使用者在升級後無法啟動應用程式** → Homebrew cask 的相依宣告於安裝階段攔截；README 系統需求章節提供事前資訊；Info.plist 校正後使直接下載 DMG 的使用者得到明確的版本訊息而非模糊失敗。三者皆為此風險的緩解措施，遺漏任一項都會讓部分使用者落回模糊失敗的路徑。

- **回歸測試僅鎖定宣告值，不驗證實際產出 binary 的 minos** → 最低版本現在存在於三個位置：兩個 Info.plist 的宣告、`Makefile` 的 `MACOS_MIN_VERSION`，以及產出 binary 的實際 minos。回歸測試只涵蓋第一項，`Makefile` 與產物之間沒有自動化的一致性檢查；若未來 Go 或 Wails 再次調整預設，測試不會察覺。此為有意識的取捨：解析建置產物需要測試依賴完整建置流程，成本與脆弱度都高於其價值。補償措施是把 otool 檢查列入本次的手動驗收條件，並在 `.claude/CLAUDE.md` 記錄「升級 Go 版本時需重新確認 minos 與宣告是否仍相等」這項關聯。

- **專案自行持有 deployment target 後與 Wails 預設脫鉤** → `MACOS_MIN_VERSION` 覆寫掉 Wails 的注入值，Wails 日後若調整自身預設（或 v3 改以 Taskfile 持有），本專案不會自動跟隨。這是刻意的：實測已確認不設定會讓 minos 隨建置機的 SDK 浮動，明確持有比自動跟隨更可預測。緩解措施同上——Go 或 Wails 升級時以 otool 重新確認。

- **跨 repo 的 cask 變更被遺漏** → cask 位於另一個 repo，本 repo 的測試與 CI 都無法涵蓋。緩解措施是在 tasks 中列為獨立且明確標示跨 repo 的項目，而非附註於其他任務之下。

- **Wails 未來版本變更 Info.plist 模板結構** → 目前檔案已納入版本控制且 Wails 僅在缺檔時生成，風險低。若確實發生，回歸測試會因解析失敗或值不符而攔截，不會靜默通過。
