## Context

LaunchPal 目前透過 `launchctl list`（不帶 domain）一次性拉回所有服務狀態，存在 `getBatchServiceStatus()` 的 map 裡，`readOnlyManager.getWithStatus` 再用 `pd.Label` 去查。但 `launchctl list` 在 user context 下只會列出 `gui/<uid>` domain 的服務，`/Library/LaunchDaemons` 與 `/System/Library/LaunchDaemons` 所屬的 `system` domain 完全查不到，造成所有 system daemon 在 UI 被誤標為 `Stopped`、PID 為 0。

以非提權方式取得 system daemon PID 的可行手段是：讀取 plist 中的 `UserName` 與 `Program`（或 `ProgramArguments[0]`），以 `pgrep -u <user> -f <program>` 篩出候選 PID，再用 `ps` 比對 `ppid == 1`（由 launchd 起的 process 其直接 parent 永遠是 PID 1）。此組合在絕大多數情況下準確，但仍存在歧義：使用者用 `nohup` / `disown` detach 的同名 process、daemon 的多重 fork、`KeepAlive` 造成的頻繁重啟，都可能讓 `ppid=1` 的候選超過一個或對應錯誤。

專案因未取得 Apple Developer ID，無法使用 `SMAppService` 註冊 privileged helper。**本 Phase 1 決定完全不引入提權通道**：歧義情境僅以 `StatusConfidence = "unverified"` 標記資訊，使用者若需 authoritative 結果則需於 Phase 2（`session-privileged-helper`）啟用 Admin Mode，屆時 `SystemManager` 會透過 helper 取得準確狀態。這讓 Phase 1 保持最小範疇、零提權行為、易於交付與審查。

現有 `getServiceStatus` 已有「`launchctl list <label>` 無 PID → pgrep fallback」邏輯，但不過濾 `UserName` 也不檢查 `ppid`，且 `commonShells` skip 邏輯僅擋常見 shell。新邏輯要重用其 shell skip 列表，但獨立成更嚴謹的偵測函式，供 system / apple system managers 共用。

## Goals / Non-Goals

**Goals:**

- system domain daemon 在 UI 顯示正確的 `Status` 與 `PID`，不需任何前置提權
- 當啟發式偵測無法唯一定位時，明確向使用者揭示信心度（避免「看起來正確但其實錯」）
- 既有 `UserManager` 行為完全不變（user domain 的 `launchctl list` 已足夠正確）
- 偵測邏輯集中於單一檔案/函式，便於 Phase 2 未來以 helper 取代時抽換

**Non-Goals:**

- 不實作 Start / Stop / Restart / Create / Update / Delete（Phase 2 `session-privileged-helper` 的範圍）
- **不提供 osascript 或任何形式的提權校準機制**；unverified 即 unverified，不試圖在 Phase 1 升級為 verified
- 不處理 `Program` 為空且 `ProgramArguments` 也為空的異常 plist（視為無法偵測，回 `Unknown`）
- 不更動備份或 plist 讀寫機制

## Decisions

### Heuristic detection algorithm (ppid=1 filter)

採用下列嚴格順序：

1. 從 plist 取得 `UserName`（預設 `root`）與 `program`：`Program` 優先，否則 `ProgramArguments[0]`
2. 若 `program` 為空字串 → 回 `StatusUnknown`、PID 0、`ConfidenceUnverified`
3. 若 `program` 落在 `commonShells` → 回 `StatusLoaded`、PID 0、`ConfidenceVerified`（延用現行保守策略，避免誤報）
4. 執行 `pgrep -u <UserName> -f <program>` 收集候選 PID
5. 對每個候選執行 `ps -o ppid= -p <pid>`，保留 `ppid == 1` 的
6. 結果：
   - 剛好 1 個 → `StatusRunning`、該 PID、`ConfidenceVerified`
   - 0 個 → `StatusStopped`、PID 0、`ConfidenceVerified`
   - 超過 1 個 → `StatusRunning`、取第一個 PID、`ConfidenceUnverified`

**為何 ppid=1 足以篩掉多數誤判**：macOS 上 launchd (PID 1) 是所有 daemon 的直接 parent；使用者從 shell 啟動的同名 process 幾乎都有非 1 的 ppid。`nohup`/`disown` 會 reparent 到 1，這是已知 edge case，由 `ConfidenceUnverified` 標示 — 使用者若需精確結果，在 Phase 2 啟用 Admin Mode 後能取得 authoritative 資訊。

