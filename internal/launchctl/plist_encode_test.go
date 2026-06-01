package launchctl

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"howett.net/plist"
)

func TestBuildPlistDict_OmitsProgramArgumentsWhenEmpty(t *testing.T) {
	cfg := &ServiceConfig{Label: "com.example", Program: "/usr/bin/true"}
	pd := BuildPlistDict(cfg, false)
	if _, ok := pd["ProgramArguments"]; ok {
		t.Errorf("ProgramArguments should be omitted when no arguments supplied")
	}
	if prog := pd["Program"]; prog != "/usr/bin/true" {
		t.Errorf("Program = %v", prog)
	}
}

func TestBuildPlistDict_KeepsArgumentsWhenNoProgram(t *testing.T) {
	// When the user only supplies Arguments (without Program), launchd uses
	// ProgramArguments[0] as the executable. BuildPlistDict must leave that
	// array untouched.
	cfg := &ServiceConfig{Label: "com.example", Arguments: []string{"/bin/sh", "-c", "echo hi"}}
	pd := BuildPlistDict(cfg, false)
	if _, ok := pd["Program"]; ok {
		t.Errorf("Program should be omitted")
	}
	args := pd["ProgramArguments"].([]string)
	if args[0] != "/bin/sh" {
		t.Errorf("argv[0] = %q, want /bin/sh", args[0])
	}
}

// TestBuildPlistDict_EncodesValidPlist round-trips the dict through the
// plist encoder and re-parses to confirm the ProgramArguments key actually
// survives marshaling.
func TestBuildPlistDict_EncodesValidPlist(t *testing.T) {
	cfg := &ServiceConfig{
		Label:     "com.example.sh",
		Arguments: []string{"/bin/sh", "-c", "echo hi"},
		RunAtLoad: true,
	}
	pd := BuildPlistDict(cfg, false)

	var sb strings.Builder
	enc := plist.NewEncoder(&sb)
	enc.Indent("\t")
	if err := enc.Encode(pd); err != nil {
		t.Fatalf("encode: %v", err)
	}
	xml := sb.String()
	for _, want := range []string{
		"<string>/bin/sh</string>",
		"<string>-c</string>",
		"<key>ProgramArguments</key>",
	} {
		if !strings.Contains(xml, want) {
			t.Errorf("plist missing %q\n%s", want, xml)
		}
	}
}

// TestBuildPlistDict_MinimalConfig covers the spec scenario "Minimal service
// config": only Label and Program are present, nothing else (no KeepAlive,
// no RunAtLoad, no ThrottleInterval).
func TestBuildPlistDict_MinimalConfig(t *testing.T) {
	cfg := &ServiceConfig{Label: "com.example", Program: "/usr/bin/true"}
	pd := BuildPlistDict(cfg, false)

	for _, unexpected := range []string{"KeepAlive", "RunAtLoad", "ThrottleInterval", "ProgramArguments", "StartInterval"} {
		if _, ok := pd[unexpected]; ok {
			t.Errorf("minimal config should omit %q, got %v", unexpected, pd[unexpected])
		}
	}
	if len(pd) != 2 {
		t.Errorf("minimal config should write exactly Label+Program, got %d keys: %v", len(pd), pd)
	}
}

// TestBuildPlistDict_KeepAlive exercises every KeepAlive encode branch from the
// "Write plist from ServiceConfig" spec scenarios.
func TestBuildPlistDict_KeepAlive(t *testing.T) {
	base := func(ka KeepAliveConfig) *ServiceConfig {
		return &ServiceConfig{Label: "com.example", Program: "/usr/bin/true", KeepAlive: ka}
	}

	t.Run("disabled omits the key", func(t *testing.T) {
		pd := BuildPlistDict(base(KeepAliveConfig{}), false)
		if _, ok := pd["KeepAlive"]; ok {
			t.Errorf("KeepAlive should be omitted when disabled, got %v", pd["KeepAlive"])
		}
	})

	t.Run("boolean mode writes true", func(t *testing.T) {
		pd := BuildPlistDict(base(KeepAliveConfig{Enabled: true, Mode: KeepAliveModeBoolean}), false)
		if v, ok := pd["KeepAlive"].(bool); !ok || v != true {
			t.Errorf("KeepAlive = %v (%T), want bool true", pd["KeepAlive"], pd["KeepAlive"])
		}
	})

	t.Run("dictionary with sub-key", func(t *testing.T) {
		pd := BuildPlistDict(base(KeepAliveConfig{
			Enabled: true, Mode: KeepAliveModeDictionary, SuccessfulExit: boolPtr(false),
		}), false)
		dict, ok := pd["KeepAlive"].(map[string]any)
		if !ok {
			t.Fatalf("KeepAlive = %v (%T), want dictionary", pd["KeepAlive"], pd["KeepAlive"])
		}
		if dict["SuccessfulExit"] != false {
			t.Errorf("SuccessfulExit = %v, want false", dict["SuccessfulExit"])
		}
	})

	t.Run("dictionary preserves PathState", func(t *testing.T) {
		pd := BuildPlistDict(base(KeepAliveConfig{
			Enabled: true, Mode: KeepAliveModeDictionary,
			PathState: map[string]bool{"/tmp/flag": true},
		}), false)
		dict := pd["KeepAlive"].(map[string]any)
		ps, ok := dict["PathState"].(map[string]any)
		if !ok {
			t.Fatalf("PathState = %v (%T), want nested dictionary", dict["PathState"], dict["PathState"])
		}
		if ps["/tmp/flag"] != true {
			t.Errorf("PathState[/tmp/flag] = %v, want true", ps["/tmp/flag"])
		}
	})

	t.Run("dictionary with no sub-key downgrades to boolean true", func(t *testing.T) {
		pd := BuildPlistDict(base(KeepAliveConfig{Enabled: true, Mode: KeepAliveModeDictionary}), false)
		if v, ok := pd["KeepAlive"].(bool); !ok || v != true {
			t.Errorf("KeepAlive = %v (%T), want bool true (no empty dict)", pd["KeepAlive"], pd["KeepAlive"])
		}
	})
}

