# Testing

This repository has no cloud CI. The safety net for the modernization
migration is **local Go tooling**: `go test`, `go vet`, `gofmt`,
`staticcheck`. The scripts in `tools/scripts/` wrap these into the
canonical entry points.

## Quick reference

| Command                                 | What it does                                  |
|-----------------------------------------|-----------------------------------------------|
| `tools/scripts/check.sh`                | fmt + vet + staticcheck + test + build (full gate, ~12s) |
| `tools/scripts/check.sh quick`          | fmt + vet + staticcheck + test -short (pre-commit, ~8s)   |
| `tools/scripts/check.sh staticcheck`    | `staticcheck ./...` honouring `staticcheck.conf` |
| `tools/scripts/test.sh`                 | `go test -race -shuffle=on` on allowlist      |
| `tools/scripts/test.sh short`           | `go test -race -shuffle=on -short` on allowlist |
| `tools/scripts/test.sh cover`           | Coverage report with threshold check          |
| `tools/scripts/test.sh cover html`      | Coverage report, opens HTML in browser        |
| `tools/scripts/test.sh bench`           | Benchmarks only                               |
| `tools/scripts/test.sh fuzz FuzzX pkg`  | Fuzz `FuzzX` in `pkg` for `$FUZZTIME`         |
| `tools/scripts/test.sh flake 10`        | Re-run 10× with `-race` to expose flakes      |
| `tools/scripts/test.sh ./pkg`           | Forward args to `go test` for one package     |
| `tools/scripts/watch.sh`                | Re-run tests on save (fswatch/inotify)        |
| `tools/scripts/cover-gaps.sh`           | Show the biggest uncovered functions per package |
| `tools/scripts/build.sh`                | Build the `microM8` binary                    |

## Develop without dread

The intended workflow as of Phase 12:

1. **Keep `watch.sh` running in a side terminal.** It re-runs the
   tests for whatever you save, on save. Install fswatch (macOS) or
   inotify-tools (Linux) for instant feedback.

2. **Before `git commit`, run `tools/scripts/check.sh quick`.**
   That's fmt + vet + staticcheck + tests with `-short` (no build).
   Typical run is ~8s on this tree. The `-short` flag skips a few
   slow tests (network-loop teardown with `client.Close()`'s 1s
   sleep, etc.).

3. **Before `gh pr create`, run the full `tools/scripts/check.sh`.**
   Adds the build and the slow tests. Typical run is ~12s.

4. **Stuck for what to test next?** `tools/scripts/cover-gaps.sh`
   prints the top-3 uncovered functions per package. Pick the one
   with the biggest test gap and write a table-driven case for it.

5. **Investigating a flake?** `tools/scripts/test.sh flake 50`
   re-runs the whole allowlist 50× under `-race`. If a flake survives
   this, it's serious — start there before assuming the fix.

## Linters

`gofmt` and `go vet` are gating in `check.sh`. So is **staticcheck** (Phase 5):

```
go install honnef.co/go/tools/cmd/staticcheck@latest
tools/scripts/check.sh staticcheck
```

The enabled check set lives in [`staticcheck.conf`](staticcheck.conf). The
default `"all"` is in effect with explicit per-check disables, each annotated
with the baseline-hit count and the reason we deferred fixing it. The tree
runs staticcheck-clean — please keep it that way when adding code. To
investigate a specific check, `staticcheck -explain SAXXXX` prints the
rationale. To temporarily ignore one finding without disabling the check
globally, add `//lint:ignore SAXXXX <reason>` directly above the offending
line.

## Allowlists

Two files at the repo root track which packages are part of the green-list:

- `.ci/vet-allowlist.txt` — packages that currently pass `go vet`
- `.ci/test-allowlist.txt` — packages whose tests build and pass

These are append-only as we fix code in later phases. The goal is for
both files to be replaceable with `./...` once Phase 4 lands.

`.ci/coverage-thresholds.txt` (TODO: created with the first real
coverage commit) holds per-package minimum coverage. The
`test.sh cover` runner fails if any package drops below its threshold.

## Conventions

### Test layout

- Tests live next to the code, in `*_test.go` files in the same package.
- Black-box tests (testing only the exported API) go in `package x_test`;
  white-box tests (touching unexported state) go in `package x`. Default
  to black-box.
