package s8webclient

import (
	"errors"
	"fmt"
	"testing"
)

// TestSentinels_AreDistinct guarantees we don't accidentally alias two
// sentinels. errors.Is on Go's `errors.errorString` falls back to pointer
// equality, so reusing the same `errors.New("X")` for two vars would silently
// collapse them. Hold our exported sentinels to that pointer-distinct
// contract.
func TestSentinels_AreDistinct(t *testing.T) {
	t.Parallel()
	all := []error{
		ErrTimeout,
		ErrIO,
		ErrNotConnected,
		ErrBadStatus,
		ErrUnexpectedMessage,
		ErrServerFailure,
		ErrInvalidDBHandle,
		ErrNoStatementHandle,
		ErrNoMoreResults,
		ErrDeleteFailed,
		ErrLockFailed,
		ErrMkdirFailed,
		ErrMetaUpdateFailed,
		ErrProjectCreateFailed,
		ErrProjectListFailed,
		ErrProjectStatusFailed,
		ErrRegistrationFailed,
	}
	for i, a := range all {
		for j, b := range all {
			if i == j {
				continue
			}
			if errors.Is(a, b) {
				t.Errorf("%v should not match %v under errors.Is", a, b)
			}
		}
	}
}

// TestSentinels_SelfMatch is the "obvious" test: every sentinel must match
// itself via errors.Is, even when wrapped through fmt.Errorf("%w").
func TestSentinels_SelfMatch(t *testing.T) {
	t.Parallel()
	cases := []error{
		ErrTimeout, ErrIO, ErrNotConnected, ErrBadStatus,
		ErrUnexpectedMessage, ErrServerFailure,
		ErrInvalidDBHandle, ErrNoStatementHandle, ErrNoMoreResults,
		ErrDeleteFailed, ErrLockFailed, ErrMkdirFailed,
		ErrMetaUpdateFailed, ErrProjectCreateFailed,
		ErrProjectListFailed, ErrProjectStatusFailed,
		ErrRegistrationFailed,
	}
	for _, sentinel := range cases {
		sentinel := sentinel
		t.Run(sentinel.Error(), func(t *testing.T) {
			t.Parallel()
			if !errors.Is(sentinel, sentinel) {
				t.Fatalf("errors.Is(%v, %v) returned false on identity", sentinel, sentinel)
			}
			wrapped := fmt.Errorf("context: %w", sentinel)
			if !errors.Is(wrapped, sentinel) {
				t.Fatalf("errors.Is unwrap failed for %v", sentinel)
			}
		})
	}
}

// TestServerError_PreservesPayloadMessage covers the most-load-bearing
// migration: the legacy `errors.New(string(msg.Payload))` returned an error
// whose Error() was exactly the payload. We must not introduce an "OP:"
// prefix into that path because anything that surfaces err.Error() to a UI
// (notably the BASIC dialect's PRINT-of-error) would suddenly show
// "GRI: file not found" instead of "file not found".
func TestServerError_PreservesPayloadMessage(t *testing.T) {
	t.Parallel()
	payload := []byte("user not found")
	err := NewServerError("GRI", payload, nil)
	if got, want := err.Error(), "user not found"; got != want {
		t.Fatalf("Error() = %q, want %q (Op must not leak into the user message when Wrapped is nil)", got, want)
	}
}

// TestServerError_WrappedFormatIncludesSentinel mirrors the legacy
// `errors.New("registration failed: PAYLOAD")` shape. When we wrap a typed
// sentinel, Error() should render as "<sentinel>: <payload>" so callers
// see the same message and ALSO get errors.Is matching for free.
func TestServerError_WrappedFormatIncludesSentinel(t *testing.T) {
	t.Parallel()
	payload := []byte("username taken")
	err := NewServerError("REG", payload, ErrRegistrationFailed)

	if got, want := err.Error(), "registration failed: username taken"; got != want {
		t.Fatalf("Error() = %q, want %q", got, want)
	}
	if !errors.Is(err, ErrRegistrationFailed) {
		t.Error("errors.Is(err, ErrRegistrationFailed) = false; sentinel matching is the whole point of the wrap")
	}
	// And the typed inspection path:
	var se *ServerError
	if !errors.As(err, &se) {
		t.Fatal("errors.As did not extract *ServerError")
	}
	if se.Op != "REG" {
		t.Errorf("ServerError.Op = %q, want REG", se.Op)
	}
	if string(se.Payload) != "username taken" {
		t.Errorf("ServerError.Payload = %q, want %q", se.Payload, "username taken")
	}
}

// TestServerError_EmptyPayload tests the corner case where the server sent
// us an "ERR" with no payload. We still want a non-empty error message.
func TestServerError_EmptyPayload(t *testing.T) {
	t.Parallel()
	t.Run("op only", func(t *testing.T) {
		t.Parallel()
		err := NewServerError("CPR", nil, nil)
		if got := err.Error(); got != "CPR" {
			t.Errorf("Error() = %q, want %q", got, "CPR")
		}
	})
	t.Run("wrapped sentinel, no payload", func(t *testing.T) {
		t.Parallel()
		err := NewServerError("REG", nil, ErrRegistrationFailed)
		if got := err.Error(); got != "registration failed" {
			t.Errorf("Error() = %q, want %q", got, "registration failed")
		}
	})
}

// TestServerError_NilWrapped covers the explicit-nil contract: passing nil
// to NewServerError should leave Unwrap() returning nil and errors.Is
// returning false against arbitrary sentinels.
func TestServerError_NilWrapped(t *testing.T) {
	t.Parallel()
	err := NewServerError("X", []byte("y"), nil)
	if u := errors.Unwrap(err); u != nil {
		t.Errorf("Unwrap() = %v, want nil", u)
	}
	if errors.Is(err, ErrTimeout) {
		t.Error("errors.Is matched a non-wrapped ServerError against ErrTimeout")
	}
}
