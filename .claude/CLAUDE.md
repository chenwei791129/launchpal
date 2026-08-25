# LaunchPal

A GUI for managing macOS LaunchAgents.

## UI Language

**Every user-facing string in the frontend (tooltips, labels, buttons, descriptions, alerts, error messages) MUST be written in English.** Even when the discussion, spec, or task notes are in Chinese, translate the text before putting it into a Vue template or TypeScript string literal.

- **Why**: The rest of the UI ("System Services", "Read-only", "Stop service", etc.) is in English. Mixing in Chinese strings breaks consistency, and because this project has no i18n framework yet, hard-coding Chinese permanently locks the wording to that language.
- **When it applies**:
  - Adding text in a `.vue` template or a string literal in `<script setup>` (tooltip, `alert()`, `throw new Error()`, `label`, etc.).
  - While editing existing strings: if you notice Chinese, translate it opportunistically.
  - When implementing UI from a spec or tasks file written in Chinese: translate the wording, do not paste it verbatim.
- **Exception**: Chinese is acceptable only once an i18n framework is introduced; English must remain the default locale.

## Tech Stack

- **Backend**: Go + Wails v2
- **Frontend**: Nuxt 4 + Vue 3 + TailwindCSS
- **Platform**: macOS only (`launchctl` is macOS-specific)

## Directory Layout

```
├── app.go                    # Wails bindings exposed to the frontend
├── admin_mode.go             # Admin Mode state machine + helper lifecycle (Enable/Disable/status)
├── main.go                   # Application entry point
├── Makefile                  # Build recipes (includes build-helper / install-helper for privhelper)
├── cmd/
│   └── launchpal-privhelper/ # Root-privileged helper binary entry point (session-scoped, no persistence)
├── internal/
│   ├── launchctl/            # launchctl command wrappers
│   │   ├── types.go          # Service, ServiceConfig, and related types (includes StatusConfidence)
│   │   ├── manager.go        # Manager interface
│   │   ├── user.go           # UserManager (~/Library/LaunchAgents)
│   │   ├── system.go         # SystemManager — dual mode: read-only by default, Admin Mode delegates writes via AdminClient
│   │   ├── apple_system.go   # AppleSystemManager (/System/Library/LaunchDaemons, always read-only)
│   │   ├── readonly.go       # Shared read-only logic for SystemManager and AppleSystemManager
│   │   └── status_detect.go  # Heuristic status detection for the system domain (single ps scan + uid/ppid=1 filter)
│   ├── privhelper/           # RPC protocol + server + client shared by app.go and the helper binary
│   │   ├── protocol.go       # Request/Response types, method name + error code constants
│   │   ├── server.go         # Newline-delimited JSON server, peer UID verification, idle/parent watchdog
│   │   ├── client.go         # Client side: Connect (retry), LaunchHelper (osascript), typed method wrappers
│   │   ├── handlers.go       # Bootstrap/Bootout/Kickstart/WritePlist/DeletePlist/List handlers, path + label validation
│   │   ├── integrity.go      # ProtectedHelperPath constant, FileSHA256, protected-copy ownership/permission verification
│   │   ├── install.go        # Root-owned protected-copy install from the running image (O_NOFOLLOW, idempotent)
│   │   ├── logpath_darwin.go # Per-component O_NOFOLLOW/openat resolution for log-path ops + backup mkdir chain
│   │   ├── logpath_other.go  # Non-darwin fallback (os.*; per-component symlink safety enforced only on darwin)
│   │   └── peer_darwin.go    # LOCAL_PEERCRED implementation (via golang.org/x/sys/unix)
│   ├── backup/               # Backup management
│   │   └── backup.go         # BackupManager implementation
│   ├── plistutil/            # plist format detection and binary→XML normalization (shared by backup, launchctl)
│   │   └── plistutil.go      # DetectFormat, NormalizeFromPath
│   └── settings/             # User preferences persisted as JSON
│       └── settings.go       # Settings struct, Default/Load/Save/Validate (~/.launchpal/settings.json)
├── frontend/                 # Nuxt 4 frontend project
│   ├── app/
│   │   ├── app.vue           # Root component
│   │   ├── pages/            # Pages (index, system, apple-system, settings, services/[name])
│   │   ├── components/       # Vue components
│   │   ├── composables/      # Composables
│   │   ├── layouts/          # Layouts
│   │   ├── assets/           # Static assets
│   │   ├── utils/            # Utility helpers
│   │   └── types/            # TypeScript types
│   └── nuxt.config.ts
├── build/
│   └── darwin/               # macOS build configuration
└── wails.json                # Wails project configuration
```

## Development Commands

```bash
make setup       # Install dependencies
make test        # Run tests (Go tests + frontend vitest + TypeScript typecheck)
make lint        # Run linters (Go golangci-lint + frontend eslint)
make build       # Build the production app
make build-debug # Build with devtools enabled
make dev         # Build and launch the app
make dmg         # Build and package as DMG
make clean       # Remove build artifacts
```

## Backup Mechanism

