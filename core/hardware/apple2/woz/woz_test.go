package woz

import (
	"bytes"
	"testing"

	"paleotronic.com/core/memory"
)

// newTestTrack returns a WOZTrack backed by a fresh in-memory byte
// slice with the given bit count. A WOZ track stores its BitCount in
// the last bytes of the data buffer, so we set those via SetBitCount
// before returning.
func newTestTrack(t *testing.T, bitCount uint16) *WOZTrack {
	t.Helper()
	// Size: bitstreamLength bytes for the bitstream + 4 bytes for
	// BytesUsed (2) + BitCount (2). Add some padding to be safe.
	data := memory.NewMemByteSlice(bitstreamLength + 16)
	w := &WOZTrack{Data: data}
	w.SetBitCount(bitCount)
	w.SetBytesUsed(uint16((bitCount + 7) / 8))
	return w
}

// TestWOZTrack_DirtyWindowInitOnFirstWrite pins the Phase 5 SA4000 fix.
// The original code was:
//
//	if w.mMin == w.mMin && w.mMax == 0 { ... } else { ... }
//
// `w.mMin == w.mMin` is always true (NaN aside), so effectively the
// condition was just `w.mMax == 0`. After the fix it is the clearer
// `w.mMin == 0 && w.mMax == 0` — identical behaviour for non-zero
// writes, but the intent is now correct (init-on-first-write).
//
// This test pins the observable contract: after the first WriteBit,
// mMin/mMax both equal the write position. Later writes shrink mMin
// or grow mMax appropriately.
func TestWOZTrack_DirtyWindowInitOnFirstWrite(t *testing.T) {
	t.Parallel()
	w := newTestTrack(t, 50000)

	// First write at offset 1000.
	w.WriteBit(1000, 1)
	if w.mMin != 1000 || w.mMax != 1000 {
		t.Fatalf("after first WriteBit(1000): mMin=%d mMax=%d, want 1000/1000",
			w.mMin, w.mMax)
	}

	// Subsequent write further along: mMax should grow, mMin stays.
	w.WriteBit(2000, 1)
	if w.mMin != 1000 || w.mMax != 2000 {
		t.Fatalf("after WriteBit(2000): mMin=%d mMax=%d, want 1000/2000",
			w.mMin, w.mMax)
	}

	// Subsequent write earlier: mMin should shrink, mMax stays.
	w.WriteBit(500, 1)
	if w.mMin != 500 || w.mMax != 2000 {
		t.Fatalf("after WriteBit(500): mMin=%d mMax=%d, want 500/2000",
			w.mMin, w.mMax)
	}
}

// TestWOZTrack_ReadBitRoundTrip is a small sanity check confirming
// WriteBit + ReadBit form a consistent pair across a few positions
// and both bit values.
func TestWOZTrack_ReadBitRoundTrip(t *testing.T) {
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

// FuzzNewWOZImage_NeverPanics asserts the WOZ file parser doesn't
// crash on arbitrary input. WOZ images come from the network or user
// disk, so a hostile or corrupt image must produce an error rather
// than a crash that takes microM8 down.
//
// Seeds: malformed headers, truncated files, oversize files.
func FuzzNewWOZImage_NeverPanics(f *testing.F) {
	f.Add([]byte(""))
	f.Add([]byte("WOZ1"))
	f.Add([]byte("WOZ1\xff\x0a\x0d\x0a"))
	f.Add(bytes.Repeat([]byte{0xff}, 16))
	f.Add(bytes.Repeat([]byte{0x00}, 256))

	f.Fuzz(func(t *testing.T, data []byte) {
		buf := memory.NewMemByteSlice(len(data) + 16)
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("NewWOZImage panicked on %d-byte input: %v", len(data), r)
			}
		}()
		_, _ = NewWOZImage(bytes.NewReader(data), buf)
	})
}
