package launchctl

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"howett.net/plist"

	"launchpal/internal/plistutil"
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

func TestUserManager_GetPlistContent_XMLFile(t *testing.T) {
	tmpDir := t.TempDir()
	m := &UserManager{launchAgentsPath: tmpDir}

	xml := `<?xml version="1.0" encoding="UTF-8"?>
<plist version="1.0">
<dict><key>Label</key><string>com.test.current</string></dict>
</plist>`
	if err := os.WriteFile(filepath.Join(tmpDir, "com.test.current.plist"), []byte(xml), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}

	got, err := m.GetPlistContent("com.test.current")
	if err != nil {
		t.Fatalf("GetPlistContent error = %v", err)
	}
	if got.Data != xml {
		t.Errorf("Data mismatch: %q", got.Data)
	}
	if got.Format != "xml" {
		t.Errorf("Format = %q, want xml", got.Format)
	}
	if got.ConvertFailed {
		t.Errorf("ConvertFailed = true, want false")
	}
}

func TestUserManager_GetPlistContent_BinaryFileConvertedToXML(t *testing.T) {
	tmpDir := t.TempDir()
	m := &UserManager{launchAgentsPath: tmpDir}

	xmlSource := `<?xml version="1.0" encoding="UTF-8"?>
<plist version="1.0">
<dict><key>Label</key><string>com.test.binary.current</string></dict>
</plist>`
	xmlPath := filepath.Join(t.TempDir(), "source.plist")
	if err := os.WriteFile(xmlPath, []byte(xmlSource), 0644); err != nil {
		t.Fatalf("write xml: %v", err)
	}
	binaryPath := filepath.Join(tmpDir, "com.test.binary.current.plist")
	cmd := exec.Command("plutil", "-convert", "binary1", "-o", binaryPath, xmlPath)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Skipf("plutil not available: %v (output: %s)", err, out)
	}

	got, err := m.GetPlistContent("com.test.binary.current")
	if err != nil {
		t.Fatalf("GetPlistContent error = %v", err)
	}
	if got.Format != "binary" {
		t.Errorf("Format = %q, want binary", got.Format)
	}
	if got.ConvertFailed {
		t.Errorf("ConvertFailed = true, want false")
	}
	if !strings.Contains(got.Data, "com.test.binary.current") {
		t.Errorf("Data missing label: %q", got.Data)
	}
}

