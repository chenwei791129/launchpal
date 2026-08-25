## 1. Regression 測試

- [x] 1.1 在 `internal/launchctl/user_test.go` 新增「Consistent user service runtime status」的 launchctl／pgrep shim regression tests，證明相同且未變動的 launchd 狀態經 `UserManager.List` 與 `UserManager.Get` 會得到相同 Status/PID，並涵蓋 positive PID、loaded without PID、unloaded label、empty Label 與 `Program=open` 遇到無關 substring process 的案例；以 `go test ./internal/launchctl -run UserServiceStatusConsistency` 驗證測試在修復前能捕捉假 PID 與狀態不一致。

## 2. 狀態判定修復

- [x] 2.1 在 `internal/launchctl/user.go` 統一 user service 的批次與單筆分類語意：僅接受 launchd 對 label 回報的正整數 PID 作為 `StatusRunning`，已載入但無 PID 回傳 `StatusLoaded`／PID 0，未載入回傳 `StatusStopped`／PID 0，empty Label 在 List/Get 皆回傳 `StatusUnknown`／PID 0，並移除單筆狀態查詢的 `pgrep -f` fallback；以 `go test ./internal/launchctl -run UserServiceStatusConsistency` 驗證所有分類及「不呼叫 pgrep」契約通過。

## 3. 完整驗證

- [x] 3.1 對 `internal/launchctl/user.go` 與 `internal/launchctl/user_test.go` 執行 `gofmt`，再執行 `go test ./internal/launchctl` 與 `go test ./...`，確認 user service 狀態修復未改變 system daemon、Run Now 或其他 launchctl manager 的既有行為，且 working tree 僅包含本 change 規劃的 application files 與 Spectra artifacts。
