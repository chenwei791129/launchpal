## Context

`internal/launchctl` 目前以 `bool` 建模 `KeepAlive`：`Service.KeepAlive bool` 與 `ServiceConfig.KeepAlive bool`（`types.go`）。讀取端 `parseKeepAlive` 把 dict 一律壓平為 `true`；寫入端 `BuildPlistDict`（`plist_encode.go`）只在 `config.KeepAlive` 為真時寫出 `KeepAlive=true`。`Service` 的 plist 解析發生在 `readonly.go`（系統域）與 `user.go`（使用者域），兩者皆呼叫 `parseKeepAlive`，且 `user.go` 的 plist struct 已用 `interface{}` 接收 `KeepAlive`，瓶頸在解析函式而非接收型別。

`Update` 流程（`user.go` 的 `Update` → `writePlist` → `BuildPlistDict`、`system.go` 的 `Update` → `encodePlist` → `BuildPlistDict`）會從 `ServiceConfig` 從零重建 plist dict，不讀舊檔合併，因此任何未建模的鍵都會在編輯時被丟棄。本 change 不修這個通用缺陷（另以 GitHub Issue 追蹤），但 `KeepAlive` 的所有子鍵會被完整建模，因此 KeepAlive dict 不會再失真。

前端在 `CreateServiceModal.vue` 以 `form.runAtLoad` / `form.keepAlive` 兩個 checkbox（約 line 71、75）呈現；`serviceToConfig.ts` 與 `ServiceSummary.vue`、`types/wails.d.ts` 也都以 boolean 對應 `keepAlive`。

## Goals / Non-Goals

**Goals:**

- 後端完整解析與 round-trip `KeepAlive` 的 boolean 與 dictionary 兩種 plist 形式，無資料失真。
- 提供 `SuccessfulExit`、`Crashed`、`AfterInitialDemand` 的可編輯設定，與頂層 `ThrottleInterval`（整數秒）。
- 前端把啟動策略改為三選項 radio（`On Demand` / `Run at Load` / `Keep Alive`），消除「KeepAlive 隱含 RunAtLoad」的 UI 誤導，並保留排程型 service 的 `RunAtLoad=false` 能力。

**Non-Goals:**

- 不提供 `PathState` / `OtherJobEnabled` 的 map 編輯 UI；不提供 `NetworkState` 編輯 UI。三者僅讀取保真並原樣寫回。
- 不實作「保留所有未建模 plist 鍵」的通用資料保真機制（橫切議題，另案）。
- 不變更系統域 / Apple 系統域的唯讀規則。

## Decisions

1. **以 `KeepAliveConfig` 取代 `bool`，新檔 `internal/launchctl/keepalive.go`**
   集中放 `KeepAliveConfig` 型別、`parseKeepAlive`（升級版）與 dict 編碼 helper，使 KeepAlive 的雙形式邏輯與 backup/round-trip 規則集中於單一 seam。`types.go` 僅保留型別引用。

2. **JSON 結構欄位（前端契約）**
   `keepAlive` 由 boolean 改為物件。`Mode` 決定寫出 boolean 還是 dictionary。boolean 子鍵以指標表示「未設定 / true / false」三態，map 子鍵以可空 map 表示保真資料。

3. **NetworkState / PathState / OtherJobEnabled 只保真不編輯**
   解析時讀入並保存於 `KeepAliveConfig`，`BuildPlistDict` 在 `Mode == "dictionary"` 時原樣寫回。前端表單不渲染這三個欄位，但 round-trip 時不得遺失。

4. **`ThrottleInterval` 為獨立頂層欄位**
   放在 `ServiceConfig`，型別 `*int`，沿用 `Schedule.Interval` 的指標模式：`nil` 不寫出，非 nil 才寫 `ThrottleInterval`。`Service` 亦帶 `*int` 供顯示。

