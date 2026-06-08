package domain

// Task is a planning task. Fields tagged `yaml:"-"` are derived by the store
// (filename, path) and are not part of the markdown frontmatter.
type Task struct {
	Slug string `yaml:"-"`
	Path string `yaml:"-"`

	Status      Status   `yaml:"status"`
	Epic        string   `yaml:"epic"`
	Description string   `yaml:"description"`
	Tier        int      `yaml:"tier"`
	Priority    string   `yaml:"priority"`
	Autonomy    int      `yaml:"autonomy_level"`
	Effort      string   `yaml:"effort"`
	Created     string   `yaml:"created"`
	Tags        []string `yaml:"tags"`
}
