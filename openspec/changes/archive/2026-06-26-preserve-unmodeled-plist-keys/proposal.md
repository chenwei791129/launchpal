## Why

當使用者編輯（Update）一個既有 service 時，目前的寫入路徑以 `BuildPlistDict` 從 `ServiceConfig` 從零重建 plist，導致任何 LaunchPal 未建模的 launchd 鍵（`ProcessType`、`Nice`、`UserName`/`GroupName`、`SoftResourceLimits`/`HardResourceLimits`、`MachServices`、`Sockets`、`LimitLoadToSessionType`、`ExitTimeOut`、`AbandonProcessGroup` 等）被靜默丟棄。這是一個影響進階使用者的資料保真（data-fidelity）缺陷：使用者只想改一個欄位，卻會在不知情下喪失原本 plist 攜帶的進階設定。對應 GitHub Issue #35。

## What Changes

- 在 **User**（`~/Library/LaunchAgents`）與 **System**（`/Library/LaunchDaemons`）兩個 Update 寫入路徑上，寫回前先讀取磁碟上既有的 plist，把 LaunchPal 未建模的鍵原封不動地保留（read-merge-write）。
- 已建模鍵維持「表單權威」：表單驅動的鍵以 `BuildPlistDict` 的輸出為準，覆蓋既有值；當使用者清除某個已建模鍵（例如把 Launch Policy 由 `Run at Load` 改為 `On Demand`），結果 plist 必須**移除**該鍵，而非從磁碟把舊值繼承回來。
- 合併以「已建模鍵的完整集合」做移除（單一真實來源 `modeledPlistKeys`），而非僅移除本次 `BuildPlistDict` 有輸出的鍵，以避免 interval ↔ calendar 切換等情境殘留舊鍵。
- 當既有 plist 讀不到或解析失敗（例如 system 服務無 Full Disk Access）時，優雅降級為現行的「全新寫入」行為，Update 不因此整個失敗。
- 純後端 Go 行為修正；不變更 `ServiceConfig`、不變更 privhelper RPC 協定、無 frontend 變更。

## Non-Goals (optional)

- 不為任何目前未建模的鍵新增 UI 編輯能力（保留 ≠ 可編輯）。
- 不修改 Create 路徑（沒有既有 plist 可合併）。
- 不變更 `KeepAlive` 的處理；其完整 round-trip 由既有的 advanced-keepalive-options 變更負責，`KeepAlive` 屬已建模鍵。
- 不在 system 服務因 Full Disk Access 讀不到既有 plist 而降級時，於 UI 主動提示使用者（先以程式碼註解記錄，留待後續）。
- 不改變既有「binary plist 編輯後輸出為 XML」的行為。

## Capabilities

### New Capabilities

- `preserve-unmodeled-plist-keys`: 在 service Update 時，跨 user 與 system domain 讀取既有 plist、保留 LaunchPal 未建模的鍵、讓已建模鍵維持表單權威，並在讀取/解析失敗時優雅降級為全新寫入的橫切資料保真行為。

### Modified Capabilities

(none)

## Impact

- Affected specs: 新增 capability `preserve-unmodeled-plist-keys`（細化既有 `core-service-management` 的 user 服務 Update 與 `system-daemon-write-ops` 的 system 服務寫入行為，但不結構性修改該兩份 spec）。
- Affected code:
  - New:
    - (none — 以下皆為既有檔案內新增函式)
  - Modified:
    - internal/launchctl/plist_encode.go
    - internal/launchctl/user.go
    - internal/launchctl/system.go
    - internal/launchctl/plist_encode_test.go
    - internal/launchctl/user_test.go
    - internal/launchctl/system_test.go
  - Removed:
    - (none)
