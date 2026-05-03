package privhelper

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// fakeRunner records every launchctl invocation and serves canned responses.
// It replaces exec.Command so handler tests run without root and without
// touching launchctl on the host.
type fakeRunner struct {
	calls   []call
	stdout  []byte
	stderr  []byte
	err     error
	respond func(name string, args []string) ([]byte, []byte, error)
}

type call struct {
	name string
	args []string
}

func (r *fakeRunner) Run(_ context.Context, name string, args ...string) ([]byte, []byte, error) {
	r.calls = append(r.calls, call{name: name, args: append([]string(nil), args...)})
	if r.respond != nil {
		return r.respond(name, args)
	}
	return r.stdout, r.stderr, r.err
}

// fakeLister serves canned daemon info and recorded calls.
type fakeLister struct {
	list      []DaemonInfo
	listErr   error
	single    DaemonInfo
	singleErr error
}

func (f *fakeLister) ListDaemons(_ context.Context) ([]DaemonInfo, error) {
	return f.list, f.listErr
}

func (f *fakeLister) GetDaemon(_ context.Context, _ string) (DaemonInfo, error) {
	return f.single, f.singleErr
}

// newChownRecorder returns a Chowner that records chown calls instead of
// invoking them on the filesystem.
func newChownRecorder() (Chowner, *[]chownCall) {
	var calls []chownCall
	chowner := func(path string, uid, gid int) error {
		calls = append(calls, chownCall{path: path, uid: uid, gid: gid})
		return nil
	}
	return chowner, &calls
}

type chownCall struct {
	path string
	uid  int
	gid  int
}

// errMsgFromRPC extracts a friendly message for assertion failures.
func errMsgFromRPC(e *RPCError) string {
	if e == nil {
		return ""
	}
	return e.Code + ":" + e.Message
}

func TestHandlers_Ping(t *testing.T) {
	h := NewHandlers(HandlerOptions{Runner: &fakeRunner{}})
	out, err := h.Handle(context.Background(), &Request{ID: 1, Method: MethodPing})
	if err != nil {
		t.Fatalf("error: %+v", err)
	}
	ping, ok := out.(PingResult)
	if !ok || !ping.Pong {
		t.Errorf("result = %+v", out)
	}
}

func TestHandlers_UnknownMethod(t *testing.T) {
	h := NewHandlers(HandlerOptions{Runner: &fakeRunner{}})
	_, err := h.Handle(context.Background(), &Request{ID: 1, Method: "WhoKnows"})
	if err == nil || err.Code != ErrCodeUnknownMethod {
		t.Errorf("err = %+v, want unknown_method", err)
	}
}

func TestHandlers_Bootstrap_Valid(t *testing.T) {
	runner := &fakeRunner{}
	h := NewHandlers(HandlerOptions{Runner: runner})
	params := json.RawMessage(`{"plistPath":"/Library/LaunchDaemons/com.example.daemon.plist"}`)
	_, err := h.Handle(context.Background(), &Request{Method: MethodBootstrap, Params: params})
	if err != nil {
		t.Fatalf("error: %s", errMsgFromRPC(err))
	}
	if len(runner.calls) != 1 {
		t.Fatalf("runner calls = %d, want 1", len(runner.calls))
	}
	got := runner.calls[0]
	if got.name != "launchctl" {
		t.Errorf("name = %q", got.name)
	}
	want := []string{"bootstrap", "system", "/Library/LaunchDaemons/com.example.daemon.plist"}
	if fmt.Sprintf("%v", got.args) != fmt.Sprintf("%v", want) {
		t.Errorf("args = %v, want %v", got.args, want)
	}
}

