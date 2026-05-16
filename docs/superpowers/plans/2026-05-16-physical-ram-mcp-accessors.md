# Physical RAM MCP Accessors

## Summary

Add five MCP tools for physical RAM access: `list_ram_regions`, `read_ram_byte`, `read_ram_range`, `write_ram_byte`, and `write_ram_range`. These tools should address named emulator RAM regions directly, bypassing Apple II soft-switch mapping, while keeping the existing `read_memory_*` / `write_memory_*` tools as 6502-visible mapped-memory accessors.

## Key Changes

- Extend the memory package with safe direct RAM helpers:
  - `MemoryBlock.DirectWrite(offset int, value uint64) bool`
  - `MemoryBlock.IsRAM() bool`
  - `MemoryManagementUnit.ListRAMBlocks() []RAMBlockInfo`
- Add `list_ram_regions`, returning JSON text with RAM block labels, base address, mux base, size, active/read/write mapping state, and RAM type. Include normal Apple II regions such as `main.main.text`, `aux.main.hgr1`, `main.languagecard`, `aux.languagecard`, plus any future expansion-card RAM blocks.
- Add MCP params:
  - `read_ram_byte`: `{ "region": string, "offset": int }`
  - `read_ram_range`: `{ "region": string, "offset": int, "count": int }`
  - `write_ram_byte`: `{ "region": string, "offset": int, "value": int }`
  - `write_ram_range`: `{ "region": string, "offset": int, "values": int[] }`
- Validate region existence, RAM-only access, offset/count bounds, and byte values in `0..255`.
- Return range reads as the same hexdump style as `read_memory_range`, prefixed with region and offset. Writes should return a concise count/region/offset confirmation.

## Implementation Notes

- Touch primarily:
  - `/Users/green/projects/microm8-cln/GoPath/src/paleotronic.com/core/memory/mmu.go`
  - `/Users/green/projects/microm8-cln/GoPath/src/paleotronic.com/octalyzer/mcp_sdk.go`
- MCP handlers should resolve the current interpreter with `backend.ProducerMain.GetInterpreter(SelectedIndex)`, then use `e.GetMemoryMap().BlockMapper[e.GetMemIndex()]`.
- Do not change Apple II soft-switch state. These tools are physical-region accessors, not CPU bus emulation.
- Keep existing `read_memory`, `read_memory_range`, `write_memory`, and `write_memory_range` behavior unchanged. Their descriptions may be clarified as "6502-visible mapped memory".

## Test Plan

- Run `go test -mod=mod -vet=off ./octalyzer`.
- Build `microM8` and verify `strings ./microM8 | rg 'list_ram_regions|read_ram_byte|read_ram_range|write_ram_byte|write_ram_range'`.
- Start MCP with `-mcp -mcp-mode streaming -mcp-port 1983 -offline`; verify the tool lab shows all five new tools.
- Manual checks:
  - `list_ram_regions` includes both main and aux Apple II RAM regions.
  - Write one byte to `aux.main.text`, read it back with `read_ram_byte`.
  - Confirm `read_memory` at the same 6502 address does not necessarily change unless soft switches map that aux region.
  - Invalid region, out-of-range offset, and invalid byte value return clear errors.

## Assumptions

- `region + offset` is the public interface, not fake linear addresses like `0x10000`.
- Expansion-card support is achieved by listing and addressing any registered RAM `MemoryBlock`; no card-specific API is needed for v1.
- Tools return text content for compatibility with the current MCP tool lab.
