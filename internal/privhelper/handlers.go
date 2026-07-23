package privhelper

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// SystemDaemonDir is the one directory under which the helper accepts writes.
// Paths outside this directory are rejected with ErrCodeInvalidParams.
// /System/Library/LaunchDaemons belongs to Apple (and SIP blocks writes there
// anyway) so only this directory is ever a valid target.
const SystemDaemonDir = "/Library/LaunchDaemons"

// launchctlPath is the absolute path the helper invokes launchctl by, so the
// resolved binary never depends on the $PATH inherited by the helper process.
const launchctlPath = "/bin/launchctl"

// SystemLogPathPrefixes bounds EnsureLogAccess (and the settings validator
// for systemLogDir) to locations where creating a world-readable file is
// harmless. macOS resolves /var and /tmp through /private, so both aliases
// are listed — filepath.Clean doesn't cross symlinks, so a path given as
// /var/log/x.log will still have that prefix when validated.
//
// Exported so internal/settings can consume the same constant. Both call
// sites SHALL reference this symbol by name; do not duplicate the literal
// slice elsewhere.
var SystemLogPathPrefixes = []string{
	"/var/log/",
	"/private/var/log/",
	"/Library/Logs/",
	"/tmp/",
	"/private/tmp/",
}

// labelPattern accepts reverse-DNS identifiers. This is deliberately strict
// so shell metacharacters (spaces, semicolons, quotes, backticks, ...) are
// rejected before any argument ever reaches launchctl.
var labelPattern = regexp.MustCompile(`^[A-Za-z0-9._-]+$`)

// Runner executes external commands; injectable so handler tests never
// actually shell out to launchctl.
type Runner interface {
	Run(ctx context.Context, name string, args ...string) (stdout, stderr []byte, err error)
}

// execRunner is the default Runner used in production.
type execRunner struct{}

// Run invokes the binary at name with args, captures both streams, and
// returns them alongside any error from exec.
func (execRunner) Run(ctx context.Context, name string, args ...string) ([]byte, []byte, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return stdout.Bytes(), stderr.Bytes(), err
}

// Chowner sets filesystem ownership. Injectable so tests don't need CAP_CHOWN.
type Chowner func(path string, uid, gid int) error

// HandlerOptions configure a set of RPC handlers.
type HandlerOptions struct {
	// Runner executes launchctl commands.
	Runner Runner

	// UserHome is the launching user's home directory; backups are placed
	// under UserHome/.launchpal/backups. Required for WritePlist/DeletePlist.
	UserHome string

	// LaunchingUID, LaunchingGID identify the user whose LaunchPal started
	// the helper. Backup files are chowned back to this uid/gid so the
	// user-side LaunchPal can read them.
	LaunchingUID int
	LaunchingGID int

	// Chown is invoked on newly-created backup files. Defaults to os.Chown.
	Chown Chowner

	// NowFn returns the current time; injectable for deterministic backup IDs.
	NowFn func() time.Time

	// ShutdownFn is called when the Shutdown RPC is received. Typically
	// wired to Server.RequestShutdown so the response is flushed before
	// the listener closes.
	ShutdownFn func()
}

// Handlers holds the set of RPC method implementations. It implements the
// Handler interface expected by Server.
type Handlers struct {
	opts HandlerOptions
}

// NewHandlers builds a Handlers from opts, filling in production defaults.
func NewHandlers(opts HandlerOptions) *Handlers {
	if opts.Runner == nil {
		opts.Runner = execRunner{}
	}
	if opts.Chown == nil {
		// Use Lchown so a symlink planted in $HOME/.launchpal/backups can't
		// trick the root helper into chowning an arbitrary target: Lchown
		// operates on the symlink itself, Chown follows it.
		opts.Chown = os.Lchown
	}
	if opts.NowFn == nil {
		opts.NowFn = time.Now
	}
	return &Handlers{opts: opts}
}

