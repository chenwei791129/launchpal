## Why

LaunchPal 目前把新建服務的日誌路徑寫死在前端（`CreateServiceModal.vue`）：user 服務固定走 `~/Library/Logs/<label>/`，system 服務固定走 `/var/log/<label>/`。使用者沒有任何方式調整預設位置，也沒有 config 持久化層可以擴充。需要引入第一個設定檔機制，讓使用者能在 Settings page 個別設定 user 與 system 域的日誌目錄，並為未來其他偏好（例如備份目錄）預留可重用的基礎設施。

## What Changes

- 新增 `internal/settings` 套件，負責讀寫 `~/.launchpal/settings.json`（atomic write、首次啟動使用內建預設值、不存在時不報錯）。
- Settings 結構暴露兩個獨立欄位：`userLogDir`（預設 `~/Library/Logs`）與 `systemLogDir`（預設 `/Library/Logs`）。
- 新增 Wails bindings：`GetSettings`、`UpdateSettings`，後者執行驗證後寫檔。
- `systemLogDir` 必須通過驗證：路徑前綴須在 privileged-helper 的 log allowlist（`/var/log/`、`/private/var/log/`、`/Library/Logs/`、`/tmp/`、`/private/tmp/`）內，且至少一層子目錄。`userLogDir` 不限制前綴但須為非空。
- 前端新增 `useSettings` composable，Settings page 在 Backup Storage 區塊之後加入 Log Storage section，提供兩欄可編輯路徑（含 Save / Reset 按鈕）。
- `CreateServiceModal.vue` 的 `logPaths` computed 改為從 settings 取值，配合 label 組成最終 stdout / stderr 路徑。
- 既有服務的 `StandardOutPath` / `StandardErrorPath` 不會被掃描或改寫；本變更只影響 `New Service` modal 的預設值。

## Non-Goals

- 不把 Backup Directory 改為可編輯（涉及既有備份搬遷，屬另一變更）。
- 不為既有 services 進行 plist 路徑遷移。
- 不在 `New Service` modal 內提供 per-service 路徑覆寫欄位。
- 不支援 path template（例如 `{label}/{stream}.log`）；最終檔名固定為 `<dir>/<label>/stdout.log` 與 `<dir>/<label>/stderr.log`。
- 不引入 YAML 或其他格式（純 JSON + stdlib `encoding/json`）。
- 不引入 i18n 框架；UI 文字維持英文（依專案 CLAUDE.md 規範）。

## Capabilities

### New Capabilities

- `app-settings`: JSON config 持久化層，定義 settings 檔位置、序列化格式、預設值、原子寫入、驗證契約，以及 GetSettings / UpdateSettings 兩個 Wails binding 的行為。
- `log-path-customization`: 使用者於 Settings page 個別設定 user / system 日誌根目錄，影響 `New Service` modal 的預設 stdout / stderr 路徑生成規則。

### Modified Capabilities

(none)

## Impact

- Affected specs:
  - New: `app-settings`, `log-path-customization`
- Affected code:
  - New:
    - internal/settings/settings.go
    - internal/settings/settings_test.go
    - frontend/app/composables/useSettings.ts
  - Modified:
    - app.go
    - frontend/app/pages/settings.vue
    - frontend/app/components/CreateServiceModal.vue
    - frontend/app/types/wails.d.ts
    - .claude/CLAUDE.md
