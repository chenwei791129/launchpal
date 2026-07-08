// Command launchpal-privhelper is a root-privileged RPC server launched on
// demand by LaunchPal when the user enables Admin Mode. See
// internal/privhelper and openspec/specs for the full protocol description.
package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"launchpal/internal/privhelper"
)

// helperConfig captures everything main needs to run, assembled from CLI
// flags. Keeping the struct separate from flag parsing makes validation
// testable without shelling out.
type helperConfig struct {
	EffectiveUID int
	Socket       string
	ParentPID    int
	LaunchingUID int
	UserHome     string
}

// parseFlags reads command-line arguments into a helperConfig. Errors are
// written to the provided writer (for CLI error reporting under osascript,
// stderr is swallowed so returning an error is just for tests / non-prod).
func parseFlags(args []string, stderr io.Writer) (*helperConfig, error) {
	fs := flag.NewFlagSet("launchpal-privhelper", flag.ContinueOnError)
	fs.SetOutput(stderr)

	socket := fs.String("socket", "", "Unix domain socket path to listen on")
	parentPID := fs.Int("parent-pid", 0, "PID of the LaunchPal parent process")
	launchingUID := fs.Int("launching-uid", -1, "UID of the user who launched the helper")
	userHome := fs.String("user-home", "", "Home directory of the launching user (for backup destination)")

	if err := fs.Parse(args); err != nil {
		return nil, err
	}

	return &helperConfig{
		EffectiveUID: syscall.Geteuid(),
		Socket:       *socket,
		ParentPID:    *parentPID,
		LaunchingUID: *launchingUID,
		UserHome:     *userHome,
	}, nil
}

// validateArgs enforces the "refuse to run without required conditions"
// requirement: root-only execution, non-empty socket, resolvable parent PID,
// a positive non-root launching UID, and — if provided — an absolute
// UserHome path.
func validateArgs(cfg *helperConfig) error {
	if cfg.EffectiveUID != 0 {
		return errors.New("helper must run as root (effective UID 0)")
	}
	if cfg.Socket == "" {
		return errors.New("--socket is required")
	}
	if cfg.ParentPID <= 0 {
		return errors.New("--parent-pid must be a positive integer")
	}
	// LaunchingUID must identify a specific non-root session user. Accepting
	// 0 here would let the server authorize any root-owned peer, defeating
	// the point of per-user peer-UID verification.
	if cfg.LaunchingUID <= 0 {
		return errors.New("--launching-uid must be a positive non-root UID")
	}
	if cfg.UserHome != "" {
		if err := validateUserHome(cfg.UserHome); err != nil {
			return err
		}
	}
	return nil
}

// validateUserHome refuses relative paths, filesystem root, and anything
// that would land inside a system directory. The helper runs as root, so
// `--user-home /` would let a compromised parent create directories at
// arbitrary locations; this check keeps the backup destination bounded.
func validateUserHome(p string) error {
	if !filepath.IsAbs(p) {
		return errors.New("--user-home must be an absolute path")
	}
	clean := filepath.Clean(p)
	if clean == "/" {
		return errors.New("--user-home must not be filesystem root")
	}
	// Defensively reject paths that sit under obvious system directories.
	// macOS user homes live under /Users or /var — never under /Library,
	// /System, /private/etc, or similar.
	forbiddenPrefixes := []string{
		"/System",
		"/Library",
		"/bin", "/sbin", "/usr",
		"/etc",
		"/private/etc", "/private/var/db",
		"/Applications",
	}
	for _, prefix := range forbiddenPrefixes {
		if clean == prefix || strings.HasPrefix(clean, prefix+"/") {
			return fmt.Errorf("--user-home must not be under %s", prefix)
		}
	}
	return nil
}

// parentAlive returns true when pid refers to a live process. A nil
// syscall.Kill(pid, 0) succeeds when the caller has permission to signal
// the process (always, for root) and it exists.
func parentAlive(pid int) bool {
	return syscall.Kill(pid, 0) == nil
}

// idleTimeout is the dormant-period after which the helper shuts itself
// down (spec: "Idle timeout"). Exposed as a var so tests could override,
// though production always uses the 30-minute default.
var idleTimeout = 30 * time.Minute