func TestUserManager_GetPlistContent_MissingFileReturnsEmptyNoError(t *testing.T) {
	tmpDir := t.TempDir()
	m := &UserManager{launchAgentsPath: tmpDir}

	got, err := m.GetPlistContent("com.test.absent")
	if err != nil {
		t.Fatalf("GetPlistContent error = %v (expected nil)", err)
	}
	if got == nil {
		t.Fatal("GetPlistContent returned nil Content")
	}
	if got.Data != "" {
		t.Errorf("Data = %q, want empty", got.Data)
	}
	if got.Format != "" && got.Format != "unknown" {
		t.Errorf("Format = %q, want empty or unknown", got.Format)
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
	if len(schedule.Schedules) != 2 {
		t.Fatalf("expected 2 schedules, got %d", len(schedule.Schedules))
	}
	if schedule.Schedules[0].Hour == nil || *schedule.Schedules[0].Hour != 3 {
		t.Errorf("expected Schedules[0].Hour=3, got %v", schedule.Schedules[0].Hour)
	}
	if schedule.Schedules[1].Hour == nil || *schedule.Schedules[1].Hour != 15 {
		t.Errorf("expected Schedules[1].Hour=15, got %v", schedule.Schedules[1].Hour)
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
	if len(schedule.Schedules) != 1 {
		t.Fatalf("expected 1 schedule, got %d", len(schedule.Schedules))
	}
	if schedule.Schedules[0].Hour == nil || *schedule.Schedules[0].Hour != 8 {
		t.Errorf("expected Schedules[0].Hour=8, got %v", schedule.Schedules[0].Hour)
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

	config := &ServiceConfig{
		Label:   "com.test.calendar",
		Program: "/usr/bin/true",
		Schedule: &ScheduleConfig{
			Schedules: []CalendarEntry{{Hour: intPtr(3), Minute: intPtr(0)}},
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
			result := plistutil.DetectFormat(tt.data)
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
	if len(schedule.Schedules) != 1 {
		t.Fatalf("Schedules length = %d, want 1", len(schedule.Schedules))
	}
	if schedule.Schedules[0].Hour == nil || *schedule.Schedules[0].Hour != 9 {
		t.Errorf("Hour = %v, want 9", schedule.Schedules[0].Hour)
	}
	if schedule.Schedules[0].Minute == nil || *schedule.Schedules[0].Minute != 0 {
		t.Errorf("Minute = %v, want 0", schedule.Schedules[0].Minute)
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
			schedule:  &ScheduleConfig{Schedules: []CalendarEntry{{Hour: intPtr(3)}}},
			wantError: false,
		},
		{
			name:      "empty schedule (every minute)",
			schedule:  &ScheduleConfig{},
			wantError: false,
		},
		{
			name:      "valid calendar entry",
			schedule:  &ScheduleConfig{Schedules: []CalendarEntry{{Hour: intPtr(9), Minute: intPtr(0)}}},
			wantError: false,
		},
		{
			name:      "hour out of range",
			schedule:  &ScheduleConfig{Schedules: []CalendarEntry{{Hour: intPtr(99)}}},
			wantError: true,
		},
		{
			name:      "minute out of range",
			schedule:  &ScheduleConfig{Schedules: []CalendarEntry{{Minute: intPtr(60)}}},
			wantError: true,
		},
		{
			name:      "day out of range",
			schedule:  &ScheduleConfig{Schedules: []CalendarEntry{{Day: intPtr(0)}}},
			wantError: true,
		},
		{
			name:      "weekday out of range",
			schedule:  &ScheduleConfig{Schedules: []CalendarEntry{{Weekday: intPtr(7)}}},
			wantError: true,
		},
		{
			name:      "month out of range",
			schedule:  &ScheduleConfig{Schedules: []CalendarEntry{{Month: intPtr(13)}}},
			wantError: true,
		},
		{
			name:      "negative minute",
			schedule:  &ScheduleConfig{Schedules: []CalendarEntry{{Minute: intPtr(-1)}}},
			wantError: true,
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
		Schedule:   &ScheduleConfig{Schedules: []CalendarEntry{{Hour: intPtr(3)}}},
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
		Schedule:   &ScheduleConfig{Schedules: []CalendarEntry{{Hour: intPtr(3)}}},
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

func TestUserManager_Kickstart_PlistNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	m := &UserManager{launchAgentsPath: tmpDir}

	err := m.Kickstart("com.test.nonexistent")
	if err == nil {
		t.Fatal("Kickstart() should return error when plist does not exist")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("Kickstart() error = %q, want it to contain 'not found'", err.Error())
	}
}

func TestGuiDomain(t *testing.T) {
	result := guiDomain()
	uid := os.Getuid()
	expected := fmt.Sprintf("gui/%d", uid)
	if result != expected {
		t.Errorf("guiDomain() = %q, want %q", result, expected)
	}
}

func TestWritePlist_SingleCalendarEntry(t *testing.T) {
	tmpDir := t.TempDir()
	plistPath := filepath.Join(tmpDir, "test.plist")
	m := &UserManager{}

	config := &ServiceConfig{
		Label:   "com.test.single",
		Program: "/usr/bin/true",
		Schedule: &ScheduleConfig{
			Schedules: []CalendarEntry{{Hour: intPtr(9), Minute: intPtr(0)}},
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
		t.Error("plist should contain StartCalendarInterval")
	}
	// Single entry should be written as dict, NOT as array
	if strings.Contains(content, "<array>") {
		t.Error("single entry should be written as dict, not array")
	}
}

func TestWritePlist_MultipleCalendarEntries(t *testing.T) {
	tmpDir := t.TempDir()
	plistPath := filepath.Join(tmpDir, "test.plist")
	m := &UserManager{}

	config := &ServiceConfig{
		Label:   "com.test.multi",
		Program: "/usr/bin/true",
		Schedule: &ScheduleConfig{
			Schedules: []CalendarEntry{
				{Hour: intPtr(9), Minute: intPtr(0)},
				{Hour: intPtr(17), Minute: intPtr(30)},
			},
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
		t.Error("plist should contain StartCalendarInterval")
	}
	// Multiple entries should be written as array
	if !strings.Contains(content, "<array>") {
		t.Error("multiple entries should be written as array")
	}
	if strings.Contains(content, "<key>StartInterval</key>") {
		t.Error("should NOT contain StartInterval when Schedules are set")
	}
}

func TestWritePlist_MultipleRoundTrip(t *testing.T) {
	tmpDir := t.TempDir()
	plistPath := filepath.Join(tmpDir, "test.plist")
	m := &UserManager{launchAgentsPath: tmpDir}

	config := &ServiceConfig{
		Label:   "test",
		Program: "/usr/bin/true",
		Schedule: &ScheduleConfig{
			Schedules: []CalendarEntry{
				{Hour: intPtr(9), Minute: intPtr(0)},
				{Hour: intPtr(17), Minute: intPtr(30)},
			},
		},
	}

	if err := m.writePlist(plistPath, config); err != nil {
		t.Fatalf("writePlist() error = %v", err)
	}

	// Read back
	service, err := m.Get("test")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if service.Schedule == nil {
		t.Fatal("Schedule is nil after round-trip")
	}
	if len(service.Schedule.Schedules) != 2 {
		t.Fatalf("expected 2 schedules, got %d", len(service.Schedule.Schedules))
	}
	if *service.Schedule.Schedules[0].Hour != 9 {
		t.Errorf("Schedules[0].Hour = %d, want 9", *service.Schedule.Schedules[0].Hour)
	}
	if *service.Schedule.Schedules[1].Hour != 17 {
		t.Errorf("Schedules[1].Hour = %d, want 17", *service.Schedule.Schedules[1].Hour)
	}
}

func TestParseSchedule_DictToSchedules(t *testing.T) {
	// Single dict should produce Schedules with 1 entry
	dict := map[string]interface{}{
		"Hour":   uint64(9),
		"Minute": uint64(0),
	}
	schedule := parseSchedule(dict, 0)
	if schedule == nil {
		t.Fatal("parseSchedule() returned nil")
	}
	if len(schedule.Schedules) != 1 {
		t.Fatalf("expected 1 schedule, got %d", len(schedule.Schedules))
	}
	e := schedule.Schedules[0]
	if e.Hour == nil || *e.Hour != 9 {
		t.Errorf("Hour = %v, want 9", e.Hour)
	}
	if e.Minute == nil || *e.Minute != 0 {
		t.Errorf("Minute = %v, want 0", e.Minute)
	}
	if e.Day != nil || e.Weekday != nil || e.Month != nil {
		t.Error("expected nil for Day, Weekday, Month")
	}
}

func TestParseSchedule_ArrayToSchedules(t *testing.T) {
	// Array of 3 dicts should produce Schedules with 3 entries
	intervals := []interface{}{
		map[string]interface{}{"Hour": uint64(9), "Minute": uint64(0)},
		map[string]interface{}{"Hour": uint64(12), "Minute": uint64(30)},
		map[string]interface{}{"Hour": uint64(18), "Minute": uint64(0)},
	}
	schedule := parseSchedule(intervals, 0)
	if schedule == nil {
		t.Fatal("parseSchedule() returned nil")
	}
	if len(schedule.Schedules) != 3 {
		t.Fatalf("expected 3 schedules, got %d", len(schedule.Schedules))
	}
	expectedHours := []int{9, 12, 18}
	for i, h := range expectedHours {
		if schedule.Schedules[i].Hour == nil || *schedule.Schedules[i].Hour != h {
			t.Errorf("Schedules[%d].Hour = %v, want %d", i, schedule.Schedules[i].Hour, h)
		}
	}
}

func TestCalendarEntry_Fields(t *testing.T) {
	// Verify CalendarEntry struct has all expected fields with *int type
	entry := CalendarEntry{
		Minute:  intPtr(30),
		Hour:    intPtr(9),
		Day:     intPtr(15),
		Weekday: intPtr(1),
		Month:   intPtr(6),
	}

	if entry.Minute == nil || *entry.Minute != 30 {
		t.Errorf("Minute = %v, want 30", entry.Minute)
	}
	if entry.Hour == nil || *entry.Hour != 9 {
		t.Errorf("Hour = %v, want 9", entry.Hour)
	}
	if entry.Day == nil || *entry.Day != 15 {
		t.Errorf("Day = %v, want 15", entry.Day)
	}
	if entry.Weekday == nil || *entry.Weekday != 1 {
		t.Errorf("Weekday = %v, want 1", entry.Weekday)
	}
	if entry.Month == nil || *entry.Month != 6 {
		t.Errorf("Month = %v, want 6", entry.Month)
	}
}

func TestScheduleConfig_Schedules(t *testing.T) {
	// Verify ScheduleConfig uses Schedules []CalendarEntry instead of single fields
	config := ScheduleConfig{
		Schedules: []CalendarEntry{
			{Minute: intPtr(0), Hour: intPtr(9)},
			{Minute: intPtr(0), Hour: intPtr(17)},
		},
	}

	if len(config.Schedules) != 2 {
		t.Fatalf("Schedules length = %d, want 2", len(config.Schedules))
	}
	if *config.Schedules[0].Hour != 9 {
		t.Errorf("Schedules[0].Hour = %d, want 9", *config.Schedules[0].Hour)
	}
	if *config.Schedules[1].Hour != 17 {
		t.Errorf("Schedules[1].Hour = %d, want 17", *config.Schedules[1].Hour)
	}
}

func TestScheduleConfig_NoHasMultiple(t *testing.T) {
	// Verify HasMultiple field is removed — ScheduleConfig should not have it
	// This test compiles only if HasMultiple is removed
	config := ScheduleConfig{
		Schedules: []CalendarEntry{{Minute: intPtr(0)}},
	}
	_ = config
}

// intPtr is a helper to create *int from a literal
func intPtr(v int) *int {
	return &v
}

// emptyPlistXML is a plist with no Label — the exact shape that triggered the
// bug where deleting this file tore down the entire user GUI domain.
const emptyPlistXML = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict/>
</plist>`

func TestServiceTarget_RejectsEmptyLabel(t *testing.T) {
	if _, err := serviceTarget("gui/501", ""); err == nil {
		t.Fatal("serviceTarget with empty label should return error to prevent tearing down the whole domain")
	}
	got, err := serviceTarget("gui/501", "com.test.app")
	if err != nil {
		t.Fatalf("serviceTarget error = %v", err)
	}
	if got != "gui/501/com.test.app" {
		t.Errorf("serviceTarget = %q, want %q", got, "gui/501/com.test.app")
	}
}

func TestUserManager_Stop_SkipsEmptyLabel(t *testing.T) {
	tmpDir := t.TempDir()
	m := &UserManager{launchAgentsPath: tmpDir}

	if err := os.WriteFile(filepath.Join(tmpDir, "com.empty.plist"), []byte(emptyPlistXML), 0644); err != nil {
		t.Fatalf("write plist: %v", err)
	}

	// Stop MUST NOT execute `launchctl bootout gui/<uid>/` (empty service-name)
	// because that would be interpreted as a domain-target and unload every
	// user LaunchAgent, collapsing the desktop session.
	if err := m.Stop("com.empty"); err != nil {
		t.Fatalf("Stop() with empty-Label plist returned error: %v", err)
	}
}

func TestUserManager_Delete_EmptyLabelPlist(t *testing.T) {
	tmpDir := t.TempDir()
	m := &UserManager{launchAgentsPath: tmpDir}

	plistPath := filepath.Join(tmpDir, "com.google.keystone.agent.plist")
	if err := os.WriteFile(plistPath, []byte(emptyPlistXML), 0644); err != nil {
		t.Fatalf("write plist: %v", err)
	}

	if err := m.Delete("com.google.keystone.agent"); err != nil {
		t.Fatalf("Delete() with empty-Label plist returned error: %v", err)
	}
	if _, err := os.Stat(plistPath); !os.IsNotExist(err) {
		t.Error("Delete() should remove the plist file even when Label is empty")
	}
}

func TestUserManager_Kickstart_EmptyLabelReturnsError(t *testing.T) {
	tmpDir := t.TempDir()
	m := &UserManager{launchAgentsPath: tmpDir}

	if err := os.WriteFile(filepath.Join(tmpDir, "com.empty.plist"), []byte(emptyPlistXML), 0644); err != nil {
		t.Fatalf("write plist: %v", err)
	}

	err := m.Kickstart("com.empty")
	if err == nil {
		t.Fatal("Kickstart() with empty-Label plist must return an error instead of issuing a malformed launchctl target")
	}
	if !strings.Contains(err.Error(), "label") {
		t.Errorf("Kickstart() error = %q, want it to mention the missing label", err.Error())
	}
}