// Handle is the Server Handler callback; it looks up the method and
// forwards to the concrete implementation.
func (h *Handlers) Handle(ctx context.Context, req *Request) (any, *RPCError) {
	switch req.Method {
	case MethodPing:
		return PingResult{Pong: true}, nil
	case MethodBootstrap:
		return h.bootstrap(ctx, req.Params)
	case MethodBootout:
		return h.bootout(ctx, req.Params)
	case MethodKickstart:
		return h.kickstart(ctx, req.Params)
	case MethodWritePlist:
		return h.writePlist(ctx, req.Params)
	case MethodDeletePlist:
		return h.deletePlist(ctx, req.Params)
	case MethodEnsureLogAccess:
		return h.ensureLogAccess(ctx, req.Params)
	case MethodTruncateLog:
		return h.truncateLog(ctx, req.Params)
	case MethodDeleteLogPaths:
		return h.deleteLogPaths(ctx, req.Params)
	case MethodShutdown:
		if h.opts.ShutdownFn != nil {
			h.opts.ShutdownFn()
		}
		return OKResult{OK: true}, nil
	default:
		return nil, &RPCError{Code: ErrCodeUnknownMethod, Message: req.Method}
	}
}

func unmarshalParams[T any](raw json.RawMessage) (T, *RPCError) {
	var p T
	if len(raw) == 0 {
		return p, &RPCError{Code: ErrCodeInvalidParams, Message: "missing params"}
	}
	if err := json.Unmarshal(raw, &p); err != nil {
		return p, &RPCError{Code: ErrCodeInvalidParams, Message: err.Error()}
	}
	return p, nil
}

// validateSystemDaemonPath rejects paths not under /Library/LaunchDaemons/,
// normalizes ".."-containing inputs, and refuses symlink escapes. Returns
// the cleaned absolute path when valid.
func validateSystemDaemonPath(path string) (string, *RPCError) {
	if path == "" {
		return "", &RPCError{Code: ErrCodeInvalidParams, Message: "plistPath is required"}
	}
	if !filepath.IsAbs(path) {
		return "", &RPCError{Code: ErrCodeInvalidParams, Message: "plistPath must be absolute"}
	}
	clean := filepath.Clean(path)
	// filepath.Clean resolves ".." textually; confirm the result still sits
	// under the allowed directory. Using a trailing slash prevents
	// "/Library/LaunchDaemonsXYZ" from matching.
	allowed := filepath.Clean(SystemDaemonDir) + string(filepath.Separator)
	if !strings.HasPrefix(clean+string(filepath.Separator), allowed) {
		return "", &RPCError{Code: ErrCodeInvalidParams, Message: "path must be under " + SystemDaemonDir}
	}
	if !strings.HasSuffix(clean, ".plist") {
		return "", &RPCError{Code: ErrCodeInvalidParams, Message: "path must end in .plist"}
	}
	return clean, nil
}

// validateLabel rejects labels containing characters outside the strict
// reverse-DNS character class, which blocks shell metacharacter injection
// on Bootout / Kickstart.
func validateLabel(label string) *RPCError {
	if label == "" {
		return &RPCError{Code: ErrCodeInvalidParams, Message: "label is required"}
	}
	if !labelPattern.MatchString(label) {
		return &RPCError{Code: ErrCodeInvalidParams, Message: "label contains invalid characters"}
	}
	return nil
}

func (h *Handlers) bootstrap(ctx context.Context, raw json.RawMessage) (any, *RPCError) {
	p, errR := unmarshalParams[BootstrapParams](raw)
	if errR != nil {
		return nil, errR
	}
	clean, errR := validateSystemDaemonPath(p.PlistPath)
	if errR != nil {
		return nil, errR
	}
	_, stderr, err := h.opts.Runner.Run(ctx, launchctlPath, "bootstrap", "system", clean)
	if err != nil {
		return nil, &RPCError{Code: ErrCodeLaunchctlFailed, Message: errMsg(err, stderr)}
	}
	return OKResult{OK: true}, nil
}

