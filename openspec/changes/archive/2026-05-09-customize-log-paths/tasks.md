## 1. 共用系統日誌 allowlist 抽出

- [x] 1.1 Helper allowlist drift is prevented：把 `internal/privhelper/handlers.go` 內既有的系統日誌 allowlist（`/var/log/`、`/private/var/log/`、`/Library/Logs/`、`/tmp/`、`/private/tmp/`）抽到單一 exported symbol（例如 `privhelper.SystemLogPathPrefixes`），讓 `EnsureLogAccess` 改為引用該常數，並補強既有 helper 測試確認 allowlist 行為不變

## 2. internal/settings 套件（TDD）

- [x] 2.1 撰寫 `internal/settings/settings_test.go`：覆蓋 Settings file location and format、Default settings values、Load settings from disk（缺檔／壞 JSON／部分欄位三個 scenario）、Atomic save to disk（atomic replace、parent dir 自動建立）、Validate settings before save（含 systemLogDir 與 userLogDir 完整 validation matrix；驗證錯誤訊息含欄位名與規則說明）
- [x] 2.2 撰寫 `internal/settings/settings.go`：實作 Settings struct（`UserLogDir`、`SystemLogDir`）、`Default()`、`Load()`、`Save()`、`Validate()`，使用 stdlib `encoding/json`；validator 透過 task 1.1 的共享常數做 systemLogDir prefix 檢查；遵循下列 design 決策——Decision 1: JSON over YAML、Decision 2: Settings 檔位於 `~/.launchpal/settings.json`、Decision 3: 兩個獨立欄位 `userLogDir` / `systemLogDir`、Decision 4: `systemLogDir` 預設值用 `/Library/Logs`、Decision 6: Atomic write via temp file + rename、Decision 7: 缺檔即用預設值，不提示使用者
- [x] 2.3 跑 `go test ./internal/settings/...` 與 `golangci-lint run ./internal/settings/...` 確保套件全綠

## 3. Wails bindings

- [x] 3.1 在 `app_test.go` 加上 GetSettings Wails binding 與 UpdateSettings Wails binding 的測試：首次呼叫 `GetSettings` 回傳 Default、`UpdateSettings` 在 validation 失敗時不寫檔且回傳 error、`UpdateSettings` 成功時寫檔並回傳 `nil`
- [x] 3.2 在 `app.go` 新增 `GetSettings()` 與 `UpdateSettings(s settings.Settings) error` 兩個 Wails 方法
- [x] 3.3 跑 `make build` 觸發 `wails generate` 重新產生 `frontend/wailsjs/go/main/App.{ts,js,d.ts}`，並同步更新 `frontend/app/types/wails.d.ts` 暴露 `Settings` 型別

## 4. Frontend useSettings composable

- [x] 4.1 [P] 撰寫 `frontend/app/composables/__tests__/useSettings.test.ts`：覆蓋 `load()` 透過 `GetSettings` 載入、`save()` 成功時更新本地 cache、`save()` 在驗證失敗時不更新本地 cache 並把 error 暴露給呼叫端
- [x] 4.2 新增 `frontend/app/composables/useSettings.ts`：暴露 reactive `settings` ref 與 `load()`、`save(next)` 方法，內部呼叫 Wails bindings

## 5. Settings page Log Storage section

- [x] 5.1 [P] 撰寫 `frontend/app/pages/__tests__/settings.test.ts` 對 Log Storage section 的測試：覆蓋 Settings page exposes log directory controls 與 Save action validates and persists——初始渲染顯示預設、reset 還原預設、Save 失敗時顯示 inline error 並保留輸入值、Save 成功觸發 `GetSettings` refresh
- [x] 5.2 在 `frontend/app/pages/settings.vue` 的 Backup Storage 區塊後方新增 Log Storage section：User Log Directory 與 System Log Directory 兩個獨立輸入欄位，各帶 Save 與 Reset to Default 按鈕，UI 文字全英文、視覺對齊既有 Backup Storage 樣式；前端在 Save 前做形式預檢查（與後端規則一致），實踐 design 決策 Decision 5: Validation 在前端與後端各做一次，後端為 source of truth

## 6. New Service modal integration

- [x] 6.1 [P] 更新 `frontend/app/components/__tests__/CreateServiceModal.test.ts`：覆蓋 New Service modal sources defaults from settings 與 design 決策 Decision 8: Settings 變更套用於下一個 `New Service` modal 開啟——每次開啟 modal 都重讀 settings、user / system serviceType 的路徑組成符合 spec 中 path composition 表
- [x] 6.2 更新 `frontend/app/components/CreateServiceModal.vue`：在 mount／open 時透過 `useSettings` 重新載入 Settings，`logPaths` computed 改以 settings 為來源組成 `<dir>/<label>/stdout.log` 與 `<dir>/<label>/stderr.log`；落實 design 決策 Decision 9: 既有服務的 log 路徑不遷移，並補一條測試確認 Existing services are not migrated（儲存 settings 不會修改既有 plist 也不會發出 helper RPC）

## 7. 文件與整合驗證

- [x] 7.1 [P] 更新 `.claude/CLAUDE.md`：新增 `~/.launchpal/settings.json` 路徑、Settings page Log Storage 區塊說明、以及系統日誌 allowlist 的單一 source of truth 位置（`privhelper.SystemLogPathPrefixes`）
- [x] 7.2 跑 `make lint && make test` 全綠後，使用 `make dev` 啟動 app，於 Settings page 修改 user / system 兩個 log dir 並 Save，再開 New Service modal（user 與 system 各一）驗證預設路徑跟著改變；額外驗證輸入無效 systemLogDir（例如 `/etc/foo`）時 Save 失敗且 inline error 正確顯示
