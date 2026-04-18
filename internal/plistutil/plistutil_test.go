package plistutil

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDetectFormat(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want string
	}{
		{"empty", []byte{}, "unknown"},
		{"xml prolog", []byte(`<?xml version="1.0"?>`), "xml"},
		{"xml with leading whitespace", []byte("  \n<?xml version=\"1.0\"?>"), "xml"},
		{"binary prefix", append([]byte("bplist00"), 0x01, 0x02), "binary"},
		{"printable ascii without xml prolog", []byte("<plist><dict/></plist>"), "xml"},
		{"control bytes indicate binary", []byte{0x62, 0x70, 0x00, 0x01, 0x02, 0x03}, "binary"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := DetectFormat(tt.data); got != tt.want {
				t.Errorf("DetectFormat() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestNormalizeFromPath_XMLPlist(t *testing.T) {
	content := `<?xml version="1.0" encoding="UTF-8"?>
<plist version="1.0">
<dict><key>Label</key><string>com.test</string></dict>
</plist>`

	path := filepath.Join(t.TempDir(), "xml.plist")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}

	got, err := NormalizeFromPath(path)
	if err != nil {
		t.Fatalf("NormalizeFromPath error = %v", err)
	}
	if got.Data != content {
		t.Errorf("Data mismatch\ngot:  %q\nwant: %q", got.Data, content)
	}
	if got.Format != "xml" {
		t.Errorf("Format = %q, want %q", got.Format, "xml")
	}
	if got.ConvertFailed {
		t.Errorf("ConvertFailed = true, want false")
	}
}

func TestNormalizeFromPath_BinaryPlistConvertedToXML(t *testing.T) {
	// Use plutil to produce a genuine binary plist, then verify Normalize converts it back.
	xmlSource := `<?xml version="1.0" encoding="UTF-8"?>
<plist version="1.0">
<dict><key>Label</key><string>com.test.binary</string></dict>
</plist>`

	xmlPath := filepath.Join(t.TempDir(), "source.plist")
	if err := os.WriteFile(xmlPath, []byte(xmlSource), 0644); err != nil {
		t.Fatalf("write xml: %v", err)
	}

	binaryPath := filepath.Join(t.TempDir(), "binary.plist")
	if err := convertXMLToBinaryForTest(xmlPath, binaryPath); err != nil {
		t.Skipf("plutil not available or conversion failed: %v", err)
	}

	got, err := NormalizeFromPath(binaryPath)
	if err != nil {
		t.Fatalf("NormalizeFromPath error = %v", err)
	}
	if got.Format != "binary" {
		t.Errorf("Format = %q, want %q", got.Format, "binary")
	}
	if got.ConvertFailed {
		t.Errorf("ConvertFailed = true, want false")
	}
	if !strings.Contains(got.Data, "com.test.binary") {
		t.Errorf("Data does not contain expected label; got: %q", got.Data)
	}
	if !strings.Contains(got.Data, "<?xml") {
		t.Errorf("converted data does not look like XML plist; got prefix: %q", firstN(got.Data, 80))
	}
}

func TestNormalizeFromPath_ConversionFailureFallsBackToRaw(t *testing.T) {
	// Corrupted binary plist: bplist prefix but invalid trailer/body.
	raw := append([]byte("bplist00"), 0xFF, 0xFE, 0xFD, 0xFC, 0xFB, 0xFA)
	path := filepath.Join(t.TempDir(), "corrupt.plist")
	if err := os.WriteFile(path, raw, 0644); err != nil {
		t.Fatalf("write: %v", err)
	}

	got, err := NormalizeFromPath(path)
	if err != nil {
		t.Fatalf("NormalizeFromPath error = %v (expected fallback, not error)", err)
	}
	if got.Format != "binary" {
		t.Errorf("Format = %q, want %q", got.Format, "binary")
	}
	if !got.ConvertFailed {
		t.Errorf("ConvertFailed = false, want true")
	}
	if got.Data != string(raw) {
		t.Errorf("fallback Data does not match raw bytes")
	}
}

func TestNormalizeFromPath_MissingFileReturnsError(t *testing.T) {
	_, err := NormalizeFromPath(filepath.Join(t.TempDir(), "does-not-exist.plist"))
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
	if !os.IsNotExist(err) {
		t.Errorf("error = %v, want os.IsNotExist", err)
	}
}

func firstN(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}
