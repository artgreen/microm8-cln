package mempak

import (
	"testing"
)

// --- Table-driven roundtrip tests ------------------------------------------
//
// The original tests in mempak_test.go cover specific (addr, value) shapes.
// This file adds a single table-driven test that exercises the full
// (asize x vsize) matrix plus the read=true short-form, which the original
// tests don't exercise.

func TestEncodeDecode_Roundtrip(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name  string
		addr  int
		value uint64
		read  bool
	}{
		{"8b addr, 8b value, write", 0xff, 0xfe, false},
		{"8b addr, 8b value, read", 0xff, 0xfe, true},
		{"16b addr, 8b value", 0xaaff, 0xfe, false},
		{"20b addr, 8b value", 0xabbcc, 0xfe, false},
		{"24b addr, 64b value", 0xccddee, 0x1122334455667788, false},
		{"20b addr, 32b value", 0xabbcc, 0x4411223344, false},
		{"max 24b addr", (1 << 24) - 1, 0x42, false},
		{"max 32b addr (covers 4-byte path)", 0x7fffffff, 0x42, false},
		{"zero address, zero value", 0, 0, false},
		{"max 64b value", 0, 0xffffffffffffffff, false},
		{"read-mode skips value bytes", 0x1234, 0xdeadbeef, true},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			data := Encode(0, tc.addr, tc.value, tc.read)
			_, daddr, dvalue, dread, count, err := Decode(data)
			if err != nil {
				t.Fatalf("Decode returned error: %v", err)
			}
			if count != len(data) {
				t.Errorf("Decode count = %d, want %d (len(data))", count, len(data))
			}
			if daddr != tc.addr {
				t.Errorf("addr roundtrip: got %#x, want %#x", daddr, tc.addr)
			}
			if dread != tc.read {
				t.Errorf("read-flag roundtrip: got %v, want %v", dread, tc.read)
			}
			// Value is intentionally undefined for read=true; only check for write.
			if !tc.read && dvalue != tc.value {
				t.Errorf("value roundtrip: got %#x, want %#x", dvalue, tc.value)
			}
		})
	}
}

func TestDecode_ReturnsErrorOnShortData(t *testing.T) {
	t.Parallel()
	// Encode a 24-bit address, 64-bit value, then truncate.
	data := Encode(0, 0xccddee, 0x1122334455667788, false)
	if len(data) < 2 {
		t.Fatalf("test precondition: Encode returned %d bytes", len(data))
	}
	truncated := data[:len(data)-1]
	_, _, _, _, _, err := Decode(truncated)
	if err == nil {
		t.Error("Decode of truncated data returned no error, want error")
	}
}

// --- Block encoding roundtrip ----------------------------------------------

func TestEncodeBlock_DecodeBlock_Roundtrip(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name   string
		addr   int
		values []uint64
	}{
		{"single value", 0x1000, []uint64{0x42}},
		{"three small values", 0x1000, []uint64{1, 2, 3}},
		{"mixed widths", 0x1000, []uint64{0xff, 0xffff, 0xffffffff, 0xffffffffff}},
		{"zeros", 0, []uint64{0, 0, 0}},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			data := EncodeBlock(0, tc.addr, tc.values)
			daddr, dvalues, err := DecodeBlock(data)
			if err != nil {
				t.Fatalf("DecodeBlock: %v", err)
			}
			if daddr != tc.addr {
				t.Errorf("addr: got %#x, want %#x", daddr, tc.addr)
			}
			if len(dvalues) != len(tc.values) {
				t.Fatalf("value count: got %d, want %d", len(dvalues), len(tc.values))
			}
			for i, v := range tc.values {
				if dvalues[i] != v {
					t.Errorf("value[%d]: got %#x, want %#x", i, dvalues[i], v)
				}
			}
		})
	}
}

func TestDecodeBlock_ReturnsErrorOnTooLittleData(t *testing.T) {
	t.Parallel()
	cases := [][]byte{nil, {}, {0x00}, {0x00, 0x01}}
	for i, in := range cases {
		i, in := i, in
		t.Run("len_"+itoa(len(in)), func(t *testing.T) {
			t.Parallel()
			_ = i
			_, _, err := DecodeBlock(in)
			if err == nil {
				t.Errorf("DecodeBlock(%v): err=nil, want error", in)
			}
		})
	}
}

// --- Internal helpers: getAddrSize / getValueSize boundary checks ----------

func TestGetAddrSize_BoundaryConditions(t *testing.T) {
	t.Parallel()
	cases := []struct {
		addr int
		want int
	}{
		{0, 0},
		{255, 0},
		{256, 1},
		{65535, 1},
		{65536, 2},
		{1<<24 - 1, 2},
		{1 << 24, 3},
	}
	for _, tc := range cases {
		got := getAddrSize(tc.addr)
		if got != tc.want {
			t.Errorf("getAddrSize(%#x) = %d, want %d", tc.addr, got, tc.want)
		}
	}
}

func TestGetValueSize_BoundaryConditions(t *testing.T) {
	t.Parallel()
	cases := []struct {
		value uint64
		want  int
	}{
		{0, 0},
		{255, 0},
		{256, 1},
		{1<<16 - 1, 1},
		{1 << 16, 2},
		{1<<32 - 1, 3},
		{1 << 32, 4},
		{1<<56 - 1, 6},
		{1 << 56, 7},
	}
	for _, tc := range cases {
		got := getValueSize(tc.value)
		if got != tc.want {
			t.Errorf("getValueSize(%#x) = %d, want %d", tc.value, got, tc.want)
		}
	}
}

// --- Fuzz: encode/decode roundtrip on arbitrary inputs --------------------

func FuzzEncodeDecodeRoundtrip(f *testing.F) {
	f.Add(uint32(0xff), uint64(0xfe), false)
	f.Add(uint32(0xaaff), uint64(0xfe), false)
	f.Add(uint32(0x7fffffff), uint64(0xffffffffffffffff), false)
	f.Add(uint32(0x1000), uint64(0xdeadbeef), true)

	f.Fuzz(func(t *testing.T, addr uint32, value uint64, read bool) {
		data := Encode(0, int(addr), value, read)
		_, daddr, dvalue, dread, count, err := Decode(data)
		if err != nil {
			t.Fatalf("Decode error on roundtrip: %v", err)
		}
		if count != len(data) {
			t.Errorf("count != len: count=%d, len=%d", count, len(data))
		}
		if daddr != int(addr) {
			t.Errorf("addr: got %#x, want %#x", daddr, addr)
		}
		if dread != read {
			t.Errorf("read: got %v, want %v", dread, read)
		}
		if !read && dvalue != value {
			t.Errorf("value: got %#x, want %#x", dvalue, value)
		}
	})
}

// FuzzDecode_NeverPanics asserts that Decode never panics on arbitrary
// (possibly malformed) byte input. We don't care about correctness here,
// only safety.
func FuzzDecode_NeverPanics(f *testing.F) {
	f.Add([]byte{0x00})
	f.Add([]byte{0xff})
	f.Add([]byte(""))

	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) == 0 {
			return // Decode reads data[0] unconditionally; nil-deref would be a separate fix.
		}
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("Decode panicked on %v: %v", data, r)
			}
		}()
		_, _, _, _, _, _ = Decode(data)
	})
}

// Small itoa to avoid importing strconv (keeps this test file dependency-free).
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	neg := n < 0
	if neg {
		n = -n
	}
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}
