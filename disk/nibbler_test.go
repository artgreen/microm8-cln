package disk

import (
	"bytes"
	"testing"
)

// FuzzNewDSKWrapperBin_NeverPanics asserts the disk image
// identification routine doesn't crash on arbitrary input. DSK files
// come from user disk, from network, from inside zip archives — any
// hostile or corrupt image must produce an error, not a crash.
func FuzzNewDSKWrapperBin_NeverPanics(f *testing.F) {
	// Seeds: each documented-valid size + a few bogus sizes.
	f.Add(bytes.Repeat([]byte{0xff}, STD_DISK_BYTES)) // 143360
	f.Add(make([]byte, 0))
	f.Add(make([]byte, 16))
	f.Add(make([]byte, 1024))
	f.Add(bytes.Repeat([]byte{0xAA}, STD_DISK_BYTES+64)) // valid alt size
	f.Add(bytes.Repeat([]byte{0x55}, PRODOS_400KB_DISK_BYTES))

	f.Fuzz(func(t *testing.T, data []byte) {
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("NewDSKWrapperBin panicked on %d-byte input: %v", len(data), r)
			}
		}()
		// nibbler can be nil — the format identification doesn't use it
		// for the cases reachable from arbitrary input.
		_, _ = NewDSKWrapperBin(nil, data, "fuzz.dsk")
	})
}
