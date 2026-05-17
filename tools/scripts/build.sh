#!/usr/bin/env bash
# Build the shipping microM8 binary.
# This is the canonical "build the thing" command for the repo.
#
# Until Phase 3 (GOPATH flatten), the module lives under GoPath/src/paleotronic.com.
# Once flattened, this script will be updated to drop the cd.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
MODULE_DIR="$REPO_ROOT/GoPath/src/paleotronic.com"

OUT="${OUT:-$REPO_ROOT/microM8}"

cd "$MODULE_DIR"
echo "Building octalyzer -> $OUT"
GOFLAGS=-mod=mod go build -o "$OUT" ./octalyzer
echo "Done: $OUT"
