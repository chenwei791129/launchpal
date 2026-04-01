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

	if err := os.WriteFile(filepath.Join(tmpDir, "com.apple.withlog.plist"), []byte(plistWithLog), 0644); err != nil {
		t.Fatalf("failed to write plist: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "com.apple.nolog.plist"), []byte(plistNoLog), 0644); err != nil {
		t.Fatalf("failed to write plist: %v", err)
	}

	m := &AppleSystemManager{readOnlyManager: readOnlyManager{basePath: tmpDir, serviceType: "apple-system"}}

	t.Run("stdout returns content", func(t *testing.T) {
		content, err := m.GetLogs("com.apple.withlog", "stdout")
		if err != nil {
			t.Fatalf("GetLogs() error = %v", err)
		}
		if content != logContent {
			t.Errorf("GetLogs() = %q, want %q", content, logContent)
		}
	})

	t.Run("no log path configured", func(t *testing.T) {
		_, err := m.GetLogs("com.apple.nolog", "stdout")
		if err == nil {
			t.Fatal("GetLogs() expected error for service with no log path")
		}
		if !strings.Contains(err.Error(), "no stdout log path configured") {
			t.Errorf("GetLogs() error = %q, want error containing 'no stdout log path configured'", err.Error())
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