- Backup directory: `~/.launchpal/backups/<service-name>/`
- Backup files:
  - `<timestamp>.plist` — plist backup
  - `<timestamp>.meta.json` — metadata (original path, etc.)
- Automatic backups are taken before `Update` and `Delete`.
- Retention: the 10 most recent backups.
- Settings → Backup History exposes a Diff button on each entry that opens a **side-by-side diff preview** (current on the left, backup on the right, red/green coloring) before the user decides whether to Restore. Binary plists are auto-converted to XML by the backend; if the service has been deleted, the left column renders as placeholders and the right column appears as pure additions. The diff is capped at 10,000 lines, and a truncation notice is shown beyond that.

## Settings

- Settings file: `~/.launchpal/settings.json` (atomic write via `os.Rename`, missing/corrupt files fall back to `Default()` so first-run never errors).
- Wails bindings: `GetSettings()` and `UpdateSettings(s)` on `App`. Validation runs in front-end (`utils/settingsValidation.ts`) and back-end (`internal/settings.Validate`) — back-end is the source of truth.
- Settings page exposes a **Log Storage** section (immediately after Backup Storage) with two independent controls — User Log Directory (default `~/Library/Logs`) and System Log Directory (default `/Library/Logs`). Each has Save and Reset to Default buttons. Save persists via `UpdateSettings`; inline error appears on validation failure; success indicator flashes briefly on success.
- New Service modal sources the default `stdoutPath` / `stderrPath` from these settings via `composeLogPaths` (`<dir>/<label>/<stream>.log`). Settings are re-read every time the modal opens (Decision 8); existing services are not migrated when settings change (Decision 9).
- The system-log path allowlist (`/var/log/`, `/private/var/log/`, `/Library/Logs/`, `/tmp/`, `/private/tmp/`) is the single source of truth at `internal/privhelper.SystemLogPathPrefixes`. Both the privileged helper (`EnsureLogAccess`/`TruncateLog`) and the settings validator import from here — never duplicate the slice.

## Service Types

LaunchPal supports three service categories:

1. **User Services** (`~/Library/LaunchAgents`)
   - Full read/write access.
   - Can start, stop, create, update, and delete services.
   - Supports immediate execution (Kickstart: `launchctl kickstart -k`).
   - Supports scheduling (`StartCalendarInterval` / `StartInterval`).
   - Cron syntax accepts ranges (`a-b`) and enumerations (`a,b,c`) and is expanded into the Cartesian product of `StartCalendarInterval` entries (capped at 50 entries).
   - Supports environment variables (`EnvironmentVariables`).
   - **Launch policy**: the create form (`CreateServiceModal.vue`) and the edit form (`pages/services/[name].vue`) present launch behavior as a single mutually-exclusive "Launch Policy" radio group — `On Demand` / `Run at Load` / `Keep Alive` — instead of two checkboxes. Mapping (shared via `app/utils/launchPolicy.ts`, rendered by `app/components/LaunchPolicyForm.vue`): `On Demand` writes neither key; `Run at Load` writes `RunAtLoad=true`; `Keep Alive` writes a `KeepAlive` value and **no** standalone `RunAtLoad` (launchd implies it). On load, KeepAlive takes precedence, so a legacy plist with both `RunAtLoad` and `KeepAlive` lands on `Keep Alive` and drops the standalone `RunAtLoad` on next save.
   - **KeepAlive fidelity**: `KeepAlive` is modeled as a structured `KeepAliveConfig` (backend `internal/launchctl/keepalive.go`), not a flattened bool. Both the boolean and dictionary plist forms round-trip without loss. The advanced section under `Keep Alive` exposes editable `SuccessfulExit` / `Crashed` / `AfterInitialDemand` (tri-state) and an integer `ThrottleInterval`; the non-editable `NetworkState` (deprecated by launchd), `PathState`, and `OtherJobEnabled` sub-keys are preserved verbatim on read and written back unchanged. A dictionary with no effective sub-key is written as the boolean `KeepAlive=true` (never an empty dict).
   - Logs tab supports a one-click Clear Logs control: truncates the active stdout/stderr file in place via `O_WRONLY|O_TRUNC|O_NOFOLLOW` (preserves inode and mode, rejects symlinks).
   - Supports cloning via a Copy button on the detail page header (next to Run Now). Clicking Copy opens the existing CreateServiceModal pre-filled with the source service's configuration (including the structured `KeepAlive` and `ThrottleInterval`); a `Keep Alive` source keeps the `Keep Alive` policy, any other source defaults to `On Demand` (so a clone never inherits a standalone `RunAtLoad`; the user can re-select before submit), and the new `Label` must be entered by the user. Duplicate labels are rejected inline by the backend's `service <label> already exists` error. The Copy button is hidden on System Services and Apple System Services detail pages.

