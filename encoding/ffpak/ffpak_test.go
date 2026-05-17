package ffpak

import (
	"bytes"
	"encoding/base64"
	"testing"
)

// --- Behavior tests ---------------------------------------------------------

func TestFFPack_PassesThroughPrintableBytes(t *testing.T) {
	t.Parallel()
	in := []byte("hello world")
	out := FFPack(in)
	if !bytes.Equal(out, in) {
		t.Errorf("FFPack(%q) = %q, want %q (no escape bytes for printable input)", in, out, in)
	}
}

func TestFFPack_EscapesControlBytes(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		in   []byte
		want []byte
	}{
		{"single null", []byte{0x00}, []byte{0xff, 0x80}},
		{"control char 0x01", []byte{0x01}, []byte{0xff, 0x81}},
		{"control char 0x1f (last escaped)", []byte{0x1f}, []byte{0xff, 0x9f}},
		{"0xff itself", []byte{0xff}, []byte{0xff, 0x7f}},
		{"space (0x20) is not escaped", []byte{0x20}, []byte{0x20}},
		{"mixed", []byte{'a', 0x00, 'b'}, []byte{'a', 0xff, 0x80, 'b'}},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := FFPack(tc.in)
			if !bytes.Equal(got, tc.want) {
				t.Errorf("FFPack(%v) = %v, want %v", tc.in, got, tc.want)
			}
		})
	}
}

func TestFFUnpack_InvertsFFPack(t *testing.T) {
	t.Parallel()
	cases := [][]byte{
		nil,
		{},
		[]byte("simple"),
		{0x00, 0x01, 0x1f, 0x20, 0x7f, 0xff},
		bytes.Repeat([]byte{0xff}, 16),
		bytes.Repeat([]byte{0x00}, 16),
		[]byte("Mixed\x00content\x01with\x1fcontrol\xffbytes"),
	}
	for i, in := range cases {
		i, in := i, in
		t.Run(testname(i, in), func(t *testing.T) {
			t.Parallel()
			packed := FFPack(in)
			unpacked := FFUnpack(packed)
			if !bytes.Equal(unpacked, in) {
				t.Errorf("roundtrip mismatch:\n  in:       %v\n  packed:   %v\n  unpacked: %v",
					in, packed, unpacked)
			}
		})
	}
}

// FFPack inflates by ~2x for fully-escaped input. Verify the size bound to
// catch any pathological growth introduced by future refactors.
func TestFFPack_NeverGrowsMoreThan2x(t *testing.T) {
	t.Parallel()
	in := bytes.Repeat([]byte{0xff}, 100)
	out := FFPack(in)
	if len(out) != len(in)*2 {
		t.Errorf("FFPack of 0xff*100: got len %d, want %d", len(out), len(in)*2)
	}
}

// --- Fuzz: roundtrip property must hold for arbitrary input -----------------

func FuzzFFPackRoundtrip(f *testing.F) {
	f.Add([]byte(""))
	f.Add([]byte("hello"))
	f.Add([]byte{0x00, 0xff, 0x01, 0xfe})
	f.Add(bytes.Repeat([]byte{0xff}, 32))

	f.Fuzz(func(t *testing.T, in []byte) {
		packed := FFPack(in)
		unpacked := FFUnpack(packed)
		if !bytes.Equal(unpacked, in) {
			t.Errorf("roundtrip lost data:\n  in:       %v\n  packed:   %v\n  unpacked: %v",
				in, packed, unpacked)
		}
	})
}

// --- Existing benchmarks (preserved) ----------------------------------------

func BenchmarkBase64(b *testing.B) {
	str := "this is an input string\r\nand this is line 2"
	for i := 0; i < b.N; i++ {
		es := base64.StdEncoding.EncodeToString([]byte(str))
		_, _ = base64.StdEncoding.DecodeString(es)
	}
}

func BenchmarkFFPack(b *testing.B) {
	str := "this is an input string\r\nand this is line 2"
	for i := 0; i < b.N; i++ {
		es := FFPack([]byte(str))
		_ = FFUnpack(es)
	}
}

// Helper.
func testname(i int, b []byte) string {
	_ = i
	if len(b) == 0 {
		return "empty"
	}
	if len(b) > 16 {
		return shorthex(b[:16]) + "..."
	}
	return shorthex(b)
}

func shorthex(b []byte) string {
	const hex = "0123456789abcdef"
	out := make([]byte, 0, len(b)*2)
	for _, x := range b {
		out = append(out, hex[x>>4], hex[x&0xf])
	}
	return string(out)
}
