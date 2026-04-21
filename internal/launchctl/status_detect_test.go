package launchctl

import (
	"errors"
	"testing"
)

// withStubPgrep temporarily overrides the pgrep function for testing.
func withStubPgrep(t *testing.T, pgrep func(user, program string) ([]int, error)) {
	t.Helper()
	orig := pgrepCandidatesFn
	pgrepCandidatesFn = pgrep
	t.Cleanup(func() { pgrepCandidatesFn = orig })
}

// withStubReadAllPPIDs temporarily overrides the ppid-table fetch.
func withStubReadAllPPIDs(t *testing.T, fn func() (map[int]int, error)) {
	t.Helper()
	orig := readAllPPIDsFn
	readAllPPIDsFn = fn
	t.Cleanup(func() { readAllPPIDsFn = orig })
}

func TestDetectSystemServiceStatus_SingleMatchingPIDWithLaunchdParent(t *testing.T) {
	withStubPgrep(t, func(user, program string) ([]int, error) {
		if user != "root" {
			t.Errorf("pgrep user = %q, want %q", user, "root")
		}
		if program != "/usr/sbin/testd" {
			t.Errorf("pgrep program = %q, want %q", program, "/usr/sbin/testd")
		}
		return []int{12345}, nil
	})

	pd := plistData{Program: "/usr/sbin/testd"}
	status, pid, confidence := DetectSystemServiceStatus(pd, map[int]int{12345: 1})
	if status != StatusRunning {
		t.Errorf("status = %q, want %q", status, StatusRunning)
	}
	if pid != 12345 {
		t.Errorf("pid = %d, want 12345", pid)
	}
	if confidence != ConfidenceVerified {
		t.Errorf("confidence = %q, want %q", confidence, ConfidenceVerified)
	}
}

func TestDetectSystemServiceStatus_NoMatchingProcess(t *testing.T) {
	withStubPgrep(t, func(user, program string) ([]int, error) { return nil, nil })

	pd := plistData{Program: "/usr/sbin/missingd"}
	status, pid, confidence := DetectSystemServiceStatus(pd, map[int]int{})
	if status != StatusStopped {
		t.Errorf("status = %q, want %q", status, StatusStopped)
	}
	if pid != 0 {
		t.Errorf("pid = %d, want 0", pid)
	}
	if confidence != ConfidenceVerified {
		t.Errorf("confidence = %q, want %q", confidence, ConfidenceVerified)
	}
}

func TestDetectSystemServiceStatus_MultipleCandidatesWithLaunchdParent(t *testing.T) {
	withStubPgrep(t, func(user, program string) ([]int, error) { return []int{100, 200, 300}, nil })

	pd := plistData{Program: "/usr/sbin/ambiguousd"}
	status, pid, confidence := DetectSystemServiceStatus(pd, map[int]int{100: 1, 200: 1, 300: 1})
	if status != StatusRunning {
		t.Errorf("status = %q, want %q", status, StatusRunning)
	}
	if pid != 100 {
		t.Errorf("pid = %d, want 100 (first candidate)", pid)
	}
	if confidence != ConfidenceUnverified {
		t.Errorf("confidence = %q, want %q", confidence, ConfidenceUnverified)
	}
}

func TestDetectSystemServiceStatus_EmptyProgram(t *testing.T) {
	withStubPgrep(t, func(user, program string) ([]int, error) {
		t.Fatal("pgrep should not be called for empty program")
		return nil, nil
	})

	pd := plistData{}
	status, pid, confidence := DetectSystemServiceStatus(pd, map[int]int{})
	if status != StatusUnknown {
		t.Errorf("status = %q, want %q", status, StatusUnknown)
	}
	if pid != 0 {
		t.Errorf("pid = %d, want 0", pid)
	}
	if confidence != ConfidenceUnverified {
		t.Errorf("confidence = %q, want %q", confidence, ConfidenceUnverified)
	}
}

func TestDetectSystemServiceStatus_UserNameDefaultsToRoot(t *testing.T) {
	var capturedUser string
	withStubPgrep(t, func(user, program string) ([]int, error) {
		capturedUser = user
		return nil, nil
	})

	pd := plistData{Program: "/usr/sbin/testd"}
	DetectSystemServiceStatus(pd, map[int]int{})
	if capturedUser != "root" {
		t.Errorf("default UserName = %q, want %q", capturedUser, "root")
	}
}

func TestDetectSystemServiceStatus_NonRootUserName(t *testing.T) {
	var capturedUser string
	withStubPgrep(t, func(user, program string) ([]int, error) {
		capturedUser = user
		return nil, nil
	})

	pd := plistData{Program: "/usr/sbin/httpd", UserName: "_www"}
	DetectSystemServiceStatus(pd, map[int]int{})
	if capturedUser != "_www" {
		t.Errorf("UserName = %q, want %q", capturedUser, "_www")
	}
}

