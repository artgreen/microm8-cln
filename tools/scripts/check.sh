#!/usr/bin/env bash
# Run the canonical pre-commit checks locally.
# This is the primary safety net for the modernization migration —
# the repo has no cloud CI, by design.
#
# Usage:
#   tools/scripts/check.sh             # fmt + vet + staticcheck + test + build (race-enabled)
#   tools/scripts/check.sh quick       # fmt + vet + staticcheck + tests with -short (no build)
#   tools/scripts/check.sh fmt         # gofmt -l (non-gating until Phase 2)
#   tools/scripts/check.sh vet         # go vet on allowlist
#   tools/scripts/check.sh staticcheck # honnef.co/go/tools/cmd/staticcheck on ./...
#   tools/scripts/check.sh test        # go test -race on allowlist
#   tools/scripts/check.sh build       # build octalyzer
#
# The "develop without dread" workflow:
#
#   tools/scripts/watch.sh             # auto-reruns tests on save (instant feedback)
#   tools/scripts/check.sh quick       # before committing (≤5s typically)
#   tools/scripts/check.sh             # before opening a PR (full gate)
#
# The staticcheck policy lives in staticcheck.conf at the repo root. The
# enabled set finds real bugs (SA*, plus some ST/S); the noisier style checks
# are explicitly disabled there with rationale. Run `staticcheck -explain SAXXX`
# to read about a check.
#
# See also:
#   tools/scripts/test.sh   for deeper test modes (cover, bench, fuzz, flake)
#   tools/scripts/watch.sh  for a file-watcher test loop
#   tools/scripts/cover-gaps.sh  for "where should I add the next test?"
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

# Vet checks we suppress, with rationale:
#   -stdmethods=false  — the emulator has memory-bus methods named
#                        ReadByte/WriteByte that take address arguments,
#                        which collide stylistically with io.ByteReader /
#                        io.ByteWriter signatures but are semantically
#                        unrelated. The names are clearer than alternatives
#                        like ReadByteAt; suppress this check rather than
#                        rename through ~30 call sites.
#   -unsafeptr=false   — gl/ contains generated OpenGL bindings (via Glow)
#                        that use unsafe.Pointer for CGO interop. The
#                        warnings are inherent to the binding pattern.
#   -cgocall=false     — same: gl/ passes Go types with embedded pointers
#                        to C across the boundary, which is the intended
#                        OpenGL binding contract.
VET_FLAGS="-stdmethods=false -unsafeptr=false -cgocall=false"

run_fmt() {
    echo "==> gofmt -l"
    cd "$REPO_ROOT"
    local unformatted
    unformatted="$(gofmt -l . || true)"
    if [ -z "$unformatted" ]; then
        echo "All files formatted."
        return 0
    fi
    local count
    count=$(echo "$unformatted" | wc -l | tr -d ' ')
    echo "FAIL: $count files unformatted. Run 'gofmt -w .' to fix:" >&2
    echo "$unformatted" >&2
    return 1
}

run_vet() {
    echo "==> go vet ./..."
    cd "$REPO_ROOT"
    GOFLAGS=-mod=mod go vet $VET_FLAGS ./...
}

# Locate staticcheck. We prefer the binary on PATH but fall back to GOPATH/bin
# because `go install` puts it there and that directory is rarely on PATH for
# casual contributors. If neither is present, print a one-line install hint.
find_staticcheck() {
    if command -v staticcheck >/dev/null 2>&1; then
        command -v staticcheck
        return 0
    fi
    local gopath_bin
    gopath_bin="$(go env GOPATH)/bin/staticcheck"
    if [ -x "$gopath_bin" ]; then
        echo "$gopath_bin"
        return 0
    fi
    return 1
}

run_staticcheck() {
    echo "==> staticcheck ./..."
    cd "$REPO_ROOT"
    local bin
    if ! bin="$(find_staticcheck)"; then
        echo "staticcheck not found. Install it with:" >&2
        echo "    go install honnef.co/go/tools/cmd/staticcheck@latest" >&2
        echo "and ensure \$(go env GOPATH)/bin is on \$PATH (or rerun via tools/scripts/check.sh)." >&2
        return 1
    fi
    # The check set is driven by ./staticcheck.conf, not flags here.
    "$bin" ./...
}

run_test() {
    "$SCRIPT_DIR/test.sh" race
}

run_test_short() {
    "$SCRIPT_DIR/test.sh" short
}

run_build() {
    "$SCRIPT_DIR/build.sh"
}

case "${1:-all}" in
    fmt)         run_fmt ;;
    vet)         run_vet ;;
    staticcheck) run_staticcheck ;;
    test)        run_test ;;
    build)       run_build ;;
    quick)
        # Pre-commit workhorse: fmt + vet + staticcheck + tests in -short
        # mode (skips fuzz seed-replay, network/lifecycle loops that sleep).
        # Skips the build because go test compiles every package's tests
        # already — a clean test run gives high confidence the binary
        # builds too.
        run_fmt
        run_vet
        run_staticcheck
        run_test_short
        ;;
    all)
        run_fmt
        run_vet
        run_staticcheck
        run_test
        run_build
        ;;
    *)
        echo "Usage: $0 [fmt|vet|staticcheck|test|build|quick|all]" >&2
        exit 2
        ;;
esac
