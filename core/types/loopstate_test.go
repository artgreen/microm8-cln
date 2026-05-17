package types

import (
	"reflect"
	"testing"
)

// TestLoopState_CodeIsPointerType pins the Phase 4b cascade: LoopState
// previously embedded Algorithm by value, which meant any time a
// LoopState got copied (e.g. stored in a map or returned from a
// function), the Algorithm's sync.Mutex was copied too — and any
// subsequent method call took the lock on the copy, providing zero
// synchronisation.
//
// The fix is structural: `Code *Algorithm`. This test pins the field
// type so a future refactor doesn't quietly re-embed the value.
func TestLoopState_CodeIsPointerType(t *testing.T) {
	t.Parallel()
	ls := LoopState{}
	field, ok := reflect.TypeOf(ls).FieldByName("Code")
	if !ok {
		t.Fatal("LoopState has no Code field")
	}
	if field.Type.Kind() != reflect.Pointer {
		t.Fatalf("LoopState.Code must be a pointer (*Algorithm), got %v. "+
			"Embedding Algorithm by value re-introduces the Phase 4b "+
			"lock-by-value cascade — the sync.Mutex gets copied with every "+
			"LoopState assignment and stops providing synchronisation.",
			field.Type)
	}
	if field.Type.Elem().Name() != "Algorithm" {
		t.Errorf("LoopState.Code element type = %v, want Algorithm",
			field.Type.Elem())
	}
}

// TestLoopState_PointerSharing confirms two LoopStates sharing the same
// Algorithm see each other's mutations — this is the *correct*
// behaviour for the FOR/NEXT machinery (multiple frames may reference
// the same source program). The value-by-Algorithm bug would have
// silently broken this because each frame got its own copy.
func TestLoopState_PointerSharing(t *testing.T) {
	t.Parallel()
	alg := NewAlgorithm()
	ls1 := LoopState{VarName: "I", Code: alg}
	ls2 := LoopState{VarName: "J", Code: alg}

	ls1.Code.Put(10, Line{NewStatement()})
	// ls2 sees the same line because they share the pointer.
	if _, ok := ls2.Code.Get(10); !ok {
		t.Fatal("ls2.Code.Get(10) returned !ok; LoopState.Code is not " +
			"sharing the underlying Algorithm (Phase 4b regression)")
	}
}

// TestLoopState_MarshalRoundTrip exercises the binary serialisation
// helpers around the LoopState fields that *are* serialised. Code is
// not part of the wire format (it's restored from context), which is
// part of why pointerising it was safe.
func TestLoopState_MarshalRoundTrip(t *testing.T) {
	t.Parallel()
	in := LoopState{
		Step:    2.5,
		Start:   1.0,
		Finish:  100.0,
		VarName: "I",
		Entry:   CodeRef{Line: 10, Statement: 2},
	}
	data, err := in.MarshalBinary()
	if err != nil {
		t.Fatalf("MarshalBinary: %v", err)
	}

	var out LoopState
	if err := out.UnmarshalBinary(data); err != nil {
		t.Fatalf("UnmarshalBinary: %v", err)
	}

	if out.VarName != in.VarName {
		t.Errorf("VarName: got %q, want %q", out.VarName, in.VarName)
	}
	// float32 round-trip — exact for these values.
	if out.Step != in.Step || out.Start != in.Start || out.Finish != in.Finish {
		t.Errorf("numeric fields differ: got Step=%v Start=%v Finish=%v, want %v/%v/%v",
			out.Step, out.Start, out.Finish, in.Step, in.Start, in.Finish)
	}
	if out.Entry != in.Entry {
		t.Errorf("Entry: got %+v, want %+v", out.Entry, in.Entry)
	}
}

// TestLoopState_UnmarshalShortData asserts the failure path: a too-short
// buffer must return an error rather than panic.
func TestLoopState_UnmarshalShortData(t *testing.T) {
	t.Parallel()
	var ls LoopState
	if err := ls.UnmarshalBinary([]uint64{0, 0, 0}); err == nil {
		t.Fatal("UnmarshalBinary([3 elements]) returned nil, want error")
	}
}
