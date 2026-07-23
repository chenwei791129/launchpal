//go:build darwin

package privhelper

import (
	"errors"
	"os"

	"golang.org/x/sys/unix"
)

// openDirNoFollow opens trustedRoot (following symlinks — the allowlist prefix
// or the user's home is a trusted, root-owned or well-known anchor), then
// descends through rel, opening each component with O_DIRECTORY|O_NOFOLLOW so a
// symlink planted at ANY intermediate component fails the traversal instead of
// silently redirecting a root-privileged operation outside the allowlist. When
// create is true a missing component is created with Mkdirat(perm) before being
// opened, so directory creation cannot reintroduce the intermediate-symlink
// escape (an issue the old os.MkdirAll had). Returns an fd for the final
// directory; the caller must unix.Close it.
func openDirNoFollow(trustedRoot string, rel []string, create bool, perm os.FileMode) (int, error) {
	fd, err := unix.Open(trustedRoot, unix.O_RDONLY|unix.O_DIRECTORY|unix.O_CLOEXEC, 0)
	if err != nil {
		return -1, err
	}
	const flags = unix.O_RDONLY | unix.O_DIRECTORY | unix.O_NOFOLLOW | unix.O_CLOEXEC
	for _, comp := range rel {
		next, err := unix.Openat(fd, comp, flags, 0)
		if err != nil && create && errors.Is(err, unix.ENOENT) {
			if mkErr := unix.Mkdirat(fd, comp, uint32(perm.Perm())); mkErr != nil && !errors.Is(mkErr, unix.EEXIST) {
				_ = unix.Close(fd)
				return -1, mkErr
			}
			next, err = unix.Openat(fd, comp, flags, 0)
		}
		if err != nil {
			_ = unix.Close(fd)
			return -1, err
		}
		_ = unix.Close(fd)
		fd = next
	}
	return fd, nil
}

// symlinkSafeEnsureLog creates any missing directories between prefix and the
// leaf's parent (symlink-safely), tightens the leaf's parent to 0755 (launchd
// often creates per-service dirs as 0744, blocking the GUI user from entering),
// and touches the leaf as 0644 if it is absent — refusing to follow a symlink
// at any component.
func symlinkSafeEnsureLog(prefix string, dirs []string, leaf string) error {
	fd, err := openDirNoFollow(prefix, dirs, true, 0755)
	if err != nil {
		return err
	}
	defer func() { _ = unix.Close(fd) }()
	if err := unix.Fchmod(fd, 0755); err != nil {
		return err
	}
	var st unix.Stat_t
	if err := unix.Fstatat(fd, leaf, &st, unix.AT_SYMLINK_NOFOLLOW); err != nil {
		if !errors.Is(err, unix.ENOENT) {
			return err
		}
		lfd, err := unix.Openat(fd, leaf, unix.O_WRONLY|unix.O_CREAT|unix.O_EXCL|unix.O_NOFOLLOW|unix.O_CLOEXEC, 0644)
		if err != nil {
			if errors.Is(err, unix.EEXIST) {
				return nil // lost the create race; honor whatever is there now
			}
			return err
		}
		_ = unix.Close(lfd)
	}
	return nil
}

// symlinkSafeTruncate opens the leaf O_WRONLY|O_TRUNC|O_NOFOLLOW relative to a
// symlink-safe parent and closes it, truncating in place. A missing directory
// or file surfaces as ENOENT (errors.Is os.ErrNotExist); a symlink leaf
// surfaces as ELOOP so the caller can distinguish "not found" from "refused".
func symlinkSafeTruncate(prefix string, dirs []string, leaf string) error {
	fd, err := openDirNoFollow(prefix, dirs, false, 0)
	if err != nil {
		return err
	}
	defer func() { _ = unix.Close(fd) }()
	lfd, err := unix.Openat(fd, leaf, unix.O_WRONLY|unix.O_TRUNC|unix.O_NOFOLLOW|unix.O_CLOEXEC, 0)
	if err != nil {
		return err
	}
	return unix.Close(lfd)
}

// symlinkSafeDelete removes the leaf via unlinkat relative to a symlink-safe
// parent, after fstatat-ing it with AT_SYMLINK_NOFOLLOW to refuse symlinks and
// non-regular files. It never follows a symlink at any path component.
func symlinkSafeDelete(prefix string, dirs []string, leaf string) error {
	fd, err := openDirNoFollow(prefix, dirs, false, 0)
	if err != nil {
		return err
	}
	defer func() { _ = unix.Close(fd) }()
	var st unix.Stat_t
	if err := unix.Fstatat(fd, leaf, &st, unix.AT_SYMLINK_NOFOLLOW); err != nil {
		return err
	}
	switch st.Mode & unix.S_IFMT {
	case unix.S_IFLNK:
		return errRefuseSymlink
	case unix.S_IFREG:
		// ok
	default:
		return errNotRegular
	}
	return unix.Unlinkat(fd, leaf, 0)
}

// symlinkSafeRemoveEmptyParent best-effort removes the leaf's parent directory
// (the last element of dirs) via unlinkat(AT_REMOVEDIR) relative to a
// symlink-safe grandparent fd, so a symlink swapped in at an intermediate
// component cannot redirect the rmdir onto a directory outside the allowlist.
// A non-empty directory (ENOTEMPTY) or any other error is ignored — the file
// removal the caller already performed is the user-visible operation.
func symlinkSafeRemoveEmptyParent(prefix string, dirs []string) {
	if len(dirs) == 0 {
		return
	}
	gpfd, err := openDirNoFollow(prefix, dirs[:len(dirs)-1], false, 0)
	if err != nil {
		return
	}
	defer func() { _ = unix.Close(gpfd) }()
	_ = unix.Unlinkat(gpfd, dirs[len(dirs)-1], unix.AT_REMOVEDIR)
}

// symlinkSafeWriteInDir creates trustedRoot/rel per component with O_NOFOLLOW,
// then writes data to the leaf `name` relative to the final directory fd (also
// O_NOFOLLOW), so neither the directory chain nor the leaf follows a symlink at
// any component. Because the write targets the openat-resolved fd rather than
// re-resolving the path by name, a symlink swapped into an intermediate
// component after the chain is created cannot redirect the write (closing the
// TOCTOU the mkdir-then-write-by-path form left open). An existing regular leaf
// is truncated; a symlink leaf is refused (O_NOFOLLOW → ELOOP).
func symlinkSafeWriteInDir(trustedRoot string, rel []string, name string, data []byte, perm os.FileMode) error {
	fd, err := openDirNoFollow(trustedRoot, rel, true, 0755)
	if err != nil {
		return err
	}
	defer func() { _ = unix.Close(fd) }()
	lfd, err := unix.Openat(fd, name, unix.O_WRONLY|unix.O_CREAT|unix.O_TRUNC|unix.O_NOFOLLOW|unix.O_CLOEXEC, uint32(perm.Perm()))
	if err != nil {
		return err
	}
	_, wErr := unix.Write(lfd, data)
	cErr := unix.Close(lfd)
	if wErr != nil {
		return wErr
	}
	return cErr
}
