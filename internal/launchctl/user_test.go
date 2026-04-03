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
	m := NewUserManager()
	path := m.getLaunchAgentsPath()

	home, _ := os.UserHomeDir()
	expected := filepath.Join(home, "Library", "LaunchAgents")

	if path != expected {
		t.Errorf("getLaunchAgentsPath() = %v, want %v", path, expected)
	}
}

func TestWritePlist_EmptyCalendarInterval(t *testing.T) {
	tmpDir := t.TempDir()
	plistPath := filepath.Join(tmpDir, "test.plist")
	m := &UserManager{}

	// Empty ScheduleConfig (all fields nil) = "every minute"
	config := &ServiceConfig{
		Label:    "com.test.everyminute",
		Program:  "/usr/bin/true",
		Schedule: &ScheduleConfig{},
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
		t.Error("plist should contain StartCalendarInterval even when all fields are empty")
	}
	if strings.Contains(content, "<key>StartInterval</key>") {
		t.Error("plist should NOT contain StartInterval")
	}
}

func TestParseSchedule_MultipleIntervals(t *testing.T) {
	intervals := []interface{}{
		map[string]interface{}{"Hour": uint64(3), "Minute": uint64(0)},
		map[string]interface{}{"Hour": uint64(15), "Minute": uint64(30)},
	}

	schedule := parseSchedule(intervals, 0)
	if schedule == nil {
		t.Fatal("parseSchedule() returned nil")
	}
	if schedule.Hour == nil || *schedule.Hour != 3 {
		t.Errorf("expected Hour=3, got %v", schedule.Hour)
	}
	if !schedule.HasMultiple {
		t.Error("expected HasMultiple=true for array with 2 entries")
	}
}

func TestParseSchedule_SingleArrayInterval(t *testing.T) {
	intervals := []interface{}{
		map[string]interface{}{"Hour": uint64(8)},
	}

	schedule := parseSchedule(intervals, 0)
	if schedule == nil {
		t.Fatal("parseSchedule() returned nil")
	}
	if schedule.HasMultiple {
		t.Error("expected HasMultiple=false for array with 1 entry")
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

	schedule := parseSchedule(pd.StartCalendarInterval, pd.StartInterval)
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


func TestUserManager_Get(t *testing.T) {
	tmpDir := t.TempDir()

	plistContent := `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>Label</key>
	<string>com.test.myapp</string>
	<key>Program</key>
	<string>/usr/local/bin/myapp</string>
	<key>RunAtLoad</key>
	<true/>
	<key>KeepAlive</key>
	<true/>
	<key>EnvironmentVariables</key>
	<dict>
		<key>PATH</key>
		<string>/usr/local/bin</string>
	</dict>
	<key>StandardOutPath</key>
	<string>/tmp/myapp.out.log</string>
	<key>StandardErrorPath</key>
	<string>/tmp/myapp.err.log</string>
	<key>WorkingDirectory</key>
	<string>/tmp</string>
</dict>
</plist>`

	// Write a valid plist file
	if err := os.WriteFile(filepath.Join(tmpDir, "com.test.myapp.plist"), []byte(plistContent), 0644); err != nil {
		t.Fatalf("failed to write plist: %v", err)
	}
	// Write a non-plist file that should be skipped by List
	if err := os.WriteFile(filepath.Join(tmpDir, "README.txt"), []byte("not a plist"), 0644); err != nil {
		t.Fatalf("failed to write non-plist file: %v", err)
	}

	m := &UserManager{launchAgentsPath: tmpDir}

	// Test Get
	service, err := m.Get("com.test.myapp")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if service.Label != "com.test.myapp" {
		t.Errorf("Label = %q, want %q", service.Label, "com.test.myapp")
	}
	if service.Program != "/usr/local/bin/myapp" {
		t.Errorf("Program = %q, want %q", service.Program, "/usr/local/bin/myapp")
	}
	if !service.RunAtLoad {
		t.Error("RunAtLoad should be true")
	}
	if !service.KeepAlive {
		t.Error("KeepAlive should be true")
	}
	if service.Type != "user" {
		t.Errorf("Type = %q, want %q", service.Type, "user")
	}
	if service.ReadOnly {
		t.Error("ReadOnly should be false")
	}
	if service.PlistFormat != "xml" {
		t.Errorf("PlistFormat = %q, want %q", service.PlistFormat, "xml")
	}
	if service.Environment["PATH"] != "/usr/local/bin" {
		t.Errorf("Environment[PATH] = %q, want %q", service.Environment["PATH"], "/usr/local/bin")
	}
	if service.StdoutPath != "/tmp/myapp.out.log" {
		t.Errorf("StdoutPath = %q, want %q", service.StdoutPath, "/tmp/myapp.out.log")
	}
	if service.StderrPath != "/tmp/myapp.err.log" {
		t.Errorf("StderrPath = %q, want %q", service.StderrPath, "/tmp/myapp.err.log")
	}
	if service.WorkingDir != "/tmp" {
		t.Errorf("WorkingDir = %q, want %q", service.WorkingDir, "/tmp")
	}

	// Test List — should return exactly 1 service (non-plist file skipped)
	services, err := m.List()
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(services) != 1 {
		t.Errorf("List() returned %d services, want 1", len(services))
	}

	// Test Get for nonexistent service returns error
	_, err = m.Get("com.test.nonexistent")
	if err == nil {
		t.Error("Get() for nonexistent service should return error")
	}
}

func TestUserManager_Get_KeepAliveDict(t *testing.T) {
	tmpDir := t.TempDir()

	plistContent := `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>Label</key>
	<string>com.test.keepalive</string>
	<key>Program</key>
	<string>/usr/bin/true</string>
	<key>KeepAlive</key>
	<dict>
		<key>SuccessfulExit</key>
		<false/>
	</dict>
</dict>
</plist>`

	if err := os.WriteFile(filepath.Join(tmpDir, "com.test.keepalive.plist"), []byte(plistContent), 0644); err != nil {
		t.Fatalf("failed to write plist: %v", err)
	}

	m := &UserManager{launchAgentsPath: tmpDir}
	service, err := m.Get("com.test.keepalive")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if !service.KeepAlive {
		t.Error("KeepAlive should be true when KeepAlive is a dict")
	}
}


func TestDetectPlistFormat(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		expected string
	}{
		{
			name:     "binary plist",
			data:     []byte("bplist00..."),
			expected: "binary",
		},
		{
			name:     "xml plist",
			data:     []byte(`<?xml version="1.0" encoding="UTF-8"?>`),
			expected: "xml",
		},
		{
			name:     "empty data",
			data:     []byte{},
			expected: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detectPlistFormat(tt.data)
			if result != tt.expected {
				t.Errorf("detectPlistFormat() = %q, want %q", result, tt.expected)
			}
		})
	}
}