2. **System Services** (`/Library/LaunchDaemons`)
   - Read-only by default. Write access (Start/Stop/Restart/Create/Update/Delete) requires Admin Mode to be enabled (see below).
   - Can view service information, status, plist contents, and logs without Admin Mode.
   - Third-party system-level services.
   - Logs tab Clear Logs control dispatches per-file: a direct `OpenFile` is tried first; only `EACCES` falls back to the helper's `TruncateLog` RPC. `ENOENT`, `ELOOP`, `EISDIR` surface to the caller and never escalate. Homebrew-style daemons whose log files are user-writable can be cleared without enabling Admin Mode.
   - Delete confirmation dialog includes an "Also delete log files" checkbox (default off). When checked, after the plist is removed the helper's `DeleteLogPaths` RPC cleans up the daemon's `StandardOutPath` / `StandardErrorPath` (and empties the parent directory if it becomes empty). Per-path failures surface as a `*launchctl.LogDeletionWarning` typed error — the overall delete still counts as success.

3. **Apple System Services** (`/System/Library/LaunchDaemons`)
   - Always read-only, even with Admin Mode enabled (SIP would block writes anyway).
   - Can view service information, status, plist contents, and logs.
   - macOS built-in services.
   - Many of these use the binary plist format and are automatically converted to XML for display.
   - The Clear Logs control is hidden entirely on apple-system pages.

## Routing Name & Schedule Input Hardening

- **Routing name confinement**: `validateRoutingName` (`internal/launchctl/user.go`, the single source of truth shared across domains) rejects any service routing `name` — and `CreateService`'s `config.Label` — that is not a single path component (contains a path separator or NUL, or equals `.` / `..`). A single component that merely contains `..` as a substring (e.g. `com.example..worker`) is deliberately **accepted** — with no separator it cannot escape the base dir, and it is a legal launchd label that `list()` surfaces and must stay manageable. Because `filepath.Join(base, name+".plist")` does not confine `..`, this guard runs at the start of every name-accepting entrypoint: user domain (`Get`/`Start`/`Update`/`Delete`/`Kickstart`/`GetPlist`/`GetPlistContent`, plus `Create` on `Label`), the read-only domain (`readonly.go` `get`/`getPlist`/`getPlistContent`, so `SystemManager`/`AppleSystemManager` reads are covered), and the system domain write ops (`system.go` `Create`/`Update`/`Delete`/`Start`/`Stop`/`Restart`, after the Admin-Mode client check so a traversal name performs no file op or helper RPC). `list()` does not route through the guard — it iterates trusted `os.ReadDir` entries. This is GUI-side defense in depth; the helper still independently re-validates system-domain write paths.
- **System-domain schedule validation parity**: `SystemManager.Create`/`Update` call `validateSystemSchedule` (`system.go`) — the shared range check (`validateSchedule`: `StartInterval >= 10`; calendar minute 0-59 / hour 0-23 / day 1-31 / weekday 0-6 / month 1-12) plus a system-domain-only 50-entry cap (`maxSystemCalendarEntries`, matching the frontend cron range-expansion limit `ScheduleForm.vue MAX_EXPANSION`). Enforced in the create/update path (returns an error and writes no plist / issues no RPC on failure), **not** in the error-less `buildCalendarInterval`/`BuildPlistDict` encoder, which is shared with the user domain — user-domain schedule behavior (its 50-entry cap stays a frontend concern) is unchanged.

## Admin Mode (session-scoped privileged helper)

Because LaunchPal is unsigned and cannot register a persistent `SMAppService` daemon, write access to `/Library/LaunchDaemons` is gated by a **session-scoped helper** the user enables explicitly.

