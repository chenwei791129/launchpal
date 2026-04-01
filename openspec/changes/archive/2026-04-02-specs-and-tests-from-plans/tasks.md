## 1. 建立 Spec 檔案

- [x] 1.1 [P] 將 `core-service-management` spec 從 change 複製到 `openspec/specs/core-service-management/spec.md`
- [x] 1.2 [P] 將 `launchdaemons-readonly` spec 從 change 複製到 `openspec/specs/launchdaemons-readonly/spec.md`

## 2. Backup 模組測試（`internal/backup/backup_test.go`）

- [x] 2.1 [P] 測試 backup creation — 驗證 Create 建立 `.plist` 和 `.meta.json` 檔案，metadata 包含 originalPath（對應 Requirement: Backup creation）
- [x] 2.2 [P] 測試 list backups — 驗證 List 回傳按時間倒序排列的備份清單，無備份時回傳空 list（對應 Requirement: List backups）
- [x] 2.3 [P] 測試 get backup content — 驗證 GetContent 讀取備份檔案內容，不存在時回傳 error（對應 Requirement: Get backup content）
- [x] 2.4 [P] 測試 restore backup — 驗證 Restore 複製備份到目標路徑，不存在時回傳 error（對應 Requirement: Restore backup）
- [x] 2.5 [P] 測試 auto-prune — 建立 12 個備份後驗證只保留最近 10 個（對應 Requirement: Backup creation 的 auto-prune 場景）

## 3. UserManager 測試（`internal/launchctl/user_test.go`）

- [x] 3.1 [P] 測試 list user services 和 get service details from plist — 用 temp dir 建立多個 XML plist，驗證 List 回傳正確數量並跳過非 plist 檔案，驗證 Get 回傳的 Service 所有欄位正確，包含 Type="user"、ReadOnly=false、PlistFormat="xml"（對應 Requirement: List user services、Requirement: Get service details from plist）
- [x] 3.2 [P] 測試 KeepAlive 為 dict 的情況 — 驗證 KeepAlive dict 被解析為 true（對應 Requirement: Get service details from plist 的 KeepAlive dict 場景）
- [x] 3.3 [P] 測試 detect plist format — 驗證 binary plist（bplist 開頭）回傳 "binary"、XML plist 回傳 "xml"、空內容回傳 "unknown"（對應 Requirement: Detect plist format）
- [x] 3.4 [P] 測試 read raw plist content 和 read service logs — 驗證 GetPlist 回傳原始檔案內容；驗證 GetLogs 讀取 stdout/stderr、無路徑時報錯、無效 logType 報錯（對應 Requirement: Read raw plist content、Requirement: Read service logs）
- [x] 3.5 [P] 測試 CRUD operations 和 write plist from ServiceConfig — 驗證 Create 建立檔案、duplicate label 報錯、empty label 報錯、Delete 移除檔案；驗證 writePlist 輸出正確的 XML plist 包含 StartInterval 和 StartCalendarInterval（對應 Requirement: CRUD operations for user services、Requirement: Write plist from ServiceConfig）
- [x] 3.6 [P] 測試 parse schedule configuration 和 validate schedule configuration — 驗證 parseSchedule 處理單一/多個 CalendarInterval 和 StartInterval；驗證 validateSchedule 拒絕小於 10 秒的 interval（對應 Requirement: Parse schedule configuration、Requirement: Validate schedule configuration）

## 4. SystemManager 測試（`internal/launchctl/system_test.go`）

- [x] 4.1 [P] 測試 get system service details 和 get raw plist content with format conversion — 用 temp dir 建立 plist，驗證 Get 回傳 Type="system"、ReadOnly=true；驗證 GetPlist 使用 plutil 轉換格式（對應 Requirement: Get system service details、Requirement: Service type and read-only fields、Requirement: Get raw plist content with format conversion）
- [x] 4.2 [P] 測試 read-only managers reject write operations — 驗證所有 write methods 回傳 ErrReadOnlyManager（對應 Requirement: Read-only managers reject write operations）
- [x] 4.3 [P] 測試 list system services — 用 temp dir 驗證 List 正確列出服務、跳過不可讀檔案（對應 Requirement: List system services）
- [x] 4.4 [P] 測試 read system service logs — 驗證 GetLogs 讀取 stdout/stderr 日誌、無路徑時報錯、無效 logType 報錯（對應 Requirement: Read system service logs）

## 5. AppleSystemManager 測試（`internal/launchctl/apple_system_test.go`）

- [x] 5.1 [P] 測試 get apple system service details — 用 temp dir 建立 plist，驗證 Get 回傳 Type="apple-system"、ReadOnly=true（對應 Requirement: Get system service details 和 Requirement: Service type and read-only fields）
- [x] 5.2 [P] 測試 read-only managers reject write operations — 驗證所有 write methods 回傳 ErrReadOnlyManager（對應 Requirement: Read-only managers reject write operations）
- [x] 5.3 [P] 測試 list apple system services — 用 temp dir 驗證 List 正確列出服務（對應 Requirement: List system services）
- [x] 5.4 [P] 測試 read apple system service logs — 驗證 GetLogs 讀取日誌（對應 Requirement: Read system service logs）

## 6. 三種 Manager 的 interface 驗證

- [x] 6.1 [P] 驗證 three manager types with distinct access levels — 確認 UserManager、SystemManager、AppleSystemManager 都實作 Manager interface（對應 Requirement: Three manager types with distinct access levels）

## 7. 驗證與清理

- [x] 7.1 執行所有測試確認通過（`make test`）
- [x] 7.2 檢查是否需要更新 README.md、.claude/CLAUDE.md
