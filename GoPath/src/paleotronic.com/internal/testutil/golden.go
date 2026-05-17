// Package testutil provides shared test helpers for the paleotronic.com module.
//
// Helpers here are deliberately small and focused. The bias is toward
// the standard library — `testing.T`, `testing.TB`, `t.Helper()`,
// `t.Cleanup()`, `t.TempDir()` — with thin wrappers only where the
// stdlib idiom needs a few more lines of plumbing in every test
// (golden files, dual-time tolerance comparisons, etc.).
package testutil

import (
	"bytes"
	"flag"
	"os"
	"path/filepath"
	"testing"
)

// updateGolden controls whether golden file comparisons should rewrite
// the .golden file on disk instead of asserting equality. Enable with:
//
//	go test ./... -update
//
// The flag is registered package-wide so any test that imports testutil
// gets it for free.
var updateGolden = flag.Bool("update", false, "rewrite .golden files instead of comparing")

// Golden compares got to the contents of testdata/<name>.golden.
//
// When the -update flag is passed to `go test`, the golden file is rewritten
// instead and the test reports as passing. This is the standard pattern
// from cmd/gofmt and the Go toolchain itself.
//
// The golden file is resolved relative to the test's working directory,
// which Go sets to the test's package directory.
func Golden(tb testing.TB, name string, got []byte) {
	tb.Helper()
	path := filepath.Join("testdata", name+".golden")

	if *updateGolden {
		if err := os.MkdirAll("testdata", 0o755); err != nil {
			tb.Fatalf("testutil.Golden: mkdir testdata: %v", err)
		}
		if err := os.WriteFile(path, got, 0o644); err != nil {
			tb.Fatalf("testutil.Golden: write %s: %v", path, err)
		}
		tb.Logf("testutil.Golden: rewrote %s (%d bytes)", path, len(got))
		return
	}

	want, err := os.ReadFile(path)
	if err != nil {
		tb.Fatalf("testutil.Golden: read %s: %v (run with -update to create)", path, err)
	}
	if !bytes.Equal(got, want) {
		tb.Errorf("testutil.Golden: %s mismatch\n--- want (%d bytes) ---\n%s\n--- got (%d bytes) ---\n%s",
			path, len(want), want, len(got), got)
	}
}

// GoldenString is the string-typed flavor of Golden.
func GoldenString(tb testing.TB, name, got string) {
	tb.Helper()
	Golden(tb, name, []byte(got))
}
