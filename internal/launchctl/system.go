package launchctl

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"syscall"

	"howett.net/plist"

	"launchpal/internal/plistutil"
	"launchpal/internal/privhelper"
)

// AdminClient is the minimal interface SystemManager needs from a
// privhelper.Client to drive write operations. It is defined here rather
// than importing the privhelper package to keep launchctl free of helper
// implementation details and easy to stub in tests.
type AdminClient interface {
	Bootstrap(ctx context.Context, plistPath string) error
	Bootout(ctx context.Context, label string) error
	Kickstart(ctx context.Context, label string) error
	WritePlist(ctx context.Context, plistPath string, data []byte) error
	DeletePlist(ctx context.Context, plistPath string) error
	EnsureLogAccess(ctx context.Context, paths []string) error
	TruncateLog(ctx context.Context, path string) error
	DeleteLogPaths(ctx context.Context, paths []string) (warnings []string, err error)
}

// SystemManager manages system LaunchDaemons in /Library/LaunchDaemons.
// Reads are always permitted. Writes require SetAdminClient to have been
// called with a live privhelper client (Admin Mode enabled); otherwise they
// return ErrReadOnlyManager.
type SystemManager struct {
	readOnlyManager

	mu          sync.RWMutex
	adminClient AdminClient
}

// NewSystemManager creates a new SystemManager
func NewSystemManager() *SystemManager {
	return &SystemManager{
		readOnlyManager: readOnlyManager{
			basePath:    "/Library/LaunchDaemons",
			serviceType: "system",
		},
	}
}

// SetAdminClient installs a privileged client so subsequent writes are
// delegated over the helper RPC. Passing nil reverts to read-only behavior
// (equivalent to ClearAdminClient).
func (m *SystemManager) SetAdminClient(c AdminClient) {
	m.mu.Lock()
	m.adminClient = c
	m.mu.Unlock()
}

// ClearAdminClient removes the installed admin client; after this call
// writes once again return ErrReadOnlyManager.
func (m *SystemManager) ClearAdminClient() {
	m.SetAdminClient(nil)
}

// client returns the currently-installed admin client, or nil if Admin Mode
// is off.
func (m *SystemManager) client() AdminClient {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.adminClient
}

func (m *SystemManager) List() ([]Service, error)             { return m.list() }
func (m *SystemManager) Get(name string) (*Service, error)    { return m.get(name) }
func (m *SystemManager) GetPlist(name string) (string, error) { return m.getPlist(name) }
func (m *SystemManager) GetPlistContent(name string) (*plistutil.Content, error) {
	return m.getPlistContent(name)
}
func (m *SystemManager) GetLogs(name string, logType string) (LogsResult, error) {
	return m.getLogs(name, logType)
}
func (m *SystemManager) GetLogClearStatus(name string, logType string) (LogClearStatus, error) {
	return m.getLogClearStatus(name, logType)
}

// ClearLogs truncates the configured stdout or stderr log for a system
// daemon. The dispatch is per-file: a direct OpenFile is tried first, and
// only EACCES falls back to the privileged helper. Any other errno
// (ENOENT, ELOOP, EISDIR, …) surfaces unchanged so the caller can
// distinguish a missing file from a permission gap. The errno comes from
// the OpenFile itself — pre-stat-then-open would race.
func (m *SystemManager) ClearLogs(name string, logType string) error {
	logPath, err := validateClearLogsArgs(m.Get, name, logType, false)
	if err != nil {
		return err
	}
	f, err := os.OpenFile(logPath, os.O_WRONLY|os.O_TRUNC|nofollowFlag, 0)
	if err == nil {
		return f.Close()
	}
	if os.IsNotExist(err) {
		return fmt.Errorf("log file does not exist: %s", logPath)
	}
	if !errors.Is(err, syscall.EACCES) {
		return err
	}
	// Re-fetch the client right before the helper call so a mid-flight
	// Admin-Mode toggle (e.g. idle timeout between status and clear) is
	// observed instead of using a stale handle.
	c := m.client()
	if c == nil {
		return ErrReadOnlyManager
	}
	return c.TruncateLog(context.Background(), logPath)
}

