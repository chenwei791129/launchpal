package launchctl

import (
	"os"
	"path/filepath"
	"testing"

	"howett.net/plist"
)

func TestAbandonProcessGroupUpdate(t *testing.T) {
	const label = "com.example.abandon-process-group"
	const key = "AbandonProcessGroup"
	for _, tc := range []struct {
		name  string
		value any
	}{
		{"enabled", true},
		{"disabled", false},
		{"absent", nil},
		{"non-boolean value", 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("PATH", t.TempDir()) // Never call the real launchctl during Update.
			dir := t.TempDir()
			path := filepath.Join(dir, label+".plist")
			existing := map[string]any{"Label": label, "Program": "/usr/bin/true", "ProcessType": "Background"}
			if tc.value != nil {
				existing[key] = tc.value
			}
			data, err := encodeDict(existing)
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(path, data, 0600); err != nil {
				t.Fatal(err)
			}
			user := &UserManager{launchAgentsPath: dir}
			system := &SystemManager{readOnlyManager: readOnlyManager{basePath: dir, serviceType: ServiceTypeSystem}}
			apple := &AppleSystemManager{readOnlyManager: readOnlyManager{basePath: dir, serviceType: ServiceTypeAppleSystem}}
			helper := &fakeAdminClient{}
			system.SetAdminClient(helper)
			wantEnabled := tc.value == true
			for _, get := range []func(string) (*Service, error){user.Get, system.Get, apple.Get} {
				service, err := get(label)
				if err != nil {
					t.Fatal(err)
				}
				if service.AbandonProcessGroup != wantEnabled {
					t.Fatalf("%s decoded %s = %v, want %v", service.Type, key, service.AbandonProcessGroup, wantEnabled)
				}
			}
			cfg := &ServiceConfig{Label: label, Program: "/usr/bin/false"}
			steps := []bool{wantEnabled}
			if wantEnabled {
				steps = append(steps, false)
			}
			for _, enabled := range steps {
				cfg.AbandonProcessGroup = enabled
				if err := system.Update(label, cfg); err != nil {
					t.Fatal(err)
				}
				if err := user.Update(label, cfg); err != nil {
					t.Fatal(err)
				}
				written, err := os.ReadFile(path)
				if err != nil {
					t.Fatal(err)
				}
				for _, data := range [][]byte{written, helper.lastWriteData} {
					var got map[string]any
					if _, err := plist.Unmarshal(data, &got); err != nil {
						t.Fatal(err)
					}
					value, present := got[key]
					if present != enabled || (enabled && value != true) || got["ProcessType"] != "Background" || got["Program"] != cfg.Program {
						t.Fatalf("after Update (enabled=%v): %v", enabled, got)
					}
				}
			}
		})
	}
}
