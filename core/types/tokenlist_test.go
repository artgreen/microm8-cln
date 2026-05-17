package types

import (
	"testing"
)

// TestTokenList_PushPop is the basic stack contract.
func TestTokenList_PushPop(t *testing.T) {
	t.Parallel()
	tl := NewTokenList()
	tl.Push(NewToken(NUMBER, "1"))
	tl.Push(NewToken(NUMBER, "2"))
	tl.Push(NewToken(NUMBER, "3"))

	if got := tl.Size(); got != 3 {
		t.Errorf("Size after 3 pushes = %d, want 3", got)
	}
	if got := tl.Pop(); got == nil || got.Content != "3" {
		t.Errorf("Pop() = %v, want token \"3\"", got)
	}
	if got := tl.Pop(); got == nil || got.Content != "2" {
		t.Errorf("second Pop() = %v, want token \"2\"", got)
	}
	if got := tl.Size(); got != 1 {
		t.Errorf("Size after 2 pops = %d, want 1", got)
	}
}

// TestTokenList_Shift mirrors Pop but from the head.
func TestTokenList_Shift(t *testing.T) {
	t.Parallel()
	tl := NewTokenList()
	tl.Push(NewToken(NUMBER, "1"))
	tl.Push(NewToken(NUMBER, "2"))

	if got := tl.Shift(); got == nil || got.Content != "1" {
		t.Errorf("Shift() = %v, want token \"1\"", got)
	}
	if got := tl.Size(); got != 1 {
		t.Errorf("Size after Shift = %d, want 1", got)
	}
	if got := tl.Shift(); got == nil || got.Content != "2" {
		t.Errorf("second Shift() = %v, want token \"2\"", got)
	}
	if got := tl.Shift(); got != nil {
		t.Errorf("Shift() on empty list = %v, want nil", got)
	}
}

// TestTokenList_UnShift_PutsAtHead.
func TestTokenList_UnShift_PutsAtHead(t *testing.T) {
	t.Parallel()
	tl := NewTokenList()
	tl.Push(NewToken(NUMBER, "2"))
	tl.UnShift(NewToken(NUMBER, "1")) // prepend
	tl.Push(NewToken(NUMBER, "3"))

	for i := 0; i < 3; i++ {
		want := []string{"1", "2", "3"}[i]
		got := tl.Get(i)
		if got == nil || got.Content != want {
			t.Errorf("[%d] = %v, want token %q", i, got, want)
		}
	}
}

// TestTokenList_Get_OutOfRangeReturnsNil pins the documented safe-fail
// behaviour.
func TestTokenList_Get_OutOfRangeReturnsNil(t *testing.T) {
	t.Parallel()
	tl := NewTokenList()
	tl.Push(NewToken(NUMBER, "1"))

	if got := tl.Get(-1); got != nil {
		t.Errorf("Get(-1) = %v, want nil", got)
	}
	if got := tl.Get(100); got != nil {
		t.Errorf("Get(100) = %v, want nil", got)
	}
}

// TestTokenList_Insert_AtPosition.
func TestTokenList_Insert_AtPosition(t *testing.T) {
	t.Parallel()
	tl := NewTokenList()
	tl.Push(NewToken(NUMBER, "a"))
	tl.Push(NewToken(NUMBER, "c"))
	tl.Insert(1, NewToken(NUMBER, "b"))

	for i, want := range []string{"a", "b", "c"} {
		if got := tl.Get(i).Content; got != want {
			t.Errorf("[%d] = %q, want %q", i, got, want)
		}
	}
}

// TestTokenList_Remove_ByIndex pins the Remove contract: removes and
// returns the token, shifts everything after down.
func TestTokenList_Remove_ByIndex(t *testing.T) {
	t.Parallel()
	tl := NewTokenList()
	for _, s := range []string{"a", "b", "c"} {
		tl.Push(NewToken(NUMBER, s))
	}

	removed := tl.Remove(1)
	if removed == nil || removed.Content != "b" {
		t.Errorf("Remove(1) = %v, want token \"b\"", removed)
	}
	if tl.Size() != 2 {
		t.Errorf("Size after Remove = %d, want 2", tl.Size())
	}
	if tl.Get(0).Content != "a" || tl.Get(1).Content != "c" {
		t.Errorf("after Remove: [%s, %s], want [a, c]",
			tl.Get(0).Content, tl.Get(1).Content)
	}
}

// TestTokenList_Clear_EmptiesList.
func TestTokenList_Clear_EmptiesList(t *testing.T) {
	t.Parallel()
	tl := NewTokenList()
	tl.Push(NewToken(NUMBER, "1"))
	tl.Push(NewToken(NUMBER, "2"))
	tl.Clear()
	if tl.Size() != 0 {
		t.Errorf("Size after Clear = %d, want 0", tl.Size())
	}
	if tl.Pop() != nil {
		t.Error("Pop after Clear returned non-nil")
	}
}

