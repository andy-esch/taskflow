package domain

import "errors"

// Sentinel errors. The CLI maps these to semantic exit codes so agents can
// route on the code without parsing text.
var (
	ErrNotFound          = errors.New("task not found")
	ErrAmbiguous         = errors.New("ambiguous match")
	ErrValidation        = errors.New("validation failed")
	ErrInvalidTransition = errors.New("invalid transition")
)
