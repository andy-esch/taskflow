package domain

// FileProblem is a non-fatal per-file load problem: the file was skipped, with
// an actionable message explaining what's wrong and how to fix it.
type FileProblem struct {
	Path    string `json:"path"`
	Message string `json:"message"`
	// EntityID and EntitySlug are optional canonical identity recovered by a
	// persistence adapter even when the document itself could not be decoded.
	// They let application-level consistency checks include unreadable records
	// without parsing Path for meaning. They are diagnostic plumbing, not part of
	// the public unreadable-file wire contract.
	EntityID   string `json:"-" yaml:"-"`
	EntitySlug string `json:"-" yaml:"-"`
}

// FixResult records the auto-repairs applied (or proposed) for one file.
type FixResult struct {
	Path    string   `json:"path"`
	Changes []string `json:"changes"`
	// Skipped marks a file the pass deliberately did NOT repair, with Changes carrying the
	// reason. Reported alongside repairs because the reason is the useful part — "this id
	// is still referenced by three files" is what the operator has to act on — but counted
	// separately, since calling a refusal a fix is how a tool loses trust.
	Skipped bool `json:"skipped,omitempty"`
}
