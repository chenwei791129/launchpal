//go:build darwin

package privhelper

import "syscall"

// syscallNoFollow is the O_NOFOLLOW flag on darwin. Isolating the value in a
// build-tagged file keeps the cross-build story clean even though LaunchPal
// only ships for macOS.
const syscallNoFollow = syscall.O_NOFOLLOW
