package launchctl

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

// fakeAdminClient records all RPC invocations so tests can assert SystemManager
// correctly delegates to the helper when Admin Mode is enabled. Every call
// returns nil by default; injecting errX via the respective *Err field makes
// the corresponding method fail.
type fakeAdminClient struct {
	mu   sync.Mutex
	logs []string

	bootstrapErr       error
	bootoutErr         error
	kickstartErr       error
	writePlistErr      error
	deletePlistErr     error
	ensureLogAccessErr error

	lastWriteData []byte
	lastLogPaths  []string
}

func (f *fakeAdminClient) record(s string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.logs = append(f.logs, s)
}

func (f *fakeAdminClient) calls() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]string(nil), f.logs...)
}

func (f *fakeAdminClient) Bootstrap(_ context.Context, path string) error {
	f.record("Bootstrap:" + path)
	return f.bootstrapErr
}

func (f *fakeAdminClient) Bootout(_ context.Context, label string) error {
	f.record("Bootout:" + label)
	return f.bootoutErr
}

func (f *fakeAdminClient) Kickstart(_ context.Context, label string) error {
	f.record("Kickstart:" + label)
	return f.kickstartErr
}

func (f *fakeAdminClient) WritePlist(_ context.Context, path string, data []byte) error {
	f.record("WritePlist:" + path)
	f.lastWriteData = append([]byte(nil), data...)
	return f.writePlistErr
}

func (f *fakeAdminClient) DeletePlist(_ context.Context, path string) error {
	f.record("DeletePlist:" + path)
	return f.deletePlistErr
}

func (f *fakeAdminClient) EnsureLogAccess(_ context.Context, paths []string) error {
	f.record("EnsureLogAccess:" + strings.Join(paths, ","))
	f.mu.Lock()
	f.lastLogPaths = append([]string(nil), paths...)
	f.mu.Unlock()
	return f.ensureLogAccessErr
}

func TestSystemManager_List(t *testing.T) {
	m := NewSystemManager()

	// Skip if directory doesn't exist (not on macOS or no permissions)
	if _, err := os.Stat("/Library/LaunchDaemons"); os.IsNotExist(err) {
		t.Skip("Skipping: /Library/LaunchDaemons not found")
	}

	services, err := m.List()
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	// Verify all services have correct type
	for _, svc := range services {
		if svc.Type != "system" {
			t.Errorf("Service %s has Type=%s, want 'system'", svc.Name, svc.Type)
		}
		if !svc.ReadOnly {
			t.Errorf("Service %s has ReadOnly=false, want true", svc.Name)
		}
	}
}

func TestSystemManager_WriteOperationsReturnError(t *testing.T) {
	m := NewSystemManager()

	tests := []struct {
		name string
		fn   func() error
	}{
		{"Start", func() error { return m.Start("test") }},
		{"Stop", func() error { return m.Stop("test") }},
		{"Restart", func() error { return m.Restart("test") }},
		{"Create", func() error { return m.Create(&ServiceConfig{Label: "test"}) }},
		{"Update", func() error { return m.Update("test", &ServiceConfig{Label: "test"}) }},
		{"Delete", func() error { return m.Delete("test") }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.fn()
			if err != ErrReadOnlyManager {
				t.Errorf("%s() error = %v, want ErrReadOnlyManager", tt.name, err)
			}
		})
	}
}

