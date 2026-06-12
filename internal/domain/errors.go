package domain

import "errors"

// Sentinel errors. The CLI maps these to semantic exit codes so agents can
// route on the code without parsing text.
//
// Exit code 12 (invalid-transition) was retired 2026-06-12: no transition
// rules exist — any status→status move is legal — so its sentinel was a dead
// documented contract. 13/14 keep their numbers; 12 stays reserved.
var (
	ErrNotFound   = errors.New("not found")
	ErrAmbiguous  = errors.New("ambiguous match")
	ErrValidation = errors.New("validation failed")
	ErrConflict   = errors.New("conflict") // already exists / write conflict
)
