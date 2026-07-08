package privhelper

import (
	"os"
	"path/filepath"
	"testing"
)

// noopChownFile / noopChownDir stand in for root-only chown calls: the test
// process is not root, so a real f.Chown(0,0) would fail with EPERM. The
// production wiring uses the real chown; these record that install requested
// root:wheel without actually needing the capability.
func installFixtures(t *testing.T) (base string, calls *chownCalls) {
	t.Helper()
	return t.TempDir(), &chownCalls{}
}

type chownCalls struct {
	fileUID, fileGID int
	fileCalled       bool
	dirUID, dirGID   int
	dirCalled        bool
}

func (c *chownCalls) file(_ *os.File, uid, gid int) error {
	c.fileCalled = true
	c.fileUID, c.fileGID = uid, gid
	return nil
}

func (c *chownCalls) dir(_ string, uid, gid int) error {
	c.dirCalled = true
	c.dirUID, c.dirGID = uid, gid
	return nil
}

func writeSource(t *testing.T, base string, content []byte) string {
	t.Helper()
	src := filepath.Join(base, "bundle-helper")
	if err := os.WriteFile(src, content, 0o755); err != nil {
		t.Fatalf("seed source: %v", err)
	}
	return src
}

func TestInstallProtectedCopy_FirstInstall(t *testing.T) {
	base, calls := installFixtures(t)
	content := []byte("#!/helper\nprivhelper image\n")
	src := writeSource(t, base, content)
	target := filepath.Join(base, "LaunchPal", "launchpal-privhelper")

	installed, err := installProtectedCopy(installConfig{
		sourcePath: src,
		targetPath: target,
		chownFile:  calls.file,
		chownDir:   calls.dir,
	})
	if err != nil {
		t.Fatalf("install: %v", err)
	}
	if !installed {
		t.Error("installed = false, want true on first install")
	}

	got, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("read target: %v", err)
	}
	if string(got) != string(content) {
		t.Errorf("target content = %q, want %q", got, content)
	}
	info, err := os.Lstat(target)
	if err != nil {
		t.Fatalf("lstat target: %v", err)
	}
	if info.Mode().Perm() != 0o755 {
		t.Errorf("target perm = %o, want 0755", info.Mode().Perm())
	}
	dirInfo, err := os.Stat(filepath.Dir(target))
	if err != nil {
		t.Fatalf("stat parent: %v", err)
	}
	if dirInfo.Mode().Perm() != 0o755 {
		t.Errorf("parent perm = %o, want 0755", dirInfo.Mode().Perm())
	}
	if !calls.fileCalled || calls.fileUID != 0 || calls.fileGID != 0 {
		t.Errorf("chownFile = %+v, want called with 0:0", calls)
	}
	if !calls.dirCalled || calls.dirUID != 0 || calls.dirGID != 0 {
		t.Errorf("chownDir = %+v, want called with 0:0", calls)
	}
}

func TestInstallProtectedCopy_IdempotentWhenCurrent(t *testing.T) {
	base, calls := installFixtures(t)
	content := []byte("identical image\n")
	src := writeSource(t, base, content)
	target := filepath.Join(base, "LaunchPal", "launchpal-privhelper")

	if _, err := installProtectedCopy(installConfig{sourcePath: src, targetPath: target, chownFile: calls.file, chownDir: calls.dir}); err != nil {
		t.Fatalf("first install: %v", err)
	}
	installed, err := installProtectedCopy(installConfig{sourcePath: src, targetPath: target, chownFile: calls.file, chownDir: calls.dir})
	if err != nil {
		t.Fatalf("second install: %v", err)
	}
	if installed {
		t.Error("installed = true on identical re-install, want false (idempotent)")
	}
}

func TestInstallProtectedCopy_RewritesWhenDifferent(t *testing.T) {
	base, calls := installFixtures(t)
	target := filepath.Join(base, "LaunchPal", "launchpal-privhelper")

	src1 := writeSource(t, base, []byte("old image\n"))
	if _, err := installProtectedCopy(installConfig{sourcePath: src1, targetPath: target, chownFile: calls.file, chownDir: calls.dir}); err != nil {
		t.Fatalf("install v1: %v", err)
	}
	// A different source (new app version) must overwrite the protected copy.
	newContent := []byte("new image\n")
	src2 := filepath.Join(base, "bundle-helper-v2")
	if err := os.WriteFile(src2, newContent, 0o755); err != nil {
		t.Fatalf("seed v2: %v", err)
	}
	installed, err := installProtectedCopy(installConfig{sourcePath: src2, targetPath: target, chownFile: calls.file, chownDir: calls.dir})
	if err != nil {
		t.Fatalf("install v2: %v", err)
	}
	if !installed {
		t.Error("installed = false, want true when source differs")
	}
	got, _ := os.ReadFile(target)
	if string(got) != string(newContent) {
		t.Errorf("target content = %q, want %q", got, newContent)
	}
}

func TestInstallProtectedCopy_RefuseSymlinkTarget(t *testing.T) {
	if syscallNoFollow == 0 {
		t.Skip("O_NOFOLLOW unavailable on this build; symlink refusal only enforced on darwin")
	}
	base, calls := installFixtures(t)
	content := []byte("image\n")
	src := writeSource(t, base, content)

	dir := filepath.Join(base, "LaunchPal")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	outside := filepath.Join(base, "outside-target")
	if err := os.WriteFile(outside, []byte("victim"), 0o644); err != nil {
		t.Fatalf("seed victim: %v", err)
	}
	target := filepath.Join(dir, "launchpal-privhelper")
	if err := os.Symlink(outside, target); err != nil {
		t.Fatalf("symlink: %v", err)
	}

	installed, err := installProtectedCopy(installConfig{sourcePath: src, targetPath: target, chownFile: calls.file, chownDir: calls.dir})
	if err == nil {
		t.Fatal("expected error installing over a symlink target")
	}
	if installed {
		t.Error("installed = true, want false when target is a symlink")
	}
	// The symlink's victim must be untouched.
	victim, _ := os.ReadFile(outside)
	if string(victim) != "victim" {
		t.Errorf("symlink target was written through: %q", victim)
	}
}
