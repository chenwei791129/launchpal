## Context

`system-daemon-status-detection` 上線後，System / Apple System 分頁的狀態顯示正確，但 Apple System Services 頁面載入 411 個服務需 2–8 秒，明顯影響使用者體驗。

瓶頸定位：`DetectSystemServiceStatus` 對每個服務執行一次 `pgrep -u <UserName> -f <program>`。macOS 上 fork+exec 一個短命 subprocess 約 5–20ms，411 × 10ms ≈ 4s。上次 change 已把 ppid 查詢批次化為單次 `ps -axo pid=,ppid=`，但 pgrep 仍是 per-service 未處理。

對 macOS 而言，「透過 uid + 完整 command line 比對尋找行程」並不需要 subprocess — 在記憶體裡對 process table 做 `strings.Contains` 即可。`pgrep` 只是對 `ps` 的便利包裝；直接從 `ps` 拿完整表格、自己做過濾，省下 N-1 次 fork。

## Goals / Non-Goals

**Goals:**

- Apple System Services 列表載入時間從 2–8s 降至 **≤ 500ms（含 UI render）**，backend 部分 ≤ 200ms
- 行為等價：對同樣的 plist 與同樣的系統狀態，新實作的 status / PID / confidence 結果與目前實作完全一致
- 保留 `DetectSystemServiceStatus` 的 testability：單元測試能直接餵 process table stub，不需 exec 真實 subprocess
- 保留 `get()` 的單服務 on-demand 語意：查單一服務時 lazy fetch process table

**Non-Goals:**

- 不加記憶體 cache / TTL — 詳見決策「時機模型選擇」
- 不加背景定期 scan — 詳見同一決策
- 不加 startup prefetch — 同上
- 不動 `UserManager.getServiceStatus` 的 pgrep fallback — 不同 domain、user context 下 `launchctl list` 已是主要來源
- 不動 Wails binding 簽名、不動前端 — 行為等價純後端優化
- 不處理 `ps` 截斷 argv 的 edge case（BSD `ps` 對超長 command line 會截斷；若某 daemon 的 argv 被截到剛好把 program path 切開則可能漏配對，視為已知侷限）

## Decisions

### 時機模型選擇：純 on-demand scan

每次前端 `List` call 觸發一次 fresh scan，不做 cache、不做背景 job、不做 startup prefetch。

**為何此設計**：

優化後 backend 處理單次 List ≈ 100ms，對使用者體感就是「即時」。在此延遲下，caching 帶來的邊際效益不足以抵銷其成本。

**Alternatives considered：**

- **TTL cache（10s）**：首次 100ms、切回 < 1ms
  - 否決：需處理過期語意、Refresh 是否清 cache、cache miss race、快取時資料與實際不一致的 debug 成本。優化後切回 100ms 已是體感即時，cache 的邊際效益不足以抵銷複雜度
- **背景定期 scan（每 5s）**：總是 < 1ms
  - 否決：LaunchPal 是間歇性桌面工具，視窗關閉時持續 fork 浪費 CPU；需處理 goroutine 生命週期、scan race、bootstrap timing；增加「使用者看到的不是當下」的認知負擔（用 tooltip 告知「5 秒前」又是另一層 UX）。沒有強烈的即時可視化需求支撐這個複雜度
- **Startup prefetch**（啟動時背景暖一次 cache）
  - 否決：和 TTL cache 同樣的複雜度 + 要處理「prefetch 還沒完成前使用者就點」的 race。首次 100ms 不需要攤到啟動時

### 偵測演算法改為 in-memory process table scan

`readAllPPIDs()` 擴充為 `readProcessTable()`，透過 `ps -axo uid=,pid=,ppid=,args=` 一次取得：

```go
type processInfo struct {
    UID  int
    PPID int
    Args string // full command line
}
type ProcessTable map[int]processInfo // key: pid
```

`DetectSystemServiceStatus` 改為掃表（不再 fork pgrep）：

```
for pid, info := range table {
    if info.UID != targetUID { continue }
    if info.PPID != 1 { continue }
    if !strings.Contains(info.Args, program) { continue }
    candidates = append(candidates, pid)
}
```

**等價性證明**：
- `pgrep -u <user>` 等於 `ps` 輸出過濾 `uid == resolve(user)`
- `pgrep -f <program>` 等於 `ps args=` 輸出過濾 `strings.Contains(args, program)`
- 原本的 ppid 過濾（ppid == 1）邏輯完全不變
- 候選排序：原本 `pgrep` 輸出按 PID 遞增；新實作走 `range map` 順序不保證，故需對候選 PID 做 sort 以保持「多候選取首個」的確定性

