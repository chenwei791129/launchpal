## Context

LaunchPal 的 service 寫入路徑集中在 `internal/launchctl`：`UserManager.writePlist` 與 `SystemManager.encodePlist` 都呼叫 `BuildPlistDict(config)`（`plist_encode.go`）把 `ServiceConfig` 轉成 plist dict。`BuildPlistDict` 只輸出它認得的鍵，因此 Update 時任何未建模鍵會被丟棄。Update 同時存在於兩個 domain：`UserManager.Update`（直接 `os.WriteFile`）與 `SystemManager.Update`（編碼後交由 privhelper 以 root 寫入）。既有讀取路徑（`getWithStatus`）以 typed struct `plistData` 解析，會遺失未建模鍵，因此合併需要一條以 `map[string]any` 為目標的原始讀取路徑。`howett.net/plist` 的 `Unmarshal` 會自動偵測 XML 或 binary，可直接解進 `map[string]any`。

## Goals / Non-Goals

**Goals**

- User 與 System 兩個 Update 路徑在寫回前保留未建模鍵。
- 已建模鍵維持表單權威，含使用者清除某鍵時的移除語意。
- 讀取/解析失敗時優雅降級為全新寫入，不中止 Update。
- `modeledPlistKeys` 為單一真實來源，並以測試防止與 `BuildPlistDict` 失同步。

**Non-Goals**

- 不為未建模鍵新增 UI 編輯能力（保留 ≠ 可編輯）。
- 不修改 Create 路徑。
- 不變更 `KeepAlive` 處理（已建模鍵，由既有變更負責其 round-trip）。
- 不變更 privhelper RPC 協定，不動 frontend。
- 不在 system 無 Full Disk Access 降級時新增 UI 提示（僅程式碼註解）。

## Decisions

### Decision: MergeUnmodeledKeys removal set uses the full modeledPlistKeys set

在 `UserManager.Update` 與 `SystemManager.Update` 寫回前插入合併步驟：讀取既有 plist 為 `map[string]any`，呼叫 `BuildPlistDict` 取得 modeled dict，再以新增的 `MergeUnmodeledKeys(modeled, existing)` 合併。

`MergeUnmodeledKeys` 以「`modeledPlistKeys` 完整集合」作為移除依據，而非僅移除本次 `BuildPlistDict` 有輸出的鍵：

1. 複製 `existing`，刪除其中所有屬於 `modeledPlistKeys` 的鍵。
2. 將 `modeled` 疊加其上。

理由：若只移除「本次有輸出的鍵」，使用者把 `Run at Load` 改成 `On Demand`（清除 `RunAtLoad`）或從 `StartInterval` 切到 `StartCalendarInterval` 時，舊鍵會從磁碟被繼承回來，表單就失去權威。**Alternative considered**：在 Update 時讀回既有值補進 `ServiceConfig` 再編碼 — 否決，因為這會讓「清除」變得不可能表達，且把保留語意混入建模欄位。

### Decision: modeledPlistKeys as the single source of truth guarded by a completeness test

在 `plist_encode.go` 定義套件層級的 `modeledPlistKeys`（集合），列出 `BuildPlistDict` 能產生的每一個鍵：`Label`、`Program`、`ProgramArguments`、`RunAtLoad`、`KeepAlive`、`ThrottleInterval`、`WorkingDirectory`、`StandardOutPath`、`StandardErrorPath`、`WakeSystem`、`EnvironmentVariables`、`StartInterval`、`StartCalendarInterval`。`StartInterval` 與 `StartCalendarInterval` 皆須納入，以涵蓋兩者互斥切換。以一個完整性測試對「最大化填滿的 `ServiceConfig`」跑 `BuildPlistDict`，斷言其輸出每個鍵都屬於 `modeledPlistKeys`。**Alternative considered**：直接用 `modeled` dict 的鍵當移除集合 — 已於上一決策否決。

### Decision: readPlistMap parses the existing plist into map[string]any for both formats

新增 `readPlistMap(path string) (map[string]any, error)`：`os.ReadFile` 後 `plist.Unmarshal` 進 `map[string]any`。`howett.net/plist` 自動處理 XML 與 binary，未建模的巢狀 dict/array 以 library 原生型別 round-trip，再經同一 encoder 寫回，確保值與型別不變。輸出一律為 XML（沿用現行 Create/Update 行為，binary 來源編輯後變 XML 非本次回歸）。

### Decision: SystemManager.Update reads the plist in the GUI process and degrades on failure

`SystemManager.Update` 在 GUI 程序端呼叫 `readPlistMap(plistPath)`。LaunchPal GUI 既已為顯示而讀取 system plist，可讀到時即可合併；無 Full Disk Access 而讀不到時，`readPlistMap` 回傳 error，合併被跳過、走全新寫入。最終位元組仍由既有 `privhelper.WritePlist` 以 root 寫入並備份，helper 介面不變。**Alternative considered**：把 read-merge-write 整個下放到 helper 內原子完成 — 否決，範圍過大且偏離現行「GUI 讀取、helper 只寫位元組」的架構。

