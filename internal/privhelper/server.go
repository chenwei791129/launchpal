package privhelper

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

// Handler processes a single decoded RPC request. Implementations return a
// result payload that will be marshaled into Response.Result, or an RPCError.
// Returning a non-RPCError error is treated as ErrCodeInternalError.
type Handler func(ctx context.Context, req *Request) (any, *RPCError)

// PeerUIDFunc resolves the peer UID for an accepted Unix socket connection.
// Injected so tests can stub out the LOCAL_PEERCRED syscall.
type PeerUIDFunc func(c *net.UnixConn) (int, error)

// ServerOptions configure a Server instance.
type ServerOptions struct {
	// SocketPath is the filesystem path of the Unix socket to create. The
	// server removes any existing file at that path on Start, chmods the
	// newly-created socket to 0600, and removes the file on Stop.
	SocketPath string

	// AuthorizedUID is the UID the connecting peer must match. Connections
	// from any other UID are closed without reading any input.
	AuthorizedUID int

	// SocketOwnerUID, SocketOwnerGID optionally chown the created socket
	// to the launching user. A root helper creating a socket without this
	// lands at root:wheel mode 0600, which the launching user cannot
	// connect to. When zero (or negative) the chown is skipped — intended
	// for tests where the helper runs as the same user as the client.
	SocketOwnerUID int
	SocketOwnerGID int

	// Handler dispatches decoded requests to their implementations.
	Handler Handler

	// PeerUID resolves the peer UID of an accepted connection. Defaults to
	// LOCAL_PEERCRED-based resolution on darwin.
	PeerUID PeerUIDFunc

	// IdleTimeout is the duration of no RPC traffic after which the server
	// stops itself. Zero disables idle timeout.
	IdleTimeout time.Duration

	// OnIdle is invoked when IdleTimeout elapses; the server calls Stop
	// after the callback. Primarily for tests.
	OnIdle func()

	// Now returns the current time; injectable for tests.
	Now func() time.Time
}

// Server listens on a Unix domain socket, authenticates peers, and dispatches
// newline-delimited JSON requests to the configured handler.
type Server struct {
	opts     ServerOptions
	listener *net.UnixListener

	mu            sync.Mutex
	stopped       bool
	stopErr       error
	conns         map[net.Conn]struct{} // accepted connections still being served
	primaryConn   net.Conn              // first accepted connection: the session client
	activityAtNs  atomic.Int64          // unix nanos of the last RPC event
	idleCancel    context.CancelFunc
	stopSignal    chan struct{}
	connWG        sync.WaitGroup
	shutdownCalls chan struct{}
}

// NewServer constructs a Server. It does not open the socket; call Start.
func NewServer(opts ServerOptions) *Server {
	if opts.Now == nil {
		opts.Now = time.Now
	}
	if opts.PeerUID == nil {
		opts.PeerUID = defaultPeerUID
	}
	s := &Server{
		opts:          opts,
		conns:         make(map[net.Conn]struct{}),
		stopSignal:    make(chan struct{}),
		shutdownCalls: make(chan struct{}, 1),
	}
	s.touch()
	return s
}

// touch records the current time as the last RPC activity, used by the idle
// watchdog to decide when the helper has been dormant long enough to exit.
func (s *Server) touch() {
	s.activityAtNs.Store(s.opts.Now().UnixNano())
}

