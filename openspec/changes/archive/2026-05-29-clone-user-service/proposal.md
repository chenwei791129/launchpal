## Why

目前要建立一個與現有 service 設定相近的新 user service，使用者得手動把所有欄位（Program、Arguments、EnvironmentVariables、Schedule、Working Directory 等）抄寫到 New Service modal。當設定欄位多、或來源 service 有 schedule / env vars 時，這個流程容易抄錯且耗時。GitHub issue #17 要求加入 Copy / Clone 動作，讓使用者一鍵複製既有 user service、只調整需要變動的部分。

## What Changes

- 在 `services/[name].vue` header 的 action 區，當 `serviceType === 'user'` 時新增一顆 Copy button（System / Apple System 不顯示）。
- 在 `CreateServiceModal.vue` 加入 optional `prefill?: ServiceConfig` prop；當 modal 以 prefill 模式開啟時，預先填入來源 service 的所有設定（Program、Arguments、WorkingDirectory、Schedule、WakeSystem、KeepAlive、EnvironmentVariables），但 `label` 強制留空、`runAtLoad` 強制設為 `false`。
- Copy button 按下後以當前 service 作為 prefill 開啟 modal；使用者輸入新 label（必要時微調其他欄位）後送出，呼叫既有的 `App.CreateService` binding 寫入 `~/Library/LaunchAgents/<new-label>.plist`。
- modal 的 `@created` event 帶上新 label；detail 頁監聽後使用 `navigateTo('/services/<new-label>?type=user')` 導向新 service 的詳細頁。
- 重複 label 由後端 `UserManager.Create` 既有檢查（`service %s already exists`）回報，前端把 error 訊息塞進 modal 既有的 error 顯示區，不關閉表單。

## Non-Goals (optional)

- 不在後端新增 `CloneService(sourceLabel, newLabel)` Wails binding 或 `UserManager.Clone` method — 純前端組裝 `ServiceConfig` 後呼叫既有 `CreateService` 即達成需求。
- 不支援 clone System service 或 Apple System service（issue AC 明文限制 user services；System daemon 通常與 Homebrew/廠商安裝路徑綁定，clone 到 `~/Library/LaunchAgents` 也無法正常啟動）。
- 不提供「name-only 對話框 → 立即建立 → 進 Edit tab 改其他欄位」的兩步式流程；single prefilled modal 已涵蓋此使用情境，且 Edit tab 目前不支援改 label，兩步式體驗較差。
- 不對 label 做前端預檢（不呼叫 `GetService` 確認是否存在），依賴後端的權威錯誤回報。
- 不為 clone 動作引入獨立的 audit log / 事件追蹤；clone 結果與一般 Create 等價，沒有額外的可追蹤性需求。

## Capabilities

### New Capabilities

(none)

### Modified Capabilities

- `core-service-management`: 在 user service 範疇內新增 "Cloning a user service" requirement，描述 Copy 動作觸發 prefilled CreateServiceModal、`RunAtLoad` 強制 false、label 必須由使用者另外輸入、建立後導向新 service detail 頁、重複 label 由後端錯誤回報的完整流程。

## Impact

- Affected specs: `core-service-management`（新增一條 requirement）
- Affected code:
  - Modified:
    - frontend/app/pages/services/[name].vue
    - frontend/app/components/CreateServiceModal.vue
    - frontend/app/components/__tests__/CreateServiceModal.test.ts
  - New:
    - frontend/app/components/__tests__/CloneUserService.test.ts
  - Removed: (none)
- Affected runtime/UX：user service detail 頁 header 新增一顆 button，且 New Service modal 在 prefill 模式下行為與目前略有差異（label 為空、`runAtLoad` 預設 false）。
- No backend code or Wails bindings change; no `.plist` schema change; no new permissions or RPC paths.
