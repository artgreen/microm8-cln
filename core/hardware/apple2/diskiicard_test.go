package apple2

import (
	"bytes"
	"encoding/binary"
	"testing"
)

// TestInt642bytes_RoundTrip pins the Phase 5 SA1003 fix. The package
// originally had a sibling `int2bytes(int)` that called
// `binary.Write(b, binary.LittleEndian, i)` with a platform `int` —
// binary.Write rejects platform-int at runtime, so int2bytes always
// returned an empty byte slice. It was unused so the bug had no
// caller-visible effect, but the dead code was removed.
//
// int642bytes(int64) is the live function and DOES work. This test
// pins the contract: int642bytes(x) returns exactly 8 little-endian
// bytes and round-trips through binary.Read.
func TestInt642bytes_RoundTrip(t *testing.T) {
	t.Parallel()
	cases := []int64{0, 1, -1, 0x7fffffffffffffff, -0x8000000000000000, 0x12345678}
	for _, in := range cases {
		in := in
		t.Run("", func(t *testing.T) {
			t.Parallel()
			got := int642bytes(in)
			if len(got) != 8 {
				t.Fatalf("int642bytes(%d): got %d bytes, want 8", in, len(got))
			}

			var out int64
			if err := binary.Read(bytes.NewReader(got), binary.LittleEndian, &out); err != nil {
				t.Fatalf("binary.Read: %v", err)
			}
			if out != in {
				t.Errorf("round-trip: in=%d, out=%d", in, out)
			}
		})
	}
}