func (h *Handlers) bootout(ctx context.Context, raw json.RawMessage) (any, *RPCError) {
	p, errR := unmarshalParams[BootoutParams](raw)
	if errR != nil {
		return nil, errR
	}
	if errR := validateLabel(p.Label); errR != nil {
		return nil, errR
	}
	_, stderr, err := h.opts.Runner.Run(ctx, launchctlPath, "bootout", "system/"+p.Label)
	if err != nil {
		return nil, &RPCError{Code: ErrCodeLaunchctlFailed, Message: errMsg(err, stderr)}
	}
	return OKResult{OK: true}, nil
}

func (h *Handlers) kickstart(ctx context.Context, raw json.RawMessage) (any, *RPCError) {
	p, errR := unmarshalParams[KickstartParams](raw)
	if errR != nil {
		return nil, errR
	}
	if errR := validateLabel(p.Label); errR != nil {
		return nil, errR
	}
	_, stderr, err := h.opts.Runner.Run(ctx, launchctlPath, "kickstart", "-k", "system/"+p.Label)
	if err != nil {
		return nil, &RPCError{Code: ErrCodeLaunchctlFailed, Message: errMsg(err, stderr)}
	}
	return OKResult{OK: true}, nil
}

func (h *Handlers) writePlist(ctx context.Context, raw json.RawMessage) (any, *RPCError) {
	_ = ctx
	p, errR := unmarshalParams[WritePlistParams](raw)
	if errR != nil {
		return nil, errR
	}
	clean, errR := validateSystemDaemonPath(p.PlistPath)
	if errR != nil {
		return nil, errR
	}
	data, err := base64.StdEncoding.DecodeString(p.Data)
	if err != nil {
		return nil, &RPCError{Code: ErrCodeInvalidParams, Message: "data must be base64: " + err.Error()}
	}
	if existing, exists, readErr := readIfExists(clean); readErr != nil {
		return nil, &RPCError{Code: ErrCodeIOError, Message: readErr.Error()}
	} else if exists {
		if errR := h.backupExisting(clean, existing); errR != nil {
			return nil, errR
		}
	}
	if errR := atomicWrite(clean, data, 0644, 0, 0); errR != nil {
		return nil, errR
	}
	return OKResult{OK: true}, nil
}

// validateLogPath accepts absolute paths under a small allowlist of
// well-known log/temp prefixes. The cleaned path and the matched allowlist
// prefix are returned so the caller can resolve the remaining components in a
// symlink-safe manner (openat-style traversal anchored at the trusted prefix).
// This lexical check is a fast pre-filter, NOT the enforcement boundary — the
// allowlist includes world-writable directories (/tmp, /private/tmp) where a
// same-UID process can plant a symlink at an intermediate component, so the
// per-component O_NOFOLLOW walk in logpath_darwin.go is what actually confines
// the operation. Rejects:
//   - non-absolute paths
//   - paths not under an allowed prefix (post-Clean, so "/var/log/../etc"
//     is normalized to "/etc" and rejected)
//   - paths ending in a separator (we require a file, not a bare directory)
//   - paths whose immediate parent is the allowlist prefix itself — e.g.
//     "/var/log/foo.log" would force a Chmod on /var/log, which we refuse
//     because /var/log is a system directory we shouldn't re-mode.
func validateLogPath(path string) (clean, prefix string, errR *RPCError) {
	if path == "" {
		return "", "", &RPCError{Code: ErrCodeInvalidParams, Message: "log path must not be empty"}
	}
	if !filepath.IsAbs(path) {
		return "", "", &RPCError{Code: ErrCodeInvalidParams, Message: "log path must be absolute"}
	}
	clean = filepath.Clean(path)
	// Each allowlist prefix ends in "/", so HasPrefix naturally rejects
	// both the prefix itself ("/tmp" has no trailing separator, so won't
	// start with "/tmp/") and look-alike siblings ("/var/logX" won't start
	// with "/var/log/").
	for _, p := range SystemLogPathPrefixes {
		if strings.HasPrefix(clean, p) {
			prefix = p
			break
		}
	}
	if prefix == "" {
		return "", "", &RPCError{Code: ErrCodeInvalidParams, Message: "log path not under an allowed prefix"}
	}
	// Require at least one sub-directory between the prefix root and the
	// file — we don't want to chmod /var/log itself. remainder is e.g.
	// "myservice/out.log" for an accepted path.
	remainder := clean[len(prefix):]
	if !strings.Contains(remainder, string(filepath.Separator)) {
		return "", "", &RPCError{Code: ErrCodeInvalidParams, Message: "log path must live in a sub-directory of " + strings.TrimSuffix(prefix, "/")}
	}
	return clean, prefix, nil
}

