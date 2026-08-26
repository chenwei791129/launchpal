# macos-minimum-version Specification

## Purpose

TBD - created by archiving change 'upgrade-go-and-macos-baseline'. Update Purpose after archive.

## Requirements

### Requirement: Declared minimum macOS version matches the toolchain's binary requirement

LaunchPal SHALL declare macOS 13.0.0 (Ventura) as its minimum supported system version. The declared value MUST equal the `LC_BUILD_VERSION` `minos` value carried by every binary the project ships — both the main application binary and the privileged helper.

`minos` is not written by a single authority. An internally linked Go binary carries the Go toolchain's own default, while an externally linked one carries the deployment target the C toolchain was given. The project ships one of each, so the declaration MUST be checked against both rather than inferred from the Go version alone.

A declared value lower than `minos` MUST NOT be used: Launch Services permits launch based on the declaration while dyld rejects the load based on `minos`, producing an unexplained launch failure instead of a version message. A declared value higher than `minos` MUST NOT be used either, as it excludes systems that are in fact capable of running the binary.

When the Go toolchain is upgraded, the declared value SHALL be re-verified against the `minos` of a freshly built binary.

#### Scenario: Declaration and binary requirement agree

- **WHEN** the project is built with the Go toolchain named by the `go` directive in the module file
- **THEN** the `minos` value of both the main binary and the privileged helper binary is 13.0
- **AND** every location declaring a minimum macOS version declares 13.0.0

##### Example: outcomes for declaration versus binary minos

| Declared value | Binary minos | Result on macOS 12 | Acceptable |
| -------------- | ------------ | ------------------ | ---------- |
| 10.13.0 | 13.0 | Launch Services permits launch, dyld rejects the load with an unexplained failure | No |
| 13.0.0 | 13.0 | Launch Services refuses with an explicit version message | Yes |
| 14.0.0 | 13.0 | macOS 13 users are refused despite the binary being loadable | No |


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

---
### Requirement: The build pins the macOS deployment target

The build SHALL pin the macOS deployment target of the Wails-built application binary to the minimum supported version, from a single value defined in `Makefile`.

The pin is required because the Wails darwin backend forces external linking, where the deployment target — not the Go toolchain's default — determines `minos`. Wails injects its own `-mmacosx-version-min` value, which is lower than the supported minimum; omitting a project value MUST NOT be treated as a way to inherit a correct default, because an externally linked build with no deployment target falls back to the build machine's SDK version and produces a `minos` that varies per machine.

The pinned value SHALL equal the `LSMinimumSystemVersion` declared by the app bundle.

#### Scenario: The application binary is built through the project's build target

- **WHEN** the application is built through the project's build target on any build machine
- **THEN** the resulting binary's `minos` is the pinned minimum version
- **AND** it does not vary with the build machine's SDK version

#### Scenario: A binary is produced without external linking

- **WHEN** a shipped binary is linked internally and therefore carries the Go toolchain's default `minos`
- **THEN** that default is still required to equal the declared minimum version
- **AND** a mismatch is resolved by re-verifying the declaration against the binaries, not by leaving the two values different


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

---
### Requirement: App bundle declares the minimum system version

Both `build/darwin/Info.plist` and `build/darwin/Info.dev.plist` SHALL declare `LSMinimumSystemVersion` as the string `13.0.0`.

Both files are tracked in version control. Wails generates its template only when the file is absent, so these declarations MUST NOT be regenerated or overwritten by a build.

#### Scenario: Production and development bundles declare the same version

- **WHEN** either Info.plist file is read
- **THEN** the value associated with the `LSMinimumSystemVersion` key is `13.0.0`

#### Scenario: A macOS 12 user opens the built application

- **WHEN** a user on macOS 12 or earlier opens the built app bundle
- **THEN** Launch Services refuses to start it and states that a newer macOS version is required
- **AND** the failure is not deferred to dyld, where it would surface as an unexplained crash or generic "cannot be opened" message


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

---
### Requirement: A regression test guards the declared version

The project SHALL include an automated test that reads both Info.plist files and asserts that the declared `LSMinimumSystemVersion` equals the expected minimum version. The expected value SHALL be defined once within the test as a single source of truth.

The test SHALL parse the files as plists rather than matching raw text, and MUST tolerate the Go template directives embedded between XML elements in those files.

Failure to read a file, failure to parse it, or absence of the key SHALL fail the test. The test MUST NOT skip, warn, or silently pass under any of these conditions — silent drift is the exact failure this requirement exists to prevent.

#### Scenario: Declaration drifts from the expected value

- **WHEN** either Info.plist declares a `LSMinimumSystemVersion` other than the expected minimum version
- **THEN** the test fails
- **AND** the failure message names the file path, the actual value, and the expected value

#### Scenario: A declaration file cannot be read or parsed

- **WHEN** an Info.plist file is missing, unreadable, unparseable, or lacks the `LSMinimumSystemVersion` key
- **THEN** the test fails rather than skipping or passing


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

---
### Requirement: README documents the system requirement

`README.md` SHALL contain a system requirements section stating that macOS 13 Ventura or later is required. The section SHALL appear before the Installation section, so a user reads the constraint before running an installation command.

The section SHALL attribute the requirement to the Go toolchain's binary load requirement, so it is not read as an advisory value that can be worked around.

#### Scenario: A user reads the README before installing

- **WHEN** a user reads `README.md` from the top
- **THEN** the system requirement of macOS 13 Ventura or later is encountered before any installation instruction

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