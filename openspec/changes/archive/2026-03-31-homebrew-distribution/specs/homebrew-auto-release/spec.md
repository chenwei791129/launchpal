## ADDED Requirements

### Requirement: Release workflow updates cask formula automatically

When a new GitHub Release is created via release-please, the release workflow SHALL automatically update the `Casks/launchpal.rb` file in the `chenwei791129/homebrew-apps` repo with:
- The new version number (extracted from the release tag, without `v` prefix)
- The new SHA256 checksum of the DMG file
- The new download URL for the DMG asset

#### Scenario: New release triggers cask formula update

- **WHEN** release-please creates a new release (e.g., `v1.2.0`)
- **AND** the build workflow produces and uploads `LaunchPal.dmg` to the release
- **THEN** the workflow downloads the DMG from the release
- **THEN** the workflow computes the SHA256 checksum of the downloaded DMG
- **THEN** the workflow updates `version`, `sha256`, and `url` in `Casks/launchpal.rb` of the `homebrew-apps` repo
- **THEN** the workflow commits and pushes the change to `homebrew-apps` main branch

### Requirement: Cross-repo authentication uses fine-grained PAT

The release workflow SHALL authenticate to the `homebrew-apps` repo using a Fine-grained Personal Access Token stored as the `HOMEBREW_TAP_TOKEN` secret in the `launchpal` repo.

The token SHALL have `contents:write` permission scoped only to the `chenwei791129/homebrew-apps` repository.

#### Scenario: Workflow authenticates and pushes to homebrew-apps

- **WHEN** the release workflow reaches the cask update step
- **THEN** it uses `HOMEBREW_TAP_TOKEN` to clone, modify, commit, and push to `chenwei791129/homebrew-apps`

#### Scenario: Missing or expired token causes workflow failure

- **WHEN** the `HOMEBREW_TAP_TOKEN` secret is missing or expired
- **THEN** the cask update step fails with a clear error message
- **THEN** the rest of the release (DMG upload to GitHub Release) is NOT affected (the cask update step runs after release asset upload)

### Requirement: README documents Homebrew installation

The `README.md` in the `launchpal` repo SHALL include a Homebrew installation section with:
- The fully qualified install command: `brew install --cask chenwei791129/apps/launchpal`
- The two-step alternative: `brew tap chenwei791129/apps` then `brew install --cask launchpal`

#### Scenario: User finds installation instructions in README

- **WHEN** user visits the LaunchPal GitHub repository
- **THEN** the README displays Homebrew Cask installation commands in an Installation section
