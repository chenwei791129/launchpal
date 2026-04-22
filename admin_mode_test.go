package main

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"launchpal/internal/launchctl"
	"launchpal/internal/privhelper"
)

// fakeSystemMgr records SetAdminClient / ClearAdminClient calls and
// satisfies the systemManagerAdapter interface.
type fakeSystemMgr struct {
	mu      sync.Mutex
	set     int
	cleared int
	current launchctl.AdminClient
}

func (f *fakeSystemMgr) SetAdminClient(c launchctl.AdminClient) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.set++
	f.current = c
}

func (f *fakeSystemMgr) ClearAdminClient() {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.cleared++
	f.current = nil
}

// newTestAdminMode builds an adminModeManager with stubs for the helper
// launch path and helper binary resolution. launchFn returns the configured
// client (or error) instead of talking to osascript.
func newTestAdminMode(t *testing.T, launch launchHelperFunc, helperErr error) (*adminModeManager, *fakeSystemMgr, *hookRecorder) {
	t.Helper()
	sysMgr := &fakeSystemMgr{}
	recorder := &hookRecorder{}
	a := newAdminModeManager(sysMgr)
	a.helperPath = func() (string, error) {
		if helperErr != nil {
			return "", helperErr
		}
		return "/fake/helper", nil
	}
	a.launchFn = launch
	a.handshakeTO = 200 * time.Millisecond
	a.shutdownTO = 200 * time.Millisecond
	a.testHook = recorder.record
	return a, sysMgr, recorder
}

type hookRecorder struct {
	mu     sync.Mutex
	states []string
}

func (h *hookRecorder) record(state string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.states = append(h.states, state)
}

func (h *hookRecorder) snapshot() []string {
	h.mu.Lock()
	defer h.mu.Unlock()
	return append([]string(nil), h.states...)
}

// liveClient returns a *privhelper.Client backed by a paired server that
// responds OK to any method. Useful when the test needs Enable to reach the
// Enabled state.
func liveClient(t *testing.T) *privhelper.Client {
	t.Helper()
	client, cleanup := paired(t, func(req privhelper.Request) privhelper.Response {
		id := req.ID
		return privhelper.Response{ID: &id, Result: []byte(`{"ok":true}`)}
	})
	t.Cleanup(cleanup)
	return client
}

func TestAdminMode_InitialStateDisabled(t *testing.T) {
	a := newAdminModeManager(&fakeSystemMgr{})
	s := a.status()
	if s.State != AdminModeDisabled {
		t.Errorf("state = %q, want %q", s.State, AdminModeDisabled)
	}
	if s.Error != nil {
		t.Errorf("error = %v, want nil", s.Error)
	}
}

func TestAdminMode_Enable_SuccessfulPath(t *testing.T) {
	client := liveClient(t)
	launch := func(ctx context.Context, opts privhelper.LaunchHelperOptions) (*privhelper.Client, error) {
		return client, nil
	}
	a, sysMgr, hook := newTestAdminMode(t, launch, nil)

	if err := a.Enable(context.Background()); err != nil {
		t.Fatalf("Enable: %v", err)
	}
	if got := a.status().State; got != AdminModeEnabled {
		t.Errorf("state = %q, want %q", got, AdminModeEnabled)
	}
	if sysMgr.set != 1 {
		t.Errorf("SetAdminClient calls = %d, want 1", sysMgr.set)
	}
	// Transitions should have been Disabled → Requesting → Enabled.
	got := hook.snapshot()
	want := []string{AdminModeRequesting, AdminModeEnabled}
	if len(got) != len(want) {
		t.Fatalf("states = %v, want %v", got, want)
	}
	for i, s := range want {
		if got[i] != s {
			t.Errorf("state[%d] = %q, want %q", i, got[i], s)
		}
	}
}

func TestAdminMode_Enable_AuthorizationCancelled(t *testing.T) {
	launch := func(ctx context.Context, opts privhelper.LaunchHelperOptions) (*privhelper.Client, error) {
		return nil, privhelper.ErrAuthorizationCanceled
	}
	a, sysMgr, hook := newTestAdminMode(t, launch, nil)

	if err := a.Enable(context.Background()); err != nil {
		t.Errorf("Enable returned error for cancel: %v", err)
	}
	s := a.status()
	if s.State != AdminModeDisabled {
		t.Errorf("state = %q, want %q", s.State, AdminModeDisabled)
	}
	if s.Error == nil || *s.Error == "" {
		t.Errorf("expected non-nil error")
	}
	if sysMgr.set != 0 {
		t.Errorf("SetAdminClient should not have been called, got %d", sysMgr.set)
	}
	got := hook.snapshot()
	want := []string{AdminModeRequesting, AdminModeDisabled}
	if len(got) != 2 || got[0] != want[0] || got[1] != want[1] {
		t.Errorf("states = %v, want %v", got, want)
	}
}

