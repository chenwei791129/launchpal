## ADDED Requirements

### Requirement: DMG image generation

The build system SHALL produce a `.dmg` disk image file using the `create-dmg` tool as the release artifact, replacing the previous `.app.gz` tar archive.

#### Scenario: DMG is created from built application

- **WHEN** the build process completes and `build/bin/launchpal.app` exists
- **THEN** the system SHALL invoke `create-dmg` to produce a DMG file containing `launchpal.app`

#### Scenario: DMG replaces gz archive

- **WHEN** the packaging step runs in CI
- **THEN** the system SHALL produce only a `.dmg` file and SHALL NOT produce a `.app.gz` file

### Requirement: Applications folder shortcut

The DMG image SHALL contain a symbolic link to `/Applications` so that users can drag the application to install it.

#### Scenario: User opens DMG and sees drag-to-install layout

- **WHEN** a user mounts the DMG file
- **THEN** the DMG SHALL display `LaunchPal.app` and an `Applications` folder shortcut side by side

### Requirement: CI workflow produces DMG

The GitHub Actions build workflow SHALL install `create-dmg` and use it to package the application into a DMG file.

#### Scenario: Build workflow packages as DMG

- **WHEN** the `build.yml` workflow runs the packaging step
- **THEN** the workflow SHALL install `create-dmg` via Homebrew and execute it against `build/bin/launchpal.app` to produce the DMG artifact

#### Scenario: Release workflow uploads DMG

- **WHEN** a new release is created by release-please
- **THEN** the `release-please.yml` workflow SHALL upload the `.dmg` file as the release asset instead of `.app.gz`

### Requirement: Local DMG build target

The Makefile SHALL provide a `dmg` target that builds the application and packages it into a DMG file locally.

#### Scenario: Developer runs make dmg

- **WHEN** a developer executes `make dmg`
- **THEN** the system SHALL first build the application via `wails build` and then run `create-dmg` to produce a DMG file in the `build/bin/` directory

#### Scenario: create-dmg not installed locally

- **WHEN** a developer executes `make dmg` without `create-dmg` installed
- **THEN** the command SHALL fail with an error indicating that `create-dmg` is required