// TestBuildPlistDict_ThrottleInterval covers the ThrottleInterval set/unset
// spec scenarios.
func TestBuildPlistDict_ThrottleInterval(t *testing.T) {
	t.Run("set writes the value", func(t *testing.T) {
		ten := 10
		pd := BuildPlistDict(&ServiceConfig{Label: "com.example", Program: "/usr/bin/true", ThrottleInterval: &ten}, false)
		if pd["ThrottleInterval"] != 10 {
			t.Errorf("ThrottleInterval = %v, want 10", pd["ThrottleInterval"])
		}
	})

	t.Run("unset omits the key", func(t *testing.T) {
		pd := BuildPlistDict(&ServiceConfig{Label: "com.example", Program: "/usr/bin/true"}, false)
		if _, ok := pd["ThrottleInterval"]; ok {
			t.Errorf("ThrottleInterval should be omitted when unset, got %v", pd["ThrottleInterval"])
		}
	})
}

// TestKeepAlive_RoundTrip reads a plist whose KeepAlive is a dictionary with a
// ThrottleInterval, then re-encodes via BuildPlistDict and confirms the keys
// survive (Task 1.5 verification).
func TestKeepAlive_RoundTrip(t *testing.T) {
	tmpDir := t.TempDir()
	plistContent := `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>Label</key>
	<string>com.test.roundtrip</string>
	<key>Program</key>
	<string>/usr/bin/true</string>
	<key>KeepAlive</key>
	<dict>
		<key>SuccessfulExit</key>
		<false/>
		<key>PathState</key>
		<dict>
			<key>/tmp/flag</key>
			<true/>
		</dict>
	</dict>
	<key>ThrottleInterval</key>
	<integer>30</integer>
</dict>
</plist>`
	if err := os.WriteFile(filepath.Join(tmpDir, "com.test.roundtrip.plist"), []byte(plistContent), 0644); err != nil {
		t.Fatalf("failed to write plist: %v", err)
	}

	m := &UserManager{launchAgentsPath: tmpDir}
	service, err := m.Get("com.test.roundtrip")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if service.ThrottleInterval == nil || *service.ThrottleInterval != 30 {
		t.Errorf("ThrottleInterval = %v, want pointer to 30", service.ThrottleInterval)
	}
	if service.KeepAlive.Mode != KeepAliveModeDictionary {
		t.Fatalf("KeepAlive.Mode = %q, want dictionary", service.KeepAlive.Mode)
	}
	if service.KeepAlive.PathState["/tmp/flag"] != true {
		t.Errorf("KeepAlive.PathState[/tmp/flag] = %v, want true", service.KeepAlive.PathState["/tmp/flag"])
	}

	// Re-encode and confirm the KeepAlive dict + ThrottleInterval survive.
	cfg := &ServiceConfig{
		Label:            service.Label,
		Program:          service.Program,
		KeepAlive:        service.KeepAlive,
		ThrottleInterval: service.ThrottleInterval,
	}
	pd := BuildPlistDict(cfg, false)
	if pd["ThrottleInterval"] != 30 {
		t.Errorf("re-encoded ThrottleInterval = %v, want 30", pd["ThrottleInterval"])
	}
	dict, ok := pd["KeepAlive"].(map[string]any)
	if !ok {
		t.Fatalf("re-encoded KeepAlive = %v (%T), want dictionary", pd["KeepAlive"], pd["KeepAlive"])
	}
	if dict["SuccessfulExit"] != false {
		t.Errorf("re-encoded SuccessfulExit = %v, want false", dict["SuccessfulExit"])
	}
	ps, ok := dict["PathState"].(map[string]any)
	if !ok || ps["/tmp/flag"] != true {
		t.Errorf("re-encoded PathState = %v, want {/tmp/flag: true}", dict["PathState"])
	}
}
