# privileged-helper-integrity Specification

## Purpose

TBD - created by archiving change 'privileged-helper-launch-integrity'. Update Purpose after archive.

## Requirements

### Requirement: Root-owned protected helper copy

When the privileged helper starts as root from a location other than the protected path, it SHALL install a copy of itself at `/Library/Application Support/LaunchPal/launchpal-privhelper`, creating any missing parent directory. The copy source SHALL be the helper's own running executable image (opened via its own executable path with `O_NOFOLLOW`), NOT a re-read of the bundle path, so that no second substitution window is opened between the launcher's verification and the helper's self-read. The installed binary and its created parent directory SHALL be owned by UID 0 / GID 0 (`root:wheel`) with mode `0755`, so that no non-root process can create, replace, or write the path. The installation SHALL be idempotent and SHALL refuse to follow a symlink at the final path component.

#### Scenario: First install on enable

- **WHEN** the helper starts as root from the app bundle and no protected copy exists
- **THEN** it copies its own running image to `/Library/Application Support/LaunchPal/launchpal-privhelper` owned `root:wheel` mode `0755`, creating the parent directory owned `root:wheel` mode `0755`

#### Scenario: Idempotent when already current

- **WHEN** the helper starts and a protected copy already exists whose contents match the running binary
- **THEN** it performs no copy and leaves the existing protected copy unchanged

#### Scenario: Skip self-install when launched from protected path

- **WHEN** the helper's own executable path equals the protected path
- **THEN** it performs no self-install

#### Scenario: Refuse symlinked target

- **WHEN** the protected path's final component is a symlink
- **THEN** the helper does not write through the symlink and reports the install as failed


<!-- @trace
source: privileged-helper-launch-integrity
updated: 2026-07-23
code:
  - Makefile
  - frontend/app/composables/useAdminMode.ts
  - internal/privhelper/integrity.go
  - admin_mode.go
  - README.md
  - frontend/app/pages/settings.vue
  - internal/launchctl/readonly.go
  - internal/privhelper/server.go
  - internal/privhelper/handlers.go
  - cmd/launchpal-privhelper/main.go
  - internal/privhelper/logpath_darwin.go
  - cmd/launchpal-privhelper/procinfo_darwin.go
  - internal/privhelper/client.go
  - internal/privhelper/install.go
  - internal/privhelper/logpath.go
  - internal/privhelper/logpath_other.go
  - main.go
  - internal/launchctl/user.go
  - cmd/launchpal-privhelper/procinfo_other.go
  - internal/launchctl/system.go
  - .github/workflows/build.yml
tests:
  - internal/launchctl/system_test.go
  - internal/privhelper/server_test.go
  - internal/privhelper/client_test.go
  - internal/privhelper/integrity_test.go
  - internal/privhelper/install_test.go
  - admin_mode_test.go
  - internal/privhelper/handlers_test.go
  - resolve_helper_test.go
  - cmd/launchpal-privhelper/helper_test.go
  - internal/launchctl/user_test.go
-->

---
### Requirement: Helper self-install does not block Admin Mode

Failure of the protected-copy installation SHALL NOT prevent the current Admin Mode session from operating. The helper SHALL log the failure and continue serving RPCs from the binary it was launched as; the protected copy SHALL be retried on a subsequent enable.

#### Scenario: Install failure is non-fatal

- **WHEN** the helper cannot write the protected copy (for example, a filesystem error)
- **THEN** the helper still binds the socket and serves RPCs for the current session, and Admin Mode reaches Enabled


<!-- @trace
source: privileged-helper-launch-integrity
updated: 2026-07-23
code:
  - Makefile
  - frontend/app/composables/useAdminMode.ts
  - internal/privhelper/integrity.go
  - admin_mode.go
  - README.md
  - frontend/app/pages/settings.vue
  - internal/launchctl/readonly.go
  - internal/privhelper/server.go
  - internal/privhelper/handlers.go
  - cmd/launchpal-privhelper/main.go
  - internal/privhelper/logpath_darwin.go
  - cmd/launchpal-privhelper/procinfo_darwin.go
  - internal/privhelper/client.go
  - internal/privhelper/install.go
  - internal/privhelper/logpath.go
  - internal/privhelper/logpath_other.go
  - main.go
  - internal/launchctl/user.go
  - cmd/launchpal-privhelper/procinfo_other.go
  - internal/launchctl/system.go
  - .github/workflows/build.yml
tests:
  - internal/launchctl/system_test.go
  - internal/privhelper/server_test.go
  - internal/privhelper/client_test.go
  - internal/privhelper/integrity_test.go
  - internal/privhelper/install_test.go
  - admin_mode_test.go
  - internal/privhelper/handlers_test.go
  - resolve_helper_test.go
  - cmd/launchpal-privhelper/helper_test.go
  - internal/launchctl/user_test.go
-->

---
### Requirement: Bundle helper hash pinning

