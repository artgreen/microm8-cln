package panic

import (
	"errors"
	"testing"
)

func TestDo_RunsFunctionAndDoesNotCallHandlerWhenNoPanic(t *testing.T) {
	t.Parallel()
	var ran bool
	var handlerCalled bool
	Do(
		func() { ran = true },
		func(_ interface{}) { handlerCalled = true },
	)
	if !ran {
		t.Error("function body did not run")
	}
	if handlerCalled {
		t.Error("handler should not be called when no panic occurs")
	}
}

func TestDo_RecoversAndCallsHandlerWithPanicValue(t *testing.T) {
	t.Parallel()
	want := errors.New("boom")
	var got interface{}
	Do(
		func() { panic(want) },
		func(r interface{}) { got = r },
	)
	if got != want {
		t.Errorf("handler received %v, want %v", got, want)
	}
}

func TestDo_HandlerSeesPanicValue_ForNonErrorTypes(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name  string
		value interface{}
	}{
		{"string", "panic-message"},
		{"int", 42},
		{"struct", struct{ X int }{X: 7}},
		{"nil-but-not-no-panic", "non-nil"},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			var got interface{}
			Do(
				func() { panic(tc.value) },
				func(r interface{}) { got = r },
			)
			if got != tc.value {
				t.Errorf("handler received %v, want %v", got, tc.value)
			}
		})
	}
}

func TestDo_DoesNotPropagatePanicToCaller(t *testing.T) {
	t.Parallel()
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("panic leaked past Do: %v", r)
		}
	}()
	Do(
		func() { panic("contained") },
		func(_ interface{}) {},
	)
}
