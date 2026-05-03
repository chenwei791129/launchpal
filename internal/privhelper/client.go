package privhelper

import (
	"bufio"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// ErrAuthorizationCanceled signals that the user dismissed the osascript
// password prompt. Callers should treat this as a benign "user said no"
// rather than an error that warrants retry.
var ErrAuthorizationCanceled = errors.New("authorization cancelled by user")

// ErrClientClosed is returned from Call after the client's connection has
// been closed (gracefully or due to helper crash). Distinguishes a closed
// client from a per-request error.
var ErrClientClosed = errors.New("privhelper: client closed")

// Client is a LaunchPal-side RPC client that talks to a running helper
// process over a connected Unix domain socket. Responses are correlated to
// requests by the monotonic id field.
type Client struct {
	conn    net.Conn
	encoder *json.Encoder
	closed  atomic.Bool

	mu      sync.Mutex
	nextID  int64
	pending map[int64]chan Response
	closeCh chan struct{}

	// onDisconnect is called once when the background reader observes that
	// the connection has ended (EOF, helper crash, or deliberate Close).
	onDisconnect func(err error)
}

// ClientOptions configure a Client.
type ClientOptions struct {
	// Conn is an already-connected Unix socket. Connect() is the usual
	// path to obtain one; tests inject a net.Pipe-based pair.
	Conn net.Conn

	// OnDisconnect is invoked once (best effort) when the background reader
	// stops — either due to EOF, an I/O error, or a deliberate Close. The
	// error argument is non-nil for unexpected disconnects and nil for a
	// clean Close.
	OnDisconnect func(err error)
}

// NewClient wraps a live Unix socket connection. The caller must have
// already verified the peer via whatever handshake is appropriate; NewClient
// itself is transport-agnostic.
func NewClient(opts ClientOptions) *Client {
	c := &Client{
		conn:         opts.Conn,
		encoder:      json.NewEncoder(opts.Conn),
		pending:      make(map[int64]chan Response),
		closeCh:      make(chan struct{}),
		onDisconnect: opts.OnDisconnect,
	}
	go c.readLoop()
	return c
}

// Call sends a request and waits for the matching response. If the client
// closes before the response arrives, ErrClientClosed is returned.
func (c *Client) Call(ctx context.Context, method string, params any) (json.RawMessage, error) {
	if c.closed.Load() {
		return nil, ErrClientClosed
	}
	ch := make(chan Response, 1)
	c.mu.Lock()
	c.nextID++
	id := c.nextID
	c.pending[id] = ch
	c.mu.Unlock()

	defer func() {
		c.mu.Lock()
		delete(c.pending, id)
		c.mu.Unlock()
	}()

	var raw json.RawMessage
	if params != nil {
		data, err := json.Marshal(params)
		if err != nil {
			return nil, fmt.Errorf("marshal params: %w", err)
		}
		raw = data
	}
	req := Request{ID: id, Method: method, Params: raw}

	c.mu.Lock()
	err := c.encoder.Encode(req)
	c.mu.Unlock()
	if err != nil {
		return nil, fmt.Errorf("send: %w", err)
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-c.closeCh:
		return nil, ErrClientClosed
	case resp := <-ch:
		if resp.Error != nil {
			return nil, resp.Error
		}
		return resp.Result, nil
	}
}

// Ping is a convenience wrapper around the Ping RPC, used by handshake.
func (c *Client) Ping(ctx context.Context) error {
	_, err := c.Call(ctx, MethodPing, nil)
	return err
}

// Bootstrap sends a Bootstrap RPC for plistPath.
func (c *Client) Bootstrap(ctx context.Context, plistPath string) error {
	_, err := c.Call(ctx, MethodBootstrap, BootstrapParams{PlistPath: plistPath})
	return err
}

// Bootout sends a Bootout RPC for the given label.
func (c *Client) Bootout(ctx context.Context, label string) error {
	_, err := c.Call(ctx, MethodBootout, BootoutParams{Label: label})
	return err
}

// Kickstart sends a Kickstart RPC for the given label.
func (c *Client) Kickstart(ctx context.Context, label string) error {
	_, err := c.Call(ctx, MethodKickstart, KickstartParams{Label: label})
	return err
}

// WritePlist sends a WritePlist RPC with base64-encoded data.
func (c *Client) WritePlist(ctx context.Context, plistPath string, data []byte) error {
	encoded := base64.StdEncoding.EncodeToString(data)
	_, err := c.Call(ctx, MethodWritePlist, WritePlistParams{PlistPath: plistPath, Data: encoded})
	return err
}

// DeletePlist sends a DeletePlist RPC for plistPath.
func (c *Client) DeletePlist(ctx context.Context, plistPath string) error {
	_, err := c.Call(ctx, MethodDeletePlist, DeletePlistParams{PlistPath: plistPath})
	return err
}

// EnsureLogAccess asks the helper to make each non-empty log path reachable
// by the unprivileged GUI process (traversable parent + readable file).
// Intended to be called right after WritePlist during Create/Update so the
// Logs tab can tail the files once the daemon starts writing them.
func (c *Client) EnsureLogAccess(ctx context.Context, paths []string) error {
	_, err := c.Call(ctx, MethodEnsureLogAccess, EnsureLogAccessParams{Paths: paths})
	return err
}

// TruncateLog asks the helper to truncate an existing log file at path to 0
// bytes. Path validation, O_NOFOLLOW, and the missing-file rejection happen
// helper-side; this wrapper just forwards the request and surfaces the
// helper's error code unchanged.
func (c *Client) TruncateLog(ctx context.Context, path string) error {
	_, err := c.Call(ctx, MethodTruncateLog, TruncateLogParams{Path: path})
	return err
}

// ListSystemDaemons sends a ListSystemDaemons RPC and returns the parsed
// list of daemons known to launchd.
func (c *Client) ListSystemDaemons(ctx context.Context) ([]DaemonInfo, error) {
	raw, err := c.Call(ctx, MethodListSystemDaemons, nil)
	if err != nil {
		return nil, err
	}
	var res ListSystemDaemonsResult
	if err := json.Unmarshal(raw, &res); err != nil {
		return nil, fmt.Errorf("decode list: %w", err)
	}
	return res.Daemons, nil
}

// GetSystemDaemon sends a GetSystemDaemon RPC for the given label.
func (c *Client) GetSystemDaemon(ctx context.Context, label string) (DaemonInfo, error) {
	raw, err := c.Call(ctx, MethodGetSystemDaemon, GetSystemDaemonParams{Label: label})
	if err != nil {
		return DaemonInfo{}, err
	}
	var info DaemonInfo
	if err := json.Unmarshal(raw, &info); err != nil {
		return DaemonInfo{}, fmt.Errorf("decode daemon: %w", err)
	}
	return info, nil
}

// Shutdown sends a Shutdown RPC; the helper is expected to reply with OK
// and close the connection shortly after. Errors that look like the helper
// already closing the connection are not returned to callers.
func (c *Client) Shutdown(ctx context.Context) error {
	_, err := c.Call(ctx, MethodShutdown, nil)
	if err == nil || errors.Is(err, ErrClientClosed) {
		return nil
	}
	return err
}

// Close terminates the client connection. Concurrent Call invocations
// unblock with ErrClientClosed. Safe to call multiple times.
func (c *Client) Close() error {
	if c.closed.Swap(true) {
		return nil
	}
	err := c.conn.Close()
	close(c.closeCh)
	return err
}

// readLoop reads responses off the wire and dispatches them to the waiting
// channel. It exits on EOF or read error; on exit it closes the client.
func (c *Client) readLoop() {
	var disconnectErr error
	defer func() {
		c.closed.Store(true)
		_ = c.conn.Close()
		select {
		case <-c.closeCh:
		default:
			close(c.closeCh)
		}
		c.mu.Lock()
		for id, ch := range c.pending {
			select {
			case ch <- Response{ID: &id, Error: &RPCError{Code: ErrCodeIOError, Message: "client closed"}}:
			default:
			}
			delete(c.pending, id)
		}
		c.mu.Unlock()
		if c.onDisconnect != nil {
			c.onDisconnect(disconnectErr)
		}
	}()

	scanner := bufio.NewScanner(c.conn)
	scanner.Buffer(make([]byte, 64*1024), 4*1024*1024)
	for scanner.Scan() {
		line := scanner.Bytes()
		var resp Response
		if err := json.Unmarshal(line, &resp); err != nil {
			// Malformed response from helper; treat like a disconnect.
			disconnectErr = fmt.Errorf("decode: %w", err)
			return
		}
		if resp.ID == nil {
			// Helper-initiated message with no correlation; ignore for now.
			continue
		}
		c.mu.Lock()
		ch, ok := c.pending[*resp.ID]
		c.mu.Unlock()
		if !ok {
			continue
		}
		select {
		case ch <- resp:
		default:
		}
	}
	if err := scanner.Err(); err != nil {
		disconnectErr = err
	}
}

// GenerateSocketPath returns $TMPDIR/launchpal-<uid>-<16 hex>.sock — the path
// convention LaunchPal uses for the per-session helper socket.
func GenerateSocketPath(uid int) (string, error) {
	dir := os.TempDir()
	var buf [8]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return "", fmt.Errorf("rand: %w", err)
	}
	name := fmt.Sprintf("launchpal-%d-%s.sock", uid, hex.EncodeToString(buf[:]))
	return filepath.Join(dir, name), nil
}

