## 1. Reuse the existing `CreateServiceModal` via an optional `prefill` prop

- [x] 1.1 為了 Reuse the existing `CreateServiceModal` via an optional `prefill` prop，在 `CreateServiceModal.vue` 新增 optional `prefill?: ServiceConfig | null` prop。當 prop 為 `undefined`/`null` 時，modal 開啟行為與目前完全一致（所有欄位空白、`runAtLoad` 預設 `true`）— 確保向後相容。驗證：`frontend/app/components/__tests__/CreateServiceModal.test.ts` 全部既有測試保持綠燈（`pnpm vitest run CreateServiceModal`）。
- [x] 1.2 當 `prefill` 為有值物件且 modal 開啟時，把 `program`、`workingDirectory`、`keepAlive`、`argumentsText`（由 `serializeShellArgs` 從 `prefill.arguments` 組成）、`envVars`、`schedule`、`wakeSystem` 全部填入，但 `form.label` 設為空字串、`form.runAtLoad` 設為 `false`（決策：`RunAtLoad` 在 clone 時強制為 `false`，無 escape hatch）。驗證：新 vitest case「mounts modal with prefill → asserts each form field equals prefill value except label='' and runAtLoad=false」。
- [x] 1.3 加入 watcher：當 `props.isOpen` 由 `false` 變為 `true` 時，依當下 `props.prefill` 重新初始化內部 state；當 `props.prefill` 物件參考改變時也重新初始化。驗證：vitest case「open→close→reopen with different prefill → form reflects the latest prefill, no stale values from previous open」。
- [x] 1.4 把 `created` event 的 signature 從 `() => void` 改為 `(label: string) => void`（決策：Modal `created` event 帶上新 label，由 detail 頁負責導頁）。`emit('created')` 全部呼叫點補上 `form.label` 作為 payload。驗證：TypeScript 編譯通過（`pnpm typecheck`）+ vitest 斷言「submit → emits 'created' with the submitted label string」。
- [x] 1.5 [P] 確認 `frontend/app/pages/index.vue` 與 `frontend/app/pages/system.vue` 對 `@created` 的既有 listener 在 signature 變更後仍能正常重新拉取 service 列表（listener 可忽略 payload）。驗證：`pnpm typecheck` 通過 + 手動 smoke test「從 list 頁開新 service → modal 關閉 → 列表出現新項目」。

## 2. User service detail 頁加入 Copy 動作（Copy button 放在 detail 頁 header action 區，僅 user service 顯示）

- [x] 2.1 在 `frontend/app/pages/services/[name].vue` header 的 action 區（Stop/Restart/Run Now 同一列）新增一顆 Copy button，icon 採 Heroicons `document-duplicate` outline、tooltip "Copy this service"、配色 `bg-surface-200 hover:bg-surface-100`。Button 僅在 `serviceType === 'user'` 條件下渲染；System / Apple System detail 頁不顯示。當 `service.value` 為 `null`（尚未載入完成）時 button `disabled`。驗證：vitest 元件測試「render with type=user → button exists and enabled when service is loaded; render with type=system or type=apple-system → button does NOT exist in DOM」。
- [x] 2.2 為 Copy 動作建立 `showCloneModal: Ref<boolean>` 與 `cloneSource: Ref<ServiceConfig | null>` 兩個 state；點擊 Copy button 時把當前 `service.value` 轉為 `ServiceConfig` 形狀（label / program / arguments / runAtLoad / keepAlive / environment / workingDirectory / schedule / wakeSystem / stdoutPath / stderrPath）寫入 `cloneSource`，並把 `showCloneModal` 設為 `true`，等同於用 prefill 開啟 `CreateServiceModal`。驗證：vitest 整合測試「click Copy → CreateServiceModal 接收到的 `prefill` prop 等於當前 service 的 ServiceConfig 投影」。
- [x] 2.3 監聽 `CreateServiceModal` 的 `@created` event（payload 為 new label），收到後呼叫 `navigateTo('/services/' + encodeURIComponent(label) + '?type=user')`（決策：Modal `created` event 帶上新 label，由 detail 頁負責導頁）。驗證：vitest case mock `navigateTo`，發出 `created('com.example.copy-dest')` → 斷言 `navigateTo` 被呼叫且參數為 `'/services/com.example.copy-dest?type=user'`。

## 3. 錯誤處理與重複名稱（重複 label 不做前端預檢）

- [x] 3.1 確認 `CreateServiceModal` 在 `App.CreateService` reject 時，把 error 訊息（含 `service <label> already exists`）寫入 `error.value` 並保留所有欄位不清空、不關閉 modal（既有行為，需新增測試以鎖定 contract）。驗證：vitest case mock `window.go.main.App.CreateService` reject with `Error('service com.example.dup already exists')` → 斷言 `error` 區塊顯示該訊息、modal 仍開啟、`form.label` 與其他欄位保留使用者輸入。

## 4. Spec 對齊與 acceptance（Clone a user service）

- [x] 4.1 對應 `Clone a user service` requirement 的所有 spec scenarios（visibility by service type / Pre-filled creation form on clone / Successful clone creates new service and navigates / Duplicate label is rejected inline / User overrides RunAtLoad before submitting）— 為每個 scenario 在 vitest 中至少一個對應測試 case 通過。驗證：`pnpm vitest run` 全綠，且測試名稱與 scenario 名稱可逐一對應（用人工 review 或 vitest `--reporter=verbose` 輸出比對）。
- [x] 4.2 [P] 手動端對端驗證：在實機建立 user service `com.example.copy-source`（任意 Program、加 1 個 env var、加 1 個 schedule、RunAtLoad=true）→ 在 detail 頁按 Copy → 命名 `com.example.copy-dest` 送出 → 確認 `~/Library/LaunchAgents/com.example.copy-dest.plist` 存在、`RunAtLoad` 為 `false`、其他欄位與來源一致、瀏覽器導向 `/services/com.example.copy-dest?type=user`。驗證：`plutil -p ~/Library/LaunchAgents/com.example.copy-dest.plist` 輸出比對來源 plist（除 Label、RunAtLoad、StandardOutPath、StandardErrorPath 之外完全一致）。

## 5. 文件與品管

- [x] 5.1 [P] 更新 `.claude/CLAUDE.md` 的 "User Services" 段落，補上「Supports cloning via Copy button on detail page; `RunAtLoad` is forced to `false` in clones」。驗證：人工 review diff + commit message 含 `docs` prefix。
- [x] 5.2 全套品管：`make lint`、`make test` 全數通過。驗證：兩個指令的 exit code 為 0。
