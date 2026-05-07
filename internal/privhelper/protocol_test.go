package privhelper

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestRequest_RoundTrip(t *testing.T) {
	original := Request{
		ID:     42,
		Method: MethodBootstrap,
		Params: json.RawMessage(`{"plistPath":"/Library/LaunchDaemons/com.example.plist"}`),
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded Request
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if decoded.ID != original.ID {
		t.Errorf("ID = %d, want %d", decoded.ID, original.ID)
	}
	if decoded.Method != original.Method {
		t.Errorf("Method = %q, want %q", decoded.Method, original.Method)
	}

	var params BootstrapParams
	if err := json.Unmarshal(decoded.Params, &params); err != nil {
		t.Fatalf("unmarshal params: %v", err)
	}
	if params.PlistPath != "/Library/LaunchDaemons/com.example.plist" {
		t.Errorf("PlistPath = %q", params.PlistPath)
	}
}

func TestResponse_ResultRoundTrip(t *testing.T) {
	resp, err := NewResultResponse(5, PingResult{Pong: true})
	if err != nil {
		t.Fatalf("NewResultResponse: %v", err)
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	// Verify the wire form has the expected shape.
	if !strings.Contains(string(data), `"id":5`) {
		t.Errorf("missing id: %s", string(data))
	}
	if !strings.Contains(string(data), `"pong":true`) {
		t.Errorf("missing pong: %s", string(data))
	}
	if strings.Contains(string(data), `"error"`) {
		t.Errorf("unexpected error field: %s", string(data))
	}
}

func TestResponse_ErrorRoundTrip(t *testing.T) {
	id := int64(7)
	resp := NewErrorResponse(&id, ErrCodeInvalidParams, "bad path")

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded Response
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if decoded.ID == nil || *decoded.ID != 7 {
		t.Errorf("ID = %v, want 7", decoded.ID)
	}
	if decoded.Error == nil {
		t.Fatal("Error is nil")
	}
	if decoded.Error.Code != ErrCodeInvalidParams {
		t.Errorf("Code = %q", decoded.Error.Code)
	}
	if decoded.Error.Message != "bad path" {
		t.Errorf("Message = %q", decoded.Error.Message)
	}
	if len(decoded.Result) != 0 {
		t.Errorf("Result should be empty: %s", string(decoded.Result))
	}
}

func TestResponse_NilIDForMalformedRequest(t *testing.T) {
	resp := NewErrorResponse(nil, ErrCodeInvalidRequest, "not valid JSON")

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	if !strings.Contains(string(data), `"id":null`) {
		t.Errorf("expected id:null in %s", string(data))
	}
}

func TestRPCError_ErrorInterface(t *testing.T) {
	var err error = &RPCError{Code: ErrCodeLaunchctlFailed, Message: "exit 1"}
	got := err.Error()
	if !strings.Contains(got, ErrCodeLaunchctlFailed) || !strings.Contains(got, "exit 1") {
		t.Errorf("Error() = %q", got)
	}
}

func TestTruncateLogParams_RoundTrip(t *testing.T) {
	original := Request{
		ID:     99,
		Method: MethodTruncateLog,
		Params: json.RawMessage(`{"path":"/var/log/myapp/out.log"}`),
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded Request
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if decoded.Method != MethodTruncateLog {
		t.Errorf("Method = %q, want %q", decoded.Method, MethodTruncateLog)
	}
	var params TruncateLogParams
	if err := json.Unmarshal(decoded.Params, &params); err != nil {
		t.Fatalf("unmarshal params: %v", err)
	}
	if params.Path != "/var/log/myapp/out.log" {
		t.Errorf("Path = %q", params.Path)
	}
}

func TestAllMethods_Coverage(t *testing.T) {
	want := []string{
		MethodPing,
		MethodBootstrap,
		MethodBootout,
		MethodKickstart,
		MethodWritePlist,
		MethodDeletePlist,
		MethodEnsureLogAccess,
		MethodTruncateLog,
		MethodShutdown,
	}
	if len(AllMethods) != len(want) {
		t.Fatalf("AllMethods has %d entries, want %d", len(AllMethods), len(want))
	}
	set := make(map[string]bool, len(AllMethods))
	for _, m := range AllMethods {
		set[m] = true
	}
	for _, m := range want {
		if !set[m] {
			t.Errorf("AllMethods missing %q", m)
		}
	}
}

func TestParseMalformedJSON(t *testing.T) {
	var req Request
	err := json.Unmarshal([]byte("{not valid json"), &req)
	if err == nil {
		t.Fatal("expected error for malformed JSON")
	}
}

func TestErrorCodes_Stable(t *testing.T) {
	// These literals are part of the wire protocol; snapshot them so an
	// accidental rename flips a test instead of silently breaking the UI.
	cases := map[string]string{
		ErrCodeInvalidRequest:   "invalid_request",
		ErrCodeUnknownMethod:    "unknown_method",
		ErrCodeInvalidParams:    "invalid_params",
		ErrCodePermissionDenied: "permission_denied",
		ErrCodeNotFound:         "not_found",
		ErrCodeLaunchctlFailed:  "launchctl_failed",
		ErrCodeIOError:          "io_error",
		ErrCodeInternalError:    "internal_error",
	}
	for got, want := range cases {
		if got != want {
			t.Errorf("error code drift: got %q, want %q", got, want)
		}
	}
}
