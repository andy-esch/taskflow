package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// researchPath resolves the flat id-led file a CLI-created research doc landed at
// (research/<minted-id>-<slug>.md) — the id is minted from `created`, so the file is
// found by its slug suffix.
func researchPath(t *testing.T, root, slug string) string {
	t.Helper()
	m, err := filepath.Glob(filepath.Join(root, "research", "*-"+slug+".md"))
	if err != nil {
		t.Fatalf("glob research %q: %v", slug, err)
	}
	if len(m) != 1 {
		t.Fatalf("expected exactly one research file for %q, got %v", slug, m)
	}
	return m[0]
}

func TestResearchNew_CreatesIDLedFile(t *testing.T) {
	root := freshRepo(t)
	runRoot(t, "-C", root, "research", "new", "Compare theming libraries",
		"--tags", "tui,color", "--description", "Weighed three libs", "--created", "2026-06-23")

	path := researchPath(t, root, "compare-theming-libraries")
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	got := string(content)
	for _, want := range []string{`created: "2026-06-23"`, "description: Weighed three libs", "tags: [tui, color]", "# Compare theming libraries"} {
		if !strings.Contains(got, want) {
			t.Errorf("created file missing %q:\n%s", want, got)
		}
	}
	// The absences ARE the contract (epic 28): no lifecycle, no cross-references.
	for _, absent := range []string{"status:", "bucket:", "epic:"} {
		if strings.Contains(got, absent) {
			t.Errorf("research must not carry %q:\n%s", absent, got)
		}
	}
}

// `init` scaffolds research/ alongside the other entity dirs, so a fresh repo can
// take a research doc without a manual mkdir.
func TestInit_ScaffoldsResearchDir(t *testing.T) {
	root := freshRepo(t)
	if fi, err := os.Stat(filepath.Join(root, "research")); err != nil || !fi.IsDir() {
		t.Errorf("init must scaffold research/: err=%v", err)
	}
}

func TestResearchList_NewestFirstAndJSON(t *testing.T) {
	root := freshRepo(t)
	runRoot(t, "-C", root, "research", "new", "Older work", "--created", "2026-01-03")
	runRoot(t, "-C", root, "research", "new", "Newer work", "--created", "2026-06-23")

	// -o name is the byte-stable order check: newest first.
	out := runRoot(t, "-C", root, "research", "list", "-o", "name")
	want := "newer-work\nolder-work\n"
	if out != want {
		t.Errorf("list -o name = %q, want %q", out, want)
	}

	js := runRoot(t, "-C", root, "research", "list", "--json")
	var env struct {
		SchemaVersion string `json:"schema_version"`
		Research      []struct {
			ID, Slug, Created string
		} `json:"research"`
	}
	if err := json.Unmarshal([]byte(js), &env); err != nil {
		t.Fatalf("invalid --json: %v\n%s", err, js)
	}
	if len(env.Research) != 2 {
		t.Fatalf("want 2 docs, got %d: %s", len(env.Research), js)
	}
	if env.Research[0].Slug != "newer-work" {
		t.Errorf("--json order = %q first, want newer-work", env.Research[0].Slug)
	}
	// Ids are minted from created, so id order matches date order.
	if env.Research[1].ID >= env.Research[0].ID {
		t.Errorf("id order must follow created order: %+v", env.Research)
	}
	if env.SchemaVersion == "" {
		t.Error("envelope missing schema_version")
	}
}

func TestResearchList_TagFilter(t *testing.T) {
	root := freshRepo(t)
	runRoot(t, "-C", root, "research", "new", "Tui doc", "--tags", "tui")
	runRoot(t, "-C", root, "research", "new", "Core doc", "--tags", "core")

	out := runRoot(t, "-C", root, "research", "list", "--tag", "tui", "-o", "name")
	if out != "tui-doc\n" {
		t.Errorf("--tag tui = %q, want tui-doc", out)
	}
}

func TestResearchShow_AndPath(t *testing.T) {
	root := freshRepo(t)
	runRoot(t, "-C", root, "research", "new", "Storage model options", "--created", "2026-01-15")

	// Fuzzy prefix resolution, like tasks/audits.
	out := runRoot(t, "-C", root, "research", "show", "storage-model")
	if !strings.Contains(out, "storage-model-options") || !strings.Contains(out, "2026-01-15") {
		t.Errorf("show missing metadata:\n%s", out)
	}

	// Compared by basename: `research path` emits the symlink-resolved absolute path
	// (/private/var/… on macOS) while the glob returns the unresolved /var/… form.
	p := strings.TrimSpace(runRoot(t, "-C", root, "research", "path", "storage-model"))
	if filepath.Base(p) != filepath.Base(researchPath(t, root, "storage-model-options")) {
		t.Errorf("path = %q, want the resolved storage-model-options file", p)
	}
}

func TestResearchShow_FrontmatterOnly(t *testing.T) {
	root := freshRepo(t)
	runRoot(t, "-C", root, "research", "new", "Some doc")

	out := runRoot(t, "-C", root, "research", "show", "some-doc", "--frontmatter-only")
	if strings.Contains(out, "## Question") {
		t.Errorf("--frontmatter-only must omit the body:\n%s", out)
	}
	if !strings.Contains(out, "some-doc") {
		t.Errorf("--frontmatter-only must keep the metadata:\n%s", out)
	}
}

// A non-id-led file in research/ is a FileProblem (the ADR-0003 carveout gate), which
// `research list` must surface rather than silently skip — this is exactly the
// pre-migration state of the 28 legacy docs.
func TestResearchList_NonIDLedFileIsReported(t *testing.T) {
	root := freshRepo(t)
	mustWrite(t, filepath.Join(root, "research", "2026-06-28-legacy.md"), "---\ncreated: 2026-06-28\n---\n# Legacy\n")

	out, err := runRootRC(t, "-C", root, "research", "list")
	if err == nil {
		t.Fatal("a non-id-led research file must make `research list` exit non-zero")
	}
	// The aggregate error carries the count; the per-file line names the fix.
	if !strings.Contains(out, "no leading id") || !strings.Contains(out, "meta/") {
		t.Errorf("output should name the carveout gate and the fix, got:\n%s", out)
	}
}

func TestResearchNew_DryRunWritesNothing(t *testing.T) {
	root := freshRepo(t)
	runRoot(t, "-C", root, "--dry-run", "research", "new", "Not written")

	m, _ := filepath.Glob(filepath.Join(root, "research", "*.md"))
	if len(m) != 0 {
		t.Errorf("dry run wrote files: %v", m)
	}
}
