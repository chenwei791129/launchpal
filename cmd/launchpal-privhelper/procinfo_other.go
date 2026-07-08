//go:build !darwin

package main

import (
	"errors"
	"time"
)

// parentStartTime is unavailable off darwin. Returning an error makes the
// watchdog degrade to a plain PID-existence check (see makeParentAlive),
// preserving the pre-existing behavior. LaunchPal ships only on macOS; this
// exists solely so cross-platform builds keep compiling.
func parentStartTime(int) (time.Time, error) {
	return time.Time{}, errors.New("parent start time not available on this platform")
}