// Connect dials the Unix socket at path, retrying with exponential backoff
// until either it succeeds or the timeout elapses. Used after LaunchHelper
// since the helper's socket may not be ready for a brief moment.
func Connect(ctx context.Context, path string, timeout time.Duration) (*net.UnixConn, error) {
	deadline := time.Now().Add(timeout)
	backoff := 20 * time.Millisecond
	maxBackoff := 500 * time.Millisecond
	var lastErr error
	for {
		conn, err := net.DialUnix("unix", nil, &net.UnixAddr{Name: path, Net: "unix"})
		if err == nil {
			return conn, nil
		}
		lastErr = err
		if time.Now().After(deadline) {
			return nil, fmt.Errorf("connect timeout: %w", lastErr)
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(backoff):
		}
		backoff *= 2
		if backoff > maxBackoff {
			backoff = maxBackoff
		}
	}
}

// LaunchHelperOptions configures a helper launch via osascript.
type LaunchHelperOptions struct {
	HelperPath   string
	SocketPath   string
	ParentPID    int
	LaunchingUID int
	UserHome     string

	// HandshakeTimeout is how long to wait for the helper to come up and
	// respond to Ping. Defaults to 10 seconds per the spec.
	HandshakeTimeout time.Duration

	// OnDisconnect is wired into the resulting Client so callers can react
	// to helper crashes without polling. See ClientOptions.OnDisconnect.
	OnDisconnect func(err error)

	// osascriptRunner runs osascript; injectable for tests. Nil uses the
	// real osascript binary.
	osascriptRunner func(ctx context.Context, script string) error
}

