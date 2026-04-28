package privhelper

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"net"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// pingHandler is a minimal handler that implements only Ping; every other
// method is an unknown_method error. Used by server tests to focus on
// transport behavior (auth, serialization, shutdown) without handler logic.
func pingHandler(_ context.Context, req *Request) (any, *RPCError) {
	switch req.Method {
	case MethodPing:
		return PingResult{Pong: true}, nil
	default:
		return nil, &RPCError{Code: ErrCodeUnknownMethod, Message: req.Method}
	}
}

// stubPeerUID returns a fixed UID so tests can exercise the auth branch
// without running under multiple users.
func stubPeerUID(uid int) PeerUIDFunc {
	return func(_ *net.UnixConn) (int, error) {
		return uid, nil
	}
}

// shortSocketPath returns a short unique socket path under /tmp. macOS limits
// Unix sockets to ~104 bytes; Go's t.TempDir() usually exceeds that on darwin
// so tests that put the socket inside TempDir fail with EINVAL.
func shortSocketPath(t *testing.T) string {
	t.Helper()
	f, err := os.CreateTemp("/tmp", "lp-*.sock")
	if err != nil {
		t.Fatalf("tmp sock: %v", err)
	}
	path := f.Name()
	_ = f.Close()
	_ = os.Remove(path)
	t.Cleanup(func() { _ = os.Remove(path) })
	return path
}

func startTestServer(t *testing.T, opts ServerOptions) *Server {
	t.Helper()
	if opts.SocketPath == "" {
		opts.SocketPath = shortSocketPath(t)
	}
	if opts.Handler == nil {
		opts.Handler = pingHandler
	}
	if opts.PeerUID == nil {
		opts.PeerUID = stubPeerUID(opts.AuthorizedUID)
	}
	s := NewServer(opts)
	if err := s.Start(); err != nil {
		t.Fatalf("Start: %v", err)
	}
	t.Cleanup(func() { s.Stop(); _ = s.Wait() })
	return s
}

func TestServer_SocketCreatedWith0600(t *testing.T) {
	sockPath := shortSocketPath(t)
	s := startTestServer(t, ServerOptions{SocketPath: sockPath, AuthorizedUID: os.Getuid()})
	_ = s

	info, err := os.Stat(sockPath)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if mode := info.Mode() & 0777; mode != 0600 {
		t.Errorf("socket mode = %o, want 0600", mode)
	}
}

func TestServer_SocketRemovedOnStop(t *testing.T) {
	sockPath := shortSocketPath(t)
	s := startTestServer(t, ServerOptions{SocketPath: sockPath, AuthorizedUID: os.Getuid()})
	s.Stop()
	_ = s.Wait()

	if _, err := os.Stat(sockPath); !os.IsNotExist(err) {
		t.Errorf("socket still exists after Stop: err=%v", err)
	}
}

func TestServer_StaleSocketCleared(t *testing.T) {
	sockPath := shortSocketPath(t)
	// Leave a stale file behind to simulate a crashed predecessor.
	if err := os.WriteFile(sockPath, []byte{}, 0600); err != nil {
		t.Fatalf("seed: %v", err)
	}
	s := startTestServer(t, ServerOptions{SocketPath: sockPath, AuthorizedUID: os.Getuid()})
	_ = s
}

func TestServer_AcceptsMatchingPeer(t *testing.T) {
	s := startTestServer(t, ServerOptions{AuthorizedUID: 42, PeerUID: stubPeerUID(42)})

	conn := dial(t, s.opts.SocketPath)
	defer func() { _ = conn.Close() }()

	write(t, conn, Request{ID: 1, Method: MethodPing})

	resp := readResponse(t, conn, 2*time.Second)
	if resp.Error != nil {
		t.Fatalf("unexpected error: %+v", resp.Error)
	}
	if resp.ID == nil || *resp.ID != 1 {
		t.Errorf("ID = %v, want 1", resp.ID)
	}
	var ping PingResult
	if err := json.Unmarshal(resp.Result, &ping); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if !ping.Pong {
		t.Errorf("pong = false, want true")
	}
}

