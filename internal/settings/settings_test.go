package settings

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"launchpal/internal/privhelper"
)

// withHome redirects $HOME so Load/Save resolve to a sandboxed
// ~/.launchpal/settings.json instead of the real one.
func withHome(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	return dir
}

func TestSettingsPath_UnderHome(t *testing.T) {
	home := withHome(t)
	got, err := Path()
	if err != nil {
		t.Fatalf("Path: %v", err)
	}
	want := filepath.Join(home, ".launchpal", "settings.json")
	if got != want {
		t.Errorf("Path = %q, want %q", got, want)
	}
}

func TestDefault_Values(t *testing.T) {
	s := Default()
	if s.UserLogDir != "~/Library/Logs" {
		t.Errorf("UserLogDir = %q, want %q", s.UserLogDir, "~/Library/Logs")
	}
	if s.SystemLogDir != "/Library/Logs" {
		t.Errorf("SystemLogDir = %q, want %q", s.SystemLogDir, "/Library/Logs")
	}
}

func TestDefault_NoFilesystemSideEffects(t *testing.T) {
	home := withHome(t)
	_ = Default()
	if _, err := os.Stat(filepath.Join(home, ".launchpal")); !os.IsNotExist(err) {
		t.Errorf(".launchpal dir created by Default(): err=%v", err)
	}
}

func TestLoad_MissingFileReturnsDefaults(t *testing.T) {
	withHome(t)
	got, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if got != Default() {
		t.Errorf("Load = %+v, want Default %+v", got, Default())
	}
}

func TestLoad_CorruptJSONReturnsDefaults(t *testing.T) {
	home := withHome(t)
	dir := filepath.Join(home, ".launchpal")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "settings.json"), []byte("{not json"), 0644); err != nil {
		t.Fatal(err)
	}
	got, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if got != Default() {
		t.Errorf("Load on corrupt JSON = %+v, want Default %+v", got, Default())
	}
}

func TestLoad_PartialJSONMergesWithDefaults(t *testing.T) {
	home := withHome(t)
	dir := filepath.Join(home, ".launchpal")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "settings.json"), []byte(`{"userLogDir": "/tmp/mylogs"}`), 0644); err != nil {
		t.Fatal(err)
	}
	got, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if got.UserLogDir != "/tmp/mylogs" {
		t.Errorf("UserLogDir = %q, want %q", got.UserLogDir, "/tmp/mylogs")
	}
	if got.SystemLogDir != "/Library/Logs" {
		t.Errorf("SystemLogDir = %q, want default %q", got.SystemLogDir, "/Library/Logs")
	}
}

func TestSave_WritesValidJSON(t *testing.T) {
	home := withHome(t)
	want := Settings{UserLogDir: "~/Library/Logs", SystemLogDir: "/Library/Logs/launchpal"}
	if err := Save(want); err != nil {
		t.Fatalf("Save: %v", err)
	}
	path := filepath.Join(home, ".launchpal", "settings.json")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if !strings.HasSuffix(string(raw), "\n") {
		t.Errorf("file does not end with newline: %q", raw)
	}
	if !strings.Contains(string(raw), "  ") {
		t.Errorf("file does not appear to be 2-space-indented: %q", raw)
	}
	var got Settings
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got != want {
		t.Errorf("round-trip = %+v, want %+v", got, want)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0644 {
		t.Errorf("perm = %o, want 0644", info.Mode().Perm())
	}
}

func TestSave_AtomicReplace(t *testing.T) {
	home := withHome(t)
	dir := filepath.Join(home, ".launchpal")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "settings.json")
	if err := os.WriteFile(path, []byte(`{"userLogDir": "old", "systemLogDir": "old"}`), 0644); err != nil {
		t.Fatal(err)
	}
	want := Settings{UserLogDir: "/tmp/u", SystemLogDir: "/Library/Logs/launchpal"}
	if err := Save(want); err != nil {
		t.Fatalf("Save: %v", err)
	}
	got, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if got != want {
		t.Errorf("Load after Save = %+v, want %+v", got, want)
	}
	// No leftover *.tmp files in the parent dir.
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".tmp") || strings.Contains(e.Name(), "settings.json.") {
			t.Errorf("leftover temp file %q", e.Name())
		}
	}
}

func TestSave_CreatesParentDir(t *testing.T) {
	home := withHome(t)
	if _, err := os.Stat(filepath.Join(home, ".launchpal")); !os.IsNotExist(err) {
		t.Fatalf("precondition: .launchpal already exists: err=%v", err)
	}
	if err := Save(Default()); err != nil {
		t.Fatalf("Save: %v", err)
	}
	info, err := os.Stat(filepath.Join(home, ".launchpal"))
	if err != nil {
		t.Fatalf("stat .launchpal: %v", err)
	}
	if !info.IsDir() {
		t.Error(".launchpal is not a directory")
	}
	if info.Mode().Perm() != 0755 {
		t.Errorf("perm = %o, want 0755", info.Mode().Perm())
	}
}

func TestSave_RejectsInvalidWithoutWriting(t *testing.T) {
	home := withHome(t)
	dir := filepath.Join(home, ".launchpal")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "settings.json")
	const original = `{"userLogDir":"/Users/a/logs","systemLogDir":"/Library/Logs/keep"}`
	if err := os.WriteFile(path, []byte(original), 0644); err != nil {
		t.Fatal(err)
	}
	bad := Settings{UserLogDir: "~/Library/Logs", SystemLogDir: "/etc/launchpal"}
	err := Save(bad)
	if err == nil {
		t.Fatal("Save: expected error, got nil")
	}
	got, _ := os.ReadFile(path)
	if string(got) != original {
		t.Errorf("file modified despite validation failure: %q", got)
	}
	// no leftover .tmp
	entries, _ := os.ReadDir(dir)
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".tmp") {
			t.Errorf("leftover temp file %q", e.Name())
		}
	}
}

