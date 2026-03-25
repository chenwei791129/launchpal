package launchctl

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"howett.net/plist"
)

func TestUserManager_List(t *testing.T) {
	m := NewUserManager()
	services, err := m.List()
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if services == nil {
		t.Error("List() returned nil, expected empty slice")
	}
	t.Logf("Found %d services", len(services))
}

func TestUserManager_GetLaunchAgentsPath(t *testing.T) {
	m := &UserManager{}
	path := m.getLaunchAgentsPath()

	home, _ := os.UserHomeDir()
	expected := filepath.Join(home, "Library", "LaunchAgents")

	if path != expected {
		t.Errorf("getLaunchAgentsPath() = %v, want %v", path, expected)
	}
}

func TestWritePlist_StartInterval(t *testing.T) {
	tmpDir := t.TempDir()
	plistPath := filepath.Join(tmpDir, "test.plist")
	m := &UserManager{}

	interval := 3600
	config := &ServiceConfig{
		Label:   "com.test.interval",
		Program: "/usr/bin/true",
		Schedule: &ScheduleConfig{
			Interval: &interval,
		},
	}

	if err := m.writePlist(plistPath, config); err != nil {
		t.Fatalf("writePlist() error = %v", err)
	}

	data, err := os.ReadFile(plistPath)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "<key>StartInterval</key>") {
		t.Error("plist should contain StartInterval key")
	}
	if !strings.Contains(content, "<integer>3600</integer>") {
		t.Error("plist should contain interval value 3600")
	}
	if strings.Contains(content, "<key>StartCalendarInterval</key>") {
		t.Error("plist should NOT contain StartCalendarInterval when StartInterval is set")
	}
}

func TestWritePlist_CalendarInterval_NotStartInterval(t *testing.T) {
	tmpDir := t.TempDir()
	plistPath := filepath.Join(tmpDir, "test.plist")
	m := &UserManager{}

	hour := 3
	minute := 0
	config := &ServiceConfig{
		Label:   "com.test.calendar",
		Program: "/usr/bin/true",
		Schedule: &ScheduleConfig{
			Hour:   &hour,
			Minute: &minute,
		},
	}

	if err := m.writePlist(plistPath, config); err != nil {
		t.Fatalf("writePlist() error = %v", err)
	}

	data, err := os.ReadFile(plistPath)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "<key>StartCalendarInterval</key>") {
		t.Error("plist should contain StartCalendarInterval key")
	}
	if strings.Contains(content, "<key>StartInterval</key>") {
		t.Error("plist should NOT contain StartInterval when CalendarInterval is set")
	}
}

func TestParseSchedule_StartInterval(t *testing.T) {
	m := &UserManager{}

	// Simulate reading a plist with StartInterval
	plistXML := `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>Label</key>
	<string>com.test.interval</string>
	<key>Program</key>
	<string>/usr/bin/true</string>
	<key>StartInterval</key>
	<integer>1800</integer>
</dict>
</plist>`

	var pd plistData
	if _, err := plist.Unmarshal([]byte(plistXML), &pd); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	schedule := m.parseSchedule(pd.StartCalendarInterval, pd.StartInterval)
	if schedule == nil {
		t.Fatal("parseSchedule() returned nil, expected schedule with Interval")
	}
	if schedule.Interval == nil {
		t.Fatal("schedule.Interval is nil, expected 1800")
	}
	if *schedule.Interval != 1800 {
		t.Errorf("schedule.Interval = %d, want 1800", *schedule.Interval)
	}
}