LaunchPal SHALL embed the SHA-256 digest of the packaged helper binary into the main application binary at build time, injected via linker flags before the main binary is linked (not after packaging). The pin SHALL gate only the launch of a bundle helper copy; it SHALL NOT be a precondition for launching an already-installed verified protected copy. When LaunchPal is about to launch a bundle copy and the pin is non-empty, LaunchPal SHALL compute the on-disk bundle helper's SHA-256 and compare it against the pin, refusing to launch on mismatch. When the pin is empty (a local development build with no injected pin), LaunchPal SHALL skip the bundle hash comparison; an empty pin SHALL NOT cause a valid protected copy to be bypassed.

#### Scenario: Bundle hash matches pin

- **WHEN** LaunchPal is about to launch a bundle copy and the bundle helper's SHA-256 equals the non-empty pin
- **THEN** LaunchPal proceeds to launch the bundle helper

#### Scenario: Bundle hash mismatch on first install

- **WHEN** no verified protected copy exists and the bundle helper's SHA-256 differs from the non-empty pin
- **THEN** LaunchPal does not launch the helper and Admin Mode returns to Disabled with an integrity error

#### Scenario: Empty pin does not bypass a valid protected copy

- **WHEN** the embedded pin is empty and a verified protected copy exists
- **THEN** LaunchPal launches the protected copy and does not fall back to the bundle copy


<!-- @trace
source: privileged-helper-launch-integrity
updated: 2026-07-23
code:
  - Makefile
  - frontend/app/composables/useAdminMode.ts
  - internal/privhelper/integrity.go
  - admin_mode.go
  - README.md
  - frontend/app/pages/settings.vue
  - internal/launchctl/readonly.go
  - internal/privhelper/server.go
  - internal/privhelper/handlers.go
  - cmd/launchpal-privhelper/main.go
  - internal/privhelper/logpath_darwin.go
  - cmd/launchpal-privhelper/procinfo_darwin.go
  - internal/privhelper/client.go
  - internal/privhelper/install.go
  - internal/privhelper/logpath.go
  - internal/privhelper/logpath_other.go
  - main.go
  - internal/launchctl/user.go
  - cmd/launchpal-privhelper/procinfo_other.go
  - internal/launchctl/system.go
  - .github/workflows/build.yml
tests:
  - internal/launchctl/system_test.go
  - internal/privhelper/server_test.go
  - internal/privhelper/client_test.go
  - internal/privhelper/integrity_test.go
  - internal/privhelper/install_test.go
  - admin_mode_test.go
  - internal/privhelper/handlers_test.go
  - resolve_helper_test.go
  - cmd/launchpal-privhelper/helper_test.go
  - internal/launchctl/user_test.go
-->

---
### Requirement: Launch prefers verified protected copy

The trust of the protected copy SHALL derive solely from its root ownership and permissions, NOT from matching any bundle hash. Before launching the helper, LaunchPal SHALL apply this decision, keeping the `resolveHelperPath` return shape unchanged (path plus error; no added return value):

1. When a verified protected copy exists, LaunchPal SHALL launch the protected copy, EXCEPT when all of the following hold, in which case it SHALL launch the bundle copy to trigger re-provisioning: the pin is non-empty, the bundle helper is readable, the bundle helper's SHA-256 equals the pin, and the bundle helper's SHA-256 differs from the protected copy's SHA-256.
2. When no verified protected copy exists, LaunchPal SHALL launch the bundle copy after it passes hash-pin verification (or, with an empty pin, without it).

#### Scenario: Steady-state launch uses protected copy

- **WHEN** a verified protected copy exists and the bundle helper's hash equals it (or the pin is empty)
- **THEN** LaunchPal launches the protected copy

#### Scenario: Missing or unreadable bundle does not disable Admin Mode

- **WHEN** a verified protected copy exists but the bundle helper is deleted or unreadable
- **THEN** LaunchPal launches the protected copy and Admin Mode still reaches Enabled

#### Scenario: Tampered bundle does not bypass protected copy

- **WHEN** a verified protected copy exists and the bundle helper has been overwritten so its hash no longer matches the non-empty pin
- **THEN** LaunchPal launches the protected copy and ignores the tampered bundle

#### Scenario: Legitimate update re-provisions protected copy

- **WHEN** a verified protected copy exists, the pin is non-empty, and the bundle helper matches the pin but differs from the protected copy
- **THEN** LaunchPal launches the bundle copy, which re-installs the protected copy

#### Scenario: First install with no protected copy

- **WHEN** no verified protected copy exists and the bundle helper matches the pin (or the pin is empty)
- **THEN** LaunchPal launches the bundle copy, from which the helper installs the protected copy


<!-- @trace
source: privileged-helper-launch-integrity
updated: 2026-07-23
code:
  - Makefile
  - frontend/app/composables/useAdminMode.ts
  - internal/privhelper/integrity.go
  - admin_mode.go
  - README.md
  - frontend/app/pages/settings.vue
  - internal/launchctl/readonly.go
  - internal/privhelper/server.go
  - internal/privhelper/handlers.go
  - cmd/launchpal-privhelper/main.go
  - internal/privhelper/logpath_darwin.go
  - cmd/launchpal-privhelper/procinfo_darwin.go
  - internal/privhelper/client.go
  - internal/privhelper/install.go
  - internal/privhelper/logpath.go
  - internal/privhelper/logpath_other.go
  - main.go
  - internal/launchctl/user.go
  - cmd/launchpal-privhelper/procinfo_other.go
  - internal/launchctl/system.go
  - .github/workflows/build.yml