func TestServer_RejectsMismatchedPeer(t *testing.T) {
	var handlerCalls atomic.Int32
	s := startTestServer(t, ServerOptions{
		AuthorizedUID: 42,
		PeerUID:       stubPeerUID(99),
		Handler: func(_ context.Context, _ *Request) (any, *RPCError) {
			handlerCalls.Add(1)
			return PingResult{Pong: true}, nil
		},
	})

	conn, err := net.DialUnix("unix", nil, &net.UnixAddr{Name: s.opts.SocketPath, Net: "unix"})
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer func() { _ = conn.Close() }()

	// The server closes without reading; writing may succeed into the pipe
	// but the read side should EOF promptly.
	_, _ = conn.Write([]byte(`{"id":1,"method":"Ping"}` + "\n"))
	_ = conn.SetReadDeadline(time.Now().Add(1 * time.Second))
	buf := make([]byte, 16)
	n, err := conn.Read(buf)
	if n > 0 || err == nil {
		t.Errorf("expected EOF/error, got n=%d err=%v", n, err)
	}
	// The handler should never have seen the request.
	if handlerCalls.Load() != 0 {
		t.Errorf("handler called %d times, want 0", handlerCalls.Load())
	}
}

func TestServer_SerialRequestProcessing(t *testing.T) {
	type order struct{ id int64 }
	var seen []order
	var mu atomicSlice[order]

	handler := func(_ context.Context, req *Request) (any, *RPCError) {
		// Use a tiny sleep to make interleaving visible in a racy impl.
		time.Sleep(5 * time.Millisecond)
		mu.Append(order{id: req.ID})
		return OKResult{OK: true}, nil
	}

	s := startTestServer(t, ServerOptions{
		AuthorizedUID: os.Getuid(),
		Handler:       handler,
	})

	conn := dial(t, s.opts.SocketPath)
	defer func() { _ = conn.Close() }()

	// Pipeline 3 requests back-to-back.
	for _, id := range []int64{1, 2, 3} {
		write(t, conn, Request{ID: id, Method: MethodPing})
	}

	reader := bufio.NewReader(conn)
	for i := 0; i < 3; i++ {
		_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
		line, err := reader.ReadBytes('\n')
		if err != nil {
			t.Fatalf("read resp %d: %v", i, err)
		}
		_ = line
	}
	seen = mu.Snapshot()

	if len(seen) != 3 {
		t.Fatalf("saw %d requests, want 3", len(seen))
	}
	for i, s := range seen {
		if s.id != int64(i+1) {
			t.Errorf("request %d id=%d, want %d", i, s.id, i+1)
		}
	}
}

func TestServer_MalformedJSON(t *testing.T) {
	s := startTestServer(t, ServerOptions{AuthorizedUID: os.Getuid()})

	conn := dial(t, s.opts.SocketPath)
	defer func() { _ = conn.Close() }()

	if _, err := conn.Write([]byte("{not json\n")); err != nil {
		t.Fatalf("write: %v", err)
	}
	resp := readResponse(t, conn, 2*time.Second)
	if resp.Error == nil || resp.Error.Code != ErrCodeInvalidRequest {
		t.Errorf("error = %+v, want invalid_request", resp.Error)
	}
	if resp.ID != nil {
		t.Errorf("id should be null for malformed request, got %v", *resp.ID)
	}
}

