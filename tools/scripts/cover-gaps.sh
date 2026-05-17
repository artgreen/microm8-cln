#!/usr/bin/env bash
# Show the largest uncovered functions in each allowlisted package so
# you know where to add the next test.
#
# Usage:
#   tools/scripts/cover-gaps.sh              # all allowlisted packages
#   tools/scripts/cover-gaps.sh ./core/types # one package
#   tools/scripts/cover-gaps.sh -n 5         # top-5 per package (default 3)
#
# Output: per-package, the top-N functions sorted by % uncovered
# multiplied by statement count (i.e. "biggest test gap by impact").
# Skips functions at 100% — those have no gap.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

# Parse args
TOP_N=3
PKGS=()
while [ $# -gt 0 ]; do
    case "$1" in
        -n) TOP_N="$2"; shift 2 ;;
        -h|--help)
            sed -n '2,/^set -e/p' "$0" | sed 's/^# \?//'
            exit 0
            ;;
        *) PKGS+=("$1"); shift ;;
    esac
done

cd "$REPO_ROOT"
export GOFLAGS="${GOFLAGS:--mod=mod}"

# Pull the package list from the allowlist (or use the user-supplied set).
if [ "${#PKGS[@]}" -gt 0 ]; then
    PACKAGES="${PKGS[*]}"
else
    PACKAGES=$(grep -v '^\s*#' "$REPO_ROOT/.ci/test-allowlist.txt" | grep -v '^\s*$')
fi

TMPDIR=$(mktemp -d)
trap 'rm -rf "$TMPDIR"' EXIT

for pkg in $PACKAGES; do
    profile="$TMPDIR/cover.out"
    # -count=1 disables the test cache so we get fresh coverage every run.
    if ! go test -count=1 -cover -coverprofile="$profile" "$pkg" >/dev/null 2>&1; then
        echo "==> $pkg: no tests or build error, skipping"
        echo
        continue
    fi
    if [ ! -s "$profile" ]; then
        # Package built fine but has no coverage data (no statements).
        continue
    fi

    # `go tool cover -func` outputs "<file>:<line>:	<func>	<pct>%"
    # We want to highlight the biggest gaps. Sort by (100-pct), drop 100%.
    summary=$(go tool cover -func="$profile" | awk -v top="$TOP_N" '
        # Last line is the total — capture and emit at the end.
        /^total:/ { total = $NF; next }
        # Format: filename:line:	funcname	pct%
        {
            # Last field has the % suffix.
            pct = $NF; gsub("%", "", pct)
            if (pct + 0 == 100) next  # skip fully-covered funcs
            print pct, $0
        }
        END { print "total", total }
    ' | sort -n | head -n "$TOP_N")

    if [ -z "$summary" ]; then
        continue
    fi

    # Always print the per-package total at the top.
    total_line=$(go tool cover -func="$profile" | awk '/^total:/ {print $NF}')
    printf "==> %s  (overall: %s)\n" "$pkg" "$total_line"
    echo "$summary" | awk '$1 != "total" { printf "    %5s%%  %s\n", $1, substr($0, index($0, $2)) }'
    echo
done
