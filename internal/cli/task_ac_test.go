package cli

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/andy-esch/taskflow/internal/domain"
)

const acTestBody = "# Alpha\n\n## Acceptance criteria\n\n- [ ] first\n- [x] second\n- [ ] third\n"

// setupRepoWithAC gives alpha a 3-item acceptance section (2 unchecked, 1 checked).
func setupRepoWithAC(t *testing.T) string {
	t.Helper()
	root := setupRepo(t)
	runRoot(t, "-C", root, "task", "set", "alpha", "--body", acTestBody)
	return root
}

func TestTaskAc_List(t *testing.T) {
	root := setupRepoWithAC(t)
	out := runRoot(t, "-C", root, "task", "ac", "alpha")
	for _, want := range []string{"1.", "first", "2.", "second", "3.", "third", "[x]", "[ ]"} {
		if !strings.Contains(out, want) {
			t.Errorf("ac list missing %q:\n%s", want, out)
		}
	}
}

func TestTaskAc_List_JSON(t *testing.T) {
	root := setupRepoWithAC(t)
	out := runRoot(t, "-C", root, "--json", "task", "ac", "alpha")
	var env struct {
		SchemaVersion string `json:"schema_version"`
		Slug          string `json:"slug"`
		Acceptance    []struct {
			Index   int    `json:"index"`
			Checked bool   `json:"checked"`
			Text    string `json:"text"`
		} `json:"acceptance"`
	}
	if err := json.Unmarshal([]byte(out), &env); err != nil {
		t.Fatalf("ac --json not parseable: %v\n%s", err, out)
	}
	if env.SchemaVersion == "" || env.Slug != "alpha" || len(env.Acceptance) != 3 {
		t.Fatalf("ac --json wrong shape:\n%s", out)
	}
	if env.Acceptance[0].Checked || !env.Acceptance[1].Checked || env.Acceptance[1].Text != "second" {
		t.Errorf("ac --json criteria wrong:\n%s", out)
	}
}

// Listing a task with no acceptance section succeeds with an empty list (distinct
// from a flip, which errors) — human says so, --json emits [] (never null).
func TestTaskAc_List_NoSection(t *testing.T) {
	root := setupRepo(t) // alpha's body is just "# Alpha"
	if out := runRoot(t, "-C", root, "task", "ac", "alpha"); !strings.Contains(out, "no acceptance criteria") {
		t.Errorf("list of a task with no AC section should say so:\n%s", out)
	}
	out := runRoot(t, "-C", root, "--json", "task", "ac", "alpha")
	if !strings.Contains(out, `"acceptance":[]`) {
		t.Errorf("empty acceptance must marshal to [] (not null):\n%s", out)
	}
}

func TestTaskAc_Check_FlipsFile(t *testing.T) {
	root := setupRepoWithAC(t)
	runRoot(t, "-C", root, "task", "ac", "alpha", "--check", "1")
	got := readFile(t, alphaPath(root))
	if !strings.Contains(got, "- [x] first") {
		t.Errorf("--check 1 should tick criterion 1:\n%s", got)
	}
	if !strings.Contains(got, "status: ready-to-start") {
		t.Errorf("frontmatter must be preserved by the flip:\n%s", got)
	}
	if !strings.Contains(got, "- [x] second") || !strings.Contains(got, "- [ ] third") {
		t.Errorf("--check 1 must not disturb other criteria:\n%s", got)
	}
}

func TestTaskAc_Uncheck_FlipsFile(t *testing.T) {
	root := setupRepoWithAC(t)
	runRoot(t, "-C", root, "task", "ac", "alpha", "--uncheck", "2")
	if got := readFile(t, alphaPath(root)); !strings.Contains(got, "- [ ] second") {
		t.Errorf("--uncheck 2 should clear criterion 2:\n%s", got)
	}
}

func TestTaskAc_Check_OutOfRange(t *testing.T) {
	root := setupRepoWithAC(t)
	if _, err := runRootRC(t, "-C", root, "task", "ac", "alpha", "--check", "99"); !errors.Is(err, domain.ErrValidation) {
		t.Errorf("out-of-range index should wrap ErrValidation (exit 11), got %v", err)
	}
}

func TestTaskAc_Check_NoSection(t *testing.T) {
	root := setupRepo(t) // alpha's body is just "# Alpha" — no acceptance section
	if _, err := runRootRC(t, "-C", root, "task", "ac", "alpha", "--check", "1"); !errors.Is(err, domain.ErrValidation) {
		t.Errorf("no AC section should wrap ErrValidation, got %v", err)
	}
}

func TestTaskAc_Check_Idempotent_NoWrite(t *testing.T) {
	root := setupRepoWithAC(t)
	before, _ := os.ReadFile(alphaPath(root))
	out := runRoot(t, "-C", root, "task", "ac", "alpha", "--check", "2") // #2 is already checked
	after, _ := os.ReadFile(alphaPath(root))
	if !bytes.Equal(before, after) {
		t.Error("checking an already-checked criterion must not write")
	}
	if !strings.Contains(out, "already checked") {
		t.Errorf("expected an 'already checked' note:\n%s", out)
	}
}

func TestTaskAc_Check_DryRun_NoWrite(t *testing.T) {
	root := setupRepoWithAC(t)
	before, _ := os.ReadFile(alphaPath(root))
	runRoot(t, "-C", root, "--dry-run", "task", "ac", "alpha", "--check", "1")
	after, _ := os.ReadFile(alphaPath(root))
	if !bytes.Equal(before, after) {
		t.Error("--dry-run --check must not write")
	}
}

func TestTaskAc_CheckUncheck_Exclusive(t *testing.T) {
	root := setupRepoWithAC(t)
	if _, err := runRootRC(t, "-C", root, "task", "ac", "alpha", "--check", "1", "--uncheck", "2"); err == nil {
		t.Fatal("--check and --uncheck should be mutually exclusive")
	}
}

// A flip returns the task_mutation envelope (it edits the body), echoing the result.
func TestTaskAc_Check_JSON_MutationEnvelope(t *testing.T) {
	root := setupRepoWithAC(t)
	out := runRoot(t, "-C", root, "--json", "task", "ac", "alpha", "--check", "1")
	var env mutationEnv // defined in body_test.go
	if err := json.Unmarshal([]byte(out), &env); err != nil {
		t.Fatalf("ac --check --json is not a task_mutation envelope: %v\n%s", err, out)
	}
	if env.Task.Slug != "alpha" || !strings.Contains(env.Body, "- [x] first") {
		t.Errorf("ac --check --json should echo the flipped body:\n%s", out)
	}
}

// The acceptance-criteria lint guard flows end-to-end through `lint` (exercising the
// body-carrying scan): a task whose acceptance section has a botched checkbox is
// flagged, so the silent under-count is caught rather than trusted.
func TestLint_FlagsMalformedAcceptance(t *testing.T) {
	root := setupRepoWithAC(t)
	// Replace with a body that has a botched checkbox in the acceptance section.
	runRoot(t, "-C", root, "task", "set", "alpha", "--body",
		"# Alpha\n\n## Acceptance criteria\n\n- [x] ok\n- [] botched\n")
	// lint exits non-zero when it finds issues, so read via runRootRC and inspect the
	// (still-emitted) JSON rather than fataling on the exit code.
	out, _ := runRootRC(t, "-C", root, "--json", "lint")
	if !strings.Contains(out, "malformed acceptance checkbox") {
		t.Errorf("lint --json should surface the malformed acceptance checkbox:\n%s", out)
	}
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}
