package launchctl

import (
	"errors"
	"os/user"
	"testing"
)

// withStubUserLookup temporarily replaces userLookupFn with a stub that
// returns entries from the given name→uid mapping.
func withStubUserLookup(t *testing.T, mapping map[string]string) {
	t.Helper()
	withCustomUserLookup(t, func(name string) (*user.User, error) {
		if uid, ok := mapping[name]; ok {
			return &user.User{Uid: uid, Username: name}, nil
		}
		return nil, errors.New("user not found")
	})
}

// withCustomUserLookup temporarily replaces userLookupFn with an arbitrary
// function. Use when the test needs to observe the name argument, count
// calls, or return stateful results.
func withCustomUserLookup(t *testing.T, fn func(name string) (*user.User, error)) {
	t.Helper()
	orig := userLookupFn
	userLookupFn = fn
	t.Cleanup(func() { userLookupFn = orig })
}

// withStubReadProcessTable temporarily overrides the process-table fetch.
func withStubReadProcessTable(t *testing.T, fn func() (ProcessTable, error)) {
	t.Helper()
	orig := readProcessTableFn
	readProcessTableFn = fn
	t.Cleanup(func() { readProcessTableFn = orig })
}

// defaultStubUsers returns a fresh copy of the stub user set used by most
// tests. Returning a fresh map prevents tests from accidentally mutating
// shared state visible to other tests.
func defaultStubUsers() map[string]string {
	return map[string]string{
		"root":           "0",
		"_www":           "70",
		"_mdnsresponder": "65",
	}
}

func TestDetectSystemServiceStatus_SingleMatchingPIDWithLaunchdParent(t *testing.T) {
	withStubUserLookup(t, defaultStubUsers())

	table := ProcessTable{
		12345: {UID: 0, PPID: 1, Args: "/usr/sbin/testd --flag"},
	}
	pd := plistData{Program: "/usr/sbin/testd"}
	status, pid, confidence := DetectSystemServiceStatus(pd, table, nil)
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
	withStubUserLookup(t, defaultStubUsers())

	pd := plistData{Program: "/usr/sbin/missingd"}
	status, pid, confidence := DetectSystemServiceStatus(pd, ProcessTable{}, nil)
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
	withStubUserLookup(t, defaultStubUsers())

	table := ProcessTable{
		300: {UID: 0, PPID: 1, Args: "/usr/sbin/ambiguousd"},
		100: {UID: 0, PPID: 1, Args: "/usr/sbin/ambiguousd"},
		200: {UID: 0, PPID: 1, Args: "/usr/sbin/ambiguousd"},
	}
	pd := plistData{Program: "/usr/sbin/ambiguousd"}
	status, pid, confidence := DetectSystemServiceStatus(pd, table, nil)
	if status != StatusRunning {
		t.Errorf("status = %q, want %q", status, StatusRunning)
	}
	if pid != 100 {
		t.Errorf("pid = %d, want 100 (lowest candidate)", pid)
	}
	if confidence != ConfidenceUnverified {
		t.Errorf("confidence = %q, want %q", confidence, ConfidenceUnverified)
	}
}

func TestDetectSystemServiceStatus_EmptyProgram(t *testing.T) {
	withStubUserLookup(t, defaultStubUsers())
	withStubReadProcessTable(t, func() (ProcessTable, error) {
		t.Fatal("process table should not be fetched for empty program")
		return nil, nil
	})

	pd := plistData{}
	status, pid, confidence := DetectSystemServiceStatus(pd, nil, nil)
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
	withCustomUserLookup(t, func(name string) (*user.User, error) {
		capturedUser = name
		return &user.User{Uid: "0", Username: name}, nil
	})

	pd := plistData{Program: "/usr/sbin/testd"}
	DetectSystemServiceStatus(pd, ProcessTable{}, nil)
	if capturedUser != "root" {
		t.Errorf("default UserName = %q, want %q", capturedUser, "root")
	}
}

func TestDetectSystemServiceStatus_NonRootUserName(t *testing.T) {
	withStubUserLookup(t, defaultStubUsers())

	// A root-owned process should not match when UserName is _www.
	table := ProcessTable{
		555: {UID: 0, PPID: 1, Args: "/usr/sbin/httpd -D FOREGROUND"},
	}
	pd := plistData{Program: "/usr/sbin/httpd", UserName: "_www"}
	status, _, _ := DetectSystemServiceStatus(pd, table, nil)
	if status != StatusStopped {
		t.Errorf("status = %q, want %q (root-owned process must not match _www)", status, StatusStopped)
	}

	// A _www-owned process should match.
	table[555] = processInfo{UID: 70, PPID: 1, Args: "/usr/sbin/httpd -D FOREGROUND"}
	status2, pid, confidence := DetectSystemServiceStatus(pd, table, nil)
	if status2 != StatusRunning || pid != 555 || confidence != ConfidenceVerified {
		t.Errorf("status/pid/confidence = %q/%d/%q, want Running/555/verified", status2, pid, confidence)
	}
}

