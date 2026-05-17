package update

import (
	"testing"

	"paleotronic.com/core/types"
)

// withDisabled temporarily forces the Disabled flag for the duration of one
// test. We snapshot the original value rather than assuming a default because
// the default could legitimately flip in either direction in the future.
func withDisabled(t *testing.T, v bool) {
	t.Helper()
	prev := Disabled
	Disabled = v
	t.Cleanup(func() { Disabled = prev })
}

// TestDisabledByDefault is the "did anyone accidentally flip the switch?"
// canary. Phase 5.5 set Disabled=true while the migration is in flight.
func TestDisabledByDefault(t *testing.T) {
	if !Disabled {
		t.Fatalf("update.Disabled = false; remote-update helpers must stay off " +
			"during the modernization migration. Flip explicitly in main if you " +
			"want them back.")
	}
}

// TestCheckVersionDisabledReturnsBuildNumber verifies the documented contract
// of CheckVersion when Disabled: skip the network and return the local build
// number, so callers that compare `online > local` see "no upgrade available."
func TestCheckVersionDisabledReturnsBuildNumber(t *testing.T) {
	withDisabled(t, true)
	got := CheckVersion()
	if got != GetBuildNumber() {
		t.Fatalf("CheckVersion() with Disabled=true = %q, want GetBuildNumber()=%q",
			got, GetBuildNumber())
	}
}

// TestGetChecksumDisabledReturnsEmpty: empty string is the well-defined "no
// checksum available" signal that downstream callers already handle
// (DownloadVersion skips the verify step if checksum == "").
func TestGetChecksumDisabledReturnsEmpty(t *testing.T) {
	withDisabled(t, true)
	if got := GetChecksum(); got != "" {
		t.Fatalf("GetChecksum() with Disabled=true = %q, want \"\"", got)
	}
}

// TestDownloadVersionDisabledIsNoop: should not touch the network and not
// blow up on the missing TextBuffer argument. The "Ok" sentinel mirrors the
// nox return path so the caller's UI doesn't surface a fake failure.
func TestDownloadVersionDisabledIsNoop(t *testing.T) {
	withDisabled(t, true)
	status, err := DownloadVersion((*types.TextBuffer)(nil))
	if err != nil {
		t.Fatalf("DownloadVersion() with Disabled=true err = %v, want nil", err)
	}
	if status != "Ok" {
		t.Fatalf("DownloadVersion() with Disabled=true status = %q, want \"Ok\"", status)
	}
}

// TestCheckAndDownloadDisabledIsNoop: returns immediately without sleeping,
// hitting CheckVersion, or scribbling to the TextBuffer. We don't assert on
// timing — just that nil txt doesn't panic.
func TestCheckAndDownloadDisabledIsNoop(t *testing.T) {
	withDisabled(t, true)
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("CheckAndDownload() with Disabled=true panicked: %v", r)
		}
	}()
	CheckAndDownload((*types.TextBuffer)(nil))
}

// TestCheckFilenameDisabledIsNoop: must not rename the binary and must not
// re-exec. The most observable property is "doesn't os.Exit(0)" — if it ran
// to the end it would have. Reaching the assertion below is the success
// signal.
func TestCheckFilenameDisabledIsNoop(t *testing.T) {
	withDisabled(t, true)
	CheckFilename()
	// If we got here, CheckFilename returned cleanly without re-exec'ing.
}

// TestVersionGettersUnaffected confirms the local-only getters keep working
// when Disabled is set; we don't want to accidentally make the version banner
// blank.
func TestVersionGettersUnaffected(t *testing.T) {
	withDisabled(t, true)
	if GetVersion() == "" {
		t.Error("GetVersion() returned empty under Disabled=true")
	}
	if GetBuildNumber() == "" {
		t.Error("GetBuildNumber() returned empty under Disabled=true")
	}
	if GetHumanVersion() == "" {
		t.Error("GetHumanVersion() returned empty under Disabled=true")
	}
}
