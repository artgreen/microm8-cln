#!/usr/bin/env bash
# Build the shipping microM8 binary.
# This is the canonical "build the thing" command for the repo.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

OUT="${OUT:-$REPO_ROOT/microM8}"

cd "$REPO_ROOT"
echo "Building octalyzer -> $OUT"
GOFLAGS=-mod=mod go build -o "$OUT" ./octalyzer
echo "Done: $OUT"
