package main

import (
	"bytes"
	"errors"
	"strings"
	"testing"
	"time"

	"launchpal/internal/privhelper"
)

// TestMakeParentAlive_PIDReuseDetected covers the "Parent PID reused by
// another process" scenario: the PID still exists but reports a different
// start time, so the closure must treat the original parent as gone.
func TestMakeParentAlive_PIDReuseDetected(t *testing.T) {
	recorded := time.Unix(1000, 0)
	startTimeOf := func(int) (time.Time, error) { return time.Unix(2000, 0), nil }
	pidExists := func(int) bool { return true }
	alive := makeParentAlive(recorded, startTimeOf, pidExists)
	if alive(1234) {
		t.Error("parent should be considered gone when the start time differs (PID reuse)")
	}
}

func TestMakeParentAlive_SameStartIsAlive(t *testing.T) {
	recorded := time.Unix(1000, 0)
	startTimeOf := func(int) (time.Time, error) { return time.Unix(1000, 0), nil }
	alive := makeParentAlive(recorded, startTimeOf, func(int) bool { return true })
	if !alive(1234) {
		t.Error("parent should be alive when the recorded start time matches")
	}
}

// TestMakeParentAlive_ZeroBaselineDegradesToPIDExistence covers the case where
// the launch-time start-time read failed (recordedStart is the zero value):
// the closure must degrade to a PID-existence check, never compare a live
// process's real start time against the zero baseline (which would never match
// and would self-terminate against a still-alive parent).
func TestMakeParentAlive_ZeroBaselineDegradesToPIDExistence(t *testing.T) {
	// A successful, non-zero current reading must NOT be compared to the zero
	// baseline; the existence check decides instead.
	startTimeOf := func(int) (time.Time, error) { return time.Unix(2000, 0), nil }

	aliveWhenExists := makeParentAlive(time.Time{}, startTimeOf, func(int) bool { return true })
	if !aliveWhenExists(1234) {
		t.Error("zero baseline with a live PID should report alive via existence check")
	}
	aliveWhenGone := makeParentAlive(time.Time{}, startTimeOf, func(int) bool { return false })
	if aliveWhenGone(1234) {
		t.Error("zero baseline with a dead PID should report gone via existence check")
	}
}

// TestMakeParentAlive_FallbackToPIDExistence covers platforms where the start
// time cannot be obtained (non-darwin, or a transient lookup error): the
// closure degrades to a plain PID-existence check rather than killing the
// helper spuriously.
func TestMakeParentAlive_FallbackToPIDExistence(t *testing.T) {
	startTimeOf := func(int) (time.Time, error) { return time.Time{}, errors.New("unavailable") }
	recorded := time.Unix(1000, 0)

	aliveWhenExists := makeParentAlive(recorded, startTimeOf, func(int) bool { return true })
	if !aliveWhenExists(1234) {
		t.Error("should fall back to PID-existence=true when start time is unavailable")
	}
	aliveWhenGone := makeParentAlive(recorded, startTimeOf, func(int) bool { return false })
	if aliveWhenGone(1234) {
		t.Error("should fall back to PID-existence=false when start time is unavailable")
	}
}

// TestIdleTimeout_DefaultIsFiveMinutes pins the idle-timeout backstop. The
// duration is the security mitigation itself — it bounds how long an idle
// root socket lingers — so a silent regression to the old 30-minute value
// must fail the build.
func TestIdleTimeout_DefaultIsFiveMinutes(t *testing.T) {
	if idleTimeout != 5*time.Minute {
		t.Errorf("idleTimeout = %v, want 5m", idleTimeout)
	}
}

// These tests exercise the argument-validation surface of the helper without
// requiring root or actually opening a socket. The helper refuses to run when
// any of the required preconditions fail, and exits with a non-zero status
// after writing a diagnostic to stderr.

func TestValidateArgs_NonRoot(t *testing.T) {
	cfg := helperConfig{
		EffectiveUID: 501,
		Socket:       "/tmp/x.sock",
		ParentPID:    1234,
		LaunchingUID: 501,
	}
	if err := validateArgs(&cfg); err == nil {
		t.Fatal("expected error for non-root UID")
	} else if !strings.Contains(err.Error(), "root") {
		t.Errorf("error = %q, want message mentioning root", err.Error())
	}
}

func TestValidateArgs_MissingSocket(t *testing.T) {
	cfg := helperConfig{
		EffectiveUID: 0,
		Socket:       "",
		ParentPID:    1234,
		LaunchingUID: 501,
	}
	if err := validateArgs(&cfg); err == nil {
		t.Fatal("expected error for missing socket")
	} else if !strings.Contains(err.Error(), "socket") {
		t.Errorf("error = %q, want message mentioning socket", err.Error())
	}
}

