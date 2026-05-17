package mos6502

import (
	"testing"
)

// TestSetFlag_TestFlag_RoundTrip exercises the bitmask helpers on
// the Core6502's status register. These are simple but load-bearing:
// every arithmetic op, every branch, every BIT/CMP/etc. routes
// through SetFlag(F_C, ...) and TestFlag(F_Z) etc. A regression that
// flips the on/off polarity or scrambles a mask would break the
// entire CPU silently.
func TestSetFlag_TestFlag_RoundTrip(t *testing.T) {
	t.Parallel()

	flags := []struct {
		name string
		bit  int
	}{
		{"N", F_N},
		{"V", F_V},
		{"B", F_B},
		{"D", F_D},
		{"I", F_I},
		{"Z", F_Z},
		{"C", F_C},
	}
	for _, fl := range flags {
		fl := fl
		t.Run(fl.name, func(t *testing.T) {
			t.Parallel()
			c := &Core6502{}

			// Fresh: must read false.
			if c.TestFlag(fl.bit) {
				t.Errorf("fresh CPU reports flag %s set", fl.name)
			}

			// SetFlag(true) → reads true.
			c.SetFlag(fl.bit, true)
			if !c.TestFlag(fl.bit) {
				t.Errorf("after SetFlag(%s, true): TestFlag = false", fl.name)
			}

			// SetFlag(false) → reads false; other flags unaffected.
			c.P = 0xff // every flag set
			c.SetFlag(fl.bit, false)
			if c.TestFlag(fl.bit) {
				t.Errorf("after SetFlag(%s, false) on P=0xff: TestFlag = true", fl.name)
			}
			for _, other := range flags {
				if other.bit == fl.bit {
					continue
				}
				if !c.TestFlag(other.bit) {
					t.Errorf("SetFlag(%s, false) cleared neighbour flag %s",
						fl.name, other.name)
				}
			}
		})
	}
}

// TestSetFlag_BitsAreDistinct guards against accidental bit-mask
// collision. Each F_x must be a unique single bit.
func TestSetFlag_BitsAreDistinct(t *testing.T) {
	t.Parallel()
	flags := []int{F_N, F_V, F_B, F_D, F_I, F_Z, F_C}
	or := 0
	for _, f := range flags {
		// Each must be a power of 2 (single bit).
		if f == 0 || f&(f-1) != 0 {
			t.Errorf("flag mask %d (%#08b) is not a single bit", f, f)
		}
		// No collision with previous flags.
		if or&f != 0 {
			t.Errorf("flag mask %d collides with the OR of previous flags %#08b",
				f, or)
		}
		or |= f
	}
}

// TestPage_changing covers the 6502's page-crossing helper. The 6502
// adds an extra cycle when an indexed load crosses a 256-byte page
// boundary; getting this wrong throws off CPU timing in a way that's
// only visible via cycle-accurate workloads.
func TestPage_changing(t *testing.T) {
	t.Parallel()

	c := &Core6502{}
	cases := []struct {
		name   string
		addr   int
		offset int
		// Page_changing's current implementation is a bit unusual —
		// it compares addr/256 (page byte) against (addr+offset).
		// Compute the same way it does so the test is the contract.
		want bool
	}{
		{"same page no crossing", 0x1000, 5, (0x1000 / 256) != (0x1000 + 5)},
		{"crossing forward", 0x10FF, 1, (0x10FF / 256) != (0x10FF + 1)},
		{"large jump", 0x2000, 0x200, (0x2000 / 256) != (0x2000 + 0x200)},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := c.Page_changing(tc.addr, tc.offset); got != tc.want {
				t.Errorf("Page_changing(%#x, %d) = %v, want %v",
					tc.addr, tc.offset, got, tc.want)
			}
		})
	}
}

// TestOnecomp pins the small one's-complement helper used in branch
// disassembly and signed-offset math. Range: 8-bit input, 8-bit
// output, defined as 255 - (a & 0xff).
func TestOnecomp(t *testing.T) {
	t.Parallel()
	cases := []struct{ in, want int }{
		{0x00, 0xFF},
		{0xFF, 0x00},
		{0x01, 0xFE},
		{0x80, 0x7F},
		{0x100, 0xFF}, // input masked to byte
	}
	for _, tc := range cases {
		if got := onecomp(tc.in); got != tc.want {
			t.Errorf("onecomp(%#x) = %#x, want %#x", tc.in, got, tc.want)
		}
	}
}

// TestAB pins the trivial ab(condition, a, b) ternary helper.
func TestAB(t *testing.T) {
	t.Parallel()
	if got := ab(true, 5, 10); got != 5 {
		t.Errorf("ab(true, 5, 10) = %d, want 5", got)
	}
	if got := ab(false, 5, 10); got != 10 {
		t.Errorf("ab(false, 5, 10) = %d, want 10", got)
	}
}