func TestValidate_SystemLogDirMatrix(t *testing.T) {
	cases := []struct {
		name      string
		systemDir string
		ok        bool
		mustSay   string
	}{
		{"valid Library/Logs sub", "/Library/Logs/launchpal", true, ""},
		{"valid var/log sub", "/var/log/myapp", true, ""},
		{"valid private/var/log sub", "/private/var/log/myapp", true, ""},
		{"valid tmp sub", "/tmp/launchpal", true, ""},
		{"valid private/tmp sub", "/private/tmp/launchpal", true, ""},
		{"valid bare allowlist root (default)", "/Library/Logs", true, ""},
		{"valid bare allowlist root with trailing slash", "/Library/Logs/", true, ""},
		{"reject prefix not in allowlist", "/etc/launchpal", false, "must start with"},
		{"reject sibling lookalike", "/var/logX/foo", false, "must start with"},
		{"reject tilde for systemLogDir", "~/Library/Logs/myapp", false, "must be absolute"},
		{"reject empty", "", false, "systemLogDir"},
		{"reject shell metacharacter", "/var/log/$(rm -rf)", false, "shell metacharacter"},
		{"reject relative", "Library/Logs/myapp", false, "must be absolute"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := Settings{UserLogDir: "~/Library/Logs", SystemLogDir: tc.systemDir}
			err := Validate(s)
			if tc.ok && err != nil {
				t.Errorf("Validate(%q) = %v, want nil", tc.systemDir, err)
				return
			}
			if !tc.ok {
				if err == nil {
					t.Errorf("Validate(%q) = nil, want error", tc.systemDir)
					return
				}
				if !strings.Contains(err.Error(), "systemLogDir") {
					t.Errorf("Validate(%q) error %q does not mention systemLogDir", tc.systemDir, err)
				}
				if tc.mustSay != "" && !strings.Contains(err.Error(), tc.mustSay) {
					t.Errorf("Validate(%q) error %q does not contain %q", tc.systemDir, err, tc.mustSay)
				}
			}
		})
	}
}

func TestValidate_SystemLogDirAllowlistShared(t *testing.T) {
	// Spec: settings validator and privileged helper must consume the same
	// shared constant. Each prefix exported by privhelper is accepted
	// (with a sub-directory appended) by Validate.
	for _, prefix := range privhelper.SystemLogPathPrefixes {
		dir := strings.TrimSuffix(prefix, "/") + "/launchpal-test"
		s := Settings{UserLogDir: "~/Library/Logs", SystemLogDir: dir}
		if err := Validate(s); err != nil {
			t.Errorf("Validate accepted by helper rejects via settings: prefix=%q dir=%q err=%v", prefix, dir, err)
		}
	}
}

func TestValidate_UserLogDirMatrix(t *testing.T) {
	cases := []struct {
		name    string
		userDir string
		ok      bool
		mustSay string
	}{
		{"valid tilde home", "~/Library/Logs", true, ""},
		{"valid absolute", "/Users/foo/logs", true, ""},
		{"reject relative", "Library/Logs", false, "must be tilde-home"},
		{"reject empty", "", false, "userLogDir"},
		{"reject shell metacharacter semicolon", "~/logs;rm", false, "shell metacharacter"},
		{"reject shell metacharacter pipe", "~/logs|x", false, "shell metacharacter"},
		{"reject shell metacharacter dollar", "~/logs$x", false, "shell metacharacter"},
		{"reject shell metacharacter backtick", "~/logs`x", false, "shell metacharacter"},
		{"reject newline", "~/logs\nx", false, "shell metacharacter"},
		{"reject null byte", "~/logs\x00x", false, "shell metacharacter"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := Settings{UserLogDir: tc.userDir, SystemLogDir: "/Library/Logs/launchpal"}
			err := Validate(s)
			if tc.ok && err != nil {
				t.Errorf("Validate(%q) = %v, want nil", tc.userDir, err)
				return
			}
			if !tc.ok {
				if err == nil {
					t.Errorf("Validate(%q) = nil, want error", tc.userDir)
					return
				}
				if !strings.Contains(err.Error(), "userLogDir") {
					t.Errorf("Validate(%q) error %q does not mention userLogDir", tc.userDir, err)
				}
				if tc.mustSay != "" && !strings.Contains(err.Error(), tc.mustSay) {
					t.Errorf("Validate(%q) error %q does not contain %q", tc.userDir, err, tc.mustSay)
				}
			}
		})
	}
}

func TestValidate_DefaultSettingsAccepted(t *testing.T) {
	if err := Validate(Default()); err != nil {
		t.Errorf("Validate(Default()) = %v, want nil", err)
	}
}

func TestSave_PropagatesNonValidationFilesystemErrors(t *testing.T) {
	// Block creation of ~/.launchpal by replacing it with a regular file.
	home := withHome(t)
	if err := os.WriteFile(filepath.Join(home, ".launchpal"), []byte("blocker"), 0644); err != nil {
		t.Fatal(err)
	}
	err := Save(Default())
	if err == nil {
		t.Fatal("Save: expected filesystem error, got nil")
	}
	// Ensure it's not a validation error masquerading as IO failure.
	var verr *ValidationError
	if errors.As(err, &verr) {
		t.Errorf("Save returned ValidationError on filesystem error: %v", err)
	}
}
