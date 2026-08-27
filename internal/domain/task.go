package domain

// Task is a planning task. Fields tagged `yaml:"-"` are derived by the store
// (filename, path) and are not part of the markdown frontmatter.
type Task struct {
	Slug string `yaml:"-"`
	Path string `yaml:"-"`
	// SourceVersion is the store-internal hash of the exact bytes that produced this
	// record. TaskGraph retains it for whole-snapshot CAS but clears it from Task()
	// projections, so planners never receive persistence tokens.
	SourceVersion string `yaml:"-"`
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

	// DependsOn is the canonical repository-global dependency set from ADR-0006.
	// Values are stable task IDs, never slugs. Valid writers serialize the semantic
	// set in sorted order; readers deliberately retain malformed duplicate values so
	// the strict graph snapshot and lint can diagnose hand-edited files precisely.
	DependsOn []string `yaml:"depends_on,omitempty"`

	// These fields are read-only legacy vocabulary. Keeping them on the typed record
	// lets the strict snapshot resolve and diagnose the live slug references without
	// treating them as canonical edges or silently dropping them during analysis. The
	// guarded dependency-migration slice removes them later.
	LegacyBlockedBy    []string `yaml:"blocked_by,omitempty"`
	LegacyDependencies []string `yaml:"dependencies,omitempty"`
	LegacyBlocks       []string `yaml:"blocks,omitempty"`
}
