## Context

LaunchPal 目前提供查看 LaunchAgent / LaunchDaemon 的 stdout / stderr log 內容（`UserManager.GetLogs`、`readOnlyManager.getLogs`），檔案讀取以 1 MB 為上限的 tail 方式呈現。issue #18 要求補上對應的清除動作：使用者一鍵把當前 log 檔案內容截斷為 0 byte，避免 daemon 長期執行造成磁碟膨脹。

跨層影響：
- 前端：`ServiceLogs.vue` 的控制列、確認 modal 風格沿用 `[name].vue` 既有 Run Now 樣式。
- Wails layer：`app.go` 已分軌成 user / system 兩組 binding（`GetLogs` vs `GetSystemLogs`），新動作沿用相同分軌。
- launchctl 內部介面：`Manager` interface 在 `manager.go` 已定義 `GetLogs`，新增方法須一致延伸。
- privileged helper：`internal/privhelper/handlers.go` 已有 `EnsureLogAccess` 對 log 路徑做 allowlist 驗證（`allowedLogPathPrefixes`），新 RPC 重用同一條 `validateLogPath`。
- 安全要求：helper 跑在 root，必須維持 O_NOFOLLOW、路徑 allowlist、TOCTOU 防護等既有不變式。

## Goals / Non-Goals

**Goals:**

- 使用者於 Logs tab 一鍵清除當前選中的 stdout 或 stderr log。
- 對 system 服務做 per-file 寫入權限分流：使用者本身可寫的 log 檔不必啟用 Admin Mode 也可清除。
- 按鈕在所有狀態下都有明確 UX：可用 / disabled + tooltip 解釋原因 / 對 apple-system 完全隱藏。
- 不破壞 helper 既有安全護欄，新 RPC 沿用 path allowlist 與 O_NOFOLLOW 模式。

**Non-Goals:**

- 不引入「同時清 stdout + stderr」的整合動作。
- 不對 user 服務做 per-file 權限檢查（user log 預設使用者可寫）。
- 不為 apple-system 服務開放任何寫入路徑，包含 log 截斷。
- 不對 List 階段做預取；權限狀態只在 detail 頁打開或 logType 切換時 lazily 查詢。
- 不做 log archive、rotation、刪除整個檔案；只做 in-place truncate 至 0 byte。

## Decisions

### 截斷 log 檔案而非僅清空畫面顯示

issue 描述「helps manage disk space」明確指向實體檔案層面的清除。實作上以 `os.OpenFile(path, O_WRONLY|O_TRUNC|O_NOFOLLOW, 0)` 開啟檔案並立即關閉，不要 `os.Remove` + 重建：保留檔案 inode 與權限位元，避免 daemon 已開啟的 fd 寫到孤兒檔案、避免重建後 log 路徑因父目錄 mode 較嚴而再次無法寫入。

替代方案：
- 寫空字串覆蓋——等價但多一次 `Write` syscall。
- `truncate(2)` syscall 直譯——Go 的 `os.Truncate` 會 follow symlink，沒有 O_NOFOLLOW 變體；用 `OpenFile` + `O_TRUNC` 才能在 helper 內阻擋 symlink attack。

### 新增三支 Wails binding 而非擴增現有 GetLogs / GetSystemLogs

`app.go` 既有的 `GetLogs(name, logType)` 與 `GetSystemLogs(name, serviceType, logType)` 已分軌 user vs system；對稱新增 `ClearLogs`、`ClearSystemLogs` 與 `GetLogClearStatus` 三支 binding 維持同一風格。

替代方案：
- 把 `ClearLogs` 收進單一 binding 接收 path——失去 Manager 內部對 service 名稱的解析與權限保護。
- 把權限狀態併進 `GetLogs` 回傳——將純粹讀取的 API 與授權狀態耦合，且每次 reload 都重算，不必要。

### `LogClearStatus` 結構而非單純 bool

回傳 `{ logPath, exists, userWritable }` 三個欄位，前端可組合出三段不同 tooltip 文案：「No log path configured」、「Log file does not exist」、「Enable Admin Mode to clear」。bool 不夠，會迫使前端再 stat 一次。

### system 服務 per-file 寫入權限分流

`SystemManager.ClearLogs` 內部以下列順序判斷：
1. 取出 plist 中對應 logType 的路徑；空路徑回 ErrNoLogPath。
2. `os.OpenFile(path, O_WRONLY|O_NOFOLLOW, 0)` 試開——成功即視為使用者可直接 truncate，沿用同一 fd 加 `O_TRUNC` 即可（或關閉後再 `OpenFile(... O_TRUNC ...)`）。
3. 試開失敗且 errno 為 EACCES：若 Admin Mode 啟用，呼叫 helper 的 `TruncateLog` RPC；否則回 `ErrReadOnlyManager`。
4. 試開失敗其他原因（ENOENT、ELOOP）：對應錯誤直接回。

這個流程與既有 `Start/Stop/Restart` 的 `m.client()` 為 nil 時 fallback 模式同構。

### 新增 helper RPC `TruncateLog` 而非重用 WritePlist

