package launchctl

import (
	"os"
	"path/filepath"
	"testing"
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
