//go:build !darwin

package launchctl

// nofollowFlag is zero on non-darwin platforms, disabling O_NOFOLLOW at the
// OpenFile call site. LaunchPal ships for macOS only; the non-darwin build
// exists solely so unit tests compile in portable CI. Symlink protection is
// silently absent there — do not run the binary on other platforms.
const nofollowFlag = 0