降級時 Bootout 不可省略：`oldLabel` 預設為 routing name（Create 以 `<Label>.plist` 命名，故 name 即 label），讀得到既有 plist 才以其 `Label` 覆寫。若降級時跳過 Bootout，launchd 會繼續沿用記憶體中的舊定義——`Bootstrap` 對已載入的 job 會失敗，而 `kickstart -k` 只重啟已載入的 job、不重讀磁碟上的 plist，使用者的編輯因而永不生效（直到手動 unload/reload 或重開機）。Bootout 為 best-effort，對從未載入的 daemon 回 "not bootstrapped" 無害。`UserManager.Update` 不受影響：它開頭即無條件呼叫 `Stop(name)`，本就以 routing name bootout 而不依賴讀取 plist。

### Decision: writePlistDict and encodeDict shared encoding helpers

把「dict → bytes / 寫檔」自 `writePlist`、`encodePlist` 抽出：user 端 `writePlistDict(path, pd)`、system 端 `encodeDict(pd) ([]byte, error)`，讓 Update 能傳入已合併的 dict，Create 維持傳 config。`BuildPlistDict` 的欄位對應邏輯不變。

## Implementation Contract

- **新增函式**：
  - `MergeUnmodeledKeys(modeled, existing map[string]any) map[string]any`（`plist_encode.go`）：回傳「existing 去除所有 `modeledPlistKeys` 後、疊上 modeled」的新 map；不修改入參。
  - `readPlistMap(path string) (map[string]any, error)`（`plist_encode.go`）：讀檔並解析為 `map[string]any`，XML/binary 通用。
  - `modeledPlistKeys`（集合，`plist_encode.go`）：列出所有已建模鍵，含 `StartInterval` 與 `StartCalendarInterval`。
  - 編碼 helper：user 端 `writePlistDict`、system 端 `encodeDict`。
- **行為**：
  - `UserManager.Update`：在現行 stop → write 之間，改為 `modeled := BuildPlistDict(config, true)`；若 `readPlistMap` 成功則 `modeled = MergeUnmodeledKeys(modeled, existing)`；以 `writePlistDict` 寫出 `modeled`。
  - `SystemManager.Update`：在 bootout 之後、`WritePlist` 之前，`modeled := BuildPlistDict(config, false)`；若 `readPlistMap` 成功則合併；`encodeDict(modeled)` 後交 helper。
  - Create 路徑兩端皆不變。
- **可觀察驗收**：
  - User round-trip 測試（`user_test.go`）：含 `ProcessType`/`Nice` 的 plist 經 Update 後仍保有該兩鍵且 `Program` 更新。
  - 移除語意測試：`RunAtLoad=true` + 未建模 `ExitTimeOut` 的 plist 以 On Demand config Update 後，結果不含 `RunAtLoad` 但保有 `ExitTimeOut`。
  - 降級測試：既有 plist 不存在/毀損時 Update 回傳 nil，且寫出內容等於純 `BuildPlistDict` 輸出。
  - System round-trip 測試（`system_test.go`）：以 stub `AdminClient` 捕捉 `WritePlist` 位元組，驗證未建模鍵保留。
  - `MergeUnmodeledKeys` 單元測試與 `modeledPlistKeys` 完整性測試（`plist_encode_test.go`）。
  - `make test`、`make lint` 全綠。
- **In scope**：`internal/launchctl` 的 `plist_encode.go`、`user.go`、`system.go` 及對應測試。
- **Out of scope**：`ServiceConfig`/`Service` 結構、`keepalive.go`、`internal/privhelper/**`、所有 frontend 檔案、`BuildPlistDict` 的欄位對應邏輯。

## Risks / Trade-offs

- [既有 binary system plist 在無 FDA 時讀不到 → 使用者以為保留卻被降級] → 已與使用者確認採降級；程式碼註解標明，必要時後續補 UI 提示。降級時仍以 routing name 執行 Bootout，確保編輯實際生效而非殘留舊的記憶體定義。
- [`modeledPlistKeys` 與 `BuildPlistDict` 失同步 → 新建模鍵未被移除致舊值殘留] → 完整性測試強制同步。
- [plist library 重編碼數值型別（如 uint64）造成格式漂移] → 未建模鍵以同一 library round-trip；round-trip 測試直接斷言鍵值。

## Migration Plan

純後端、無資料遷移、無相依變更。部署即生效於後續 Update。Rollback：`git revert` 對應 commit 即恢復為「全新寫入」行為。

## Open Questions

- 是否需在 system 服務因 FDA 讀不到既有 plist 而降級時，於 UI 主動提示「未建模鍵可能未保留」？目前決議不在本次處理，先以程式碼註解記錄。
