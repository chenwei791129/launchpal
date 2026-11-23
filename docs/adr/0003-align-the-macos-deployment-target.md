# Align the macOS deployment target with the Go runtime

The minimum macOS version declared by the app bundle must match the actual minimum of both shipped binaries. Adopt macOS 13 rather than lowering the binary version marker to keep macOS 12 nominally supported: a linker override does not make newer runtime APIs available, and the privileged helper must not run in an unsupported configuration.

The Go linker supplies the helper's minimum. The Wails app is externally linked, so pin its deployment target explicitly rather than relying on Wails defaults or the build machine's SDK. See [Minimum macOS Version](../../.claude/CLAUDE.md#minimum-macos-version) for the checks to run when changing Go or Wails.
