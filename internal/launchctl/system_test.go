package launchctl

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

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

	m := &SystemManager{launchDaemonsPath: tmpDir}

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

		m := &SystemManager{launchDaemonsPath: tmpDir}
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

		m := &SystemManager{launchDaemonsPath: tmpDir}
		services, err := m.List()
		if err != nil {
			t.Fatalf("List() error = %v", err)
		}

		if len(services) != 0 {
			t.Errorf("List() returned %d services, want 0", len(services))
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

	m := &SystemManager{launchDaemonsPath: tmpDir}

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
