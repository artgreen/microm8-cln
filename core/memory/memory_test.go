package memory

import (
	"testing"
)

// TestNewMemoryMap_AllocatesSlotZero verifies the constructor's
// documented behaviour: NewMemoryMap pre-allocates slot 0's backing
// slice but leaves later slots nil-allocated (callers populate them
// via Alloc on the corresponding VM).
func TestNewMemoryMap_AllocatesSlotZero(t *testing.T) {
	t.Parallel()
	mm := NewMemoryMap()

	if mm.Data[0] == nil {
		t.Fatal("Data[0] is nil after NewMemoryMap; slot 0 should be pre-allocated")
	}
	if len(mm.Data[0]) != OCTALYZER_INTERPRETER_SIZE {
		t.Errorf("Data[0] length = %d, want %d", len(mm.Data[0]), OCTALYZER_INTERPRETER_SIZE)
	}

	if mm.BlockMapper[0] == nil {
		t.Error("BlockMapper[0] is nil; should be initialised by NewMemoryMap")
	}
}

// TestWriteReadInterpreterMemory_HighAddrRoundTrip exercises the
// direct-slice path (addresses ≥ 64K bypass the MMU). A bug in the
// modulo or slice indexing would surface here.
func TestWriteReadInterpreterMemory_HighAddrRoundTrip(t *testing.T) {
	t.Parallel()
	mm := NewMemoryMap()

	// Pick a high address that bypasses the MMU but is within the
	// interpreter slice.
	const addr = 0x20000 // 128K — well past the MMU-mapped 64K window
	mm.WriteInterpreterMemory(0, addr, 0xdeadbeef)

	got := mm.ReadInterpreterMemory(0, addr)
	if got != 0xdeadbeef {
		t.Errorf("Write/Read round-trip at 0x%x: got %#x, want 0xdeadbeef", addr, got)
	}

	// Distinct addresses don't collide.
	mm.WriteInterpreterMemory(0, addr+1, 0xcafef00d)
	if got := mm.ReadInterpreterMemory(0, addr); got != 0xdeadbeef {
		t.Errorf("addr untouched: got %#x, want 0xdeadbeef (write to addr+1 leaked)", got)
	}
	if got := mm.ReadInterpreterMemory(0, addr+1); got != 0xcafef00d {
		t.Errorf("addr+1: got %#x, want 0xcafef00d", got)
	}
}

// TestWriteReadInterpreterMemory_BlockMapped exercises the MMU path
// (addresses < 64K go through BlockMapper). Activation requires both
// Register (cache index) and SetBankREAD/WRITE (per-page resolution
// table) — they're two layers, and Register alone isn't enough to
// make a block live for reads/writes. The test pins both layers.
func TestWriteReadInterpreterMemory_BlockMapped(t *testing.T) {
	t.Parallel()
	mm := NewMemoryMap()

	// Register a 1KB RAM block at 0x2000.
	const base = 0x2000
	const size = 0x400
	block := NewMemoryBlockRAM(mm, 0, mm.MEMBASE(0), base, size, true, "test-ram", 0, false, 0)
	mm.BlockMapper[0].Register(block)

	// Activate read+write resolution for the block's banks (0x20..0x23
	// covers 0x2000..0x23FF).
	startBank := base / 256
	endBank := (base + size) / 256
	mm.BlockMapper[0].SetBankREAD(startBank, endBank, block)
	mm.BlockMapper[0].SetBankWRITE(startBank, endBank, block)

	// Write/read through the mapped region.
	mm.WriteInterpreterMemory(0, base+0x10, 0x12345678)
	if got := mm.ReadInterpreterMemory(0, base+0x10); got != 0x12345678 {
		t.Errorf("MMU round-trip at 0x%x: got %#x, want 0x12345678", base+0x10, got)
	}

	// Boundary: first and last addresses in the block.
	mm.WriteInterpreterMemory(0, base, 0x111)
	mm.WriteInterpreterMemory(0, base+size-1, 0x222)
	if got := mm.ReadInterpreterMemory(0, base); got != 0x111 {
		t.Errorf("MMU at block start 0x%x: got %#x, want 0x111", base, got)
	}
	if got := mm.ReadInterpreterMemory(0, base+size-1); got != 0x222 {
		t.Errorf("MMU at block end 0x%x: got %#x, want 0x222", base+size-1, got)
	}
}

// TestReadInterpreterMemory_NilSlotReturnsZero confirms the safety
// guard: Data[index] is nil for un-allocated slots, and the read path
// must return 0 rather than panic on a nil dereference.
func TestReadInterpreterMemory_NilSlotReturnsZero(t *testing.T) {
	t.Parallel()
	mm := NewMemoryMap()

	// Slot 1+ is uninitialised by default (only slot 0 gets the
	// pre-allocated backing slice). A read on the nil slot must be
	// safe and return 0.
	if got := mm.ReadInterpreterMemory(1, 0x100); got != 0 {
		t.Errorf("ReadInterpreterMemory(1, 0x100) on nil slot = %#x, want 0", got)
	}

	// WriteInterpreterMemory on a nil slot is documented to no-op.
	// Confirm it doesn't panic.
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("WriteInterpreterMemory on nil slot panicked: %v", r)
		}
	}()
	mm.WriteInterpreterMemory(1, 0x100, 0xff)
}

// TestSlotRestartFlag_RoundTrip exercises the slot-restart flag used
// by the Phase 7 RebootService. A bug in the underlying ReadGlobal/
// WriteGlobal would silently break clean-boot routing.
func TestSlotRestartFlag_RoundTrip(t *testing.T) {
	t.Parallel()
	mm := NewMemoryMap()

	// Initial state: not requesting restart.
	if mm.IntGetSlotRestart(0) {
		t.Error("freshly-constructed mm has slot-restart already set")
	}

	mm.IntSetSlotRestart(0, true)
	if !mm.IntGetSlotRestart(0) {
		t.Error("after Set(true): Get returned false")
	}

	mm.IntSetSlotRestart(0, false)
	if mm.IntGetSlotRestart(0) {
		t.Error("after Set(false): Get returned true")
	}
}
