## Context

安全稽核 run-1 找出一組低風險、彼此獨立的檔案系統/輸入強化缺口，皆已對照程式碼確認：

- root helper 的 log 路徑處理（`ensureLogAccess` 的 `os.MkdirAll` + `os.Chmod(parent)`、`truncateLog`、`deleteOneLogPath`、`backupExisting`）僅以 `validateLogPath`（`filepath.Clean` + 前綴比對）與末端 `O_NOFOLLOW`/`Lstat` 防護；允許清單含世界可寫的 `/tmp/`、`/private/tmp/`，故同 UID 攻擊者可在**中間目錄**植入 symlink 逃逸。實測確認 `Chmod` 會跟隨 symlinked parent、`O_TRUNC|O_NOFOLLOW` 會截斷 symlinked parent 後的真實檔、`Lstat` 對 symlinked parent 後的葉節點回報為一般檔。
- GUI 的 `name`（與 Create 的 `config.Label`）流入 `filepath.Join(base, name+".plist")`，`filepath.Join` 不 confine 到 base。使用者域（`user.go`）、系統域（`system.go` 的 Create/Update/Delete/Start/Stop/Restart）與唯讀域（`readonly.go`）皆有此形態。
- `SystemManager.Create/Update`（`internal/launchctl/system.go`）未呼叫 `validateSchedule`（使用者域有），行事曆展開上限（50 筆）僅在前端 `ScheduleForm.vue`。後端展開後的計數位於 `config.Schedule.Schedules`，`buildCalendarInterval`/`BuildPlistDict` 僅編碼傳入的項目且**無錯誤通道**。
- helper 以 bare name `exec.CommandContext(ctx, "launchctl", ...)` 呼叫（`internal/privhelper/handlers.go`），目前被提權 trampoline 的 PATH 重置緩解。

這些多屬縱深防禦；特別是 log 路徑 symlink 逃逸對能觸達 socket 的呼叫者而言，與 by-design 的 WritePlist+Bootstrap root RCE 完全冗餘，因此是正確性/縱深防禦缺陷而非提權。本 change 只做這組防禦性修正（Change 2）。

稽核 #6（`readLogTail`/`copyFile` 補 `O_NOFOLLOW`）經對抗式 review 後**不納入**：屬使用者權限內（無跨權限提升），安全效益極微，且會造成真實 UX 迴歸——`O_NOFOLLOW` 對 symlink 回傳 `ELOOP`（非 `ENOENT`），使「symlink 但目標尚未建立」的 log 從中性 not-found placeholder 變成紅色錯誤；`copyFile` 加 `O_NOFOLLOW` 會破壞把 plist 以 symlink 指向 dotfiles repo 的還原工作流。淨負面，故排除。

## Goals / Non-Goals

**Goals:**

- 讓 root helper 的 log 路徑操作對「中間目錄 symlink」安全（含建立中間目錄）。
- 讓 GUI 的路由 `name`/`Label` 在使用者域、系統域、唯讀域皆無法穿越到目標目錄之外。
- 讓系統域的排程範圍驗證與使用者域對稱，並於系統域 create/update 路徑強制 50 筆上限。
- 讓 helper 對 `launchctl` 的呼叫不依賴 `$PATH`。

**Non-Goals:**

- Admin Mode 生命週期（Change 1，已 park）。
- helper 二進位簽章/完整性（Finding 1，另案）。
- `readLogTail`/`copyFile` 的 `O_NOFOLLOW`（稽核 #6，經評估排除，理由見 Context）。
- 更動使用者域既有排程行為（其 50 筆上限維持現狀，不新增後端拒絕）。
- 更動 socket 驗證模型、RPC 方法集、或 WritePlist/Bootstrap 的既有語意。

## Decisions

### helper log 路徑以 openat 逐段 O_NOFOLLOW 解析