func TestUserManager_GetPlist(t *testing.T) {
	tmpDir := t.TempDir()

	plistContent := `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>Label</key>
	<string>com.test.getplist</string>
	<key>Program</key>
	<string>/usr/bin/true</string>
</dict>
</plist>`

	if err := os.WriteFile(filepath.Join(tmpDir, "com.test.getplist.plist"), []byte(plistContent), 0644); err != nil {
		t.Fatalf("failed to write plist: %v", err)
	}

	m := &UserManager{launchAgentsPath: tmpDir}
	content, err := m.GetPlist("com.test.getplist")
	if err != nil {
		t.Fatalf("GetPlist() error = %v", err)
	}
	if content != plistContent {
		t.Errorf("GetPlist() content mismatch.\ngot:\n%s\nwant:\n%s", content, plistContent)
	}
}

func TestUserManager_GetLogs(t *testing.T) {
	tmpDir := t.TempDir()
	logDir := t.TempDir()

	// Create log files
	stdoutLog := filepath.Join(logDir, "stdout.log")
	stderrLog := filepath.Join(logDir, "stderr.log")
	if err := os.WriteFile(stdoutLog, []byte("stdout output line 1\nstdout output line 2\n"), 0644); err != nil {
		t.Fatalf("failed to write stdout log: %v", err)
	}
	if err := os.WriteFile(stderrLog, []byte("stderr output line 1\n"), 0644); err != nil {
		t.Fatalf("failed to write stderr log: %v", err)
	}

	// Create a plist that references the log files with absolute paths
	plistContent := `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>Label</key>
	<string>com.test.logs</string>
	<key>Program</key>
	<string>/usr/bin/true</string>
	<key>StandardOutPath</key>
	<string>` + stdoutLog + `</string>
	<key>StandardErrorPath</key>
	<string>` + stderrLog + `</string>
</dict>
</plist>`

	if err := os.WriteFile(filepath.Join(tmpDir, "com.test.logs.plist"), []byte(plistContent), 0644); err != nil {
		t.Fatalf("failed to write plist: %v", err)
	}

	m := &UserManager{launchAgentsPath: tmpDir}

	// Test stdout log retrieval
	content, err := m.GetLogs("com.test.logs", "stdout")
	if err != nil {
		t.Fatalf("GetLogs(stdout) error = %v", err)
	}
	if !strings.Contains(content, "stdout output line 1") {
		t.Errorf("GetLogs(stdout) missing expected content, got: %q", content)
	}

	// Test stderr log retrieval
	content, err = m.GetLogs("com.test.logs", "stderr")
	if err != nil {
		t.Fatalf("GetLogs(stderr) error = %v", err)
	}
	if !strings.Contains(content, "stderr output line 1") {
		t.Errorf("GetLogs(stderr) missing expected content, got: %q", content)
	}

	// Test with no log path configured
	noLogPlist := `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>Label</key>
	<string>com.test.nologs</string>
	<key>Program</key>
	<string>/usr/bin/true</string>
</dict>
</plist>`

	if err := os.WriteFile(filepath.Join(tmpDir, "com.test.nologs.plist"), []byte(noLogPlist), 0644); err != nil {
		t.Fatalf("failed to write plist: %v", err)
	}

	_, err = m.GetLogs("com.test.nologs", "stdout")
	if err == nil {
		t.Error("GetLogs() should return error when no log path configured")
	}

	// Test with invalid log type
	_, err = m.GetLogs("com.test.logs", "debug")
	if err == nil {
		t.Fatal("GetLogs() should return error for invalid log type 'debug'")
	}
	if !strings.Contains(err.Error(), "invalid log type") {
		t.Errorf("GetLogs() error = %q, want it to contain 'invalid log type'", err.Error())
	}
}


