package launchctl

import "testing"

// boolPtr is a small helper for building *bool literals in table tests.
func boolPtr(b bool) *bool { return &b }

func TestParseKeepAlive_ScalarForms(t *testing.T) {
	tests := []struct {
		name        string
		in          any
		wantEnabled bool
		wantMode    string
	}{
		{"nil yields disabled", nil, false, ""},
		{"bool true yields boolean mode", true, true, KeepAliveModeBoolean},
		{"bool false yields disabled boolean", false, false, KeepAliveModeBoolean},
		{"unrecognized type yields disabled", 42, false, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseKeepAlive(tt.in)
			if got.Enabled != tt.wantEnabled {
				t.Errorf("Enabled = %v, want %v", got.Enabled, tt.wantEnabled)
			}
			if got.Mode != tt.wantMode {
				t.Errorf("Mode = %q, want %q", got.Mode, tt.wantMode)
			}
		})
	}
}

// TestParseKeepAlive_Dictionary covers the spec example "KeepAlive dictionary
// round-trips every sub-key".
func TestParseKeepAlive_Dictionary(t *testing.T) {
	in := map[string]any{
		"SuccessfulExit":     false,
		"Crashed":            true,
		"AfterInitialDemand": false,
		"NetworkState":       true,
		"PathState":          map[string]any{"/tmp/flag": true},
		"OtherJobEnabled":    map[string]any{"com.other.job": true},
	}

	got := parseKeepAlive(in)

	if !got.Enabled {
		t.Error("Enabled should be true for a dictionary")
	}
	if got.Mode != KeepAliveModeDictionary {
		t.Errorf("Mode = %q, want %q", got.Mode, KeepAliveModeDictionary)
	}
	assertBoolPtr(t, "SuccessfulExit", got.SuccessfulExit, false)
	assertBoolPtr(t, "Crashed", got.Crashed, true)
	assertBoolPtr(t, "AfterInitialDemand", got.AfterInitialDemand, false)
	assertBoolPtr(t, "NetworkState", got.NetworkState, true)
	if got.PathState["/tmp/flag"] != true {
		t.Errorf("PathState[/tmp/flag] = %v, want true", got.PathState["/tmp/flag"])
	}
	if got.OtherJobEnabled["com.other.job"] != true {
		t.Errorf("OtherJobEnabled[com.other.job] = %v, want true", got.OtherJobEnabled["com.other.job"])
	}
}

// TestParseKeepAlive_IgnoresUnparseableInnerValues confirms the contract that
// type-mismatched inner values are skipped rather than failing the whole parse.
func TestParseKeepAlive_IgnoresUnparseableInnerValues(t *testing.T) {
	in := map[string]any{
		"SuccessfulExit": "not-a-bool",
		"Crashed":        true,
		"PathState":      map[string]any{"/tmp/x": "nope", "/tmp/y": false},
	}

	got := parseKeepAlive(in)

	if got.SuccessfulExit != nil {
		t.Errorf("SuccessfulExit should be nil when the inner value is not a bool, got %v", *got.SuccessfulExit)
	}
	assertBoolPtr(t, "Crashed", got.Crashed, true)
	if _, ok := got.PathState["/tmp/x"]; ok {
		t.Error("PathState should drop the non-bool entry /tmp/x")
	}
	if got.PathState["/tmp/y"] != false {
		t.Errorf("PathState[/tmp/y] = %v, want false", got.PathState["/tmp/y"])
	}
}

func assertBoolPtr(t *testing.T, name string, got *bool, want bool) {
	t.Helper()
	if got == nil {
		t.Errorf("%s is nil, want %v", name, want)
		return
	}
	if *got != want {
		t.Errorf("%s = %v, want %v", name, *got, want)
	}
}