`validateLogPath` 的字面比對保留（快速拒絕明顯越界），但實際 chmod/open/truncate/remove/建目錄改為以「開啟 parent 目錄（`O_DIRECTORY|O_NOFOLLOW`，逐段）後對葉節點以 `*at` 系統呼叫操作」進行，經 `golang.org/x/sys/unix` 的 `Openat`/`Fchmodat`（`AT_SYMLINK_NOFOLLOW`）/`Unlinkat` 完成。`ensureLogAccess` 需建立缺少的中間目錄，改以 `Mkdirat` 逐段建立（相對於已驗證的 parent fd），取代會跟隨 symlink 的 `os.MkdirAll`，否則建目錄步驟會重新引入中間 symlink 逃逸。逐段 `O_NOFOLLOW` 可同時封住中間與末端元件的 symlink，且避免「先 `EvalSymlinks` 再開啟」的 TOCTOU 窗口。並修正 `ensureLogAccess`/`truncateLog` 中誤稱 `O_NOFOLLOW` 已關閉 symlink 替換競態的註解。

替代方案：對 parent 做 `filepath.EvalSymlinks` 後重新確認仍在允許清單內——已否決為主要方案（解析與開啟之間有 TOCTOU 窗口），僅在 `*at` 難以套用處作為次選。移除 `/tmp/`、`/private/tmp/` 允許清單項——已否決，會使把 log 放在 `/tmp` 的系統 daemon 失去 log 存取/清除功能（可被使用者察覺的功能倒退）。

### 路由 name/Label 以共用 validateRoutingName 限制為單一元件（含系統域）

新增共用 `validateRoutingName`：拒絕含 `/`、NUL，或**等於** `.`/`..` 的值（單一元件即可通過）。於使用者域（`user.go`）、**系統域（`system.go` 的 Create 用 `config.Label`，Update/Delete/Start/Stop/Restart 用 `name`）**、唯讀域（`readonly.go`）各接受 `name`/`Label` 的方法起始處呼叫。只需拒絕分隔符／NUL 與恰為 `.`/`..` 的值即可完全 confine：無分隔符時 `filepath.Join(base, name+".plist")` 不可能逃出 base。**不**以子字串方式拒絕 `..`——合法 launchd label 可含連續點（如 `com.example..worker`），且 `list()` 仍會列出它，若拒絕會使該服務可見卻無法管理。系統域寫入雖由 helper 端 `validateSystemDaemonPath` 再驗證，但 GUI 端讀取與操作仍應自我 confine（縱深防禦），不倚賴 helper 單點。

### 系統域排程範圍驗證對稱，50 筆上限置於系統域 create/update 路徑

`SystemManager.Create/Update` 補呼叫既有 `validateSchedule`（`StartInterval >= 10`；行事曆 minute 0-59／hour 0-23／day 1-31／weekday 0-6／month 1-12）。行事曆數量上限（50，與 cron-range-expansion 一致）於**系統域 create/update 路徑**（可回傳錯誤處，檢查 `len(config.Schedule.Schedules) > 50`）強制，回傳驗證錯誤且不寫 plist。

替代方案：把上限放進 `buildCalendarInterval`/`BuildPlistDict`——已否決：該層回傳 `any`/`map[string]any` 無錯誤通道，只能靜默截斷或 panic，無法達成「回傳錯誤、不寫 plist」；且該 encoder 為使用者域共用，會對使用者域造成未揭露的迴歸（使用者域上限維持前端現狀，本 change 不變更）。

### helper 以絕對路徑 /bin/launchctl 呼叫

bootstrap/bootout/kickstart 改以 `/bin/launchctl` 呼叫，讓安全性顯式，不依賴提權 trampoline 的 PATH 重置。以單一常數表示路徑。

## Implementation Contract

**行為（可觀察）：**

