package main

import (
	"os"
	"testing"
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