- **Binary**: `launchpal-privhelper`, shipped inside the app bundle at `LaunchPal.app/Contents/MacOS/launchpal-privhelper`. Built by `make build-helper`, installed into the bundle by `make install-helper`. The `make build` / `build-debug` order is load-bearing: build-helper → hash it → `wails build -ldflags "-X main.helperPin=<sha256>"` (pin injected **before** the main binary is linked) → copy the same helper into the bundle. CI (`.github/workflows/build.yml` "Build application" step) mirrors this, appending `-X main.helperPin=<sha256>` alongside the existing `-X main.version` injection.
- **Launch**: triggered by the `EnableAdminMode` Wails binding. `resolveHelperPath` (GUI, unprivileged; `admin_mode.go`) decides which binary to launch, then LaunchPal runs `osascript -e 'do shell script "... with administrator privileges"'`, prompting the user for their password or Touch ID once per session.
- **Launch integrity** (`internal/privhelper/integrity.go` + `install.go`, `admin_mode.go`): defends the unsigned helper against plant-and-wait tampering.
  - **Trust anchor = root-owned protected copy** at `internal/privhelper.ProtectedHelperPath` (`/Library/Application Support/LaunchPal/launchpal-privhelper`, the single source of truth shared by GUI resolution and helper install). On first enable the helper self-installs a copy of its own running image there (source opened via `os.Executable()` with `O_NOFOLLOW`, not a re-read of the bundle, to shrink the TOCTOU window), owned `root:wheel` mode `0755` with a `root:wheel 0755` parent. A copy is verified only when it is a regular file (not a symlink), owner UID 0, and `mode & 022 == 0` (`IsVerifiedProtectedCopy`).
  - **Resolution priority** (`resolveHelperLaunchPath`, signature kept `func() (string, error)`): a verified protected copy is launched by default; the **only** exception is a legitimate update — non-empty pin, readable bundle, `bundleHash == pin`, and `bundleHash != protectedHash` — which launches the bundle copy to re-provision. Key invariant: an existing verified protected copy is **never** bypassed by an empty pin, a missing/unreadable bundle, or a tampered bundle (closes the "clear the pin to force a malicious bundle" downgrade and the "delete the bundle to DoS Admin Mode" failure). With no verified protected copy, the bundle copy launches after hash-pin verification (or, with an empty dev-build pin, without it).
  - **Hash pinning = defense-in-depth, not the anchor**: `main.helperPin` is the bundle helper's SHA-256, injected at build time; empty in dev builds. It gates only the launch of a bundle copy — it is never a precondition for launching the protected copy. Honestly bounded: the pin lives in the user-writable main binary, so an attacker who can write the helper can usually patch the pin — hence it only blocks unsophisticated overwrites during the bootstrap window.
  - **Integrity failure**: when no verified protected copy exists and hash-pin verification fails (bundle missing/unreadable, or SHA-256 ≠ non-empty pin), `Enable` does **not** run osascript or launch the helper; it returns to `Disabled` via `failFromRequesting` with code `helper_integrity_failed` (message carries only the code). Helper self-install failure is non-fatal — the current session still serves from the launched binary; the protected copy is retried on the next enable.
  - **Residual risk**: the first enable (and the first enable after each app update) still runs the bundle copy once — this bootstrap window, including the self-install TOCTOU, cannot be fully closed without a paid Apple Developer signing identity.
- **IPC**: Unix domain socket at `$TMPDIR/launchpal-<uid>-<16-hex-random>.sock`, `chmod 0600`, peer UID verified via `LOCAL_PEERCRED` before any RPC is processed. Messages are newline-delimited JSON.
- **Scope**: the helper can only Bootstrap/Bootout/Kickstart and WritePlist/DeletePlist under `/Library/LaunchDaemons/`, plus EnsureLogAccess/`TruncateLog`/`DeleteLogPaths` under the log allowlist (`/var/log/`, `/private/var/log/`, `/Library/Logs/`, `/tmp/`, `/private/tmp/`, sub-directory required). Bootstrap/Bootout/Kickstart invoke launchctl by its **absolute path `/bin/launchctl`** (single constant `launchctlPath`), so the resolved binary is independent of the inherited `$PATH`. Paths outside these prefixes and labels with shell metacharacters are rejected before any `launchctl` invocation.
- **Symlink-safe log-path resolution** (`logpath_darwin.go`): because the allowlist includes world-writable `/tmp/` and `/private/tmp/`, a same-UID process can plant a symlink at an **intermediate** directory, not just the leaf. `validateLogPath` (lexical `filepath.Clean` + prefix match) is a fast pre-filter, **not** the enforcement boundary — it now also returns the matched allowlist prefix. `EnsureLogAccess`/`TruncateLog`/`DeleteLogPaths` (and the backup mkdir chain in `backupExisting`) resolve **every path component** below the trusted prefix with `O_NOFOLLOW`: the parent chain is walked via `openat(O_DIRECTORY|O_NOFOLLOW)` anchored at the trusted (root-owned or well-known) prefix, and the leaf is operated on relative to that parent fd (`Fchmod`/`Openat`/`Fstatat`+`AT_SYMLINK_NOFOLLOW`/`Unlinkat`). Missing intermediate directories are created with `Mkdirat` per component (replacing the old symlink-following `os.MkdirAll`), so directory creation cannot reintroduce the escape. A symlink at any component fails the operation instead of redirecting the root-privileged chmod/create/truncate/delete outside the allowlist. `TruncateLog` still opens `O_WRONLY|O_TRUNC|O_NOFOLLOW` without `O_CREATE`, so it cannot be coerced into materializing a 0-byte root-owned file. `DeleteLogPaths` refuses symlinks/non-regular files and best-effort removes the now-empty parent (`ENOTEMPTY` ignored); per-path failures return as a partial-success warning slice, not an RPC-level error. The non-darwin `logpath_other.go` preserves the pre-openat `os.*` behavior (LaunchPal ships only on macOS; the per-component guarantee is darwin-only and the tests `t.Skip` when `syscallNoFollow == 0`).
- **Log access**: after a successful Create/Update, `SystemManager` sends an `EnsureLogAccess` RPC with the plist's `StandardOutPath` / `StandardErrorPath`. The helper creates the parent chain symlink-safely (see above), tightens the leaf's parent to `0755` if it already existed with a stricter mode (common launchd default is `0744`, which blocks user traversal), and touches the log file as `0644` with `O_NOFOLLOW` if missing. Paths are restricted to the allowlist and must live at least one sub-directory deep — `/var/log/foo.log` is rejected to prevent re-moding a system log root.
- **Lifecycle**:
  - **Disconnect = self-terminate (primary teardown)**: the helper serves a single client. `acceptLoop` triggers `Server.Stop()` after the **primary** connection's `handleConn` returns for **any** reason — clean EOF, a read/scan error, or a failed response write (`encoder.Encode`) — not only the post-scan EOF path. The primary connection is the first one accepted (the GUI connects immediately after launch and holds one long-lived connection); a stray same-UID connection opening and closing does **not** tear down the live session (`primaryConn` guard). This is the main mechanism that closes the "root socket outlives the authorized session" window, because the unprivileged GUI cannot signal the root helper. It only fires on Disable, GUI exit, GUI crash, or a real disconnect; a normal multi-step operation keeps the single long-lived connection open. `Stop()` is skipped when the server is already stopping for another reason. The explicit `Shutdown` RPC stops the server synchronously regardless of which connection sent it.
  - **`Stop()` closes active connections**: `Server.Stop()` closes the listener **and** every accepted connection still being served (tracked in a `conns` set). Closing the connection unblocks `handleConn`'s `scanner.Scan()` so the process actually exits; closing the listener alone would leave a connected-but-idle helper running. This is what lets an idle-timeout or parent-watchdog stop terminate a helper the GUI is still connected to.
  - Parent watchdog polls LaunchPal's PID once per second and treats the parent as alive **only when a process with that PID exists AND its recorded start time still matches** — a PID that exists with a different start time (PID reuse) counts as gone, so the helper self-exits within ~2 seconds. Start time is read on darwin via `kinfo_proc` (`SysctlKinfoProc` "kern.proc.pid", `P_starttime`) in `procinfo_darwin.go`; `procinfo_other.go` returns an error off darwin so the closure degrades to a plain `syscall.Kill(pid, 0)` existence check. `StartParentWatchdog` and its `alive func(int) bool` signature are **unchanged** — the recorded start time is captured into the `alive` closure via `makeParentAlive`.
  - Idle timeout: **5 minutes** without RPC traffic triggers self-exit (single adjustable constant `idleTimeout` in `cmd/launchpal-privhelper/main.go`; pinned by `TestIdleTimeout_DefaultIsFiveMinutes`). Every RPC resets the timer, so an active session is unaffected; the timeout only bounds an idle-but-still-connected session. The idle stop closes the active connection (see above) so the process — not just the listener — exits.
  - Graceful shutdown via the `Shutdown` RPC (sent by `DisableAdminMode`) with a 3-second ack timeout.
  - **Best-effort Shutdown on handshake failure (client side)**: `privhelper.LaunchHelper` sends a best-effort `Shutdown` (short timeout) over the established connection before `Close()` when `Connect` succeeded but the `Ping` handshake failed — `admin_mode.go` receives `(nil, err)` on that path and has no client to send on, so the Shutdown must live in `LaunchHelper`. When `Connect` itself failed (no connection), no Shutdown is sent and the idle timeout + parent watchdog are the backstop. (`Client.Close()` routes its `closeCh` close through a `sync.Once` so the Close-vs-readLoop race cannot double-close.)
