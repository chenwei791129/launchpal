## 1. 後端版本注入

- [x] 1.1 [P] 在 `main.go` 新增 `var version = "dev"` 變數，並將 version 傳遞給 `NewApp()`（Version variable with build-time injection）
- [x] 1.2 在 `app.go` 的 `App` struct 新增 `version` 欄位，並實作 `GetVersion() string` 方法作為 Wails binding（Version API binding）

## 2. 前端動態版本顯示

- [x] 2.1 [P] 修改 `frontend/app/components/StatusBar.vue`，移除硬寫的 `v0.1.0`，改為呼叫 `GetVersion()` Wails binding 取得版本（Frontend dynamic version display）
- [x] 2.2 [P] 修改 `frontend/app/pages/settings.vue`，移除硬寫的 `appVersion = 'v0.1.0'`，改為呼叫 `GetVersion()` Wails binding 取得版本（Frontend dynamic version display）

## 3. CI Pipeline 版本傳遞

- [x] 3.1 修改 `.github/workflows/build.yml`，新增 `version` 的 `workflow_call` input 參數，在建置步驟中使用 `-ldflags "-X main.version=$VERSION"` 注入版本（CI pipeline version propagation）
- [x] 3.2 修改 `.github/workflows/release-please.yml`，將 `tag_name` 作為 `version` input 傳遞給 `build.yml`（CI pipeline version propagation）