5. **前端 radio 模型（建立與編輯共用）**
   單一 `launchPolicy` 狀態，值域 `"onDemand" | "runAtLoad" | "keepAlive"`。
   - **提交映射（radio → plist）**：`onDemand` → 兩鍵皆不寫；`runAtLoad` → `RunAtLoad=true`、無 KeepAlive；`keepAlive` → 寫出 KeepAlive（boolean 或 dictionary），不另外寫 `RunAtLoad`（由 launchd 隱含）。
   - **載入映射（既有 service → radio）**：以 KeepAlive 優先。`KeepAlive.Enabled == true` → `keepAlive`（無論 `RunAtLoad` 為何，因 KeepAlive 隱含 RunAtLoad）；否則 `RunAtLoad == true` → `runAtLoad`；兩者皆否 → `onDemand`。此規則明確處理舊版雙 checkbox 產生的「`RunAtLoad=true` 且 `KeepAlive` 同時存在」服務——一律落在 `keepAlive`，且儲存時不再寫出獨立 `RunAtLoad`，由 launchd 隱含。
   - 此 radio 與進階區同時套用於 `CreateServiceModal.vue`（建立 / clone）與 `pages/services/[name].vue` 的編輯表單；兩者共用相同的 launchPolicy 映射與 KeepAliveConfig round-trip（含非編輯子鍵保真）。所有 UI 字串為英文。

## Implementation Contract

- **資料型別（Go，`internal/launchctl/keepalive.go`）**：
  ```
  type KeepAliveConfig struct {
      Enabled            bool            `json:"enabled"`
      Mode               string          `json:"mode"` // "boolean" | "dictionary"
      SuccessfulExit     *bool           `json:"successfulExit,omitempty"`
      Crashed            *bool           `json:"crashed,omitempty"`
      AfterInitialDemand *bool           `json:"afterInitialDemand,omitempty"`
      NetworkState       *bool           `json:"networkState,omitempty"`
      PathState          map[string]bool `json:"pathState,omitempty"`
      OtherJobEnabled    map[string]bool `json:"otherJobEnabled,omitempty"`
  }
  ```
  `Service.KeepAlive` 與 `ServiceConfig.KeepAlive` 型別由 `bool` 改為 `KeepAliveConfig`。`ServiceConfig` 新增 `ThrottleInterval *int`；`Service` 新增 `ThrottleInterval *int`。
  此外，plist 解碼結構 `plistData`（`user.go`，及 `readonly.go` 對應的解碼路徑）目前沒有 `ThrottleInterval` 欄位，**必須新增** `ThrottleInterval *int` `plist:"ThrottleInterval"` tag，否則 launchd plist 中的該鍵不會被 unmarshal，`Service.ThrottleInterval` 將恆為 nil。

- **解析契約（`parseKeepAlive(v any) KeepAliveConfig`）**：
  - `v == nil` 或鍵不存在 → `{Enabled:false}`。
  - `v` 為 `bool` → `{Enabled:v, Mode:"boolean"}`。
  - `v` 為 `map[string]any` → `{Enabled:true, Mode:"dictionary", ...}`，逐一解析 `SuccessfulExit`/`Crashed`/`AfterInitialDemand`/`NetworkState`（bool→指標）與 `PathState`/`OtherJobEnabled`（map[string]any→map[string]bool）。無法辨識的內層值忽略而非報錯。

- **編碼契約（`BuildPlistDict`）**：
  - `KeepAlive.Enabled == false` → 不寫 `KeepAlive` 鍵。
  - `Mode == "boolean"` → 寫 `KeepAlive = true`。
  - `Mode == "dictionary"` → 寫 `KeepAlive` 為 dict，僅放入非 nil 的子鍵與非空 map。**若 dict 最終沒有任何有效子鍵，則降級寫出 `KeepAlive = true`（boolean 形式），不寫空 dict。** 這是單一確定規則：避免「空 dict 在語意上等同 true 卻看似有條件」的歧義；前端進階區亦以此為準（見下方空 dict 規則）。
  - `ThrottleInterval != nil` → 寫 `ThrottleInterval = *v`，否則略過。

