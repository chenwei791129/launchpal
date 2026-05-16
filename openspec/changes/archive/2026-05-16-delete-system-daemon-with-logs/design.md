## Context

LaunchPal 的 Admin Mode 透過 session-scoped privhelper 對 `/Library/LaunchDaemons/` 執行寫入操作。現有的刪除流程為：

1. `launchctl bootout system/<label>` — 將 daemon 從 launchd 卸載
2. `DeletePlist` RPC → helper 備份 plist 後刪除 `/Library/LaunchDaemons/<name>.plist`

刪除後，plist 中 `StandardOutPath` / `StandardErrorPath` 指向的 log 檔案（通常位於 `/var/log/<service>/`）會留在磁碟上。

**先前的相關工作（在本 change 暫存期間落地）**：

- `session-privileged-helper` 已建立 `EnsureLogAccess` RPC 及共享的 log path allowlist（`/var/log/`, `/private/var/log/`, `/Library/Logs/`, `/tmp/`, `/private/tmp/`），存於 `internal/privhelper.SystemLogPathPrefixes`，並且供 `internal/settings.Validate` 與 helper 端共用。
- `feat(logs)` (#29) 已新增 `TruncateLog` RPC，沿用相同 allowlist 與 `validateLogPath` 驗證器，並引入 `syscallNoFollow`（darwin 為 `syscall.O_NOFOLLOW`，其他平台為 0）作為共用 symlink 防護常數。`DeleteLogPaths` 是這套成熟模式的第三個應用，可直接重用 `validateLogPath`、`SystemLogPathPrefixes`、`syscallNoFollow`，不需新增驗證層。
- `fix-program-required-validation`（最新落地）示範了把跨 manager 共用的驗證 helper 放在 `internal/launchctl/user.go` 並由 SystemManager 引用的做法；本 change 的 `DeleteServiceOptions` 型別與 `DeleteWithOptions` 方法在程式碼放置上採同一原則（型別放 `internal/launchctl/types.go`，方法僅在 `system.go` 暴露）。

本 change 延伸相同安全模型，新增 `DeleteLogPaths` RPC。

## Goals / Non-Goals

**Goals:**

- 讓使用者在刪除 system daemon 時，可選擇性地一併刪除 log 檔案與清空後的 parent 目錄
- 以 helper 執行刪除（root 身分），沿用 `EnsureLogAccess` 的 allowlist 與 `O_NOFOLLOW` 安全模式
- UI 預設不勾選，保持保守行為（log 常有除錯價值）
- 刪除 log 不建立備份（log 體積大、屬執行期產物，使用者勾選即視同知情同意）

**Non-Goals:**

- User services 的 log 清除（user 有寫入權，可自行處理；列為 future extension）
- Apple System Services 的 log 清除（SIP 保護，無意義）
- Log 路徑預覽或確認 UI（避免 UX 過重）
- Log 備份（刻意排除）

## Decisions

### DeleteLogPaths RPC 設計

**決策**：新增獨立的 `DeleteLogPaths` RPC，params 為 `{ Paths []string }`。

**理由**：與 `DeletePlist` 分開呼叫，讓兩者各自有獨立的錯誤語意。若 `DeletePlist` 成功但 `DeleteLogPaths` 失敗（如 log 已不存在），不影響刪除 daemon 的主要結果；前端可選擇顯示 warning 而非 error。合併進 `DeletePlist` 會讓 handler 責任混雜。

**替代方案**：在 `DeletePlist` params 中加入 `logPaths []string` 欄位。拒絕原因：handler 責任混雜，且無法獨立重試。

### 路徑驗證策略（allowlist + lstat）

**決策**：`handleDeleteLogPaths` 對每條路徑套用與 `validateLogPath`（`EnsureLogAccess` 所用）相同的 allowlist 驗證，同時使用 `os.Lstat` + 型別判斷取代 `os.Stat`，確保不跟隨 symlink。

**理由**：`O_NOFOLLOW` 語意在 Go 標準函式庫的 `os.Remove` 不直接支援；使用 `lstat` 先確認路徑為普通檔案（`ModeSymlink` 未設定）再 `os.Remove` 可達到等效防護。Parent 目錄清除也先 `lstat` 確認為目錄後再 `os.Remove`（只有空目錄才成功）。

**Allowlist**（沿用 `EnsureLogAccess`）：
- `/var/log/`
- `/private/var/log/`
- `/Library/Logs/`
- `/tmp/`
- `/private/tmp/`

**額外守則**：路徑必須在 allowlist 子目錄一層以上（`/var/log/foo.log` 拒絕，`/var/log/foo/bar.log` 接受），與 `EnsureLogAccess` 一致。

### Parent 目錄清除策略

**決策**：刪除 log 檔後，嘗試 `os.Remove(parent)`。若 parent 目錄非空，`Remove` 會回傳 `syscall.ENOTEMPTY` 並靜默忽略（不視為錯誤）。只清一層 parent（不遞迴向上）。

**理由**：只清一層足以處理典型情境（如 `/var/log/jeff.test/`）。遞迴向上刪除風險過高（例如誤刪 `/var/log/` 本身）。

**替代方案**：遞迴向上刪除。拒絕原因：安全風險，且 allowlist 的「至少一層子目錄深」規則已阻止刪到 root log 目錄，但向上遞迴可能刪到 allowlist 邊界。

### `Manager` interface 相容性

**決策**：`Manager` interface 的 `Delete(name string) error` 簽名**不改變**。`SystemManager` 新增獨立的 `DeleteWithOptions(name string, opts DeleteServiceOptions) error` 方法。`App.DeleteSystemService` 呼叫 `DeleteWithOptions`，`UserManager.Delete` 不受影響。

**理由**：更動 interface 簽名會要求所有 Manager 實作（UserManager、SystemManager、AppleSystemManager）同步更新，工程量大且 UserManager 暫不需要此功能。新增 `DeleteWithOptions` 只在 SystemManager 層提供，清楚表達「這是 system-only 功能」。

**替代方案**：更新 interface 加入 `options`。拒絕原因：User / Apple manager 不需要此 options，強制實作空方法違反最小介面原則。

### UI 預設行為

**決策**：checkbox 預設 `false`（不勾選）。

**理由**：log 常包含除錯資訊；刪除後不可復原（不備份）。保守預設降低使用者誤操作風險。

### `DeleteLogPaths` 部分失敗處理

**決策**：helper 逐一處理每條路徑，收集錯誤後一次回傳（partial success）。client 端若有任何錯誤，以 warning 方式向前端回報，不阻斷整體刪除結果。

**理由**：log 檔可能因 service 從未啟動而不存在（`ENOENT` 是合理情境），不應因此讓整個刪除操作顯示為失敗。

## Risks / Trade-offs

- **[Risk] 使用者誤刪仍有用的 log** → 預設不勾選 + dialog 明確說明「log files will be permanently deleted」；使用者必須主動選擇。
- **[Risk] `StandardOutPath` 指向 allowlist 外的非預期路徑** → allowlist 驗證在 helper 端強制執行，拒絕非 allowlist 路徑並回傳錯誤，前端顯示 warning。
- **[Trade-off] Log 不備份** → 明確設計決策，proposal 已記錄 trade-off。未來若有需求可加入 log 備份（但需處理大檔案問題）。
- **[Risk] Parent 目錄含有非 log 的其他檔案** → 使用 `os.Remove`（非 `RemoveAll`），只有空目錄才會被刪除；有其他檔案時靜默跳過，不影響那些檔案。
- **[Risk] Symlink 攻擊** → `lstat` + 型別檢查確保目標為普通檔案或目錄，symlink 被拒絕，不跟隨。
