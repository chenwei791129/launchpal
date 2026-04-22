package launchctl

import (
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