// TestTokenList_RPeekLPeek_DontMutate confirms RPeek/LPeek return a
// valid token without modifying the list, and return an INVALID
// sentinel when empty.
func TestTokenList_RPeekLPeek_DontMutate(t *testing.T) {
	t.Parallel()
	tl := NewTokenList()
	tl.Push(NewToken(NUMBER, "1"))
	tl.Push(NewToken(NUMBER, "2"))

	if got := tl.RPeek().Content; got != "2" {
		t.Errorf("RPeek = %q, want \"2\"", got)
	}
	if got := tl.LPeek().Content; got != "1" {
		t.Errorf("LPeek = %q, want \"1\"", got)
	}
	if tl.Size() != 2 {
		t.Error("peek mutated the list")
	}

	// Empty list returns sentinel.
	empty := NewTokenList()
	if got := empty.RPeek().Type; got != INVALID {
		t.Errorf("RPeek on empty = %v, want INVALID", got)
	}
	if got := empty.LPeek().Type; got != INVALID {
		t.Errorf("LPeek on empty = %v, want INVALID", got)
	}
}

// TestTokenList_Equals_PinsEqualFoldSemantics: the Phase 5 SA6005
// migration replaced ToLower(a)==ToLower(b) with EqualFold. Equals
// uses EqualFold (case-insensitive) so "FOO" should equal "foo".
func TestTokenList_Equals_CaseInsensitive(t *testing.T) {
	t.Parallel()
	a := NewTokenList()
	a.Push(NewToken(KEYWORD, "PRINT"))
	a.Push(NewToken(STRING, "Hello"))

	b := NewTokenList()
	b.Push(NewToken(KEYWORD, "print"))
	b.Push(NewToken(STRING, "hello"))

	if !a.Equals(b) {
		t.Error("Equals returned false for case-different but otherwise identical lists")
	}

	c := NewTokenList()
	c.Push(NewToken(KEYWORD, "PRINT"))
	if a.Equals(c) {
		t.Error("Equals returned true for different-sized lists")
	}
}

// TestTokenList_IndexOf pins the documented (and possibly-quirky)
// behaviour of IndexOf/IndexOfN. The implementation uses `i > start`
// rather than `>= start`, which means IndexOf (which calls
// IndexOfN(0, ...)) skips index 0. Callers in core/interpreter rely
// on this off-by-one — see core/interpreter/interpreter.go:3161 —
// so we pin the current behaviour and flag the suspicion.
//
// TODO: revisit whether this `> start` is a bug or a feature; if a
// bug, fix and update both interpreter usages and this test.
func TestTokenList_IndexOf(t *testing.T) {
	t.Parallel()
	tl := NewTokenList()
	tl.Push(NewToken(KEYWORD, "PRINT")) // index 0 — UNFINDABLE by IndexOf
	tl.Push(NewToken(STRING, "hello"))
	tl.Push(NewToken(SEPARATOR, ","))
	tl.Push(NewToken(STRING, "world"))

	if got := tl.IndexOf(SEPARATOR, ","); got != 2 {
		t.Errorf("IndexOf(SEPARATOR, \",\") = %d, want 2", got)
	}
	// Index 0 is silently unreachable from IndexOf.
	if got := tl.IndexOf(KEYWORD, "PRINT"); got != -1 {
		t.Errorf("IndexOf(KEYWORD, \"PRINT\") at idx 0 = %d, "+
			"want -1 (the start=0 off-by-one in IndexOfN); if you "+
			"fixed the off-by-one, also fix callers in "+
			"core/interpreter/interpreter.go:3161-3162", got)
	}
	if got := tl.IndexOf(KEYWORD, "MISSING"); got != -1 {
		t.Errorf("IndexOf(KEYWORD, \"MISSING\") = %d, want -1", got)
	}
}

// TestTokenList_IndexOfN_StartIsExclusive documents the strict-
// inequality contract — start=N skips indices ≤ N. If this gets
// "fixed" to inclusive, several interpreter parsing paths need to
// adjust their caller arguments.
func TestTokenList_IndexOfN_StartIsExclusive(t *testing.T) {
	t.Parallel()
	tl := NewTokenList()
	tl.Push(NewToken(KEYWORD, "A"))
	tl.Push(NewToken(KEYWORD, "B"))
	tl.Push(NewToken(KEYWORD, "A"))

	// Search for "A" starting at index 0 — index 0 itself is skipped.
	if got := tl.IndexOfN(0, KEYWORD, "A"); got != 2 {
		t.Errorf("IndexOfN(0, A) = %d, want 2 (start is exclusive)", got)
	}
	// Search starting at index 1: should find index 2.
	if got := tl.IndexOfN(1, KEYWORD, "A"); got != 2 {
		t.Errorf("IndexOfN(1, A) = %d, want 2", got)
	}
}
