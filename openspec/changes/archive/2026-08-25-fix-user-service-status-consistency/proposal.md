## Problem

同一個 user LaunchAgent 在 service list 與 service detail 可能顯示不同的執行狀態與 PID。已載入但沒有 launchd PID 的 job，會在詳情查詢中因程序名稱的寬鬆比對而被誤判為 `running`，甚至顯示無關程序的 PID；列表則正確顯示為 `loaded`。

## Root Cause

`UserManager.List()` 使用批次 `launchctl list` 結果，但 `UserManager.Get()` 在單筆 launchctl 查詢沒有 PID 時，會執行 `pgrep -f <program>`。當 program 是 `open` 等短字串時，程序搜尋可能命中命令列中包含該字串的無關程序，導致兩條查詢路徑採用不同標準並產生假 PID。

## Proposed Solution

統一 user service 的列表與詳情狀態判定，僅使用 launchd 對該 label 回報的載入狀態與 PID：

- launchd 回報正整數 PID 時為 `running` 並回傳該 PID。
- job 已載入但 launchd 未回報 PID 時為 `loaded` 且 PID 為 0。
- label 未載入時為 `stopped` 且 PID 為 0。
- plist 的 `Label` 為空時，列表與詳情皆回傳 `unknown` 且 PID 為 0，不向 launchd 查詢空 label。
- 移除 user service 單筆狀態查詢的程序名稱 fallback，不再從無法歸屬於該 launchd job 的程序推導 PID。

## Success Criteria

- 在 service 狀態未改變的前提下，`ListServices` 與 `GetService` 對同一個 user service 回傳相同的 Status 與 PID。
- program 為 `open`、job 已載入且 launchd 沒有 PID 時，列表與詳情皆回傳 `loaded`、PID 0。
- plist 的 `Label` 為空時，列表與詳情皆回傳 `unknown`、PID 0。
- 單筆查詢不會因無關程序的命令列包含 program 字串而回傳 `running` 或假 PID。
- frontend 現有的狀態配色、Stop/Start 顯示及 Run Now 確認條件不需修改，並持續依 backend 回傳狀態運作。
- backend regression tests 驗證批次與單筆狀態分類使用相同 launchd 語意。

## Non-Goals

- 不追蹤由 wrapper command 啟動的下游 application process；例如 `open` 結束後，Chrome 是否仍執行不代表 LaunchAgent job 為 `running`。
- 不修改 system 或 apple-system daemon 的 heuristic status detection。
- 不調整 frontend 的狀態顏色、操作按鈕、Run Now 確認流程或自動刷新策略。
- 不新增 Wails binding、Service JSON 欄位或新的 runtime status 值。

## Capabilities

### New Capabilities

（無）

### Modified Capabilities

- `core-service-management`：明確規範 user service 的 List 與 Get 使用一致的 launchd 狀態分類，且不得以無關程序名稱推導 PID。

## Impact

- Affected specs: `core-service-management`
- Affected code:
  - Modified:
    - `internal/launchctl/user.go`
    - `internal/launchctl/user_test.go`
  - New: （無）
  - Removed: （無）
