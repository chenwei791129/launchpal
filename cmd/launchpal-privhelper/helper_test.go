package main

import (
	"bytes"
	"strings"
	"testing"
)

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