func TestSystemManager_Get(t *testing.T) {
	tmpDir := t.TempDir()

	plistContent := `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>Label</key>
	<string>com.test.daemon</string>
	<key>Program</key>
	<string>/usr/sbin/testd</string>
	<key>RunAtLoad</key>
	<true/>
	<key>StandardOutPath</key>
	<string>/var/log/testd.log</string>
</dict>
</plist>`

	plistPath := filepath.Join(tmpDir, "com.test.daemon.plist")
	if err := os.WriteFile(plistPath, []byte(plistContent), 0644); err != nil {
		t.Fatalf("failed to write plist file: %v", err)
	}

	m := &SystemManager{readOnlyManager: readOnlyManager{basePath: tmpDir, serviceType: "system"}}

	// Test Get
	svc, err := m.Get("com.test.daemon")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if svc.Type != "system" {
		t.Errorf("Type = %q, want %q", svc.Type, "system")
	}
	if !svc.ReadOnly {
		t.Error("ReadOnly = false, want true")
	}
	if svc.Label != "com.test.daemon" {
		t.Errorf("Label = %q, want %q", svc.Label, "com.test.daemon")
	}
	if svc.Program != "/usr/sbin/testd" {
		t.Errorf("Program = %q, want %q", svc.Program, "/usr/sbin/testd")
	}
	if !svc.RunAtLoad {
		t.Error("RunAtLoad = false, want true")
	}
	if svc.StdoutPath != "/var/log/testd.log" {
		t.Errorf("StdoutPath = %q, want %q", svc.StdoutPath, "/var/log/testd.log")
	}
	if svc.PlistFormat != "xml" {
		t.Errorf("PlistFormat = %q, want %q", svc.PlistFormat, "xml")
	}

	// Test GetPlist (plutil works with valid XML plist files)
	plistOutput, err := m.GetPlist("com.test.daemon")
	if err != nil {
		t.Fatalf("GetPlist() error = %v", err)
	}
	if !strings.Contains(plistOutput, "com.test.daemon") {
		t.Errorf("GetPlist() output does not contain Label, got: %s", plistOutput)
	}
}

func TestSystemManager_List_TempDir(t *testing.T) {
	t.Run("with plists and non-plist files", func(t *testing.T) {
		tmpDir := t.TempDir()

		plistTemplate := `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>Label</key>
	<string>%s</string>
	<key>Program</key>
	<string>/usr/bin/true</string>
</dict>
</plist>`

		// Write two valid plist files
		for _, name := range []string{"com.test.first", "com.test.second"} {
			content := strings.Replace(plistTemplate, "%s", name, 1)
			path := filepath.Join(tmpDir, name+".plist")
			if err := os.WriteFile(path, []byte(content), 0644); err != nil {
				t.Fatalf("failed to write plist: %v", err)
			}
		}

		// Write a non-plist file that should be ignored
		nonPlist := filepath.Join(tmpDir, "readme.txt")
		if err := os.WriteFile(nonPlist, []byte("not a plist"), 0644); err != nil {
			t.Fatalf("failed to write non-plist file: %v", err)
		}

		m := &SystemManager{readOnlyManager: readOnlyManager{basePath: tmpDir, serviceType: "system"}}
		services, err := m.List()
		if err != nil {
			t.Fatalf("List() error = %v", err)
		}

		if len(services) != 2 {
			t.Fatalf("List() returned %d services, want 2", len(services))
		}

		for _, svc := range services {
			if svc.Type != "system" {
				t.Errorf("Service %s Type = %q, want %q", svc.Name, svc.Type, "system")
			}
			if !svc.ReadOnly {
				t.Errorf("Service %s ReadOnly = false, want true", svc.Name)
			}
		}
	})

	t.Run("empty directory", func(t *testing.T) {
		tmpDir := t.TempDir()

		m := &SystemManager{readOnlyManager: readOnlyManager{basePath: tmpDir, serviceType: "system"}}
		services, err := m.List()
		if err != nil {
			t.Fatalf("List() error = %v", err)
		}

		if len(services) != 0 {
			t.Errorf("List() returned %d services, want 0", len(services))
		}
	})
}

