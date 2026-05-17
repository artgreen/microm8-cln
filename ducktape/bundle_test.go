package ducktape

import (
	"bytes"
	"testing"
)

// TestDuckTapeBundle_UnmarshalBinary_HappyText: text-mode (non-binary)
// messages: ID is 3 chars, byte 3 is a space, payload is the rest.
func TestDuckTapeBundle_UnmarshalBinary_HappyText(t *testing.T) {
	t.Parallel()
	d := &DuckTapeBundle{}
	err := d.UnmarshalBinary([]byte("LGN hello world"))
	if err != nil {
		t.Fatalf("UnmarshalBinary: %v", err)
	}
	if d.ID != "LGN" {
		t.Errorf("ID = %q, want LGN", d.ID)
	}
	if d.Binary {
		t.Error("text-mode bundle reports Binary=true")
	}
	if string(d.Payload) != "hello world" {
		t.Errorf("Payload = %q, want %q", d.Payload, "hello world")
	}
}

// TestDuckTapeBundle_UnmarshalBinary_ShortReturnsError: < 4 bytes is
// the documented error path.
func TestDuckTapeBundle_UnmarshalBinary_ShortReturnsError(t *testing.T) {
	t.Parallel()
	cases := [][]byte{
		nil,
		{},
		[]byte("A"),
		[]byte("AB"),
		[]byte("ABC"), // 3 bytes — still short
	}
	for _, in := range cases {
		d := &DuckTapeBundle{}
		if err := d.UnmarshalBinary(in); err == nil {
			t.Errorf("UnmarshalBinary(%q) returned nil error, want short-input error", in)
		}
	}
}

// TestDuckTapeBundle_UnmarshalBinary_BinaryEmpty: a binary-mode header
// with no payload bytes is valid (empty payload).
func TestDuckTapeBundle_UnmarshalBinary_BinaryEmpty(t *testing.T) {
	t.Parallel()
	d := &DuckTapeBundle{}
	err := d.UnmarshalBinary([]byte("FOO#"))
	if err != nil {
		t.Fatalf("UnmarshalBinary: %v", err)
	}
	if d.ID != "FOO" {
		t.Errorf("ID = %q, want FOO", d.ID)
	}
	if !d.Binary {
		t.Error("Binary = false, want true")
	}
	if len(d.Payload) != 0 {
		t.Errorf("Payload = %v, want empty", d.Payload)
	}
}

// FuzzDuckTapeBundle_UnmarshalBinary_NeverPanics is the contract test:
// the wire-format parser must reject malformed input cleanly (return
// an error) and never panic, regardless of input.
//
// Seeds:
//   - empty / nil
//   - short headers (1-3 bytes)
//   - text-mode bundles with various payloads
//   - binary-mode (4th byte = '#') with corrupt / truncated / overflow
//     base64 bodies
//   - very long inputs (size-bomb resistance)
func FuzzDuckTapeBundle_UnmarshalBinary_NeverPanics(f *testing.F) {
	// Happy-path seeds.
	f.Add([]byte("LGN hello"))
	f.Add([]byte("MSG "))
	f.Add([]byte("FOO#"))
	// Boundaries.
	f.Add([]byte(""))
	f.Add([]byte("A"))
	f.Add([]byte("AB"))
	f.Add([]byte("ABC"))
	f.Add([]byte("ABCD"))
	// Binary mode with garbage payload.
	f.Add([]byte("BIN#@@@invalid_base64@@@"))
	f.Add([]byte("BIN#" + string(bytes.Repeat([]byte("="), 64))))
	// Large.
	f.Add(append([]byte("BIG "), bytes.Repeat([]byte{0xff}, 10000)...))

	f.Fuzz(func(t *testing.T, data []byte) {
		d := &DuckTapeBundle{}
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("UnmarshalBinary(%q) panicked: %v", data, r)
			}
		}()
		_ = d.UnmarshalBinary(data) // error is fine; panic is not
	})
}

// TestDuckTapeBundle_RoundTrip_TextMode is the happy-path round-trip:
// marshal then unmarshal returns the same fields. Tests the wire
// format itself.
func TestDuckTapeBundle_RoundTrip_TextMode(t *testing.T) {
	t.Parallel()
	original := DuckTapeBundle{
		ID:      "MSG",
		Payload: []byte("hello world"),
		Binary:  false,
	}
	encoded, err := original.MarshalBinaryUDP()
	if err != nil {
		t.Fatalf("MarshalBinaryUDP: %v", err)
	}

	decoded := &DuckTapeBundle{}
	if err := decoded.UnmarshalBinary(encoded); err != nil {
		t.Fatalf("UnmarshalBinary: %v", err)
	}
	if decoded.ID != original.ID {
		t.Errorf("ID: got %q, want %q", decoded.ID, original.ID)
	}
	if decoded.Binary != original.Binary {
		t.Errorf("Binary: got %v, want %v", decoded.Binary, original.Binary)
	}
	if !bytes.Equal(decoded.Payload, original.Payload) {
		t.Errorf("Payload: got %q, want %q", decoded.Payload, original.Payload)
	}
}

// TestDuckTapeBundle_MarshalBinaryUDP_RejectsBadID: ID must be exactly
// 3 chars. Anything else is an error rather than producing garbage.
func TestDuckTapeBundle_MarshalBinaryUDP_RejectsBadID(t *testing.T) {
	t.Parallel()
	for _, badID := range []string{"", "A", "AB", "ABCD", "TOOLONG"} {
		d := DuckTapeBundle{ID: badID}
		if _, err := d.MarshalBinaryUDP(); err == nil {
			t.Errorf("MarshalBinaryUDP with ID=%q returned nil error", badID)
		}
	}
}
