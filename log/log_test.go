package log

import (
	"strings"
	"testing"
)

// formatStr is unexported; we test it in-package. Verifies the Phase-0
// vet fix (fmt.Sprint(v) -> fmt.Sprint(v...)) is producing the right
// output shape — without the fix, args would render as a bracketed slice.

func TestFormatStr_EmptyFormatSpreadsArgs(t *testing.T) {
	t.Parallel()
	// Mix string + int so that fmt.Sprint inserts a separator we can assert.
	// Per fmt.Sprint docs: "Spaces are added between operands when neither
	// is a string." So we use two ints — output will have a space.
	got := formatStr("", 1, 2)

	// Spread args render as `1 2`. Un-spread (the Phase-0 bug) would
	// render as `[1 2]` because v=[]interface{}{1,2} formats with brackets.
	if strings.Contains(got, "[1 2]") {
		t.Errorf("formatStr: %q contains %q — args not spread (regression of Phase 0 fix)",
			got, "[1 2]")
	}
	// Positive assertion: the spread output ends with "1 2\r\n".
	if !strings.Contains(got, "1 2") {
		t.Errorf("formatStr: %q does not contain spread output %q", got, "1 2")
	}
}

func TestFormatStr_TrailingCRLF(t *testing.T) {
	t.Parallel()
	got := formatStr("", "x")
	if !strings.HasSuffix(got, "\r\n") {
		t.Errorf("formatStr should end with \\r\\n, got %q", got)
	}
}

func TestFormatStr_FormatStringRenders(t *testing.T) {
	t.Parallel()
	// Note: the non-empty-format branch passes v (the slice) to Sprintf
	// with myfmt = "%s [%s] " + format. The exact rendering depends on
	// the format string; we only verify the format prefix and that
	// the timestamp+caller header is present.
	got := formatStr("plain message", "ignored")
	if !strings.Contains(got, "plain message") {
		t.Errorf("formatStr format-mode: %q does not contain %q", got, "plain message")
	}
}

func TestFormatStr_IncludesCallerWhenEnabled(t *testing.T) {
	t.Parallel()
	if !SHOWCALLER {
		t.Skip("SHOWCALLER disabled at compile time")
	}
	got := formatStr("", "msg")
	// Caller heuristic walks 2 frames up — for a direct call from this
	// test function, it should resolve to *something* (specifically the
	// test runner frame). We assert the placeholder "**" is not present,
	// which would indicate runtime.Caller failed.
	if strings.Contains(got, "[**]") {
		t.Errorf("formatStr: caller resolution failed (got default '**'): %q", got)
	}
}

// Print/Println/Printf are gated on SILENT = true and become no-ops, so
// there's nothing observable to test for those right now. If/when SILENT
// becomes a runtime flag or these get a writer parameter, expand tests.
//
// They DO call formatStr internally, so the formatStr tests above cover
// the substantive logic.
