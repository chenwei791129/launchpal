## 1. 合併核心與單一真實來源（plist_encode.go）

- [x] 1.1 在 `plist_encode.go` 定義套件層級 `modeledPlistKeys` 集合，涵蓋 `BuildPlistDict` 能輸出的每一個鍵，並同時納入互斥的 `StartInterval` 與 `StartCalendarInterval`。實作需求 "Single source of truth for the modeled key set"；對應設計決策 "Decision: modeledPlistKeys as the single source of truth guarded by a completeness test"。行為：存在單一可查詢的「已建模鍵集合」供移除邏輯使用。驗證：`plist_encode_test.go` 新增完整性測試，對一組共同涵蓋所有已建模欄位的 `ServiceConfig` 跑 `BuildPlistDict`（因 `StartInterval` 與 `StartCalendarInterval` 在單一 config 內互斥，至少需含 Interval 版與 calendar 版各一），聯集所有輸出鍵後斷言每個鍵都是 `modeledPlistKeys` 的成員（`go test ./internal/launchctl -run TestModeledPlistKeys`）。
- [x] 1.2 實作 `MergeUnmodeledKeys(modeled, existing map[string]any) map[string]any`：回傳「`existing` 去除所有 `modeledPlistKeys` 後、再疊上 `modeled`」的新 map，且不修改任一入參。實作需求 "Modeled keys remain form-authoritative on Update"；對應設計決策 "Decision: MergeUnmodeledKeys removal set uses the full modeledPlistKeys set"。行為：未建模鍵被保留、已建模鍵以 `modeled` 為權威、`existing` 有而 `modeled` 無的已建模鍵被移除。驗證：`plist_encode_test.go` 單元測試涵蓋三種情形——(a) 未建模鍵 `ProcessType` 保留、(b) 已建模鍵 `Program` 由 `modeled` 覆蓋、(c) `existing` 有 `RunAtLoad` 但 `modeled` 無時被移除（`go test ./internal/launchctl -run TestMergeUnmodeledKeys`）。
- [x] 1.3 實作 `readPlistMap(path string) (map[string]any, error)`：`os.ReadFile` 後以 `plist.Unmarshal` 解析為 `map[string]any`。支撐需求 "Preserve unmodeled plist keys on service Update" 的讀取側；對應設計決策 "Decision: readPlistMap parses the existing plist into map[string]any for both formats"。行為：可從既有 plist 取回含未建模鍵的完整 map，讀取或解析失敗時回傳 error 供呼叫端降級。驗證：`plist_encode_test.go` 對一個含未建模鍵 `ProcessType` 的 XML plist 斷言回傳的 map 含該鍵；對不存在路徑斷言回傳 error（`go test ./internal/launchctl -run TestReadPlistMap`）。

## 2. 共用編碼 helper（避免 Create 與 Update 重複編碼邏輯）

對應設計決策 "Decision: writePlistDict and encodeDict shared encoding helpers"。

- [x] 2.1 [P] 自 `UserManager.writePlist` 抽出 `writePlistDict(path string, pd map[string]any) error`（負責 encode + `os.WriteFile` `0644`），`writePlist` 改為先 `BuildPlistDict` 再呼叫它。行為：既有 Create 寫入行為與輸出位元組維持不變。驗證：既有 `user_test.go` 的 Create/Write 相關測試仍全數通過（`go test ./internal/launchctl -run TestUser`）。
- [x] 2.2 [P] 自 `encodePlist` 抽出 `encodeDict(pd map[string]any) ([]byte, error)`，`encodePlist` 改為先 `BuildPlistDict` 再呼叫它。行為：system 端既有編碼輸出維持不變。驗證：既有 `system_test.go` 的 Create/encode 相關測試仍全數通過（`go test ./internal/launchctl -run TestSystem`）。

## 3. User domain Update 保留未建模鍵（user.go）

對應設計決策 "Decision: MergeUnmodeledKeys removal set uses the full modeledPlistKeys set"。

- [x] 3.1 [P] 在 `UserManager.Update` 寫回前插入合併：`modeled := BuildPlistDict(config, true)`；`readPlistMap(plistPath)` 成功則 `modeled = MergeUnmodeledKeys(modeled, existing)`，失敗則跳過合併（降級）；最後以 `writePlistDict` 寫出。實作需求 "Preserve unmodeled plist keys on service Update" 與 "Graceful degradation when the existing plist is unavailable"（user domain）。行為：編輯 user 服務時未建模鍵被保留、被清除的已建模鍵從磁碟移除、既有 plist 讀取/解析失敗時降級為全新寫入且 Update 不報錯。驗證：`user_test.go` 新增 round-trip 測試——(a) 含 `ProcessType=Background`/`Nice=5` 的 plist 經只改 `Program` 的 Update 後仍保有該兩鍵且 `Program` 已更新；(b) `RunAtLoad=true` + 未建模 `ExitTimeOut=30` 的 plist 經 On Demand config Update 後結果不含 `RunAtLoad` 但保有 `ExitTimeOut`；(c) 對毀損/不存在既有 plist 的 Update 回傳 nil 且寫出內容等於純 `BuildPlistDict` 輸出（`go test ./internal/launchctl -run TestUserManagerUpdatePreserve`）。

## 4. System domain Update 保留未建模鍵（system.go）

對應設計決策 "Decision: SystemManager.Update reads the plist in the GUI process and degrades on failure"。

- [x] 4.1 [P] 在 `SystemManager.Update` 的 bootout 之後、`WritePlist` 之前插入合併：`modeled := BuildPlistDict(config, false)`；`readPlistMap(plistPath)` 成功則合併、失敗則降級；`encodeDict(modeled)` 後交 helper 寫入；並加程式碼註解說明無 Full Disk Access 時讀不到既有 plist 而降級。實作需求 "Preserve unmodeled plist keys on service Update" 與 "Graceful degradation when the existing plist is unavailable"（system domain）。行為：編輯 system 服務時未建模鍵在送往 helper 的位元組中被保留；讀不到既有 plist 時降級為全新寫入。驗證：`system_test.go` 以 stub `AdminClient` 捕捉 `WritePlist` 收到的位元組，斷言 (a) 含 `ProcessType=Standard` 的既有 plist 經 Update 後位元組仍含該鍵、(b) 既有 plist 不可讀時寫出的位元組等於純 `BuildPlistDict` 輸出（`go test ./internal/launchctl -run TestSystemManagerUpdatePreserve`）。

## 5. 驗證與文件

- [x] 5.1 執行完整測試與靜態檢查。行為：整包後端與前端測試、lint 在含新行為下全綠。驗證：`make test` 與 `make lint` 皆 exit 0。
- [x] 5.2 依 `.claude/CLAUDE.md` 的 Post-Implementation Checklist 檢查 `README.md` 與 `.claude/CLAUDE.md` 是否需補述「Update 保留未建模 plist 鍵」的資料保真行為，並視需要更新。行為：使用者文件與專案說明與新行為一致。驗證：人工審閱 diff，確認已補述或確認確實無需變更。
