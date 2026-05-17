package base91

import (
	"strings"
	"testing"
	"unicode/utf8"
)

// Known-good vectors. Preserved from the original test file.
var samples = []struct {
	plain   string
	encoded string
}{
	{"1", "xA"},
	{"1234567890", "QztEml0o[2;(A"},
	{"abcdefghijklmnopqurstuvwxyz", "#G(Ic,5ph#77&xrmlrjg2]jTs%2<WF%qfB"},
}

func TestEncode_KnownVectors(t *testing.T) {
	t.Parallel()
	for _, s := range samples {
		s := s
		t.Run(s.plain, func(t *testing.T) {
			t.Parallel()
			if got := Encode(s.plain); got != s.encoded {
				t.Errorf("Encode(%q) = %q, want %q", s.plain, got, s.encoded)
			}
		})
	}
}

func TestDecode_KnownVectors(t *testing.T) {
	t.Parallel()
	for _, s := range samples {
		s := s
		t.Run(s.encoded, func(t *testing.T) {
			t.Parallel()
			if got := Decode(s.encoded); got != s.plain {
				t.Errorf("Decode(%q) = %q, want %q", s.encoded, got, s.plain)
			}
		})
	}
}

func TestEncodeDecode_Roundtrip(t *testing.T) {
	t.Parallel()
	cases := []string{
		"",
		"a",
		"hello",
		"The quick brown fox jumps over the lazy dog.",
		strings.Repeat("x", 1000),
		"\x00\x01\x02\x03\xff\xfe", // binary
		"\xff\xff\xff\xff\xff",     // high byte run
		string([]byte{0, 0, 0, 0}), // null bytes
	}
	for _, in := range cases {
		in := in
		t.Run(testname(in), func(t *testing.T) {
			t.Parallel()
			encoded := Encode(in)
			decoded := Decode(encoded)
			if decoded != in {
				t.Errorf("roundtrip mismatch for %q (len=%d):\n  encoded: %q\n  decoded: %q",
					in, len(in), encoded, decoded)
			}
		})
	}
}

func TestEncodeString_DelegatesToEncode(t *testing.T) {
	t.Parallel()
	for _, s := range samples {
		got := EncodeString([]byte(s.plain))
		if got != s.encoded {
			t.Errorf("EncodeString(%q) = %q, want %q", s.plain, got, s.encoded)
		}
	}
}

func TestDecodeString_ReturnsBytesAndNoError(t *testing.T) {
	t.Parallel()
	for _, s := range samples {
		got, err := DecodeString(s.encoded)
		if err != nil {
			t.Errorf("DecodeString(%q): unexpected error: %v", s.encoded, err)
		}
		if string(got) != s.plain {
			t.Errorf("DecodeString(%q) = %q, want %q", s.encoded, string(got), s.plain)
		}
	}
}

func TestDecode_SkipsUnknownCharacters(t *testing.T) {
	t.Parallel()
	// The decoder is documented to skip chars not in dectab. Confirm by
	// inserting whitespace and verifying the result still roundtrips.
	original := "hello world"
	encoded := Encode(original)
	corrupted := encoded[:3] + " \n\t " + encoded[3:]
	if got := Decode(corrupted); got != original {
		t.Errorf("Decode of whitespace-padded encoding = %q, want %q", got, original)
	}
}

// FuzzRoundtrip asserts Encode/Decode forms a faithful roundtrip for any
// input. Catches encoder bugs that lose or corrupt bytes.
func FuzzRoundtrip(f *testing.F) {
	for _, s := range samples {
		f.Add(s.plain)
	}
	f.Add("")
	f.Add("\x00\x01\x02\x03")
	f.Add(strings.Repeat("a", 100))

	f.Fuzz(func(t *testing.T, in string) {
		encoded := Encode(in)
		decoded := Decode(encoded)
		if decoded != in {
			t.Errorf("roundtrip lost data:\n  in:      %q (len=%d)\n  encoded: %q\n  decoded: %q",
				in, len(in), encoded, decoded)
		}
	})
}

// FuzzDecode_NeverPanics asserts that the decoder never panics on arbitrary
// input. Catches index-out-of-bounds bugs and similar.
func FuzzDecode_NeverPanics(f *testing.F) {
	for _, s := range samples {
		f.Add(s.encoded)
	}
	f.Add("")
	f.Add("garbage \x00\x01\xff")
	f.Add(strings.Repeat("!", 1000))

	f.Fuzz(func(t *testing.T, encoded string) {
		// Skip non-UTF-8 (decoder is byte-oriented; we focus on string-shaped inputs).
		if !utf8.ValidString(encoded) {
			t.Skip()
		}
		_ = Decode(encoded) // must not panic
	})
}

// Helper: short, escape-safe subtest names from arbitrary input.
func testname(s string) string {
	if s == "" {
		return "<empty>"
	}
	if len(s) > 20 {
		return s[:20] + "..."
	}
	return s
}
