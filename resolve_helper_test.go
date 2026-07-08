package main

import (
	"context"
	"errors"
	"os"
	"testing"

	"launchpal/internal/privhelper"
)

const (
	testBundlePath    = "/Applications/LaunchPal.app/Contents/MacOS/launchpal-privhelper"
	testProtectedPath = privhelper.ProtectedHelperPath
)

// fakeHash returns a fileHash func backed by a path→digest map; unknown paths
// report a not-exist error (missing/unreadable bundle).
func fakeHash(m map[string]string) func(string) (string, error) {
	return func(p string) (string, error) {
		if h, ok := m[p]; ok {
			return h, nil
		}
		return "", os.ErrNotExist
	}
}

// isVerified returns an isVerified func that reports true only for the given
// path (empty ""→ nothing verified).
func isVerified(verified string) func(string) bool {
	return func(p string) bool { return verified != "" && p == verified }
}

func TestResolveHelperLaunchPath(t *testing.T) {
	cases := []struct {
		name          string
		protectedOK   bool
		pin           string
		hashes        map[string]string
		wantPath      string
		wantErr       error
		wantIntegrity bool
	}{
		{
			// (a) valid protected + empty pin → protected
			name:        "valid protected, empty pin",
			protectedOK: true,
			pin:         "",
			hashes:      map[string]string{testBundlePath: "aaa", testProtectedPath: "aaa"},
			wantPath:    testProtectedPath,
		},
		{
			// (b) valid protected + bundle missing → protected (no error, no DoS)
			name:        "valid protected, bundle missing",
			protectedOK: true,
			pin:         "pinvalue",
			hashes:      map[string]string{testProtectedPath: "aaa"},
			wantPath:    testProtectedPath,
		},
		{
			// (c) tampered bundle (≠ pin) + valid protected → protected (no downgrade)
			name:        "valid protected, tampered bundle mismatching pin",
			protectedOK: true,
			pin:         "pinvalue",
			hashes:      map[string]string{testBundlePath: "tampered", testProtectedPath: "aaa"},
			wantPath:    testProtectedPath,
		},
		{
			// (d) legitimate update: valid protected, bundle==pin, bundle≠protected → bundle
			name:        "legitimate update re-provisions",
			protectedOK: true,
			pin:         "newpin",
			hashes:      map[string]string{testBundlePath: "newpin", testProtectedPath: "oldhash"},
			wantPath:    testBundlePath,
		},
		{
			// pin matches but bundle == protected (already current) → protected
			name:        "matching pin but bundle equals protected",
			protectedOK: true,
			pin:         "samehash",
			hashes:      map[string]string{testBundlePath: "samehash", testProtectedPath: "samehash"},
			wantPath:    testProtectedPath,
		},
		{
			// Empty pin does not bypass a valid protected copy
			name:        "empty pin does not bypass protected",
			protectedOK: true,
			pin:         "",
			hashes:      map[string]string{testBundlePath: "whatever", testProtectedPath: "aaa"},
			wantPath:    testProtectedPath,
		},
		{
			// (e) no protected + non-empty pin, bundle mismatch → integrity error
			name:          "no protected, bundle mismatch pin",
			protectedOK:   false,
			pin:           "pinvalue",
			hashes:        map[string]string{testBundlePath: "tampered"},
			wantIntegrity: true,
		},
		{
			// no protected + non-empty pin + bundle missing → integrity error
			name:          "no protected, non-empty pin, bundle missing",
			protectedOK:   false,
			pin:           "pinvalue",
			hashes:        map[string]string{},
			wantIntegrity: true,
		},
		{
			// first install: no protected + bundle matches pin → bundle
			name:        "no protected, bundle matches pin",
			protectedOK: false,
			pin:         "pinvalue",
			hashes:      map[string]string{testBundlePath: "pinvalue"},
			wantPath:    testBundlePath,
		},
		{
			// dev build: no protected + empty pin → bundle (no hash gate)
			name:        "no protected, empty pin",
			protectedOK: false,
			pin:         "",
			hashes:      map[string]string{testBundlePath: "anything"},
			wantPath:    testBundlePath,
		},
		{
			// dev build with no bundle at all → plain not-found, not integrity
			name:        "no protected, empty pin, bundle missing",
			protectedOK: false,
			pin:         "",
			hashes:      map[string]string{},
			wantErr:     os.ErrNotExist,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			verified := ""
			if tc.protectedOK {
				verified = testProtectedPath
			}
			got, err := resolveHelperLaunchPath(helperResolution{
				bundlePath:    testBundlePath,
				protectedPath: testProtectedPath,
				pin:           tc.pin,
				isVerified:    isVerified(verified),
				fileHash:      fakeHash(tc.hashes),
			})
			switch {
			case tc.wantIntegrity:
				if !errors.Is(err, errHelperIntegrity) {
					t.Fatalf("err = %v, want errHelperIntegrity", err)
				}
			case tc.wantErr != nil:
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("err = %v, want %v", err, tc.wantErr)
				}
			default:
				if err != nil {
					t.Fatalf("unexpected err: %v", err)
				}
				if got != tc.wantPath {
					t.Errorf("path = %q, want %q", got, tc.wantPath)
				}
			}
		})
	}
}

func TestAdminMode_Enable_PassesResolvedPathToLaunch(t *testing.T) {
	client := liveClient(t)
	var gotPath string
	launch := func(ctx context.Context, opts privhelper.LaunchHelperOptions) (*privhelper.Client, error) {
		gotPath = opts.HelperPath
		return client, nil
	}
	a, _, _ := newTestAdminMode(t, launch, nil)
	// The resolved launch path (protected or bundle) must reach LaunchHelper
	// unchanged via the existing LaunchHelperOptions.HelperPath field.
	a.helperPath = func() (string, error) { return testProtectedPath, nil }

	if err := a.Enable(context.Background()); err != nil {
		t.Fatalf("Enable: %v", err)
	}
	if gotPath != testProtectedPath {
		t.Errorf("LaunchHelper HelperPath = %q, want %q", gotPath, testProtectedPath)
	}
}

func TestAdminMode_Enable_IntegrityFailureRefusesLaunch(t *testing.T) {
	launch := func(ctx context.Context, opts privhelper.LaunchHelperOptions) (*privhelper.Client, error) {
		t.Fatal("launchFn (osascript) must not run on integrity failure")
		return nil, nil
	}
	a, sysMgr, _ := newTestAdminMode(t, launch, nil)
	// Override the injected helperPath to report an integrity failure.
	a.helperPath = func() (string, error) { return "", errHelperIntegrity }

	err := a.Enable(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
	s := a.status()
	if s.State != AdminModeDisabled {
		t.Errorf("state = %q, want %q", s.State, AdminModeDisabled)
	}
	if s.Error == nil || *s.Error != "helper_integrity_failed" {
		t.Errorf("error = %v, want helper_integrity_failed", s.Error)
	}
	if sysMgr.set != 0 {
		t.Errorf("SetAdminClient calls = %d, want 0", sysMgr.set)
	}
}
