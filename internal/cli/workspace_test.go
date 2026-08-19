package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/andy-esch/taskflow/internal/testutil"
	"github.com/andy-esch/taskflow/internal/wire"
)

// runIn executes the root command anchored at root. stdout and stderr are captured
// SEPARATELY — ambient ⚠ warnings land on stderr, and mixing them into stdout would
// make any --json assertion here a lie about the machine contract.
func runIn(t *testing.T, root string, args ...string) (stdout, stderr string, err error) {
	t.Helper()
	var out, errOut bytes.Buffer
	cmd := NewRootCmd(strings.NewReader(""), &out, &errOut)
	cmd.SetArgs(append([]string{"-C", root}, args...))
	cmd.SetOut(&out)
	cmd.SetErr(&errOut)
	// Execute FIRST: Go evaluates return operands left to right, so returning
	// `out.String(), cmd.Execute()` would capture the buffer before it is written.
	err = cmd.Execute()
	return out.String(), errOut.String(), err
}

func decodeWorkspace(t *testing.T, s string) wire.WorkspaceEnvelope {
	t.Helper()
	var env wire.WorkspaceEnvelope
	if err := json.Unmarshal([]byte(s), &env); err != nil {
		t.Fatalf("decode workspace envelope: %v\n%s", err, s)
	}
	return env
}

// TestWorkspace_ReportsResolvedRoot: the cheap read that answers "which planning
// tree would a mutation hit from here?" before running one.
func TestWorkspace_ReportsResolvedRoot(t *testing.T) {
	root := testutil.NewRepo(t).Task("ready-to-start", "a.md", "---\nstatus: ready-to-start\ndescription: a\ntags: [x]\n---\n").Root

	out, errOut, err := runIn(t, root, "workspace", "--json")
	if err != nil {
		t.Fatalf("workspace: %v\n%s%s", err, out, errOut)
	}
	ws := decodeWorkspace(t, out).Workspace
	if ws.PlanningRoot == "" || !strings.HasSuffix(ws.PlanningRoot, filepath.Base(root)) {
		t.Errorf("planning_root = %q, want the resolved root of %q", ws.PlanningRoot, root)
	}
	if ws.Source != wire.WorkspaceSourceDiscovered {
		t.Errorf("source = %q, want %q for a bare tasks/ tree", ws.Source, wire.WorkspaceSourceDiscovered)
	}
}

// TestWorkspace_PointerIsCalledOut is the case the whole finding is about: the
// directory you stand in is NOT the tree you would change.
func TestWorkspace_PointerIsCalledOut(t *testing.T) {
	planning := testutil.NewRepo(t).Task("ready-to-start", "a.md", "---\nstatus: ready-to-start\ndescription: a\ntags: [x]\n---\n").Root
	if err := os.WriteFile(filepath.Join(planning, ".tskflwctl.toml"), []byte("taskflow_root = \".\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	impl := t.TempDir()
	if err := os.WriteFile(filepath.Join(impl, ".tskflwctl.toml"),
		[]byte("planning_repo = "+strconvQuote(planning)+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	out, errOut, err := runIn(t, impl, "workspace", "--json")
	if err != nil {
		t.Fatalf("workspace: %v\n%s%s", err, out, errOut)
	}
	ws := decodeWorkspace(t, out).Workspace
	if ws.Source != wire.WorkspaceSourcePointer {
		t.Errorf("source = %q, want %q — a planning_repo routing must be visible", ws.Source, wire.WorkspaceSourcePointer)
	}
	if !strings.Contains(ws.PlanningRoot, filepath.Base(planning)) {
		t.Errorf("planning_root = %q, want the POINTED-AT repo %q, not the cwd", ws.PlanningRoot, planning)
	}
}

// TestMutationReceipt_CarriesWorkspace: a successful mutation must prove which
// planning tree it changed, without a separate pre-read.
func TestMutationReceipt_CarriesWorkspace(t *testing.T) {
	root := testutil.NewRepo(t).Task("ready-to-start", "a.md", "---\nstatus: ready-to-start\ndescription: a\ntags: [x]\n---\n").Root

	out, errOut, err := runIn(t, root, "--json", "task", "start", "a")
	if err != nil {
		t.Fatalf("task start: %v\n%s%s", err, out, errOut)
	}
	var env struct {
		Workspace wire.WorkspaceJSON `json:"workspace"`
	}
	if err := json.Unmarshal([]byte(out), &env); err != nil {
		t.Fatalf("decode: %v\n%s", err, out)
	}
	if env.Workspace.PlanningRoot == "" {
		t.Errorf("mutation receipt must name the planning tree it changed, got:\n%s", out)
	}
}

// strconvQuote is strconv.Quote without the import churn in this file.
func strconvQuote(s string) string { return `"` + strings.ReplaceAll(s, `\`, `\\`) + `"` }

// TestCreatedID_IsTheDurableHandle is the H2 contract: `new --json` must hand back an
// id you can still resolve later, not the slug (which changes on rename). Asserting
// the filename prefix alone would be weak — the point is that the id WORKS as a
// lookup key, so this round-trips it through `info` and then renames the entity and
// looks it up again.
func TestCreatedID_IsTheDurableHandle(t *testing.T) {
	root := testutil.NewRepo(t).
		Epic("e1.md", "---\nstatus: active\ndescription: e\n---\n# E1\n").
		Task("ready-to-start", "seed.md", "---\nstatus: ready-to-start\ndescription: seed\nepic: e1\ntags: [x]\n---\n").Root

	out, errOut, err := runIn(t, root, "--json", "task", "new", "Durable Handle", "--epic", "e1", "--tags", "x")
	if err != nil {
		t.Fatalf("task new: %v\n%s%s", err, out, errOut)
	}
	var env struct {
		Created struct{ ID, Slug, Path string } `json:"created"`
	}
	if err := json.Unmarshal([]byte(out), &env); err != nil {
		t.Fatalf("decode: %v\n%s", err, out)
	}
	c := env.Created
	if c.ID == "" || c.ID == c.Slug {
		t.Fatalf("created.id must be the minted id, not the slug: id=%q slug=%q", c.ID, c.Slug)
	}
	if !strings.HasPrefix(c.Path, "tasks/"+c.ID+"-") {
		t.Errorf("created.id %q must lead the filename %q", c.ID, c.Path)
	}

	// The id resolves as a handle right away...
	if out, errOut, err := runIn(t, root, "task", "info", c.ID); err != nil {
		t.Fatalf("created.id must resolve via `task info`: %v\n%s%s", err, out, errOut)
	}

	// ...and, unlike the slug, still resolves after a rename. This is the whole
	// reason the field matters: an agent that stored the slug would now be broken.
	if out, errOut, err := runIn(t, root, "task", "rename", c.Slug, "renamed-handle"); err != nil {
		t.Fatalf("rename: %v\n%s%s", err, out, errOut)
	}
	if out, errOut, err := runIn(t, root, "task", "info", c.ID); err != nil {
		t.Errorf("created.id must survive a rename: %v\n%s%s", err, out, errOut)
	}
	if _, _, err := runIn(t, root, "task", "info", c.Slug); err == nil {
		t.Log("note: the old slug still resolves after rename — id is still the safer handle")
	}
}
