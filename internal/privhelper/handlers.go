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
	"strconv"
	"strings"
	"time"
)

// SystemDaemonDir is the one directory under which the helper accepts writes.
// Paths outside this directory are rejected with ErrCodeInvalidParams.
// /System/Library/LaunchDaemons belongs to Apple (and SIP blocks writes there
// anyway) so only this directory is ever a valid target.
const SystemDaemonDir = "/Library/LaunchDaemons"

// allowedLogPathPrefixes bounds EnsureLogAccess to locations where creating
// a world-readable file is harmless. macOS resolves /var and /tmp through
// /private, so both aliases are listed — filepath.Clean doesn't cross
// symlinks, so a path given as /var/log/x.log will still have that prefix
// when validated.
var allowedLogPathPrefixes = []string{
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

// DaemonLister is the subset of launchctl-list behavior the helper exposes.
// Injectable for tests.
type DaemonLister interface {
	ListDaemons(ctx context.Context) ([]DaemonInfo, error)
	GetDaemon(ctx context.Context, label string) (DaemonInfo, error)
}

// Chowner sets filesystem ownership. Injectable so tests don't need CAP_CHOWN.
type Chowner func(path string, uid, gid int) error

// HandlerOptions configure a set of RPC handlers.
type HandlerOptions struct {
	// Runner executes launchctl commands.
	Runner Runner

	// Lister fetches daemon status/pid. Defaults to launchctlLister which
	// parses `launchctl print system`.
	Lister DaemonLister

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
	if opts.Lister == nil {
		opts.Lister = &launchctlLister{runner: opts.Runner}
	}
	return &Handlers{opts: opts}
}

// Handle is the Server Handler callback; it looks up the method and
// forwards to the concrete implementation.
func (h *Handlers) Handle(ctx context.Context, req *Request) (any, *RPCError) {
	switch req.Method {
	case MethodPing:
		return PingResult{Pong: true}, nil
	case MethodListSystemDaemons:
		return h.listDaemons(ctx)
	case MethodGetSystemDaemon:
		return h.getDaemon(ctx, req.Params)
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
	_, stderr, err := h.opts.Runner.Run(ctx, "launchctl", "bootstrap", "system", clean)
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
	_, stderr, err := h.opts.Runner.Run(ctx, "launchctl", "bootout", "system/"+p.Label)
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
	_, stderr, err := h.opts.Runner.Run(ctx, "launchctl", "kickstart", "-k", "system/"+p.Label)
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
// well-known log/temp prefixes. The cleaned path is returned so the caller
// works with a canonical form. Rejects:
//   - non-absolute paths
//   - paths not under an allowed prefix (post-Clean, so "/var/log/../etc"
//     is normalized to "/etc" and rejected)
//   - paths ending in a separator (we require a file, not a bare directory)
//   - paths whose immediate parent is the allowlist prefix itself — e.g.
//     "/var/log/foo.log" would force a Chmod on /var/log, which we refuse
//     because /var/log is a system directory we shouldn't re-mode.
func validateLogPath(path string) (string, *RPCError) {
	if path == "" {
		return "", &RPCError{Code: ErrCodeInvalidParams, Message: "log path must not be empty"}
	}
	if !filepath.IsAbs(path) {
		return "", &RPCError{Code: ErrCodeInvalidParams, Message: "log path must be absolute"}
	}
	clean := filepath.Clean(path)
	// Each allowlist prefix ends in "/", so HasPrefix naturally rejects
	// both the prefix itself ("/tmp" has no trailing separator, so won't
	// start with "/tmp/") and look-alike siblings ("/var/logX" won't start
	// with "/var/log/").
	var matchedPrefix string
	for _, prefix := range allowedLogPathPrefixes {
		if strings.HasPrefix(clean, prefix) {
			matchedPrefix = prefix
			break
		}
	}
	if matchedPrefix == "" {
		return "", &RPCError{Code: ErrCodeInvalidParams, Message: "log path not under an allowed prefix"}
	}
	// Require at least one sub-directory between the prefix root and the
	// file — we don't want to chmod /var/log itself. remainder is e.g.
	// "myservice/out.log" for an accepted path.
	remainder := clean[len(matchedPrefix):]
	if !strings.Contains(remainder, string(filepath.Separator)) {
		return "", &RPCError{Code: ErrCodeInvalidParams, Message: "log path must live in a sub-directory of " + strings.TrimSuffix(matchedPrefix, "/")}
	}
	return clean, nil
}

// ensureLogAccess prepares log-file locations so the unprivileged GUI can
// tail them. For each validated path:
//   - MkdirAll(parent, 0755) so traversal bits are present.
//   - Chmod(parent, 0755) in case the parent already existed with a
//     restrictive mode (launchd creates per-service log dirs as root
//     0744, which blocks the user from even entering the directory).
//   - If the file does not yet exist, touch it as 0644 with O_NOFOLLOW so a
//     pre-planted symlink can't redirect the root-privileged create.
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
		clean, errR := validateLogPath(path)
		if errR != nil {
			return nil, errR
		}
		dir := filepath.Dir(clean)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, &RPCError{Code: ErrCodeIOError, Message: err.Error()}
		}
		if err := os.Chmod(dir, 0755); err != nil {
			return nil, &RPCError{Code: ErrCodeIOError, Message: err.Error()}
		}
		if _, err := os.Lstat(clean); os.IsNotExist(err) {
			f, err := os.OpenFile(clean, os.O_WRONLY|os.O_CREATE|os.O_EXCL|syscallNoFollow, 0644)
			if err != nil {
				if os.IsExist(err) {
					// Lost the race with another writer; honor whatever is
					// there now.
					continue
				}
				return nil, &RPCError{Code: ErrCodeIOError, Message: err.Error()}
			}
			_ = f.Close()
		} else if err != nil {
			return nil, &RPCError{Code: ErrCodeIOError, Message: err.Error()}
		}
	}
	return OKResult{OK: true}, nil
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

func (h *Handlers) listDaemons(ctx context.Context) (any, *RPCError) {
	daemons, err := h.opts.Lister.ListDaemons(ctx)
	if err != nil {
		return nil, &RPCError{Code: ErrCodeLaunchctlFailed, Message: err.Error()}
	}
	return ListSystemDaemonsResult{Daemons: daemons}, nil
}

func (h *Handlers) getDaemon(ctx context.Context, raw json.RawMessage) (any, *RPCError) {
	p, errR := unmarshalParams[GetSystemDaemonParams](raw)
	if errR != nil {
		return nil, errR
	}
	if errR := validateLabel(p.Label); errR != nil {
		return nil, errR
	}
	info, err := h.opts.Lister.GetDaemon(ctx, p.Label)
	if err != nil {
		if errors.Is(err, errDaemonNotFound) {
			return nil, &RPCError{Code: ErrCodeNotFound, Message: p.Label}
		}
		return nil, &RPCError{Code: ErrCodeLaunchctlFailed, Message: err.Error()}
	}
	return info, nil
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
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return &RPCError{Code: ErrCodeIOError, Message: err.Error()}
	}
	// chown newly-created directories up the chain to the user so they can
	// traverse. MkdirAll may or may not have created intermediate dirs; we
	// chown idempotently — no-op when they already belonged to the user.
	h.chownPath(filepath.Join(h.opts.UserHome, ".launchpal"))
	h.chownPath(filepath.Join(h.opts.UserHome, ".launchpal", "backups"))
	h.chownPath(backupDir)

	id := h.opts.NowFn().Format("20060102-150405")
	backupPath := filepath.Join(backupDir, id+".plist")
	metaPath := filepath.Join(backupDir, id+".meta.json")
	// O_NOFOLLOW refuses to open a pre-existing symlink — without this,
	// a user-owned symlink at backupPath would redirect a root write to an
	// arbitrary path.
	if err := writeNoFollow(backupPath, existingData, 0644); err != nil {
		return &RPCError{Code: ErrCodeIOError, Message: err.Error()}
	}
	h.chownPath(backupPath)
	metaDoc := map[string]string{"originalPath": plistPath}
	if raw, err := json.Marshal(metaDoc); err == nil {
		_ = writeNoFollow(metaPath, raw, 0644)
		h.chownPath(metaPath)
	}
	return nil
}

