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
│   │   └── status_detect.go  # Heuristic status detection for the system domain (pgrep -u + ppid=1 filter)
│   ├── privhelper/           # RPC protocol + server + client shared by app.go and the helper binary
│   │   ├── protocol.go       # Request/Response types, method name + error code constants
│   │   ├── server.go         # Newline-delimited JSON server, peer UID verification, idle/parent watchdog
│   │   ├── client.go         # Client side: Connect (retry), LaunchHelper (osascript), typed method wrappers
│   │   ├── handlers.go       # Bootstrap/Bootout/Kickstart/WritePlist/DeletePlist/List handlers, path + label validation
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

## Admin Mode (session-scoped privileged helper)

Because LaunchPal is unsigned and cannot register a persistent `SMAppService` daemon, write access to `/Library/LaunchDaemons` is gated by a **session-scoped helper** the user enables explicitly.

- **Binary**: `launchpal-privhelper`, shipped inside the app bundle at `LaunchPal.app/Contents/MacOS/launchpal-privhelper`. Built by `make build-helper`, installed into the bundle by `make install-helper` (both run automatically from `make build`).
- **Launch**: triggered by the `EnableAdminMode` Wails binding. LaunchPal runs `osascript -e 'do shell script "... with administrator privileges"'`, prompting the user for their password or Touch ID once per session.
- **IPC**: Unix domain socket at `$TMPDIR/launchpal-<uid>-<16-hex-random>.sock`, `chmod 0600`, peer UID verified via `LOCAL_PEERCRED` before any RPC is processed. Messages are newline-delimited JSON.
- **Scope**: the helper can only Bootstrap/Bootout/Kickstart and WritePlist/DeletePlist under `/Library/LaunchDaemons/`, plus EnsureLogAccess/`TruncateLog`/`DeleteLogPaths` under the log allowlist (`/var/log/`, `/private/var/log/`, `/Library/Logs/`, `/tmp/`, `/private/tmp/`, sub-directory required). Paths outside these prefixes and labels with shell metacharacters are rejected before any `launchctl` invocation. `TruncateLog` opens with `O_WRONLY|O_TRUNC|O_NOFOLLOW` and refuses to create a missing file, so the helper cannot be coerced into materializing a 0-byte root-owned file in `/tmp/`. `DeleteLogPaths` accepts a list of log files, validates each against the allowlist, uses `os.Lstat` to refuse symlinks, removes each regular file, and best-effort removes the now-empty parent directory (`ENOTEMPTY` is silently ignored); per-path failures are returned as a partial-success warning slice rather than an RPC-level error so callers can treat them as non-fatal.
- **Log access**: after a successful Create/Update, `SystemManager` sends an `EnsureLogAccess` RPC with the plist's `StandardOutPath` / `StandardErrorPath`. The helper `MkdirAll`s the parent directory as `0755`, `Chmod`s it to `0755` if it already existed with a stricter mode (common launchd default is `0744` which blocks user traversal), and touches the log file as `0644` with `O_NOFOLLOW` if missing. Paths are restricted to the allowlist `/var/log/`, `/private/var/log/`, `/Library/Logs/`, `/tmp/`, `/private/tmp/` and must live at least one sub-directory deep — `/var/log/foo.log` is rejected to prevent re-moding a system log root.
- **Lifecycle**:
  - Parent watchdog polls LaunchPal's PID once per second via `syscall.Kill(pid, 0)`; when the parent is gone the helper removes the socket and exits within 2 seconds.
  - Idle timeout: 30 minutes without RPC traffic triggers self-exit.
  - Graceful shutdown via the `Shutdown` RPC (sent by `DisableAdminMode`) with a 3-second ack timeout.
- **Backups**: when the helper overwrites/deletes a plist it first writes a backup under `<user-home>/.launchpal/backups/<label>/` and `lchown`s the created files to the launching user's UID/GID so the user-side LaunchPal can read them. `O_NOFOLLOW` is used to block symlink attacks against the backup path.
- **State machine (frontend-visible)**:
  ```
  Disabled → Requesting → Enabled     (successful authorization + handshake)
                        ↓
                     Disabled          (user cancel / helper handshake failure / helper crash)
  Enabled → ShuttingDown → Disabled   (DisableAdminMode)
  Enabled → Disabled                   (helper crash — error: "helper_crashed")
  ```
  State changes emit the `admin_mode:state` Wails event. The frontend composable `useAdminMode()` exposes `state`, `lastError`, `isEnabled`, and `enable()`/`disable()` methods.

## Status Detection Logic

### User domain (`UserManager`)

1. Query `launchctl list <label>` for the service entry.
2. Parse a PID from the output (present ⇒ running).
3. If no PID, fall back to `pgrep -f <program>`.
4. Skip the pgrep fallback for common shells (bash, sh, zsh) to avoid false matches.
5. `StatusConfidence` is always `verified` because `launchctl list` is authoritative in the user domain.

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

## Log ANSI Color Rendering

- `ServiceLogs.vue` renders log content via `v-html` bound to a `renderedLogs` computed, which runs the raw log string through `frontend/app/utils/ansiToHtml.ts` (a hand-written, dependency-free SGR parser) — this is the single controlled `v-html` XSS surface for all three service types.
- Supported SGR subset: reset (`0`), bold (`1`), underline (`4`), foreground 30–37 / 90–97, background 40–47 / 100–107. Colors map to an OneDark-derived palette (`SGR_COLOR_MAP`) chosen for contrast on the dark `bg-surface-500`.
- Everything else — 256-color/truecolor, other SGR codes, non-SGR CSI, OSC/DCS/etc., and malformed sequences — is stripped silently (no throw, no UI error). Plain text is HTML-escaped; emitted `<span>` style attributes are limited to a four-property whitelist (`color`, `background-color`, `font-weight`, `text-decoration`) and never interpolate log text.
- The loading / error / "No logs available" branches are unchanged; empty or null logs short-circuit to the placeholder.

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
