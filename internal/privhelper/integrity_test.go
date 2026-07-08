package privhelper

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"
)

func TestFileSHA256(t *testing.T) {
	dir := t.TempDir()

	t.Run("computes digest of an existing file", func(t *testing.T) {
		content := []byte("launchpal-privhelper binary bytes\n")
		p := filepath.Join(dir, "bin")
		if err := os.WriteFile(p, content, 0o644); err != nil {
			t.Fatalf("seed: %v", err)
		}
		got, err := FileSHA256(p)
		if err != nil {
			t.Fatalf("FileSHA256: %v", err)
		}
		want := hex.EncodeToString(func() []byte { s := sha256.Sum256(content); return s[:] }())
		if got != want {
			t.Errorf("digest = %s, want %s", got, want)
		}
	})

	t.Run("missing file returns error", func(t *testing.T) {
		if _, err := FileSHA256(filepath.Join(dir, "does-not-exist")); err == nil {
			t.Fatal("expected error for missing file")
		}
	})

	t.Run("symlink is refused", func(t *testing.T) {
		if syscallNoFollow == 0 {
			t.Skip("O_NOFOLLOW unavailable on this build")
		}
		target := filepath.Join(dir, "real-bin")
		if err := os.WriteFile(target, []byte("real"), 0o644); err != nil {
			t.Fatalf("seed target: %v", err)
		}
		link := filepath.Join(dir, "link-bin")
		if err := os.Symlink(target, link); err != nil {
			t.Fatalf("symlink: %v", err)
		}
		if _, err := FileSHA256(link); err == nil {
			t.Error("expected error hashing a symlink with O_NOFOLLOW")
		}
	})
}

func TestVerifyProtectedIdentity(t *testing.T) {
	cases := []struct {
		name string
		id   protectedIdentity
		want bool
	}{
		{"root-owned 0755", protectedIdentity{exists: true, regular: true, ownerUID: 0, perm: 0o755}, true},
		{"root-owned 0700", protectedIdentity{exists: true, regular: true, ownerUID: 0, perm: 0o700}, true},
		{"non-root owner", protectedIdentity{exists: true, regular: true, ownerUID: 501, perm: 0o755}, false},
		{"group-writable", protectedIdentity{exists: true, regular: true, ownerUID: 0, perm: 0o775}, false},
		{"other-writable", protectedIdentity{exists: true, regular: true, ownerUID: 0, perm: 0o757}, false},
		{"non-regular (symlink/dir)", protectedIdentity{exists: true, regular: false, ownerUID: 0, perm: 0o755}, false},
		{"missing", protectedIdentity{exists: false}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := verifyProtectedIdentity(tc.id); got != tc.want {
				t.Errorf("verifyProtectedIdentity(%+v) = %v, want %v", tc.id, got, tc.want)
			}
		})
	}
}

func TestIsVerifiedProtectedCopy(t *testing.T) {
	dir := t.TempDir()

	t.Run("regular file owned by the test user is unverified (not root)", func(t *testing.T) {
		p := filepath.Join(dir, "user-owned")
		if err := os.WriteFile(p, []byte("x"), 0o755); err != nil {
			t.Fatalf("seed: %v", err)
		}
		// The test process is not root, so a file it creates is never
		// UID 0 — the ownership gate must reject it.
		if IsVerifiedProtectedCopy(p) {
			t.Error("expected unverified for a non-root-owned file")
		}
	})

	t.Run("missing path is unverified", func(t *testing.T) {
		if IsVerifiedProtectedCopy(filepath.Join(dir, "nope")) {
			t.Error("expected unverified for a missing path")
		}
	})

	t.Run("symlink is unverified", func(t *testing.T) {
		if syscallNoFollow == 0 {
			t.Skip("symlink identity check only meaningful with O_NOFOLLOW semantics")
		}
		target := filepath.Join(dir, "target")
		if err := os.WriteFile(target, []byte("x"), 0o755); err != nil {
			t.Fatalf("seed: %v", err)
		}
		link := filepath.Join(dir, "link")
		if err := os.Symlink(target, link); err != nil {
			t.Fatalf("symlink: %v", err)
		}
		if IsVerifiedProtectedCopy(link) {
			t.Error("expected unverified for a symlink")
		}
	})
}
