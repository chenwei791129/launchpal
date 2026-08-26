## 1. Go toolchain 升級

- [x] 1.1 將 `go.mod` 的 go directive 由 1.26.0 提升至 1.27.0，使專案以 Go 1.27 建置且無任何原始碼變更。此步驟落實決策「採用 macOS 13 為新基準，不以 linker 選項覆寫版本標記」——不得加入任何 linker 的 macOS 版本覆寫參數。驗證：go build ./... 與 go test ./... 全套件通過。
- [x] 1.2 確認升級後靜態檢查與建置工具鏈維持可用，Wails CLI 與 golangci-lint 皆可由 Go 1.27 編譯執行。驗證：go tool golangci-lint run ./... 回報零 issue，且 go tool wails version 正常輸出 v2.15.0。

## 2. 最低版本宣告與回歸測試

- [x] 2.1 依決策「以 plist 解析的回歸測試鎖定最低版本宣告」，新增 `build_metadata_test.go`（package main），實作 requirement「A regression test guards the declared version」：以 howett.net/plist 將兩個 Info.plist 解碼後取出 LSMinimumSystemVersion，與測試內單一定義的期望版本常數比對；檔案讀取失敗、解析失敗或鍵不存在皆須判定為測試失敗，不得跳過或靜默通過。失敗訊息須同時包含檔案路徑、實際值與期望值。驗證：此時兩個 plist 仍為 10.13.0，執行 go test 時該測試必須失敗，且失敗訊息可辨識出是哪個檔案、實際值為何。
- [x] 2.2 依決策「Info.plist 宣告值校正為 13.0.0 並與 Go toolchain 對齊」，實作 requirement「App bundle declares the minimum system version」：將 `build/darwin/Info.plist` 與 `build/darwin/Info.dev.plist` 的 LSMinimumSystemVersion 由 10.13.0 改為 13.0.0，使 macOS 12 以下的使用者由 Launch Services 以明確版本訊息攔截，而非通過啟動後由 dyld 以模糊錯誤攔截。驗證：2.1 的測試由失敗轉為通過。
- [x] 2.3 反向確認回歸測試確實具備攔截能力：暫時將任一 Info.plist 的宣告值改為其他數值，確認測試失敗後還原。僅確認測試通過不足以證明它有作用。驗證：改動期間 go test 失敗，還原後通過，且工作區無殘留變更。

## 3. 對外文件與專案說明

- [x] 3.1 [P] 依決策「README 新增系統需求章節置於 Installation 之前」，實作 requirement「README documents the system requirement」：在 `README.md` 的 Installation 章節之前新增系統需求章節，標示需要 macOS 13 Ventura 或以上，並說明此為 Go toolchain 產出 binary 的載入需求而非可繞過的建議值，使使用者在執行安裝指令前即得知限制。驗證：檢視 README 章節順序，系統需求出現在 Installation 之前，且內容明確標示版本與成因。
- [x] 3.2 [P] 在 `.claude/CLAUDE.md` 記錄最低 macOS 版本的維護關聯：說明宣告值必須等於 Go toolchain 寫入 binary 的 minos，且升級 Go 版本時須重新確認兩者是否仍相等。此為回歸測試不涵蓋實際建置產物的補償措施。驗證：內容審閱，該段落明確指出 Go 版本升級與最低版本宣告之間的相依關係。

## 4. 建置產物驗證

- [x] 4.1 依決策「於 Makefile 明確釘死 Wails 建置的 macOS deployment target」，實作 requirement「The build pins the macOS deployment target」：於 `Makefile` 新增單一來源變數持有 13.0，並以 `-mmacosx-version-min` 透過 `CGO_CFLAGS` / `CGO_LDFLAGS` 傳入 `build` 與 `build-debug` 的 wails 建置。此為必要步驟而非最佳化——Go 1.27 的 13.0 預設只在內部連結時生效，主 binary 因 Wails 的 darwin ObjC 後端走外部連結，不釘死則停在 Wails 硬編碼的 11.0，使宣告值高於實際 minos。註記 Wails 的 UpsertEnv 在偵測到既有 flag 時不覆寫，`-framework UniformTypeIdentifiers` 仍會被附加。驗證：`make build` 無 `ld: warning: ... than being linked` 警告，且 `otool -L` 確認 UniformTypeIdentifiers 仍正常連結。
- [x] 4.2 實作 requirement「Declared minimum macOS version matches the toolchain's binary requirement」的實際驗證：執行 make build 產出 app bundle，以 otool 檢視主 binary 與 launchpal-privhelper 的 LC_BUILD_VERSION，確認兩者 minos 皆為 13.0，與 Info.plist 宣告的 13.0.0 一致。驗證：otool 輸出的 minos 為 13.0，且與宣告值相等。若不相等，先判斷該 binary 走內部或外部連結再決定處置：外部連結者調整 4.1 的釘定值，內部連結者以 Go 預設為準並回頭校正宣告值；不得為了保住較低的支援範圍而把標記往下壓（決策一）。

## 5. 跨 repo：Homebrew cask

- [x] 5.1 [P] 依決策「Homebrew cask 以 depends_on 於安裝階段攔截」，於 chenwei791129/homebrew-apps 的 cask formula 實作 requirement「Cask declares the minimum macOS version」：加入要求 Ventura 或以上的 macOS 相依宣告，使不相容系統於安裝階段即被拒絕而非安裝完成後才在啟動時失敗。此變更位於另一個 repo，本 repo 的測試與 CI 皆無法涵蓋，須以獨立提交完成，不得因本 repo 變更合併而視為已完成。驗證：cask formula 內含 macOS 版本相依宣告且版本與 app bundle 宣告一致；於支援版本上執行安裝流程仍可正常完成，包含既有的 postflight quarantine 移除行為。

## 6. 提交與發布

- [x] 6.1 依決策「go.mod 升級與宣告校正分為兩個 commit」切分提交：go directive 升級使用 chore 前綴（對使用者不可見，不觸發發版），最低系統需求校正（含 Makefile 的 deployment target 釘定）、README、回歸測試與專案說明使用 fix 前綴（縮減支援範圍的使用者可見變更，須進入 CHANGELOG）。兩者屬同一批 release，不得只出貨前者。驗證：git log 顯示兩個獨立 commit 且前綴符合上述分類，暫存檔案依關注點分組而非一次全數加入。
