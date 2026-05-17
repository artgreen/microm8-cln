package interpreter

import (
	"testing"

	"paleotronic.com/ducktape"
)

// TestAJumpCounter_DecodesFullTargetAddress pins the Phase 4b fix
// in analyzer.go:
//
//	to := int(msg.Payload[5]<<16) | …
//
// `msg.Payload[5]` is a byte; `byte<<16` overflows the byte and
// produces 0. So the bit-16 nibble of `to` was always 0 — the
// analyser silently dropped the third byte of every target address.
// The fix shifts in int space: `int(msg.Payload[5])<<16`.
//
// This test encodes a JMP message whose payload[5] is non-zero and
// verifies the recovered address has the right value at bits 16–23.
// Under the bug the assertion fails because bit-16's nibble is 0.
func TestAJumpCounter_DecodesFullTargetAddress(t *testing.T) {
	t.Parallel()

	// Target address = 0x12345678. Encoded big-endian into Payload[4:8].
	const target = 0x12345678
	msg := &ducktape.DuckTapeBundle{
		ID: "JMP",
		Payload: []byte{
			0x00, 0x00, 0x00, 0x00, // from (ignored)
			0x12, 0x34, 0x56, 0x78, // to
			0x01, // isJump (ignored by current code, kept for header parity)
		},
	}

	ajc := NewAJumpCounter(1)
	if cont := ajc.Process(msg); !cont {
		t.Fatal("Process returned false; want true (JMP is the only ID that updates the counter)")
	}

	if count, ok := ajc.Jumps[target]; !ok {
		t.Fatalf("Jumps[%#x] = %d, want presence; full address was lost. "+
			"Pre-fix the byte<<16 overflow zeroed bits 16-23 of the address, "+
			"so this lookup landed at %#x instead.",
			target, count, 0x12005678)
	} else if count != 1 {
		t.Errorf("Jumps[%#x] = %d, want 1", target, count)
	}

	// Negative case: a non-JMP message must not affect the counter.
	other := &ducktape.DuckTapeBundle{ID: "XYZ", Payload: msg.Payload}
	if cont := ajc.Process(other); !cont {
		t.Error("Process returned false for non-JMP message")
	}
	if ajc.Jumps[target] != 1 {
		t.Errorf("non-JMP message modified the counter: Jumps[%#x] = %d, want 1",
			target, ajc.Jumps[target])
	}
}

// TestAJumpCounter_GetAddressesByJumpCounts is the small helper test.
// Asserts the by-count grouping works for a couple of straightforward
// inputs.
func TestAJumpCounter_GetAddressesByJumpCounts(t *testing.T) {
	t.Parallel()
	ajc := NewAJumpCounter(1)
	ajc.Jumps[0x100] = 1
	ajc.Jumps[0x200] = 2
	ajc.Jumps[0x300] = 1

	got := ajc.GetAddressesByJumpCounts(1)
	if len(got) != 2 {
		t.Errorf("GetAddressesByJumpCounts(1) returned %v, want 2 entries", got)
	}
	if got := ajc.GetAddressesByJumpCounts(2); len(got) != 1 || got[0] != 0x200 {
		t.Errorf("GetAddressesByJumpCounts(2) = %v, want [0x200]", got)
	}
	if got := ajc.GetAddressesByJumpCounts(99); len(got) != 0 {
		t.Errorf("GetAddressesByJumpCounts(99) = %v, want empty", got)
	}
}
