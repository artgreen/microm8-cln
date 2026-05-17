package files

import (
	"testing"
)

// TestPackageFileProvider_Exists_FilenameLookup pins the Phase 5 SA4014
// fix in Exists(p, f). The original code had two `else if` branches
// gated on the same `f == ""` condition; the second branch was
// unreachable. The fix changed the second branch's condition to
// `f != ""`, which is what the body clearly meant (file-inside-package
// lookup).
//
// Test contract:
//   - empty filename + empty path  → "root" exists
//   - filename empty + non-empty path → package-as-directory exists iff
//     the package is registered
//   - filename present + non-empty path → file-inside-package lookup
//
// Before the fix, the third case was UNREACHABLE — Exists would always
// return false for file-inside-package because the dup'd `f == ""`
// condition never matched.
func TestPackageFileProvider_Exists_FilenameLookup(t *testing.T) {
	t.Parallel()
	pfp := NewPackageFileProvider("", 0)
	pfp.pack = map[string]*Package{
		"games": {
			Name: "games",
			Content: []PackageDirEntry{
				{Name: "frogger.bin"},
				{Name: "pacman.bin"},
			},
		},
	}

	cases := []struct {
		name string
		p    string
		f    string
		want bool
	}{
		{"root path empty filename", "/", "", true},
		{"empty path empty filename", "", "", true},
		{"package dir exists", "games", "", true},
		{"unknown package dir", "missing", "", false},
		// THE pinned branch — file-inside-package lookup. Pre-fix this
		// hit the dup'd `f == ""` condition and unconditionally
		// returned false, breaking every file-in-pak load.
		{"file in package exists", "games", "frogger.bin", true},
		{"file in package missing", "games", "tetris.bin", false},
		{"file in unknown package", "missing", "frogger.bin", false},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := pfp.Exists(tc.p, tc.f)
			if got != tc.want {
				t.Errorf("Exists(%q, %q) = %v, want %v", tc.p, tc.f, got, tc.want)
			}
		})
	}
}
