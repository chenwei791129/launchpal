## Context

LaunchPal 至今沒有任何 user-facing 設定檔。所有偏好（備份目錄、預設日誌路徑）都是寫死在程式碼裡的常數：`internal/backup/backup.go` 的 `~/.launchpal/backups`、`frontend/app/components/CreateServiceModal.vue` 的 `~/Library/Logs/<label>/...` 與 `/var/log/<label>/...`。Settings page 的 Backup Directory 區塊同樣只是顯示一個寫死字串 `~/.launchpal/backups/`。

本變更要做兩件事：

1. **第一次**為專案引入 config 持久化層，能在未來擴充其他偏好（例如把 backup dir 也搬進來）。
2. 在這個基礎上提供 user / system 日誌目錄個別設定。

關鍵約束：
- 既存的 privileged-helper RPC（`internal/privhelper/handlers.go` 的 `EnsureLogAccess`）對 system 域日誌路徑有 allowlist 限制：路徑必須以 `/var/log/`、`/private/var/log/`、`/Library/Logs/`、`/tmp/`、`/private/tmp/` 之一為前綴，且**至少一層子目錄**。`systemLogDir` 設定必須通過同等驗證，否則建立 system service 時 `EnsureLogAccess` 會在 helper 端 reject。
- 既有服務的 plist（`StandardOutPath` / `StandardErrorPath`）一律不動。本變更只影響 `New Service` modal 的預設值。
- 專案無 i18n 框架，UI 字串維持英文（依 `.claude/CLAUDE.md` 規範）。

## Goals / Non-Goals

**Goals:**

- 提供可重用的 `internal/settings` 套件：載入、儲存、預設值、原子寫入、驗證。
- Settings JSON 檔案位置與既有 `~/.launchpal/` 命名空間一致：`~/.launchpal/settings.json`。
- Settings page UI 比照 Backup Storage 區塊樣式，新增 Log Storage section，user / system 兩欄獨立可編輯，含 Save / Reset 行為。
- `New Service` modal 讀取 settings 來生成預設 stdout / stderr 路徑；label 仍為唯一插值點。
- `systemLogDir` 驗證與 helper allowlist 對齊，前端在 Save 時即拒絕無效值，避免延後到 create service 時才報錯。

**Non-Goals:**

- 不把 Backup Directory 改為可編輯（涉及既有備份搬遷，獨立變更處理）。
- 不為既有 services 重寫 plist 中的 log 路徑。
- 不在 modal 內提供 per-service 路徑覆寫。
- 不支援 path template / 自訂檔名格式。
- 不引入 YAML / TOML / 第三方 config 套件；純 JSON + stdlib。
- 不引入 i18n 框架。

## Decisions

### Decision 1: JSON over YAML

選 JSON（`encoding/json`，stdlib）而非 YAML。
- **Why**: 主要編輯入口是 Settings UI，hand-editing 體驗的優先級低；專案目前依賴極簡，避免新增 direct dep。
- **Alternatives considered**: YAML（hand-edit 友善但需引入 `gopkg.in/yaml.v3` 為 direct dep）；TOML（同樣需新 dep）。

### Decision 2: Settings 檔位於 `~/.launchpal/settings.json`

與既有 `~/.launchpal/backups/` 共用同一個應用根目錄，不採用 `~/Library/Application Support/LaunchPal/`。
- **Why**: 一致性 > macOS 慣例。使用者已經熟悉 `~/.launchpal/` 是 LaunchPal 的工作區。
- **Alternatives considered**: `~/Library/Application Support/LaunchPal/settings.json`（Apple 慣例，但分散應用狀態）。

### Decision 3: 兩個獨立欄位 `userLogDir` / `systemLogDir`

不合併為單一欄位、不使用 path template。最終路徑固定為 `<dir>/<label>/<stream>.log`。
- **Why**: User / system 域天然落在不同根目錄（`~/Library/Logs` vs `/Library/Logs`）；template 是 YAGNI，99% 使用者只想換目錄。
- **Alternatives considered**: 單一 template 欄位含 `{label}` / `{stream}` 變數（過度設計）；單一 base dir 同時用於兩域（強迫一邊用奇怪路徑）。

### Decision 4: `systemLogDir` 預設值用 `/Library/Logs`

