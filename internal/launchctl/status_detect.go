package launchctl

import (
	"os/exec"
	"strconv"
	"strings"
)

// commonShells lists shell executable paths that should not be matched by pgrep,
// because the launchctl list output for a shell-wrapped daemon would otherwise
// collide with unrelated user shell processes.
var commonShells = map[string]bool{
	"/bin/bash":     true,
	"/bin/sh":       true,
	"/bin/zsh":      true,
	"/usr/bin/bash": true,
	"/usr/bin/sh":   true,
	"/usr/bin/zsh":  true,
}

// Package-level function variables so tests can substitute stub implementations
// without spawning real processes.
var (
	pgrepCandidatesFn = runPgrepCandidates
	readAllPPIDsFn    = readAllPPIDs
)

// DetectSystemServiceStatus returns the runtime status, PID, and detection
// confidence for a system-domain daemon. If ppidTable is non-nil, parent-PID
// lookups read from it (avoiding a `ps` fork per candidate); otherwise the
// function fetches the full ppid table once. Callers that invoke this for many
// services (e.g. List) should pre-fetch the table via readAllPPIDs and pass it
// in to amortize the cost.
func DetectSystemServiceStatus(pd plistData, ppidTable map[int]int) (status string, pid int, confidence string) {
	user := pd.UserName
	if user == "" {
		user = "root"
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

	candidates, err := pgrepCandidatesFn(user, program)
	if err != nil || len(candidates) == 0 {
		return StatusStopped, 0, ConfidenceVerified
	}

	table := ppidTable
	if table == nil {
		var fetchErr error
		table, fetchErr = readAllPPIDsFn()
		if fetchErr != nil || table == nil {
			// Cannot determine parent PIDs; candidates exist but cannot be
			// filtered. Reporting Stopped would be a confident false negative,
			// so degrade to Running with unverified confidence.
			return StatusRunning, candidates[0], ConfidenceUnverified
		}
	}

	kept := make([]int, 0, len(candidates))
	for _, c := range candidates {
		if table[c] == 1 {
			kept = append(kept, c)
		}
	}

	switch len(kept) {
	case 0:
		return StatusStopped, 0, ConfidenceVerified
	case 1:
		return StatusRunning, kept[0], ConfidenceVerified
	default:
		return StatusRunning, kept[0], ConfidenceUnverified
	}
}

// runPgrepCandidates executes `pgrep -u <user> -f <program>` as an argv slice
// (no shell interpretation) and returns the matched PIDs.
func runPgrepCandidates(user, program string) ([]int, error) {
	cmd := exec.Command("pgrep", "-u", user, "-f", program)
	out, err := cmd.Output()
	if err != nil {
		// Exit code 1 from pgrep means "no matches", which is not a real error
		// for our purposes. Return empty slice in that case.
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return nil, nil
		}
		return nil, err
	}

	lines := strings.Split(string(out), "\n")
	pids := make([]int, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if pid, err := strconv.Atoi(line); err == nil && pid > 0 {
			pids = append(pids, pid)
		}
	}
	return pids, nil
}

// readAllPPIDs returns a pid→ppid map for every process visible to the current
// user, via a single `ps -axo pid=,ppid=` invocation. Callers amortize the cost
// across many daemon checks by sharing one result.
func readAllPPIDs() (map[int]int, error) {
	cmd := exec.Command("ps", "-axo", "pid=,ppid=")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	table := make(map[int]int)
	for _, line := range strings.Split(string(out), "\n") {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		pid, err1 := strconv.Atoi(fields[0])
		ppid, err2 := strconv.Atoi(fields[1])
		if err1 != nil || err2 != nil {
			continue
		}
		table[pid] = ppid
	}
	return table, nil
}
