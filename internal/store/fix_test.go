package store

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/andy-esch/taskflow/internal/domain"

	"github.com/andy-esch/taskflow/internal/testutil"
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

func TestFixFrontmatterText_PreservesInlineComments(t *testing.T) {
	// A colon inside a trailing comment must NOT drag the comment into the value.
	in := []byte("---\n" +
		"priority: medium # note: double check\n" + // no fix needed; comment intact
		"description: Phase 1: do it # ref: TICKET-9\n" + // value quoted, comment kept outside
		"---\nbody\n")
	out, _ := fixFrontmatterText(in)
	s := string(out)

	if !strings.Contains(s, "priority: medium # note: double check") {
		t.Errorf("inline comment on an unchanged value was altered:\n%s", s)
	}
	if !strings.Contains(s, `description: "Phase 1: do it" # ref: TICKET-9`) {
		t.Errorf("comment folded into the quoted value:\n%s", s)
	}
	// And it parses (comments are legal YAML).
	fm, _ := splitFrontmatter(out)
	var m map[string]any
	if err := yaml.Unmarshal(fm, &m); err != nil {
		t.Fatalf("fixed frontmatter invalid: %v\n%s", err, fm)
	}
	if m["description"] != "Phase 1: do it" {
		t.Errorf("description value wrong: %#v", m["description"])
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
	path, out := testutil.TaskFixture(root, "ready-to-start", "bad.md", "---\nstatus: ready-to-start\ntags: a,b\n---\n# B\n")
	testutil.Write(t, path, out)

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

// A misspelled id is repairable in place because i/l/o have a canonical Crockford decode:
// the repaired id keeps the same decoded value, so it is the same identity spelled legally.
// Both the filename and the co-located `id:` field must move together, or lint immediately
// reports drift between them.
func TestFixRepairsMisspelledIDInFilenameAndFrontmatter(t *testing.T) {
	r := testutil.NewRepo(t)
	r.Epic("01-e.md", "---\nstatus: active\ndescription: e\n---\n# E\n")
	testutil.Write(t, filepath.Join(r.Root, "tasks", "6fbj87000lt6-bad-id.md"), "---\nschema: 1\nid: 6fbj87000lt6\nstatus: ready-to-start\nepic: 01-e\ndescription: d\n---\n# Bad\n")

	results, err := NewFS(r.Root).FixFrontmatter(false)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 || results[0].Skipped {
		t.Fatalf("expected one repair, got %+v", results)
	}
	fixed := filepath.Join(r.Root, "tasks", "6fbj870001t6-bad-id.md")
	body, err := os.ReadFile(fixed)
	if err != nil {
		t.Fatalf("repaired file not at its canonical name: %v", err)
	}
	if !strings.Contains(string(body), "id: 6fbj870001t6") {
		t.Errorf("frontmatter id was not moved with the filename:\n%s", body)
	}
	if _, err := os.Stat(filepath.Join(r.Root, "tasks", "6fbj87000lt6-bad-id.md")); !os.IsNotExist(err) {
		t.Error("the misspelled filename survived the repair")
	}
}

// The guard that matters: renaming a referenced id would leave those links dangling, and
// this repo has no rename cascade. Refusing loudly beats trading one broken file for
// several broken references.
func TestFixRefusesToRepairAnIDThatIsReferencedElsewhere(t *testing.T) {
	r := testutil.NewRepo(t)
	r.Epic("01-e.md", "---\nstatus: active\ndescription: e\n---\n# E\n")
	testutil.Write(t, filepath.Join(r.Root, "tasks", "6fbj87000lt6-bad-id.md"), "---\nschema: 1\nid: 6fbj87000lt6\nstatus: ready-to-start\nepic: 01-e\ndescription: d\n---\n# Bad\n")
	testutil.Write(t, filepath.Join(r.Root, "tasks", "6fbj870009t6-refers.md"), "---\nschema: 1\nid: 6fbj870009t6\nstatus: ready-to-start\nepic: 01-e\ndescription: d\n---\n# Ref\n\nSee [bad](6fbj87000lt6-bad-id.md).\n")

	results, err := NewFS(r.Root).FixFrontmatter(false)
	if err != nil {
		t.Fatal(err)
	}
	var refusal *domain.FixResult
	for i := range results {
		if strings.Contains(results[i].Path, "6fbj87000lt6") {
			refusal = &results[i]
		}
	}
	if refusal == nil || !refusal.Skipped {
		t.Fatalf("a referenced id must be skipped, not repaired: %+v", results)
	}
	if !strings.Contains(refusal.Changes[0], "6fbj870009t6-refers.md") {
		t.Errorf("the refusal must name the referring file: %q", refusal.Changes[0])
	}
	if _, err := os.Stat(filepath.Join(r.Root, "tasks", "6fbj87000lt6-bad-id.md")); err != nil {
		t.Error("the file was renamed despite being referenced")
	}
}

// u is excluded from Crockford with no digit it stands for, so there is nothing to repair
// it TO — "fixing" it would mean choosing a different identity.
func TestFixRefusesAnIDContainingU(t *testing.T) {
	r := testutil.NewRepo(t)
	r.Epic("01-e.md", "---\nstatus: active\ndescription: e\n---\n# E\n")
	testutil.Write(t, filepath.Join(r.Root, "tasks", "6fbj87000ut6-u-id.md"), "---\nschema: 1\nid: 6fbj87000ut6\nstatus: ready-to-start\nepic: 01-e\ndescription: d\n---\n# U\n")

	results, err := NewFS(r.Root).FixFrontmatter(false)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 || !results[0].Skipped {
		t.Fatalf("a u id must be skipped: %+v", results)
	}
	if !strings.Contains(results[0].Changes[0], "canonical decode") {
		t.Errorf("the refusal should explain why u cannot be repaired: %q", results[0].Changes[0])
	}
}