// logPathComponents splits a validated clean path into the directory
// components below the trusted prefix and the leaf file name. validateLogPath
// guarantees at least one sub-directory, so dirs is non-empty and leaf is a
// bare file name.
func logPathComponents(clean, prefix string) (dirs []string, leaf string) {
	parts := strings.Split(strings.TrimPrefix(clean, prefix), string(filepath.Separator))
	return parts[:len(parts)-1], parts[len(parts)-1]
}

// ensureLogAccess prepares log-file locations so the unprivileged GUI can
// tail them. For each validated path the leaf's parent chain is created and
// tightened to 0755 (launchd creates per-service log dirs as root 0744, which
// blocks the user from even entering the directory) and the file is touched as
// 0644 if absent. Every path component is resolved with O_NOFOLLOW
// (openat-style traversal anchored at the allowlist prefix, in
// symlinkSafeEnsureLog), so a symlink planted at ANY component — intermediate
// or leaf — fails the operation instead of redirecting the root-privileged
// create/chmod outside the allowlist.
//
// Empty paths in the list are silently skipped (plists routinely omit
// StandardErrorPath and we don't want to force the GUI to filter).
func (h *Handlers) ensureLogAccess(_ context.Context, raw json.RawMessage) (any, *RPCError) {
	p, errR := unmarshalParams[EnsureLogAccessParams](raw)
	if errR != nil {
		return nil, errR
	}
	for _, path := range p.Paths {
		if path == "" {
			continue
		}
		clean, prefix, errR := validateLogPath(path)
		if errR != nil {
			return nil, errR
		}
		dirs, leaf := logPathComponents(clean, prefix)
		if err := symlinkSafeEnsureLog(prefix, dirs, leaf); err != nil {
			return nil, &RPCError{Code: ErrCodeIOError, Message: err.Error()}
		}
	}
	return OKResult{OK: true}, nil
}

// truncateLog opens an existing log file with O_WRONLY|O_TRUNC|O_NOFOLLOW
// relative to a symlink-safe parent and immediately closes it. Every path
// component is resolved with O_NOFOLLOW (openat-style, in symlinkSafeTruncate),
// so a symlink at an intermediate directory or the leaf cannot redirect the
// truncate onto a real file outside the allowlist. Without O_CREATE the call
// surfaces ENOENT for missing files unchanged, so the helper cannot be coerced
// into materializing a 0-byte root-owned file in /tmp/.
func (h *Handlers) truncateLog(_ context.Context, raw json.RawMessage) (any, *RPCError) {
	p, errR := unmarshalParams[TruncateLogParams](raw)
	if errR != nil {
		return nil, errR
	}
	clean, prefix, errR := validateLogPath(p.Path)
	if errR != nil {
		return nil, errR
	}
	dirs, leaf := logPathComponents(clean, prefix)
	if err := symlinkSafeTruncate(prefix, dirs, leaf); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, &RPCError{Code: ErrCodeNotFound, Message: clean}
		}
		return nil, &RPCError{Code: ErrCodeIOError, Message: err.Error()}
	}
	return OKResult{OK: true}, nil
}

// deleteLogPaths removes each path in p.Paths after validating it against
// the log allowlist and refusing to follow symlinks. Per-path failures are
// collected in the result's Errors slice rather than bubbled up as an RPC
// error — a partial success (some paths deleted, others rejected/missing)
// is a valid response and lets callers like SystemManager.DeleteWithOptions
// surface a non-fatal warning instead of failing the whole delete.
//
// After a successful file removal the handler attempts to remove the parent
// directory; an error there is swallowed (typically ENOTEMPTY when other
// log files share the dir, occasionally ENOENT if the parent was already
// gone). Only one level up is collapsed — never recurse beyond the
// immediate parent.
func (h *Handlers) deleteLogPaths(_ context.Context, raw json.RawMessage) (any, *RPCError) {
	p, errR := unmarshalParams[DeleteLogPathsParams](raw)
	if errR != nil {
		return nil, errR
	}
	result := DeleteLogPathsResult{Errors: []string{}}
	for _, path := range p.Paths {
		if err := deleteOneLogPath(path); err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("%s: %s", path, err.Error()))
		}
	}
	return result, nil
}