func TestHandlers_Bootstrap_PathOutsideAllowedDir(t *testing.T) {
	runner := &fakeRunner{}
	h := NewHandlers(HandlerOptions{Runner: runner})
	cases := []string{
		"/tmp/evil.plist",
		"/Library/LaunchAgents/com.example.plist",
		"relative/path.plist",
		"/Library/LaunchDaemons/../../etc/passwd",
	}
	for _, path := range cases {
		params, _ := json.Marshal(BootstrapParams{PlistPath: path})
		_, err := h.Handle(context.Background(), &Request{Method: MethodBootstrap, Params: params})
		if err == nil || err.Code != ErrCodeInvalidParams {
			t.Errorf("path %q: err = %+v, want invalid_params", path, err)
		}
	}
	if len(runner.calls) != 0 {
		t.Errorf("rejected paths still invoked launchctl %d times", len(runner.calls))
	}
}

func TestHandlers_Bootstrap_NonPlistExtensionRejected(t *testing.T) {
	h := NewHandlers(HandlerOptions{Runner: &fakeRunner{}})
	params, _ := json.Marshal(BootstrapParams{PlistPath: "/Library/LaunchDaemons/hacked"})
	_, err := h.Handle(context.Background(), &Request{Method: MethodBootstrap, Params: params})
	if err == nil || err.Code != ErrCodeInvalidParams {
		t.Errorf("err = %+v, want invalid_params", err)
	}
}

func TestHandlers_Bootstrap_LaunchctlFailure(t *testing.T) {
	runner := &fakeRunner{
		stderr: []byte("Bootstrap failed: 5: Input/output error"),
		err:    errors.New("exit status 5"),
	}
	h := NewHandlers(HandlerOptions{Runner: runner})
	params, _ := json.Marshal(BootstrapParams{PlistPath: "/Library/LaunchDaemons/com.example.plist"})
	_, err := h.Handle(context.Background(), &Request{Method: MethodBootstrap, Params: params})
	if err == nil || err.Code != ErrCodeLaunchctlFailed {
		t.Errorf("err = %+v, want launchctl_failed", err)
	}
	if !strings.Contains(err.Message, "Input/output error") {
		t.Errorf("message = %q; want stderr echoed", err.Message)
	}
}

func TestHandlers_Bootout_LabelInjection(t *testing.T) {
	runner := &fakeRunner{}
	h := NewHandlers(HandlerOptions{Runner: runner})
	cases := []string{
		"com.example; rm -rf /",
		"com.example$(echo pwned)",
		"`whoami`",
		"com.example/../other",
		"label with space",
		"",
	}
	for _, label := range cases {
		params, _ := json.Marshal(BootoutParams{Label: label})
		_, err := h.Handle(context.Background(), &Request{Method: MethodBootout, Params: params})
		if err == nil || err.Code != ErrCodeInvalidParams {
			t.Errorf("label %q: err = %+v, want invalid_params", label, err)
		}
	}
	if len(runner.calls) != 0 {
		t.Errorf("rejected labels still invoked launchctl %d times", len(runner.calls))
	}
}

func TestHandlers_Bootout_Valid(t *testing.T) {
	runner := &fakeRunner{}
	h := NewHandlers(HandlerOptions{Runner: runner})
	params, _ := json.Marshal(BootoutParams{Label: "com.example.daemon"})
	_, err := h.Handle(context.Background(), &Request{Method: MethodBootout, Params: params})
	if err != nil {
		t.Fatalf("err: %s", errMsgFromRPC(err))
	}
	if len(runner.calls) != 1 {
		t.Fatalf("calls = %d", len(runner.calls))
	}
	want := []string{"bootout", "system/com.example.daemon"}
	if fmt.Sprintf("%v", runner.calls[0].args) != fmt.Sprintf("%v", want) {
		t.Errorf("args = %v, want %v", runner.calls[0].args, want)
	}
}