而非沿用 modal 既有預設 `/var/log`。
- **Why**: 對齊 user 端的 `~/Library/Logs`；`/Library/Logs` 是 Apple 文件推薦的應用程式日誌目錄，比 `/var/log`（Unix daemon 風格）更貼近 macOS 慣例。Helper allowlist 已涵蓋 `/Library/Logs/`，無需額外調整。
- **Alternatives considered**: 沿用 `/var/log`（與 user 端命名不對齊）；不設預設讓使用者必填（違反「合理預設」原則）。

### Decision 5: Validation 在前端與後端各做一次，後端為 source of truth

`UpdateSettings` 後端 binding 在寫檔前必驗；前端在 Save 按下時也預先檢查並顯示錯誤。
- **Why**: 後端驗證是 source of truth（避免 bypass）；前端預檢查提供即時 UX 回饋，避免 round-trip 才看到錯誤。
- **`systemLogDir` 規則**：絕對路徑、前綴在 `{/var/log/, /private/var/log/, /Library/Logs/, /tmp/, /private/tmp/}` 內、至少一層子目錄（即不可僅為 allowlist root 本身）。Allowlist 常數從 `internal/privhelper` 套件 export 共用，避免兩處硬編造成漂移。
- **`userLogDir` 規則**：非空字串、不含 shell metacharacters；允許 `~` 前綴（執行時展開），允許絕對路徑；不限制 allowlist。
- **Alternatives considered**: 只前端驗證（容易 bypass）；只後端驗證（UX 差，需 round-trip）。

### Decision 6: Atomic write via temp file + rename

`Save()` 寫入到 `<settings>.tmp`，`fsync`，再 `rename` 覆蓋目標。
- **Why**: 避免在寫檔過程崩潰留下半截 JSON 導致下次啟動 panic。
- **Alternatives considered**: 直接覆寫（不安全）；使用 `lockedfile`（過度設計）。

### Decision 7: 缺檔即用預設值，不提示使用者

`Load()` 在檔案不存在時回傳 `Default()` 結果且不報錯；首次 `Save()` 才實際建立檔案。
- **Why**: 首次啟動體驗應該透明。使用者不需要知道 settings.json 存在，除非他們改了任何值。
- **Alternatives considered**: 啟動時自動建立預設檔（產生不必要的檔案系統寫入）；缺檔報錯（破壞首次體驗）。

### Decision 8: Settings 變更套用於下一個 `New Service` modal 開啟

不向已開啟的 modal 廣播變更事件。
- **Why**: Settings 與 New Service 不會同時操作；簡化實作。
- **Alternatives considered**: Wails event bus 廣播（增加一條訊息路徑換取極少價值）。

### Decision 9: 既有服務的 log 路徑不遷移

只影響從本變更生效後**新建立**的服務。既存 plist 的 `StandardOutPath` / `StandardErrorPath` 不會被掃描或改寫。
- **Why**: 遷移既有服務需要備份、diff、helper 寫入權限（system 域），實作成本與失敗風險高，不在 scope 內。
- **Alternatives considered**: 提供「Migrate existing services」按鈕（屬於另一個變更，未來再評估）。

## Risks / Trade-offs

- **Risk**: 使用者把 `systemLogDir` 設成 helper allowlist 外的路徑（例如自家 `/Users/foo/logs`）。
  → Mitigation: Save 時前端 + 後端雙重驗證，配合明確錯誤訊息（「Path must start with one of: ...」）。

- **Risk**: 使用者升級到本版本後，原本 `New Service` 預設 `/var/log` 變成 `/Library/Logs`，可能造成困惑（既有腳本/監控以為 daemon 會寫到 `/var/log`）。
  → Mitigation: 既有服務不受影響（plist 內路徑已固定）；CHANGELOG / release notes 明確說明預設值變更，使用者若仍偏好 `/var/log` 可在 Settings page 改回。

- **Risk**: `~/.launchpal/settings.json` 損毀（手動編輯失誤）導致 `Load()` 失敗。
  → Mitigation: `Load()` 在 JSON parse 失敗時退回預設值並 log warning（非 fatal），讓 app 仍可啟動；使用者下次儲存時會覆寫掉壞檔。

- **Trade-off**: 不支援 path template ⇒ 無法自訂檔名（例如改成 `out.log` / `err.log`）。
  → Acceptable：99% 使用者只想換目錄；若日後有需要再演進。

- **Trade-off**: `systemLogDir` allowlist 與 helper 同步需要兩處保持一致。
  → Mitigation: 把 allowlist 常數定義在 `internal/privhelper`（或新的 shared package），由 settings 套件 import 使用，單一 source of truth。
