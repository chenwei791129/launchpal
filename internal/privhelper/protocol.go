// Package privhelper defines the RPC protocol shared between the LaunchPal
// GUI process (user-privileged client) and the launchpal-privhelper binary
// (root-privileged server) exchanged as newline-delimited JSON messages over
// a Unix domain socket.
package privhelper

import "encoding/json"

// Supported RPC methods. These are the exhaustive set of method names the
// helper accepts; anything outside this list returns ErrCodeUnknownMethod.
const (
	MethodPing            = "Ping"
	MethodBootstrap       = "Bootstrap"
	MethodBootout         = "Bootout"
	MethodKickstart       = "Kickstart"
	MethodWritePlist      = "WritePlist"
	MethodDeletePlist     = "DeletePlist"
	MethodEnsureLogAccess = "EnsureLogAccess"
	MethodTruncateLog     = "TruncateLog"
	MethodDeleteLogPaths  = "DeleteLogPaths"
	MethodShutdown        = "Shutdown"
)

// AllMethods is the complete set of supported methods, useful for validation
// and documentation.
var AllMethods = []string{
	MethodPing,
	MethodBootstrap,
	MethodBootout,
	MethodKickstart,
	MethodWritePlist,
	MethodDeletePlist,
	MethodEnsureLogAccess,
	MethodTruncateLog,
	MethodDeleteLogPaths,
	MethodShutdown,
}

// Error codes returned in RPC responses. The taxonomy is fixed; every new
// error scenario must map onto one of these codes.
const (
	ErrCodeInvalidRequest   = "invalid_request"
	ErrCodeUnknownMethod    = "unknown_method"
	ErrCodeInvalidParams    = "invalid_params"
	ErrCodePermissionDenied = "permission_denied"
	ErrCodeNotFound         = "not_found"
	ErrCodeLaunchctlFailed  = "launchctl_failed"
	ErrCodeIOError          = "io_error"
	ErrCodeInternalError    = "internal_error"
)

// Request is a single RPC request framed as a JSON line.
type Request struct {
	ID     int64           `json:"id"`
	Method string          `json:"method"`
	Params json.RawMessage `json:"params,omitempty"`
}

// Response is a single RPC response framed as a JSON line. Exactly one of
// Result or Error should be set; both zero-valued responses are invalid.
type Response struct {
	ID     *int64          `json:"id"`
	Result json.RawMessage `json:"result,omitempty"`
	Error  *RPCError       `json:"error,omitempty"`
}

// RPCError is the error payload returned inside a Response.
type RPCError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Error implements the error interface so RPCError can flow through Go's
// standard error handling.
func (e *RPCError) Error() string {
	if e == nil {
		return "<nil>"
	}
	return e.Code + ": " + e.Message
}

// PingResult is the payload of a successful Ping.
type PingResult struct {
	Pong bool `json:"pong"`
}

// BootstrapParams are the parameters for MethodBootstrap.
type BootstrapParams struct {
	PlistPath string `json:"plistPath"`
}

// BootoutParams are the parameters for MethodBootout.
type BootoutParams struct {
	Label string `json:"label"`
}

// KickstartParams are the parameters for MethodKickstart.
type KickstartParams struct {
	Label string `json:"label"`
}

// WritePlistParams are the parameters for MethodWritePlist. Data is the plist
// contents encoded as base64; the helper decodes, writes atomically, sets
// ownership, and chowns any backup it creates to the launching user.
type WritePlistParams struct {
	PlistPath string `json:"plistPath"`
	Data      string `json:"data"`
}

// DeletePlistParams are the parameters for MethodDeletePlist.
type DeletePlistParams struct {
	PlistPath string `json:"plistPath"`
}

// EnsureLogAccessParams are the parameters for MethodEnsureLogAccess. Each
// path is a StandardOutPath or StandardErrorPath pulled from the plist the
// caller just wrote; the helper ensures the parent directory is traversable
// (0755) and the log file exists with 0644 so the launching user can tail it
// without Full Disk Access.
type EnsureLogAccessParams struct {
	Paths []string `json:"paths"`
}

// TruncateLogParams are the parameters for MethodTruncateLog. Path must be an
// absolute log file path under the same allowlist used by EnsureLogAccess.
// The helper truncates the existing file to 0 bytes preserving inode, owner,
// group, and mode; it does NOT create a missing file.
type TruncateLogParams struct {
	Path string `json:"path"`
}

// DeleteLogPathsParams are the parameters for MethodDeleteLogPaths. Each
// path is an absolute log file path under the same allowlist used by
// EnsureLogAccess. The helper deletes each file and then tries to remove
// the parent directory if it is empty.
type DeleteLogPathsParams struct {
	Paths []string `json:"paths"`
}

// DeleteLogPathsResult is the success payload for MethodDeleteLogPaths.
// Errors holds one entry per path the helper could not process; the slice
// is empty on a fully successful run. A non-empty Errors with no top-level
// RPCError is a valid partial-success response.
type DeleteLogPathsResult struct {
	Errors []string `json:"errors"`
}

// OKResult is the empty success payload used by methods that have no return
// value beyond "it worked" (Bootstrap, Bootout, Kickstart, WritePlist,
// DeletePlist, Shutdown).
type OKResult struct {
	OK bool `json:"ok"`
}

// NewErrorResponse builds a Response carrying an error with the given request
// id. A nil id signals the helper couldn't parse the request.
func NewErrorResponse(id *int64, code, message string) Response {
	return Response{
		ID:    id,
		Error: &RPCError{Code: code, Message: message},
	}
}

// NewResultResponse builds a Response with the given result payload.
func NewResultResponse(id int64, result any) (Response, error) {
	data, err := json.Marshal(result)
	if err != nil {
		return Response{}, err
	}
	rid := id
	return Response{ID: &rid, Result: data}, nil
}