**為何此設計**：
- 消滅 N-1 次 subprocess fork（從 411 次 → 1 次）
- 邏輯與 `pgrep -u -f` 完全等價，行為可驗證
- 記憶體成本極小（macOS 上 process 數量通常 < 1000，每個約 200 bytes → 總量 < 200KB）

**Alternatives considered：**

- **pgrep 平行化（16-goroutine pool）**：仍 411 次 fork，只是並行
  - 否決：fork 本身就是資源浪費；對系統不友善；不如直接根除
- **用 cgo 呼叫 `sysctl(KERN_PROC_ALL)`**：真正原生 API、無 subprocess
  - 否決：引入 cgo 編譯複雜度、跨 macOS 版本 struct layout 風險、對一個桌面工具過度工程。`ps` 輸出格式在 macOS 跨版本穩定已久

### UserName 解析：per-List-call 的 local cache

plist 的 `UserName` 欄位是字串（如 `"_www"`、`"_mdnsresponder"`），process table 的 uid 是整數。需要 `os/user.Lookup(name)` 做一次轉換。

411 個服務中若重複 user（大多是 root），重複呼叫 `user.Lookup` 是浪費。用 local `map[string]int` cache，生命週期與單次 List call 相同（call 結束就釋放）。

**為何 per-call 而非全域 cache**：使用者帳號／系統 user 可能在 app 存活期間變動（雖然罕見）；per-call cache 不需處理失效，拿到錯誤 uid 的風險為零。411 次 lookup 首 call 成本一次性攤提，總共只是多幾 ms。

**Alternatives considered：**

- **全域 sync.Map cache**：跨 List call 共享
  - 否決：多了一層 staleness 顧慮；對問題規模無顯著收益
- **不 cache，每次都 Lookup**：簡單
  - 否決：411 次 `user.Lookup` 約 5–10ms（走 Open Directory / dsmemberd），仍是可觀成本

### 移除 `pgrepCandidatesFn` 測試注入點

`pgrepCandidatesFn` 與 `readAllPPIDsFn` 是當前測試用的 package-level function 變數。改動後：

- `readAllPPIDsFn` → `readProcessTableFn`（保留 stub 點）
- `pgrepCandidatesFn` **移除**（不再有 pgrep 呼叫可 stub）

測試從「stub pgrep 回傳 candidates + stub ppid table」簡化為「直接餵完整 process table」。測試碼更直白，且等價情境（single match、no match、multi-candidate、non-launchd parent 過濾等）覆蓋率不變。

## Risks / Trade-offs

- **[`ps -axo args=` 對超長 command line 會截斷]**
  → 目前 macOS `ps` 預設截斷至約 2048 chars。若某 daemon 的 `Program` 路徑（通常 < 100 chars）被截斷到剛好切開、導致 `strings.Contains` 漏配對 → daemon 被誤判為 Stopped。視為已知侷限；實際 system daemon 的 program path 都很短，觸發機率極低
- **[`ps` 跨 macOS 版本輸出格式變化]**
  → macOS 的 BSD `ps` 格式穩定數十年未變；`-axo uid=,pid=,ppid=,args=` 的欄位順序與空白分隔是 POSIX 定義。風險可忽略
- **[在 `range map[int]processInfo` 的不確定順序下，多候選情境取「首個」變成非確定]**
  → 實作時對候選 PID slice 做 `sort.Ints` 後取 `[0]`，與舊行為（pgrep 輸出按 PID 遞增取首個）等價
- **[root user 的 uid 解析若失敗會讓所有 system daemon 都變 Stopped]**
  → `user.Lookup("root")` 幾乎不會失敗；若失敗應 log error 並讓該次 List 降級：所有 daemon 的 StatusConfidence 變 `unverified`，status 依 `candidates[0]` 或 `Stopped` 決定。具體降級策略在 tasks 中定義
- **[Args 內容可能含敏感資訊]**
  → `ps args=` 對 system daemon 不含 user 機密；但全 process table 讀入記憶體時要避免日誌化。實作確保 process table 不寫 log

## Migration Plan

- 無資料遷移（記憶體狀態變更，無 persistent storage）
- 前端無需更新（Wails binding 簽名不變）
- Release 即生效；release-please 走 `refactor` 或 `perf` 不觸發版本 minor bump（走 patch）

## Open Questions

（無）