func TestAdminMode_Enable_HelperBinaryNotFound(t *testing.T) {
	launch := func(ctx context.Context, opts privhelper.LaunchHelperOptions) (*privhelper.Client, error) {
		t.Fatal("launchFn should not be called when helperPath fails")
		return nil, nil
	}
	a, _, _ := newTestAdminMode(t, launch, errors.New("missing"))
	err := a.Enable(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
	if a.status().State != AdminModeDisabled {
		t.Errorf("state = %q", a.status().State)
	}
}

func TestAdminMode_Enable_HandshakeFailure(t *testing.T) {
	launch := func(ctx context.Context, opts privhelper.LaunchHelperOptions) (*privhelper.Client, error) {
		return nil, errors.New("helper handshake failed: connect timeout")
	}
	a, sysMgr, _ := newTestAdminMode(t, launch, nil)
	err := a.Enable(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
	if a.status().State != AdminModeDisabled {
		t.Errorf("state = %q", a.status().State)
	}
	if sysMgr.set != 0 {
		t.Errorf("SetAdminClient should not be called, got %d", sysMgr.set)
	}
}

func TestAdminMode_Enable_Idempotent(t *testing.T) {
	client := liveClient(t)
	calls := 0
	launch := func(ctx context.Context, opts privhelper.LaunchHelperOptions) (*privhelper.Client, error) {
		calls++
		return client, nil
	}
	a, _, _ := newTestAdminMode(t, launch, nil)

	if err := a.Enable(context.Background()); err != nil {
		t.Fatalf("Enable: %v", err)
	}
	if err := a.Enable(context.Background()); err != nil {
		t.Fatalf("Enable #2: %v", err)
	}
	if calls != 1 {
		t.Errorf("launchFn called %d times, want 1", calls)
	}
}

func TestAdminMode_Disable(t *testing.T) {
	client := liveClient(t)
	launch := func(ctx context.Context, opts privhelper.LaunchHelperOptions) (*privhelper.Client, error) {
		return client, nil
	}
	a, sysMgr, hook := newTestAdminMode(t, launch, nil)

	if err := a.Enable(context.Background()); err != nil {
		t.Fatalf("Enable: %v", err)
	}
	if err := a.Disable(context.Background()); err != nil {
		t.Fatalf("Disable: %v", err)
	}
	if a.status().State != AdminModeDisabled {
		t.Errorf("state = %q", a.status().State)
	}
	if sysMgr.cleared != 1 {
		t.Errorf("ClearAdminClient calls = %d, want 1", sysMgr.cleared)
	}
	got := hook.snapshot()
	want := []string{AdminModeRequesting, AdminModeEnabled, AdminModeShuttingDown, AdminModeDisabled}
	if len(got) != len(want) {
		t.Fatalf("states = %v, want %v", got, want)
	}
	for i, s := range want {
		if got[i] != s {
			t.Errorf("states[%d] = %q, want %q", i, got[i], s)
		}
	}
}

func TestAdminMode_HelperCrashSetsState(t *testing.T) {
	client := liveClient(t)
	launch := func(ctx context.Context, opts privhelper.LaunchHelperOptions) (*privhelper.Client, error) {
		return client, nil
	}
	a, sysMgr, _ := newTestAdminMode(t, launch, nil)

	if err := a.Enable(context.Background()); err != nil {
		t.Fatalf("Enable: %v", err)
	}

	a.handleHelperCrash(errors.New("boom"))
	s := a.status()
	if s.State != AdminModeDisabled {
		t.Errorf("state = %q", s.State)
	}
	if s.Error == nil || *s.Error != "helper_crashed" {
		t.Errorf("error = %v, want helper_crashed", s.Error)
	}
	if sysMgr.cleared != 1 {
		t.Errorf("ClearAdminClient calls = %d", sysMgr.cleared)
	}
}

// paired returns a privhelper.Client backed by net.Pipe with a goroutine
// implementing respond. Similar to the helper inside the privhelper package
// tests but duplicated here so admin_mode_test stays standalone.
func paired(t *testing.T, respond func(req privhelper.Request) privhelper.Response) (*privhelper.Client, func()) {
	t.Helper()
	return pairedInternal(t, respond)
}
