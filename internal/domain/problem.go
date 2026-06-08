package domain

// FileProblem is a non-fatal per-file load problem: the file was skipped, with
// an actionable message explaining what's wrong and how to fix it.
type FileProblem struct {
	Path    string `json:"path"`
	Message string `json:"message"`
}

// FixResult records the auto-repairs applied (or proposed) for one file.
type FixResult struct {
	Path    string   `json:"path"`
	Changes []string `json:"changes"`
}
