## Why

LaunchPal 目前把 `KeepAlive` 當成單一布林勾選框，且讀取時把字典形式壓平成 `true`（見 `internal/launchctl/types.go` 的 `parseKeepAlive`）。這導致兩個問題：使用者無法表達 launchd 支援的條件式保活（例如「非正常結束才重啟」），且既有使用字典形式的 plist 在 LaunchPal 編輯後會被降級成 `KeepAlive=true`，遺失原本的細部設定。同時 `Run at Load` 與 `Keep Alive` 以兩個獨立 checkbox 呈現，與 launchd「KeepAlive 隱含 RunAtLoad」的語意不符，會誤導使用者。

## What Changes

- 後端以結構化的 `KeepAliveConfig` 取代 `Service.KeepAlive` / `ServiceConfig.KeepAlive` 的純 `bool`，完整解析並 round-trip `KeepAlive` 的 boolean 與 dictionary 兩種形式。
- 字典子鍵支援：`SuccessfulExit`、`Crashed`、`AfterInitialDemand` 提供可編輯 UI；`NetworkState`（launchd 已標記失效）、`PathState`、`OtherJobEnabled` 僅做讀取保真與原樣寫回，本次不提供編輯 UI。
- 新增頂層 `ThrottleInterval`（整數秒）支援，僅在使用者明確設定時寫入 plist。
- **BREAKING**（前端型別）：`Service` / `ServiceConfig` 的 `keepAlive` 由 boolean 改為物件結構，連帶 Wails binding 型別與前端表單需更新。
- 前端把 `Run at Load` / `Keep Alive` 兩個 checkbox 改為單一啟動策略 radio group，三個選項：`On Demand`、`Run at Load`、`Keep Alive`；選 `Keep Alive` 時展開進階選項區，並標註多條件為 OR 且 KeepAlive 隱含 RunAtLoad。同一 radio 與進階區同時套用於**建立**（`CreateServiceModal.vue`）與**編輯**（`pages/services/[name].vue` 的編輯表單）兩個入口，確保既有服務在編輯時能 round-trip。

## Non-Goals

- 不提供 `PathState` / `OtherJobEnabled` 的 map 編輯 UI（讀取保真即可，map 編輯器的 UI 成本與本 issue 效益不成比例，留待後續）。
- 不把 `NetworkState` 做成可編輯欄位（launchd 官方標記其「no longer implemented」）。
- 不處理「保留所有未建模 plist 鍵」的通用資料保真缺陷——該缺陷橫切 `ProcessType`、`Nice`、`MachServices` 等大量鍵，屬獨立議題，另以 GitHub Issue 追蹤。

## Capabilities

### New Capabilities

(none)

### Modified Capabilities

- `core-service-management`: `KeepAlive` 由「bool / dict 壓平成 true」改為完整結構化解析與 round-trip；新增 `ThrottleInterval` 解析與寫入；建立服務的啟動策略由雙 checkbox 改為三選項 radio group，並新增 KeepAlive 進階選項表單。

## Impact

- Affected specs: `core-service-management`
- Affected code:
  - Modified:
    - internal/launchctl/types.go
    - internal/launchctl/plist_encode.go
    - internal/launchctl/user.go
    - internal/launchctl/readonly.go
    - internal/launchctl/user_test.go
    - frontend/app/types/wails.d.ts
    - frontend/wailsjs/go/models.ts
    - frontend/app/utils/serviceToConfig.ts
    - frontend/app/components/CreateServiceModal.vue
    - frontend/app/components/ServiceSummary.vue
    - frontend/app/components/ServiceRow.vue
    - frontend/app/pages/services/[name].vue
  - New:
    - internal/launchctl/keepalive.go
