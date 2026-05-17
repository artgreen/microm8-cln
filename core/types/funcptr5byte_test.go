package types

import (
	"testing"
)

// TestFuncPtr5b_GetPointer_PreservesHighByte pins the Phase 4b
// shift-too-large fix. The original code was:
//
//	v := int(i.hi<<8) | int(i.lo)
//
// where i.hi is a byte. `byte << 8` overflows the byte and produces 0,
// so GetPointer only ever returned the low byte. The fix shifts in
// int space: `int(i.hi)<<8 | int(i.lo)`. This test asserts every
// high-byte value lands in the right slot.
func TestFuncPtr5b_GetPointer_PreservesHighByte(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		hi   byte
		lo   byte
		want int
	}{
		{"zero", 0x00, 0x00, 0x0000},
		{"low only", 0x00, 0x42, 0x0042},
		{"high only", 0x12, 0x00, 0x1200},
		{"both", 0xAB, 0xCD, 0xABCD},
		{"max", 0xFF, 0xFF, 0xFFFF},
		// Without the fix every high-byte case would collapse to the lo
		// value because hi<<8 == 0 in byte arithmetic.
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			f := &FuncPtr5b{hi: tc.hi, lo: tc.lo}
			if got := f.GetPointer(); got != tc.want {
				t.Errorf("GetPointer() with hi=%#x lo=%#x = %#x, want %#x",
					tc.hi, tc.lo, got, tc.want)
			}
		})
	}
}

// TestFuncPtr5b_GetArgPointer_PreservesHighByte mirrors the above for
// the argument pointer pair (vhi/vlo). Same bug, same fix.
func TestFuncPtr5b_GetArgPointer_PreservesHighByte(t *testing.T) {
	t.Parallel()
	f := &FuncPtr5b{vhi: 0xC0, vlo: 0x80}
	if got := f.GetArgPointer(); got != 0xC080 {
		t.Errorf("GetArgPointer() with vhi=0xC0 vlo=0x80 = %#x, want 0xC080", got)
	}
}

// TestFuncPtr5b_SetPointer_RoundTrip exercises the setter/getter pair.
// The setter splits an int into hi/lo bytes; the getter reassembles
// them. A regression in either direction surfaces here.
func TestFuncPtr5b_SetPointer_RoundTrip(t *testing.T) {
	t.Parallel()
	for _, addr := range []int{0, 1, 0x100, 0x8000, 0xC080, 0xFFFF} {
		addr := addr
		t.Run("", func(t *testing.T) {
			t.Parallel()
			f := &FuncPtr5b{}
			f.SetPointer(addr)
			if got := f.GetPointer(); got != addr {
				t.Errorf("SetPointer(%#x) → GetPointer = %#x", addr, got)
			}
		})
	}
}

// TestFuncPtr5b_SetFirstBytePersists pins the Phase 5 SA4005 fix:
// SetFirstByte used a value receiver (`func (i FuncPtr5b)`), so the
// write `i.fb = l` mutated a copy and was discarded. After fixing to
// `*FuncPtr5b`, the write persists. This test would have failed before
// the fix because the second assertion would still see the old value.
func TestFuncPtr5b_SetFirstBytePersists(t *testing.T) {
	t.Parallel()
	f := &FuncPtr5b{}
	f.SetFirstByte(0x42)
	if got := f.GetFirstByte(); got != 0x42 {
		t.Fatalf("SetFirstByte(0x42) did not persist: GetFirstByte = %#x", got)
	}
	f.SetFirstByte(0xAB)
	if got := f.GetFirstByte(); got != 0xAB {
		t.Fatalf("second SetFirstByte did not persist: GetFirstByte = %#x", got)
	}
}

// (FuncPtr5b doesn't promise thread-safety — its callers serialise
// access via the interpreter. A race-clean test would be misleading.
// The persistence test above is the real Phase 5 pin.)
