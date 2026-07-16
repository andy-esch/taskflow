package cli

import (
	"encoding/json"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/andy-esch/taskflow/internal/domain"
)

// wantAlphaPath is alpha's canonical absolute path — resolving symlinks the way
// discovery does (macOS /var → /private/var), so it matches the store-emitted path.
func wantAlphaPath(t *testing.T, root string) string {
	t.Helper()
	p, err := filepath.EvalSymlinks(alphaPath(root))
	if err != nil {
		t.Fatal(err)
	}
	return p
}

// `task path` prints exactly the absolute file path (pipe-friendly, no globbing).
func TestTaskPath(t *testing.T) {
	root := setupRepo(t)
	out := runRoot(t, "-C", root, "task", "path", "alpha")
	want := wantAlphaPath(t, root)
	if strings.TrimSpace(out) != want {
		t.Errorf("task path = %q, want %q", strings.TrimSpace(out), want)
	}
	if strings.Contains(strings.TrimSpace(out), "\n") {
		t.Errorf("task path should print exactly one line:\n%s", out)
	}
}

// `task path --json` wraps the same path so the schema_version contract holds.
func TestTaskPath_JSON(t *testing.T) {
	root := setupRepo(t)
	out := runRoot(t, "-C", root, "--json", "task", "path", "alpha")
	var env struct {
		SchemaVersion string `json:"schema_version"`
		Path          string `json:"path"`
	}
	if err := json.Unmarshal([]byte(out), &env); err != nil {
		t.Fatalf("task path --json not parseable: %v\n%s", err, out)
	}
	want := wantAlphaPath(t, root)
	if env.SchemaVersion == "" || env.Path != want {
		t.Errorf("task path --json wrong: %+v, want path %q", env, want)
	}
}

// `task info --json` is the token-cheap metadata read: path + triage fields + the
// acceptance-criteria tally, no body.
func TestTaskInfo_JSON(t *testing.T) {
	root := setupRepo(t)
	// Give alpha an acceptance-criteria section: 1 of 2 checked.
	runRoot(t, "-C", root, "task", "set", "alpha", "--body",
		"# Alpha\n\n## Acceptance criteria\n\n- [x] done\n- [ ] not yet\n")
	out := runRoot(t, "-C", root, "--json", "task", "info", "alpha")
	var env struct {
		SchemaVersion string `json:"schema_version"`
		TaskInfo      struct {
			Slug   string `json:"slug"`
			Status string `json:"status"`
			Path   string `json:"path"`
			AC     struct {
				Checked int `json:"checked"`
				Total   int `json:"total"`
			} `json:"ac"`
		} `json:"task_info"`
	}
	if err := json.Unmarshal([]byte(out), &env); err != nil {
		t.Fatalf("task info --json not parseable: %v\n%s", err, out)
	}
	want := wantAlphaPath(t, root)
	ti := env.TaskInfo
	if env.SchemaVersion == "" || ti.Slug != "alpha" || ti.Status != "ready-to-start" || ti.Path != want {
		t.Errorf("task info --json metadata wrong:\n%s", out)
	}
	if ti.AC.Checked != 1 || ti.AC.Total != 2 {
		t.Errorf("ac tally = %d/%d, want 1/2\n%s", ti.AC.Checked, ti.AC.Total, out)
	}
}

// A task with no acceptance-criteria section reports a zero tally, not an error.
func TestTaskInfo_NoAcceptanceSection(t *testing.T) {
	root := setupRepo(t) // alpha's fixture body is just "# Alpha"
	out := runRoot(t, "-C", root, "--json", "task", "info", "alpha")
	if !strings.Contains(out, `"ac":{"checked":0,"total":0}`) {
		t.Errorf("no AC section should be a 0/0 tally:\n%s", out)
	}
}

// `task show --section` narrows the human output to one section.
func TestTaskShow_Section(t *testing.T) {
	root := setupRepo(t)
	runRoot(t, "-C", root, "task", "set", "alpha", "--body",
		"# Alpha\n\n## Objective\n\nDo the thing.\n\n## Acceptance criteria\n\n- [ ] a\n- [ ] b\n")
	out := runRoot(t, "-C", root, "task", "show", "alpha", "--section", "acceptance", "--raw")
	if !strings.Contains(out, "## Acceptance criteria") || !strings.Contains(out, "- [ ] a") {
		t.Errorf("--section acceptance should show that section:\n%s", out)
	}
	if strings.Contains(out, "## Objective") || strings.Contains(out, "Do the thing.") {
		t.Errorf("--section acceptance must NOT show other sections:\n%s", out)
	}
}

// Under --json, --section narrows the body field (envelope shape unchanged).
func TestTaskShow_SectionJSON_NarrowsBody(t *testing.T) {
	root := setupRepo(t)
	runRoot(t, "-C", root, "task", "set", "alpha", "--body",
		"# Alpha\n\n## Objective\n\ntop\n\n## Acceptance criteria\n\n- [ ] a\n")
	out := runRoot(t, "-C", root, "--json", "task", "show", "alpha", "--section", "acceptance")
	var env struct {
		Body string `json:"body"`
	}
	if err := json.Unmarshal([]byte(out), &env); err != nil {
		t.Fatalf("task show --section --json not parseable: %v\n%s", err, out)
	}
	if !strings.HasPrefix(env.Body, "## Acceptance criteria") || strings.Contains(env.Body, "Objective") {
		t.Errorf("--json --section should narrow body to the section:\n%q", env.Body)
	}
}

// A missing section is a clean not-found (exit 10), not a silent empty body.
func TestTaskShow_SectionNotFound(t *testing.T) {
	root := setupRepo(t)
	if _, err := runRootRC(t, "-C", root, "task", "show", "alpha", "--section", "nonexistent"); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("a missing section should wrap ErrNotFound (exit 10), got %v", err)
	}
}

// `task show --frontmatter-only` keeps the metadata, drops the body.
func TestTaskShow_FrontmatterOnly(t *testing.T) {
	root := setupRepo(t)
	out := runRoot(t, "-C", root, "task", "show", "alpha", "--frontmatter-only")
	if !strings.Contains(out, "alpha") {
		t.Errorf("--frontmatter-only should still show metadata:\n%s", out)
	}
	if strings.Contains(out, "# Alpha") {
		t.Errorf("--frontmatter-only must omit the body:\n%s", out)
	}
	// No body → no trailing blank line under the metadata fields.
	if strings.HasSuffix(out, "\n\n") {
		t.Errorf("--frontmatter-only must not leave a trailing blank line:\n%q", out)
	}
}

func TestTaskShow_SectionAndFrontmatterOnly_Exclusive(t *testing.T) {
	root := setupRepo(t)
	if _, err := runRootRC(t, "-C", root, "task", "show", "alpha", "--section", "x", "--frontmatter-only"); err == nil {
		t.Fatal("--section and --frontmatter-only should be mutually exclusive")
	}
}