- Table-driven tests with subtests are the default. Each row gets a
  `t.Run(name, ...)` so failures point at the exact case.
- Use `t.Parallel()` in the outer test and inside each `t.Run` whenever
  the cases don't share mutable state.

```go
func TestEncode(t *testing.T) {
    t.Parallel()
    cases := []struct {
        name string
        in   string
        want string
    }{
        {"empty", "", ""},
        {"ascii", "hello", "encoded(hello)"},
    }
    for _, tc := range cases {
        tc := tc
        t.Run(tc.name, func(t *testing.T) {
            t.Parallel()
            got := Encode(tc.in)
            if got != tc.want {
                t.Errorf("Encode(%q) = %q, want %q", tc.in, got, tc.want)
            }
        })
    }
}
```

### Helpers

The `paleotronic.com/internal/testutil` package provides:

- `Golden(t, name, got)` — compare `got` to `testdata/<name>.golden`.
  Run with `-update` to rewrite golden files.
- `GoldenString(t, name, got)` — string-typed equivalent.
- `NoGoroutineLeaks(t)` — `defer testutil.NoGoroutineLeaks(t)()` to
  assert the test cleans up its goroutines.
- `Eventually(t, timeout, poll, fn, msg)` — assert a condition
  eventually becomes true (replaces tight `time.Sleep` loops).

### Fuzz tests

For any code that parses or decodes external input — file formats,
network frames, disk images, encoded data — write a `FuzzX` alongside
the unit tests:

```go
func FuzzDecode(f *testing.F) {
    f.Add([]byte("seed1"))
    f.Add([]byte{0xff, 0xfe, 0xfd})
    f.Fuzz(func(t *testing.T, data []byte) {
        // Should never panic on arbitrary input.
        _, _ = Decode(data)
    })
}
```

Run with `tools/scripts/test.sh fuzz FuzzDecode ./encoding/base91`.
Seed corpora go in `testdata/fuzz/FuzzDecode/`. Failing inputs
auto-saved by the fuzzer go in the same place and **should be
committed**.

### Benchmarks

For hot paths — CPU emulators, memory access, decoders — add a
`BenchmarkX` next to the unit tests. Use `-benchmem` to track
allocations, and `benchstat` (`go install golang.org/x/perf/cmd/benchstat`)
to compare before/after across refactors.

### Race detector

All test runs use `-race` by default. Don't suppress it. If a test is
expensive enough that `-race` adds intolerable overhead, that's a code
smell — the production code is probably doing work the test doesn't
need to exercise.

### Short mode

Tests that take more than ~500ms wall time should be marked skippable
in `-short` mode:

```go
func TestMyExpensiveThing(t *testing.T) {
    if testing.Short() {
        t.Skip("slow: spins up a full producer + waits for shutdown")
    }
    // ...
}
```

The watch loop and `check.sh quick` pass `-short` so they stay
fast (under ~10s). The pre-PR `check.sh` does NOT pass `-short` and
exercises everything.

### Goroutine leaks

Any test that spawns goroutines (directly or transitively) should
defer `testutil.NoGoroutineLeaks(t)()`. This catches the class of
bugs where cancellation paths fail to drain workers, which is one
of the modernization migration's main concerns (`MonitorNetwork`,
`MusicService`, `RebootService` all leaked goroutines under the
old GOPATH-era code).

### Golden files

Use sparingly — golden files are great for renderers, formatters,
and protocol output, but terrible for anything with non-determinism
(maps, timestamps, goroutine output). When using:

- Store in `testdata/<name>.golden`.
- Update via `go test ./pkg -update`.
- Always commit the resulting files.
- If the format is human-readable (JSON, text), keep it human-readable
  so diffs review cleanly.

## Adding tests to a new package

1. Add the package path to `.ci/test-allowlist.txt`.
2. Write tests until `tools/scripts/test.sh ./your-pkg` passes.
3. If introducing coverage thresholds, add a line to
   `.ci/coverage-thresholds.txt`:
   ```
   ./your-pkg 80.0
   ```
4. Run `tools/scripts/test.sh cover` to confirm.