func TestUserManager_CRUD(t *testing.T) {
	tmpDir := t.TempDir()
	m := &UserManager{launchAgentsPath: tmpDir}

	// Create with valid config
	config := &ServiceConfig{
		Label:   "com.test.crud",
		Program: "/usr/bin/true",
	}
	if err := m.Create(config); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	// Verify file exists
	plistPath := filepath.Join(tmpDir, "com.test.crud.plist")
	if _, err := os.Stat(plistPath); os.IsNotExist(err) {
		t.Fatal("Create() did not create plist file")
	}

	// Create with empty label → error
	emptyConfig := &ServiceConfig{
		Label:   "",
		Program: "/usr/bin/true",
	}
	err := m.Create(emptyConfig)
	if err == nil {
		t.Fatal("Create() with empty label should return error")
	}
	if !strings.Contains(err.Error(), "service label is required") {
		t.Errorf("Create() error = %q, want it to contain 'service label is required'", err.Error())
	}

	// Create with duplicate label → error
	err = m.Create(config)
	if err == nil {
		t.Fatal("Create() with duplicate label should return error")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("Create() error = %q, want it to contain 'already exists'", err.Error())
	}

	// Delete existing service
	if err := m.Delete("com.test.crud"); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if _, err := os.Stat(plistPath); !os.IsNotExist(err) {
		t.Error("Delete() did not remove plist file")
	}

	// Delete nonexistent service → error
	err = m.Delete("com.test.nonexistent")
	if err == nil {
		t.Fatal("Delete() nonexistent should return error")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("Delete() error = %q, want it to contain 'not found'", err.Error())
	}
}


func TestParseSchedule_SingleDict(t *testing.T) {
	// Single dict (not wrapped in an array)
	dict := map[string]interface{}{
		"Hour":   uint64(9),
		"Minute": uint64(0),
	}

	schedule := parseSchedule(dict, 0)
	if schedule == nil {
		t.Fatal("parseSchedule() returned nil for single dict")
	}
	if schedule.Hour == nil || *schedule.Hour != 9 {
		t.Errorf("Hour = %v, want 9", schedule.Hour)
	}
	if schedule.Minute == nil || *schedule.Minute != 0 {
		t.Errorf("Minute = %v, want 0", schedule.Minute)
	}
	if schedule.HasMultiple {
		t.Error("HasMultiple should be false for single dict")
	}
}

func TestParseSchedule_NilAndZero(t *testing.T) {
	// Both nil and 0 → nil result
	schedule := parseSchedule(nil, 0)
	if schedule != nil {
		t.Errorf("parseSchedule(nil, 0) = %v, want nil", schedule)
	}
}

func TestValidateSchedule_Comprehensive(t *testing.T) {
	tests := []struct {
		name      string
		schedule  *ScheduleConfig
		wantError bool
	}{
		{
			name:      "nil schedule is valid",
			schedule:  nil,
			wantError: false,
		},
		{
			name:      "valid interval",
			schedule:  &ScheduleConfig{Interval: intPtr(60)},
			wantError: false,
		},
		{
			name:      "interval too small",
			schedule:  &ScheduleConfig{Interval: intPtr(5)},
			wantError: true,
		},
		{
			name:      "interval exactly 10 is valid",
			schedule:  &ScheduleConfig{Interval: intPtr(10)},
			wantError: false,
		},
		{
			name:      "interval of 9 is invalid",
			schedule:  &ScheduleConfig{Interval: intPtr(9)},
			wantError: true,
		},
		{
			name:      "calendar schedule with hour only",
			schedule:  &ScheduleConfig{Hour: intPtr(3)},
			wantError: false,
		},
		{
			name:      "empty schedule (every minute)",
			schedule:  &ScheduleConfig{},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateSchedule(tt.schedule)
			if (err != nil) != tt.wantError {
				t.Errorf("validateSchedule() error = %v, wantError = %v", err, tt.wantError)
			}
		})
	}
}