func TestDetectSystemServiceStatus_NonLaunchdParentFilteredOut(t *testing.T) {
	withStubUserLookup(t, defaultStubUsers())

	// Parent is a shell, not launchd.
	table := ProcessTable{
		555: {UID: 0, PPID: 9999, Args: "/usr/sbin/testd"},
	}
	pd := plistData{Program: "/usr/sbin/testd"}
	status, pid, confidence := DetectSystemServiceStatus(pd, table, nil)
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
	withStubUserLookup(t, defaultStubUsers())

	table := ProcessTable{
		100: {UID: 0, PPID: 9999, Args: "/usr/sbin/mixedd"}, // shell child
		200: {UID: 0, PPID: 1, Args: "/usr/sbin/mixedd"},    // launchd child
		300: {UID: 0, PPID: 500, Args: "/usr/sbin/mixedd"},  // other parent
	}
	pd := plistData{Program: "/usr/sbin/mixedd"}
	status, pid, confidence := DetectSystemServiceStatus(pd, table, nil)
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
			// The shell skip path must not fetch the process table or look up
			// the uid — fail fast if either is invoked.
			withCustomUserLookup(t, func(string) (*user.User, error) {
				t.Fatal("user.Lookup should not be called for shell skip")
				return nil, nil
			})

			pd := plistData{Program: shell}
			status, pid, confidence := DetectSystemServiceStatus(pd, nil, nil)
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
	withStubUserLookup(t, defaultStubUsers())

	table := ProcessTable{
		42: {UID: 0, PPID: 1, Args: "/usr/bin/somehelper --flag"},
	}
	pd := plistData{ProgramArguments: []string{"/usr/bin/somehelper", "--flag"}}
	status, pid, _ := DetectSystemServiceStatus(pd, table, nil)
	if status != StatusRunning || pid != 42 {
		t.Errorf("status/pid = %q/%d, want Running/42 (fallback to ProgramArguments[0])", status, pid)
	}
}

func TestDetectSystemServiceStatus_ArgsSubstringMatch(t *testing.T) {
	withStubUserLookup(t, defaultStubUsers())

	// `ps args=` output includes arguments after the program path; Contains
	// matches as long as the program path appears anywhere in the argv string.
	table := ProcessTable{
		77: {UID: 0, PPID: 1, Args: "/usr/sbin/real --config /etc/x.conf"},
	}
	pd := plistData{Program: "/usr/sbin/real"}
	status, pid, _ := DetectSystemServiceStatus(pd, table, nil)
	if status != StatusRunning || pid != 77 {
		t.Errorf("status/pid = %q/%d, want Running/77", status, pid)
	}
}

// TestDetectSystemServiceStatus_NilTableFetchesLazily verifies the Get path:
// when the caller does not pre-fetch the table, DetectSystemServiceStatus
// fetches it on demand.
func TestDetectSystemServiceStatus_NilTableFetchesLazily(t *testing.T) {
	withStubUserLookup(t, defaultStubUsers())

	called := false
	withStubReadProcessTable(t, func() (ProcessTable, error) {
		called = true
		return ProcessTable{42: {UID: 0, PPID: 1, Args: "/usr/sbin/lazyd"}}, nil
	})

	pd := plistData{Program: "/usr/sbin/lazyd"}
	status, pid, _ := DetectSystemServiceStatus(pd, nil, nil)
	if !called {
		t.Fatal("readProcessTableFn was not called when table == nil")
	}
	if status != StatusRunning || pid != 42 {
		t.Errorf("status/pid = %q/%d, want Running/42", status, pid)
	}
}

// TestDetectSystemServiceStatus_ProcessTableFetchFailureDegrades verifies
// that a fetch failure degrades to StatusStopped/unverified rather than a
// confident Stopped/verified result.
func TestDetectSystemServiceStatus_ProcessTableFetchFailureDegrades(t *testing.T) {
	withStubUserLookup(t, defaultStubUsers())
	withStubReadProcessTable(t, func() (ProcessTable, error) {
		return nil, errors.New("ps failed")
	})

	pd := plistData{Program: "/usr/sbin/degradable"}
	status, pid, confidence := DetectSystemServiceStatus(pd, nil, nil)
	if status != StatusStopped {
		t.Errorf("status = %q, want %q (cannot confidently confirm Running without a table)", status, StatusStopped)
	}
	if pid != 0 {
		t.Errorf("pid = %d, want 0", pid)
	}
	if confidence != ConfidenceUnverified {
		t.Errorf("confidence = %q, want %q (degraded, not verified)", confidence, ConfidenceUnverified)
	}
}