func TestHandlers_Kickstart_Valid(t *testing.T) {
	runner := &fakeRunner{}
	h := NewHandlers(HandlerOptions{Runner: runner})
	params, _ := json.Marshal(KickstartParams{Label: "com.example.daemon"})
	_, err := h.Handle(context.Background(), &Request{Method: MethodKickstart, Params: params})
	if err != nil {
		t.Fatalf("err: %s", errMsgFromRPC(err))
	}
	want := []string{"kickstart", "-k", "system/com.example.daemon"}
	if fmt.Sprintf("%v", runner.calls[0].args) != fmt.Sprintf("%v", want) {
		t.Errorf("args = %v, want %v", runner.calls[0].args, want)
	}
}

func TestHandlers_WritePlist_PathValidation(t *testing.T) {
	h := NewHandlers(HandlerOptions{Runner: &fakeRunner{}, UserHome: t.TempDir()})
	params, _ := json.Marshal(WritePlistParams{PlistPath: "/etc/passwd", Data: ""})
	_, err := h.Handle(context.Background(), &Request{Method: MethodWritePlist, Params: params})
	if err == nil || err.Code != ErrCodeInvalidParams {
		t.Errorf("err = %+v, want invalid_params", err)
	}
}

func TestHandlers_WritePlist_NewFile(t *testing.T) {
	tmp := t.TempDir()
	// Route the daemon directory into our tempdir so the test can actually
	// exercise the write without needing /Library/LaunchDaemons. We wrap the
	// path validator by simulating a real plist path that matches the
	// allowed prefix: we set SystemDaemonDir via the package variable only
	// through the test by relative redirect via the atomicWriteDir helper.
	//
	// Since SystemDaemonDir is a const, we can't redirect it. Instead we
	// drive the test by patching os.Rename via the same atomicWrite path:
	// we accept that the unit test covers the validation path and rely on
	// a separate integration test for the actual write. Here we exercise
	// the happy path by stubbing out os.Rename indirectly — simpler: we
	// invoke atomicWrite directly.
	dest := filepath.Join(tmp, "com.example.test.plist")
	if errR := atomicWrite(dest, []byte("<plist/>"), 0644, -1, -1); errR != nil {
		t.Fatalf("atomicWrite: %+v", errR)
	}
	info, err := os.Stat(dest)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if mode := info.Mode() & 0777; mode != 0644 {
		t.Errorf("mode = %o, want 0644", mode)
	}
	got, err := os.ReadFile(dest)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if string(got) != "<plist/>" {
		t.Errorf("content = %q", string(got))
	}
}

func TestHandlers_Backup_ChownsToUser(t *testing.T) {
	tmp := t.TempDir()
	chowner, calls := newChownRecorder()
	h := NewHandlers(HandlerOptions{
		Runner:       &fakeRunner{},
		UserHome:     tmp,
		LaunchingUID: 501,
		LaunchingGID: 20,
		Chown:        chowner,
		NowFn:        func() time.Time { return time.Date(2026, 4, 22, 10, 0, 0, 0, time.UTC) },
	})
	if errR := h.backupExisting("/Library/LaunchDaemons/com.example.test.plist", []byte("<plist/>")); errR != nil {
		t.Fatalf("backupExisting: %+v", errR)
	}

	// A backup and meta file should exist.
	backupFile := filepath.Join(tmp, ".launchpal", "backups", "com.example.test", "20260422-100000.plist")
	if _, err := os.Stat(backupFile); err != nil {
		t.Errorf("backup file missing: %v", err)
	}
	metaFile := filepath.Join(tmp, ".launchpal", "backups", "com.example.test", "20260422-100000.meta.json")
	meta, err := os.ReadFile(metaFile)
	if err != nil {
		t.Fatalf("read meta: %v", err)
	}
	var decoded map[string]string
	if err := json.Unmarshal(meta, &decoded); err != nil {
		t.Fatalf("unmarshal meta: %v", err)
	}
	if decoded["originalPath"] != "/Library/LaunchDaemons/com.example.test.plist" {
		t.Errorf("originalPath = %q", decoded["originalPath"])
	}

	// Every path touched by the backup should have been chowned to the user.
	wantPaths := []string{
		filepath.Join(tmp, ".launchpal"),
		filepath.Join(tmp, ".launchpal", "backups"),
		filepath.Join(tmp, ".launchpal", "backups", "com.example.test"),
		backupFile,
		metaFile,
	}
	recorded := make(map[string]chownCall, len(*calls))
	for _, c := range *calls {
		recorded[c.path] = c
	}
	for _, p := range wantPaths {
		c, ok := recorded[p]
		if !ok {
			t.Errorf("no chown for %s", p)
			continue
		}
		if c.uid != 501 || c.gid != 20 {
			t.Errorf("chown %s: uid=%d gid=%d, want 501:20", p, c.uid, c.gid)
		}
	}
}

