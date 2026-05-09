package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"launchpal/internal/launchctl"
	"launchpal/internal/settings"
)

func TestGetVersion_Default(t *testing.T) {
	app := NewApp()
	got := app.GetVersion()
	if got != "dev" {
		t.Errorf("GetVersion() = %q, want %q", got, "dev")
	}
}

func TestGetVersion_Injected(t *testing.T) {
	app := NewAppWithVersion("v1.6.0")
	got := app.GetVersion()
	if got != "v1.6.0" {
		t.Errorf("GetVersion() = %q, want %q", got, "v1.6.0")
	}
}

func TestRevealInFinder_EmptyPath(t *testing.T) {
	app := NewApp()
	err := app.RevealInFinder("")
	if err == nil {
		t.Error("RevealInFinder(\"\") should return an error for empty path")
	}
}

func TestRevealInFinder_ValidPath(t *testing.T) {
	if os.Getenv("CI") != "" {
		t.Skip("skipping GUI test in CI environment")
	}

	// Create a temp file to use as the target path
	f, err := os.CreateTemp("", "launchpal-test-*.txt")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer func() { _ = os.Remove(f.Name()) }()
	_ = f.Close()

	app := NewApp()
	if err := app.RevealInFinder(f.Name()); err != nil {
		t.Errorf("RevealInFinder(%q) returned unexpected error: %v", f.Name(), err)
	}
}

func TestRevealInFinder_NonexistentPath(t *testing.T) {
	app := NewApp()
	err := app.RevealInFinder("/nonexistent/launchpal-test-path.plist")
	if err == nil {
		t.Error("RevealInFinder with nonexistent path should return an error")
	}
}

func TestClearSystemLogs_InvalidServiceType(t *testing.T) {
	app := NewApp()
	err := app.ClearSystemLogs("com.x", "bogus", launchctl.LogTypeStdout)
	if err == nil {
		t.Fatal("expected error for invalid serviceType")
	}
	if !strings.Contains(err.Error(), "invalid service type") {
		t.Errorf("err = %q", err.Error())
	}
}

func TestClearSystemLogs_AppleSystemRejected(t *testing.T) {
	app := NewApp()
	err := app.ClearSystemLogs("com.apple.x", launchctl.ServiceTypeAppleSystem, launchctl.LogTypeStdout)
	if err == nil {
		t.Fatal("expected error for apple-system")
	}
	if !strings.Contains(err.Error(), "apple-system") && !strings.Contains(err.Error(), "read-only") {
		t.Errorf("err = %q, want apple-system/read-only mention", err.Error())
	}
}

func TestGetLogClearStatus_InvalidServiceType(t *testing.T) {
	app := NewApp()
	_, err := app.GetLogClearStatus("com.x", "bogus", launchctl.LogTypeStdout)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestClearLogs_UserDispatchPropagatesError(t *testing.T) {
	app := NewApp()
	if err := app.ClearLogs("com.test.never-exists-launchpal", launchctl.LogTypeStdout); err == nil {
		t.Error("expected error for missing service via ClearLogs binding")
	}
}

func TestGetSettings_FirstRunReturnsDefaults(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	app := NewApp()
	got, err := app.GetSettings()
	if err != nil {
		t.Fatalf("GetSettings: %v", err)
	}
	if got != settings.Default() {
		t.Errorf("GetSettings = %+v, want Default %+v", got, settings.Default())
	}
}

func TestUpdateSettings_ValidationFailureDoesNotWrite(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	app := NewApp()
	bad := settings.Settings{UserLogDir: "~/Library/Logs", SystemLogDir: "/etc/foo"}
	err := app.UpdateSettings(bad)
	if err == nil {
		t.Fatal("UpdateSettings: expected validation error, got nil")
	}
	if !strings.Contains(err.Error(), "systemLogDir") {
		t.Errorf("error %q does not mention systemLogDir", err)
	}
	if _, statErr := os.Stat(filepath.Join(home, ".launchpal", "settings.json")); !os.IsNotExist(statErr) {
		t.Errorf("settings.json created despite validation failure: err=%v", statErr)
	}
}

func TestUpdateSettings_SuccessPersists(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	app := NewApp()
	want := settings.Settings{UserLogDir: "/tmp/userlogs", SystemLogDir: "/Library/Logs/launchpal"}
	if err := app.UpdateSettings(want); err != nil {
		t.Fatalf("UpdateSettings: %v", err)
	}
	got, err := app.GetSettings()
	if err != nil {
		t.Fatalf("GetSettings: %v", err)
	}
	if got != want {
		t.Errorf("after UpdateSettings: GetSettings = %+v, want %+v", got, want)
	}
}

func TestIsSystemDaemonPath(t *testing.T) {
	cases := []struct {
		path string
		want bool
	}{
		{"/Library/LaunchDaemons/com.example.plist", true},
		{"/Library/LaunchDaemons/nested/com.example.plist", true},
		{"/Library/LaunchDaemons//trailing.plist", true}, // filepath.Clean normalizes
		{"/Library/LaunchAgents/com.example.plist", false},
		{"/System/Library/LaunchDaemons/com.apple.plist", false},
		{"/Users/foo/Library/LaunchAgents/com.example.plist", false},
		{"/Library/LaunchDaemonsX/sneaky.plist", false}, // must not match sibling dir
		{"", false},
	}
	for _, tc := range cases {
		if got := isSystemDaemonPath(tc.path); got != tc.want {
			t.Errorf("isSystemDaemonPath(%q) = %v, want %v", tc.path, got, tc.want)
		}
	}
}
