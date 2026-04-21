## 1. 型別與常數

- [x] 1.1 在 `internal/launchctl/types.go` 新增 `StatusConfidence` 欄位（string）到 `Service` struct，及常數 `ConfidenceVerified = "verified"`、`ConfidenceUnverified = "unverified"` — 實作「StatusConfidence field on Service」並落實「StatusConfidence 模型」決策
- [x] 1.2 [P] 在 `frontend/app/types/service.ts`（或對應型別檔）新增 `statusConfidence: 'verified' | 'unverified'` 欄位（實際檔案：`frontend/app/types/wails.d.ts`）

## 2. 狀態偵測核心邏輯（TDD）

- [x] 2.1 新增 `internal/launchctl/status_detect_test.go`，為「Heuristic status detection for system daemons」寫測試：single match、no match、multiple candidates、empty program、UserName 預設 root、非 root UserName、launchd parent filter 等情境
- [x] 2.2 新增 `internal/launchctl/status_detect.go`，實作 `DetectSystemServiceStatus(plist plistData) (status string, pid int, confidence string)`，落實「Heuristic detection algorithm (ppid=1 filter)」：pgrep -u + ps ppid=1 過濾（「偵測邏輯落腳位置」決策）
- [x] 2.3 在 status_detect.go 中實作「Shell skip list」：移入並共用 `commonShells` map，命中時回 StatusLoaded + verified
- [x] 2.4 跑 2.1 的測試確認通過

## 3. 整合進 readOnlyManager

- [x] 3.1 為 `readOnlyManager.getWithStatus` 新增測試，覆蓋「Status detection replaces Stopped fallback」：batch map miss 時呼叫 DetectSystemServiceStatus；batch map hit 時走原邏輯並標 verified
- [x] 3.2 修改 `internal/launchctl/readonly.go`，在 statusMap 未命中時改呼叫 `DetectSystemServiceStatus`，並將結果寫入 Service 的 Status / PID / StatusConfidence
- [x] 3.3 修改 `readonly.go` 中 batch map 命中分支，顯式設定 `StatusConfidence = ConfidenceVerified`
- [x] 3.4 確認 `SystemManager` 與 `AppleSystemManager` 的 Get / List 皆走新邏輯，符合 `launchdaemons-readonly` 的「Get system service details」MODIFIED 要求（兩者皆委派給 `readOnlyManager.getWithStatus`）
- [x] 3.5 [P] 為 `UserManager` 產出的 Service 預設 `StatusConfidence = ConfidenceVerified`（在 user.go 相關路徑設定）
- [x] 3.6 跑 3.1、以及既有 `system_test.go`、`apple_system_test.go`、`user_test.go` 確認無回歸

## 4. 前端 unverified 資訊性呈現

- [x] 4.1 [P] 在 `frontend/app/pages/system.vue` 列表 Status 欄位旁，`statusConfidence === 'unverified'` 時渲染 info icon + tooltip — 對應決策「前端 unverified 呈現」（實作位置：共用元件 `ServiceRow.vue`，system.vue 透過 `ReadOnlyServiceList` 間接使用）
- [x] 4.2 [P] 在 `frontend/app/pages/apple-system.vue` 套用相同規則（共用 `ServiceRow.vue`）
- [x] 4.3 在 `frontend/app/pages/services/[name].vue` 詳情頁 Status 區塊以相同圖示呈現
- [x] 4.4 tooltip 文案確定：「偵測到多個符合 Program 且 parent 為 launchd 的行程；顯示的 PID 可能不是該 daemon 實際對應的那個。」（不引導至 Phase 2 功能，避免 dead link）
- [x] 4.5 確認 unverified 狀態下 UI 無任何按鈕或可點擊動作（純資訊性）— svg 僅作為 tooltip trigger，無 click handler

## 5. 審核與掃尾

- [x] 5.1 執行 `make test` 確認所有測試通過
- [x] 5.2 執行 `make build-debug` 驗證 production 可編譯、Wails binding 生成正確
- [x] 5.3 手動測試：`/Library/LaunchDaemons` 下已知 running daemon 顯示 Running + 正確 PID、已知停止 daemon 顯示 Stopped、多候選情境顯示 info icon — 已於 `make dev` 執行後實機驗證 20 個 system services（9 Running + 11 Stopped），全部符合 `pgrep -u <user> -f <program>` + `ppid=1` 過濾結果；涵蓋 `Program` 欄位缺失改走 `ProgramArguments[0]` 的 case，及 shell-wrapped launcher（com.wazuh.agent）正確匹配
- [x] 5.4 Audit：`pgrep` 傳入的 program path 未經 shell interpretation（用 exec.Command argv slice 而非 sh -c 拼接）— `exec.Command("pgrep", "-u", user, "-f", program)` / `exec.Command("ps", "-o", "ppid=", "-p", strconv.Itoa(pid))` 均為 argv slice
- [x] 5.5 更新 `.claude/CLAUDE.md`：補充新的狀態偵測機制與 StatusConfidence 欄位說明
- [x] 5.6 確認 `README.md` 是否需更新（若 feature 對使用者可感知）— 已在 Features 增列 system services heuristic status detection 說明