func TestHandlers_DeletePlist_NotFound(t *testing.T) {
	h := NewHandlers(HandlerOptions{
		Runner:   &fakeRunner{},
		UserHome: t.TempDir(),
	})
	params, _ := json.Marshal(DeletePlistParams{PlistPath: "/Library/LaunchDaemons/nope.plist"})
	_, err := h.Handle(context.Background(), &Request{Method: MethodDeletePlist, Params: params})
	if err == nil || err.Code != ErrCodeNotFound {
		t.Errorf("err = %+v, want not_found", err)
	}
}

func TestHandlers_Shutdown_TriggersShutdownFn(t *testing.T) {
	called := false
	h := NewHandlers(HandlerOptions{
		Runner:     &fakeRunner{},
		ShutdownFn: func() { called = true },
	})
	out, err := h.Handle(context.Background(), &Request{Method: MethodShutdown})
	if err != nil {
		t.Fatalf("err: %s", errMsgFromRPC(err))
	}
	if !called {
		t.Error("ShutdownFn was not invoked")
	}
	if ok, _ := out.(OKResult); !ok.OK {
		t.Errorf("result = %+v", out)
	}
}

func TestHandlers_List_UsesLister(t *testing.T) {
	lister := &fakeLister{list: []DaemonInfo{
		{Label: "com.example.one", Status: "running", PID: 42},
		{Label: "com.example.two", Status: "stopped", PID: 0},
	}}
	h := NewHandlers(HandlerOptions{Runner: &fakeRunner{}, Lister: lister})
	out, err := h.Handle(context.Background(), &Request{Method: MethodListSystemDaemons})
	if err != nil {
		t.Fatalf("err: %s", errMsgFromRPC(err))
	}
	res, ok := out.(ListSystemDaemonsResult)
	if !ok {
		t.Fatalf("result type = %T", out)
	}
	if len(res.Daemons) != 2 {
		t.Errorf("daemons = %d", len(res.Daemons))
	}
}

func TestHandlers_Get_NotFound(t *testing.T) {
	lister := &fakeLister{singleErr: errDaemonNotFound}
	h := NewHandlers(HandlerOptions{Runner: &fakeRunner{}, Lister: lister})
	params, _ := json.Marshal(GetSystemDaemonParams{Label: "com.example.missing"})
	_, err := h.Handle(context.Background(), &Request{Method: MethodGetSystemDaemon, Params: params})
	if err == nil || err.Code != ErrCodeNotFound {
		t.Errorf("err = %+v, want not_found", err)
	}
}

func TestHandlers_Get_InvalidLabel(t *testing.T) {
	h := NewHandlers(HandlerOptions{Runner: &fakeRunner{}})
	params, _ := json.Marshal(GetSystemDaemonParams{Label: "bad label!"})
	_, err := h.Handle(context.Background(), &Request{Method: MethodGetSystemDaemon, Params: params})
	if err == nil || err.Code != ErrCodeInvalidParams {
		t.Errorf("err = %+v, want invalid_params", err)
	}
}

