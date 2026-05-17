#!/usr/bin/env bash
# Run the test suite in various modes.
#
# Usage:
#   tools/scripts/test.sh                    # quick: -race on the test allowlist
#   tools/scripts/test.sh unit               # all unit tests on the allowlist
#   tools/scripts/test.sh race               # -race on the allowlist
#   tools/scripts/test.sh cover              # coverage report
#   tools/scripts/test.sh cover html         # coverage report + open HTML
#   tools/scripts/test.sh bench              # benchmarks only (-run=^$ -bench=.)
#   tools/scripts/test.sh fuzz <fn> [pkg]    # fuzz <fn> for $FUZZTIME (default 30s)
#   tools/scripts/test.sh flake [N]          # rerun -race -count=N (default 10) to expose flakes
#   tools/scripts/test.sh all                # everything: race + cover. No fuzz, no bench.
#   tools/scripts/test.sh <pkg> [-run X]     # forward args to `go test` for a single package
#
# Environment:
#   FUZZTIME=30s    duration for `fuzz` subcommand
#   COVER_THRESHOLD path to .ci/coverage-thresholds.txt (override for testing)
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

# Read allowlist file into space-separated package paths; skip comments and blanks.
read_allowlist() {
    grep -v '^\s*#' "$1" | grep -v '^\s*$' | tr '\n' ' '
}

TEST_PKGS="$(read_allowlist "$REPO_ROOT/.ci/test-allowlist.txt")"
COVER_THRESHOLD="${COVER_THRESHOLD:-$REPO_ROOT/.ci/coverage-thresholds.txt}"

cd "$REPO_ROOT"
export GOFLAGS="${GOFLAGS:--mod=mod}"

run_unit() {
    echo "==> go test (allowlist)"
    go test -shuffle=on $TEST_PKGS
}

run_race() {
    echo "==> go test -race (allowlist)"
    go test -race -shuffle=on $TEST_PKGS
}

run_cover() {
    echo "==> go test -coverprofile (allowlist)"
    local profile="$REPO_ROOT/coverage.out"
    go test -race -coverprofile="$profile" -covermode=atomic $TEST_PKGS

    echo
    echo "==> Coverage by package:"
    go tool cover -func="$profile" | tail -n +1

    if [ -f "$COVER_THRESHOLD" ]; then
        check_thresholds "$profile" "$COVER_THRESHOLD"
    fi

    if [ "${1:-}" = "html" ]; then
        go tool cover -html="$profile" -o "$REPO_ROOT/coverage.html"
        echo "==> Wrote $REPO_ROOT/coverage.html"
        # macOS-friendly open; ignored elsewhere
        command -v open >/dev/null 2>&1 && open "$REPO_ROOT/coverage.html"
    fi
}

# Parse coverage.out + thresholds file; fail if any package drops below target.
check_thresholds() {
    local profile="$1" thresholds="$2"
    echo
    echo "==> Coverage threshold check:"
    local fails=0
    # `go tool cover -func` last line is total; we want per-file rolled up to package
    # Simpler: derive package coverage by re-running `go test -cover` per package
    while IFS= read -r line; do
        # Skip comments and blanks
        case "$line" in ''|\#*) continue ;; esac
        local pkg=$(echo "$line" | awk '{print $1}')
        local target=$(echo "$line" | awk '{print $2}')
        # Get actual coverage for this package
        local actual
        actual=$(go test -cover -count=1 "$pkg" 2>&1 | awk '/coverage:/ {gsub("%",""); print $5; exit}' || echo "0.0")
        if [ -z "$actual" ] || [ "$actual" = "0.0" ]; then
            # Package has no tests or no statements — treat as 0%
            actual="0.0"
        fi
        local pass="ok"
        if awk -v a="$actual" -v t="$target" 'BEGIN{exit !(a+0 < t+0)}'; then
            pass="FAIL"
            fails=$((fails+1))
        fi
        printf "  %-40s actual=%-7s target=%-7s %s\n" "$pkg" "$actual" "$target" "$pass"
    done < "$thresholds"
    if [ "$fails" -gt 0 ]; then
        echo "==> $fails package(s) below coverage threshold."
        return 1
    fi
}

run_bench() {
    echo "==> go test -bench (allowlist)"
    go test -run='^$' -bench=. -benchmem -benchtime=2s $TEST_PKGS
}

run_fuzz() {
    local fn="${1:-}" pkg="${2:-}"
    if [ -z "$fn" ]; then
        echo "Usage: $0 fuzz <FuzzFuncName> [pkg]" >&2
        return 2
    fi
    # Default fuzz package is the first one in the allowlist
    if [ -z "$pkg" ]; then
        pkg=$(echo "$TEST_PKGS" | awk '{print $1}')
    fi
    local time="${FUZZTIME:-30s}"
    echo "==> go test -fuzz=$fn -fuzztime=$time $pkg"
    go test -fuzz="$fn" -fuzztime="$time" -run='^$' "$pkg"
}

run_flake() {
    local n="${1:-10}"
    echo "==> Running tests $n times to expose flakes"
    go test -race -count="$n" -shuffle=on $TEST_PKGS
}

# Forward arbitrary args to `go test` against a single package.
run_pkg() {
    local pkg="$1"; shift
    echo "==> go test $pkg $*"
    go test -race "$pkg" "$@"
}

case "${1:-race}" in
    unit)  run_unit ;;
    race)  run_race ;;
    cover) shift; run_cover "${1:-}" ;;
    bench) run_bench ;;
    fuzz)  shift; run_fuzz "$@" ;;
    flake) shift; run_flake "${1:-10}" ;;
    all)
        run_race
        echo
        run_cover
        ;;
    -h|--help|help)
        sed -n '2,/^set -e/p' "$0" | sed 's/^# \?//'
        ;;
    *)
        # Treat first arg as a package path
        run_pkg "$@"
        ;;
esac
