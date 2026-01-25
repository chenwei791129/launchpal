package backup

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBackupManager_GetBackupDir(t *testing.T) {
	m := NewBackupManager()
	dir := m.getBackupDir("com.test.service")

	home, _ := os.UserHomeDir()
	expected := filepath.Join(home, ".launchpal", "backups", "com.test.service")

	if dir != expected {
		t.Errorf("getBackupDir() = %v, want %v", dir, expected)
	}
}