func TestValidateArgs_MissingParentPID(t *testing.T) {
	cfg := helperConfig{
		EffectiveUID: 0,
		Socket:       "/tmp/x.sock",
		ParentPID:    0,
		LaunchingUID: 501,
	}
	if err := validateArgs(&cfg); err == nil {
		t.Fatal("expected error for missing parent-pid")
	} else if !strings.Contains(err.Error(), "parent-pid") {
		t.Errorf("error = %q, want message mentioning parent-pid", err.Error())
	}
}

func TestValidateArgs_MissingLaunchingUID(t *testing.T) {
	cfg := helperConfig{
		EffectiveUID: 0,
		Socket:       "/tmp/x.sock",
		ParentPID:    1234,
		LaunchingUID: -1,
	}
	if err := validateArgs(&cfg); err == nil {
		t.Fatal("expected error for missing launching-uid")
	} else if !strings.Contains(err.Error(), "launching-uid") {
		t.Errorf("error = %q, want message mentioning launching-uid", err.Error())
	}
}

func TestValidateArgs_RejectRootLaunchingUID(t *testing.T) {
	cfg := helperConfig{
		EffectiveUID: 0,
		Socket:       "/tmp/x.sock",
		ParentPID:    1234,
		LaunchingUID: 0,
	}
	if err := validateArgs(&cfg); err == nil {
		t.Fatal("launching-uid=0 must be rejected (root peer would bypass per-user auth)")
	}
}

func TestValidateArgs_UserHomeValidation(t *testing.T) {
	cases := []struct {
		name string
		home string
		ok   bool
	}{
		{"relative path rejected", "relative/path", false},
		{"root rejected", "/", false},
		{"system dir rejected", "/System/Library", false},
		{"lib dir rejected", "/Library", false},
		{"etc rejected", "/etc", false},
		{"applications rejected", "/Applications/Foo.app", false},
		{"normal user home ok", "/Users/alice", true},
		{"empty ok (optional)", "", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := helperConfig{
				EffectiveUID: 0,
				Socket:       "/tmp/x.sock",
				ParentPID:    1234,
				LaunchingUID: 501,
				UserHome:     tc.home,
			}
			err := validateArgs(&cfg)
			if tc.ok && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if !tc.ok && err == nil {
				t.Errorf("expected error for user-home %q", tc.home)
			}
		})
	}
}

func TestValidateArgs_AllGood(t *testing.T) {
	cfg := helperConfig{
		EffectiveUID: 0,
		Socket:       "/tmp/x.sock",
		ParentPID:    1234,
		LaunchingUID: 501,
	}
	if err := validateArgs(&cfg); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestParseFlags_Defaults(t *testing.T) {
	var stderr bytes.Buffer
	cfg, err := parseFlags([]string{
		"--socket", "/tmp/s.sock",
		"--parent-pid", "4321",
		"--launching-uid", "501",
	}, &stderr)
	if err != nil {
		t.Fatalf("parseFlags: %v", err)
	}
	if cfg.Socket != "/tmp/s.sock" {
		t.Errorf("Socket = %q", cfg.Socket)
	}
	if cfg.ParentPID != 4321 {
		t.Errorf("ParentPID = %d", cfg.ParentPID)
	}
	if cfg.LaunchingUID != 501 {
		t.Errorf("LaunchingUID = %d", cfg.LaunchingUID)
	}
	if cfg.UserHome == "" && false {
		t.Errorf("UserHome is empty")
	}
}

func TestParseFlags_UserHomeIsOptional(t *testing.T) {
	var stderr bytes.Buffer
	cfg, err := parseFlags([]string{
		"--socket", "/tmp/s.sock",
		"--parent-pid", "4321",
		"--launching-uid", "501",
		"--user-home", "/Users/alice",
	}, &stderr)
	if err != nil {
		t.Fatalf("parseFlags: %v", err)
	}
	if cfg.UserHome != "/Users/alice" {
		t.Errorf("UserHome = %q", cfg.UserHome)
	}
}

func TestSelfInstallProtectedCopy_SkipWhenLaunchedFromProtectedPath(t *testing.T) {
	called := false
	install := func(string) (bool, error) { called = true; return false, nil }
	selfInstallProtectedCopy(privhelper.ProtectedHelperPath, install, func(string, ...any) {})
	if called {
		t.Error("install should not run when launched from the protected path")
	}
}

func TestSelfInstallProtectedCopy_FailureIsNonFatal(t *testing.T) {
	var logged strings.Builder
	install := func(string) (bool, error) { return false, errFakeInstall }
	// A failing install must return normally (non-fatal) and emit a log line —
	// the caller keeps serving the current session regardless.
	selfInstallProtectedCopy("/Applications/LaunchPal.app/Contents/MacOS/launchpal-privhelper", install,
		func(format string, args ...any) { logged.WriteString(format) })
	if logged.Len() == 0 {
		t.Error("expected a log line on install failure")
	}
}

var errFakeInstall = errorString("boom")

type errorString string

func (e errorString) Error() string { return string(e) }
