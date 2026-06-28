package cli

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/andy-esch/taskflow/internal/domain"
)

func alphaPath(root string) string {
	return filepath.Join(root, "tasks", "ready-to-start", "alpha.md")
}

// `task append` adds a section to the body through the tool, atomically.
func TestTaskAppend_Body(t *testing.T) {
	root := setupRepo(t) // alpha (ready-to-start)
	runRoot(t, "-C", root, "task", "append", "alpha", "--body", "## Review\n- looks good")
	b, err := os.ReadFile(alphaPath(root))
	if err != nil {
		t.Fatal(err)
	}
	got := string(b)
	if !strings.Contains(got, "# Alpha") || !strings.Contains(got, "## Review") {
		t.Errorf("append should keep the old body and add the section:\n%s", got)
	}
	if !strings.Contains(got, "updated_at:") {
		t.Errorf("append should stamp updated_at:\n%s", got)
	}
}

// `task append --body-file -` reads the appended content from the injected stdin.
// The reader is passed via NewRootCmd's `in` param alone (no separate cmd.SetIn) —
// proving M12: the one injected reader reaches resolveBody's cmd.InOrStdin().
func TestTaskAppend_Stdin(t *testing.T) {
	root := setupRepo(t)
	var out bytes.Buffer
	cmd := NewRootCmd(strings.NewReader("## Notes\n- from stdin\n"), &out, &out)
	cmd.SetArgs([]string{"-C", root, "task", "append", "alpha", "--body-file", "-"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	b, _ := os.ReadFile(alphaPath(root))
	if !strings.Contains(string(b), "from stdin") {
		t.Errorf("append body should come from the injected stdin:\n%s", b)
	}
}

// TestNewRootCmd_UnifiesStdin pins M12: the single `in` param is the one stdin
// owner. Before the fix, root.go never called SetIn, so cmd.InOrStdin() (the
// source resolveBody reads for `--body-file -`) fell back to os.Stdin — a DIFFERENT
// handle than app.In (which the prompt gate, prompter, and editor read). Now both
// are the injected reader: this asserts the cobra side directly, and app.In is set
// to the same `in` in the constructor, so every input path agrees.
func TestNewRootCmd_UnifiesStdin(t *testing.T) {
	in := strings.NewReader("piped input\n")
	var out bytes.Buffer
	cmd := NewRootCmd(in, &out, &out)
	if cmd.InOrStdin() != in {
		t.Error("cmd.InOrStdin() must be the injected reader (resolveBody's source), not os.Stdin")
	}
}

// Empty append input is a clean validation error, not an empty write.
func TestTaskAppend_Empty_Errors(t *testing.T) {
	root := setupRepo(t)
	var out bytes.Buffer
	cmd := NewRootCmd(strings.NewReader(""), &out, &out)
	cmd.SetArgs([]string{"-C", root, "task", "append", "alpha", "--body", "   "})
	if err := cmd.Execute(); !errors.Is(err, domain.ErrValidation) {
		t.Errorf("empty append should wrap ErrValidation (exit 11), got %v", err)
	}
}

// Passing both --body and --body-file is a usage error, not a silent precedence
// pick (resolveBody would otherwise quietly prefer --body-file and drop --body).
func TestTaskAppend_BodyAndBodyFile_Exclusive(t *testing.T) {
	root := setupRepo(t)
	if _, err := runRootRC(t, "-C", root, "task", "append", "alpha", "--body", "x", "--body-file", "-"); err == nil {
		t.Fatal("`task append --body … --body-file -` should be rejected (mutually exclusive)")
	}
}

func TestTaskSet_BodyAndBodyFile_Exclusive(t *testing.T) {
	root := setupRepo(t)
	if _, err := runRootRC(t, "-C", root, "task", "set", "alpha", "--body", "x", "--body-file", "-"); err == nil {
		t.Fatal("`task set --body … --body-file -` should be rejected (mutually exclusive)")
	}
}

// `task set --body` replaces the whole body.
func TestTaskSet_BodyReplace(t *testing.T) {
	root := setupRepo(t)
	runRoot(t, "-C", root, "task", "set", "alpha", "--body", "# Rewritten\n\nbrand new")
	b, _ := os.ReadFile(alphaPath(root))
	got := string(b)
	if !strings.Contains(got, "brand new") || strings.Contains(got, "# Alpha") {
		t.Errorf("set --body should replace the body wholesale:\n%s", got)
	}
	if !strings.Contains(got, "status: ready-to-start") {
		t.Errorf("frontmatter must be preserved by set --body:\n%s", got)
	}
}

// `set --body` mixed with field flags is rejected (body edits are their own call).
func TestTaskSet_BodyWithFields_Rejected(t *testing.T) {
	root := setupRepo(t)
	var out bytes.Buffer
	cmd := NewRootCmd(strings.NewReader(""), &out, &out)
	cmd.SetArgs([]string{"-C", root, "task", "set", "alpha", "--tier", "2", "--body", "x"})
	if err := cmd.Execute(); !errors.Is(err, domain.ErrValidation) {
		t.Errorf("combining --body with field flags should wrap ErrValidation, got %v", err)
	}
}

// --dry-run previews a body edit without writing (both append and set --body).
func TestTaskAppend_DryRun_NoWrite(t *testing.T) {
	root := setupRepo(t)
	before, _ := os.ReadFile(alphaPath(root))
	runRoot(t, "-C", root, "--dry-run", "task", "append", "alpha", "--body", "## Nope")
	runRoot(t, "-C", root, "--dry-run", "task", "set", "alpha", "--body", "# Nope")
	after, _ := os.ReadFile(alphaPath(root))
	if !bytes.Equal(before, after) {
		t.Error("--dry-run body edits must not write")
	}
}

// The task_mutation --json envelope carries dry_run + the resulting body, and the
// task with a stamped updated_at.
type mutationEnv struct {
	SchemaVersion string `json:"schema_version"`
	DryRun        bool   `json:"dry_run"`
	Body          string `json:"body"`
	Task          struct {
		Slug    string `json:"slug"`
		Updated string `json:"updated_at"`
		Tier    int    `json:"tier"`
	} `json:"task"`
}

func TestTaskAppend_JSON(t *testing.T) {
	root := setupRepo(t)
	out := runRoot(t, "-C", root, "--json", "task", "append", "alpha", "--body", "## Notes")
	var env mutationEnv
	if err := json.Unmarshal([]byte(out), &env); err != nil {
		t.Fatalf("append --json is not a parseable envelope: %v\n%s", err, out)
	}
	if env.SchemaVersion == "" || env.Task.Slug != "alpha" || env.Task.Updated == "" {
		t.Errorf("append --json envelope wrong:\n%s", out)
	}
	if env.DryRun {
		t.Errorf("a real append should report dry_run=false")
	}
	// The resulting body is echoed and contains the appended section.
	if !strings.Contains(env.Body, "## Notes") {
		t.Errorf("append --json should echo the resulting body:\n%s", out)
	}
}

// --dry-run is now distinguishable in JSON: dry_run=true and nothing written.
func TestTaskSetBody_DryRun_JSON(t *testing.T) {
	root := setupRepo(t)
	before, _ := os.ReadFile(alphaPath(root))
	out := runRoot(t, "-C", root, "--json", "--dry-run", "task", "set", "alpha", "--body", "# Preview")
	var env mutationEnv
	if err := json.Unmarshal([]byte(out), &env); err != nil {
		t.Fatalf("set --body --dry-run --json not parseable: %v\n%s", err, out)
	}
	if !env.DryRun {
		t.Errorf("a --dry-run mutation must report dry_run=true:\n%s", out)
	}
	if !strings.Contains(env.Body, "# Preview") {
		t.Errorf("dry-run should still echo the would-be body:\n%s", out)
	}
	if after, _ := os.ReadFile(alphaPath(root)); !bytes.Equal(before, after) {
		t.Error("--dry-run must not write")
	}
}

// Field-only `task set` is a mutation too: dry_run is present, body is omitted.
func TestTaskSetFields_DryRun_JSON_NoBody(t *testing.T) {
	root := setupRepo(t)
	out := runRoot(t, "-C", root, "--json", "--dry-run", "task", "set", "alpha", "--tier", "2")
	if !strings.Contains(out, `"dry_run":true`) {
		t.Errorf("field-set --dry-run should carry dry_run=true:\n%s", out)
	}
	if strings.Contains(out, `"body"`) {
		t.Errorf("field-set should omit body (omitempty):\n%s", out)
	}
}
