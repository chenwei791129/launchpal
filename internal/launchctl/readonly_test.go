package launchctl

import (
	"os"
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
// launchctl map does not contain the service label, heuristic detection is
// invoked (replacing the old "default to Stopped" fallback).
func TestReadOnly_GetWithStatus_MissInvokesDetection(t *testing.T) {
	tmpDir := t.TempDir()
	writeSystemPlist(t, tmpDir, "com.test.sampled", sampleSystemPlist)

	detectCalled := false
	withStubPgrep(t, func(user, program string) ([]int, error) {
		detectCalled = true
		if program != "/usr/sbin/sampled" {
			t.Errorf("program = %q, want %q", program, "/usr/sbin/sampled")
		}
		return []int{42}, nil
	})

	m := &readOnlyManager{basePath: tmpDir, serviceType: ServiceTypeSystem}

	// Pass an empty statusMap to simulate a batch miss (label not present).
	svc, err := m.getWithStatus("com.test.sampled", map[string]serviceStatus{}, map[int]int{42: 1})
	if err != nil {
		t.Fatalf("getWithStatus error = %v", err)
	}

	if !detectCalled {
		t.Fatal("heuristic detection was not called on batch miss")
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

	withStubPgrep(t, func(user, program string) ([]int, error) { return []int{10, 20}, nil })

	m := &readOnlyManager{basePath: tmpDir, serviceType: ServiceTypeSystem}
	svc, err := m.getWithStatus("com.test.sampled", map[string]serviceStatus{}, map[int]int{10: 1, 20: 1})
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
// calling heuristic detection.
func TestReadOnly_GetWithStatus_HitSkipsDetectionAndSetsVerified(t *testing.T) {
	tmpDir := t.TempDir()
	writeSystemPlist(t, tmpDir, "com.test.sampled", sampleSystemPlist)

	withStubPgrep(t, func(user, program string) ([]int, error) {
		t.Fatal("heuristic detection should NOT be called when batch map has the label")
		return nil, nil
	})

	statusMap := map[string]serviceStatus{
		"com.test.sampled": {status: StatusRunning, pid: 7777},
	}
	m := &readOnlyManager{basePath: tmpDir, serviceType: ServiceTypeSystem}
	svc, err := m.getWithStatus("com.test.sampled", statusMap, nil)
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
// individual-query path (statusMap == nil) also uses heuristic detection for
// read-only managers; with a nil ppidTable, detection falls back to fetching.
func TestReadOnly_GetWithStatus_NilStatusMapInvokesDetection(t *testing.T) {
	tmpDir := t.TempDir()
	writeSystemPlist(t, tmpDir, "com.test.sampled", sampleSystemPlist)

	withStubPgrep(t, func(user, program string) ([]int, error) { return nil, nil })
	withStubReadAllPPIDs(t, func() (map[int]int, error) { return map[int]int{}, nil })

	m := &readOnlyManager{basePath: tmpDir, serviceType: ServiceTypeSystem}
	svc, err := m.getWithStatus("com.test.sampled", nil, nil)
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
