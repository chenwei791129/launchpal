# build-version-injection Specification

## Purpose

TBD - created by archiving change 'auto-version-injection'. Update Purpose after archive.

## Requirements

### Requirement: Version variable with build-time injection

The application SHALL declare a package-level `version` variable in `main.go` with a default value of `"dev"`. The variable SHALL be overridable at build time via Go linker flags (`-ldflags "-X main.version=<value>"`).

#### Scenario: Default version in local development

- **WHEN** the application is built without `-ldflags` version override
- **THEN** the version variable SHALL have the value `"dev"`

#### Scenario: Version injected during CI build

- **WHEN** the application is built with `-ldflags "-X main.version=v1.6.0"`
- **THEN** the version variable SHALL have the value `v1.6.0`


<!-- @trace
source: auto-version-injection
updated: 2026-03-31
code:
  - frontend/wailsjs/go/main/App.js
  - .github/workflows/build.yml
  - .github/workflows/release-please.yml
  - frontend/app/pages/settings.vue
  - app.go
  - main.go
  - frontend/app/components/StatusBar.vue
  - frontend/wailsjs/go/main/App.d.ts
tests:
  - app_test.go
-->

---
### Requirement: Version API binding

The `App` struct SHALL expose a `GetVersion() string` method that returns the current version value. This method SHALL be available as a Wails binding callable from the frontend.

#### Scenario: Frontend retrieves version

- **WHEN** the frontend calls the `GetVersion` Wails binding
- **THEN** the method SHALL return the current value of the version variable


<!-- @trace
source: auto-version-injection
updated: 2026-03-31
code:
  - frontend/wailsjs/go/main/App.js
  - .github/workflows/build.yml
  - .github/workflows/release-please.yml
  - frontend/app/pages/settings.vue
  - app.go
  - main.go
  - frontend/app/components/StatusBar.vue
  - frontend/wailsjs/go/main/App.d.ts
tests:
  - app_test.go
-->

---
### Requirement: CI pipeline version propagation

The `build.yml` workflow SHALL accept a `version` input parameter via `workflow_call`. When the `version` input is provided, the build step SHALL pass it to the Go linker via `-ldflags "-X main.version=<version>"`. The `release-please.yml` workflow SHALL pass the `tag_name` output to `build.yml` as the `version` input.

#### Scenario: Release build receives version from release-please

- **WHEN** release-please creates a release with tag `v1.7.0`
- **THEN** `release-please.yml` SHALL pass `v1.7.0` to `build.yml` as the version input
- **AND** `build.yml` SHALL build with `-ldflags "-X main.version=v1.7.0"`

#### Scenario: PR build without version input

- **WHEN** `build.yml` is triggered by a pull request (no version input)
- **THEN** the build SHALL proceed without version ldflags, resulting in the default `"dev"` value


<!-- @trace
source: auto-version-injection
updated: 2026-03-31
code:
  - frontend/wailsjs/go/main/App.js
  - .github/workflows/build.yml
  - .github/workflows/release-please.yml
  - frontend/app/pages/settings.vue
  - app.go
  - main.go
  - frontend/app/components/StatusBar.vue
  - frontend/wailsjs/go/main/App.d.ts
tests:
  - app_test.go
-->

---
### Requirement: Frontend dynamic version display

The `StatusBar.vue` component and `settings.vue` page SHALL retrieve the version string by calling the `GetVersion` Wails binding instead of using a hardcoded value. The hardcoded `v0.1.0` strings SHALL be removed.

#### Scenario: StatusBar displays injected version

- **WHEN** the application is running with version `v1.6.0`
- **THEN** the StatusBar SHALL display `v1.6.0`

#### Scenario: StatusBar displays dev version

- **WHEN** the application is running in local development (no version injected)
- **THEN** the StatusBar SHALL display `dev`

#### Scenario: Settings page displays version

- **WHEN** the user navigates to the Settings page
- **THEN** the page SHALL display the same version string as the StatusBar

<!-- @trace
source: auto-version-injection
updated: 2026-03-31
code:
  - frontend/wailsjs/go/main/App.js
  - .github/workflows/build.yml
  - .github/workflows/release-please.yml
  - frontend/app/pages/settings.vue
  - app.go
  - main.go
  - frontend/app/components/StatusBar.vue
  - frontend/wailsjs/go/main/App.d.ts
tests:
  - app_test.go
-->