func TestServer_IdleTimeout(t *testing.T) {
	var idleFired atomic.Bool
	sockPath := shortSocketPath(t)
	s := NewServer(ServerOptions{
		SocketPath:    sockPath,
		AuthorizedUID: os.Getuid(),
		Handler:       pingHandler,
		PeerUID:       stubPeerUID(os.Getuid()),
		IdleTimeout:   200 * time.Millisecond,
		OnIdle:        func() { idleFired.Store(true) },
	})
	if err := s.Start(); err != nil {
		t.Fatalf("Start: %v", err)
	}

	// Wait for the idle watchdog to fire.
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if idleFired.Load() {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	if !idleFired.Load() {
		t.Fatal("OnIdle did not fire within deadline")
	}
	_ = s.Wait()
	if _, err := os.Stat(sockPath); !os.IsNotExist(err) {
		t.Errorf("socket should be removed, err=%v", err)
	}
}

func TestServer_RequestShutdown(t *testing.T) {
	sockPath := shortSocketPath(t)
	var s *Server
	handler := func(_ context.Context, req *Request) (any, *RPCError) {
		if req.Method == MethodShutdown {
			s.RequestShutdown()
			return OKResult{OK: true}, nil
		}
		return PingResult{Pong: true}, nil
	}
	s = NewServer(ServerOptions{
		SocketPath:    sockPath,
		AuthorizedUID: os.Getuid(),
		Handler:       handler,
		PeerUID:       stubPeerUID(os.Getuid()),
	})
	if err := s.Start(); err != nil {
		t.Fatalf("Start: %v", err)
	}

	conn := dial(t, sockPath)
	write(t, conn, Request{ID: 1, Method: MethodShutdown})
	resp := readResponse(t, conn, 2*time.Second)
	if resp.Error != nil {
		t.Fatalf("unexpected error: %+v", resp.Error)
	}
	_ = conn.Close()

	// Wait should return promptly after the shutdown.
	done := make(chan error, 1)
	go func() { done <- s.Wait() }()
	select {
	case err := <-done:
		if err != nil {
			t.Errorf("Wait: %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("Wait did not return after shutdown")
	}
	if _, err := os.Stat(sockPath); !os.IsNotExist(err) {
		t.Errorf("socket should be gone, err=%v", err)
	}
}

func TestServer_ParentWatchdog(t *testing.T) {
	sockPath := shortSocketPath(t)
	s := NewServer(ServerOptions{
		SocketPath:    sockPath,
		AuthorizedUID: os.Getuid(),
		Handler:       pingHandler,
		PeerUID:       stubPeerUID(os.Getuid()),
	})
	if err := s.Start(); err != nil {
		t.Fatalf("Start: %v", err)
	}

	var exitCalled atomic.Bool
	dead := make(chan struct{})
	aliveCount := 0
	cancel := s.StartParentWatchdog(99999, 50*time.Millisecond,
		func(_ int) bool {
			aliveCount++
			if aliveCount == 1 {
				return true
			}
			close(dead)
			return false
		},
		func() { exitCalled.Store(true) },
	)
	defer cancel()

	select {
	case <-dead:
	case <-time.After(2 * time.Second):
		t.Fatal("watchdog did not detect dead parent")
	}
	_ = s.Wait()
	if !exitCalled.Load() {
		t.Error("onExit should have been called")
	}
}

// atomicSlice is a tiny thread-safe append helper for concurrent test
// assertions on ordered operations.
type atomicSlice[T any] struct {
	mu    sync.Mutex
	items []T
}

func (s *atomicSlice[T]) Append(v T) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items = append(s.items, v)
}

func (s *atomicSlice[T]) Snapshot() []T {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]T, len(s.items))
	copy(out, s.items)
	return out
}

// dial connects to a Unix socket and fails the test on error.
func dial(t *testing.T, path string) *net.UnixConn {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for {
		conn, err := net.DialUnix("unix", nil, &net.UnixAddr{Name: path, Net: "unix"})
		if err == nil {
			return conn
		}
		if time.Now().After(deadline) {
			t.Fatalf("dial %s: %v", path, err)
		}
		time.Sleep(20 * time.Millisecond)
	}
}

// write encodes req as a JSON line and sends it to the connection.
func write(t *testing.T, conn *net.UnixConn, req Request) {
	t.Helper()
	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	data = append(data, '\n')
	if _, err := conn.Write(data); err != nil {
		t.Fatalf("write: %v", err)
	}
}

// readResponse reads a single JSON-line response with a deadline.
func readResponse(t *testing.T, conn *net.UnixConn, deadline time.Duration) Response {
	t.Helper()
	_ = conn.SetReadDeadline(time.Now().Add(deadline))
	reader := bufio.NewReader(conn)
	line, err := reader.ReadBytes('\n')
	if err != nil && !errors.Is(err, net.ErrClosed) {
		// A partial read that ended in EOF still may have a valid line.
		if len(line) == 0 {
			t.Fatalf("read: %v", err)
		}
	}
	var resp Response
	if err := json.Unmarshal(line, &resp); err != nil {
		t.Fatalf("unmarshal %q: %v", string(line), err)
	}
	return resp
}
