package privhelper

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// installConfig parameterizes installProtectedCopy so the copy logic is
// testable without root: chownFile / chownDir are injected. Production wires
// the real f.Chown / os.Chown (which require root); tests record the requested
// UID/GID instead.
type installConfig struct {
	sourcePath string
	targetPath string
	chownFile  func(*os.File, int, int) error
	chownDir   func(string, int, int) error
}

// installProtectedCopy copies the source image to the protected target path,
// owned root:wheel (via the injected chown) with mode 0755, creating the
// parent directory the same way. It is idempotent: when the target already
// matches the source byte-for-byte it performs no copy and returns
// installed=false. The final path component is opened with O_NOFOLLOW so a
// symlink planted at the target is refused rather than written through.
//
// Returns installed=true when a copy was performed.
func installProtectedCopy(cfg installConfig) (bool, error) {
	// Idempotency: skip the copy when a regular target already matches the
	// source. The source is hashed only when there is a regular target to
	// compare against — first install and non-regular targets read the source
	// exactly once, via io.Copy below. A symlink or non-regular target falls
	// through to the write path, where O_NOFOLLOW rejects it.
	if info, err := os.Lstat(cfg.targetPath); err == nil && info.Mode().IsRegular() {
		srcHash, err := FileSHA256(cfg.sourcePath)
		if err != nil {
			return false, fmt.Errorf("hash source: %w", err)
		}
		if tgtHash, err := FileSHA256(cfg.targetPath); err == nil && tgtHash == srcHash {
			return false, nil
		}
	}

	dir := filepath.Dir(cfg.targetPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return false, fmt.Errorf("create parent dir: %w", err)
	}
	// MkdirAll honors the umask and a pre-existing directory may be stricter
	// (launchd defaults to 0744), so force 0755 explicitly.
	if err := os.Chmod(dir, 0o755); err != nil {
		return false, fmt.Errorf("chmod parent dir: %w", err)
	}
	if cfg.chownDir != nil {
		if err := cfg.chownDir(dir, 0, 0); err != nil {
			return false, fmt.Errorf("chown parent dir: %w", err)
		}
	}

	src, err := os.OpenFile(cfg.sourcePath, os.O_RDONLY|syscallNoFollow, 0)
	if err != nil {
		return false, fmt.Errorf("open source: %w", err)
	}
	defer func() { _ = src.Close() }()

	// O_NOFOLLOW on the target refuses a pre-existing symlink at the final
	// component (ELOOP), so a planted symlink cannot redirect the root write.
	dst, err := os.OpenFile(cfg.targetPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC|syscallNoFollow, 0o755)
	if err != nil {
		return false, fmt.Errorf("open target: %w", err)
	}
	if _, err := io.Copy(dst, src); err != nil {
		_ = dst.Close()
		return false, fmt.Errorf("copy image: %w", err)
	}
	// Force mode 0755 regardless of umask applied at create time.
	if err := dst.Chmod(0o755); err != nil {
		_ = dst.Close()
		return false, fmt.Errorf("chmod target: %w", err)
	}
	if cfg.chownFile != nil {
		if err := cfg.chownFile(dst, 0, 0); err != nil {
			_ = dst.Close()
			return false, fmt.Errorf("chown target: %w", err)
		}
	}
	if err := dst.Close(); err != nil {
		return false, fmt.Errorf("close target: %w", err)
	}
	return true, nil
}

// InstallProtectedCopy copies the running helper image at exePath to the
// protected path (ProtectedHelperPath), owned root:wheel with mode 0755. It
// wraps installProtectedCopy with the real root-only chown calls. Intended to
// be called by the root-privileged helper at startup; the caller decides
// whether to invoke it (only when launched from a path other than the
// protected path) and treats any error as non-fatal.
func InstallProtectedCopy(exePath string) (bool, error) {
	return installProtectedCopy(installConfig{
		sourcePath: exePath,
		targetPath: ProtectedHelperPath,
		chownFile:  func(f *os.File, uid, gid int) error { return f.Chown(uid, gid) },
		chownDir:   os.Chown,
	})
}
