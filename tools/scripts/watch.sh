#!/usr/bin/env bash
# Watch the module tree and re-run tests on save.
#
# Usage:
#   tools/scripts/watch.sh           # watch everything in the test allowlist
#   tools/scripts/watch.sh ./buffer  # watch + test a single package
#
# Cross-platform: uses fswatch on macOS, inotifywait on Linux, falls back to
# a polling loop if neither is installed. To install on macOS: brew install fswatch.
# To install on Linux: apt install inotify-tools.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

PKG="${1:-}"

run_tests() {
    echo "=== $(date '+%H:%M:%S') ==="
    if [ -n "$PKG" ]; then
        (cd "$REPO_ROOT" && GOFLAGS=-mod=mod go test -race -count=1 -shuffle=on "$PKG") || true
    else
        "$SCRIPT_DIR/test.sh" race || true
    fi
    echo
    echo "(watching for changes...)"
}

# Initial run
run_tests

if command -v fswatch >/dev/null 2>&1; then
    # macOS / BSD: fswatch with one-event-per-batch semantics
    fswatch -or --event Updated --event Created --event Removed \
        --exclude '\.git/' --exclude '/\.' \
        --include '\.go$' \
        "$REPO_ROOT" |
    while read -r _; do
        run_tests
    done
elif command -v inotifywait >/dev/null 2>&1; then
    # Linux
    while inotifywait -qre modify,create,delete --include '\.go$' \
        --exclude '\.git' "$REPO_ROOT" >/dev/null; do
        run_tests
    done
else
    # Pure-bash fallback: poll mtime hashes every 2s
    echo "(no fswatch or inotifywait; falling back to 2s polling — install fswatch/inotify-tools for instant feedback)"
    last=""
    while true; do
        cur=$(find "$REPO_ROOT" -name '*.go' -not -path '*/.git/*' \
              -exec stat -f '%m %N' {} + 2>/dev/null | md5 2>/dev/null || \
              find "$REPO_ROOT" -name '*.go' -not -path '*/.git/*' \
              -printf '%T@ %p\n' 2>/dev/null | md5sum | awk '{print $1}')
        if [ "$cur" != "$last" ]; then
            [ -n "$last" ] && run_tests
            last="$cur"
        fi
        sleep 2
    done
fi
