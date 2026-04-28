package privhelper

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"os"
	"strings"
	"testing"
	"time"
)

// pairedClientServer sets up a Client backed by a net.Pipe-style connected
// pair so unit tests don't need a Unix socket. The "server" side is a
// goroutine that reads requests and sends canned responses via respond.
func pairedClientServer(t *testing.T, respond func(req Request) Response) (*Client, func()) {
	t.Helper()
	clientEnd, serverEnd := net.Pipe()
	done := make(chan struct{})
	go func() {
		defer close(done)
		defer func() { _ = serverEnd.Close() }()
		decoder := json.NewDecoder(serverEnd)
		encoder := json.NewEncoder(serverEnd)
		for {
			var req Request
			if err := decoder.Decode(&req); err != nil {
				return
			}
			resp := respond(req)
			if err := encoder.Encode(resp); err != nil {
				return
			}
		}
	}()

	client := NewClient(ClientOptions{Conn: clientEnd})
	cleanup := func() {
		_ = client.Close()
		_ = serverEnd.Close()
		<-done
	}
	return client, cleanup
}

func TestClient_CallResponseCorrelation(t *testing.T) {
	client, cleanup := pairedClientServer(t, func(req Request) Response {
		// Echo the id with a method-dependent result.
		id := req.ID
		result, _ := json.Marshal(map[string]any{"method": req.Method, "id": req.ID})
		return Response{ID: &id, Result: result}
	})
	defer cleanup()

	raw, err := client.Call(context.Background(), "A", nil)
	if err != nil {
		t.Fatalf("call A: %v", err)
	}
	var obj map[string]any
	if err := json.Unmarshal(raw, &obj); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if obj["method"] != "A" {
		t.Errorf("method = %v", obj["method"])
	}
}

func TestClient_ErrorResponse(t *testing.T) {
	client, cleanup := pairedClientServer(t, func(req Request) Response {
		id := req.ID
		return Response{ID: &id, Error: &RPCError{Code: ErrCodeInvalidParams, Message: "nope"}}
	})
	defer cleanup()

	_, err := client.Call(context.Background(), "Bad", nil)
	if err == nil {
		t.Fatal("expected error")
	}
	var rpcErr *RPCError
	if !errors.As(err, &rpcErr) {
		t.Fatalf("err type = %T", err)
	}
	if rpcErr.Code != ErrCodeInvalidParams {
		t.Errorf("code = %q", rpcErr.Code)
	}
}

func TestClient_CloseUnblocksCall(t *testing.T) {
	client, cleanup := pairedClientServer(t, func(req Request) Response {
		// Server never responds — simulate a hung helper.
		time.Sleep(5 * time.Second)
		id := req.ID
		return Response{ID: &id}
	})
	defer cleanup()

	errCh := make(chan error, 1)
	go func() {
		_, err := client.Call(context.Background(), "Slow", nil)
		errCh <- err
	}()

	time.Sleep(50 * time.Millisecond)
	_ = client.Close()

	select {
	case err := <-errCh:
		if !errors.Is(err, ErrClientClosed) {
			// I/O errors also acceptable since net.Pipe may report differently.
			if err == nil {
				t.Errorf("expected error on Close")
			}
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Call did not unblock after Close")
	}
}

func TestConnect_RetryAndTimeout(t *testing.T) {
	// Dial a non-existent socket; expect timeout.
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	path := "/tmp/does-not-exist-launchpal-test.sock"
	_ = os.Remove(path)
	_, err := Connect(ctx, path, 200*time.Millisecond)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "timeout") {
		t.Errorf("err = %v, want timeout", err)
	}
}

func TestConnect_SucceedsOnceListenerStarts(t *testing.T) {
	path := shortSocketPath(t)
	// Start a listener a bit late to force at least one retry.
	ready := make(chan struct{})
	go func() {
		time.Sleep(100 * time.Millisecond)
		l, err := net.ListenUnix("unix", &net.UnixAddr{Name: path, Net: "unix"})
		if err != nil {
			panic(err)
		}
		close(ready)
		conn, err := l.Accept()
		if err != nil {
			return
		}
		defer func() { _ = conn.Close() }()
		_ = l.Close()
	}()

	conn, err := Connect(context.Background(), path, 2*time.Second)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer func() { _ = conn.Close() }()
	<-ready
}

func TestIsAuthorizationCanceled(t *testing.T) {
	cases := map[string]bool{
		"execution error: User canceled. (-128)": true,
		"User cancelled.":                        true,
		"osascript: error (-128)":                true,
		"unexpected error doing shell script":    false,
		"":                                       false,
	}
	for input, want := range cases {
		got := isAuthorizationCanceled(errors.New(input))
		if got != want {
			t.Errorf("input %q: got %v, want %v", input, got, want)
		}
	}
	if isAuthorizationCanceled(nil) {
		t.Error("nil error should not be considered cancelled")
	}
}

func TestLaunchHelper_UserCancelsAuthorization(t *testing.T) {
	ran := false
	opts := LaunchHelperOptions{
		HelperPath:   "/fake/helper",
		SocketPath:   "/tmp/fake.sock",
		ParentPID:    1234,
		LaunchingUID: 501,
		osascriptRunner: func(_ context.Context, _ string) error {
			ran = true
			return errors.New("execution error: User canceled. (-128)")
		},
	}
	_, err := LaunchHelper(context.Background(), opts)
	if !errors.Is(err, ErrAuthorizationCanceled) {
		t.Errorf("err = %v, want ErrAuthorizationCanceled", err)
	}
	if !ran {
		t.Error("osascript runner should have been invoked")
	}
}

func TestLaunchHelper_OsascriptFailure(t *testing.T) {
	opts := LaunchHelperOptions{
		HelperPath:   "/fake/helper",
		SocketPath:   "/tmp/fake.sock",
		ParentPID:    1234,
		LaunchingUID: 501,
		osascriptRunner: func(_ context.Context, _ string) error {
			return errors.New("command not found: osascript")
		},
	}
	_, err := LaunchHelper(context.Background(), opts)
	if errors.Is(err, ErrAuthorizationCanceled) {
		t.Errorf("a generic error should not be treated as cancel: %v", err)
	}
	if err == nil {
		t.Error("expected an error")
	}
}

func TestGenerateSocketPath_Shape(t *testing.T) {
	uid := os.Getuid()
	p, err := GenerateSocketPath(uid)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	if !strings.HasSuffix(p, ".sock") {
		t.Errorf("suffix = %q", p)
	}
	if !strings.Contains(p, "launchpal-") {
		t.Errorf("missing launchpal prefix: %q", p)
	}
	// Two successive calls should produce different paths.
	q, err := GenerateSocketPath(uid)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	if p == q {
		t.Error("generated paths collided")
	}
}

func TestShellQuote_EscapesSingleQuote(t *testing.T) {
	got := shellQuote(`can't`)
	want := `'can'\''t'`
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestAppleScriptQuote_EscapesBackslashAndQuote(t *testing.T) {
	got := appleScriptQuote(`a "b" \c`)
	// expect: "a \"b\" \\c"
	want := `"a \"b\" \\c"`
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