`WritePlist` 限制路徑必須是 `/Library/LaunchDaemons/*.plist`，不適合用來寫 log 檔。新 RPC 限定 log 路徑 allowlist（重用 `validateLogPath`），實際操作就是 `os.OpenFile(path, O_WRONLY|O_TRUNC|O_NOFOLLOW, 0)`。重用 `validateLogPath` 還可繼承既有「prefix 不能是 allowlist root 本身」的防呆——避免不小心截斷 `/var/log/system.log` 這種 log 根目錄裡的核心檔。

不過為了這個動作，我們會 **限縮** allowlist 條件：`TruncateLog` 只接受**已存在**的檔案，缺檔回 `ErrCodeNotFound`，避免 helper 被誘導在 allowlist 路徑下憑空建立 root-owned 檔。`EnsureLogAccess` 的「不存在則建立」語義不會擴散到 truncate。

### 後端在實際 truncate 時必再驗權限

`GetLogClearStatus` 是給 UI 用的，前端據此決定按鈕 enable/disable。但使用者按確認的瞬間，檔案 mode、owner、Admin Mode 狀態都可能已變動（TOCTOU 風險）。後端的 `ClearLogs` / `ClearSystemLogs` 必須以實際 OpenFile 的成敗為準，而不是先 `stat` 再 `open`：先 stat 後 open 之間如果有 race，會把 stat 的結論誤套到變動後的檔案。

這也代表 `LogClearStatus.userWritable` 的計算應該用 `os.OpenFile(path, O_WRONLY|O_NOFOLLOW, 0)` 試開後立刻關閉的 access test，而非 `stat` + `os.Geteuid()` 比對 mode bits——前者是真實的 kernel 判定，後者要自己處理 ACL、setgid 群組等邊界。

### 按鈕可見性 / 可用性矩陣

| serviceType | UserWritable | Admin Mode | 按鈕 |
|-------------|--------------|------------|------|
| user | — | — | enabled |
| system | true | — | enabled（直接路徑） |
| system | false | enabled | enabled（helper 路徑） |
| system | false | disabled | **disabled** + tooltip「Enable Admin Mode to clear」 |
| apple-system | — | — | **隱藏** |

`ServiceLogs.vue` 不再只看單一 `canWrite` prop，而是同時讀取 `serviceType`、`admin.isEnabled` 與 `LogClearStatus`，計算出 `(visible, enabled, tooltipReason)`。為了讓元件易測，這些值由 `[name].vue` 傳入而非元件內 import composable。

### 確認 modal 沿用 Run Now 風格

直接複用 `[name].vue:320-348` 的 Teleport pattern，紅色強調按鈕（破壞性操作 vs Run Now 用橘色）。文案：標題 `Clear Logs`、內文 `This will permanently truncate the {stdout|stderr} log file for {service-name}. The file is reset to 0 bytes; existing entries cannot be recovered. Continue?`、按鈕 `Cancel` / `Clear Logs`。

## Risks / Trade-offs

- **Risk**: `os.OpenFile(path, O_WRONLY|O_TRUNC, ...)` 立刻把檔案截斷，daemon 若使用 `O_APPEND` 會自動把新寫入接到 0 offset；但若 daemon 用 `lseek` + `write` 的非 append 模式並快取了 offset，新 log 會寫到 sparse hole。 → Mitigation: launchd 把 daemon 的 stdout/stderr 重新導向到檔案時，預設都是 `O_APPEND`，這個情況實務上罕見；但仍在 design 與 spec 中註記，並在 ClearLogs error 文案上加上「If logs do not appear after clearing, restart the service」的 hint。
- **Risk**: 使用者對 log 檔可寫但對父目錄無 traverse 權限時，`OpenFile` 會回 EACCES，被誤判為「需要 Admin Mode」。 → Mitigation: 接受這個誤判；fallback 到 helper 路徑仍能成功 truncate；UX 損失只是多一次授權，不會出錯。
- **Risk**: helper allowlist 仍有 `/tmp/`、`/private/tmp/`，理論上可以是世界可寫，攻擊者若能在 daemon 啟動前替換 log 路徑為 `/tmp/...` 並把預期被 truncate 的 victim 檔放在那邊，root 會幫他截掉。 → Mitigation: helper 既有設計即假設 plist 路徑值由 root 寫入並驗證，本 change 不放寬該前提；`TruncateLog` 也維持 plist 路徑寫入受限於 `/Library/LaunchDaemons` 不變。並在 spec 內顯式註記 `TruncateLog` 不接受不存在的檔案，避免 root 在 `/tmp/` 憑空建立 0 byte 檔案。
- **Risk**: 把 `LogClearStatus` lazy 查詢綁在 `logType` 切換上，可能在切換瞬間出現舊狀態閃爍。 → Mitigation: 在 `ServiceLogs.vue` 內把按鈕的 disabled 狀態跟 status query 的 pending 狀態 OR 起來，loading 期間先 disable。
- **Risk**: Admin Mode 在使用者按下確認的瞬間被自動關閉（idle timeout）。 → Mitigation: backend 收到 `helper not connected` 時回明確錯誤，前端 alert 文案提醒重新啟用 Admin Mode；不必在前端額外做 polling。
