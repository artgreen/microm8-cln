package fmt

import (
	"testing"
)

// Sprintf is the only function in this package that doesn't gate on the
// global settings.Verbose. The Print* family writes to os.Stdout when verbose
// and is otherwise a no-op; we don't currently test those because they touch
// process-global state (the settings package + os.Stdout) and the package is
// thin enough that the indirection has no behavior worth catching.
//
// If/when we refactor this package to take a writer parameter, expand tests
// to cover Println/Printf/etc.

func TestSprintf_DelegatesToStdlibFmt(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name   string
		format string
		args   []interface{}
		want   string
	}{
		{"plain string", "hello", nil, "hello"},
		{"single %s", "hello %s", []interface{}{"world"}, "hello world"},
		{"int + string", "%d items in %s", []interface{}{3, "queue"}, "3 items in queue"},
		{"percent literal", "100%%", nil, "100%"},
		{"missing arg", "%s", nil, "%!s(MISSING)"},
		{"empty format", "", nil, ""},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := Sprintf(tc.format, tc.args...)
			if got != tc.want {
				t.Errorf("Sprintf(%q, %v) = %q, want %q", tc.format, tc.args, got, tc.want)
			}
		})
	}
}