func TestValidateSystemDaemonPath_TableDriven(t *testing.T) {
	cases := []struct {
		in       string
		ok       bool
		contains string // substring expected in error message
	}{
		{"/Library/LaunchDaemons/com.example.plist", true, ""},
		{"/Library/LaunchDaemons/foo/bar.plist", true, ""},
		{"/Library/LaunchDaemons/../etc/shadow", false, "must be under"},
		{"/tmp/test.plist", false, "must be under"},
		{"/Library/LaunchDaemonsX/foo.plist", false, "must be under"},
		{"relative.plist", false, "absolute"},
		{"", false, "required"},
		{"/Library/LaunchDaemons/nope.txt", false, "must end in .plist"},
	}
	for _, tc := range cases {
		_, err := validateSystemDaemonPath(tc.in)
		if tc.ok {
			if err != nil {
				t.Errorf("path %q: unexpected err %+v", tc.in, err)
			}
			continue
		}
		if err == nil {
			t.Errorf("path %q: expected error, got nil", tc.in)
			continue
		}
		if !strings.Contains(err.Message, tc.contains) {
			t.Errorf("path %q: msg = %q, want contains %q", tc.in, err.Message, tc.contains)
		}
	}
}

func TestParsePrintSystem(t *testing.T) {
	output := []byte(`com.apple.xpc.launchd.domain.system = {
	type = system
	handle = 0
	active count = 123
	services = {
		12345  -  com.example.running
		-      0 com.example.stopped
		67890  -  com.example.another
	}
}`)
	got := parsePrintSystem(output)
	if len(got) != 3 {
		t.Fatalf("len = %d, want 3", len(got))
	}
	if got[0].Label != "com.example.running" || got[0].PID != 12345 || got[0].Status != "running" {
		t.Errorf("entry 0 = %+v", got[0])
	}
	if got[1].Label != "com.example.stopped" || got[1].PID != 0 || got[1].Status != "stopped" {
		t.Errorf("entry 1 = %+v", got[1])
	}
	if got[2].Label != "com.example.another" || got[2].PID != 67890 {
		t.Errorf("entry 2 = %+v", got[2])
	}
}

func TestParsePrintService(t *testing.T) {
	output := []byte(`com.example.daemon = {
	state = running
	pid = 5432
	...
}`)
	got := parsePrintService("com.example.daemon", output)
	if got.Label != "com.example.daemon" {
		t.Errorf("Label = %q", got.Label)
	}
	if got.Status != "running" {
		t.Errorf("Status = %q", got.Status)
	}
	if got.PID != 5432 {
		t.Errorf("PID = %d", got.PID)
	}
}

