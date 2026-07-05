package domain

// Task is a planning task. Fields tagged `yaml:"-"` are derived by the store
// (filename, path) and are not part of the markdown frontmatter.
type Task struct {
	Slug string `yaml:"-"`
	Path string `yaml:"-"`
	// StatusFellBack is set by the store when the frontmatter status is missing or
	// unrecognized — under the flat layout (ADR-0003 §4) there is no directory to fall
	// back to, so Status keeps its raw value; the task still lists and lint flags it
	// (FrontmatterStatusIssues).
	StatusFellBack bool `yaml:"-"`

	// ID is the stable 12-char identifier (ADR-0003 §3): it leads the flat filename
	// (tasks/<id>-<slug>.md) and is the primary resolution key.
	ID string `yaml:"id"`

	// FilenameID is that same id as parsed from the flat filename's leading field
	// (set by the store via splitFlatName). It is the canonical key resolveID/CAS
	// match on; the frontmatter `id:` above is a co-located copy that must equal it,
	// and lint flags any drift (IDDriftIssue). Derived, not frontmatter.
	FilenameID string `yaml:"-"`

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
