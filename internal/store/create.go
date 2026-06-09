package store

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"

	yaml "go.yaml.in/yaml/v3"

	"github.com/andy-esch/taskflow/internal/domain"
)

// fmField is one frontmatter key/value, written in declared order for new files.
type fmField struct {
	key string
	val any
}

// buildFile serializes ordered frontmatter fields + a markdown body into a
// complete file. Values go through valueNode, so a description containing a
// colon is correctly quoted (the pm non-conformant-YAML trap, avoided at the
// source).
func buildFile(fields []fmField, body string) ([]byte, error) {
	mapping := &yaml.Node{Kind: yaml.MappingNode}
	for _, f := range fields {
		node, err := valueNode(f.val)
		if err != nil {
			return nil, fmt.Errorf("field %q: %w", f.key, err)
		}
		mapping.Content = append(mapping.Content,
			&yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: f.key}, node)
	}
	return assembleFile(mapping, []byte(body))
}

// taskFields is the canonical frontmatter order for a new task.
func taskFields(t domain.Task) []fmField {
	return []fmField{
		{"status", string(t.Status)},
		{"epic", t.Epic},
		{"description", t.Description},
		{"effort", t.Effort},
		{"tier", t.Tier},
		{"priority", t.Priority},
		{"autonomy_level", t.Autonomy},
		{"tags", t.Tags},
		{"created", t.Created},
	}
}

// CreateTask writes a new task file under tasks/<status>/<slug>.md. It refuses
// to clobber an existing file. The slug and status are taken from t.
func (s *FS) CreateTask(t domain.Task, body string) (domain.Task, error) {
	if t.Slug == "" {
		return domain.Task{}, fmt.Errorf("%w: empty task slug", domain.ErrValidation)
	}
	dir := filepath.Join(s.tasksDir, t.Status.Dir())
	path := filepath.Join(dir, t.Slug+".md")
	if _, err := os.Stat(path); err == nil {
		return domain.Task{}, fmt.Errorf("task %q already exists: %w", t.Slug, domain.ErrValidation)
	}
	content, err := buildFile(taskFields(t), body)
	if err != nil {
		return domain.Task{}, err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return domain.Task{}, fmt.Errorf("mkdir %s: %w", dir, err)
	}
	if err := writeFileAtomic(path, content, 0o644); err != nil {
		return domain.Task{}, err
	}
	t.Path = path
	return t, nil
}

var epicNumRe = regexp.MustCompile(`^(\d+)-`)

// nextEpicNumber returns max(existing NN- prefix)+1, or 1 if none.
func (s *FS) nextEpicNumber() (int, error) {
	entries, err := os.ReadDir(s.epicsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return 1, nil
		}
		return 0, fmt.Errorf("read epics dir: %w", err)
	}
	next := 1
	for _, e := range entries {
		m := epicNumRe.FindStringSubmatch(e.Name())
		if m == nil {
			continue
		}
		if n, _ := strconv.Atoi(m[1]); n+1 > next {
			next = n + 1
		}
	}
	return next, nil
}

// epicFields is the canonical frontmatter order for a new epic.
func epicFields(e domain.Epic) []fmField {
	return []fmField{
		{"status", e.Status},
		{"description", e.Description},
		{"priority", e.Priority},
		{"tags", e.Tags},
		{"created", e.Created},
	}
}

// CreateEpic writes a new epic at epics/NN-<slug>.md, auto-assigning the next
// number. It refuses to clobber an existing file.
func (s *FS) CreateEpic(slug string, e domain.Epic, body string) (domain.Epic, error) {
	if slug == "" {
		return domain.Epic{}, fmt.Errorf("%w: empty epic slug", domain.ErrValidation)
	}
	num, err := s.nextEpicNumber()
	if err != nil {
		return domain.Epic{}, err
	}
	id := fmt.Sprintf("%02d-%s", num, slug)
	path := filepath.Join(s.epicsDir, id+".md")
	if _, err := os.Stat(path); err == nil {
		return domain.Epic{}, fmt.Errorf("epic %q already exists: %w", id, domain.ErrValidation)
	}
	content, err := buildFile(epicFields(e), body)
	if err != nil {
		return domain.Epic{}, err
	}
	if err := os.MkdirAll(s.epicsDir, 0o755); err != nil {
		return domain.Epic{}, fmt.Errorf("mkdir %s: %w", s.epicsDir, err)
	}
	if err := writeFileAtomic(path, content, 0o644); err != nil {
		return domain.Epic{}, err
	}
	e.ID = id
	e.Path = path
	return e, nil
}