// Start listens on the configured socket path, chmods it to 0600, and begins
// accepting connections in background goroutines.
func (s *Server) Start() error {
	// Fresh socket each run: clear any stale file left over from a crashed
	// predecessor. Ignoring ENOENT keeps the clean case quiet.
	if err := os.Remove(s.opts.SocketPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("clear stale socket: %w", err)
	}
	addr, err := net.ResolveUnixAddr("unix", s.opts.SocketPath)
	if err != nil {
		return fmt.Errorf("resolve socket: %w", err)
	}
	l, err := net.ListenUnix("unix", addr)
	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}
	if err := os.Chmod(s.opts.SocketPath, 0600); err != nil {
		_ = l.Close()
		_ = os.Remove(s.opts.SocketPath)
		return fmt.Errorf("chmod socket: %w", err)
	}
	// Hand the socket to the launching user so their connect(2) succeeds.
	// A root-owned mode-0600 socket is not reachable by the user that LaunchPal
	// runs as, which is the exact caller we want to let in. Peer-UID
	// verification in the accept loop is the second line of defense.
	if s.opts.SocketOwnerUID > 0 {
		if err := os.Lchown(s.opts.SocketPath, s.opts.SocketOwnerUID, s.opts.SocketOwnerGID); err != nil {
			_ = l.Close()
			_ = os.Remove(s.opts.SocketPath)
			return fmt.Errorf("chown socket: %w", err)
		}
	}
	s.listener = l

	go s.acceptLoop()
	if s.opts.IdleTimeout > 0 {
		ctx, cancel := context.WithCancel(context.Background())
		s.idleCancel = cancel
		go s.idleWatchdog(ctx)
	}
	return nil
}

// Wait blocks until Stop has been called and all accepted connections have
// finished processing. It returns the first error encountered (if any).
func (s *Server) Wait() error {
	<-s.stopSignal
	s.connWG.Wait()
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.stopErr
}

// Stop closes the listener, removes the socket file, and signals any waiters.
// Subsequent calls are no-ops.
func (s *Server) Stop() {
	s.mu.Lock()
	if s.stopped {
		s.mu.Unlock()
		return
	}
	s.stopped = true
	if s.listener != nil {
		_ = s.listener.Close()
	}
	// Close any accepted connection still being served. Without this, a stop
	// triggered while the GUI holds a long-lived idle connection (idle timeout
	// or parent watchdog) would leave handleConn blocked on scanner.Scan() and
	// the process would never exit. Closing unblocks the scan so handleConn
	// returns and connWG drains.
	for c := range s.conns {
		_ = c.Close()
	}
	if s.idleCancel != nil {
		s.idleCancel()
	}
	// The stopped guard above ensures close() runs exactly once, so no
	// sync.Once ceremony is needed.
	close(s.stopSignal)
	s.mu.Unlock()

	_ = os.Remove(s.opts.SocketPath)
}

// RequestShutdown signals a graceful shutdown from inside a handler: the
// handler returns success to the client, then the server stops after the
// response has been written.
func (s *Server) RequestShutdown() {
	select {
	case s.shutdownCalls <- struct{}{}:
	default:
	}
}

func (s *Server) isStopped() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.stopped
}

func (s *Server) acceptLoop() {
	for {
		conn, err := s.listener.AcceptUnix()
		if err != nil {
			if s.isStopped() {
				return
			}
			if errors.Is(err, net.ErrClosed) {
				return
			}
			s.mu.Lock()
			if s.stopErr == nil {
				s.stopErr = fmt.Errorf("accept: %w", err)
			}
			s.mu.Unlock()
			continue
		}

		uid, err := s.opts.PeerUID(conn)
		if err != nil || uid != s.opts.AuthorizedUID {
			// Close without reading input: mismatched peers never get to
			// send us a single byte. Spec: "Peer UID verification".
			_ = conn.Close()
			continue
		}

		s.mu.Lock()
		s.conns[conn] = struct{}{}
		// The first accepted connection is the session client (the GUI, which
		// connects immediately after launch and holds one long-lived
		// connection). Only its ending self-terminates the helper.
		isPrimary := s.primaryConn == nil
		if isPrimary {
			s.primaryConn = conn
		}
		s.mu.Unlock()

		s.connWG.Add(1)
		go func() {
			defer s.connWG.Done()
			defer func() {
				s.mu.Lock()
				delete(s.conns, conn)
				s.mu.Unlock()
				_ = conn.Close()
			}()
			s.handleConn(conn)
			// Single-client design: once the session client's connection ends
			// for any reason — clean EOF, a read/scan error, or a failed
			// response write — self-terminate rather than lingering as a root
			// socket. Covering every handleConn return path (not just the
			// post-scan EOF path) is the primary teardown mechanism, since the
			// unprivileged GUI cannot signal the root helper. Only the primary
			// connection triggers this: a stray same-UID connection opening and
			// closing must not tear down the live session. Skip when the server
			// is already stopping for another reason (idle, watchdog, or an
			// in-flight shutdown request).
			if isPrimary && !s.isStopped() {
				s.Stop()
			}
		}()
	}
}