func TestSystemManager_WriteOperationsWithAdminModeEnabled(t *testing.T) {
	tmp := t.TempDir()
	// Seed a plist so Get (used by Stop/Restart to resolve label) works.
	// RunAtLoad=true makes Start's Bootstrap path a single-step operation;
	// the already-loaded fallback path is covered by a dedicated test.
	plist := `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0"><dict>
  <key>Label</key><string>com.example.testdaemon</string>
  <key>Program</key><string>/usr/bin/true</string>
  <key>RunAtLoad</key><true/>
</dict></plist>`
	path := filepath.Join(tmp, "com.example.testdaemon.plist")
	if err := os.WriteFile(path, []byte(plist), 0644); err != nil {
		t.Fatalf("seed: %v", err)
	}

	m := &SystemManager{readOnlyManager: readOnlyManager{basePath: tmp, serviceType: "system"}}
	fake := &fakeAdminClient{}
	m.SetAdminClient(fake)

	t.Run("Start with RunAtLoad=true only bootstraps", func(t *testing.T) {
		fake.logs = nil
		if err := m.Start("com.example.testdaemon"); err != nil {
			t.Fatalf("Start: %v", err)
		}
		// Bootstrap succeeded and the plist has RunAtLoad=true, so launchd
		// spawns the job on its own. An extra Kickstart here would trigger
		// the 10-second throttle (minimum_runtime backoff).
		want := []string{"Bootstrap:" + path}
		got := fake.calls()
		if strings.Join(got, "|") != strings.Join(want, "|") {
			t.Errorf("calls = %v, want %v", got, want)
		}
	})

	t.Run("Stop delegates to Bootout with label", func(t *testing.T) {
		fake.logs = nil
		if err := m.Stop("com.example.testdaemon"); err != nil {
			t.Fatalf("Stop: %v", err)
		}
		want := "Bootout:com.example.testdaemon"
		got := fake.calls()
		if len(got) != 1 || got[0] != want {
			t.Errorf("calls = %v, want [%s]", got, want)
		}
	})

	t.Run("Restart delegates to Kickstart with label", func(t *testing.T) {
		fake.logs = nil
		if err := m.Restart("com.example.testdaemon"); err != nil {
			t.Fatalf("Restart: %v", err)
		}
		want := "Kickstart:com.example.testdaemon"
		got := fake.calls()
		if len(got) != 1 || got[0] != want {
			t.Errorf("calls = %v, want [%s]", got, want)
		}
	})

	t.Run("Create writes plist then bootstraps", func(t *testing.T) {
		fake.logs = nil
		fake.lastWriteData = nil
		cfg := &ServiceConfig{
			Label:   "com.example.new",
			Program: "/usr/bin/foo",
		}
		if err := m.Create(cfg); err != nil {
			t.Fatalf("Create: %v", err)
		}
		got := fake.calls()
		wantPath := filepath.Join(tmp, "com.example.new.plist")
		want := []string{"WritePlist:" + wantPath, "Bootstrap:" + wantPath}
		if strings.Join(got, "|") != strings.Join(want, "|") {
			t.Errorf("calls = %v, want %v", got, want)
		}
		if !strings.Contains(string(fake.lastWriteData), "com.example.new") {
			t.Errorf("plist body missing label: %s", string(fake.lastWriteData))
		}
	})

	t.Run("Update boots out then writes plist", func(t *testing.T) {
		// Bootout is required so launchd drops its in-memory config and
		// picks up the new plist on the next Start; WritePlist alone
		// leaves launchd running the previous definition forever.
		fake.logs = nil
		cfg := &ServiceConfig{
			Label:   "com.example.testdaemon",
			Program: "/usr/bin/bar",
		}
		if err := m.Update("com.example.testdaemon", cfg); err != nil {
			t.Fatalf("Update: %v", err)
		}
		want := []string{"Bootout:com.example.testdaemon", "WritePlist:" + path}
		got := fake.calls()
		if strings.Join(got, "|") != strings.Join(want, "|") {
			t.Errorf("calls = %v, want %v", got, want)
		}
	})

	t.Run("Delete boots out then removes plist", func(t *testing.T) {
		fake.logs = nil
		if err := m.Delete("com.example.testdaemon"); err != nil {
			t.Fatalf("Delete: %v", err)
		}
		want := []string{"Bootout:com.example.testdaemon", "DeletePlist:" + path}
		got := fake.calls()
		if strings.Join(got, "|") != strings.Join(want, "|") {
			t.Errorf("calls = %v, want %v", got, want)
		}
	})

	t.Run("ClearAdminClient reverts to read-only", func(t *testing.T) {
		m.ClearAdminClient()
		if err := m.Start("com.example.testdaemon"); !errors.Is(err, ErrReadOnlyManager) {
			t.Errorf("Start err = %v, want ErrReadOnlyManager", err)
		}
	})
}

