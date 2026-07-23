package privhelper

import "errors"

// errRefuseSymlink / errNotRegular are the sentinel errors symlinkSafeDelete
// returns for a symlink or non-regular leaf. They live in this untagged file
// (not duplicated per platform) so the darwin and non-darwin implementations
// share one message text — callers and tests substring-match these strings, so
// drift between two copies would silently break that contract.
var (
	errRefuseSymlink = errors.New("refusing to delete symlink")
	errNotRegular    = errors.New("not a regular file")
)