- 對 helper log 路徑 RPC 傳入「中間目錄為 symlink 指向允許清單外」的路徑時，root 不會 chmod/截斷/刪除/建檔於允許清單外的真實目標（操作被拒或被限制在允許清單內的真實路徑）；合法的（非 symlink、位於允許清單子目錄下的）log 路徑仍可運作，含建立中間目錄。
- 以含 `/`、`..`、NUL 的 `name`/`Label` 呼叫任何綁定（使用者域或系統域）時回傳驗證錯誤且不觸及 base 目錄外的檔案；一般 label 照常運作。
- 以超出範圍或超過 50 筆的排程建立/更新系統 daemon 時回傳驗證錯誤且不寫入 plist；合法排程照常；使用者域排程行為不變。
- helper 執行 bootstrap/bootout/kickstart 時呼叫 `/bin/launchctl`，與 `$PATH` 無關。

**介面/契約：**

- `internal/privhelper/handlers.go`：`ensureLogAccess`/`truncateLog`/`deleteOneLogPath`/`backupExisting` 的檔案操作改用逐段 `O_NOFOLLOW`/`*at`（含 `Mkdirat` 建中間目錄）；`validateLogPath` 字面檢查保留為前置；bootstrap/bootout/kickstart 的 launchctl 呼叫改絕對路徑常數。RPC 方法集、參數形狀不變。
- `internal/launchctl`：新增共用 `validateRoutingName`，於 `user.go`、`system.go`（Create/Update/Delete/Start/Stop/Restart）、`readonly.go` 接受 `name`/`Label` 的方法起始呼叫；`system.go` Create/Update 加 `validateSchedule` 與 `len(Schedules) > 50` 檢查。
- 對外 Wails 綁定簽章不變。不修改 `types.go`、`internal/backup/backup.go`、`plist_encode.go`、前端。

**失敗模式：**

- 非 darwin 平台：`*at` 逐段解析以既有平台守衛處理；本產品僅 macOS 發佈。
- `*at` 逐段開啟遇任一元件為 symlink：回傳錯誤，操作中止（不部分完成）。

**驗收標準：**

- 新增/更新 Go 測試涵蓋：中間目錄 symlink（含建目錄情境）無法使 helper 於允許清單外 chmod/截斷/刪除/建檔；`validateRoutingName` 拒絕 `..`/`/`/NUL 且接受一般 label，且系統域 Create/Update/Delete 亦受保護；系統域超範圍排程與超過 50 筆被 create/update 路徑以錯誤拒絕且不寫 plist；helper 以 `/bin/launchctl` 呼叫（以注入的 Runner 斷言）。
- `make test` 綠燈；`make lint` 無新違規。

**範圍邊界：**

- 僅本 change 列出的防禦性修正與其測試。
- 不動 Admin Mode 生命週期、helper 簽章、log 讀取/備份還原的 `O_NOFOLLOW`、使用者域排程行為、log 載入前端狀態機、或 WritePlist/Bootstrap 語意。

## Risks / Trade-offs

- [逐段 `*at` 解析改動 helper 檔案操作路徑] → 以測試涵蓋合法路徑（含建目錄）仍運作與中間 symlink 被封；`validateLogPath` 字面前置檢查保留。
- [name/Label 驗證可能擋掉既有含特殊字元的服務名] → 合法 launchd label 僅用 `[A-Za-z0-9._-]`，不含 `/`/`..`；驗證只擋穿越字元，不影響正常命名。
- [系統域排程驗證新增可能拒絕過去可寫入的畸形/超量排程] → 與使用者域範圍檢查一致，launchd 本就忽略畸形值；上限 50 與前端一致；屬正確化，且不影響使用者域。
- [絕對路徑 `/bin/launchctl` 若未來 macOS 變更 launchctl 位置] → 為既有標準位置；風險極低，且比 bare name 依賴 `$PATH` 更安全。

## Migration Plan

無資料遷移。純執行期行為變更，隨版本更新即生效。回退策略：還原本 change 的程式碼變更即可（無持久化狀態）。

## Open Questions

無。
