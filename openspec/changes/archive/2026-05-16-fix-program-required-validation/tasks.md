## 1. 後端驗證：CRUD operations for user services 的 invariant 防線

- [x] 1.1 [CRUD operations for user services] 在 `internal/launchctl/user.go` 的 `UserManager.Create` 入口加入「`Program` 與 `Arguments` 至少擇一非空」的檢查，未通過時回傳 `errors.New("service must specify either Program or at least one argument in Arguments")`；以單元測試覆蓋 Create 拒絕路徑（驗證：回傳該錯誤、且 `~/Library/LaunchAgents/<label>.plist` 不存在）。
- [x] 1.2 在 `UserManager.Update` 入口加入相同檢查；以單元測試覆蓋 Update 拒絕路徑（驗證：回傳該錯誤、且原 plist 內容未被改寫）。
- [x] 1.3 在 `internal/launchctl/system.go` 的 `SystemManager.Create` 與 `SystemManager.Update` 入口加入相同檢查並回傳相同錯誤；以單元測試或 fake helper 驗證拒絕路徑不會觸發任何 helper RPC、且不寫入 `/Library/LaunchDaemons/` 任何檔案。
- [x] 1.4 [P] 補一個正面案例的單元測試：以 `Program=""`、`Arguments=["/usr/bin/open", "/Applications/Foo.app"]` 呼叫 `UserManager.Create`，驗證寫出的 plist 只含 `Label` 與 `ProgramArguments`、不含 `Program` 鍵（讀回再用 `plist.Unmarshal` 比對 keys）。

## 2. 前端：Create 表單

- [x] 2.1 在 `frontend/app/components/CreateServiceModal.vue` 移除 Program Path input 的 `required` 屬性；提交按鈕 disabled 條件改為 `loading || !form.label || (!form.program && !argumentsText.trim())`；`handleSubmit` 早退 guard 同步更新為 `if (!form.label || (!form.program && !argumentsText.trim())) return`。驗證：vitest 元件測試覆蓋（a）只填 label 與 arguments 時按鈕 enabled 且 submit 會呼叫 `CreateService`、（b）label 填、program 與 arguments 都空時按鈕 disabled。
- [x] 2.2 [P] 在同檔案的 Program Path label 下方加 hint 文字 `Optional. Leave empty if the executable is provided as the first item in Arguments.`，沿用既有 `text-xs text-gray-500 mt-1` 樣式。驗證：vitest 快照或文字斷言確認 hint 存在於 modal 中。

## 3. 前端：Edit Tab

- [x] 3.1 在 `frontend/app/pages/services/[name].vue` 加 cross-field 驗證：當 `editForm.program` 與 `editArgumentsText.trim()` 都空白時，Save Changes 按鈕 disabled 並顯示錯誤訊息（重用既有 `saveError` 顯示位置或新增 inline error）。驗證：vitest（或對應頁面測試）斷言「兩者皆空時按鈕 disabled」與「至少一個非空時按鈕 enabled」。
- [x] 3.2 [P] 在同檔案 Edit Tab 的 Program Path label 下方加上與 Create 表單一致的 hint 文字 `Optional. Leave empty if the executable is provided as the first item in Arguments.`。驗證：頁面渲染後該文字存在。

## 4. 整合驗證

- [x] 4.1 `make test` 通過（Go tests + frontend vitest + tsc）。
- [x] 4.2 `make lint` 通過（golangci-lint + eslint）。
- [x] 4.3 請使用者執行 `make dev` 並手動驗證：（a）以 Synology Drive 為例（Program 空、Arguments 含 `/usr/bin/open '/Applications/Synology Drive Client.app'`）能透過 Create 表單建立服務並成功啟動、（b）Edit Tab 將 Program 與 Arguments 都清空時無法 Save、（c）原本 Program 有填的既存服務行為不變。
