#!/usr/bin/env bash
# Run the canonical pre-commit checks locally.
# This is the primary safety net for the modernization migration —
# the repo has no cloud CI, by design.
#
# Usage:
#   tools/scripts/check.sh           # fmt + vet + test + build (race-enabled)
#   tools/scripts/check.sh fmt       # gofmt -l (non-gating until Phase 2)
#   tools/scripts/check.sh vet       # go vet on allowlist
#   tools/scripts/check.sh test      # go test -race on allowlist
#   tools/scripts/check.sh build     # build octalyzer
#
# See also:
#   tools/scripts/test.sh   for deeper test modes (cover, bench, fuzz, flake)
#   tools/scripts/watch.sh  for a file-watcher test loop
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
MODULE_DIR="$REPO_ROOT/GoPath/src/paleotronic.com"

read_allowlist() {
    grep -v '^\s*#' "$1" | grep -v '^\s*$' | tr '\n' ' '
}

VET_PKGS="$(read_allowlist "$REPO_ROOT/.ci/vet-allowlist.txt")"

run_fmt() {
    echo "==> gofmt -l"
    cd "$MODULE_DIR"
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
    echo "==> go vet (allowlist)"
    cd "$MODULE_DIR"
    GOFLAGS=-mod=mod go vet $VET_PKGS
}

run_test() {
    "$SCRIPT_DIR/test.sh" race
}

run_build() {
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