func TestSystemManager_CreateAndUpdateCallEnsureLogAccess(t *testing.T) {
	// When the user provides log paths, Create/Update must tell the helper
	// to make those paths traversable by the unprivileged GUI — otherwise a
	// root-owned /var/log/<svc>/ created by launchd at 0744 blocks the
	// Logs tab. EnsureLogAccess is best-effort so errors are swallowed, but
	// the call itself must still happen.
	tmp := t.TempDir()
	// Seed a plist so Update's Bootout-first step has something to resolve.
	seed := `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0"><dict>
  <key>Label</key><string>com.example.loggy</string>
</dict></plist>`
	path := filepath.Join(tmp, "com.example.loggy.plist")
	if err := os.WriteFile(path, []byte(seed), 0644); err != nil {
		t.Fatalf("seed: %v", err)
	}

	m := &SystemManager{readOnlyManager: readOnlyManager{basePath: tmp, serviceType: "system"}}
	fake := &fakeAdminClient{}
	m.SetAdminClient(fake)

	t.Run("Create with both log paths", func(t *testing.T) {
		fake.logs = nil
		fake.lastLogPaths = nil
		cfg := &ServiceConfig{
			Label:      "com.example.loggy",
			Program:    "/usr/bin/true",
			StdoutPath: "/var/log/loggy/stdout.log",
			StderrPath: "/var/log/loggy/stderr.log",
		}
		if err := m.Create(cfg); err != nil {
			t.Fatalf("Create: %v", err)
		}
		want := []string{
			"WritePlist:" + path,
			"EnsureLogAccess:/var/log/loggy/stdout.log,/var/log/loggy/stderr.log",
			"Bootstrap:" + path,
		}
		got := fake.calls()
		if strings.Join(got, "|") != strings.Join(want, "|") {
			t.Errorf("calls = %v, want %v", got, want)
		}
	})

	t.Run("Update with stdout only", func(t *testing.T) {
		fake.logs = nil
		fake.lastLogPaths = nil
		cfg := &ServiceConfig{
			Label:      "com.example.loggy",
			Program:    "/usr/bin/true",
			StdoutPath: "/var/log/loggy/out.log",
		}
		if err := m.Update("com.example.loggy", cfg); err != nil {
			t.Fatalf("Update: %v", err)
		}
		// EnsureLogAccess only receives the non-empty paths.
		want := []string{
			"Bootout:com.example.loggy",
			"WritePlist:" + path,
			"EnsureLogAccess:/var/log/loggy/out.log",
		}
		got := fake.calls()
		if strings.Join(got, "|") != strings.Join(want, "|") {
			t.Errorf("calls = %v, want %v", got, want)
		}
	})

	t.Run("Create without log paths skips EnsureLogAccess", func(t *testing.T) {
		fake.logs = nil
		cfg := &ServiceConfig{
			Label:   "com.example.nopath",
			Program: "/usr/bin/true",
		}
		if err := m.Create(cfg); err != nil {
			t.Fatalf("Create: %v", err)
		}
		for _, c := range fake.calls() {
			if strings.HasPrefix(c, "EnsureLogAccess") {
				t.Errorf("unexpected EnsureLogAccess call when no log paths provided: %q", c)
			}
		}
	})

	t.Run("EnsureLogAccess failure is non-fatal", func(t *testing.T) {
		fake.logs = nil
		fake.ensureLogAccessErr = errors.New("helper had a bad day")
		defer func() { fake.ensureLogAccessErr = nil }()
		cfg := &ServiceConfig{
			Label:      "com.example.loggy",
			Program:    "/usr/bin/true",
			StdoutPath: "/var/log/loggy/out.log",
		}
		if err := m.Create(cfg); err != nil {
			t.Errorf("Create should not fail when EnsureLogAccess errors: %v", err)
		}
	})
}