// deleteOneLogPath validates path and removes the file, resolving every path
// component with O_NOFOLLOW (openat-style, in symlinkSafeDelete) so a symlink
// at an intermediate directory or the leaf cannot redirect the removal onto a
// real file outside the allowlist. It then best-effort removes the
// now-possibly-empty parent dir. Errors bubble up as-is so callers can
// substring-match the underlying errno (e.g. "no such file or directory").
func deleteOneLogPath(path string) error {
	clean, prefix, errR := validateLogPath(path)
	if errR != nil {
		return fmt.Errorf("%s", errR.Message)
	}
	dirs, leaf := logPathComponents(clean, prefix)
	if err := symlinkSafeDelete(prefix, dirs, leaf); err != nil {
		return err
	}
	// Best-effort parent cleanup, resolved per-component with O_NOFOLLOW
	// (symlinkSafeRemoveEmptyParent) so a symlink swapped into an intermediate
	// directory cannot redirect the rmdir outside the allowlist — matching the
	// symlink safety of the delete above rather than re-resolving the parent
	// path by name. A non-empty directory (ENOTEMPTY) or any other failure is
	// ignored because the file delete (the user-visible operation) succeeded.
	symlinkSafeRemoveEmptyParent(prefix, dirs)
	return nil
}

func (h *Handlers) deletePlist(ctx context.Context, raw json.RawMessage) (any, *RPCError) {
	_ = ctx
	p, errR := unmarshalParams[DeletePlistParams](raw)
	if errR != nil {
		return nil, errR
	}
	clean, errR := validateSystemDaemonPath(p.PlistPath)
	if errR != nil {
		return nil, errR
	}
	existing, exists, err := readIfExists(clean)
	if err != nil {
		return nil, &RPCError{Code: ErrCodeIOError, Message: err.Error()}
	}
	if !exists {
		return nil, &RPCError{Code: ErrCodeNotFound, Message: clean}
	}
	if errR := h.backupExisting(clean, existing); errR != nil {
		return nil, errR
	}
	if err := os.Remove(clean); err != nil {
		return nil, &RPCError{Code: ErrCodeIOError, Message: err.Error()}
	}
	return OKResult{OK: true}, nil
}

// readIfExists reads path; returns (data, true, nil) if the file is present,
// (nil, false, nil) if it doesn't exist, or (nil, false, err) on other I/O
// errors. Used to decide whether to run the backup path before a write.
//
// The read uses O_NOFOLLOW (on darwin) so a symlink planted at the daemon
// path redirects into a no-such-file error rather than silently leaking
// the target's contents into the user-owned backup file. On non-darwin
// platforms O_NOFOLLOW is a no-op (LaunchPal only ships on macOS).
func readIfExists(path string) ([]byte, bool, error) {
	f, err := os.OpenFile(path, os.O_RDONLY|syscallNoFollow, 0)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, false, nil
		}
		return nil, false, err
	}
	defer func() { _ = f.Close() }()
	data, err := io.ReadAll(f)
	if err != nil {
		return nil, false, err
	}
	return data, true, nil
}

