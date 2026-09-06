# homebrew-cask-formula Specification

## Purpose

The `Casks/launchpal.rb` formula in `chenwei791129/homebrew-apps` is the canonical Homebrew installation source for LaunchPal. It pins to the latest GitHub Release DMG and lets users install via `brew install --cask chenwei791129/apps/launchpal` (or via the two-step `brew tap` + `brew install` flow).

## Requirements

### Requirement: Cask formula provides standard Homebrew installation

The `Casks/launchpal.rb` file in `chenwei791129/homebrew-apps` repo SHALL define a valid Homebrew Cask formula that installs LaunchPal.app to `/Applications/`.

The formula SHALL include:
- `cask` name: `launchpal`
- `version`: matching the latest GitHub Release tag (without `v` prefix)
- `url`: pointing to `LaunchPal.dmg` asset on the GitHub Release
- `sha256`: matching the SHA256 checksum of the DMG file
- `name`: `LaunchPal`
- `homepage`: pointing to the GitHub repository URL

#### Scenario: User installs LaunchPal via Homebrew

- **WHEN** user runs `brew install --cask chenwei791129/apps/launchpal`
- **THEN** Homebrew downloads the DMG from the GitHub Release URL
- **THEN** LaunchPal.app is installed to `/Applications/`

#### Scenario: User installs via two-step tap method

- **WHEN** user runs `brew tap chenwei791129/apps` followed by `brew install --cask launchpal`
- **THEN** the result is identical to using the fully qualified name


<!-- @trace
source: homebrew-distribution
updated: 2026-03-31
code:
  - .github/workflows/release-please.yml
  - README.md
-->

---
### Requirement: Postflight steps remove quarantine attribute

The cask formula SHALL include a `postflight_steps` stanza whose `run` step executes `xattr -dr com.apple.quarantine` on the installed app bundle. The app path SHALL be written as the install-steps template token `{{appdir}}/launchpal.app`, because a steps block serialises `args` verbatim and does not evaluate Ruby interpolation.

This is required because the app is not code-signed or notarized, and macOS Gatekeeper would otherwise block the app from opening.

#### Scenario: App opens without Gatekeeper warning after installation

- **WHEN** the cask installation completes (including the postflight steps)
- **THEN** the quarantine extended attribute is removed from `LaunchPal.app`
- **THEN** user can open the app without macOS displaying a Gatekeeper warning


<!-- @trace
source: homebrew-distribution
updated: 2026-03-31
code:
  - .github/workflows/release-please.yml
  - README.md
-->

---
### Requirement: Caveats inform user about quarantine removal

The cask formula SHALL include a `caveats` string that explains:
1. The app is not code-signed
2. The quarantine attribute has been automatically removed during installation
3. This is safe because the app is open-source

#### Scenario: User sees caveats after installation

- **WHEN** cask installation completes
- **THEN** Homebrew displays the caveats text to the user in the terminal


<!-- @trace
source: homebrew-distribution
updated: 2026-03-31
code:
  - .github/workflows/release-please.yml
  - README.md
-->

---
### Requirement: Cask supports uninstallation

The cask formula SHALL include an `uninstall` stanza using `quit` and `delete` to cleanly remove the app.

#### Scenario: User uninstalls LaunchPal

- **WHEN** user runs `brew uninstall --cask launchpal`
- **THEN** LaunchPal.app is quit (if running) and deleted from `/Applications/`

<!-- @trace
source: homebrew-distribution
updated: 2026-03-31
code:
  - .github/workflows/release-please.yml
  - README.md
-->

---
### Requirement: Cask declares the minimum macOS version

The `Casks/launchpal.rb` formula in `chenwei791129/homebrew-apps` SHALL include a `depends_on macos:` stanza requiring Ventura or later, matching the minimum system version declared by the app bundle.

This stanza rejects the installation on an incompatible system at install time. Without it, Homebrew installs an app the user cannot launch, and the resulting failure surfaces only when the user opens it — where an unsigned app makes the version problem easy to misread as a Gatekeeper or quarantine issue.

The declared version SHALL be kept consistent with the app bundle's `LSMinimumSystemVersion`. The cask lives in a separate repository, so no test or CI job in the LaunchPal repository can enforce this consistency; it MUST be updated as an explicit step whenever the minimum system version changes.

#### Scenario: User on an unsupported macOS version installs via Homebrew

- **WHEN** a user on macOS 12 or earlier runs `brew install --cask chenwei791129/apps/launchpal`
- **THEN** Homebrew refuses the installation and reports that a newer macOS version is required
- **AND** no DMG is downloaded and no app is placed in `/Applications/`

#### Scenario: User on a supported macOS version installs via Homebrew

- **WHEN** a user on macOS 13 or later runs `brew install --cask chenwei791129/apps/launchpal`
- **THEN** the installation proceeds exactly as before, including the postflight quarantine removal

<!-- @trace
source: upgrade-go-and-macos-baseline
updated: 2026-08-26
code:
  - go.mod
  - Makefile
  - build/darwin/Info.plist
  - build/darwin/Info.dev.plist
  - README.md
tests:
  - build_metadata_test.go
-->