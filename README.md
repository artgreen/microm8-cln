# microM8 - Apple II Emulator

microM8 is a comprehensive Apple II emulator with advanced features for both casual users and developers. It provides accurate emulation of Apple II hardware along with modern integration capabilities including MCP (Model Context Protocol) support for AI assistant interaction.

## Features

- Full Apple II/II+ emulation
- 6502 CPU emulation with cycle-accurate timing
- Support for disk operations and various disk image formats
- Built-in assembler (asm65xx)
- MCP server integration for AI-powered control and automation
- Cross-platform support (Linux, macOS, Windows)
- Remote interface capabilities
- Debugging and development tools

## Project Structure

```
microm8-cln/
├── octalyzer/        # Main emulator client (binary entry point)
├── server/           # Server components
├── core/             # Core emulation logic
├── debugger/         # Debugging tools
├── api/              # Network/RPC API
├── internal/         # Internal-only support packages
├── tools/scripts/    # Build/test/watch entry points
├── .ci/              # Allowlists + coverage thresholds
└── ...
```

The module path is `paleotronic.com`. All Go source lives at the repository root — the historical `GoPath/src/paleotronic.com/` wrapper was removed in Phase 3 of the modernization migration.

## Building from Source

### Prerequisites

- Go 1.24 or later
- Git
- C toolchain (gcc/clang), `pkg-config`
- System libraries for CGO deps:
  - **macOS**: `brew install pkg-config portaudio`
  - **Linux**: `sudo apt-get install pkg-config portaudio19-dev libgl1-mesa-dev xorg-dev libxi-dev libxcursor-dev libxinerama-dev libxrandr-dev libasound2-dev`

### Build

From the repository root:

```bash
tools/scripts/build.sh
```

Produces a `microM8` binary at the repository root. This is equivalent to:

```bash
GOFLAGS=-mod=mod go build -o microM8 ./octalyzer
```

### Additional build targets via `lmake.sh`

The `octalyzer/lmake.sh` script supports several build variants:

```bash
cd octalyzer
./lmake.sh build     # standard build (default)
./lmake.sh run       # build and run
./lmake.sh asm       # build the assembler (asm65xx)
./lmake.sh remint    # build the remote-interface variant
./lmake.sh nox       # build without X11 (noxarchaist)
./lmake.sh macos     # cross-build for macOS x86_64 (requires xgo)
./lmake.sh profile   # build + run with profiling
```

Messages about `xgo` not being found can be safely ignored unless cross-compiling.

### Running tests

```bash
tools/scripts/check.sh         # full pre-commit (fmt + vet + test + build)
tools/scripts/test.sh          # race tests on the allowlist
tools/scripts/test.sh cover    # coverage report with threshold check
tools/scripts/test.sh fuzz FuzzRoundtrip ./encoding/base91
tools/scripts/watch.sh         # auto-rerun on save
```

See [TESTING.md](TESTING.md) for testing conventions and how to extend coverage.

## Running microM8

After building, you can run the emulator:

```bash
./microM8
```

### MCP Server Mode

To enable MCP server for AI integration:

**Stdio mode (default):**
```bash
./microM8 -mcp
```

**SSE mode (for web-based clients):**
```bash
./microM8 -mcp -mcp-mode sse
./microM8 -mcp -mcp-mode sse -mcp-port 8080
```

## Development

### Building components separately

**Server:**
```bash
go build ./server
```

**Client (octalyzer):**
```bash
go build ./octalyzer
```

**Remote interface:**
```bash
go build -tags remint ./octalyzer
```

## License

MIT License

## Contributing

[Contribution guidelines to be added]

## Support

[Support information to be added]
