## Why

目前 `SystemManager.List` / `AppleSystemManager.List` 載入 411 個 `/System/Library/LaunchDaemons` 服務需花費 2–8 秒，對使用者是明顯等待。瓶頸為 `DetectSystemServiceStatus` 對每個服務 fork 一次 `pgrep -u <user> -f <program>` subprocess — 411 × 5–20ms = 2–8s。上次 change（`system-daemon-status-detection`）已將 `ps` 用於 ppid lookup 批次化為單次呼叫，但 `pgrep` 仍是 per-service 未批次。

## What Changes

- 以單次 `ps -axo uid=,pid=,ppid=,args=` 取代 N 次 `pgrep` fork，讀回完整 process table 到記憶體
- `DetectSystemServiceStatus` 改為 in-memory scan：過濾 `uid == target_uid && strings.Contains(args, program) && ppid == 1`
- `UserName` → uid 解析透過 `os/user.Lookup`，List 呼叫期間以 local map cache 去重
- `readOnlyManager.list()` 在 List 開始時呼叫一次 `readProcessTable()`，傳給所有 `getWithStatus`
- `readOnlyManager.get()` 保留 nil table → lazy fetch 語意（單服務查詢時 on-demand 建表）
- 行為等價：所有偵測結果（status、PID、confidence）與目前實作結果一致；spec scenarios 不變

## Non-Goals

- **不加 cache / TTL**：優化後單次 List ≈ 100ms 已是即時感，cache 帶來生命週期複雜度（過期語意、Refresh 互動、fresh-read race）不划算；使用者按 Refresh 預期「現在就去查」的直觀語意保留
- **不加背景定期 scan**：app 為間歇性打開的桌面工具，視窗隱藏時持續 fork subprocess 浪費 CPU；且需處理 goroutine 生命週期、scan race、bootstrap timing 等複雜度
- **不加 startup prefetch**：首次載入 100ms 已足夠，不需攤平到啟動時
- **不動 `UserManager.getServiceStatus`**：user domain 的 pgrep fallback 是獨立 code path、不同域、不同問題，不在本次範圍
- **不動前端**：`ReadOnlyServiceList.vue` / `ServiceRow.vue` / `StatusConfidenceIcon.vue` 無需改動，Wails binding 簽名不變

## Capabilities

### New Capabilities

（無）

### Modified Capabilities

- `system-daemon-status-detection`：偵測演算法的 detection step（步驟 4–5）從「`pgrep -u <user> -f <program>` + 逐 PID `ps` ppid 查詢」改為「單次 `ps -axo uid=,pid=,ppid=,args=` 讀完整 process table，in-memory 過濾 uid 符合、args 含 program、ppid=1」。對外行為（回傳 status / PID / confidence）完全等價；所有既有 scenarios 行為不變

## Impact

- Affected specs：
  - 修改 `openspec/specs/system-daemon-status-detection/spec.md`（Requirement "Heuristic status detection for system daemons" 的 algorithm 描述）
- Affected code：
  - `internal/launchctl/status_detect.go`：`readAllPPIDs` → `readProcessTable`；移除 `pgrepCandidatesFn`、`runPgrepCandidates`、`runParentPID`；`DetectSystemServiceStatus` 簽名與實作改用 process table
  - `internal/launchctl/status_detect_test.go`：更新 test doubles（從 stub pgrep/readAllPPIDs 改為直接餵 process table）
  - `internal/launchctl/readonly.go`：`list()` 呼叫 `readProcessTable()`；`get()` 沿用 nil fallback
  - `internal/launchctl/readonly_test.go`：更新對應 test doubles
- 相依工具：需 `ps` 支援 `-axo uid=,pid=,ppid=,args=`（macOS 內建 BSD `ps` 原生支援）
