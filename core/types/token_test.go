package types

import (
	"math"
	"testing"
)

// TestToken_AsInteger pins type-coercion-to-int across every TokenType.
// Every BASIC dialect routes integer arithmetic through this path.
func TestToken_AsInteger(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		tok  *Token
		want int
	}{
		{"nil token → 0", nil, 0},
		{"NUMBER int-valued", NewToken(NUMBER, "42"), 42},
		{"NUMBER negative", NewToken(NUMBER, "-17"), -17},
		{"NUMBER floor of float", NewToken(NUMBER, "3.7"), 3},
		{"INTEGER", NewToken(INTEGER, "100"), 100},
		// STRING-typed token: any non-empty string is truthy → 1; empty → 0
		{"STRING non-empty → 1", NewToken(STRING, "anything"), 1},
		{"STRING empty → 0", NewToken(STRING, ""), 0},
		// BOOLEAN's AsInteger uses ParseFloat on Content, NOT a
		// truthy/falsy check on the literal "true"/"false". This is a
		// quirk worth pinning: "1" → 1, "true" → 0 (ParseFloat fails).
		// AsNumeric is the one that recognises "true" / "yes".
		{"BOOLEAN content=1", NewToken(BOOLEAN, "1"), 1},
		{"BOOLEAN content=true → 0 (ParseFloat fails)", NewToken(BOOLEAN, "true"), 0},
		// WORD treated as a numeric expression in AsInteger
		{"WORD numeric", NewToken(WORD, "42"), 42},
		{"empty content NUMBER → 0", NewToken(NUMBER, ""), 0},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := tc.tok.AsInteger(); got != tc.want {
				t.Errorf("AsInteger() = %d, want %d", got, tc.want)
			}
		})
	}
}

// TestToken_AsFloat pins type-coercion-to-float64.
func TestToken_AsFloat(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		tok  *Token
		want float64
	}{
		{"nil → 0", nil, 0},
		{"NUMBER float", NewToken(NUMBER, "3.14"), 3.14},
		{"NUMBER integer", NewToken(NUMBER, "10"), 10},
		{"INTEGER", NewToken(INTEGER, "100"), 100},
		{"STRING non-empty → 1", NewToken(STRING, "x"), 1},
		{"STRING empty → 0", NewToken(STRING, ""), 0},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := tc.tok.AsFloat()
			if math.Abs(got-tc.want) > 1e-9 {
				t.Errorf("AsFloat() = %v, want %v", got, tc.want)
			}
		})
	}
}

// TestToken_AsString covers the inverse: rendering a token to text.
// BOOLEAN tokens stringify by their truthiness, STRING tokens are
// wrapped in double quotes, everything else uses Content directly.
func TestToken_AsString(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		tok  *Token
		want string
	}{
		{"nil → empty", nil, ""},
		{"BOOLEAN truthy → true", NewToken(BOOLEAN, "1"), "true"},
		{"BOOLEAN falsy → false", NewToken(BOOLEAN, "0"), "false"},
		{`STRING wrapped in quotes`, NewToken(STRING, "hello"), `"hello"`},
		{"NUMBER passthrough", NewToken(NUMBER, "3.14"), "3.14"},
		{"INTEGER passthrough", NewToken(INTEGER, "42"), "42"},
		{"KEYWORD passthrough", NewToken(KEYWORD, "PRINT"), "PRINT"},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := tc.tok.AsString(); got != tc.want {
				t.Errorf("AsString() = %q, want %q", got, tc.want)
			}
		})
	}
}

// TestToken_NewToken_DefaultsNumericContentToZero pins the constructor
// quirk: a NUMBER or INTEGER token with empty Content is coerced to
// "0" at creation. (BASIC parsers rely on this so an uninitialised
// numeric variable evaluates to 0 instead of producing a parse
// error.)
func TestToken_NewToken_DefaultsNumericContentToZero(t *testing.T) {
	t.Parallel()
	if got := NewToken(NUMBER, "").Content; got != "0" {
		t.Errorf("NewToken(NUMBER, \"\").Content = %q, want \"0\"", got)
	}
	if got := NewToken(INTEGER, "").Content; got != "0" {
		t.Errorf("NewToken(INTEGER, \"\").Content = %q, want \"0\"", got)
	}
	// STRING is not coerced — uninitialised STRING has empty content.
	if got := NewToken(STRING, "").Content; got != "" {
		t.Errorf("NewToken(STRING, \"\").Content = %q, want \"\"", got)
	}
}

// TestToken_IsIn covers the membership helper used by dialect
// dispatch ("is this token a KEYWORD or DYNAMICKEYWORD?" etc.).
func TestToken_IsIn(t *testing.T) {
	t.Parallel()
	tok := NewToken(KEYWORD, "PRINT")
	if !tok.IsIn([]TokenType{KEYWORD, FUNCTION}) {
		t.Error("KEYWORD token IsIn([KEYWORD, FUNCTION]) = false")
	}
	if tok.IsIn([]TokenType{STRING, NUMBER}) {
		t.Error("KEYWORD token IsIn([STRING, NUMBER]) = true")
	}
	if tok.IsIn([]TokenType{}) {
		t.Error("IsIn(empty list) = true")
	}
	// IsNotIn is the inverse.
	if tok.IsNotIn([]TokenType{KEYWORD, FUNCTION}) {
		t.Error("KEYWORD token IsNotIn([KEYWORD]) = true")
	}
}

// TestToken_Copy_DeepCopiesList confirms Copy() produces an
// independent token. The List field is also copied so mutating the
// copy's list doesn't bleed back into the original.
func TestToken_Copy_DeepCopiesList(t *testing.T) {
	t.Parallel()
	original := NewToken(LIST, "outer")
	original.List = NewTokenList()
	original.List.Push(NewToken(NUMBER, "1"))

	cp := original.Copy()
	if cp == original {
		t.Fatal("Copy() returned the same pointer")
	}
	if cp.Content != original.Content {
		t.Errorf("Copy().Content = %q, original.Content = %q", cp.Content, original.Content)
	}
	if cp.List == original.List {
		t.Error("Copy().List shares pointer with original.List")
	}
	if cp.List.Size() != 1 {
		t.Errorf("Copy().List.Size() = %d, want 1 (deep-copied)", cp.List.Size())
	}
}
