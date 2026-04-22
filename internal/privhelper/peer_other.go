//go:build !darwin

package privhelper

import (
	"errors"
	"net"
)

// defaultPeerUID is a stub on non-darwin platforms so the package compiles
// for CI / cross-dev; LaunchPal only ships on macOS.
func defaultPeerUID(_ *net.UnixConn) (int, error) {
	return -1, errors.New("LOCAL_PEERCRED is only supported on darwin")
}