func TestSystemManager_StartIgnoresBootstrapErrorWhenKickstartSucceeds(t *testing.T) {
	// Common case: user re-clicks Start on an already-bootstrapped daemon.
	// Bootstrap fails with "already loaded" but Kickstart still fires and
	// succeeds, so Start returns nil.
	tmp := t.TempDir()
	plistContent := `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0"><dict>
  <key>Label</key><string>com.example.loaded</string>
</dict></plist>`
	path := filepath.Join(tmp, "com.example.loaded.plist")
	if err := os.WriteFile(path, []byte(plistContent), 0644); err != nil {
		t.Fatalf("seed: %v", err)
	}

	m := &SystemManager{readOnlyManager: readOnlyManager{basePath: tmp, serviceType: "system"}}
	fake := &fakeAdminClient{bootstrapErr: errors.New("already loaded")}
	m.SetAdminClient(fake)
	if err := m.Start("com.example.loaded"); err != nil {
		t.Errorf("Start: expected nil when Kickstart succeeds, got %v", err)
	}
	// Both Bootstrap (failed) and Kickstart (succeeded) should have been
	// invoked in that order.
	got := fake.calls()
	want := []string{"Bootstrap:" + path, "Kickstart:com.example.loaded"}
	if strings.Join(got, "|") != strings.Join(want, "|") {
		t.Errorf("calls = %v, want %v", got, want)
	}
}

func TestSystemManager_StartKickstartsWhenRunAtLoadFalse(t *testing.T) {
	// Fresh bootstrap but RunAtLoad=false: launchd won't spawn the process
	// on its own, so Start must follow up with Kickstart.
	tmp := t.TempDir()
	plistContent := `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0"><dict>
  <key>Label</key><string>com.example.lazy</string>
</dict></plist>`
	path := filepath.Join(tmp, "com.example.lazy.plist")
	if err := os.WriteFile(path, []byte(plistContent), 0644); err != nil {
		t.Fatalf("seed: %v", err)
	}

	m := &SystemManager{readOnlyManager: readOnlyManager{basePath: tmp, serviceType: "system"}}
	fake := &fakeAdminClient{}
	m.SetAdminClient(fake)
	if err := m.Start("com.example.lazy"); err != nil {
		t.Fatalf("Start: %v", err)
	}
	got := fake.calls()
	want := []string{"Bootstrap:" + path, "Kickstart:com.example.lazy"}
	if strings.Join(got, "|") != strings.Join(want, "|") {
		t.Errorf("calls = %v, want %v", got, want)
	}
}

func TestSystemManager_StartReturnsBootstrapErrorWhenBothFail(t *testing.T) {
	tmp := t.TempDir()
	m := &SystemManager{readOnlyManager: readOnlyManager{basePath: tmp, serviceType: "system"}}
	fake := &fakeAdminClient{
		bootstrapErr: errors.New("bootstrap failed: file not found"),
		kickstartErr: errors.New("service not loaded"),
	}
	m.SetAdminClient(fake)
	err := m.Start("nope")
	if err == nil {
		t.Fatal("expected error")
	}
	// Bootstrap error is preferred — it's the more informative of the two.
	if !strings.Contains(err.Error(), "bootstrap failed") {
		t.Errorf("err = %v, want bootstrap failed", err)
	}
}

