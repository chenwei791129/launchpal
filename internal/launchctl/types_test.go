package launchctl

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestCanWriteLogFile(t *testing.T) {
	tmp := t.TempDir()

	t.Run("writable file returns true", func(t *testing.T) {
		path := filepath.Join(tmp, "writable.log")
		if err := os.WriteFile(path, []byte("data"), 0644); err != nil {
			t.Fatalf("seed: %v", err)
		}
		if !canWriteLogFile(path) {
			t.Error("expected true for writable file")
		}
	})

	t.Run("read-only file returns false", func(t *testing.T) {
		// Skip when running as root: 0444 still permits root to open for
		// writing, so the test would always succeed and silently lose its
		// signal value.
		if os.Geteuid() == 0 {
			t.Skip("root bypasses mode bits")
		}
		path := filepath.Join(tmp, "readonly.log")
		if err := os.WriteFile(path, []byte("data"), 0444); err != nil {
			t.Fatalf("seed: %v", err)
		}
		if canWriteLogFile(path) {
			t.Error("expected false for read-only file")
		}
	})

	t.Run("missing file returns false", func(t *testing.T) {
		if canWriteLogFile(filepath.Join(tmp, "never-existed.log")) {
			t.Error("expected false for missing file")
		}
	})

	t.Run("symlink returns false on darwin", func(t *testing.T) {
		// O_NOFOLLOW only takes effect on darwin (LaunchPal's only
		// supported platform). On other GOOS targets nofollowFlag is 0 so
		// the symlink is silently followed; skip there to keep portable CI
		// honest about what is and isn't enforced.
		if runtime.GOOS != "darwin" {
			t.Skip("O_NOFOLLOW behavior is darwin-specific")
		}
		target := filepath.Join(tmp, "symlink-target.log")
		if err := os.WriteFile(target, []byte("data"), 0644); err != nil {
			t.Fatalf("seed target: %v", err)
		}
		link := filepath.Join(tmp, "link.log")
		if err := os.Symlink(target, link); err != nil {
			t.Fatalf("symlink: %v", err)
		}
		if canWriteLogFile(link) {
			t.Error("expected false for symlink (O_NOFOLLOW should reject)")
		}
	})

	t.Run("empty path returns false", func(t *testing.T) {
		if canWriteLogFile("") {
			t.Error("expected false for empty path")
		}
	})
}

func TestTruncateLogFile(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "trunc.log")
	if err := os.WriteFile(path, []byte("noisy log content\n"), 0640); err != nil {
		t.Fatalf("seed: %v", err)
	}
	beforeMode, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if err := truncateLogFile(path); err != nil {
		t.Fatalf("truncateLogFile: %v", err)
	}
	after, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat after: %v", err)
	}
	if after.Size() != 0 {
		t.Errorf("size = %d, want 0", after.Size())
	}
	if after.Mode().Perm() != beforeMode.Mode().Perm() {
		t.Errorf("mode changed: %o -> %o", beforeMode.Mode().Perm(), after.Mode().Perm())
	}
}

func TestSelectLogPath(t *testing.T) {
	svc := &Service{StdoutPath: "/var/log/out.log", StderrPath: "/var/log/err.log"}
	if got := selectLogPath(svc, LogTypeStdout); got != "/var/log/out.log" {
		t.Errorf("stdout = %q", got)
	}
	if got := selectLogPath(svc, LogTypeStderr); got != "/var/log/err.log" {
		t.Errorf("stderr = %q", got)
	}
	if got := selectLogPath(svc, "trace"); got != "" {
		t.Errorf("invalid type = %q, want empty", got)
	}
}

func TestLogClearStatusFor(t *testing.T) {
	tmp := t.TempDir()

	t.Run("missing file: exists=false, writable=false", func(t *testing.T) {
		path := filepath.Join(tmp, "missing.log")
		got := logClearStatusFor(path)
		if got.LogPath != path || got.Exists || got.UserWritable {
			t.Errorf("got = %+v, want path=%s exists=false writable=false", got, path)
		}
	})

	t.Run("existing writable file", func(t *testing.T) {
		path := filepath.Join(tmp, "ok.log")
		if err := os.WriteFile(path, []byte("x"), 0644); err != nil {
			t.Fatalf("seed: %v", err)
		}
		got := logClearStatusFor(path)
		if !got.Exists {
			t.Errorf("expected exists=true")
		}
		if !got.UserWritable {
			t.Errorf("expected writable=true")
		}
	})

	t.Run("empty path", func(t *testing.T) {
		got := logClearStatusFor("")
		if got.LogPath != "" || got.Exists || got.UserWritable {
			t.Errorf("got = %+v, want all zero", got)
		}
	})
}