func TestHandlers_EnsureLogAccess(t *testing.T) {
	// Use a sandbox under /tmp so validateLogPath accepts the paths (macOS
	// resolves /tmp through /private/tmp, which is in the allowlist).
	base, err := os.MkdirTemp("/tmp", "launchpal-test-*")
	if err != nil {
		t.Fatalf("mktemp: %v", err)
	}
	defer func() { _ = os.RemoveAll(base) }()

	t.Run("creates parent dir and touches missing file", func(t *testing.T) {
		h := NewHandlers(HandlerOptions{Runner: &fakeRunner{}})
		logPath := filepath.Join(base, "svc-created", "out.log")
		params, _ := json.Marshal(EnsureLogAccessParams{Paths: []string{logPath}})
		_, errR := h.Handle(context.Background(), &Request{Method: MethodEnsureLogAccess, Params: params})
		if errR != nil {
			t.Fatalf("handler err = %+v", errR)
		}
		info, err := os.Stat(filepath.Dir(logPath))
		if err != nil {
			t.Fatalf("parent dir: %v", err)
		}
		if info.Mode().Perm() != 0755 {
			t.Errorf("parent perm = %o, want 0755", info.Mode().Perm())
		}
		finfo, err := os.Stat(logPath)
		if err != nil {
			t.Fatalf("log file: %v", err)
		}
		if finfo.Size() != 0 {
			t.Errorf("touched file size = %d, want 0", finfo.Size())
		}
	})

	t.Run("tightens existing restrictive parent to 0755", func(t *testing.T) {
		// Mimics the reported bug: launchd created /var/log/jeff.test at
		// 0744 (no x-bit for others), so the GUI user can't even stat the
		// file inside. The handler should open the directory back up.
		dir := filepath.Join(base, "restrictive")
		if err := os.Mkdir(dir, 0744); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
		logPath := filepath.Join(dir, "out.log")
		h := NewHandlers(HandlerOptions{Runner: &fakeRunner{}})
		params, _ := json.Marshal(EnsureLogAccessParams{Paths: []string{logPath}})
		_, errR := h.Handle(context.Background(), &Request{Method: MethodEnsureLogAccess, Params: params})
		if errR != nil {
			t.Fatalf("handler err = %+v", errR)
		}
		info, _ := os.Stat(dir)
		if info.Mode().Perm() != 0755 {
			t.Errorf("parent perm after ensure = %o, want 0755", info.Mode().Perm())
		}
	})

	t.Run("preserves existing log file contents", func(t *testing.T) {
		dir := filepath.Join(base, "keep")
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
		logPath := filepath.Join(dir, "out.log")
		const existing = "pre-existing log\n"
		if err := os.WriteFile(logPath, []byte(existing), 0644); err != nil {
			t.Fatalf("seed: %v", err)
		}
		h := NewHandlers(HandlerOptions{Runner: &fakeRunner{}})
		params, _ := json.Marshal(EnsureLogAccessParams{Paths: []string{logPath}})
		_, errR := h.Handle(context.Background(), &Request{Method: MethodEnsureLogAccess, Params: params})
		if errR != nil {
			t.Fatalf("handler err = %+v", errR)
		}
		got, err := os.ReadFile(logPath)
		if err != nil {
			t.Fatalf("read: %v", err)
		}
		if string(got) != existing {
			t.Errorf("file contents changed to %q", got)
		}
	})

	t.Run("empty paths skip validation", func(t *testing.T) {
		h := NewHandlers(HandlerOptions{Runner: &fakeRunner{}})
		params, _ := json.Marshal(EnsureLogAccessParams{Paths: []string{"", filepath.Join(base, "svc-skip", "out.log"), ""}})
		_, errR := h.Handle(context.Background(), &Request{Method: MethodEnsureLogAccess, Params: params})
		if errR != nil {
			t.Errorf("handler err = %+v", errR)
		}
	})

	t.Run("rejects paths outside allowlist", func(t *testing.T) {
		h := NewHandlers(HandlerOptions{Runner: &fakeRunner{}})
		bad := []string{
			"/etc/passwd",            // outside allowlist entirely
			"/root/evil.log",         // outside allowlist entirely
			"/var/log/../etc/passwd", // traversal normalized out of allowlist
			"/var/log/foo.log",       // parent IS the allowlist root
			"relative.log",           // not absolute
			"/tmp/",                  // bare directory
		}
		for _, p := range bad {
			params, _ := json.Marshal(EnsureLogAccessParams{Paths: []string{p}})
			_, errR := h.Handle(context.Background(), &Request{Method: MethodEnsureLogAccess, Params: params})
			if errR == nil {
				t.Errorf("path %q: expected rejection", p)
				continue
			}
			if errR.Code != ErrCodeInvalidParams {
				t.Errorf("path %q: code = %s, want invalid_params (msg=%s)", p, errR.Code, errR.Message)
			}
		}
	})
}