// LaunchHelper executes the helper via osascript with administrator
// privileges, then connects and performs a Ping handshake. Returns a live
// Client on success.
//
// osascript's "do shell script ... with administrator privileges" throws a
// specific AppleScript error when the user dismisses the prompt; we pattern
// match the exit error message to surface that as ErrAuthorizationCanceled.
func LaunchHelper(ctx context.Context, opts LaunchHelperOptions) (*Client, error) {
	if opts.HandshakeTimeout <= 0 {
		opts.HandshakeTimeout = 10 * time.Second
	}
	if opts.osascriptRunner == nil {
		opts.osascriptRunner = runOsascript
	}

	// Build the shell command the helper runs under root. Arguments are
	// individually shell-quoted by shellQuote (since osascript's `do shell
	// script` interprets a single string, not an argv array).
	helperCmd := shellQuote(opts.HelperPath) +
		" --socket " + shellQuote(opts.SocketPath) +
		" --parent-pid " + fmt.Sprint(opts.ParentPID) +
		" --launching-uid " + fmt.Sprint(opts.LaunchingUID)
	if opts.UserHome != "" {
		helperCmd += " --user-home " + shellQuote(opts.UserHome)
	}
	helperCmd += " &> /dev/null &"

	script := fmt.Sprintf(
		`do shell script %s with administrator privileges`,
		appleScriptQuote(helperCmd),
	)

	if err := opts.osascriptRunner(ctx, script); err != nil {
		if isAuthorizationCanceled(err) {
			return nil, ErrAuthorizationCanceled
		}
		return nil, fmt.Errorf("osascript: %w", err)
	}

	conn, err := Connect(ctx, opts.SocketPath, opts.HandshakeTimeout)
	if err != nil {
		return nil, fmt.Errorf("helper handshake failed: %w", err)
	}
	client := NewClient(ClientOptions{Conn: conn, OnDisconnect: opts.OnDisconnect})

	pingCtx, cancel := context.WithTimeout(ctx, opts.HandshakeTimeout)
	defer cancel()
	if err := client.Ping(pingCtx); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("helper handshake failed: %w", err)
	}
	return client, nil
}

// runOsascript runs `osascript -e <script>` and returns any error. The
// default osascript error text is captured for cancel detection.
func runOsascript(ctx context.Context, script string) error {
	cmd := exec.CommandContext(ctx, "osascript", "-e", script)
	// Combine stdout + stderr so cancel-detection sees either stream.
	out, err := cmd.CombinedOutput()
	if err != nil {
		return &osascriptError{err: err, output: strings.TrimSpace(string(out))}
	}
	return nil
}

type osascriptError struct {
	err    error
	output string
}

func (e *osascriptError) Error() string {
	if e.output != "" {
		return e.output
	}
	return e.err.Error()
}

// isAuthorizationCanceled inspects an osascript error for the AppleScript
// cancel string. When the user clicks Cancel in the password prompt,
// osascript emits "User canceled. (-128)".
func isAuthorizationCanceled(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "user canceled") ||
		strings.Contains(msg, "user cancelled") ||
		strings.Contains(msg, "(-128)")
}

// shellQuote wraps a string in single quotes, escaping embedded single
// quotes. Only used inside osascript's shell-script string.
func shellQuote(s string) string {
	return `'` + strings.ReplaceAll(s, `'`, `'\''`) + `'`
}

// appleScriptQuote wraps a string in double quotes, escaping embedded double
// quotes and backslashes per AppleScript string literal rules.
func appleScriptQuote(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	return `"` + s + `"`
}