**為何不用 `-u root` 寫死**：部分 LaunchDaemon 會在 plist 中指定 `UserName` 為 `_www`、`_mdnsresponder` 等非 root system user；寫死 `root` 會把這些服務誤判為 `Stopped`。

**Alternatives considered:**
- `launchctl procinfo <pid>` (non-root)：對 system domain process 會回 permission denied，派不上用場
- 讀 `/proc` 類似介面：macOS 沒有 `/proc`，只有 `sysctl KERN_PROC` — 需 cgo，複雜度不值
- **單次 osascript 提權（舊草案）**：被否決 — 增加維護複雜度、TCC 對話框副作用、且與 Phase 2 的 Admin Mode 定位重疊；Phase 1 的目標是「零提權的資訊可視化」，不該把提權邏輯埋進唯讀路徑

### StatusConfidence 模型

在 `Service` struct 新增 `StatusConfidence string` 欄位，常數：

- `ConfidenceVerified` = `"verified"` — 偵測結果可信（唯一匹配或無匹配）
- `ConfidenceUnverified` = `"unverified"` — 多重候選或 `program` 無法判定

`UserManager` 產出的 Service 一律為 `ConfidenceVerified`（`launchctl list` 在 user domain 是 authoritative）。`SystemManager` / `AppleSystemManager` 依偵測結果填入。

Phase 1 內 `unverified` 為終態，**沒有升級路徑**。Phase 2 實作完成後，`SystemManager` 在 Admin Mode 開啟時會從 helper 取得 authoritative 資料並標為 `verified`；但這是 Phase 2 的職責，Phase 1 spec / code 不預留任何 Verify 通道。

**為何是字串欄位而非 bool**：保留未來擴充空間（例如 Phase 2 可能新增 `"helper-verified"` 來源區分），且對 JSON 序列化與前端 UI 分支更自然。

**Alternatives considered:**
- 只用布林 `Verified`：擴充性差
- 用 enum int：跨 Go/TypeScript 邊界比字串容易對不齊

### 偵測邏輯落腳位置

新增 `internal/launchctl/status_detect.go`，公開 `DetectSystemServiceStatus(plist plistData) (status string, pid int, confidence string)`。`readOnlyManager.getWithStatus` 於 `statusMap` 未命中時呼叫此函式。`commonShells` 常數從 `user.go` 移到 `status_detect.go`（user.go 透過 import 重用）。

**為何不直接擴充 `getServiceStatus`**：`getServiceStatus` 以 `label` 為輸入、以 `launchctl list <label>` 為主邏輯，適合 user domain；system domain 完全不需 `launchctl list`，邏輯形狀不同，共用會增加分支複雜度。

### 前端 unverified 呈現

- `pages/system.vue` / `pages/apple-system.vue` 列表中，Status 欄位旁，`statusConfidence === 'unverified'` 時額外渲染 info icon
- hover 時顯示 tooltip：「偵測到多個符合 Program 且 parent 為 launchd 的行程；顯示的 PID 可能不是該 daemon 實際對應的那個。啟用 Admin Mode（Phase 2）可取得 authoritative 結果。」
- `pages/services/[name].vue` 詳情頁 Status 區塊類似呈現
- **沒有按鈕、沒有 Wails method 呼叫** — 純顯示性元素

**為何不給按鈕引導去 Phase 2**：Phase 2 還未實作，UI 指向未存在的功能會造成 dead link；Phase 2 實作時會新增「Enable Admin Mode」入口，屆時 tooltip 可再更新。Phase 1 保持資訊性。

## Risks / Trade-offs

- **[nohup/disown edge case 造成誤判]** → 由 `ConfidenceUnverified` 顯式標記；使用者若需精確結果需等 Phase 2 的 Admin Mode
- **[`pgrep` 的 `-f` 以整條 command line 比對，可能匹配到 grep 自己或其他含 program 字串的 process]** → `-f` 的 pattern 使用完整路徑（`/usr/bin/foo` 而非 `foo`）大幅降低碰撞；仍有碰撞時 `ppid=1` 過濾能再縮一層
- **[某些 daemon 使用 wrapper shell script，真正工作 process 的 ppid 是該 wrapper 而非 1]** → 此類服務會被判為 Stopped，Phase 1 已知侷限；Phase 2 可解
- **[`StatusConfidence` 新欄位對 JSON 序列化相容性]** → 前端舊版本會忽略未知欄位；後端送出新欄位不會影響既有消費者；相容性 OK
- **[unverified 使用者觀感]** → 可能被誤解為「bug」；靠 tooltip 文案明確說明「為何不確定、有甚麼方法可以變確定（Phase 2 的 Admin Mode）」