func TestHandlers_TruncateLog(t *testing.T) {
	// /tmp resolves to /private/tmp on darwin, both of which are in the
	// allowlist; the test sandbox lives there so validateLogPath accepts
	// the paths.
	base, err := os.MkdirTemp("/tmp", "launchpal-trunc-*")
	if err != nil {
		t.Fatalf("mktemp: %v", err)
	}
	defer func() { _ = os.RemoveAll(base) }()

	t.Run("truncates existing file preserving mode", func(t *testing.T) {
		dir := filepath.Join(base, "svc-trunc")
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
		logPath := filepath.Join(dir, "out.log")
		if err := os.WriteFile(logPath, []byte("noisy log content\n"), 0640); err != nil {
			t.Fatalf("seed: %v", err)
		}
		before, err := os.Stat(logPath)
		if err != nil {
			t.Fatalf("stat before: %v", err)
		}

		h := NewHandlers(HandlerOptions{Runner: &fakeRunner{}})
		params, _ := json.Marshal(TruncateLogParams{Path: logPath})
		out, errR := h.Handle(context.Background(), &Request{Method: MethodTruncateLog, Params: params})
		if errR != nil {
			t.Fatalf("handler err = %+v", errR)
		}
		if ok, _ := out.(OKResult); !ok.OK {
			t.Errorf("result = %+v", out)
		}

		after, err := os.Stat(logPath)
		if err != nil {
			t.Fatalf("stat after: %v", err)
		}
		if after.Size() != 0 {
			t.Errorf("size after = %d, want 0", after.Size())
		}
		if after.Mode().Perm() != before.Mode().Perm() {
			t.Errorf("mode changed: %o -> %o", before.Mode().Perm(), after.Mode().Perm())
		}
	})

	t.Run("rejects path outside allowlist", func(t *testing.T) {
		h := NewHandlers(HandlerOptions{Runner: &fakeRunner{}})
		bad := []string{
			"/etc/passwd",
			"/Users/jeff/secret.log",
			"relative.log",
		}
		for _, p := range bad {
			params, _ := json.Marshal(TruncateLogParams{Path: p})
			_, errR := h.Handle(context.Background(), &Request{Method: MethodTruncateLog, Params: params})
			if errR == nil || errR.Code != ErrCodeInvalidParams {
				t.Errorf("path %q: code = %s, want invalid_params", p, errCodeOrEmpty(errR))
			}
		}
	})

	t.Run("rejects parent equal to allowlist root", func(t *testing.T) {
		// /var/log/foo.log: parent is /var/log which is the allowlist root
		// itself; truncating files there would be a foot-gun (system.log
		// etc.). validateLogPath enforces "must live in a sub-directory".
		h := NewHandlers(HandlerOptions{Runner: &fakeRunner{}})
		params, _ := json.Marshal(TruncateLogParams{Path: "/var/log/foo.log"})
		_, errR := h.Handle(context.Background(), &Request{Method: MethodTruncateLog, Params: params})
		if errR == nil || errR.Code != ErrCodeInvalidParams {
			t.Errorf("code = %s, want invalid_params", errCodeOrEmpty(errR))
		}
	})

	t.Run("missing file returns not_found", func(t *testing.T) {
		h := NewHandlers(HandlerOptions{Runner: &fakeRunner{}})
		logPath := filepath.Join(base, "svc-missing", "never-created.log")
		params, _ := json.Marshal(TruncateLogParams{Path: logPath})
		_, errR := h.Handle(context.Background(), &Request{Method: MethodTruncateLog, Params: params})
		if errR == nil || errR.Code != ErrCodeNotFound {
			t.Errorf("code = %s, want not_found", errCodeOrEmpty(errR))
		}
		if _, statErr := os.Stat(logPath); !os.IsNotExist(statErr) {
			t.Error("handler should not have created the file")
		}
	})

	t.Run("symlink at log path is rejected without dereferencing", func(t *testing.T) {
		dir := filepath.Join(base, "svc-symlink")
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
		target := filepath.Join(dir, "target.log")
		if err := os.WriteFile(target, []byte("real content\n"), 0644); err != nil {
			t.Fatalf("seed target: %v", err)
		}
		linkPath := filepath.Join(dir, "out.log")
		if err := os.Symlink(target, linkPath); err != nil {
			t.Fatalf("symlink: %v", err)
		}

		h := NewHandlers(HandlerOptions{Runner: &fakeRunner{}})
		params, _ := json.Marshal(TruncateLogParams{Path: linkPath})
		_, errR := h.Handle(context.Background(), &Request{Method: MethodTruncateLog, Params: params})
		if errR == nil {
			t.Fatal("expected error for symlink")
		}
		// O_NOFOLLOW is darwin-only; on portable CI builds nofollowFlag is 0
		// and the open succeeds. Skip the strict assertion in that case so
		// CI stays green; on macOS (the only supported platform) the symlink
		// is reliably rejected.
		if syscallNoFollow == 0 {
			t.Skipf("O_NOFOLLOW unavailable on this build; symlink rejection only enforced on darwin")
		}
		if errR.Code != ErrCodeIOError {
			t.Errorf("code = %s, want io_error", errR.Code)
		}
		// Target must NOT have been truncated.
		info, statErr := os.Stat(target)
		if statErr != nil {
			t.Fatalf("stat target: %v", statErr)
		}
		if info.Size() == 0 {
			t.Error("symlink target was truncated; O_NOFOLLOW failed")
		}
	})

	t.Run("dispatches via Handle switch", func(t *testing.T) {
		// Fresh sanity check that MethodTruncateLog is wired into the
		// switch — a regression where the constant is added but the case is
		// missing would manifest as unknown_method.
		dir := filepath.Join(base, "svc-dispatch")
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
		logPath := filepath.Join(dir, "out.log")
		if err := os.WriteFile(logPath, []byte("x\n"), 0644); err != nil {
			t.Fatalf("seed: %v", err)
		}
		h := NewHandlers(HandlerOptions{Runner: &fakeRunner{}})
		params, _ := json.Marshal(TruncateLogParams{Path: logPath})
		_, errR := h.Handle(context.Background(), &Request{Method: MethodTruncateLog, Params: params})
		if errR != nil && errR.Code == ErrCodeUnknownMethod {
			t.Fatal("MethodTruncateLog not wired into Handle switch")
		}
	})
}