- **行為（使用者觀察）**：
  - 既有 `KeepAlive=true` 的 service 仍顯示為 Keep Alive 並可正常儲存。
  - 既有 `KeepAlive` 為 dict（如 `{SuccessfulExit:false}`）的 service 顯示其細項，且編輯儲存後 plist 仍保留該 dict 與其中的 NetworkState/PathState/OtherJobEnabled。
  - 使用者可建立 / 編輯帶 `SuccessfulExit=false`、`Crashed`、`AfterInitialDemand` 與 `ThrottleInterval` 的 service。
  - 建立表單顯示三選項 radio；選 Keep Alive 才出現進階區。

- **失敗模式**：解析端對未知 / 型別不符的內層值採忽略策略，不使整筆 service 解析失敗（維持既有「跳過無法解析的 service」之上一層 robustness）。前端 radio 必有一個選中值，預設 `runAtLoad`（沿用現有 `runAtLoad: true` 預設）。

- **驗收**：
  - Go 單元測試覆蓋 `parseKeepAlive` 與 `BuildPlistDict` 對 boolean、dictionary（含 NetworkState/PathState/OtherJobEnabled 保真）、與 ThrottleInterval 寫出/略過的 round-trip。
  - 前端 component 測試覆蓋 radio 三選項切換、Keep Alive 進階區顯示、以及提交時 launchPolicy → config 的映射。
  - 既有 clone 相關測試（`CloneUserService.test.ts`、`CreateServiceModal.test.ts`）更新為 radio 模型後通過。
  - `make test` 與 `make lint` 通過。

- **Scope 邊界**：
  - In scope：
    - 後端：`internal/launchctl` 的 types、新檔 keepalive、plist_encode、user（含 `plistData` 新增 `ThrottleInterval` 欄位）、readonly 的解析串接，以及 `user_test.go` 等既有 Go 測試中以 bool 斷言 `KeepAlive` 之處的更新。
    - 前端型別：手寫的 `types/wails.d.ts` 與 Wails 產生的 `wailsjs/go/models.ts`（須以 `wails generate` 重新產生或同步編輯，否則 runtime 綁定型別 stale）。
    - 前端 UI：`CreateServiceModal.vue`（建立 / clone）、`pages/services/[name].vue` 的編輯表單（radio + 進階區 + round-trip）、`serviceToConfig.ts`、`ServiceSummary.vue` 顯示、`ServiceRow.vue` 啟動策略 badge。
  - Out of scope：通用未建模鍵保真（另案，GitHub Issue 追蹤）、PathState/OtherJobEnabled/NetworkState 的編輯 UI、系統/Apple 域唯讀規則變動。`app.go` 僅 by-value 轉送 `ServiceConfig`、無 KeepAlive 直接引用，會自動重編譯，不需手動修改。

## Risks / Trade-offs

- **前端型別 BREAKING**：`keepAlive` 由 boolean 變物件，所有讀寫點（modal、serviceToConfig、ServiceSummary、wails.d.ts、clone 邏輯與其測試）必須同步更新，遺漏會造成 TypeScript 編譯或執行期錯誤。以 `make test`（含 typecheck）作為守門。
- **clone 行為**：clone 既有規則是強制 `RunAtLoad=false`。在 radio 模型下，若來源為 Keep Alive，clone 後應落在哪個 radio 需明確：採「clone 保留來源的 launchPolicy，但若為 runAtLoad 則降為 onDemand 以維持現有『RunAtLoad 強制 false』語意」——細節於 spec scenario 釘死。
- **空 dict 規則（確定行為）**：`Mode=="dictionary"` 但所有子鍵皆未設定時，**統一降級為 boolean `KeepAlive=true`**（見編碼契約），不寫空 dict。前端進階區在使用者把所有條件清空時，亦回退為 boolean 模式呈現。此規則同時拘束 spec scenario，避免實作者在「空 dict / true / 驗證錯誤」三者間各自臆測。
- **既有 Go / 前端測試需同步更新**：`keepAlive` 由 bool 變物件後，`user_test.go` 等以 `service.KeepAlive` 作 bool 斷言之處、`wailsjs/go/models.ts`、clone 測試的 fixture（`keepAlive: true` → 物件）與斷言都會編譯或斷言失敗，須一併更新，以 `make test`（含 typecheck）為守門。
