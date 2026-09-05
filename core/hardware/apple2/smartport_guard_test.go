package apple2

import (
	"bytes"
	"testing"
)

// TestHas2IMGMagic guards the 4-byte magic read against short/truncated images
// (regression for the SmartPort slice-bounds panic on undersized -drive1 files).
func TestHas2IMGMagic(t *testing.T) {
	cases := []struct {
		name string
		data []byte
		want bool
	}{
		{"nil", nil, false},
		{"empty", []byte{}, false},
		{"one byte", []byte("2"), false},
		{"three bytes (prefix of magic)", []byte("2IM"), false},
		{"exactly magic", []byte("2IMG"), true},
		{"magic with trailer", append([]byte("2IMG"), 0, 0, 0, 0), true},
		{"non-2img headerless", []byte("PRODOSBLOCKS...."), false},
	}
	for _, c := range cases {
		if got := has2IMGMagic(c.data); got != c.want {
			t.Errorf("%s: has2IMGMagic(%q) = %v, want %v", c.name, c.data, got, c.want)
		}
	}
}

// TestNewSmartPortBlockDeviceRejectsTruncated ensures a too-short 2IMG image is
// rejected with an error rather than panicking while decoding its header.
func TestNewSmartPortBlockDeviceRejectsTruncated(t *testing.T) {
	// Starts with the magic but far shorter than the 64-byte header.
	short := append([]byte("2IMG"), bytes.Repeat([]byte{0}, 8)...)
	if _, err := NewSmartPortBlockDevice(short, "local:truncated.2mg"); err == nil {
		t.Fatalf("expected error for truncated 2IMG image, got nil")
	}
}