// RestorePlist writes raw plist bytes to plistPath via the privileged helper.
// The backup-restore flow uses this: the caller has the archived plist bytes
// and just needs them back on disk under /Library/LaunchDaemons/. We match
// the user-domain behaviour (BackupManager.Restore is a plain copyFile) and
// deliberately do NOT bootout + bootstrap — restoring is an explicit user
// action, and the detail page's Restart button is the right place to pick
// up the new config. Returns ErrReadOnlyManager when Admin Mode is off.
func (m *SystemManager) RestorePlist(plistPath string, data []byte) error {
	c := m.client()
	if c == nil {
		return ErrReadOnlyManager
	}
	return c.WritePlist(context.Background(), plistPath, data)
}

// Start makes a system daemon run. The order matters because launchd
// throttles fast-exiting jobs (minimum_runtime = 10s by default): issuing
// Bootstrap + Kickstart unconditionally means launchd spawns the job from
// RunAtLoad and then Kickstart kills it before it has lived long enough,
// causing a 10-second backoff before the next spawn.
//
//   - Bootstrap succeeds, RunAtLoad=true → launchd already spawned, done.
//   - Bootstrap succeeds, RunAtLoad=false → Kickstart to force spawn.
//   - Bootstrap fails (already loaded) → Kickstart to ensure running.
//
// Returns ErrReadOnlyManager when Admin Mode is disabled.
func (m *SystemManager) Start(name string) error {
	c := m.client()
	if c == nil {
		return ErrReadOnlyManager
	}
	if err := validateRoutingName(name); err != nil {
		return err
	}
	plistPath := filepath.Join(m.basePath, name+".plist")
	ctx := context.Background()

	bootstrapErr := c.Bootstrap(ctx, plistPath)

	svc, getErr := m.Get(name)
	if bootstrapErr == nil && getErr == nil && svc.RunAtLoad {
		// Fresh bootstrap + RunAtLoad=true: launchd will spawn the job on
		// its own. Kickstarting here would just trigger throttle.
		return nil
	}

	label := ""
	if getErr == nil {
		label = svc.Label
	}
	if label == "" {
		label = name
	}

	if err := c.Kickstart(ctx, label); err != nil {
		// Prefer the bootstrap error when both failed — it's usually the
		// more informative one (file not found, bad plist, etc.).
		if bootstrapErr != nil {
			return bootstrapErr
		}
		return err
	}
	return nil
}

// Stop boots out the daemon via the privileged helper when Admin Mode is
// enabled; otherwise returns ErrReadOnlyManager.
func (m *SystemManager) Stop(name string) error {
	c := m.client()
	if c == nil {
		return ErrReadOnlyManager
	}
	if err := validateRoutingName(name); err != nil {
		return err
	}
	svc, err := m.Get(name)
	if err != nil {
		return err
	}
	label := svc.Label
	if label == "" {
		label = name
	}
	return c.Bootout(context.Background(), label)
}

// Restart kickstarts (-k) the daemon via the privileged helper when Admin
// Mode is enabled; otherwise returns ErrReadOnlyManager.
func (m *SystemManager) Restart(name string) error {
	c := m.client()
	if c == nil {
		return ErrReadOnlyManager
	}
	if err := validateRoutingName(name); err != nil {
		return err
	}
	svc, err := m.Get(name)
	if err != nil {
		return err
	}
	label := svc.Label
	if label == "" {
		label = name
	}
	return c.Kickstart(context.Background(), label)
}

// Create writes a new daemon plist via the privileged helper and bootstraps
// it. Returns ErrReadOnlyManager when Admin Mode is disabled.
func (m *SystemManager) Create(config *ServiceConfig) error {
	c := m.client()
	if c == nil {
		return ErrReadOnlyManager
	}
	if config == nil || config.Label == "" {
		return fmt.Errorf("service label is required")
	}
	if err := validateRoutingName(config.Label); err != nil {
		return err
	}
	if err := validateProgramOrArguments(config); err != nil {
		return err
	}
	if err := validateSystemSchedule(config.Schedule); err != nil {
		return err
	}
	plistPath := filepath.Join(m.basePath, config.Label+".plist")
	data, err := encodePlist(config)
	if err != nil {
		return err
	}
	ctx := context.Background()
	if err := c.WritePlist(ctx, plistPath, data); err != nil {
		return err
	}
	// Make stdout/stderr tailable by the GUI after launchd creates them.
	// Failure is non-fatal: the daemon still runs and the Logs tab will just
	// surface "permission denied" if the path ends up unreachable.
	ensureLogPathsBestEffort(ctx, c, config)
	return c.Bootstrap(ctx, plistPath)
}

