//go:build darwin

package launchctl

import "syscall"

// nofollowFlag is the O_NOFOLLOW flag on darwin. Splitting the value across a
// pair of build-tagged files mirrors the privhelper package so log
// truncation refuses to dereference a planted symlink on macOS while still
// letting the package compile on portable CI.
const nofollowFlag = syscall.O_NOFOLLOW
