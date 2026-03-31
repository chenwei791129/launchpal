## ADDED Requirements

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

### Requirement: Postflight removes quarantine attribute

The cask formula SHALL include a `postflight` stanza that executes `xattr -dr com.apple.quarantine` on the installed app bundle.

This is required because the app is not code-signed or notarized, and macOS Gatekeeper would otherwise block the app from opening.

#### Scenario: App opens without Gatekeeper warning after installation

- **WHEN** the cask installation completes (including postflight)
- **THEN** the quarantine extended attribute is removed from `LaunchPal.app`
- **THEN** user can open the app without macOS displaying a Gatekeeper warning

### Requirement: Caveats inform user about quarantine removal

The cask formula SHALL include a `caveats` string that explains:
1. The app is not code-signed
2. The quarantine attribute has been automatically removed during installation
3. This is safe because the app is open-source

#### Scenario: User sees caveats after installation

- **WHEN** cask installation completes
- **THEN** Homebrew displays the caveats text to the user in the terminal

### Requirement: Cask supports uninstallation

The cask formula SHALL include an `uninstall` stanza using `quit` and `delete` to cleanly remove the app.

#### Scenario: User uninstalls LaunchPal

- **WHEN** user runs `brew uninstall --cask launchpal`
- **THEN** LaunchPal.app is quit (if running) and deleted from `/Applications/`