// Update overwrites the daemon plist via the privileged helper (the helper
// backs up the prior content) and forces launchd to drop its in-memory
// config. Without the Bootout step launchd continues running the previously
// bootstrapped definition forever — editing the plist on disk alone does
// nothing. Returns ErrReadOnlyManager when Admin Mode is disabled.
func (m *SystemManager) Update(name string, config *ServiceConfig) error {
	c := m.client()
	if c == nil {
		return ErrReadOnlyManager
	}
	if config == nil {
		return fmt.Errorf("config is required")
	}
	if err := validateRoutingName(name); err != nil {
		return err
	}
	if err := validateProgramOrArguments(config); err != nil {
		return err
	}
	if err := validateSystemSchedule(config.Schedule); err != nil {
		return err
	}
	plistPath := filepath.Join(m.basePath, name+".plist")
	ctx := context.Background()
	// Read the existing plist once, in the GUI process: it supplies both the
	// unmodeled keys to preserve and the label for Bootout. Modeled keys stay
	// form-authoritative. When the GUI lacks Full Disk Access it cannot read
	// the existing system plist, so readPlistMap fails and we degrade to a
	// fresh write. Even then we still Bootout using the routing name, which
	// equals the label by construction (Create writes <Label>.plist) — skipping
	// it would leave launchd running the old in-memory definition, and the new
	// plist on disk would never take effect until an explicit unload/reload or
	// reboot (Bootstrap fails on an already-loaded job and `kickstart -k` only
	// restarts the stale job rather than re-reading the plist).
	modeled := BuildPlistDict(config, false)
	oldLabel := name
	if existing, rerr := readPlistMap(plistPath); rerr == nil {
		modeled = mergeUnmodeledKeys(modeled, existing)
		if label, ok := existing["Label"].(string); ok && label != "" {
			oldLabel = label
		}
	}
	// Encode before Bootout so a failed encode leaves the running daemon
	// untouched rather than booted out with a stale plist still on disk.
	data, err := encodeDict(modeled)
	if err != nil {
		return err
	}
	// Bootout so launchd drops its in-memory config; writing the plist alone
	// leaves it running the previous definition forever. Best-effort: a "not
	// bootstrapped" error (e.g. the daemon was never loaded) is fine.
	if oldLabel != "" {
		_ = c.Bootout(ctx, oldLabel)
	}
	if err := c.WritePlist(ctx, plistPath, data); err != nil {
		return err
	}
	ensureLogPathsBestEffort(ctx, c, config)
	return nil
}

// ensureLogPathsBestEffort asks the helper to make StdoutPath/StderrPath
// traversable by the GUI user. Called from Create/Update after a successful
// WritePlist. Failures are intentionally swallowed: the daemon is already
// configured and the GUI will simply fall back to "permission denied" on
// the Logs tab if the preparation fails.
func ensureLogPathsBestEffort(ctx context.Context, c AdminClient, config *ServiceConfig) {
	if config == nil {
		return
	}
	paths := make([]string, 0, 2)
	if config.StdoutPath != "" {
		paths = append(paths, config.StdoutPath)
	}
	if config.StderrPath != "" {
		paths = append(paths, config.StderrPath)
	}
	if len(paths) == 0 {
		return
	}
	_ = c.EnsureLogAccess(ctx, paths)
}

// Delete removes the daemon plist via the privileged helper (after taking a
// backup). Returns ErrReadOnlyManager when Admin Mode is disabled. A
// not_found from the helper is treated as success so Delete is idempotent
// from the GUI's perspective — probing the filesystem here would be
// unreliable because the GUI process often can't read /Library/LaunchDaemons
// without Full Disk Access.
func (m *SystemManager) Delete(name string) error {
	return m.DeleteWithOptions(name, DeleteServiceOptions{})
}