- **Backups**: when the helper overwrites/deletes a plist it first writes a backup under `<user-home>/.launchpal/backups/<label>/` and `lchown`s the created files to the launching user's UID/GID so the user-side LaunchPal can read them. The backup directory chain is created per component with `O_NOFOLLOW` (`symlinkSafeMkdirChain`, anchored at the launching user's home) so a symlink pre-planted at `.launchpal` / `backups` / `<label>` cannot redirect the root-privileged write; the leaf backup/meta writes use `O_NOFOLLOW` (`writeNoFollow`) so the leaf itself is likewise not followed.
- **Enable commit is race-safe**: `Enable`'s success path resolves the commit under `a.mu` via a switch — if `disableRequested` it tears down and lands on `Disabled` (""); if `helperDisconnected` (the helper connection ended during the Requesting window, recorded by `handleHelperCrash` when it sees state `Requesting`) it tears down and lands on `Disabled` (`admin_session_ended`) rather than storing a dead client as `Enabled`; otherwise it commits `Enabled`. Both early exits tear the client down **outside** the lock (via `setStateLocked` for the final transition). Because `handleHelperCrash` and the commit both hold `a.mu`, a disconnect either sets the flag before the commit (caught by the switch) or arrives after `Enabled` is committed (caught by the `Enabled` branch) — the window is fully closed.
- **Pending-disable during Requesting**: a Disable click while the authorization prompt is in flight (state `Requesting`) is **not** dropped — `Disable` records a `disableRequested` flag (guarded by `a.mu`) instead of no-op'ing. When the in-flight Enable's handshake then succeeds, it snapshots the client + flag under the lock, **releases the lock**, and — if the flag is set — tears the just-launched helper down via the shared `teardownClient` helper and lands on `Disabled` rather than `Enabled` (tearing down outside the lock avoids freezing `GetAdminModeStatus` behind the blocking Shutdown RPC). The flag is cleared **only** at the start of a fresh `Disabled` → `Requesting` Enable; a no-op Enable (already `Requesting`/`Enabled`) does **not** clear it, so multi-click sequences are deterministic. The teardown sequence (short-timeout `Shutdown` → `Close` → clear client/`ClearAdminClient`) is centralized in `teardownClient`, shared by `Disable` and the pending-disable path.
- **State machine (frontend-visible)**:
  ```
  Disabled → Requesting → Enabled     (successful authorization + handshake, no pending disable)
                        ↓
                     Disabled          (user cancel / helper handshake failure / pending-disable honored)
  Enabled → ShuttingDown → Disabled   (DisableAdminMode)
  Enabled → Disabled                   (helper connection ends while Enabled — reason: "admin_session_ended")
  ```
  A connection ending while `Enabled` (idle self-termination, clean teardown, or a real crash — the GUI cannot tell them apart) surfaces the **neutral** `admin_session_ended` reason, not a red `helper_crashed` error. State changes emit the `admin_mode:state` Wails event. The frontend composable `useAdminMode()` exposes `state`, `lastError`, `isEnabled`, `isRequesting`, `isShuttingDown`, `isSessionEnded`, `displayMessage`, and `enable()`/`disable()`; when `isSessionEnded` is true the Settings page renders `displayMessage` ("Admin Mode session ended — re-enable to continue") in neutral gray instead of red.

