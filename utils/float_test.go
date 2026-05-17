package utils

import (
	"testing"
)

func TestStrToFloatStrApple(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"zero", "0", "0"},
		{"zero literal point", "0.0", "0"},
		{"small positive", "1.5", "1.5"},
		{"small negative", "-1.5", "-1.5"},
		// 33.199999999 in float64 rounds to ~33.2 in significant-digit
		// formatting (9 digits).
		{"truncates trailing 9s", "33.199999999", "33.2"},
		// Very large numbers switch to exponential 'g' format.
		{"very large", "1e10", "1e+10"},
		// Very small numbers switch to exponential 'e' format.
		{"very small", "1e-5", "1e-05"},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := StrToFloatStrApple(tc.in)
			if got != tc.want {
				t.Errorf("StrToFloatStrApple(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}