// DeleteWithOptions extends Delete with optional log-file cleanup. The
// Manager interface deliberately does NOT include this method: only
// SystemManager owns paths under the system-domain log allowlist that the
// helper can safely act on, so widening the interface would force the
// user-domain and apple-system managers to implement a flow they don't
// support. Returns ErrReadOnlyManager when Admin Mode is disabled, and a
// *LogDeletionWarning when the plist was removed cleanly but one or more
// log paths failed — callers should treat that as overall success and
// surface the entries as a non-fatal warning.
func (m *SystemManager) DeleteWithOptions(name string, opts DeleteServiceOptions) error {
	c := m.client()
	if c == nil {
		return ErrReadOnlyManager
	}
	if err := validateRoutingName(name); err != nil {
		return err
	}
	plistPath := filepath.Join(m.basePath, name+".plist")
	// Capture log paths from the parsed plist BEFORE deletion. After
	// DeletePlist succeeds the file is gone and a fresh Get would fail; the
	// helper's backup is also out of reach for the GUI without the helper's
	// cooperation, so reading them now is the simplest source. If Get fails
	// (typically Full Disk Access denied) we record getFailed so the
	// log-cleanup path can surface an auditable warning instead of silently
	// skipping the user's explicit request.
	var (
		logPaths  []string
		getFailed bool
	)
	if svc, err := m.Get(name); err == nil {
		if svc.Label != "" {
			_ = c.Bootout(context.Background(), svc.Label)
		}
		if opts.DeleteLogs {
			if svc.StdoutPath != "" {
				logPaths = append(logPaths, svc.StdoutPath)
			}
			if svc.StderrPath != "" {
				logPaths = append(logPaths, svc.StderrPath)
			}
		}
	} else if opts.DeleteLogs {
		getFailed = true
	}

	if err := c.DeletePlist(context.Background(), plistPath); err != nil {
		var rpcErr *privhelper.RPCError
		if !errors.As(err, &rpcErr) || rpcErr.Code != privhelper.ErrCodeNotFound {
			return err
		}
	}

	if !opts.DeleteLogs {
		return nil
	}
	if getFailed {
		return &LogDeletionWarning{Errors: []string{
			"could not read plist to determine log paths; log files were not deleted (Full Disk Access may be required)",
		}}
	}
	if len(logPaths) == 0 {
		return nil
	}
	warnings, dlErr := c.DeleteLogPaths(context.Background(), logPaths)
	if dlErr != nil {
		// Transport / param failure: surface as warning so the daemon delete
		// (the operation the user actually asked for) still reads as success.
		return &LogDeletionWarning{Errors: []string{dlErr.Error()}}
	}
	if len(warnings) > 0 {
		return &LogDeletionWarning{Errors: warnings}
	}
	return nil
}

// maxSystemCalendarEntries bounds the number of StartCalendarInterval entries a
// system daemon create/update may write, matching the frontend cron
// range-expansion cap (ScheduleForm.vue MAX_EXPANSION). It is enforced in the
// create/update path — which can return an error and write no plist — NOT in
// buildCalendarInterval/BuildPlistDict, which have no error channel and are
// shared with the user domain whose behavior this change must not alter.
const maxSystemCalendarEntries = 50

// validateSystemSchedule brings system daemon schedule validation to parity
// with the user domain: the shared range check (validateSchedule) plus the
// system-domain-only 50-entry cap. On failure the caller writes no plist.
func validateSystemSchedule(s *ScheduleConfig) error {
	if err := validateSchedule(s); err != nil {
		return err
	}
	if s != nil && len(s.Schedules) > maxSystemCalendarEntries {
		return fmt.Errorf("schedule has %d calendar entries, exceeding the %d-entry limit", len(s.Schedules), maxSystemCalendarEntries)
	}
	return nil
}

// encodePlist marshals a ServiceConfig into an XML plist suitable for
// /Library/LaunchDaemons. Uses BuildPlistDict to share the field-mapping
// logic with UserManager; the expandPaths=false argument keeps paths
// verbatim because a root-written daemon should not resolve `~`.
func encodePlist(config *ServiceConfig) ([]byte, error) {
	return encodeDict(BuildPlistDict(config, false))
}

// encodeDict marshals an already-built plist dict into indented XML bytes.
// Split out from encodePlist so Update can encode a dict that has already been
// merged with the existing plist's unmodeled keys.
func encodeDict(pd map[string]any) ([]byte, error) {
	var buf bytes.Buffer
	enc := plist.NewEncoder(&buf)
	enc.Indent("\t")
	if err := enc.Encode(pd); err != nil {
		return nil, fmt.Errorf("encode plist: %w", err)
	}
	return buf.Bytes(), nil
}

// Ensure SystemManager implements Manager interface
var _ Manager = (*SystemManager)(nil)