## Status Detection Logic

### User domain (`UserManager`)

Status and PID derive **exclusively** from what launchd reports for the service label, so `List` (batch) and `Get` (single-service) classify an unchanged job identically.

1. `List` runs `launchctl list` once (`getBatchServiceStatus`, tabular output) and reuses the resulting label → status/PID map for every service; `Get` runs `launchctl list <label>` (`getServiceStatus`, dict output) for the one label.
2. Both parsers hand their parsed PID to the shared `classifyPID` helper — the single source of truth for the running/loaded boundary, so the two output formats cannot drift back into disagreeing about the same job.
3. Classification: positive PID ⇒ `StatusRunning` with that PID; loaded without a positive PID ⇒ `StatusLoaded` / PID 0; label unknown to launchd ⇒ `StatusStopped` / PID 0.
4. An empty `Label` is classified as `StatusUnknown` / PID 0 / `unverified` in `getWithStatus` — the single chokepoint covering both paths (the batch path never reaches `getServiceStatus`), and launchd is never queried for an empty label. `unverified` mirrors the system domain's `StatusUnknown`/`unverified` pairing in `status_detect.go` — nothing was actually confirmed with launchd, so labeling it `verified` would be a lie the frontend's confidence tooltip is supposed to catch.
5. There is deliberately **no** `pgrep -f <program>` fallback. It attributed unrelated processes to a job whenever the program was a short wrapper command (e.g. `open` matches any command line containing that substring), so the detail view showed a fake PID and `running` while the list view correctly showed `loaded`. `TestUserServiceStatusConsistency` pins this: it shims `pgrep` to return a fake PID and fails if the fallback is reintroduced. (`commonShells` still exists, but is now used only by the system-domain heuristic in `status_detect.go`.)
6. `StatusConfidence` is `verified` for every other case, because `launchctl list` is authoritative in the user domain once a label exists to query.

### System domain (`SystemManager` / `AppleSystemManager`)

`launchctl list` in a user context only surfaces services in the `gui/<uid>` domain; it cannot see system daemons under `/Library/LaunchDaemons` or `/System/Library/LaunchDaemons`. `status_detect.go` performs a heuristic detection instead:

1. Read `UserName` from the plist (defaults to `root`) and resolve `program` (`Program` if present, otherwise `ProgramArguments[0]`).
2. `program` empty → `unknown` / PID 0 / `unverified`.
3. `program` in `commonShells` → `loaded` / PID 0 / `verified`.
4. Resolve `UserName` to a numeric UID via `os/user.Lookup` (cached per List call).
5. Scan the process table (a `map[int]processInfo{UID, PPID, Args}` built from a single `ps -axo uid=,pid=,ppid=,args=` call) for entries where `UID` matches the resolved UID, `PPID == 1` (launchd), and `Args` contains the program path. Sort the matches ascending.
6. Exactly 1 kept → `running` / PID / `verified`; 0 kept → `stopped` / 0 / `verified`; more than 1 → `running` / lowest PID / `unverified`.

`readOnlyManager.list` issues exactly one `ps` invocation and one UID cache per List call, shared across every detection call. This collapses the former per-service `pgrep` + per-candidate `ps ppid=` fan-out into a single subprocess fork — on a machine with 411 Apple system services, list latency drops from several seconds to ~100 ms. For the single-service `get()` path, `DetectSystemServiceStatus` fetches the table lazily when called with a nil argument. If the process-table fetch or UID lookup fails, detection degrades to `stopped` / `unverified` rather than a confident false negative.

### `StatusConfidence` field

The `Service` struct carries a `StatusConfidence string` (`verified` / `unverified`). When the value is `unverified`, the frontend renders an info icon next to the Status column with a tooltip explaining that the displayed status and PID may not correspond to the exact process launchd started. The icon is purely informational — there is no click handler, no Wails call, and no action button behind it.

