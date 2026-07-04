package launchctl

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAppleSystemManager_List(t *testing.T) {
	m := NewAppleSystemManager()

	// Skip if directory doesn't exist (not on macOS or no permissions)
	if _, err := os.Stat("/System/Library/LaunchDaemons"); os.IsNotExist(err) {
		t.Skip("Skipping: /System/Library/LaunchDaemons not found")
	}

	services, err := m.List()
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	// Verify all services have correct type
	for _, svc := range services {
		if svc.Type != "apple-system" {
			t.Errorf("Service %s has Type=%s, want 'apple-system'", svc.Name, svc.Type)
		}
		if !svc.ReadOnly {
			t.Errorf("Service %s has ReadOnly=false, want true", svc.Name)
		}
	}
}

func TestAppleSystemManager_WriteOperationsReturnError(t *testing.T) {
	m := NewAppleSystemManager()

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

func TestAppleSystemManager_Get(t *testing.T) {
	tmpDir := t.TempDir()

	plistContent := `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>Label</key>
	<string>com.apple.test</string>
	<key>Program</key>
	<string>/usr/libexec/testd</string>
	<key>RunAtLoad</key>
	<true/>
</dict>
</plist>`

	err := os.WriteFile(filepath.Join(tmpDir, "com.apple.test.plist"), []byte(plistContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write test plist: %v", err)
	}

	m := &AppleSystemManager{readOnlyManager: readOnlyManager{basePath: tmpDir, serviceType: "apple-system"}}
	svc, err := m.Get("com.apple.test")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if svc.Type != "apple-system" {
		t.Errorf("Type = %q, want %q", svc.Type, "apple-system")
	}
	if !svc.ReadOnly {
		t.Error("ReadOnly = false, want true")
	}
	if svc.Label != "com.apple.test" {
		t.Errorf("Label = %q, want %q", svc.Label, "com.apple.test")
	}
	if svc.Program != "/usr/libexec/testd" {
		t.Errorf("Program = %q, want %q", svc.Program, "/usr/libexec/testd")
	}
	if svc.PlistFormat != "xml" {
		t.Errorf("PlistFormat = %q, want %q", svc.PlistFormat, "xml")
	}
}

func TestAppleSystemManager_List_TempDir(t *testing.T) {
	t.Run("with services", func(t *testing.T) {
		tmpDir := t.TempDir()

		plist1 := `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>Label</key>
	<string>com.apple.svc1</string>
	<key>Program</key>
	<string>/usr/libexec/svc1</string>
</dict>
</plist>`

		plist2 := `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>Label</key>
	<string>com.apple.svc2</string>
	<key>Program</key>
	<string>/usr/libexec/svc2</string>
</dict>
</plist>`

		if err := os.WriteFile(filepath.Join(tmpDir, "com.apple.svc1.plist"), []byte(plist1), 0644); err != nil {
			t.Fatalf("failed to write plist: %v", err)
		}
		if err := os.WriteFile(filepath.Join(tmpDir, "com.apple.svc2.plist"), []byte(plist2), 0644); err != nil {
			t.Fatalf("failed to write plist: %v", err)
		}

		m := &AppleSystemManager{readOnlyManager: readOnlyManager{basePath: tmpDir, serviceType: "apple-system"}}
		services, err := m.List()
		if err != nil {
			t.Fatalf("List() error = %v", err)
		}

		if len(services) != 2 {
			t.Fatalf("List() returned %d services, want 2", len(services))
		}

		for _, svc := range services {
			if svc.Type != "apple-system" {
				t.Errorf("Service %s has Type=%q, want %q", svc.Name, svc.Type, "apple-system")
			}
		}
	})

	t.Run("empty directory", func(t *testing.T) {
		tmpDir := t.TempDir()

		m := &AppleSystemManager{readOnlyManager: readOnlyManager{basePath: tmpDir, serviceType: "apple-system"}}
		services, err := m.List()
		if err != nil {
			t.Fatalf("List() error = %v", err)
		}

		if len(services) != 0 {
			t.Errorf("List() returned %d services, want 0", len(services))
		}
	})
}

func TestAppleSystemManager_ClearLogs(t *testing.T) {
	m := NewAppleSystemManager()
	err := m.ClearLogs("com.apple.anything", "stdout")
	if err == nil {
		t.Fatal("ClearLogs should return error for apple-system services")
	}
	msg := err.Error()
	if !strings.Contains(msg, "apple-system") && !strings.Contains(msg, "read-only") {
		t.Errorf("err = %q, want it to mention apple-system or read-only", msg)
	}
}

func TestAppleSystemManager_GetLogClearStatus(t *testing.T) {
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "log.log")
	if err := os.WriteFile(logFile, []byte("x"), 0644); err != nil {
		t.Fatalf("seed: %v", err)
	}
	plistContent := `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>Label</key><string>com.apple.x</string>
	<key>Program</key><string>/usr/libexec/x</string>
	<key>StandardOutPath</key><string>` + logFile + `</string>
</dict>
</plist>`
	if err := os.WriteFile(filepath.Join(tmpDir, "com.apple.x.plist"), []byte(plistContent), 0644); err != nil {
		t.Fatalf("seed: %v", err)
	}
	m := &AppleSystemManager{readOnlyManager: readOnlyManager{basePath: tmpDir, serviceType: "apple-system"}}
	got, err := m.GetLogClearStatus("com.apple.x", "stdout")
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if got.LogPath != logFile {
		t.Errorf("LogPath = %q, want %q", got.LogPath, logFile)
	}
}

func TestAppleSystemManager_GetLogs(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a log file with content
	logContent := "2024-01-01 test log output\n"
	logPath := filepath.Join(tmpDir, "test.stdout.log")
	if err := os.WriteFile(logPath, []byte(logContent), 0644); err != nil {
		t.Fatalf("failed to write log file: %v", err)
	}

	// Create a plist that references the log file
	plistWithLog := `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>Label</key>
	<string>com.apple.withlog</string>
	<key>Program</key>
	<string>/usr/libexec/withlog</string>
	<key>StandardOutPath</key>
	<string>` + logPath + `</string>
</dict>
</plist>`

	// Create a plist without log paths
	plistNoLog := `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>Label</key>
	<string>com.apple.nolog</string>
	<key>Program</key>
	<string>/usr/libexec/nolog</string>
</dict>
</plist>`

	// Plist whose StandardOutPath points at a never-created file
	missingLog := filepath.Join(tmpDir, "missing.log")
	plistMissing := `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>Label</key>
	<string>com.apple.missing</string>
	<key>Program</key>
	<string>/usr/libexec/missing</string>
	<key>StandardOutPath</key>
	<string>` + missingLog + `</string>
</dict>
</plist>`

	// Plist referencing an empty log file
	emptyLog := filepath.Join(tmpDir, "empty.log")
	if err := os.WriteFile(emptyLog, []byte(""), 0644); err != nil {
		t.Fatalf("failed to write empty log file: %v", err)
	}
	plistEmpty := `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>Label</key>
	<string>com.apple.empty</string>
	<key>Program</key>
	<string>/usr/libexec/empty</string>
	<key>StandardOutPath</key>
	<string>` + emptyLog + `</string>
</dict>
</plist>`

	if err := os.WriteFile(filepath.Join(tmpDir, "com.apple.withlog.plist"), []byte(plistWithLog), 0644); err != nil {
		t.Fatalf("failed to write plist: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "com.apple.nolog.plist"), []byte(plistNoLog), 0644); err != nil {
		t.Fatalf("failed to write plist: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "com.apple.missing.plist"), []byte(plistMissing), 0644); err != nil {
		t.Fatalf("failed to write plist: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "com.apple.empty.plist"), []byte(plistEmpty), 0644); err != nil {
		t.Fatalf("failed to write plist: %v", err)
	}

	m := &AppleSystemManager{readOnlyManager: readOnlyManager{basePath: tmpDir, serviceType: "apple-system"}}

	t.Run("stdout returns content", func(t *testing.T) {
		got, err := m.GetLogs("com.apple.withlog", "stdout")
		if err != nil {
			t.Fatalf("GetLogs() error = %v", err)
		}
		if got.Status != "ok" {
			t.Errorf("GetLogs() Status = %q, want %q", got.Status, "ok")
		}
		if got.Content != logContent {
			t.Errorf("GetLogs() Content = %q, want %q", got.Content, logContent)
		}
		if got.Path != logPath {
			t.Errorf("GetLogs() Path = %q, want %q", got.Path, logPath)
		}
	})

	t.Run("no log path configured returns no-path status", func(t *testing.T) {
		got, err := m.GetLogs("com.apple.nolog", "stdout")
		if err != nil {
			t.Fatalf("GetLogs() error = %v, want nil", err)
		}
		if got.Status != "no-path" {
			t.Errorf("GetLogs() Status = %q, want %q", got.Status, "no-path")
		}
		if got.Path != "" {
			t.Errorf("GetLogs() Path = %q, want empty", got.Path)
		}
	})

	t.Run("path configured but file missing returns not-found status", func(t *testing.T) {
		got, err := m.GetLogs("com.apple.missing", "stdout")
		if err != nil {
			t.Fatalf("GetLogs() error = %v, want nil", err)
		}
		if got.Status != "not-found" {
			t.Errorf("GetLogs() Status = %q, want %q", got.Status, "not-found")
		}
		if got.Path != missingLog {
			t.Errorf("GetLogs() Path = %q, want %q", got.Path, missingLog)
		}
	})

	t.Run("empty file returns ok status with empty content", func(t *testing.T) {
		got, err := m.GetLogs("com.apple.empty", "stdout")
		if err != nil {
			t.Fatalf("GetLogs() error = %v, want nil", err)
		}
		if got.Status != "ok" {
			t.Errorf("GetLogs() Status = %q, want %q", got.Status, "ok")
		}
		if got.Content != "" {
			t.Errorf("GetLogs() Content = %q, want empty", got.Content)
		}
		if got.Path != emptyLog {
			t.Errorf("GetLogs() Path = %q, want %q", got.Path, emptyLog)
		}
	})

	t.Run("permission denied returns error with path", func(t *testing.T) {
		if os.Geteuid() == 0 {
			t.Skip("running as root; mode 000 does not block reads")
		}

		noPermLog := filepath.Join(tmpDir, "noperm.log")
		if err := os.WriteFile(noPermLog, []byte("secret\n"), 0000); err != nil {
			t.Fatalf("failed to write no-perm log file: %v", err)
		}
		plistNoPerm := `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>Label</key>
	<string>com.apple.noperm</string>
	<key>Program</key>
	<string>/usr/libexec/noperm</string>
	<key>StandardOutPath</key>
	<string>` + noPermLog + `</string>
</dict>
</plist>`
		if err := os.WriteFile(filepath.Join(tmpDir, "com.apple.noperm.plist"), []byte(plistNoPerm), 0644); err != nil {
			t.Fatalf("failed to write plist: %v", err)
		}

		_, err := m.GetLogs("com.apple.noperm", "stdout")
		if err == nil {
			t.Fatal("GetLogs() expected error for permission-denied log file")
		}
		if !strings.Contains(err.Error(), "permission denied") {
			t.Errorf("GetLogs() error = %q, want error containing 'permission denied'", err.Error())
		}
		if !strings.Contains(err.Error(), noPermLog) {
			t.Errorf("GetLogs() error = %q, want it to contain path %q", err.Error(), noPermLog)
		}
	})

	t.Run("invalid log type", func(t *testing.T) {
		_, err := m.GetLogs("com.apple.withlog", "invalid")
		if err == nil {
			t.Fatal("GetLogs() expected error for invalid log type")
		}
		if !strings.Contains(err.Error(), "invalid log type") {
			t.Errorf("GetLogs() error = %q, want error containing 'invalid log type'", err.Error())
		}
	})
}
