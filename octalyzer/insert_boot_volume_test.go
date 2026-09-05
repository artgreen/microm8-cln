//go:build !remint
// +build !remint

package main

import (
	gotest "testing"

	"paleotronic.com/core/settings"
)

// TestRecordInsertedBootVolume verifies that a just-inserted volume is recorded
// in the correct PureBoot slot and that the competing primary boot slot is
// cleared, so a following reboot boots exactly the inserted volume.
func TestRecordInsertedBootVolume(t *gotest.T) {
	const idx = 0
	reset := func() {
		settings.PureBootVolume[idx] = ""
		settings.PureBootVolume2[idx] = ""
		settings.PureBootSmartVolume[idx] = ""
	}

	t.Run("high-capacity routes to SmartPort and clears Disk II boot", func(t *gotest.T) {
		reset()
		settings.PureBootVolume[idx] = "local:stale-5.25.dsk" // leftover boot disk
		recordInsertedBootVolume(idx, 0, "local:big.po", true)
		if got := settings.PureBootSmartVolume[idx]; got != "local:big.po" {
			t.Errorf("PureBootSmartVolume = %q, want %q", got, "local:big.po")
		}
		if got := settings.PureBootVolume[idx]; got != "" {
			t.Errorf("PureBootVolume = %q, want cleared", got)
		}
	})

	t.Run("drive 0 low-capacity sets primary boot and clears SmartPort", func(t *gotest.T) {
		reset()
		settings.PureBootSmartVolume[idx] = "local:stale.po" // leftover SmartPort boot
		recordInsertedBootVolume(idx, 0, "local:game.dsk", false)
		if got := settings.PureBootVolume[idx]; got != "local:game.dsk" {
			t.Errorf("PureBootVolume = %q, want %q", got, "local:game.dsk")
		}
		if got := settings.PureBootSmartVolume[idx]; got != "" {
			t.Errorf("PureBootSmartVolume = %q, want cleared", got)
		}
	})

	t.Run("drive 1 low-capacity sets secondary and leaves primary boot alone", func(t *gotest.T) {
		reset()
		settings.PureBootVolume[idx] = "local:boot.dsk"
		recordInsertedBootVolume(idx, 1, "local:data.dsk", false)
		if got := settings.PureBootVolume2[idx]; got != "local:data.dsk" {
			t.Errorf("PureBootVolume2 = %q, want %q", got, "local:data.dsk")
		}
		if got := settings.PureBootVolume[idx]; got != "local:boot.dsk" {
			t.Errorf("PureBootVolume = %q, want left untouched", got)
		}
	})

	reset()
}
