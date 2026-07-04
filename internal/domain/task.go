package domain

// Task is a planning task. Fields tagged `yaml:"-"` are derived by the store
// (filename, path, declared) and are not part of the markdown frontmatter.
type Task struct {
	Slug string `yaml:"-"`
	Path string `yaml:"-"`
	// FolderStatus is the status the file's directory implies (the mirror). Status
	// is the authoritative one (frontmatter, ADR-0003 Phase A); the two diverge only
	// when a file is misfiled — the directory hasn't caught up to a frontmatter
	// status change. See Misfiled.
	FolderStatus Status `yaml:"-"`
	// StatusFellBack is set by the store when the frontmatter status was missing or
	// unrecognized, so Status above is the folder fallback rather than a real
	// frontmatter value — lint flags it (FrontmatterStatusIssues).
	StatusFellBack bool `yaml:"-"`

	// ID is the stable 12-char identifier (ADR-0003), minted on create by the core
	// service. Additive for now: written to new files but not yet resolved on,
	// exposed in --json, or in the field registry — those land when the id moves
	// into the filename (the stable-key layout change). Empty on pre-rollout files.
	ID string `yaml:"id"`

	Status      Status   `yaml:"status"`
	Epic        string   `yaml:"epic"`
	Description string   `yaml:"description"`
	Tier        int      `yaml:"tier"`
	Priority    string   `yaml:"priority"`
	Autonomy    int      `yaml:"autonomy_level"`
	Effort      string   `yaml:"effort"`
	Created     string   `yaml:"created"`
	Updated     string   `yaml:"updated_at"`
	StartedAt   string   `yaml:"started_at"`           // stamped when a task enters in-progress (incl. `new --start`)
	RevisitAt   string   `yaml:"revisit_at,omitempty"` // optional "snooze until" date for a deferred task (set by `task defer`)
	Tags        []string `yaml:"tags"`
}

// Misfiled reports whether the file sits in a directory that disagrees with its
// authoritative (frontmatter) status — the mirror is stale and the file should be
// moved to match. Guarded on a valid FolderStatus so a Task with no folder context
// (e.g. one built in a test, or before parseTask records the directory) is never
// misfiled; a frontmatter word that names no recognized status falls back to the
// folder in parseTask, so those arrive here equal and aren't misfiled either.
func (t Task) Misfiled() bool {
	return t.FolderStatus.Valid() && t.Status != t.FolderStatus
}
