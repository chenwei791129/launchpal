## 1. 定義新的 process table 結構

- [x] 1.1 在 `internal/launchctl/status_detect.go` 定義 `processInfo struct { UID int; PPID int; Args string }` 與 `type ProcessTable map[int]processInfo`，取代現有的 `map[int]int` ppid table — 對應決策「偵測演算法改為 in-memory process table scan」

## 2. 實作 process table 讀取（TDD）

- [x] 2.1 在 `internal/launchctl/status_detect_test.go` 新增 `TestReadProcessTable_ParsesPsOutput` 等測試：覆蓋正常行、含空白的 args、截斷長 args、空白行、無法解析的行（stub `exec.Command` 或透過可注入的輸出 parser 函式）— 實作為 `TestParseProcessTable_*` 5 個測試，透過分離 `parseProcessTable` 純函式使 parser 邏輯可直接測試，毋須 stub exec
- [x] 2.2 在 `internal/launchctl/status_detect.go` 實作 `readProcessTable() (ProcessTable, error)`：呼叫 `exec.Command("ps", "-axo", "uid=,pid=,ppid=,args=")`，逐行 parse 前 3 個整數欄位 + 其餘視為 args
- [x] 2.3 新增 package-level 變數 `readProcessTableFn = readProcessTable` 作為測試注入點；移除舊的 `readAllPPIDsFn` 與移除 `pgrepCandidatesFn` 測試注入點（對應決策）
- [x] 2.4 跑 2.1 的測試確認通過

## 3. 重寫 Heuristic status detection for system daemons（TDD）

- [x] 3.1 更新 `status_detect_test.go` 既有 10 個情境為直接餵 `ProcessTable` 的形式：覆蓋「Heuristic status detection for system daemons」的 9 個新 scenarios（single match、no match、multiple candidates、non-launchd parent 過濾、UserName defaults to root、non-root UserName、empty program、snapshot shared across List、process table fetch failure degrades）
- [x] 3.2 在 `status_detect.go` 修改 `DetectSystemServiceStatus` 簽名為 `DetectSystemServiceStatus(pd plistData, table ProcessTable, uidCache map[string]int) (status string, pid int, confidence string)`；移除 `pgrepCandidatesFn`、`runPgrepCandidates`、`runParentPID`（簽名較 tasks 原文多 `uidCache`，以落實 task 3.4 caller-passes-cache 決策）
- [x] 3.3 實作新的 detection 邏輯：resolve UserName→UID、iterate table 過濾 `info.UID == targetUID && info.PPID == 1 && strings.Contains(info.Args, program)`、對候選 PID 做 `sort.Ints`、依數量回 verified/unverified
- [x] 3.4 實作 UserName→UID 解析：接受 caller 傳入的 `uidCache map[string]int`，miss 時呼叫 `os/user.Lookup`（透過 `userLookupFn` 注入點），命中則直接回；lookup 失敗時本次偵測降級為 Stopped/Unverified（spec 明確要求「not a confident Stopped」— 比 tasks 原寫的 "Stopped/Verified" 保守）— 對應決策「UserName 解析：per-List-call 的 local cache」
- [x] 3.5 跑 3.1 的測試確認通過，且等價性驗證：對同一 plist + 同一 process table，新舊實作 status/PID/confidence 完全一致

## 4. Shell skip list 維持行為等價

- [x] 4.1 確認 `Shell skip list` requirement 在新實作中仍於 `DetectSystemServiceStatus` 第 4 步命中（早於 UID 解析與 table scan），對應 scenario「Daemon invokes /bin/bash as Program」仍回 `StatusLoaded` / `verified`（已由 `TestDetectSystemServiceStatus_ShellProgramSkipped` 驗證並在測試中 fail-fast 若 user.Lookup 被呼叫）

## 5. 整合進 readOnlyManager

- [x] 5.1 修改 `internal/launchctl/readonly.go` `list()`：在批次取 statusMap 後呼叫 `readProcessTableFn()` 一次，建立 local `uidCache`，傳給所有 `getWithStatus` — 對應決策「偵測演算法改為 in-memory process table scan」的 List 共用 snapshot 規則
- [x] 5.2 修改 `getWithStatus` 簽名：ppidTable 參數改為 `table ProcessTable`，新增 `uidCache map[string]int`，同步更新 `get()` 呼叫以傳 nil（lazy fetch）
- [x] 5.3 `DetectSystemServiceStatus` 接收 nil table 時，內部呼叫 `readProcessTableFn()` 建立 per-call table；對應 Scenario「Process table snapshot shared across a single List call」在 get 路徑下退化為單次建表
- [x] 5.4 更新 `internal/launchctl/readonly_test.go` 既有 4 個測試為直接餵 `ProcessTable`（移除舊的 pgrep/ppid-table stub 方式）

## 6. 失敗降級路徑

- [x] 6.1 在 `status_detect_test.go` 新增測試 `TestDetectSystemServiceStatus_ProcessTableFetchFailureDegrades`：stub `readProcessTableFn` 回 error、且 program 非 shell → 預期 `StatusRunning` / candidate[0]（若有推測依據）或 `StatusStopped`；任一情況 confidence 為 `ConfidenceUnverified`
- [x] 6.2 在 `status_detect.go` 對 `readProcessTableFn` error 實作降級：回 Stopped/Unverified（與 UID lookup 失敗路徑一致；無 candidates 依據時 Stopped 是唯一合理選擇，confidence=Unverified 表示不確定性）

## 7. 測試與驗證

- [x] 7.1 執行 `make test` 確認所有測試通過
- [x] 7.2 執行 `make build-debug` 驗證 production 可編譯、Wails binding 生成無變化
- [x] 7.3 [P] 手動測試：`/Library/LaunchDaemons` 下已知 running daemon 顯示 Running + 正確 PID；對照 `ps -axo pid=,ppid=,args=` 輸出驗證結果等價 — 使用者實機驗證通過
- [x] 7.4 [P] 手動測試：Apple System Services（411 服務）載入時間 ≤ 500ms（用碼錶或 UI load spinner 消失時間量測），backend 部分 ≤ 200ms（加 debug log 量測）— 使用者實機驗證通過
- [x] 7.5 Audit：`ps` 傳入的 argv 皆 slice 形式（無 shell interpretation）；process table 不寫入任何 log／stderr；UserName 解析失敗的 error log 不含 args 內容（驗證：`grep -n exec.Command status_detect.go` → `exec.Command("ps", "-axo", "uid=,pid=,ppid=,args=")` 為 argv slice；實作未 log process table 或 error 詳情，降級路徑直接回 Stopped/Unverified）

## 8. 文件同步

- [x] 8.1 更新 `.claude/CLAUDE.md` 的「Status Detection Logic → System domain」章節：把「pgrep -u + ps ppid=」替換為「single `ps -axo` process table + in-memory scan」

## 9. 範圍確認（不做 cache / 背景 scan）

- [x] 9.1 實作完成後明確驗收：確認本次未引入 TTL cache、未引入背景定期 scan goroutine、未引入 startup prefetch — 對應決策「時機模型選擇：純 on-demand scan」。Grep 結果中 "cache" 僅出現於 `uidCache`（per-call UserName→uid 暫存，與 design 一致；非 service status TTL cache），無 goroutine / prefetch / timer
