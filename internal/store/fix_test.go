package store

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	yaml "go.yaml.in/yaml/v3"
)

func TestFixFrontmatterText(t *testing.T) {
	in := []byte("---\nstatus: ready-to-start\ndescription: Phase 1: do the thing\ntags: a,b,c\n---\n# Body\nkeep: this: line\n")
	out, changes := fixFrontmatterText(in)
	s := string(out)

	if !strings.Contains(s, `description: "Phase 1: do the thing"`) {
		t.Errorf("description not quoted:\n%s", s)
	}
	if !strings.Contains(s, "tags: [a, b, c]") {
		t.Errorf("tags not normalized to a list:\n%s", s)
	}
	if !strings.Contains(s, "# Body\nkeep: this: line\n") {
		t.Errorf("body not preserved verbatim:\n%s", s)
	}
	if len(changes) != 2 {
		t.Errorf("want 2 changes, got %v", changes)
	}

	// The whole point: the result now parses as valid YAML.
	fm, _ := splitFrontmatter(out)
	var m map[string]any
	if err := yaml.Unmarshal(fm, &m); err != nil {
		t.Fatalf("fixed frontmatter is still invalid: %v\n%s", err, fm)
	}
	if tags, ok := m["tags"].([]any); !ok || len(tags) != 3 {
		t.Errorf("tags not a 3-element list: %#v", m["tags"])
	}
}

func TestFixFrontmatterText_NoOp(t *testing.T) {
	in := []byte("---\nstatus: ready-to-start\ndescription: clean\ntags: [a, b]\n---\n# Body\n")
	out, changes := fixFrontmatterText(in)
	if len(changes) != 0 || string(out) != string(in) {
		t.Errorf("clean file changed: %v\n%s", changes, out)
	}
}

func TestFixFrontmatterText_Idempotent(t *testing.T) {
	in := []byte("---\ndescription: a: b\ntags: x\n---\nbody\n")
	once, _ := fixFrontmatterText(in)
	twice, changes := fixFrontmatterText(once)
	if len(changes) != 0 || string(twice) != string(once) {
		t.Errorf("not idempotent; second pass changed %v\n%s", changes, once)
	}
}

func TestFS_FixFrontmatter_DryRunThenWrite(t *testing.T) {
	root := t.TempDir()
	writeTask(t, root, "ready-to-start", "bad.md", "---\nstatus: ready-to-start\ntags: a,b\n---\n# B\n")
	path := filepath.Join(root, "tasks", "ready-to-start", "bad.md")

	res, err := NewFS(root).FixFrontmatter(true) // dry-run
	if err != nil {
		t.Fatal(err)
	}
	if len(res) != 1 {
		t.Fatalf("want 1 fixable file, got %d", len(res))
	}
	if raw, _ := os.ReadFile(path); !strings.Contains(string(raw), "tags: a,b") {
		t.Errorf("dry-run must not modify the file:\n%s", raw)
	}

	if _, err := NewFS(root).FixFrontmatter(false); err != nil { // real
		t.Fatal(err)
	}
	tasks, problems, err := NewFS(root).ListTasks()
	if err != nil {
		t.Fatal(err)
	}
	if len(problems) != 0 {
		t.Errorf("still unreadable after fix: %+v", problems)
	}
	if len(tasks) != 1 || len(tasks[0].Tags) != 2 {
		t.Errorf("tags not fixed: %+v", tasks)
	}
}
