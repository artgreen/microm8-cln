#!/usr/bin/env bash
# Run the local pre-flight checks that CI also runs.
# Mirrors .github/workflows/ci.yml so contributors can reproduce CI locally.
#
# Usage:
#   tools/scripts/check.sh           # run all checks
#   tools/scripts/check.sh fmt       # gofmt check only
#   tools/scripts/check.sh vet       # go vet on allowlist only
#   tools/scripts/check.sh test      # go test on allowlist only
#   tools/scripts/check.sh build     # build octalyzer only
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
MODULE_DIR="$REPO_ROOT/GoPath/src/paleotronic.com"

# Read allowlist file into space-separated package paths; skip comments and blanks.
read_allowlist() {
    grep -v '^\s*#' "$1" | grep -v '^\s*$' | tr '\n' ' '
}

VET_PKGS="$(read_allowlist "$REPO_ROOT/.ci/vet-allowlist.txt")"
TEST_PKGS="$(read_allowlist "$REPO_ROOT/.ci/test-allowlist.txt")"

run_fmt() {
    echo "==> gofmt -l (allowed to fail until Phase 2)"
    cd "$MODULE_DIR"
    local unformatted
    unformatted="$(gofmt -l . | grep -v '^vendor/' || true)"
    if [ -n "$unformatted" ]; then
        echo "Unformatted files:"
        echo "$unformatted"
        echo "(run 'gofmt -w .' to fix; not gating yet)"
    else
        echo "All formatted."
    fi
}

run_vet() {
    echo "==> go vet (allowlist)"
    cd "$MODULE_DIR"
    GOFLAGS=-mod=mod go vet $VET_PKGS
}

run_test() {
    echo "==> go test (allowlist)"
    cd "$MODULE_DIR"
    GOFLAGS=-mod=mod go test $TEST_PKGS
}

run_build() {
    echo "==> go build ./octalyzer"
    "$SCRIPT_DIR/build.sh"
}

case "${1:-all}" in
    fmt)   run_fmt ;;
    vet)   run_vet ;;
    test)  run_test ;;
    build) run_build ;;
    all)
        run_fmt
        run_vet
        run_test
        run_build
        ;;
    *)
        echo "Usage: $0 [fmt|vet|test|build|all]" >&2
        exit 2
        ;;
esac
