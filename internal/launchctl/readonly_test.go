package launchctl

import (
	"os"
	"os/user"
	"path/filepath"
	"testing"
)

const sampleSystemPlist = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>Label</key>
	<string>com.test.sampled</string>
	<key>Program</key>
	<string>/usr/sbin/sampled</string>
</dict>
</plist>`

func writeSystemPlist(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name+".plist"), []byte(content), 0644); err != nil {
		t.Fatalf("write plist: %v", err)
	}
}

// TestReadOnly_GetWithStatus_MissInvokesDetection verifies that when the batch
// launchctl map does not contain the service label, heuristic detection runs
// against the shared process table.
func TestReadOnly_GetWithStatus_MissInvokesDetection(t *testing.T) {
	tmpDir := t.TempDir()
	writeSystemPlist(t, tmpDir, "com.test.sampled", sampleSystemPlist)

	withStubUserLookup(t, defaultStubUsers())

	table := ProcessTable{
		42: {UID: 0, PPID: 1, Args: "/usr/sbin/sampled"},
	}
	m := &readOnlyManager{basePath: tmpDir, serviceType: ServiceTypeSystem}

	svc, err := m.getWithStatus("com.test.sampled", map[string]serviceStatus{}, table, map[string]int{})
	if err != nil {
		t.Fatalf("getWithStatus error = %v", err)
	}

	if svc.Status != StatusRunning {
		t.Errorf("Status = %q, want %q", svc.Status, StatusRunning)
	}
	if svc.PID != 42 {
		t.Errorf("PID = %d, want 42", svc.PID)
	}
	if svc.StatusConfidence != ConfidenceVerified {
		t.Errorf("StatusConfidence = %q, want %q", svc.StatusConfidence, ConfidenceVerified)
	}
}

// TestReadOnly_GetWithStatus_MissWithMultipleCandidates verifies the
// unverified path when detection finds ambiguous matches.
func TestReadOnly_GetWithStatus_MissWithMultipleCandidates(t *testing.T) {
	tmpDir := t.TempDir()
	writeSystemPlist(t, tmpDir, "com.test.sampled", sampleSystemPlist)

	withStubUserLookup(t, defaultStubUsers())

	table := ProcessTable{
		10: {UID: 0, PPID: 1, Args: "/usr/sbin/sampled"},
		20: {UID: 0, PPID: 1, Args: "/usr/sbin/sampled"},
	}
	m := &readOnlyManager{basePath: tmpDir, serviceType: ServiceTypeSystem}
	svc, err := m.getWithStatus("com.test.sampled", map[string]serviceStatus{}, table, map[string]int{})
	if err != nil {
		t.Fatalf("getWithStatus error = %v", err)
	}

	if svc.Status != StatusRunning {
		t.Errorf("Status = %q, want %q", svc.Status, StatusRunning)
	}
	if svc.StatusConfidence != ConfidenceUnverified {
		t.Errorf("StatusConfidence = %q, want %q", svc.StatusConfidence, ConfidenceUnverified)
	}
}

// TestReadOnly_GetWithStatus_HitSkipsDetectionAndSetsVerified verifies that a
// batch map hit keeps the fast path and marks confidence = verified without
// consulting the process table or the user lookup.
func TestReadOnly_GetWithStatus_HitSkipsDetectionAndSetsVerified(t *testing.T) {
	tmpDir := t.TempDir()
	writeSystemPlist(t, tmpDir, "com.test.sampled", sampleSystemPlist)

	// Any call into user.Lookup or the process-table fetch would be a bug.
	withCustomUserLookup(t, func(string) (*user.User, error) {
		t.Fatal("user.Lookup should not be called when batch map has the label")
		return nil, nil
	})

	statusMap := map[string]serviceStatus{
		"com.test.sampled": {status: StatusRunning, pid: 7777},
	}
	m := &readOnlyManager{basePath: tmpDir, serviceType: ServiceTypeSystem}
	svc, err := m.getWithStatus("com.test.sampled", statusMap, nil, nil)
	if err != nil {
		t.Fatalf("getWithStatus error = %v", err)
	}

	if svc.Status != StatusRunning {
		t.Errorf("Status = %q, want %q", svc.Status, StatusRunning)
	}
	if svc.PID != 7777 {
		t.Errorf("PID = %d, want 7777", svc.PID)
	}
	if svc.StatusConfidence != ConfidenceVerified {
		t.Errorf("StatusConfidence = %q, want %q", svc.StatusConfidence, ConfidenceVerified)
	}
}

// TestReadOnly_GetWithStatus_NilStatusMapInvokesDetection ensures the
// individual-query path (statusMap == nil, table == nil) triggers lazy
// detection: readProcessTableFn is called once for this single service.
func TestReadOnly_GetWithStatus_NilStatusMapInvokesDetection(t *testing.T) {
	tmpDir := t.TempDir()
	writeSystemPlist(t, tmpDir, "com.test.sampled", sampleSystemPlist)

	withStubUserLookup(t, defaultStubUsers())
	withStubReadProcessTable(t, func() (ProcessTable, error) { return ProcessTable{}, nil })

	m := &readOnlyManager{basePath: tmpDir, serviceType: ServiceTypeSystem}
	svc, err := m.getWithStatus("com.test.sampled", nil, nil, nil)
	if err != nil {
		t.Fatalf("getWithStatus error = %v", err)
	}

	if svc.Status != StatusStopped {
		t.Errorf("Status = %q, want %q", svc.Status, StatusStopped)
	}
	if svc.StatusConfidence != ConfidenceVerified {
		t.Errorf("StatusConfidence = %q, want %q", svc.StatusConfidence, ConfidenceVerified)
	}
}