// parentCheckInterval is how often the parent-watchdog polls (spec: "Parent
// PID watchdog"). Exposed as a var for the same reason as idleTimeout.
var parentCheckInterval = 1 * time.Second

func main() {
	cfg, err := parseFlags(os.Args[1:], os.Stderr)
	if err != nil {
		os.Exit(2)
	}
	if err := validateArgs(cfg); err != nil {
		fmt.Fprintln(os.Stderr, "launchpal-privhelper:", err)
		os.Exit(1)
	}
	os.Exit(run(cfg))
}

// run wires the privhelper server to the configured socket, starts the
// parent-PID watchdog, and blocks until shutdown. Returns the process exit
// code. Split out from main() so tests can exercise the wiring without
// calling os.Exit.
func run(cfg *helperConfig) int {
	// Provision the root-owned protected copy before binding the socket. This
	// runs in the root helper process so the privileged write never crosses an
	// RPC boundary. Failure is non-fatal (see selfInstallProtectedCopy).
	if exe, err := os.Executable(); err != nil {
		fmt.Fprintln(os.Stderr, "launchpal-privhelper: os.Executable:", err)
	} else {
		selfInstallProtectedCopy(exe, privhelper.InstallProtectedCopy,
			func(format string, args ...any) { fmt.Fprintf(os.Stderr, format+"\n", args...) })
	}

	launchingGID, _ := lookupGID(cfg.LaunchingUID)

	// Handlers must be declared before the server so the server's handler
	// callback can reference them; but handlers need the server to wire
	// ShutdownFn. Build in two steps with a mutable reference.
	var server *privhelper.Server
	handlers := privhelper.NewHandlers(privhelper.HandlerOptions{
		UserHome:     cfg.UserHome,
		LaunchingUID: cfg.LaunchingUID,
		LaunchingGID: launchingGID,
		ShutdownFn:   func() { server.RequestShutdown() },
	})

	server = privhelper.NewServer(privhelper.ServerOptions{
		SocketPath:     cfg.Socket,
		AuthorizedUID:  cfg.LaunchingUID,
		SocketOwnerUID: cfg.LaunchingUID,
		SocketOwnerGID: launchingGID,
		Handler:        handlers.Handle,
		IdleTimeout:    idleTimeout,
	})

	if err := server.Start(); err != nil {
		fmt.Fprintln(os.Stderr, "launchpal-privhelper: start server:", err)
		return 1
	}

	watchdogCancel := server.StartParentWatchdog(cfg.ParentPID, parentCheckInterval, parentAlive, nil)
	defer watchdogCancel()

	if err := server.Wait(); err != nil {
		fmt.Fprintln(os.Stderr, "launchpal-privhelper:", err)
		return 1
	}
	return 0
}

// selfInstallProtectedCopy provisions the root-owned protected helper copy at
// startup. It runs only when the helper was launched from a path other than
// the protected path (i.e. from the app bundle) — launching from the protected
// path already means the copy is in place. The skip check and the install
// target derive from the same privhelper.ProtectedHelperPath constant, so they
// cannot diverge. Failure is non-fatal: the helper logs it and keeps serving
// from the binary it was launched as, and the protected copy is retried on the
// next enable. A provisioning failure must not make the current Admin Mode
// session unusable.
func selfInstallProtectedCopy(exePath string, install func(string) (bool, error), logf func(string, ...any)) {
	if exePath == privhelper.ProtectedHelperPath {
		return
	}
	if _, err := install(exePath); err != nil {
		logf("launchpal-privhelper: protected-copy install failed: %v", err)
	}
}

// lookupGID resolves the primary GID for uid. A zero GID is fine if lookup
// fails (helper keeps running as root-owned files); backup chown will just
// fall back to uid:0, which the user can still read.
func lookupGID(uid int) (int, error) {
	u, err := user.LookupId(strconv.Itoa(uid))
	if err != nil {
		return 0, err
	}
	gid, err := strconv.Atoi(u.Gid)
	if err != nil {
		return 0, err
	}
	return gid, nil
}
