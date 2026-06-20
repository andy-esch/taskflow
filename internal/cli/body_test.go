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

// `task append --body-file -` reads the appended content from stdin.
func TestTaskAppend_Stdin(t *testing.T) {
	root := setupRepo(t)
	var out bytes.Buffer
	cmd := NewRootCmd(&out, &out)
	cmd.SetIn(strings.NewReader("## Notes\n- from stdin\n"))
	cmd.SetArgs([]string{"-C", root, "task", "append", "alpha", "--body-file", "-"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	b, _ := os.ReadFile(alphaPath(root))
	if !strings.Contains(string(b), "from stdin") {
		t.Errorf("append body should come from stdin:\n%s", b)
	}
}

// Empty append input is a clean validation error, not an empty write.
func TestTaskAppend_Empty_Errors(t *testing.T) {
	root := setupRepo(t)
	var out bytes.Buffer
	cmd := NewRootCmd(&out, &out)
	cmd.SetArgs([]string{"-C", root, "task", "append", "alpha", "--body", "   "})
	if err := cmd.Execute(); !errors.Is(err, domain.ErrValidation) {
		t.Errorf("empty append should wrap ErrValidation (exit 11), got %v", err)
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
	cmd := NewRootCmd(&out, &out)
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

// The body-mutation --json output is a parseable task envelope with updated_at.
func TestTaskAppend_JSON(t *testing.T) {
	root := setupRepo(t)
	out := runRoot(t, "-C", root, "--json", "task", "append", "alpha", "--body", "## Notes")
	var env struct {
		SchemaVersion string `json:"schema_version"`
		Task          struct {
			Slug    string `json:"slug"`
			Updated string `json:"updated_at"`
		} `json:"task"`
	}
	if err := json.Unmarshal([]byte(out), &env); err != nil {
		t.Fatalf("append --json is not a parseable envelope: %v\n%s", err, out)
	}
	if env.SchemaVersion == "" || env.Task.Slug != "alpha" {
		t.Errorf("append --json envelope wrong:\n%s", out)
	}
	if env.Task.Updated == "" {
		t.Errorf("append should stamp updated_at in the JSON task:\n%s", out)
	}
}
