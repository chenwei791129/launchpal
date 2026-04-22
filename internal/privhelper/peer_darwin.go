//go:build darwin

package privhelper

import (
	"fmt"
	"net"

	"golang.org/x/sys/unix"
)

// defaultPeerUID queries LOCAL_PEERCRED on a connected Unix socket to obtain
// the peer's effective UID. macOS does not support SO_PEERCRED; the platform
// idiom is xucred + getsockopt(SOL_LOCAL, LOCAL_PEERCRED).
func defaultPeerUID(c *net.UnixConn) (int, error) {
	raw, err := c.SyscallConn()
	if err != nil {
		return -1, fmt.Errorf("SyscallConn: %w", err)
	}
	var xu *unix.Xucred
	var opErr error
	if err := raw.Control(func(fd uintptr) {
		xu, opErr = unix.GetsockoptXucred(int(fd), unix.SOL_LOCAL, unix.LOCAL_PEERCRED)
	}); err != nil {
		return -1, fmt.Errorf("control: %w", err)
	}
	if opErr != nil {
		return -1, fmt.Errorf("LOCAL_PEERCRED: %w", opErr)
	}
	if xu == nil {
		return -1, fmt.Errorf("LOCAL_PEERCRED returned nil")
	}
	return int(xu.Uid), nil
}
