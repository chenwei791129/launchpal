## MODIFIED Requirements

### Requirement: Helper binary packaged in app bundle

The `launchpal-privhelper` binary SHALL be compiled and placed at `LaunchPal.app/Contents/MacOS/launchpal-privhelper` during the build. At runtime LaunchPal SHALL resolve the helper to launch by locating the bundle helper via `os.Executable()` (the main binary directory joined with `launchpal-privhelper`) as the packaged source, and SHALL launch a verified root-owned protected copy whenever one exists — its trust deriving from root ownership and permissions, not from matching the bundle. LaunchPal SHALL launch the bundle copy only when no verified protected copy exists, or when a non-empty pin proves the bundle copy is a legitimate update differing from the protected copy.

#### Scenario: Bundle helper is the packaged source

- **WHEN** LaunchPal is built
- **THEN** the helper is packaged at `<main-binary-dir>/launchpal-privhelper` and is used as the source from which the protected copy is provisioned

#### Scenario: Runtime launch prefers verified protected copy

- **WHEN** LaunchPal enables Admin Mode and a verified protected copy exists
- **THEN** the protected copy is launched instead of the bundle copy, regardless of whether the bundle copy is present, altered, or hash-matched

#### Scenario: Bundle copy used only for first install or legitimate update

- **WHEN** no verified protected copy exists, or a non-empty pin proves the bundle copy is a legitimate update differing from the protected copy
- **THEN** LaunchPal launches the bundle copy after hash-pin verification

#### Scenario: Helper binary missing

- **WHEN** no verified protected copy exists and the bundle helper is not found
- **THEN** Admin Mode enablement fails with an error identifying the missing binary path