## Plist Format Handling

- Plist format is auto-detected (XML or binary).
- Binary plists are converted to XML via `plutil` for display.
- The Summary page shows the original format.

## Update Preserves Unmodeled Plist Keys

- On **Update** (not Create), both `UserManager.Update` and `SystemManager.Update` perform read-merge-write: the existing on-disk plist is read into a `map[string]any` (`readPlistMap` in `plist_encode.go`), and keys LaunchPal does not model (`ProcessType`, `Nice`, `MachServices`, `Sockets`, resource limits, `ExitTimeOut`, …) are preserved verbatim. Modeled keys stay form-authoritative.
- The merge (`mergeUnmodeledKeys`) removes keys using the full `modeledPlistKeys` set — the single source of truth listing every key `BuildPlistDict` can emit (guarded by `TestModeledPlistKeys`). This is why clearing a key (e.g. `Run at Load` → `On Demand`) or toggling `StartInterval` ↔ `StartCalendarInterval` strips the stale key instead of inheriting the old value back from disk.
- Encoding is shared via `writePlistDict` (user) / `encodeDict` (system) so Create passes a fresh dict and Update passes the merged dict.
- If the existing plist cannot be read or parsed (e.g. a system daemon plist unreadable without Full Disk Access, or a corrupt file), the merge is skipped and Update degrades to a fresh `BuildPlistDict`-only write rather than failing. The system-domain read happens in the GUI process; the degradation point is annotated in `SystemManager.Update`. Even on this degrade path `SystemManager.Update` still issues `Bootout` using the routing name as the label (Create writes `<Label>.plist`, so name == label by construction) — skipping it would leave launchd running the stale in-memory definition, because `Bootstrap` fails on an already-loaded job and `kickstart -k` only restarts the loaded job rather than re-reading the on-disk plist. `Bootout` is best-effort, so a "not bootstrapped" error on a never-loaded daemon is harmless. (`UserManager.Update` is unaffected: it unconditionally calls `Stop(name)` up front, which already boots out by routing name without reading the plist.)
- Because `Disabled` is unmodeled, it is preserved like any other advanced key — editing a disabled service keeps it disabled, and there is no UI control to clear it (this is intended; re-enable via `launchctl enable`). This was a deliberate decision over the legacy "edit silently re-enables" behavior.

## Log Load Classification (LogsResult)

- `Manager.GetLogs` returns a structured `launchctl.LogsResult{Content, Status, Path}` instead of a bare string. `Status` is one of `ok` / `no-path` / `not-found` (constants `LogStatusOK` / `LogStatusNoPath` / `LogStatusNotFound` in `types.go`). Structural states — no log path configured, log file not created yet — travel in `Status` with a nil error; only real failures (invalid log type, missing service, permission denied, other I/O) use the error channel. The not-found state is derived from the file-open result (`os.IsNotExist`), never from matching error message text.
- The classification lives in the shared `getServiceLogs` helper (`types.go`), used by `UserManager.GetLogs` (tilde-expanded path) and `readOnlyManager.getLogs` (plist-literal path, shared by SystemManager and AppleSystemManager). `Path` is the path actually opened; empty when `Status` is `no-path`.
- Frontend: Wails serializes with lowercase json keys (`content` / `status` / `path`); `ServiceLogs.vue` branches on `status` — `no-path` and `not-found` render neutral placeholders ("No {logType} log path configured for this service" / "Log file does not exist yet" + path), never the red error branch. Promise rejections show the backend message verbatim (Wails v2 rejects with the Go error as a plain string, so the string check comes before `instanceof Error`).

## Log ANSI Color Rendering

- `ServiceLogs.vue` renders log content via `v-html` bound to a `renderedLogs` computed, which runs the `LogsResult` content string through `frontend/app/utils/ansiToHtml.ts` (a hand-written, dependency-free SGR parser) — this is the single controlled `v-html` XSS surface for all three service types.
- Supported SGR subset: reset (`0`), bold (`1`), underline (`4`), foreground 30–37 / 90–97, background 40–47 / 100–107. Colors map to an OneDark-derived palette (`SGR_COLOR_MAP`) chosen for contrast on the dark `bg-surface-500`.
- Everything else — 256-color/truecolor, other SGR codes, non-SGR CSI, OSC/DCS/etc., and malformed sequences — is stripped silently (no throw, no UI error). Plain text is HTML-escaped; emitted `<span>` style attributes are limited to a four-property whitelist (`color`, `background-color`, `font-weight`, `text-decoration`) and never interpolate log text.
- The `<pre>` branch keys on the raw `content` (not on `renderedLogs`): content made of pure ANSI/control sequences renders to an empty string but must still get the `<pre>`, not the placeholder. Empty content falls through to the shared placeholder block, whose wording follows the log-load classification above.

## Log Auto-refresh