// writeNoFollow writes data to path with O_NOFOLLOW semantics so an
// attacker-planted symlink cannot redirect a root-privileged write to an
// arbitrary target. The file is truncated if it exists as a regular file.
func writeNoFollow(path string, data []byte, mode os.FileMode) error {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC|syscallNoFollow, mode)
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

// errDaemonNotFound is returned by DaemonLister implementations when a label
// is not found, so handlers can map it to ErrCodeNotFound.
var errDaemonNotFound = errors.New("daemon not found")

// launchctlLister implements DaemonLister by parsing `launchctl print system`
// (list) and `launchctl print system/<label>` (single).
type launchctlLister struct {
	runner Runner
}

// ListDaemons returns the bootstrapped daemons under system/ by parsing the
// "services" block of `launchctl print system`.
func (l *launchctlLister) ListDaemons(ctx context.Context) ([]DaemonInfo, error) {
	stdout, stderr, err := l.runner.Run(ctx, "launchctl", "print", "system")
	if err != nil {
		return nil, fmt.Errorf("launchctl print system: %s", errMsg(err, stderr))
	}
	return parsePrintSystem(stdout), nil
}

// GetDaemon returns the single daemon entry for label by parsing
// `launchctl print system/<label>`.
func (l *launchctlLister) GetDaemon(ctx context.Context, label string) (DaemonInfo, error) {
	stdout, stderr, err := l.runner.Run(ctx, "launchctl", "print", "system/"+label)
	if err != nil {
		s := strings.ToLower(strings.TrimSpace(string(stderr)))
		if strings.Contains(s, "could not find service") || strings.Contains(s, "no such") {
			return DaemonInfo{}, errDaemonNotFound
		}
		return DaemonInfo{}, fmt.Errorf("launchctl print system/%s: %s", label, errMsg(err, stderr))
	}
	return parsePrintService(label, stdout), nil
}

