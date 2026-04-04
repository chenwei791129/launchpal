## Context

LaunchPal 的 `UserManager` 目前使用 legacy `launchctl load`/`unload` 來啟停使用者層級的 LaunchAgent 服務。這些指令在 macOS man page 中被歸類為 LEGACY SUBCOMMANDS，Apple 建議改用 `bootstrap`/`bootout`。新指令從 macOS 10.11 El Capitan 開始完整可用，而 LaunchPal 的目標平台（Apple Silicon Mac）最低為 macOS 11.0 Big Sur，不存在向下相容問題。

目前受影響的程式碼位於 `internal/launchctl/user.go`：
- `Start()` 呼叫 `launchctl load <plistPath>`（第 298 行）
- `Stop()` 呼叫 `launchctl unload <plistPath>`（第 321 行）
- `Restart()` 透過 `Stop()` + `Start()` 組合，無直接 launchctl 呼叫

## Goals / Non-Goals

**Goals:**

- 將 `Start()` 和 `Stop()` 遷移至 `bootstrap`/`bootout`
- 利用新指令的正確 exit code 改善錯誤偵測
- 維持現有行為：Stop 失敗時不中斷流程（如服務未載入）

**Non-Goals:**

- 不遷移 `launchctl list`（狀態查詢），留待 Phase 2
- 不改變 `Restart()`、`Create()`、`Update()`、`Delete()` 的邏輯（它們呼叫 `Stop()`/`Start()`，會自動受益）
- 不處理 System / Apple System 管理器（唯讀，不涉及 load/unload）

## Decisions

### 使用 `gui/<uid>` 作為 domain target

新版 launchctl 要求明確指定 domain target。對 LaunchAgent（使用者層級服務），有三種 domain 可選：

| Domain | 說明 |
|---|---|
| `user/<uid>/` | 使用者 domain，可獨立於 GUI 登入存在 |
| `gui/<uid>/` | GUI login domain，與使用者登入 session 綁定 |
| `login/<asid>/` | 用 audit session ID 指定，不方便 |

選擇 `gui/<uid>/`，因為 `~/Library/LaunchAgents` 中的服務屬於使用者登入 session 的 Aqua context，與 GUI domain 語意最吻合。UID 透過 Go 的 `os.Getuid()` 取得，在 `Start()`/`Stop()` 呼叫時格式化為字串。

### bootstrap 語法：傳入 plist 路徑

```
launchctl bootstrap gui/<uid> <plistPath>
```

與 legacy `load` 一樣接受 plist 路徑。若服務已載入，`bootstrap` 會回傳非零 exit code 和錯誤訊息「service already loaded」，`Start()` 應將此錯誤傳遞給呼叫端。

### bootout 語法：使用 service-target（label）

```
launchctl bootout gui/<uid>/<label>
```

`bootout` 支援兩種形式：路徑或 label。選擇 label 形式，因為 `Stop()` 已經透過 `Get()` 取得了 `service.Label`，使用 label 更精確且不依賴檔案路徑。

若服務未載入，`bootout` 會回傳非零 exit code（error 3: No such process）。`Stop()` 目前已忽略 `unload` 的錯誤，遷移後應維持相同行為：忽略 `bootout` 的錯誤，繼續後續的 pgrep/kill 流程。

### UID 格式化輔助函數

新增一個 package-level 的輔助函數 `guiDomain()` 回傳 `fmt.Sprintf("gui/%d", os.Getuid())`，避免在 `Start()` 和 `Stop()` 中重複格式化邏輯。

## Risks / Trade-offs

- **已載入服務重複 bootstrap** → `bootstrap` 會回報錯誤「service already loaded」。`Start()` 原本在 `load` 失敗時就回傳錯誤，行為一致，無需特殊處理。
- **未載入服務 bootout** → `bootout` 回傳非零 exit code。`Stop()` 已忽略此類錯誤，維持現狀。
- **測試限制** → `bootstrap`/`bootout` 需要真實的 launchd 環境，無法在 unit test 中呼叫。現有測試不直接測試 `Start()`/`Stop()`（它們呼叫外部指令），遷移後測試策略不變。