- `ServiceLogs.vue` exposes an **Auto-refresh** checkbox next to Auto-scroll for all three service types. It is component-local and session-only (never persisted to settings) and defaults to unchecked — polling is an explicit opt-in because each tick re-reads the whole ≤1MB tail and re-runs `ansiToHtml`.
- While checked, a `setInterval` (`AUTO_REFRESH_INTERVAL_MS = 2000`, handle in `pollTimer`) reloads the current stream through the same `loadLogs` path as the manual Refresh button. The tick is guarded by `if (loading.value) return` so a slow load never stacks concurrent polls. A single `watch(autoRefresh)` owns the timer lifecycle (start/stop); nothing else calls `setInterval`/`clearInterval`.
- **Auto-disable on non-ok outcome**: `loadLogs` computes `loadOk = logs.value?.status === 'ok'` (reading the lowercase runtime key) and, on any non-ok result — `no-path` / `not-found` / promise rejection / development fallback with no `LogsResult` — sets `autoRefresh = false`, which routes through the watcher to `stopPolling`. The check lives on the shared post-load path so polling, manual Refresh, and stream switches behave identically. There is no auto-resume; the user re-checks the box. The `loadOk` flag is required rather than re-deriving from `logs.value` because the catch path leaves `logs.value` holding the previous successful result.
- **Orthogonal to Auto-scroll** (Decision 4): Auto-refresh controls *when* to reload; Auto-scroll controls *whether* to scroll to bottom after a load. Follow mode = both checked.
- **Lifecycle**: `onBeforeUnmount` calls `stopPolling` alongside the existing `clearSuccessTimeout` cleanup. `watch(logType)` preserves the toggle across a stdout/stderr switch (loadLogs auto-disables if the new stream is non-ok). `watch(() => props.serviceName)` resets the toggle to unchecked and stops polling — the detail page has no `:key` on `ServiceLogs`, so a hypothetical in-place `serviceName` change would not remount it; Auto-refresh is a per-service choice and must not silently carry over.

## Concurrent Load Request Sequencing

- `loadLogs()` and `loadLogClearStatus()` are shared async functions invoked by several triggers (mount, the `logType` stdout/stderr switch, manual Refresh, the Auto-refresh poll tick, and `confirmClear`'s post-clear reload) and two invocations can be in flight at once. Each function carries its own **monotonic module-scoped counter** (`loadSeq` for `loadLogs`, `clearStatusSeq` for `loadLogClearStatus`); every call claims `const seq = ++<counter>` synchronously **before its first `await`** and applies its result only while `seq === <counter>` — a superseded (older) call discards its outcome without mutating shared state on both the resolve and reject paths.
- In `loadLogs` the guard gates the `logs.value` assignment, the `error.value` catch assignment, the `finally` `loading.value = false` reset (only the newest load clears the spinner), and the `loadOk`-driven Auto-refresh auto-disable; the `await nextTick(); scrollToBottom()` micro-window is intentionally left unguarded (a stale scroll is harmless). In `loadLogClearStatus` it gates the `logClearStatus.value` assignment on both the resolve and the silent-fail (`null`) paths so the Clear button's enabled state/tooltip always describes the stream currently shown. The two counters are independent — a log-content load must not supersede a status query or vice versa. The `if (loading.value) return` poll-tick skip only prevents poll-versus-poll stacking; the sequence tokens are what prevent cross-trigger overlap from applying stale results.

## Commit Message Conventions

This project uses [release-please](https://github.com/googleapis/release-please) to automate version bumps and releases (see `.github/workflows/release-please.yml`). A `feat` commit triggers a minor bump; `fix` triggers a patch.

- For changes that **do not affect production behavior** (docs, CI, configuration, tests, refactors), use `chore`, `docs`, `ci`, `test`, `refactor`, etc., so no unnecessary release is cut.
- Use `feat` **only** when introducing or changing user-visible functionality.
- Use `fix` **only** when fixing an actual bug.
- When the user asks to commit pending changes, **split by concern** — do not bundle unrelated work into a single commit.

## Homebrew Distribution

- Homebrew tap: `chenwei791129/homebrew-apps` (a standalone repo, shareable across multiple apps).
- Install command: `brew install --cask chenwei791129/apps/launchpal`.
- The cask formula lives at `Casks/launchpal.rb` inside the `homebrew-apps` repo.
- Because the app is not code-signed, a `postflight` block automatically clears the quarantine attribute.
- The `update-homebrew` job in `release-please.yml` updates the formula (version + SHA256) on each release.
- Cross-repo writes use `HOMEBREW_TAP_TOKEN` (a fine-grained PAT scoped to `homebrew-apps`).

## Known Limitations

- User services (`~/Library/LaunchAgents`) can be managed without any authorization.
- System services (`/Library/LaunchDaemons`) require Admin Mode; enabling it prompts for authorization **once per LaunchPal session** — there is no cross-session credential cache by design.
- Apple system services (`/System/Library/LaunchDaemons`) are always read-only (SIP).
- Some system services require Full Disk Access to be readable.

## Worktree Setup

Use the `.worktrees/` directory (already in `.gitignore`).
