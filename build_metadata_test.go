package main

import (
	"os"
	"testing"

	"howett.net/plist"
)

// wantMinimumSystemVersion is the single source of truth for the minimum macOS
// version LaunchPal declares. It must equal the LC_BUILD_VERSION minos value
// the Go toolchain writes into the produced binaries: a lower declaration lets
// Launch Services start the app only for dyld to reject the load with an
// unexplained failure, and a higher one excludes systems that could in fact run
// it. Re-verify against a freshly built binary whenever the Go toolchain is
// upgraded.
const wantMinimumSystemVersion = "13.0.0"

// infoPlistPaths are the app bundle declarations that must carry the version
// above. Both are tracked in version control; Wails only generates its template
// when the file is absent, so a build never overwrites them.
var infoPlistPaths = []string{
	"build/darwin/Info.plist",
	"build/darwin/Info.dev.plist",
}

// TestInfoPlistMinimumSystemVersion guards the declared minimum macOS version
// against silent drift — the exact failure that let the Wails template default
// of 10.13.0 sit unnoticed for years. The files are parsed as plists rather
// than text-matched; the Go template directives they embed between XML elements
// do not disturb the parser. Every way of not reaching a comparison (unreadable
// file, unparseable content, missing key, non-string value) is a failure, never
// a skip.
func TestInfoPlistMinimumSystemVersion(t *testing.T) {
	const key = "LSMinimumSystemVersion"

	for _, path := range infoPlistPaths {
		t.Run(path, func(t *testing.T) {
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("%s: read failed: %v", path, err)
			}

			var declared map[string]any
			if _, err := plist.Unmarshal(data, &declared); err != nil {
				t.Fatalf("%s: parse failed: %v", path, err)
			}

			value, ok := declared[key]
			if !ok {
				t.Fatalf("%s: %s key is absent, want %q", path, key, wantMinimumSystemVersion)
			}

			got, ok := value.(string)
			if !ok {
				t.Fatalf("%s: %s is %T (%v), want the string %q", path, key, value, value, wantMinimumSystemVersion)
			}

			if got != wantMinimumSystemVersion {
				t.Errorf("%s: %s is %q, want %q", path, key, got, wantMinimumSystemVersion)
			}
		})
	}
}
