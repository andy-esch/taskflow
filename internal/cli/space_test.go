package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/andy-esch/taskflow/internal/config"
	"github.com/andy-esch/taskflow/internal/userconfig"
	"github.com/andy-esch/taskflow/internal/wire"
)

func spaceConfigHome(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	t.Setenv(userconfig.DirEnv, dir)
	return dir
}

func initializedSpaceRepo(t *testing.T) string {
	t.Helper()
	repo := t.TempDir()
	if _, err := config.Init(repo, "", false); err != nil {
		t.Fatalf("init planning repo: %v", err)
	}
	return repo
}

func decodeSpaces(t *testing.T, text string) wire.SpacesEnvelope {
	t.Helper()
	var env wire.SpacesEnvelope
	if err := json.Unmarshal([]byte(text), &env); err != nil {
		t.Fatalf("decode spaces JSON: %v\n%s", err, text)
	}
	return env
}

// TestSpaceAddAndList_JSON exercises the public surface end-to-end. Adding from a nested
// directory stores the repo marker directory, carries the durable verification id, and
// list reports an explicit healthy state plus the global JSON schema version.
func TestSpaceAddAndList_JSON(t *testing.T) {
	spaceConfigHome(t)
	repo := initializedSpaceRepo(t)
	nested := filepath.Join(repo, "tasks", "nested")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatal(err)
	}

	out, errOut, err := runIn(t, nested, "space", "add", "--id", "primary", "--json")
	if err != nil {
		t.Fatalf("space add: %v\n%s%s", err, out, errOut)
	}
	var mutation wire.SpaceMutationEnvelope
	if err := json.Unmarshal([]byte(out), &mutation); err != nil {
		t.Fatalf("decode mutation: %v\n%s", err, out)
	}
	cfg, err := config.Discover(repo)
	if err != nil {
		t.Fatal(err)
	}
	if !mutation.Changed || mutation.Space.ID != "primary" || mutation.Space.Path != userconfig.TildePath(cfg.Dir) {
		t.Errorf("unexpected add receipt: %+v", mutation)
	}
	if mutation.Space.VerifyID == "" || mutation.Space.VerifyID != cfg.ID {
		t.Errorf("verify_id = %q, want repo id %q", mutation.Space.VerifyID, cfg.ID)
	}

	out, errOut, err = runIn(t, t.TempDir(), "space", "list", "--json")
	if err != nil {
		t.Fatalf("space list outside a repo: %v\n%s%s", err, out, errOut)
	}
	env := decodeSpaces(t, out)
	if env.SchemaVersion != wire.SchemaVersion || len(env.Spaces) != 1 {
		t.Fatalf("unexpected list envelope: %+v", env)
	}
	if got := env.Spaces[0]; got.ID != "primary" || got.Path != mutation.Space.Path || got.State != wire.SpaceStateOK {
		t.Errorf("listed space = %+v", got)
	}

	human, errOut, err := runIn(t, t.TempDir(), "space", "list", "--color=never")
	if err != nil {
		t.Fatalf("human space list: %v\n%s%s", err, human, errOut)
	}
	for _, want := range []string{"primary", "ok", mutation.Space.Path} {
		if !strings.Contains(human, want) {
			t.Errorf("human list must show id, state, and path; missing %q:\n%s", want, human)
		}
	}
}

// TestSpaceAdd_PointerStoresPointerRepo pins the repo-vs-root distinction: registering
// an implementation pointer records the implementation checkout, while discovery and the
// durable id still describe its external planning repo.
func TestSpaceAdd_PointerStoresPointerRepo(t *testing.T) {
	spaceConfigHome(t)
	planning := initializedSpaceRepo(t)
	impl := t.TempDir()
	if _, err := config.InitPointer(impl, planning, false); err != nil {
		t.Fatal(err)
	}
	out, errOut, err := runIn(t, impl, "space", "add", "--id", "pointer", "--json")
	if err != nil {
		t.Fatalf("add pointer: %v\n%s%s", err, out, errOut)
	}
	var env wire.SpaceMutationEnvelope
	if err := json.Unmarshal([]byte(out), &env); err != nil {
		t.Fatal(err)
	}
	resolved, err := config.Discover(impl)
	if err != nil {
		t.Fatal(err)
	}
	if env.Space.Path != userconfig.TildePath(resolved.Dir) || env.Space.Root != resolved.Root {
		t.Errorf("pointer registration lost repo/root distinction: %+v; config=%+v", env.Space, resolved)
	}
	if env.Space.VerifyID != resolved.ID {
		t.Errorf("pointer verify_id = %q, want target repo id %q", env.Space.VerifyID, resolved.ID)
	}
}

