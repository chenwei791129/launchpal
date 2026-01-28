package launchctl

import (
	"os"
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
