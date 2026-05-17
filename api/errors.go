package s8webclient

import (
	"errors"

	"paleotronic.com/fmt"
)

// Sentinel errors returned by the API client. Callers should match with
// errors.Is rather than string comparison so that wrapping with additional
// context (server payload, operation name) still matches:
//
//	if errors.Is(err, s8webclient.ErrTimeout) { ... }
//
// All wrapping in this package goes through fmt.Errorf with the %w verb,
// or NewServerError which embeds an Unwrap target.
var (
	// ErrTimeout is returned when a server response doesn't arrive within
	// the request's timeout window.
	ErrTimeout = errors.New("timeout")

	// ErrIO is returned when the server sends an "ERR" message with no
	// payload (i.e. we know it failed but the server didn't say why).
	ErrIO = errors.New("i/o error")

	// ErrNotConnected indicates the underlying DuckTape connection is nil
	// or has been torn down.
	ErrNotConnected = errors.New("not connected")

	// ErrBadStatus is returned when a server response carries an
	// unexpected status code in the message header.
	ErrBadStatus = errors.New("bad status")

	// ErrUnexpectedMessage indicates the server returned a message ID we
	// didn't expect for the current request.
	ErrUnexpectedMessage = errors.New("unexpected message")

	// ErrServerFailure is the generic "the server said no" sentinel used
	// when the server response carries no further detail (legacy
	// `errors.New("failed")`).
	ErrServerFailure = errors.New("server failure")

	// ─── dbapi sentinels ────────────────────────────────────────────────

	// ErrInvalidDBHandle is returned from dbapi when a caller passes a
	// handle that doesn't exist on the server.
	ErrInvalidDBHandle = errors.New("invalid DB handle")

	// ErrNoStatementHandle is returned from dbapi when a caller passes a
	// statement handle that doesn't exist on the server.
	ErrNoStatementHandle = errors.New("no such statement handle")

	// ErrNoMoreResults is returned from dbapi when a result-set iterator
	// has been exhausted.
	ErrNoMoreResults = errors.New("no more results")

	// ─── operation-failed sentinels (filecache / userfiles / spaces) ────

	// ErrDeleteFailed is returned when a server-side delete fails.
	ErrDeleteFailed = errors.New("delete failed")

	// ErrLockFailed is returned when a server-side lock acquisition fails.
	ErrLockFailed = errors.New("lock failed")

	// ErrMkdirFailed is returned when a server-side mkdir fails.
	ErrMkdirFailed = errors.New("mkdir failed")

	// ErrMetaUpdateFailed is returned when a server-side metadata update fails.
	ErrMetaUpdateFailed = errors.New("meta update failed")

	// ErrProjectCreateFailed is returned when project creation fails.
	ErrProjectCreateFailed = errors.New("project create failed")

	// ErrProjectListFailed is returned when project list fetch fails.
	ErrProjectListFailed = errors.New("project fetch list failed")

	// ErrProjectStatusFailed is returned when project status fetch fails.
	ErrProjectStatusFailed = errors.New("project status failed")

	// ErrRegistrationFailed is returned when user registration fails. The
	// server's payload may carry a human-readable explanation; in that
	// case the wrapped error is a *ServerError whose Unwrap() returns
	// ErrRegistrationFailed.
	ErrRegistrationFailed = errors.New("registration failed")
)

// ServerError wraps an error response from the server. It is returned
// whenever the server's "ERR" message carries a payload we want to surface
// to callers without losing the ability to match a typed sentinel.
//
// Implements errors.Is / errors.Unwrap via the Wrapped field, so:
//
//	err := someAPICall()
//	if errors.Is(err, s8webclient.ErrRegistrationFailed) {
//	    // handle registration failure
//	}
//	var se *s8webclient.ServerError
//	if errors.As(err, &se) {
//	    log.Printf("server said: %s", se.Payload)
//	}
type ServerError struct {
	// Op is the operation name (typically the request message ID).
	Op string
	// Payload is the server's raw error response body.
	Payload []byte
	// Wrapped is an optional underlying sentinel that callers can match
	// with errors.Is (e.g. ErrRegistrationFailed).
	Wrapped error
}

// Error renders the user-visible message. When the error is wrapping a
// sentinel and carries a payload, the format is "<sentinel>: <payload>"
// (matching the pre-migration `errors.New("X failed: " + payload)` shape).
// When there's no sentinel, the payload alone is the message — this
// preserves the original `errors.New(string(msg.Payload))` behaviour so
// callers that surface err.Error() to users see no change. Op is
// structured metadata available via errors.As; it intentionally does NOT
// appear in Error() to avoid noisy "REG: details" messages.
func (e *ServerError) Error() string {
	if e.Wrapped != nil {
		if len(e.Payload) > 0 {
			return fmt.Sprintf("%s: %s", e.Wrapped.Error(), e.Payload)
		}
		return e.Wrapped.Error()
	}
	if len(e.Payload) > 0 {
		return string(e.Payload)
	}
	return e.Op
}

// Unwrap returns the wrapped sentinel for use with errors.Is. Returns nil
// when no sentinel was supplied.
func (e *ServerError) Unwrap() error {
	return e.Wrapped
}

// NewServerError returns a *ServerError wrapping the given operation
// name, server payload, and (optional) underlying sentinel. Pass a
// non-nil wrapped sentinel to allow errors.Is matching; pass nil for the
// generic "server returned this opaque payload" case.
func NewServerError(op string, payload []byte, wrapped error) error {
	return &ServerError{Op: op, Payload: payload, Wrapped: wrapped}
}
