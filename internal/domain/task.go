package domain

// Task is a planning task. Fields tagged `yaml:"-"` are derived by the store
// (filename, path, declared) and are not part of the markdown frontmatter.
type Task struct {
	Slug string `yaml:"-"`
	Path string `yaml:"-"`
	// Declared is the status the frontmatter literally claimed. Status is the
	// authoritative one (its directory); the two diverge only when a file is
	// misfiled — see Misfiled.
	Declared Status `yaml:"-"`

	Status      Status   `yaml:"status"`
	Epic        string   `yaml:"epic"`
	Description string   `yaml:"description"`
	Tier        int      `yaml:"tier"`
	Priority    string   `yaml:"priority"`
	Autonomy    int      `yaml:"autonomy_level"`
	Effort      string   `yaml:"effort"`
	Created     string   `yaml:"created"`
	Updated     string   `yaml:"updated_at"`
	Tags        []string `yaml:"tags"`
}

// Misfiled reports whether the frontmatter declared a *recognized* status that
// disagrees with the file's directory (the authoritative Status). A foreign or
// invalid status word (e.g. legacy "superseded") is not misfiled — the folder
// simply governs.
func (t Task) Misfiled() bool {
	return t.Declared.Valid() && t.Declared != t.Status
}
