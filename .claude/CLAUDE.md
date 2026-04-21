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
├── main.go                   # Application entry point
├── Makefile                  # Build recipes
├── internal/
│   ├── launchctl/            # launchctl command wrappers
│   │   ├── types.go          # Service, ServiceConfig, and related types (includes StatusConfidence)
│   │   ├── manager.go        # Manager interface
│   │   ├── user.go           # UserManager (~/Library/LaunchAgents)
│   │   ├── system.go         # SystemManager (/Library/LaunchDaemons, read-only)
│   │   ├── apple_system.go   # AppleSystemManager (/System/Library/LaunchDaemons, read-only)
│   │   ├── readonly.go       # Shared read-only logic for SystemManager and AppleSystemManager
│   │   └── status_detect.go  # Heuristic status detection for the system domain (pgrep -u + ppid=1 filter)
│   ├── backup/               # Backup management
│   │   └── backup.go         # BackupManager implementation
│   └── plistutil/            # plist format detection and binary→XML normalization (shared by backup, launchctl)
│       └── plistutil.go      # DetectFormat, NormalizeFromPath
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
make test        # Run tests
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

## Service Types

LaunchPal supports three service categories:

1. **User Services** (`~/Library/LaunchAgents`)
   - Full read/write access.
   - Can start, stop, create, update, and delete services.
   - Supports immediate execution (Kickstart: `launchctl kickstart -k`).
   - Supports scheduling (`StartCalendarInterval` / `StartInterval`).
   - Cron syntax accepts ranges (`a-b`) and enumerations (`a,b,c`) and is expanded into the Cartesian product of `StartCalendarInterval` entries (capped at 50 entries).
   - Supports environment variables (`EnvironmentVariables`).

2. **System Services** (`/Library/LaunchDaemons`)
   - Read-only.
   - Can view service information, status, plist contents, and logs.
   - Third-party system-level services.

3. **Apple System Services** (`/System/Library/LaunchDaemons`)
   - Read-only.
   - Can view service information, status, plist contents, and logs.
   - macOS built-in services.
   - Many of these use the binary plist format and are automatically converted to XML for display.

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

- Write operations are limited to user-level services (`~/Library/LaunchAgents`).
- System services (`/Library/LaunchDaemons`, `/System/Library/LaunchDaemons`) are read-only.
- Cannot stop services running as root (would require sudo).
- Some system services require Full Disk Access to be readable.

## Worktree Setup

Use the `.worktrees/` directory (already in `.gitignore`).
