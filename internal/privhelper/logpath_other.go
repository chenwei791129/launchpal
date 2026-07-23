//go:build !darwin

package privhelper

import (
	"fmt"
	"os"
	"path/filepath"
)

// The non-darwin build has no openat/*at path-safety guarantees (LaunchPal
// only ships on macOS; this file exists so the package still compiles and the
// non-privileged code paths run under `go test` on other platforms). It
// preserves the pre-openat os.* behavior; the per-component symlink safety is
// only enforced on darwin, matching the syscallNoFollow==0 skips in the tests.
// The errRefuseSymlink / errNotRegular sentinels this file returns are shared
// with the darwin build from the untagged logpath.go.

func joinLog(prefix string, dirs []string, leaf string) string {
	parts := append([]string{prefix}, dirs...)
	parts = append(parts, leaf)
	return filepath.Join(parts...)
}

func symlinkSafeEnsureLog(prefix string, dirs []string, leaf string) error {
	dir := filepath.Join(append([]string{prefix}, dirs...)...)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	if err := os.Chmod(dir, 0755); err != nil {
		return err
	}
	full := filepath.Join(dir, leaf)
	if _, err := os.Lstat(full); os.IsNotExist(err) {
		f, err := os.OpenFile(full, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
		if err != nil {
			if os.IsExist(err) {
				return nil
			}
			return err
		}
		return f.Close()
	} else if err != nil {
		return err
	}
	return nil
}

func symlinkSafeTruncate(prefix string, dirs []string, leaf string) error {
	f, err := os.OpenFile(joinLog(prefix, dirs, leaf), os.O_WRONLY|os.O_TRUNC, 0)
	if err != nil {
		return err
	}
	return f.Close()
}

func symlinkSafeDelete(prefix string, dirs []string, leaf string) error {
	full := joinLog(prefix, dirs, leaf)
	info, err := os.Lstat(full)
	if err != nil {
		return err
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return errRefuseSymlink
	}
	if !info.Mode().IsRegular() {
		return errNotRegular
	}
	return os.Remove(full)
}

func symlinkSafeRemoveEmptyParent(prefix string, dirs []string) {
	if len(dirs) == 0 {
		return
	}
	_ = os.Remove(filepath.Join(append([]string{prefix}, dirs...)...))
}

func symlinkSafeWriteInDir(trustedRoot string, rel []string, name string, data []byte, perm os.FileMode) error {
	dir := filepath.Join(append([]string{trustedRoot}, rel...)...)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("mkdir chain: %w", err)
	}
	f, err := os.OpenFile(filepath.Join(dir, name), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, perm)
	if err != nil {
		return err
	}
	_, wErr := f.Write(data)
	cErr := f.Close()
	if wErr != nil {
		return wErr
	}
	return cErr
}
