package domain

// Epic is a long-lived domain goal. Tasks reference it by ID via their
// `epic:` field. Epics live flat in epics/<id>.md (no status directories;
// status is a frontmatter field with its own vocabulary).
type Epic struct {
	ID   string `yaml:"-"`
	Path string `yaml:"-"`

	Status      string   `yaml:"status"`
	Description string   `yaml:"description"`
	Priority    string   `yaml:"priority"`
	Tags        []string `yaml:"tags"`
}
