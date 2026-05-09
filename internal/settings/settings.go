// Package settings persists user-editable application preferences as a
// JSON document at ~/.launchpal/settings.json. Defaults are returned in
// memory until the user explicitly saves; missing or corrupt files fall
// back to Default() so first-run startup is transparent.
package settings

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"launchpal/internal/privhelper"
)

// Settings holds every user-editable preference persisted to disk.
//
// Add fields with care: forward-compatible decoding means an older binary
// reading a newer file silently drops unknown fields; a newer binary
// reading an older file gets defaults for any field absent from the JSON.
type Settings struct {
	UserLogDir   string `json:"userLogDir"`
	SystemLogDir string `json:"systemLogDir"`
}

const (
	defaultUserLogDir   = "~/Library/Logs"
	defaultSystemLogDir = "/Library/Logs"
)

// Path returns the absolute path of the on-disk settings document.
func Path() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".launchpal", "settings.json"), nil
}

// Default returns the in-memory defaults. It performs no filesystem I/O,
// so first-run callers can render the UI without ever touching disk.
func Default() Settings {
	return Settings{
		UserLogDir:   defaultUserLogDir,
		SystemLogDir: defaultSystemLogDir,
	}
}

// Load reads ~/.launchpal/settings.json and returns the parsed Settings.
//
// Recovery contract:
//   - file does not exist → Default(), no error (first-run experience).
//   - file exists but JSON is invalid → Default(), no error, warning logged
//     so the failure is observable in application logs.
//   - file exists with partial fields → defaults filled in for absent
//     fields (forward-compatible decoding).
//
// Filesystem errors other than "not exist" (e.g., EACCES on the parent)
// are returned to the caller.
func Load() (Settings, error) {
	path, err := Path()
	if err != nil {
		return Default(), err
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return Default(), nil
		}
		return Default(), err
	}
	out := Default()
	if err := json.Unmarshal(raw, &out); err != nil {
		log.Printf("settings: invalid JSON at %s, falling back to defaults: %v", path, err)
		return Default(), nil
	}
	return out, nil
}

// Save validates s and persists it atomically to ~/.launchpal/settings.json.
//
// On validation failure no temp file is left on disk and no rename occurs;
// the on-disk settings remain whatever they were before the call. On
// filesystem failure the temp file is best-effort removed.
func Save(s Settings) error {
	if err := Validate(s); err != nil {
		return err
	}
	path, err := Path()
	if err != nil {
		return err
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	body, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	body = append(body, '\n')

	tmp, err := os.CreateTemp(dir, "settings-*.json.tmp")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	cleanup := func() { _ = os.Remove(tmpPath) }
	if _, err := tmp.Write(body); err != nil {
		_ = tmp.Close()
		cleanup()
		return err
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		cleanup()
		return err
	}
	if err := tmp.Close(); err != nil {
		cleanup()
		return err
	}
	if err := os.Chmod(tmpPath, 0644); err != nil {
		cleanup()
		return err
	}
	if err := os.Rename(tmpPath, path); err != nil {
		cleanup()
		return err
	}
	return nil
}

// ValidationError signals a user-correctable Settings problem. The Field
// is exposed separately so the UI can render the error inline next to
// the offending input.
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// shellMetaChars rejects characters that would let a malicious value
// inject a shell command if the path were ever interpolated unquoted.
// We block them at the validator boundary — defense in depth, since the
// real path consumers also quote.
var shellMetaChars = []string{";", "&", "|", "$", "`", "\n", "\r", "\x00"}

func containsShellMeta(s string) bool {
	for _, c := range shellMetaChars {
		if strings.Contains(s, c) {
			return true
		}
	}
	return false
}

// Validate returns nil for an acceptable Settings, or a *ValidationError
// describing the first violated rule. Field-level rules:
//
//   - userLogDir: non-empty, no shell metacharacters, starts with "~/" or "/".
//   - systemLogDir: absolute path under one of privhelper.SystemLogPathPrefixes,
//     with at least one path component beyond the prefix.
//
// The systemLogDir prefix list is sourced from privhelper so a single edit
// updates both the helper handler and this validator.
func Validate(s Settings) error {
	if err := validateUserLogDir(s.UserLogDir); err != nil {
		return err
	}
	if err := validateSystemLogDir(s.SystemLogDir); err != nil {
		return err
	}
	return nil
}

func validateUserLogDir(v string) error {
	if v == "" {
		return &ValidationError{Field: "userLogDir", Message: "must not be empty"}
	}
	if containsShellMeta(v) {
		return &ValidationError{Field: "userLogDir", Message: "contains shell metacharacter"}
	}
	if !strings.HasPrefix(v, "~/") && !strings.HasPrefix(v, "/") {
		return &ValidationError{Field: "userLogDir", Message: "must be tilde-home (~/...) or absolute (/...)"}
	}
	return nil
}

func validateSystemLogDir(v string) error {
	if v == "" {
		return &ValidationError{Field: "systemLogDir", Message: "must not be empty"}
	}
	if containsShellMeta(v) {
		return &ValidationError{Field: "systemLogDir", Message: "contains shell metacharacter"}
	}
	if !filepath.IsAbs(v) {
		return &ValidationError{Field: "systemLogDir", Message: "must be absolute (no tilde, no relative paths)"}
	}
	// Compare against clean+"/" so a bare allowlist root (e.g.
	// "/Library/Logs") still satisfies HasPrefix — the New Service modal
	// always interpolates "<systemLogDir>/<label>/<stream>.log", so the
	// label provides the depth the helper's per-file allowlist requires.
	cleanWithSlash := filepath.Clean(v) + "/"
	for _, prefix := range privhelper.SystemLogPathPrefixes {
		if strings.HasPrefix(cleanWithSlash, prefix) {
			return nil
		}
	}
	return &ValidationError{
		Field:   "systemLogDir",
		Message: "must start with one of: " + strings.Join(privhelper.SystemLogPathPrefixes, ", "),
	}
}

// IsValidationError is a convenience for callers that don't want to
// import errors.As just to type-assert.
func IsValidationError(err error) bool {
	var v *ValidationError
	return errors.As(err, &v)
}
