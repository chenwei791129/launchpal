//go:build darwin

package main

import (
	"fmt"
	"time"

	"golang.org/x/sys/unix"
)

// parentStartTime returns the start time of the process identified by pid,
// read from the kernel's kinfo_proc via sysctl(KERN_PROC_PID). The parent
// watchdog compares this against the value recorded at launch to tell the
// original LaunchPal parent apart from an unrelated process that later reused
// the same PID. A non-existent PID surfaces as an error.
func parentStartTime(pid int) (time.Time, error) {
	kp, err := unix.SysctlKinfoProc("kern.proc.pid", pid)
	if err != nil {
		return time.Time{}, fmt.Errorf("sysctl kern.proc.pid %d: %w", pid, err)
	}
	tv := kp.Proc.P_starttime
	return time.Unix(0, tv.Nano()), nil
}
