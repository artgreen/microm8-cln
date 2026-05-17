package apple2

import (
	"testing"
)

// TestMockingboardChipBaseOffset_ChipsHaveDistinctBases pins the
// Phase 5 SA4028 fix. The original formula `(i%1)*0x80` always
// evaluated to 0 for every chip — Mockingboard's two AY-3-8910 chips
// would have collided on the same I/O window, with register writes
// for chip 1 silently affecting chip 0.
//
// The fix uses `(i%2)*0x80`. This pin asserts the INVARIANT we
// actually care about — chip 0 and chip 1 must have distinct base
// addresses, with chip 1 at 0x80 — so even if someone "simplifies"
// the formula in a way that breaks it, the test catches it.
func TestMockingboardChipBaseOffset_ChipsHaveDistinctBases(t *testing.T) {
	t.Parallel()

	if mockingboardChipBaseOffset(0) == mockingboardChipBaseOffset(1) {
		t.Fatalf("chip 0 and chip 1 share base address 0x%x — Mockingboard "+
			"register writes for the second AY-3-8910 will collide with "+
			"the first (Phase 5 SA4028 regression)",
			mockingboardChipBaseOffset(0))
	}
	if got := mockingboardChipBaseOffset(0); got != 0x00 {
		t.Errorf("chip 0 base = 0x%x, want 0x00", got)
	}
	if got := mockingboardChipBaseOffset(1); got != 0x80 {
		t.Errorf("chip 1 base = 0x%x, want 0x80", got)
	}
}
