## ADDED Requirements

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