// handleConn processes RPC requests serially on a single accepted connection.
// Requests are framed as newline-delimited JSON; responses are written in the
// same order and one at a time.
func (s *Server) handleConn(conn *net.UnixConn) {
	scanner := bufio.NewScanner(conn)
	// Plist bodies can be hundreds of KB; accommodate up to 4 MiB per line.
	scanner.Buffer(make([]byte, 64*1024), 4*1024*1024)

	encoder := json.NewEncoder(conn)

	for scanner.Scan() {
		if s.isStopped() {
			return
		}
		line := scanner.Bytes()

		var req Request
		var reqIDPtr *int64
		if err := json.Unmarshal(line, &req); err != nil {
			_ = encoder.Encode(NewErrorResponse(nil, ErrCodeInvalidRequest, err.Error()))
			continue
		}
		rid := req.ID
		reqIDPtr = &rid

		s.touch()
		result, rpcErr := s.opts.Handler(context.Background(), &req)
		s.touch()

		var resp Response
		if rpcErr != nil {
			resp = NewErrorResponse(reqIDPtr, rpcErr.Code, rpcErr.Message)
		} else {
			var err error
			resp, err = NewResultResponse(req.ID, result)
			if err != nil {
				resp = NewErrorResponse(reqIDPtr, ErrCodeInternalError, err.Error())
			}
		}

		if err := encoder.Encode(resp); err != nil {
			return
		}

		// If a handler requested shutdown during this iteration, finish the
		// response then tear down the server. The client's last response was
		// already flushed by the encoder. Stop synchronously: this is the
		// explicit-shutdown path (works regardless of which connection sent
		// it) and setting stopped here means the acceptLoop's post-return
		// primary check sees isStopped() and does not double-stop.
		select {
		case <-s.shutdownCalls:
			s.Stop()
			return
		default:
		}
	}

	if err := scanner.Err(); err != nil && !errors.Is(err, io.EOF) {
		s.mu.Lock()
		if s.stopErr == nil {
			s.stopErr = fmt.Errorf("scan: %w", err)
		}
		s.mu.Unlock()
	}
}

// idleWatchdog stops the server once IdleTimeout elapses without activity.
func (s *Server) idleWatchdog(ctx context.Context) {
	tick := s.opts.IdleTimeout / 4
	if tick < 50*time.Millisecond {
		tick = 50 * time.Millisecond
	}
	t := time.NewTicker(tick)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			last := time.Unix(0, s.activityAtNs.Load())
			if s.opts.Now().Sub(last) >= s.opts.IdleTimeout {
				if s.opts.OnIdle != nil {
					s.opts.OnIdle()
				}
				s.Stop()
				return
			}
		}
	}
}

// StartParentWatchdog polls for the existence of parentPID at interval and
// calls onExit (and Stop) when the parent goes away. Runs until the returned
// cancel is called or the parent disappears.
//
// Using syscall.Kill(pid, 0) is the standard "is this process alive" test on
// POSIX: zero signal performs the permission check without delivering a
// signal, returning an error (ESRCH) when no such process exists.
func (s *Server) StartParentWatchdog(parentPID int, interval time.Duration, alive func(int) bool, onExit func()) context.CancelFunc {
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		t := time.NewTicker(interval)
		defer t.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				if !alive(parentPID) {
					if onExit != nil {
						onExit()
					}
					s.Stop()
					return
				}
			}
		}
	}()
	return cancel
}
