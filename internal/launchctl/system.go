package launchctl

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"path/filepath"
	"sync"

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
func (m *SystemManager) GetLogs(name string, logType string) (string, error) {
	return m.getLogs(name, logType)
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
	plistPath := filepath.Join(m.basePath, name+".plist")
	data, err := encodePlist(config)
	if err != nil {
		return err
	}
	ctx := context.Background()
	// Resolve the old label from disk before we overwrite; Bootout is a
	// best-effort step so an error (e.g. "not bootstrapped") is OK.
	if svc, getErr := m.Get(name); getErr == nil && svc.Label != "" {
		_ = c.Bootout(ctx, svc.Label)
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
	c := m.client()
	if c == nil {
		return ErrReadOnlyManager
	}
	plistPath := filepath.Join(m.basePath, name+".plist")
	// Ignore bootout error if the daemon isn't loaded — the following
	// DeletePlist is the operation the user actually asked for.
	if svc, err := m.Get(name); err == nil && svc.Label != "" {
		_ = c.Bootout(context.Background(), svc.Label)
	}
	err := c.DeletePlist(context.Background(), plistPath)
	if err == nil {
		return nil
	}
	var rpcErr *privhelper.RPCError
	if errors.As(err, &rpcErr) && rpcErr.Code == privhelper.ErrCodeNotFound {
		return nil
	}
	return err
}

// encodePlist marshals a ServiceConfig into an XML plist suitable for
// /Library/LaunchDaemons. Uses BuildPlistDict to share the field-mapping
// logic with UserManager; the expandPaths=false argument keeps paths
// verbatim because a root-written daemon should not resolve `~`.
func encodePlist(config *ServiceConfig) ([]byte, error) {
	pd := BuildPlistDict(config, false)
	var buf bytes.Buffer
	enc := plist.NewEncoder(&buf)
	enc.Indent("\t")
	if err := enc.Encode(pd); err != nil {
		return nil, fmt.Errorf("encode plist: %w", err)
	}
	return buf.Bytes(), nil
}

// encodePlistBase64 is a convenience used by tests and the admin-client
// wrapper in app.go: marshal + base64 in one step.
func encodePlistBase64(config *ServiceConfig) (string, error) {
	data, err := encodePlist(config)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(data), nil
}

// Ensure SystemManager implements Manager interface
var _ Manager = (*SystemManager)(nil)
