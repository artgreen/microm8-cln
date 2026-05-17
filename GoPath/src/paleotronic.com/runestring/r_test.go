package runestring

import (
	"testing"
)

// --- Construction ------------------------------------------------------------

func TestNewRuneString_IsEmpty(t *testing.T) {
	t.Parallel()
	rs := NewRuneString()
	if rs.Length() != 0 {
		t.Errorf("NewRuneString().Length() = %d, want 0", rs.Length())
	}
	if rs.String() != "" {
		t.Errorf("NewRuneString().String() = %q, want %q", rs.String(), "")
	}
}

func TestCast_PreservesUnicode(t *testing.T) {
	t.Parallel()
	cases := []struct {
		in       string
		wantLen  int
		wantBack string
	}{
		{"", 0, ""},
		{"hello", 5, "hello"},
		{"héllo", 5, "héllo"}, // accented chars are single runes
		{"日本語", 3, "日本語"},     // CJK 3-byte UTF-8 chars
		{"á", 2, "á"},       // combining mark (separate runes)
		{"🎉", 1, "🎉"},         // surrogate-pair emoji is one rune
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.in, func(t *testing.T) {
			t.Parallel()
			rs := Cast(tc.in)
			if rs.Length() != tc.wantLen {
				t.Errorf("Cast(%q).Length() = %d, want %d", tc.in, rs.Length(), tc.wantLen)
			}
			if rs.String() != tc.wantBack {
				t.Errorf("Cast(%q).String() = %q, want %q", tc.in, rs.String(), tc.wantBack)
			}
		})
	}
}

// --- Mutation: Assign / Append ----------------------------------------------

func TestAssign_ReplacesContent(t *testing.T) {
	t.Parallel()
	rs := Cast("original")
	rs.Assign("replaced")
	if rs.String() != "replaced" {
		t.Errorf("after Assign: got %q, want %q", rs.String(), "replaced")
	}
}

func TestAppend_AppendsRunes(t *testing.T) {
	t.Parallel()
	rs := Cast("foo")
	rs.Append("bar")
	if rs.String() != "foobar" {
		t.Errorf("after Append: got %q, want %q", rs.String(), "foobar")
	}
}

func TestAppendRunes_AppendsAnotherRuneString(t *testing.T) {
	t.Parallel()
	a := Cast("alpha")
	b := Cast("beta")
	a.AppendRunes(b)
	if a.String() != "alphabeta" {
		t.Errorf("got %q, want %q", a.String(), "alphabeta")
	}
	// Source unchanged.
	if b.String() != "beta" {
		t.Errorf("source mutated: got %q, want %q", b.String(), "beta")
	}
}

func TestAppendSlice_AppendsSliceOfRunes(t *testing.T) {
	t.Parallel()
	rs := Cast("x")
	rs.AppendSlice([]rune{'y', 'z'})
	if rs.String() != "xyz" {
		t.Errorf("got %q, want %q", rs.String(), "xyz")
	}
}

// --- SubString ---------------------------------------------------------------

func TestSubString(t *testing.T) {
	t.Parallel()
	rs := Cast("abcdef")
	cases := []struct {
		name string
		s, e int
		want string
	}{
		{"full", 0, 6, "abcdef"},
		{"prefix", 0, 3, "abc"},
		{"middle", 2, 5, "cde"},
		{"empty range", 3, 3, ""},
		{"clamped end", 4, 100, "ef"},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := rs.SubString(tc.s, tc.e).String()
			if got != tc.want {
				t.Errorf("SubString(%d,%d) = %q, want %q", tc.s, tc.e, got, tc.want)
			}
		})
	}
}

// --- Copy --------------------------------------------------------------------

// Copy uses 1-based start, similar to Pascal/BASIC string semantics. Verify
// that contract because callers may rely on it.
func TestCopy_OneBasedStart(t *testing.T) {
	t.Parallel()
	rs := Cast("abcdef")
	cases := []struct {
		name         string
		start, count int
		want         string
	}{
		{"start at 1 (first char)", 1, 3, "abc"},
		{"start at 3 (middle)", 3, 2, "cd"},
		{"start past end", 100, 5, ""},
		{"count past end (clamps)", 4, 100, "def"},
		{"zero count", 1, 0, ""},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := Copy(rs, tc.start, tc.count).String()
			if got != tc.want {
				t.Errorf("Copy(%q, %d, %d) = %q, want %q",
					rs.String(), tc.start, tc.count, got, tc.want)
			}
		})
	}
}

// --- Concat ------------------------------------------------------------------

// Concat has a subtle aliasing bug worth pinning down: `tmp := s1` copies the
// struct header but shares the underlying slice. Document the current
// behavior so future refactors don't change it silently.
func TestConcat_ProducesConcatenation(t *testing.T) {
	t.Parallel()
	a := Cast("foo")
	b := Cast("bar")
	c := Concat(a, b)
	if c.String() != "foobar" {
		t.Errorf("Concat = %q, want %q", c.String(), "foobar")
	}
}

// --- Delete ------------------------------------------------------------------

func TestDelete_OneBasedStart(t *testing.T) {
	t.Parallel()
	rs := Cast("abcdef")
	cases := []struct {
		name         string
		start, count int
		want         string
	}{
		{"delete from start", 1, 2, "cdef"},
		{"delete middle", 3, 2, "abef"},
		{"delete past end (no-op-ish)", 100, 2, "abcdef"},
		{"delete count past end (truncates)", 4, 100, "abc"},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := Delete(rs, tc.start, tc.count).String()
			if got != tc.want {
				t.Errorf("Delete(%q, %d, %d) = %q, want %q",
					rs.String(), tc.start, tc.count, got, tc.want)
			}
		})
	}
}

// --- Pos: returns 1-based index, 0 for not-found ---------------------------

func TestPos(t *testing.T) {
	t.Parallel()
	rs := Cast("hello")
	cases := []struct {
		ch   rune
		want int
	}{
		{'h', 1},
		{'l', 3}, // first occurrence
		{'o', 5},
		{'z', 0}, // not present
	}
	for _, tc := range cases {
		tc := tc
		t.Run(string(tc.ch), func(t *testing.T) {
			t.Parallel()
			if got := Pos(tc.ch, rs); got != tc.want {
				t.Errorf("Pos(%q) = %d, want %d", tc.ch, got, tc.want)
			}
		})
	}
}

// --- HasPrefix: case-insensitive by design (via internal l()) -------------

func TestHasPrefix_IsCaseInsensitive(t *testing.T) {
	t.Parallel()
	cases := []struct {
		hay, needle string
		want        bool
	}{
		{"hello world", "hello", true},
		{"hello world", "HELLO", true},
		{"HELLO world", "hello", true},
		{"hello", "hellos", false}, // needle longer than haystack
		{"hello", "", false},       // empty needle returns false (by current contract)
		{"hello", "world", false},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.hay+"/"+tc.needle, func(t *testing.T) {
			t.Parallel()
			hay := Cast(tc.hay)
			needle := Cast(tc.needle)
			if got := hay.HasPrefix(needle); got != tc.want {
				t.Errorf("%q.HasPrefix(%q) = %v, want %v", tc.hay, tc.needle, got, tc.want)
			}
		})
	}
}
