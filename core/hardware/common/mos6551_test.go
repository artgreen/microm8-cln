package common

import (
	"testing"
)

// TestMOS6551StatusRegister_ResetPersistsViaPointerReceiver pins the
// Phase 5 SA4005 cluster fix on mos6551StatusRegister. Reset(),
// ProgramReset(), GetValue(), and String() previously had value
// receivers — so calls like `m.Status.Reset()` mutated a *copy* of
// the register and the change was discarded. After the fix all methods
// use pointer receivers; the writes persist.
//
// This test would have FAILED before the fix: `r.state` after Reset()
// would still be the zero value because Reset() wrote to a copy.
func TestMOS6551StatusRegister_ResetPersistsViaPointerReceiver(t *testing.T) {
	t.Parallel()
	var r mos6551StatusRegister
	r.Reset()
	// Reset() sets state to 0x10 (srTXO — Tx OK). If the write didn't
	// persist (value-receiver bug), r.state would still be 0.
	if r.state != 0x10 {
		t.Fatalf("after Reset(): state=%#x, want 0x10. The write didn't "+
			"persist — has Reset()'s receiver been changed back to value?",
			r.state)
	}
}

// TestMOS6551StatusRegister_ProgramReset pins the same fix on the
// other mutator. ProgramReset clears bit 1 (srFRM) from the current
// state; if the receiver is a value, the AND-mask doesn't persist.
func TestMOS6551StatusRegister_ProgramReset(t *testing.T) {
	t.Parallel()
	r := mos6551StatusRegister{state: 0xff}
	r.ProgramReset()
	// 0xff & 0xfd = 0xfd (clear bit 1).
	if r.state != 0xfd {
		t.Fatalf("after ProgramReset() with state=0xff: state=%#x, want 0xfd",
			r.state)
	}
}

// TestMOS6551StatusRegister_SetClearViaPointer covers the Set/Clear
// helpers (which already had pointer receivers, but live in the same
// state-machine cluster — if anyone consolidates the receiver style
// later we want a test that screams).
func TestMOS6551StatusRegister_SetClearViaPointer(t *testing.T) {
	t.Parallel()
	var r mos6551StatusRegister
	r.Set(srIRQ) // 0x80
	if !r.IsSet(srIRQ) {
		t.Errorf("Set(srIRQ) did not persist (state = %#x)", r.state)
	}
	r.Clear(srIRQ)
	if r.IsSet(srIRQ) {
		t.Errorf("Clear(srIRQ) did not persist (state = %#x)", r.state)
	}
}

// TestMOS6551StatusRegister_GetValueFiresCallback exercises the
// callback hook on a *pointer*-receiver GetValue. Previously GetValue
// took a value receiver; the call worked but any callback that mutated
// the register via this receiver would have hit the lock-by-value
// trap. The fix is that GetValue runs on the same register the caller
// owns, so a callback that writes through r persists.
func TestMOS6551StatusRegister_GetValueFiresCallback(t *testing.T) {
	t.Parallel()
	fired := 0
	r := mos6551StatusRegister{
		state: 0x42,
		callback: func() {
			fired++
		},
	}
	got := r.GetValue()
	if fired != 1 {
		t.Errorf("callback fired %d times, want 1", fired)
	}
	if got != 0x42 {
		t.Errorf("GetValue() = %#x, want 0x42", got)
	}
}
