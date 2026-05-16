## Problem

LaunchPal 的 Create New Service 表單把 Program Path 標記為必填，但這違反 launchd 的實際語意。launchd 規則是 `Program` 與 `ProgramArguments` 至少擇一即可；若 plist 把完整命令放進 Arguments（例如 Synology Drive 的 `/usr/bin/open '/Applications/Synology Drive Client.app'`），Program 必須留空，否則 launchd 會把 `ProgramArguments[0]` 當成 `argv[0]`（依慣例是程式名），執行不會符合預期。

目前的症狀：
- `frontend/app/components/CreateServiceModal.vue` 在 Program Path input 加了 `required` 屬性，提交按鈕在 `!form.program` 時 disabled，`handleSubmit` 也有對應的 guard。使用者完全無法透過 Create 表單建立「Program 空白、命令放在 Arguments」的服務。
- `frontend/app/pages/services/[name].vue` 的 Edit 頁雖沒有 `required`，但同樣缺少「Program 與 Arguments 至少擇一」的 cross-field 驗證，可以兩者皆空白存檔，產生 launchd 無法載入的壞 plist。
- 後端 `internal/launchctl/` 的 Create/Update 也沒有檢查，任何繞過前端的路徑（例如 IPC 直呼）都能寫出壞 plist。

## Root Cause

最初設計表單時把 Program 與 Arguments 視為「程式」與「附加參數」兩種不同性質的欄位，所以理所當然地把 Program 設成必填。但 launchd 的 plist 規格其實把這兩個欄位視為「描述同一個程序的兩種寫法」：

- 只有 `Program`：執行 `Program`，無參數。
- 只有 `ProgramArguments`：`ProgramArguments[0]` 即 executable，後面是參數。
- 兩者都有：`Program` 是 executable，`ProgramArguments` 變成 `argv`（含 `argv[0]`）。

第三種寫法在實務上很罕見，第二種寫法則很常見（特別是命令包含空格、需要 shell-like 引號處理時，例如 `open '/Applications/...'`）。後端 `internal/launchctl/plist_encode.go` 的 `BuildPlistDict` 已經正確支援這三種組合 —— 兩個欄位都以 `if non-empty` 條件式輸出。**問題只出在前端表單與缺乏後端最低限度的 invariant 防線。**

## Proposed Solution

### 前端：將「Program OR Arguments」改為跨欄位驗證

1. **CreateServiceModal**：
   - 移除 Program Path input 的 `required` 屬性。
   - 將 Submit 按鈕 disabled 條件由 `loading || !form.label || !form.program` 改為 `loading || !form.label || (!form.program && !argumentsText.trim())`。
   - `handleSubmit` 的早退 guard 同步改為「label 必填，且 Program 與 Arguments 至少擇一」。

2. **services/[name].vue（Edit Tab）**：
   - 同樣加入 cross-field 驗證：當 program 與 arguments 都空白時，Save Changes 按鈕 disabled 並顯示錯誤提示，禁止寫入。

3. **UX hint（兩處共用文案）**：
   - 在 Program Path label 下加說明：`Optional. Leave empty if the executable is provided as the first item in Arguments.`
   - 用語為英文，遵守 `.claude/CLAUDE.md` 規則。

### 後端：最後一道 invariant 防線

在 `internal/launchctl/` 的 user 與 system Create/Update 入口加 validate：當 `Program == "" && len(Arguments) == 0` 時直接回傳錯誤（訊息明確：`service must specify either Program or at least one argument in Arguments`）。這是 defense-in-depth：避免任何繞過前端的呼叫路徑（單元測試、IPC、未來新介面）寫出壞 plist。

## Non-Goals

- **不**重新設計 Program / Arguments 的 UI 排列順序或語意分隔（例如不把兩欄合併為單一指令輸入框）。維持現有 input 結構，僅修驗證。
- **不**改動後端 `BuildPlistDict` 行為 —— 它已是正確語意。
- **不**對既存的 plist 做資料遷移或回填（壞 plist 由 launchd 在載入時自己拒絕，不在本次處理範圍）。
- **不**處理「Program 與 Arguments 同時填寫」的進階教學或警告 —— 那是 launchd 合法用法，現行 UI 不阻擋亦不額外提示。
- **不**新增 i18n 框架；新加的 hint 文字保持英文硬編碼。

## Success Criteria

- 使用者能在 Create New Service 表單中只填 Label 與 Arguments、Program Path 留空，並成功建立服務；建出的 plist 只含 `Label` 與 `ProgramArguments`，launchd 能正常載入。
- 使用者在 Edit Tab 把 Program 與 Arguments 都清空時，Save Changes 按鈕 disabled 且顯示錯誤訊息，無法寫入壞 plist。
- 後端 `UserManager.Create` / `UserManager.Update` / `SystemManager.Create` / `SystemManager.Update` 在 `Program == ""` 且 `len(Arguments) == 0` 時回傳具語意的錯誤。
- 既存「Program 有填」的服務行為完全不變，Create/Update/list/read 路徑無回歸。
- `make test` 與 `make lint` 通過；新增至少一個後端單元測試覆蓋「兩者皆空」拒絕路徑。

## Capabilities

### New Capabilities

(none)

### Modified Capabilities

- `core-service-management`: 將 `CRUD operations for user services` 的 Create/Update 行為擴充，新增「ServiceConfig 必須 Program 或 Arguments 至少擇一非空」的驗證；其他既有 scenario 不變。

## Impact

- Affected specs: `core-service-management`
- Affected code:
  - Modified:
    - frontend/app/components/CreateServiceModal.vue
    - frontend/app/pages/services/[name].vue
    - internal/launchctl/user.go
    - internal/launchctl/system.go
    - internal/launchctl/user_test.go
  - New:
    - (none)
  - Removed:
    - (none)
