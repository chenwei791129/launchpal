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
		// Integrity failure is distinct from a plain missing binary: refuse to
		// launch osascript and surface a dedicated code. The message carries
		// only the code so it reveals nothing beyond the existing errors.
		if errors.Is(err, errHelperIntegrity) {
			a.failFromRequesting("helper_integrity_failed", "")
			return err
		}
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

// errHelperIntegrity signals that no verified protected copy exists and the
// bundle helper fails hash-pin verification (missing/unreadable, or its
// SHA-256 differs from a non-empty pin). Enable maps it to the
// helper_integrity_failed error code and refuses to launch osascript.
var errHelperIntegrity = errors.New("helper_integrity_failed")

// helperResolution carries the inputs of the launch-path decision so the
// logic is unit-testable without touching os.Executable, the real protected
// path, or the embedded pin. Production wiring is in resolveHelperPath.
type helperResolution struct {
	bundlePath    string
	protectedPath string
	pin           string
	isVerified    func(string) bool
	fileHash      func(string) (string, error)
}

// resolveHelperPath returns the path of the helper binary LaunchPal should
// launch. The trust of the protected copy derives solely from its root
// ownership and permissions, never from matching the bundle hash — see
// resolveHelperLaunchPath for the full decision.
func resolveHelperPath() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("os.Executable: %w", err)
	}
	bundle := filepath.Join(filepath.Dir(exe), "launchpal-privhelper")
	return resolveHelperLaunchPath(helperResolution{
		bundlePath:    bundle,
		protectedPath: privhelper.ProtectedHelperPath,
		pin:           helperPin,
		isVerified:    privhelper.IsVerifiedProtectedCopy,
		fileHash:      privhelper.FileSHA256,
	})
}

// resolveHelperLaunchPath applies the launch-path priority:
//
//  1. A verified protected copy is launched by default. The only exception is
//     a legitimate update — the pin is non-empty, the bundle is readable, the
//     bundle hash equals the pin, and the bundle differs from the protected
//     copy — in which case the bundle copy is launched to re-provision.
//  2. With no verified protected copy, the bundle copy is launched after
//     hash-pin verification (or, with an empty pin, without it). A non-empty
//     pin that the bundle fails (missing/unreadable, or hash mismatch) yields
//     errHelperIntegrity; an empty pin with a missing bundle yields a plain
//     not-found error naming the path.
//
// The key invariant: an existing verified protected copy is never bypassed by
// an empty pin, a missing bundle, or a tampered bundle.
func resolveHelperLaunchPath(r helperResolution) (string, error) {
	if r.isVerified(r.protectedPath) {
		if isLegitimateBundleUpdate(r) {
			// Bundle proven current and differs from the protected copy →
			// launch it to re-provision.
			return r.bundlePath, nil
		}
		return r.protectedPath, nil
	}

	// No verified protected copy → first-install path via the bundle.
	bundleHash, err := r.fileHash(r.bundlePath)
	if err != nil {
		if r.pin != "" {
			// Non-empty pin but the bundle can't be verified.
			return "", errHelperIntegrity
		}
		return "", fmt.Errorf("helper binary not found at %s: %w", r.bundlePath, err)
	}
	if r.pin != "" && bundleHash != r.pin {
		return "", errHelperIntegrity
	}
	return r.bundlePath, nil
}

// isLegitimateBundleUpdate reports whether the bundle copy is a proven-current
// app update that differs from the existing protected copy — the sole case in
// which a verified protected copy is bypassed. It requires a non-empty pin, a
// readable bundle whose hash equals the pin, and a protected copy whose hash
// differs from that pin (equivalently, differs from the bundle).
func isLegitimateBundleUpdate(r helperResolution) bool {
	if r.pin == "" {
		return false
	}
	// Hash the protected copy first and bail out in the steady state
	// (protected copy already current, or unreadable) — this skips the
	// multi-MB bundle hash on the common Enable path, which touches the bundle
	// only for an actual update.
	protectedHash, err := r.fileHash(r.protectedPath)
	if err != nil || protectedHash == r.pin {
		return false
	}
	bundleHash, err := r.fileHash(r.bundlePath)
	return err == nil && bundleHash == r.pin
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