func TestDetectSystemServiceStatus_NonLaunchdParentFilteredOut(t *testing.T) {
	withStubPgrep(t, func(user, program string) ([]int, error) { return []int{555}, nil })

	pd := plistData{Program: "/usr/sbin/testd"}
	// Parent is a shell, not launchd.
	status, pid, confidence := DetectSystemServiceStatus(pd, map[int]int{555: 9999})
	if status != StatusStopped {
		t.Errorf("status = %q, want %q", status, StatusStopped)
	}
	if pid != 0 {
		t.Errorf("pid = %d, want 0", pid)
	}
	if confidence != ConfidenceVerified {
		t.Errorf("confidence = %q, want %q", confidence, ConfidenceVerified)
	}
}

func TestDetectSystemServiceStatus_MixedPPIDKeepsOnlyLaunchdChildren(t *testing.T) {
	withStubPgrep(t, func(user, program string) ([]int, error) { return []int{100, 200, 300}, nil })

	// 100: shell child (ppid=9999), 200: launchd (ppid=1), 300: other (ppid=500)
	pd := plistData{Program: "/usr/sbin/mixedd"}
	status, pid, confidence := DetectSystemServiceStatus(pd, map[int]int{100: 9999, 200: 1, 300: 500})
	if status != StatusRunning {
		t.Errorf("status = %q, want %q", status, StatusRunning)
	}
	if pid != 200 {
		t.Errorf("pid = %d, want 200 (only launchd child)", pid)
	}
	if confidence != ConfidenceVerified {
		t.Errorf("confidence = %q, want %q", confidence, ConfidenceVerified)
	}
}

func TestDetectSystemServiceStatus_ShellProgramSkipped(t *testing.T) {
	for _, shell := range []string{"/bin/bash", "/bin/sh", "/bin/zsh", "/usr/bin/bash", "/usr/bin/sh", "/usr/bin/zsh"} {
		t.Run(shell, func(t *testing.T) {
			withStubPgrep(t, func(user, program string) ([]int, error) {
				t.Fatalf("pgrep should not be called for shell %q", shell)
				return nil, nil
			})

			pd := plistData{Program: shell}
			status, pid, confidence := DetectSystemServiceStatus(pd, map[int]int{})
			if status != StatusLoaded {
				t.Errorf("status = %q, want %q", status, StatusLoaded)
			}
			if pid != 0 {
				t.Errorf("pid = %d, want 0", pid)
			}
			if confidence != ConfidenceVerified {
				t.Errorf("confidence = %q, want %q", confidence, ConfidenceVerified)
			}
		})
	}
}

func TestDetectSystemServiceStatus_FallbackToProgramArgumentsFirst(t *testing.T) {
	var capturedProgram string
	withStubPgrep(t, func(user, program string) ([]int, error) {
		capturedProgram = program
		return nil, nil
	})

	pd := plistData{ProgramArguments: []string{"/usr/bin/somehelper", "--flag"}}
	DetectSystemServiceStatus(pd, map[int]int{})
	if capturedProgram != "/usr/bin/somehelper" {
		t.Errorf("program = %q, want %q (ProgramArguments[0])", capturedProgram, "/usr/bin/somehelper")
	}
}

// TestDetectSystemServiceStatus_NilPPIDTableFetchesLazily verifies that when
// the caller does not pre-fetch the ppid table, DetectSystemServiceStatus
// fetches it on demand via readAllPPIDsFn.
func TestDetectSystemServiceStatus_NilPPIDTableFetchesLazily(t *testing.T) {
	withStubPgrep(t, func(user, program string) ([]int, error) { return []int{42}, nil })
	called := false
	withStubReadAllPPIDs(t, func() (map[int]int, error) {
		called = true
		return map[int]int{42: 1}, nil
	})

	pd := plistData{Program: "/usr/sbin/lazyd"}
	status, pid, _ := DetectSystemServiceStatus(pd, nil)
	if !called {
		t.Fatal("readAllPPIDsFn was not called when ppidTable == nil")
	}
	if status != StatusRunning || pid != 42 {
		t.Errorf("status/pid = %q/%d, want Running/42", status, pid)
	}
}

// TestDetectSystemServiceStatus_PPIDFetchFailureDegrades verifies that when
// pgrep finds candidates but the ppid table cannot be read, the result
// degrades to Running/Unverified rather than producing a confident false
// negative (Stopped/Verified).
func TestDetectSystemServiceStatus_PPIDFetchFailureDegrades(t *testing.T) {
	withStubPgrep(t, func(user, program string) ([]int, error) { return []int{777}, nil })
	withStubReadAllPPIDs(t, func() (map[int]int, error) {
		return nil, errFake
	})

	pd := plistData{Program: "/usr/sbin/degradable"}
	status, pid, confidence := DetectSystemServiceStatus(pd, nil)
	if status != StatusRunning {
		t.Errorf("status = %q, want %q (degraded, not false Stopped)", status, StatusRunning)
	}
	if pid != 777 {
		t.Errorf("pid = %d, want 777 (first candidate)", pid)
	}
	if confidence != ConfidenceUnverified {
		t.Errorf("confidence = %q, want %q", confidence, ConfidenceUnverified)
	}
}

var errFake = errors.New("fake error")