func TestSystemManager_RestorePlist(t *testing.T) {
	tmp := t.TempDir()
	m := &SystemManager{readOnlyManager: readOnlyManager{basePath: tmp, serviceType: "system"}}

	t.Run("returns ErrReadOnlyManager when admin mode off", func(t *testing.T) {
		err := m.RestorePlist(filepath.Join(tmp, "foo.plist"), []byte("<plist/>"))
		if !errors.Is(err, ErrReadOnlyManager) {
			t.Errorf("err = %v, want ErrReadOnlyManager", err)
		}
	})

	t.Run("delegates to AdminClient.WritePlist", func(t *testing.T) {
		fake := &fakeAdminClient{}
		m.SetAdminClient(fake)
		defer m.ClearAdminClient()

		target := filepath.Join(tmp, "com.example.loggy.plist")
		data := []byte("<plist>restored</plist>")
		if err := m.RestorePlist(target, data); err != nil {
			t.Fatalf("RestorePlist: %v", err)
		}
		got := fake.calls()
		if len(got) != 1 || got[0] != "WritePlist:"+target {
			t.Errorf("calls = %v, want [WritePlist:%s]", got, target)
		}
		if string(fake.lastWriteData) != string(data) {
			t.Errorf("data = %q, want %q", fake.lastWriteData, data)
		}
	})
}

func TestSystemManager_GetPlistContent(t *testing.T) {
	tmp := t.TempDir()
	m := &SystemManager{readOnlyManager: readOnlyManager{basePath: tmp, serviceType: "system"}}

	t.Run("returns XML for existing plist", func(t *testing.T) {
		body := `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0"><dict>
  <key>Label</key><string>com.example.exist</string>
</dict></plist>`
		if err := os.WriteFile(filepath.Join(tmp, "com.example.exist.plist"), []byte(body), 0644); err != nil {
			t.Fatalf("seed: %v", err)
		}
		got, err := m.GetPlistContent("com.example.exist")
		if err != nil {
			t.Fatalf("GetPlistContent: %v", err)
		}
		if !strings.Contains(got.Data, "com.example.exist") {
			t.Errorf("content missing label: %q", got.Data)
		}
	})

	t.Run("returns empty content (not error) when plist missing", func(t *testing.T) {
		got, err := m.GetPlistContent("not-present")
		if err != nil {
			t.Fatalf("GetPlistContent missing: %v", err)
		}
		if got == nil || got.Data != "" {
			t.Errorf("missing plist = %+v, want empty Content", got)
		}
	})
}

func TestSystemManager_GetLogs(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a temp log file with content
	logFile := filepath.Join(tmpDir, "testd.log")
	logContent := "2024-01-01 daemon started\n2024-01-02 processing request\n"
	if err := os.WriteFile(logFile, []byte(logContent), 0644); err != nil {
		t.Fatalf("failed to write log file: %v", err)
	}

	// Create a plist that references the log file
	plistContent := `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>Label</key>
	<string>com.test.logging</string>
	<key>Program</key>
	<string>/usr/bin/true</string>
	<key>StandardOutPath</key>
	<string>` + logFile + `</string>
</dict>
</plist>`

	plistPath := filepath.Join(tmpDir, "com.test.logging.plist")
	if err := os.WriteFile(plistPath, []byte(plistContent), 0644); err != nil {
		t.Fatalf("failed to write plist: %v", err)
	}

	m := &SystemManager{readOnlyManager: readOnlyManager{basePath: tmpDir, serviceType: "system"}}

	t.Run("stdout returns log content", func(t *testing.T) {
		got, err := m.GetLogs("com.test.logging", "stdout")
		if err != nil {
			t.Fatalf("GetLogs(stdout) error = %v", err)
		}
		if got != logContent {
			t.Errorf("GetLogs(stdout) = %q, want %q", got, logContent)
		}
	})

	t.Run("stderr returns error when not configured", func(t *testing.T) {
		_, err := m.GetLogs("com.test.logging", "stderr")
		if err == nil {
			t.Fatal("GetLogs(stderr) expected error, got nil")
		}
		if !strings.Contains(err.Error(), "no stderr log path configured") {
			t.Errorf("GetLogs(stderr) error = %q, want error about no stderr path", err.Error())
		}
	})

	t.Run("invalid log type returns error", func(t *testing.T) {
		_, err := m.GetLogs("com.test.logging", "invalid")
		if err == nil {
			t.Fatal("GetLogs(invalid) expected error, got nil")
		}
		if !strings.Contains(err.Error(), "invalid log type") {
			t.Errorf("GetLogs(invalid) error = %q, want error about invalid log type", err.Error())
		}
	})
}
