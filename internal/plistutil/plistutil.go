// Package plistutil provides helpers for reading and normalizing macOS plist
// files. It detects binary-format plists and converts them to XML so downstream
// consumers (diff viewers, text displays) can work with human-readable content.
package plistutil

import (
	"os"
	"os/exec"
)

// Content is a normalized plist payload with metadata describing how it was produced.
type Content struct {
	// Data is the plist body. For binary-format plists this holds the XML
	// representation produced by plutil; for XML plists it is the raw file
	// content. When ConvertFailed is true this falls back to the raw bytes.
	Data string `json:"data"`

	// Format describes the ORIGINAL on-disk format: "xml", "binary", or "unknown".
	Format string `json:"format"`

	// ConvertFailed is true when the file was detected as binary but the
	// binary-to-XML conversion via plutil did not succeed. Data will contain
	// the raw bytes in that case.
	ConvertFailed bool `json:"convertFailed"`
}

// DetectFormat classifies plist bytes as "xml", "binary", or "unknown".
func DetectFormat(data []byte) string {
	if len(data) == 0 {
		return "unknown"
	}
	if len(data) >= 6 && string(data[0:6]) == "bplist" {
		return "binary"
	}
	for i := 0; i < len(data) && i < 100; i++ {
		if data[i] == '<' {
			if len(data) > i+5 && string(data[i:i+5]) == "<?xml" {
				return "xml"
			}
			break
		}
	}
	for i := 0; i < len(data) && i < 100; i++ {
		if data[i] < 32 && data[i] != '\n' && data[i] != '\r' && data[i] != '\t' {
			return "binary"
		}
	}
	return "xml"
}

// NormalizeFromPath reads a plist file and returns a Content describing its
// normalized XML payload. Binary plists are converted to XML via plutil.
// If the binary-to-XML conversion fails the raw bytes are returned with
// ConvertFailed set to true (no error).
//
// Read errors (including os.IsNotExist) are returned as-is so callers can
// distinguish "file missing" from other failures.
func NormalizeFromPath(path string) (*Content, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	format := DetectFormat(data)
	if format != "binary" {
		return &Content{
			Data:   string(data),
			Format: format,
		}, nil
	}

	cmd := exec.Command("plutil", "-convert", "xml1", "-o", "-", path)
	output, cmdErr := cmd.Output()
	if cmdErr != nil {
		return &Content{
			Data:          string(data),
			Format:        "binary",
			ConvertFailed: true,
		}, nil
	}
	return &Content{
		Data:   string(output),
		Format: "binary",
	}, nil
}
