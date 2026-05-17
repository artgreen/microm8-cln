package core

import (
	"testing"

	"paleotronic.com/core/hardware/spectrum/snapshot"
)

// TestConfigHasZXState_NilConfigSafe pins the Phase 5 SA5011 fix. The
// original ApplyLaunchConfig had:
//
//	if config.ZXState != nil || strings.Contains(...) { ... }
//	...
//	if config == nil { /* fallback to shell */ }
//
// The first line dereferenced config.ZXState BEFORE the nil check on
// config — so passing nil panicked. The fix moved the nil-aware
// check into a tiny helper that's now testable in isolation.
func TestConfigHasZXState_NilConfigSafe(t *testing.T) {
	t.Parallel()

	t.Run("nil config", func(t *testing.T) {
		t.Parallel()
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("configHasZXState(nil) panicked: %v", r)
			}
		}()
		if configHasZXState(nil) {
			t.Error("configHasZXState(nil) = true, want false")
		}
	})

	t.Run("config without ZXState", func(t *testing.T) {
		t.Parallel()
		if configHasZXState(&VMLauncherConfig{}) {
			t.Error("configHasZXState({}) = true, want false")
		}
	})

	t.Run("config with ZXState", func(t *testing.T) {
		t.Parallel()
		c := &VMLauncherConfig{ZXState: &snapshot.Z80{}}
		if !configHasZXState(c) {
			t.Error("configHasZXState({ZXState: &Z80{}}) = false, want true")
		}
	})
}