// atomicWrite writes data to path atomically with the given mode, setting
// ownership to uid:gid. The write goes to a sibling temp file which is then
// renamed; this keeps system daemons consistent even if the helper crashes
// mid-write.
func atomicWrite(path string, data []byte, mode os.FileMode, uid, gid int) *RPCError {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".launchpal-privhelper-*")
	if err != nil {
		return &RPCError{Code: ErrCodeIOError, Message: err.Error()}
	}
	tmpPath := tmp.Name()
	defer func() { _ = os.Remove(tmpPath) }() // only removes if rename didn't happen
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return &RPCError{Code: ErrCodeIOError, Message: err.Error()}
	}
	if err := tmp.Chmod(mode); err != nil {
		_ = tmp.Close()
		return &RPCError{Code: ErrCodeIOError, Message: err.Error()}
	}
	if err := tmp.Close(); err != nil {
		return &RPCError{Code: ErrCodeIOError, Message: err.Error()}
	}
	// In production uid/gid are always 0/0 (root:wheel), which matches the
	// effective owner already — this Chown is a no-op on the real write
	// path. Suppressing the error keeps the call cheap for tests that run
	// without CAP_CHOWN.
	if err := os.Chown(tmpPath, uid, gid); err != nil {
		_ = err
	}
	if err := os.Rename(tmpPath, path); err != nil {
		return &RPCError{Code: ErrCodeIOError, Message: err.Error()}
	}
	return nil
}

// backupExisting copies existingData to ~user/.launchpal/backups/<label>/
// following the same convention as the user-side BackupManager. Newly created
// files are chowned to the launching user so they're readable by the GUI.
func (h *Handlers) backupExisting(plistPath string, existingData []byte) *RPCError {
	if h.opts.UserHome == "" {
		return &RPCError{Code: ErrCodeInternalError, Message: "user-home not configured"}
	}
	label := strings.TrimSuffix(filepath.Base(plistPath), ".plist")
	if label == "" {
		return &RPCError{Code: ErrCodeInvalidParams, Message: "cannot derive backup label"}
	}
	backupDir := filepath.Join(h.opts.UserHome, ".launchpal", "backups", label)
	rel := []string{".launchpal", "backups", label}
	id := h.opts.NowFn().Format("20060102-150405")
	backupPath := filepath.Join(backupDir, id+".plist")
	metaPath := filepath.Join(backupDir, id+".meta.json")
	// Create the backup directory chain AND write the leaf relative to the
	// per-component O_NOFOLLOW-resolved directory fd (symlinkSafeWriteInDir), so
	// a symlink at .launchpal / backups / <label> — whether pre-planted or
	// swapped in mid-operation (same-UID attack against the user's own home) —
	// cannot redirect the root-privileged write outside the backup tree. The
	// leaf is opened O_NOFOLLOW too. (Writing by absolute path would only guard
	// the leaf and re-follow intermediate symlinks.)
	if err := symlinkSafeWriteInDir(h.opts.UserHome, rel, id+".plist", existingData, 0644); err != nil {
		return &RPCError{Code: ErrCodeIOError, Message: err.Error()}
	}
	// chown the directory chain and the backup file to the launching user so the
	// user-side LaunchPal can read them. symlinkSafeWriteInDir created any
	// missing dirs; we chown idempotently — no-op when they already belonged to
	// the user.
	h.chownPath(filepath.Join(h.opts.UserHome, ".launchpal"))
	h.chownPath(filepath.Join(h.opts.UserHome, ".launchpal", "backups"))
	h.chownPath(backupDir)
	h.chownPath(backupPath)
	metaDoc := map[string]string{"originalPath": plistPath}
	if raw, err := json.Marshal(metaDoc); err == nil {
		_ = symlinkSafeWriteInDir(h.opts.UserHome, rel, id+".meta.json", raw, 0644)
		h.chownPath(metaPath)
	}
	return nil
}

// chownPath sets ownership to the launching user when configured. Non-fatal;
// a chown failure should not take down a write operation.
func (h *Handlers) chownPath(path string) {
	if h.opts.Chown == nil {
		return
	}
	if h.opts.LaunchingUID < 0 {
		return
	}
	_ = h.opts.Chown(path, h.opts.LaunchingUID, h.opts.LaunchingGID)
}

// errMsg renders a command failure for the RPC response: prefer stderr when
// present, fall back to the error string.
func errMsg(err error, stderr []byte) string {
	s := strings.TrimSpace(string(stderr))
	if s != "" {
		return s
	}
	if err != nil {
		return err.Error()
	}
	return ""
}
