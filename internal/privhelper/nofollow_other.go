//go:build !darwin

package privhelper

// syscallNoFollow is zero on non-darwin platforms, intentionally disabling
// O_NOFOLLOW at the OpenFile call site. LaunchPal ships for macOS only; the
// non-darwin build exists solely so unit tests compile in portable CI.
// Running the helper on another OS would silently lose symlink protection
// — do not deploy this binary outside macOS.
const syscallNoFollow = 0
