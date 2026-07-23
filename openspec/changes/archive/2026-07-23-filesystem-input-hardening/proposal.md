## Why

安全稽核（run-1）確認了一組低風險但真實的檔案系統/輸入強化缺口：root helper 的 log 路徑處理可經「中間目錄 symlink」逃逸允許清單（其 `O_NOFOLLOW`/`Lstat` 只保護末端元件，且註解誤稱已關閉 symlink 替換競態）；GUI 的 `name` 路由參數無路徑穿越防護；系統 daemon 的建立/更新略過排程驗證；root helper 以 bare name 呼叫 `launchctl`（目前被提權 trampoline 的 PATH 重置緩解，屬潛在向量）。這些多為縱深防禦（例如 symlink 逃逸對能觸達 socket 的呼叫者而言，與 by-design 的 WritePlist+Bootstrap root RCE 完全冗餘，非提權），但值得一次收斂以維持一致的防護面與誠實的程式碼註解。（稽核 #6 的 `readLogTail`/`copyFile` 補 `O_NOFOLLOW` 經評估後不納入：屬使用者權限內、安全效益極微，且會使「symlink 但目標尚未建立」的 log 從中性 not-found 變成錯誤、並破壞把 plist 以 symlink 指向 dotfiles 的還原工作流——淨負面。）

## What Changes

- **helper log 路徑逐元件 symlink 安全解析**：`validateLogPath` 目前僅 `filepath.Clean` + 前綴比對（允許清單含世界可寫的 `/tmp/`、`/private/tmp/`），而 `ensureLogAccess`（`os.MkdirAll` + `os.Chmod` parent）、`truncateLog`、`deleteOneLogPath`、`backupExisting` 的 `O_NOFOLLOW`/`Lstat` 只保護末端元件。改為逐段以 `O_NOFOLLOW` 開啟（parent 以 `O_DIRECTORY|O_NOFOLLOW`，葉節點以 `*at`；建立缺少的中間目錄以 `Mkdirat`）再進行 chmod/open/remove。修正誤導性註解。
- **`name` 路由參數穿越防護**：GUI 的 `name`（與 Create 的 `config.Label`）流入 `filepath.Join(basePath, name+".plist")` 且無 confine，`name="../../etc/passwd"` 會解析到 base 之外。改為拒絕含 `/`、`..`、NUL 的名稱，要求 `name == filepath.Base(name)` 且非 `.`/`..`；使用者域、系統域（含 Create/Update/Delete/Start/Stop/Restart）與唯讀域皆套用。
- **系統 daemon 排程驗證對稱**：`SystemManager.Create/Update` 未呼叫 `validateSchedule`（使用者域有）。補上排程範圍驗證，並在**系統域 create/update 路徑**（可回傳錯誤處）強制行事曆展開數量上限（50，與 cron 展開一致）——不放進無錯誤通道且為使用者域共用的 encoder，以免誤改使用者域行為。
- **helper 以絕對路徑呼叫 `launchctl`**：bootstrap/bootout/kickstart 目前以 bare name 經 `$PATH` 解析。改為 `/bin/launchctl`，讓安全性顯式而非仰賴提權 trampoline 的 PATH 重置。

## Capabilities

### New Capabilities

(none)

### Modified Capabilities

- `privileged-helper-rpc`: 新增需求——log 路徑相關 RPC（EnsureLogAccess/TruncateLog/DeleteLogPaths）與備份寫入 SHALL 對每一路徑元件做 symlink 安全解析，不得只保護末端元件。
- `system-daemon-write-ops`: 新增需求——helper SHALL 以絕對路徑 `/bin/launchctl` 呼叫；系統 daemon 的建立/更新 SHALL 施加與使用者域相同的排程範圍驗證，並於 create/update 路徑（可回傳錯誤處）強制 50 筆行事曆上限。
- `core-service-management`: 新增需求——路由 `name`／`Label` SHALL 於使用者域、系統域與唯讀域限制為單一路徑元件（拒絕穿越）。

## Impact

- Affected specs: `privileged-helper-rpc`, `system-daemon-write-ops`, `core-service-management`
- Affected code:
  - Modified:
    - internal/privhelper/handlers.go
    - internal/launchctl/user.go
    - internal/launchctl/readonly.go
    - internal/launchctl/types.go
    - internal/launchctl/system.go
  - New: (none)
  - Removed: (none)
- Tests affected: internal/privhelper/handlers_test.go, internal/launchctl/user_test.go, internal/launchctl/system_test.go
- Behavior change surfaced to users: minimal — invalid traversal names and out-of-range / over-cap system-daemon schedules are now rejected with an error (consistent with existing user-domain behavior). User-domain schedule behavior is unchanged (its 50-entry cap remains as-is). Normal (non-traversal, in-range) operations are unchanged.
