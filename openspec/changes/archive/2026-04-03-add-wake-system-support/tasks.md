## 1. 後端型別與資料模型

- [x] 1.1 在 `internal/launchctl/types.go` 的 `Service` struct 新增 `WakeSystem bool` 欄位（`json:"wakeSystem"`），實作 WakeSystem field in Service and ServiceConfig 規格
- [x] [P] 1.2 在 `internal/launchctl/types.go` 的 `ServiceConfig` struct 新增 `WakeSystem bool` 欄位（`json:"wakeSystem"`）
- [x] [P] 1.3 在 `internal/launchctl/user.go` 的 `plistData` struct 新增 `WakeSystem bool` 欄位（`plist:"WakeSystem"`）

## 2. 後端讀取支援

- [x] [P] 2.1 在 `internal/launchctl/user.go` 的 `Get()` 方法中，將 `pd.WakeSystem` 賦值給 `service.WakeSystem`，實作 WakeSystem plist read support across all managers（UserManager）
- [x] [P] 2.2 在 `internal/launchctl/system.go` 的 `Get()` 方法中，解析 plist 的 `WakeSystem` 欄位並賦值給 `Service.WakeSystem`，實作 WakeSystem plist read support across all managers（SystemManager）
- [x] [P] 2.3 在 `internal/launchctl/apple_system.go` 的 `Get()` 方法中，解析 plist 的 `WakeSystem` 欄位並賦值給 `Service.WakeSystem`，實作 WakeSystem plist read support across all managers（AppleSystemManager）

## 3. 後端寫入支援

- [x] 3.1 在 `internal/launchctl/user.go` 的 `writePlist()` 方法中，當 `config.WakeSystem` 為 `true` 時寫入 `pd["WakeSystem"] = true`，實作 WakeSystem plist write support 規格

## 4. 後端測試

- [x] 4.1 在 `internal/launchctl/user_test.go` 新增測試：建立含 WakeSystem 的服務，驗證 plist 輸出包含 `WakeSystem` key
- [x] [P] 4.2 在 `internal/launchctl/user_test.go` 新增測試：讀取含 WakeSystem 的 plist，驗證 `Service.WakeSystem` 為 `true`
- [x] [P] 4.3 在 `internal/launchctl/user_test.go` 新增測試：讀取不含 WakeSystem 的 plist，驗證 `Service.WakeSystem` 為 `false`

## 5. 前端排程表單

- [x] 5.1 在 `frontend/app/components/ScheduleForm.vue` 新增 WakeSystem toggle（checkbox），僅在排程啟用時顯示，實作 WakeSystem toggle in schedule form 規格
- [x] 5.2 `ScheduleForm` 新增 `wakeSystem` prop（`modelValue` 之外）與 `update:wakeSystem` emit，將 toggle 狀態傳遞給父元件，實作 WakeSystem state is emitted with schedule config

## 6. 前端整合

- [x] [P] 6.1 在 `frontend/app/components/CreateServiceModal.vue` 接收 `ScheduleForm` 的 `wakeSystem` 事件，將值帶入 `ServiceConfig.wakeSystem` 提交給後端
- [x] [P] 6.2 在 `frontend/app/pages/services/[name].vue` 的編輯模式中，初始化 `wakeSystem` 狀態並接收 `ScheduleForm` 的更新，提交時帶入 `ServiceConfig.wakeSystem`

## 7. 前端顯示

- [x] 7.1 在 `frontend/app/components/ServiceSummary.vue` 新增 "Wake System" 欄位顯示，值為 "Yes" 或 "No"，實作 WakeSystem display in ServiceSummary 規格

## 8. 型別更新

- [x] [P] 8.1 更新 `frontend/app/types/wails.d.ts`，在 `Service` 和 `ServiceConfig` 型別中新增 `wakeSystem` 欄位
- [x] [P] 8.2 執行 Wails binding 生成或手動更新 `frontend/wailsjs/go/models.ts`