// TestDetectSystemServiceStatus_UIDLookupFailureDegrades verifies that a
// failure to resolve UserName → uid degrades to Stopped/Unverified.
func TestDetectSystemServiceStatus_UIDLookupFailureDegrades(t *testing.T) {
	withStubUserLookup(t, map[string]string{}) // no users → every lookup fails

	pd := plistData{Program: "/usr/sbin/unresolvable", UserName: "nobodyghost"}
	status, _, confidence := DetectSystemServiceStatus(pd, ProcessTable{}, nil)
	if status != StatusStopped {
		t.Errorf("status = %q, want %q", status, StatusStopped)
	}
	if confidence != ConfidenceUnverified {
		t.Errorf("confidence = %q, want %q", confidence, ConfidenceUnverified)
	}
}

// TestDetectSystemServiceStatus_UIDCacheAvoidsRepeatedLookups verifies that
// consecutive detections for the same UserName only call user.Lookup once
// when a shared uidCache is provided.
func TestDetectSystemServiceStatus_UIDCacheAvoidsRepeatedLookups(t *testing.T) {
	lookupCount := 0
	withCustomUserLookup(t, func(name string) (*user.User, error) {
		lookupCount++
		return &user.User{Uid: "0", Username: name}, nil
	})

	table := ProcessTable{}
	uidCache := map[string]int{}
	pd := plistData{Program: "/usr/sbin/x"}

	for range 5 {
		DetectSystemServiceStatus(pd, table, uidCache)
	}
	if lookupCount != 1 {
		t.Errorf("user.Lookup called %d times, want 1 (should be cached)", lookupCount)
	}
}

func TestParseProcessTable_NormalLine(t *testing.T) {
	raw := "    0    1    0 /sbin/launchd\n    0  100    1 /usr/sbin/testd --flag\n"
	table := parseProcessTable(raw)
	if len(table) != 2 {
		t.Fatalf("got %d rows, want 2", len(table))
	}
	info1, ok := table[1]
	if !ok {
		t.Fatal("pid 1 missing")
	}
	if info1.UID != 0 || info1.PPID != 0 || info1.Args != "/sbin/launchd" {
		t.Errorf("pid 1 info = %+v, want {0, 0, /sbin/launchd}", info1)
	}
	info100, ok := table[100]
	if !ok {
		t.Fatal("pid 100 missing")
	}
	if info100.UID != 0 || info100.PPID != 1 || info100.Args != "/usr/sbin/testd --flag" {
		t.Errorf("pid 100 info = %+v", info100)
	}
}

func TestParseProcessTable_ArgsWithMultipleSpaces(t *testing.T) {
	raw := "  70  200    1 /usr/sbin/httpd  -D FOREGROUND   --multi\n"
	table := parseProcessTable(raw)
	info, ok := table[200]
	if !ok {
		t.Fatal("pid 200 missing")
	}
	// Args must preserve the inner whitespace as-is.
	if info.Args != "/usr/sbin/httpd  -D FOREGROUND   --multi" {
		t.Errorf("Args = %q (internal whitespace collapsed)", info.Args)
	}
}

func TestParseProcessTable_BlankLinesSkipped(t *testing.T) {
	raw := "\n   0   1   0 /sbin/launchd\n\n\n"
	table := parseProcessTable(raw)
	if len(table) != 1 {
		t.Errorf("got %d rows, want 1 (blank lines should be skipped)", len(table))
	}
}

func TestParseProcessTable_MalformedLineSkipped(t *testing.T) {
	raw := "not_a_number 1 0 /sbin/launchd\n0 100 1 /usr/sbin/testd\n"
	table := parseProcessTable(raw)
	if _, ok := table[1]; ok {
		t.Error("pid 1 should have been skipped (uid is not numeric)")
	}
	if _, ok := table[100]; !ok {
		t.Error("pid 100 should be present (line is valid)")
	}
}

func TestParseProcessTable_EmptyOutput(t *testing.T) {
	table := parseProcessTable("")
	if len(table) != 0 {
		t.Errorf("empty output produced %d rows, want 0", len(table))
	}
}