func TestWritePlist_WakeSystem(t *testing.T) {
	tmpDir := t.TempDir()
	plistPath := filepath.Join(tmpDir, "test.plist")
	m := &UserManager{}

	config := &ServiceConfig{
		Label:      "com.test.wake",
		Program:    "/usr/bin/true",
		WakeSystem: true,
		Schedule:   &ScheduleConfig{Hour: intPtr(3)},
	}

	if err := m.writePlist(plistPath, config); err != nil {
		t.Fatalf("writePlist() error = %v", err)
	}

	data, err := os.ReadFile(plistPath)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "<key>WakeSystem</key>") {
		t.Error("plist should contain WakeSystem key when WakeSystem is true")
	}
}

func TestWritePlist_WakeSystemFalse(t *testing.T) {
	tmpDir := t.TempDir()
	plistPath := filepath.Join(tmpDir, "test.plist")
	m := &UserManager{}

	config := &ServiceConfig{
		Label:      "com.test.nowake",
		Program:    "/usr/bin/true",
		WakeSystem: false,
	}

	if err := m.writePlist(plistPath, config); err != nil {
		t.Fatalf("writePlist() error = %v", err)
	}

	data, err := os.ReadFile(plistPath)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	content := string(data)
	if strings.Contains(content, "<key>WakeSystem</key>") {
		t.Error("plist should NOT contain WakeSystem key when WakeSystem is false")
	}
}

func TestUserManager_Get_WakeSystemTrue(t *testing.T) {
	tmpDir := t.TempDir()

	plistContent := `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>Label</key>
	<string>com.test.wake</string>
	<key>Program</key>
	<string>/usr/bin/true</string>
	<key>WakeSystem</key>
	<true/>
</dict>
</plist>`

	if err := os.WriteFile(filepath.Join(tmpDir, "com.test.wake.plist"), []byte(plistContent), 0644); err != nil {
		t.Fatalf("failed to write plist: %v", err)
	}

	m := &UserManager{launchAgentsPath: tmpDir}
	service, err := m.Get("com.test.wake")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if !service.WakeSystem {
		t.Error("WakeSystem should be true when plist contains WakeSystem key with true value")
	}
}

func TestUserManager_Get_WakeSystemAbsent(t *testing.T) {
	tmpDir := t.TempDir()

	plistContent := `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>Label</key>
	<string>com.test.nowake</string>
	<key>Program</key>
	<string>/usr/bin/true</string>
</dict>
</plist>`

	if err := os.WriteFile(filepath.Join(tmpDir, "com.test.nowake.plist"), []byte(plistContent), 0644); err != nil {
		t.Fatalf("failed to write plist: %v", err)
	}

	m := &UserManager{launchAgentsPath: tmpDir}
	service, err := m.Get("com.test.nowake")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if service.WakeSystem {
		t.Error("WakeSystem should be false when plist does not contain WakeSystem key")
	}
}

func TestUserManager_RoundTrip_WakeSystemDisable(t *testing.T) {
	tmpDir := t.TempDir()
	m := &UserManager{launchAgentsPath: tmpDir}

	// Create a service with WakeSystem enabled
	config := &ServiceConfig{
		Label:      "com.test.roundtrip",
		Program:    "/usr/bin/true",
		WakeSystem: true,
		Schedule:   &ScheduleConfig{Hour: intPtr(3)},
	}
	plistPath := filepath.Join(tmpDir, "com.test.roundtrip.plist")
	if err := m.writePlist(plistPath, config); err != nil {
		t.Fatalf("writePlist() error = %v", err)
	}

	// Read it back and verify WakeSystem is true
	service, err := m.Get("com.test.roundtrip")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if !service.WakeSystem {
		t.Fatal("WakeSystem should be true after initial write")
	}

	// Update with WakeSystem disabled
	config.WakeSystem = false
	if err := m.writePlist(plistPath, config); err != nil {
		t.Fatalf("writePlist() error = %v", err)
	}

	// Verify WakeSystem key is absent from the plist
	data, err := os.ReadFile(plistPath)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if strings.Contains(string(data), "<key>WakeSystem</key>") {
		t.Error("plist should NOT contain WakeSystem key after disabling")
	}

	// Read back and verify Service.WakeSystem is false
	service, err = m.Get("com.test.roundtrip")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if service.WakeSystem {
		t.Error("WakeSystem should be false after update")
	}
}

// intPtr is a helper to create *int from a literal
func intPtr(v int) *int {
	return &v
}
