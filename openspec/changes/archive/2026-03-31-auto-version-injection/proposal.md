## Why

應用程式中版本號目前是硬寫的 `v0.1.0`（StatusBar 和 Settings 頁面），但實際最新版已是 v1.6.0。版本號由 release-please 在 CI 中自動產生 tag，手動維護原始碼中的版本字串容易遺漏且不可靠。需要在建置時自動注入版本號。

## What Changes

- 在 Go 後端新增 `version` 變數，透過 `-ldflags` 在建置時注入版本號
- 新增 `GetVersion()` Wails binding API，讓前端可以動態取得版本
- 修改 CI build workflow，從 release-please 的 `tag_name` 取得版本並傳入建置指令
- 前端 StatusBar 和 Settings 頁面改為呼叫後端 API 取得版本，移除硬寫的字串
- 本地開發時預設顯示 `dev`

## Capabilities

### New Capabilities

- `build-version-injection`: 建置時版本號注入機制，涵蓋 Go ldflags 注入、Wails binding API、CI pipeline 傳遞

### Modified Capabilities

（無）

## Impact

- 受影響程式碼：
  - `main.go` — 新增 version 變數，傳遞給 App
  - `app.go` — 新增 `GetVersion()` 方法
  - `frontend/app/components/StatusBar.vue` — 改用動態版本
  - `frontend/app/pages/settings.vue` — 改用動態版本
  - `.github/workflows/build.yml` — 接收 version input，加入 ldflags
  - `.github/workflows/release-please.yml` — 傳遞 tag_name 給 build workflow
  - `Makefile` — 可能不需改動（本地用預設值 `dev`）