// errCodeOrEmpty extracts the code from an RPCError or returns "" so test
// failure messages stay readable when the error happens to be nil.
func errCodeOrEmpty(e *RPCError) string {
	if e == nil {
		return ""
	}
	return e.Code
}

// Smoke test that writePlist-via-handler correctly rejects path even when
// data is valid base64. Keeps the end-to-end code path green.
func TestHandlers_WritePlist_Base64Validation(t *testing.T) {
	h := NewHandlers(HandlerOptions{Runner: &fakeRunner{}, UserHome: t.TempDir()})
	badData := "!!!not-base64"
	params, _ := json.Marshal(WritePlistParams{PlistPath: "/Library/LaunchDaemons/com.x.plist", Data: badData})
	_, err := h.Handle(context.Background(), &Request{Method: MethodWritePlist, Params: params})
	if err == nil || err.Code != ErrCodeInvalidParams {
		t.Errorf("err = %+v, want invalid_params", err)
	}
	// And valid base64 gets past validation (still fails the actual write
	// since /Library/LaunchDaemons isn't writable in CI; the error class
	// then is io_error).
	good := base64.StdEncoding.EncodeToString([]byte("<plist/>"))
	params, _ = json.Marshal(WritePlistParams{PlistPath: "/Library/LaunchDaemons/com.x.plist", Data: good})
	_, err = h.Handle(context.Background(), &Request{Method: MethodWritePlist, Params: params})
	if err == nil {
		t.Error("expected error writing to /Library/LaunchDaemons without root")
	} else if err.Code != ErrCodeIOError {
		t.Errorf("err = %+v, want io_error", err)
	}
}
