package privhelper

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
	"syscall"
)

// ProtectedHelperPath is the single source of truth for the root-owned
// protected helper copy. Both the GUI (path resolution in admin_mode.go) and
// the helper (self-install in cmd/launchpal-privhelper) reference this
// constant. Its ancestor directories (/Library, /Library/Application Support)
// are root-writable only, so a non-root attacker cannot forge the path — this
// is the trust anchor that ends steady-state plant-and-wait.
const ProtectedHelperPath = "/Library/Application Support/LaunchPal/launchpal-privhelper"

// FileSHA256 returns the hex-encoded SHA-256 digest of the file at path. The
// file is opened with O_NOFOLLOW (on darwin) so a symlink planted at path is
// refused rather than silently hashed through to another target.
func FileSHA256(path string) (string, error) {
	f, err := os.OpenFile(path, os.O_RDONLY|syscallNoFollow, 0)
	if err != nil {
		return "", err
	}
	defer func() { _ = f.Close() }()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// protectedIdentity captures the observable properties of a candidate
// protected copy that determine whether it can be trusted. Separated from the
// filesystem lookup so verifyProtectedIdentity is a pure, table-testable
// decision.
type protectedIdentity struct {
	exists   bool
	regular  bool
	ownerUID int
	perm     os.FileMode
}

// verifyProtectedIdentity reports whether a protected copy is trustworthy: it
// must exist, be a regular file (not a symlink or directory), be owned by
// UID 0, and carry no group/other write bit (mode & 022 == 0).
func verifyProtectedIdentity(id protectedIdentity) bool {
	return id.exists && id.regular && id.ownerUID == 0 && id.perm&0o022 == 0
}

// IsVerifiedProtectedCopy reports whether the file at path is a valid
// root-owned protected copy. The lookup uses Lstat so a symlink is observed as
// a non-regular file and rejected rather than followed.
func IsVerifiedProtectedCopy(path string) bool {
	info, err := os.Lstat(path)
	if err != nil {
		return false
	}
	uid, ok := statOwnerUID(info)
	if !ok {
		return false
	}
	return verifyProtectedIdentity(protectedIdentity{
		exists:   true,
		regular:  info.Mode().IsRegular(),
		ownerUID: uid,
		perm:     info.Mode().Perm(),
	})
}

// statOwnerUID extracts the owning UID from a FileInfo. Returns ok=false when
// the platform does not expose a unix stat (never happens on the darwin ship
// target; the portable CI build runs on unix too).
func statOwnerUID(info os.FileInfo) (int, bool) {
	st, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return 0, false
	}
	return int(st.Uid), true
}