// TestSpaceAdd_InvalidPathLeavesNoRegistry proves validation happens before the first
// home-scope write. A bad explicit path must not even create spaces.toml.
func TestSpaceAdd_InvalidPathLeavesNoRegistry(t *testing.T) {
	home := spaceConfigHome(t)
	bad := t.TempDir() // existing, but no marker and no tasks/
	out, errOut, err := runIn(t, t.TempDir(), "space", "add", bad)
	if err == nil {
		t.Fatalf("bad space add should fail:\n%s%s", out, errOut)
	}
	if _, statErr := os.Stat(filepath.Join(home, userconfig.SpacesFile)); !os.IsNotExist(statErr) {
		t.Errorf("bad add must leave no registry behind, stat err=%v", statErr)
	}
}

// TestSpaceForget_RemovesOnlyRegistryEntry exercises the CLI mutation and the disk-safety
// promise together: the entry disappears, while an arbitrary repo file remains untouched.
func TestSpaceForget_RemovesOnlyRegistryEntry(t *testing.T) {
	spaceConfigHome(t)
	repo := initializedSpaceRepo(t)
	marker := filepath.Join(repo, "keep-me.txt")
	if err := os.WriteFile(marker, []byte("still here"), 0o644); err != nil {
		t.Fatal(err)
	}
	if out, errOut, err := runIn(t, repo, "space", "add", "--id", "temporary"); err != nil {
		t.Fatalf("add: %v\n%s%s", err, out, errOut)
	}
	out, errOut, err := runIn(t, t.TempDir(), "space", "forget", "temporary", "--json")
	if err != nil {
		t.Fatalf("forget: %v\n%s%s", err, out, errOut)
	}
	var env wire.SpaceMutationEnvelope
	if err := json.Unmarshal([]byte(out), &env); err != nil || !env.Changed {
		t.Fatalf("forget receipt: %+v / %v\n%s", env, err, out)
	}
	if b, err := os.ReadFile(marker); err != nil || string(b) != "still here" {
		t.Errorf("forget touched the repo: %q / %v", b, err)
	}
	if spaces, err := userconfig.Spaces(); err != nil || len(spaces) != 0 {
		t.Errorf("registry entry remains: %v / %v", spaces, err)
	}
}

// TestSpaceList_ReportsMissingEntry ensures a broken path stays data in the listing — it
// is neither fatal nor silently forgotten.
func TestSpaceList_ReportsMissingEntry(t *testing.T) {
	spaceConfigHome(t)
	missing := filepath.Join(t.TempDir(), "gone")
	if _, _, err := userconfig.AddSpace(userconfig.Space{ID: "missing", Path: missing}); err != nil {
		t.Fatal(err)
	}
	out, errOut, err := runIn(t, t.TempDir(), "space", "list", "--json")
	if err != nil {
		t.Fatalf("broken entry must not fail list: %v\n%s%s", err, out, errOut)
	}
	env := decodeSpaces(t, out)
	if len(env.Spaces) != 1 || env.Spaces[0].State != wire.SpaceStateMissing {
		t.Fatalf("missing entry diagnosis = %+v", env.Spaces)
	}
	if spaces, err := userconfig.Spaces(); err != nil || len(spaces) != 1 {
		t.Errorf("listing auto-removed the broken entry: %v / %v", spaces, err)
	}
}

// TestSpaceRegistry_DoesNotAffectDiscover is the advisory invariant as an equality test:
// populating the home registry cannot change any field cwd-anchored discovery resolves.
func TestSpaceRegistry_DoesNotAffectDiscover(t *testing.T) {
	spaceConfigHome(t)
	local := initializedSpaceRepo(t)
	other := initializedSpaceRepo(t)
	before, err := config.Discover(local)
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := userconfig.AddSpace(userconfig.Space{ID: "other", Path: other}); err != nil {
		t.Fatal(err)
	}
	after, err := config.Discover(local)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(before, after) {
		t.Errorf("registry changed local discovery:\nbefore=%+v\nafter=%+v", before, after)
	}
}
