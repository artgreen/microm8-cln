#!/usr/bin/env bash
# Build the microM8 binary and its companion targets.
#
# Subcommands (mirrors the legacy octalyzer/lmake.sh):
#   build.sh                # build microM8 (default)
#   build.sh remint         # build with -tags remint    -> ./remint
#   build.sh nox            # build with -tags nox       -> ./noxarchaist
#   build.sh asm            # build the 6502 assembler   -> ./asm65xx
#   build.sh macos          # cross-compile darwin/amd64 -> ./microM8-darwin-amd64
#   build.sh run            # build + launch microM8
#   build.sh profile        # build + launch with -inst-vms
#   build.sh assets         # regenerate octalyzer/assets/assets.go via go-bindata
#
# Environment:
#   OUT       output path for the default microM8 build (default: ./microM8)
#   GOFLAGS   defaults to -mod=mod
#
# Notes vs the legacy octalyzer/lmake.sh:
#   - Dropped `go build -i` (deprecated; no-op since Go 1.20).
#   - Dropped `xgo` for the macos target. Modern Go cross-compiles natively
#     via GOOS/GOARCH. However, the GL cgo deps mean cross-compiling
#     to a different darwin arch still needs a working cross-toolchain
#     (clang with the right SDK). The `macos` target is included for parity
#     and will likely require local tooling tweaks to actually run; we do
#     not exercise it in `check.sh`.
#   - Dropped `unset GOROOT`; no longer needed in module mode.
#   - `assets` invokes `go-bindata` directly. Install with:
#         go install github.com/kevinburke/go-bindata/v4/go-bindata@latest
#     The generated assets/assets.go is committed; only re-run when the
#     bundled files under octalyzer/{fonts,images,profile,data,debugger,...}
#     change.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

# Windows host detection (Git Bash sets OS=Windows_NT). Matches lmake.sh.
EXE_SUFFIX=""
if [ "${OS:-}" = "Windows_NT" ]; then
    EXE_SUFFIX=".exe"
fi

DEFAULT_OUT="$REPO_ROOT/microM8${EXE_SUFFIX}"
OUT="${OUT:-$DEFAULT_OUT}"

cd "$REPO_ROOT"
export GOFLAGS="${GOFLAGS:--mod=mod}"

build_microm8() {
    echo "==> Building microM8 -> $OUT"
    go build -o "$OUT" ./octalyzer
    echo "Done: $OUT"
}

build_remint() {
    local out="$REPO_ROOT/remint${EXE_SUFFIX}"
    echo "==> Building remint (tags: remint) -> $out"
    go build -tags remint -o "$out" ./octalyzer
    echo "Done: $out"
}

build_nox() {
    local out="$REPO_ROOT/noxarchaist${EXE_SUFFIX}"
    echo "==> Building noxarchaist (tags: nox) -> $out"
    go build -tags nox -o "$out" ./octalyzer
    echo "Done: $out"
}

build_asm() {
    local out="$REPO_ROOT/asm65xx${EXE_SUFFIX}"
    echo "==> Building 6502 assembler -> $out"
    go build -o "$out" ./core/hardware/cpu/mos6502/asm/cmd
    echo "Done: $out"
}

build_macos() {
    # Cross-compile to darwin/amd64. The CGO/OpenGL surface means this
    # needs a working clang cross-toolchain — vanilla `go build` on an
    # arm64 mac won't pick up the right linker flags without help. We
    # invoke the build for parity with the legacy script; if the toolchain
    # isn't set up, the build will fail at the link step with a clear
    # message rather than silently producing a broken binary.
    local out="$REPO_ROOT/microM8-darwin-amd64"
    echo "==> Cross-compiling for darwin/amd64 -> $out"
    echo "    (requires a clang darwin/amd64 cross-toolchain; not exercised by check.sh)"
    GOOS=darwin GOARCH=amd64 CGO_ENABLED=1 go build -o "$out" ./octalyzer
    echo "Done: $out"
}

run_microm8() {
    build_microm8
    echo "==> Launching $OUT"
    "$OUT"
}

run_profile() {
    build_microm8
    echo "==> Launching $OUT -inst-vms"
    "$OUT" -inst-vms
}

build_assets() {
    # The legacy assetbuild.sh lives in octalyzer/ and uses paths relative
    # to that directory, so cd in before running it.
    if ! command -v go-bindata >/dev/null 2>&1; then
        echo "go-bindata not found. Install with:" >&2
        echo "    go install github.com/kevinburke/go-bindata/v4/go-bindata@latest" >&2
        echo "and ensure \$(go env GOPATH)/bin is on \$PATH." >&2
        return 1
    fi
    echo "==> Regenerating octalyzer/assets/assets.go"
    ( cd "$REPO_ROOT/octalyzer" && ./assetbuild.sh )
    echo "Done. Review the diff in octalyzer/assets/assets.go before committing."
}

case "${1:-build}" in
    build|microm8|microM8)
        build_microm8
        ;;
    remint)
        build_remint
        ;;
    nox)
        build_nox
        ;;
    asm)
        build_asm
        ;;
    macos)
        build_macos
        ;;
    run)
        run_microm8
        ;;
    profile)
        run_profile
        ;;
    assets)
        build_assets
        ;;
    -h|--help|help)
        sed -n '2,/^set -e/p' "$0" | sed 's/^# \?//'
        ;;
    *)
        echo "Usage: $0 [build|remint|nox|asm|macos|run|profile|assets]" >&2
        exit 2
        ;;
esac