// parsePrintSystem extracts labels/status/pid from the "services" block of
// `launchctl print system`. Format (roughly):
//
//	services = {
//		12345  -  com.example.running
//		-      0 com.example.stopped
//		...
//	}
//
// We look for the `services = {` opening, read until the closing `}`, and
// parse each line within.
func parsePrintSystem(output []byte) []DaemonInfo {
	lines := strings.Split(string(output), "\n")
	inBlock := false
	out := []DaemonInfo{}
	for _, raw := range lines {
		line := strings.TrimSpace(raw)
		if !inBlock {
			if strings.HasPrefix(line, "services") && strings.Contains(line, "{") {
				inBlock = true
			}
			continue
		}
		if strings.HasPrefix(line, "}") {
			break
		}
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}
		info := DaemonInfo{Label: fields[len(fields)-1]}
		// fields[0] is PID or "-"; fields[1] is exit code ("0" / "-"). A
		// numeric PID means the daemon is running.
		if pid, ok := parsePIDField(fields[0]); ok {
			info.PID = pid
			info.Status = "running"
		} else {
			info.Status = "stopped"
		}
		out = append(out, info)
	}
	return out
}

// parsePrintService pulls PID (if any) from `launchctl print system/<label>`
// output. The output has a top-level `state = <running|not running>` and a
// `pid = N` line when running.
func parsePrintService(label string, output []byte) DaemonInfo {
	info := DaemonInfo{Label: label, Status: "stopped"}
	for _, raw := range strings.Split(string(output), "\n") {
		line := strings.TrimSpace(raw)
		if strings.HasPrefix(line, "state = ") {
			state := strings.TrimPrefix(line, "state = ")
			if strings.Contains(state, "running") {
				info.Status = "running"
			}
		}
		if strings.HasPrefix(line, "pid = ") {
			if pid, ok := parsePIDField(strings.TrimPrefix(line, "pid = ")); ok {
				info.PID = pid
			}
		}
	}
	return info
}

// parsePIDField returns (pid, true) for a positive numeric field, (0, false)
// for "-" or any unparseable value.
func parsePIDField(s string) (int, bool) {
	n, err := strconv.Atoi(strings.TrimSpace(s))
	if err != nil || n <= 0 {
		return 0, false
	}
	return n, true
}
