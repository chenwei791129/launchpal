package launchctl

import (
	"fmt"
	"os/exec"
	"os/user"
	"slices"
	"strconv"
	"strings"
)

// commonShells lists shell executable paths that should not be matched by the
// process-table scan. For a shell-wrapped daemon the shell would otherwise
// collide with unrelated user shell processes.
var commonShells = map[string]bool{
	"/bin/bash":     true,
	"/bin/sh":       true,
	"/bin/zsh":      true,
	"/usr/bin/bash": true,
	"/usr/bin/sh":   true,
	"/usr/bin/zsh":  true,
}

// processInfo holds the uid, ppid, and full argv string for a single running
// process. ProcessTable is keyed by pid.
type processInfo struct {
	UID  int
	PPID int
	Args string
}

// ProcessTable maps pid → processInfo. Callers that detect many services at
// once should build the table once and share it.
type ProcessTable map[int]processInfo

// Package-level function variables so tests can substitute stub implementations
// without spawning real subprocesses.
var (
	readProcessTableFn = readProcessTable
	userLookupFn       = user.Lookup
)

// DetectSystemServiceStatus returns the runtime status, PID, and detection
// confidence for a system-domain daemon. The algorithm mirrors
// `pgrep -u <user> -f <program>` + `ps ppid=1` filtering but operates on a
// pre-fetched process table to avoid per-service subprocess forks.
//
// The program match uses plain substring (strings.Contains) rather than
// pgrep's regex semantics. In practice daemon Program paths are full absolute
// paths without regex metacharacters, so this is strictly narrower and
// cannot produce false positives that pgrep would have rejected.
//
// If table is nil, it is fetched lazily via readProcessTableFn. If uidCache is
// nil, a single-shot cache is created for the call; callers doing many
// detections in one List should share a cache to avoid repeated
// `os/user.Lookup` calls.
func DetectSystemServiceStatus(pd plistData, table ProcessTable, uidCache map[string]int) (status string, pid int, confidence string) {
	userName := pd.UserName
	if userName == "" {
		userName = "root"
	}

	program := pd.Program
	if program == "" && len(pd.ProgramArguments) > 0 {
		program = pd.ProgramArguments[0]
	}

	if program == "" {
		return StatusUnknown, 0, ConfidenceUnverified
	}

	if commonShells[program] {
		return StatusLoaded, 0, ConfidenceVerified
	}

	uid, uidErr := resolveUID(userName, uidCache)
	if uidErr != nil {
		// Without a resolvable uid we cannot filter candidates; degrade to
		// unverified rather than reporting confident Stopped.
		return StatusStopped, 0, ConfidenceUnverified
	}

	if table == nil {
		var err error
		table, err = readProcessTableFn()
		if err != nil || table == nil {
			// Cannot distinguish "not running" from "couldn't check"; a
			// confident Stopped would be misleading, so report unverified.
			return StatusStopped, 0, ConfidenceUnverified
		}
	}

	candidates := make([]int, 0)
	for pid, info := range table {
		if info.UID != uid {
			continue
		}
		if info.PPID != 1 {
			continue
		}
		if !strings.Contains(info.Args, program) {
			continue
		}
		candidates = append(candidates, pid)
	}
	slices.Sort(candidates)

	switch len(candidates) {
	case 0:
		return StatusStopped, 0, ConfidenceVerified
	case 1:
		return StatusRunning, candidates[0], ConfidenceVerified
	default:
		return StatusRunning, candidates[0], ConfidenceUnverified
	}
}

// resolveUID turns a system user name into a numeric uid, consulting the
// cache first. On a cache miss it calls userLookupFn and stores the result.
// If cache is nil a scratch cache is used.
func resolveUID(name string, cache map[string]int) (int, error) {
	if cache != nil {
		if uid, ok := cache[name]; ok {
			return uid, nil
		}
	}
	u, err := userLookupFn(name)
	if err != nil {
		return 0, err
	}
	uid, err := strconv.Atoi(u.Uid)
	if err != nil {
		return 0, fmt.Errorf("invalid uid %q for user %s: %w", u.Uid, name, err)
	}
	if cache != nil {
		cache[name] = uid
	}
	return uid, nil
}

// readProcessTable runs a single `ps -axo uid=,pid=,ppid=,args=` and returns
// the parsed table. Output format: each line has three whitespace-separated
// integers followed by the full argv string (which may itself contain
// whitespace and runs to end-of-line).
func readProcessTable() (ProcessTable, error) {
	cmd := exec.Command("ps", "-axo", "uid=,pid=,ppid=,args=")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	return parseProcessTable(string(out)), nil
}

// parseProcessTable parses raw `ps -axo uid=,pid=,ppid=,args=` output. It is
// split out from readProcessTable so tests can exercise parsing deterministically.
//
// Each line has three numeric fields followed by the full argv string, which
// may itself contain whitespace and runs to end-of-line. `strings.Fields` on
// the whole line would collapse spaces inside argv, so the three leading
// fields are carved off with `strings.Cut` and the remainder kept verbatim.
func parseProcessTable(raw string) ProcessTable {
	table := make(ProcessTable, 512)
	for _, line := range strings.Split(raw, "\n") {
		uidStr, rest, ok := cutField(line)
		if !ok {
			continue
		}
		pidStr, rest, ok := cutField(rest)
		if !ok {
			continue
		}
		ppidStr, args, ok := cutField(rest)
		if !ok {
			continue
		}
		uid, err1 := strconv.Atoi(uidStr)
		pid, err2 := strconv.Atoi(pidStr)
		ppid, err3 := strconv.Atoi(ppidStr)
		if err1 != nil || err2 != nil || err3 != nil {
			continue
		}
		table[pid] = processInfo{UID: uid, PPID: ppid, Args: args}
	}
	return table
}

// cutField trims leading spaces and returns the first space-delimited field
// plus the remainder (leading spaces stripped). Returns ok=false if no field
// is present. macOS `ps` right-aligns numeric columns with spaces, never
// tabs — so splitting on " " is sufficient.
func cutField(s string) (field, rest string, ok bool) {
	s = strings.TrimLeft(s, " ")
	if s == "" {
		return "", "", false
	}
	before, after, found := strings.Cut(s, " ")
	if !found {
		return before, "", true
	}
	return before, strings.TrimLeft(after, " "), true
}
