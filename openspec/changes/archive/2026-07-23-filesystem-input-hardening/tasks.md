<!-- 各任務描述陳述「交付的行為/契約」與「驗證方式」；檔案路徑僅為定位脈絡。技術符號、路徑、指令保留原文。 -->

## 1. Symlink-safe resolution of log-path arguments（helper log 路徑以 openat 逐段 O_NOFOLLOW 解析）

- [x] [P] 1.1 在 `internal/privhelper/handlers_test.go` 新增測試涵蓋 spec「Symlink-safe resolution of log-path arguments」：以中間目錄為 symlink（指向允許清單外，例如 `/tmp/<link>` → `/etc` 的暫存佈置）分別驅動 EnsureLogAccess（含需建立中間目錄的情境）／TruncateLog／DeleteLogPaths，斷言 root 不會 chmod／截斷／刪除／建檔於允許清單外的真實目標；並斷言合法的巢狀（非 symlink）log 路徑仍成功。驗證：`go test ./internal/privhelper -run` 對應測試綠燈。
- [x] 1.2 於 `internal/privhelper/handlers.go` 將 `ensureLogAccess`（建目錄/parent chmod/touch）、`truncateLog`、`deleteOneLogPath`、`backupExisting` 改為逐段開啟（parent 以 `O_DIRECTORY|O_NOFOLLOW`，葉節點以 `*at`＋`AT_SYMLINK_NOFOLLOW`，建立缺少的中間目錄以 `Mkdirat`，經 `golang.org/x/sys/unix`），取代會跟隨 symlink 的 `os.MkdirAll`/`os.Chmod`；保留 `validateLogPath` 字面檢查為前置；修正誤稱 `O_NOFOLLOW` 已關閉 symlink 替換競態的註解（設計：helper log 路徑以 openat 逐段 O_NOFOLLOW 解析）。行為：中間目錄 symlink（含建目錄）無法逃逸允許清單。驗證：1.1 測試通過。

## 2. Routing name path-traversal confinement（路由 name/Label 以共用 validateRoutingName 限制為單一元件（含系統域））

- [x] [P] 2.1 在 `internal/launchctl/user_test.go` 與 `internal/launchctl/system_test.go` 新增測試涵蓋 spec「Routing name path-traversal confinement」：`validateRoutingName` 對含 `/`、`..`、NUL 或等於 `.`/`..` 的值回傳錯誤、對一般 label（如 `com.example.foo`）通過；並斷言以穿越 `name`/`Label` 呼叫使用者域與系統域（Create/Update/Delete）方法皆不觸及 base 目錄外的檔案。驗證：`go test ./internal/launchctl -run` 對應測試綠燈。
- [x] 2.2 於 `internal/launchctl` 新增共用 `validateRoutingName`，並在 `user.go`、`readonly.go`、以及 `system.go` 的 Create（`config.Label`）／Update／Delete／Start／Stop／Restart（`name`）各方法起始呼叫（設計：路由 name/Label 以共用 validateRoutingName 限制為單一元件（含系統域））。行為：穿越名稱於使用者域與系統域皆被拒且不進行檔案操作。驗證：2.1 測試通過。

## 3. System daemon schedule validation parity（系統域排程範圍驗證對稱，50 筆上限置於系統域 create/update 路徑）

- [x] [P] 3.1 在 `internal/launchctl/system_test.go` 新增測試涵蓋 spec「System daemon schedule validation parity」：`SystemManager.Create/Update` 對超範圍排程（`StartInterval` < 10 或行事曆欄位越界）回傳驗證錯誤且不寫入 plist；對超過 50 筆行事曆項目亦於 create/update 路徑回傳錯誤且不寫入。驗證：`go test ./internal/launchctl -run` 對應測試綠燈。
- [x] 3.2 於 `internal/launchctl/system.go` 的 Create/Update 呼叫既有 `validateSchedule`，並在同一路徑（可回傳錯誤處）檢查 `len(config.Schedule.Schedules) > 50` 即回傳驗證錯誤且不寫 plist；不修改無錯誤通道且為使用者域共用的 `buildCalendarInterval`，以免變更使用者域行為（設計：系統域排程範圍驗證對稱，50 筆上限置於系統域 create/update 路徑）。行為：系統域畸形/超量排程被拒，使用者域行為不變。驗證：3.1 測試通過。

## 4. Helper invokes launchctl by absolute path（helper 以絕對路徑 /bin/launchctl 呼叫）

- [x] [P] 4.1 在 `internal/privhelper/handlers_test.go` 以既有可注入的 `Runner` 斷言涵蓋 spec「Helper invokes launchctl by absolute path」：bootstrap／bootout／kickstart 傳入的命令為 `/bin/launchctl`。驗證：對應 `go test` 綠燈。
- [x] 4.2 於 `internal/privhelper/handlers.go` 以單一常數 `/bin/launchctl` 取代 bootstrap／bootout／kickstart 的 bare name `launchctl`（設計：helper 以絕對路徑 /bin/launchctl 呼叫）。行為：呼叫與 `$PATH` 無關。驗證：4.1 測試通過。

## 5. 驗證與文件

- [x] 5.1 執行 `make test`（Go 測試 + 前端）與 `make lint`，確認全綠且無新違規。驗證：指令輸出無失敗。
- [x] 5.2 更新 `.claude/CLAUDE.md` 相關說明（helper log 路徑逐元件 symlink 安全含 `Mkdirat`、路由 name/Label confinement 含系統域、系統域排程範圍驗證與 create/update 路徑 50 筆上限、helper 以 `/bin/launchctl` 呼叫），使文件與行為一致。驗證：內容審閱對照本 change 的 design 契約。