tests:
  - internal/launchctl/system_test.go
  - internal/privhelper/server_test.go
  - internal/privhelper/client_test.go
  - internal/privhelper/integrity_test.go
  - internal/privhelper/install_test.go
  - admin_mode_test.go
  - internal/privhelper/handlers_test.go
  - resolve_helper_test.go
  - cmd/launchpal-privhelper/helper_test.go
  - internal/launchctl/user_test.go
-->

---
### Requirement: Ownership and permission verification of protected copy

LaunchPal SHALL treat a protected copy as unverified when any of the following holds: it does not exist, it is a symlink or non-regular file, its owner UID is not 0, or it is writable by group or other (`mode & 022 != 0`). An unverified protected copy SHALL NOT be launched; LaunchPal SHALL treat the situation as "no verified protected copy" and take the first-install path.

#### Scenario: Protected copy owned by non-root

- **WHEN** the protected copy exists but its owner UID is not 0
- **THEN** LaunchPal treats it as unverified and takes the first-install path

#### Scenario: Protected copy group- or world-writable

- **WHEN** the protected copy's mode has any group or other write bit set
- **THEN** LaunchPal treats it as unverified and takes the first-install path


<!-- @trace
source: privileged-helper-launch-integrity
updated: 2026-07-23
code:
  - Makefile
  - frontend/app/composables/useAdminMode.ts
  - internal/privhelper/integrity.go
  - admin_mode.go
  - README.md
  - frontend/app/pages/settings.vue
  - internal/launchctl/readonly.go
  - internal/privhelper/server.go
  - internal/privhelper/handlers.go
  - cmd/launchpal-privhelper/main.go
  - internal/privhelper/logpath_darwin.go
  - cmd/launchpal-privhelper/procinfo_darwin.go
  - internal/privhelper/client.go
  - internal/privhelper/install.go
  - internal/privhelper/logpath.go
  - internal/privhelper/logpath_other.go
  - main.go
  - internal/launchctl/user.go
  - cmd/launchpal-privhelper/procinfo_other.go
  - internal/launchctl/system.go
  - .github/workflows/build.yml
tests:
  - internal/launchctl/system_test.go
  - internal/privhelper/server_test.go
  - internal/privhelper/client_test.go
  - internal/privhelper/integrity_test.go
  - internal/privhelper/install_test.go
  - admin_mode_test.go
  - internal/privhelper/handlers_test.go
  - resolve_helper_test.go
  - cmd/launchpal-privhelper/helper_test.go
  - internal/launchctl/user_test.go
-->

---
### Requirement: Integrity failure refuses helper launch

When no verified protected copy exists and bundle hash pinning fails (the bundle helper is missing/unreadable or its SHA-256 differs from a non-empty pin), Admin Mode enablement SHALL NOT execute osascript or launch the helper, and state SHALL return to `Disabled` with a distinct integrity error code (`helper_integrity_failed`). The error message SHALL NOT disclose information beyond what the existing helper-not-found and handshake errors already reveal.

#### Scenario: Enable aborts on integrity failure

- **WHEN** no verified protected copy exists and the bundle helper hash does not match the non-empty pin during Enable
- **THEN** osascript is not executed, no helper is launched, and Admin Mode ends in `Disabled` with error `helper_integrity_failed`

<!-- @trace
source: privileged-helper-launch-integrity
updated: 2026-07-23
code:
  - Makefile
  - frontend/app/composables/useAdminMode.ts
  - internal/privhelper/integrity.go
  - admin_mode.go
  - README.md
  - frontend/app/pages/settings.vue
  - internal/launchctl/readonly.go
  - internal/privhelper/server.go
  - internal/privhelper/handlers.go
  - cmd/launchpal-privhelper/main.go
  - internal/privhelper/logpath_darwin.go
  - cmd/launchpal-privhelper/procinfo_darwin.go
  - internal/privhelper/client.go
  - internal/privhelper/install.go
  - internal/privhelper/logpath.go
  - internal/privhelper/logpath_other.go
  - main.go
  - internal/launchctl/user.go
  - cmd/launchpal-privhelper/procinfo_other.go
  - internal/launchctl/system.go
  - .github/workflows/build.yml
tests:
  - internal/launchctl/system_test.go
  - internal/privhelper/server_test.go
  - internal/privhelper/client_test.go
  - internal/privhelper/integrity_test.go
  - internal/privhelper/install_test.go
  - admin_mode_test.go
  - internal/privhelper/handlers_test.go
  - resolve_helper_test.go
  - cmd/launchpal-privhelper/helper_test.go
  - internal/launchctl/user_test.go
-->