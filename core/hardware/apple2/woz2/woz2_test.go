package woz2

import (
	"testing"

	"paleotronic.com/core/memory"
)

// newTestTrack returns a WOZ2Track with `bitCount` writable bits and
// realBitCount preconfigured so BitCount() returns a sensible value.
func newTestTrack(t *testing.T, bitCount int) *WOZ2Track {
	t.Helper()
	const blocks = 13 // generous; one block = 512 bytes
	data := memory.NewMemByteSlice(blocks * 512)
	return &WOZ2Track{
		Header:       memory.NewMemByteSlice(8),
		Data:         data,
		realBitCount: bitCount,
	}
}

// TestWOZ2Track_DirtyWindowInitOnFirstWrite mirrors the WOZ1 pin for
// the v2 image variant: Phase 5 SA4000 fix. The condition
// `w.mMin == w.mMin && w.mMax == 0` evaluated to `w.mMax == 0` and
// was replaced with the clearer `w.mMin == 0 && w.mMax == 0`.
func TestWOZ2Track_DirtyWindowInitOnFirstWrite(t *testing.T) {
	t.Parallel()
	w := newTestTrack(t, 50000)

	w.WriteBit(1000, 1)
	if w.mMin != 1000 || w.mMax != 1000 {
		t.Fatalf("after first WriteBit(1000): mMin=%d mMax=%d, want 1000/1000",
			w.mMin, w.mMax)
	}

	w.WriteBit(2000, 1)
	if w.mMin != 1000 || w.mMax != 2000 {
		t.Fatalf("after WriteBit(2000): mMin=%d mMax=%d, want 1000/2000",
			w.mMin, w.mMax)
	}

	w.WriteBit(500, 1)
	if w.mMin != 500 || w.mMax != 2000 {
		t.Fatalf("after WriteBit(500): mMin=%d mMax=%d, want 500/2000",
			w.mMin, w.mMax)
	}
}

// TestWOZ2Track_ReadBitRoundTrip: same sanity check as the WOZ1 file.
func TestWOZ2Track_ReadBitRoundTrip(t *testing.T) {
	t.Parallel()
	w := newTestTrack(t, 50000)

	for _, ptr := range []int{0, 1, 7, 8, 1000, 49999} {
		w.WriteBit(ptr, 1)
		if got := w.ReadBit(ptr); got != 1 {
			t.Errorf("after WriteBit(%d, 1): ReadBit = %d, want 1", ptr, got)
		}
		w.WriteBit(ptr, 0)
		if got := w.ReadBit(ptr); got != 0 {
			t.Errorf("after WriteBit(%d, 0): ReadBit = %d, want 0", ptr, got)
		}
	}
}
