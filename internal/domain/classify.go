package domain

import (
	"errors"
	"strings"
)

// Class is a domain error's outcome category — the single, adapter-neutral
// classification of a sentinel-wrapped error. It lives here, beside the sentinels
// it reads (errors.go), so every consumer maps it to its own surface without
// importing a sibling adapter: the CLI → exit code + JSON `code` name
// (cli/exit.go), the TUI → {inline field error, reload-and-retry, red flash}
// (tui/model.go), and a future web adapter → HTTP status (404/409/422). Before
// this, the CLI owned the only errors.Is table and the TUI re-derived outcomes
// from raw error text — see audit H4.
type Class int

const (
	// ClassUnknown is anything not wrapping a domain sentinel (incl. a nil error,
	// which has no failure class). Adapters map it to their generic failure.
	ClassUnknown Class = iota
	ClassNotFound
	ClassValidation
	ClassAmbiguous
	ClassConflict
)

// Classify reports a domain error's outcome category via errors.Is, so a wrapped
// sentinel (the `%w: …` form every service/store error uses) classifies the same
// as the bare sentinel. A nil error or one wrapping no sentinel is ClassUnknown.
//
// Conflict is checked before validation: the compare-and-swap "changed on disk"
// error (store/fsstore.go) wraps ErrConflict, and an adapter must be able to route
// it to a reload rather than treating it as a fix-in-place validation failure.
func Classify(err error) Class {
	switch {
	case err == nil:
		return ClassUnknown
	case errors.Is(err, ErrNotFound):
		return ClassNotFound
	case errors.Is(err, ErrConflict):
		return ClassConflict
	case errors.Is(err, ErrAmbiguous):
		return ClassAmbiguous
	case errors.Is(err, ErrValidation):
		return ClassValidation
	default:
		return ClassUnknown
	}
}

// classSentinels pairs each classifiable Class with its sentinel, so Reason can
// strip the exact `<sentinel>: ` prefix the error wraps (not merely the first
// ": ", which would mangle a reason that itself contains a colon, e.g. "revisit
// offset %q: %v").
var classSentinels = map[Class]error{
	ClassNotFound:   ErrNotFound,
	ClassConflict:   ErrConflict,
	ClassAmbiguous:  ErrAmbiguous,
	ClassValidation: ErrValidation,
}

// Reason returns the human-facing detail of a sentinel-wrapped error with the
// sentinel prefix stripped — e.g. `fmt.Errorf("%w: at least one tag is required",
// ErrValidation)` yields "at least one tag is required". The service/store errors
// are uniformly wrapped as `<sentinel>: <reason>` (errorf with a leading `%w`), so
// the exact `"<sentinel>: "` prefix is removed; a reason that itself contains a
// colon survives intact. An error with no classifiable sentinel (ClassUnknown) is
// returned verbatim.
//
// This is the single, tested home for the unwrap the TUI used to do inline with a
// raw strings.TrimPrefix on ErrValidation (audit H4): the prefix-coupling lives
// here once instead of at each surface.
func Reason(err error) string {
	if err == nil {
		return ""
	}
	msg := err.Error()
	if sentinel, ok := classSentinels[Classify(err)]; ok {
		return strings.TrimPrefix(msg, sentinel.Error()+": ")
	}
	return msg
}
