package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"launchpal/internal/launchctl"
	"launchpal/internal/privhelper"

	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// Admin Mode state constants. These strings are what the frontend compares
// against, so the wire values are part of the UI contract.
const (
	AdminModeDisabled     = "disabled"
	AdminModeRequesting   = "requesting"
	AdminModeEnabled      = "enabled"
	AdminModeShuttingDown = "shutting_down"
)

// AdminModeStatus is the shape exposed to the frontend. Error is a pointer
// so it can round-trip through the Wails JSON binding as `null` when empty.
type AdminModeStatus struct {
	State string  `json:"state"`
	Error *string `json:"error"`
}

// adminModeManager coordinates Admin Mode state and the helper client.
// Methods are safe for concurrent use; the mutex protects all mutable
// fields (state, err, client).
type adminModeManager struct {
	mu          sync.Mutex
	state       string
	lastErr     string // "" when no error
	client      *privhelper.Client
	systemMgr   systemManagerAdapter
	event       adminModeEventEmitter
	helperPath  func() (string, error) // resolves the helper binary path at runtime
	launchFn    launchHelperFunc       // injectable for tests
	nowFn       func() time.Time
	currentUID  int
	currentGID  int
	userHome    string
	currentPID  int
	handshakeTO time.Duration
	shutdownTO  time.Duration
	testHook    func(state string) // optional; called on every transition
}

// adminModeEventEmitter notifies the frontend when the state changes. The
// Wails runtime implements it via EventsEmit; tests use a no-op.
type adminModeEventEmitter interface {
	Emit(name string, data ...any)
}

// launchHelperFunc is the signature of privhelper.LaunchHelper; injectable
// for tests so we don't run osascript.
type launchHelperFunc func(ctx context.Context, opts privhelper.LaunchHelperOptions) (*privhelper.Client, error)

// systemManagerAdapter is the subset of SystemManager Admin Mode needs.
// SetAdminClient accepts a launchctl.AdminClient so the package is free to
// swap in a test double; *privhelper.Client implements the interface.
type systemManagerAdapter interface {
	SetAdminClient(c launchctl.AdminClient)
	ClearAdminClient()
}

// newAdminModeManager constructs an Admin Mode manager with production
// defaults. systemMgr receives the live client when Admin Mode enables.
func newAdminModeManager(systemMgr systemManagerAdapter) *adminModeManager {
	return &adminModeManager{
		state:       AdminModeDisabled,
		systemMgr:   systemMgr,
		helperPath:  resolveHelperPath,
		launchFn:    privhelper.LaunchHelper,
		nowFn:       time.Now,
		currentUID:  os.Getuid(),
		currentGID:  os.Getgid(),
		currentPID:  os.Getpid(),
		handshakeTO: 10 * time.Second,
		shutdownTO:  3 * time.Second,
		userHome: func() string {
			if h, err := os.UserHomeDir(); err == nil {
				return h
			}
			return os.Getenv("HOME")
		}(),
	}
}

// status returns the current AdminModeStatus; safe for concurrent callers.
func (a *adminModeManager) status() AdminModeStatus {
	a.mu.Lock()
	defer a.mu.Unlock()
	var err *string
	if a.lastErr != "" {
		s := a.lastErr
		err = &s
	}
	return AdminModeStatus{State: a.state, Error: err}
}

// setState transitions and notifies the frontend. Must be called with mu held.
func (a *adminModeManager) setState(state string, errMsg string) {
	a.state = state
	a.lastErr = errMsg
	if a.testHook != nil {
		a.testHook(state)
	}
	if a.event != nil {
		a.event.Emit("admin_mode:state", AdminModeStatus{State: state, Error: maybeStrPtr(errMsg)})
	}
}

func maybeStrPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// Enable starts the helper if Admin Mode is currently Disabled. If already
// Enabled it is a no-op (spec: EnableAdminMode while already enabled).
func (a *adminModeManager) Enable(ctx context.Context) error {
	a.mu.Lock()
	if a.state == AdminModeEnabled || a.state == AdminModeRequesting {
		a.mu.Unlock()
		return nil
	}
	a.setState(AdminModeRequesting, "")
	a.mu.Unlock()

	socketPath, err := privhelper.GenerateSocketPath(a.currentUID)
	if err != nil {
		a.failFromRequesting("socket_path_failed", err.Error())
		return err
	}

	helperPath, err := a.helperPath()
	if err != nil {
		a.failFromRequesting("helper_binary_not_found", err.Error())
		return err
	}

	client, err := a.launchFn(ctx, privhelper.LaunchHelperOptions{
		HelperPath:       helperPath,
		SocketPath:       socketPath,
		ParentPID:        a.currentPID,
		LaunchingUID:     a.currentUID,
		UserHome:         a.userHome,
		HandshakeTimeout: a.handshakeTO,
		OnDisconnect:     a.handleHelperCrash,
	})
	if err != nil {
		if errors.Is(err, privhelper.ErrAuthorizationCanceled) {
			a.failFromRequesting("authorization_cancelled", "User cancelled authorization")
			return nil
		}
		a.failFromRequesting("helper_handshake_failed", err.Error())
		return err
	}

	// Install the crash callback so we drop Admin Mode if the helper dies.
	// We can't do this in NewClient because the current design builds the
	// client inside LaunchHelper; a future refactor could expose it. For
	// now, poll via Ping-on-error is sufficient and keeps the Client struct
	// smaller. (The server-side tests already cover OnDisconnect behavior.)
	a.mu.Lock()
	a.client = client
	a.systemMgr.SetAdminClient(client)
	a.setState(AdminModeEnabled, "")
	a.mu.Unlock()

	return nil
}

// Disable sends Shutdown, waits up to shutdownTO, and returns to Disabled.
// If the helper is unresponsive the connection is closed anyway.
func (a *adminModeManager) Disable(ctx context.Context) error {
	a.mu.Lock()
	if a.state != AdminModeEnabled {
		a.mu.Unlock()
		return nil
	}
	a.setState(AdminModeShuttingDown, "")
	client := a.client
	a.mu.Unlock()

	if client != nil {
		shutdownCtx, cancel := context.WithTimeout(ctx, a.shutdownTO)
		_ = client.Shutdown(shutdownCtx)
		cancel()
		_ = client.Close()
	}

	a.mu.Lock()
	a.client = nil
	a.systemMgr.ClearAdminClient()
	a.setState(AdminModeDisabled, "")
	a.mu.Unlock()
	return nil
}

// handleHelperCrash is wired to Client.OnDisconnect for unexpected exits.
// Clean shutdowns (state == ShuttingDown) don't surface as crashes.
func (a *adminModeManager) handleHelperCrash(_ error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.state != AdminModeEnabled {
		return
	}
	a.client = nil
	a.systemMgr.ClearAdminClient()
	a.setState(AdminModeDisabled, "helper_crashed")
}

// failFromRequesting transitions from Requesting → Disabled with a reason
// code. Used by every failure path out of Enable.
func (a *adminModeManager) failFromRequesting(code, message string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.state != AdminModeRequesting {
		return
	}
	msg := code
	if message != "" {
		msg = code + ": " + message
	}
	a.setState(AdminModeDisabled, msg)
}

// resolveHelperPath returns the sibling launchpal-privhelper binary path.
// Resolved at runtime so we locate it inside the .app bundle regardless of
// how the app was invoked.
func resolveHelperPath() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("os.Executable: %w", err)
	}
	dir := filepath.Dir(exe)
	helper := filepath.Join(dir, "launchpal-privhelper")
	if _, err := os.Stat(helper); err != nil {
		return "", fmt.Errorf("helper binary not found at %s: %w", helper, err)
	}
	return helper, nil
}

// wailsEventEmitter adapts the Wails runtime's EventsEmit to our emitter
// interface so tests can use a no-op without pulling the Wails runtime.
type wailsEventEmitter struct {
	ctx context.Context
}

func (e *wailsEventEmitter) Emit(name string, data ...any) {
	if e.ctx == nil {
		return
	}
	wailsruntime.EventsEmit(e.ctx, name, data...)
}
