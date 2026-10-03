package domain

// Task is a planning task. The source identity lives in the adapter read;
// Path is retained temporarily for legacy local callers.
type Task struct {
	Slug string `yaml:"-"`
	Path string `yaml:"-"` // transitional local source location; removed in the next slice
	// Title is the first non-fenced H1 in the Markdown body. It is derived by
	// adapters rather than duplicated in frontmatter; callers must fall back to
	// Slug when an adapter cannot provide body-derived presentation data.
	Title string `yaml:"-"`
	// StatusFellBack is set by the store when the frontmatter status is missing or
	// unrecognized — under the flat layout (ADR-0003 §4) there is no directory to fall
	// back to, so Status keeps its raw value; the task still lists and lint flags it
	// (FrontmatterStatusIssues).
	StatusFellBack bool `yaml:"-"`

	// ID is the declared stable identifier (ADR-0003 §3). A local file repeats
	// it in tasks/<id>-<slug>.md; the adapter's source ID is the resolution key.
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
	// LegacyDependencyFields preserves field presence separately from values so an
	// explicitly empty legacy key remains diagnosable and migratable. Values are
	// the canonical field names and are populated by the store parser.
	LegacyDependencyFields []string `yaml:"-"`
}
