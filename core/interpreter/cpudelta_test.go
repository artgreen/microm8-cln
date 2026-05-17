package interpreter

import (
	"testing"

	"paleotronic.com/freeze"
)

// TestSimpleDelta is an unfinished scaffolding test (pre-Phase-1).
// It ends with `t.Fail()` and only logs the encoded bytes — no
// assertion. Skipping until the test is rewritten to assert
// round-trip equivalence between the two encodings.
func TestSimpleDelta(t *testing.T) {
	t.Skip("CPUDelta round-trip test needs rewrite: only logs bytes then " +
		"t.Fail()s without comparing them")

	s1 := &freeze.CPURegs{A: 0x01, X: 0x02, Y: 0x03, PC: 0x4000, P: 128}
	s2 := &freeze.CPURegs{A: 0xff, X: 0x02, Y: 0x03, PC: 0x4002, P: 159}

	d := getDelta(s1, s2)
	b := d.ToBytes()
	t.Logf("Delta encoded bytes 1 = %+v", b)

	d2 := &CPUDelta{}
	d2.FromBytes(b)
	b2 := d.ToBytes()
	t.Logf("Delta encoded bytes 2 = %+v", b2)
}
