package cli

import (
	"encoding/json"
	"errors"
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
// list reports an explicit healthy-empty state plus the global JSON schema version.
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
	if got := env.Spaces[0]; got.ID != "primary" || got.Path != mutation.Space.Path || got.State != wire.SpaceStateEmpty || got.Root == "" {
		t.Errorf("listed space = %+v", got)
	}

	human, errOut, err := runIn(t, t.TempDir(), "space", "list", "--color=never")
	if err != nil {
		t.Fatalf("human space list: %v\n%s%s", err, human, errOut)
	}
	for _, want := range []string{"primary", "empty", "no planning entities yet", mutation.Space.Path} {
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

func TestSpaceList_GroupsDirectAndPointerEntryPoints(t *testing.T) {
	spaceConfigHome(t)
	planning := initializedSpaceRepo(t)
	implementation := t.TempDir()
	deployment := t.TempDir()
	if _, err := config.InitPointer(implementation, planning, false); err != nil {
		t.Fatal(err)
	}
	if _, err := config.InitPointer(deployment, planning, false); err != nil {
		t.Fatal(err)
	}
	for _, fixture := range []struct {
		id   string
		path string
	}{
		{id: "desirelines", path: implementation},
		{id: "desirelines-planning", path: planning},
		{id: "desirelines-deploy", path: deployment},
	} {
		if out, errOut, err := runIn(t, t.TempDir(), "space", "add", fixture.path, "--id", fixture.id); err != nil {
			t.Fatalf("add %s: %v\n%s%s", fixture.id, err, out, errOut)
		}
	}

	out, errOut, err := runIn(t, t.TempDir(), "space", "list", "--json")
	if err != nil {
		t.Fatalf("JSON list: %v\n%s%s", err, out, errOut)
	}
	env := decodeSpaces(t, out)
	if len(env.Spaces) != 3 {
		t.Fatalf("spaces = %+v", env.Spaces)
	}
	wantRoles := []string{wire.SpaceRolePointer, wire.SpaceRoleDirect, wire.SpaceRolePointer}
	planningID := env.Spaces[0].PlanningID
	if planningID == "" {
		t.Fatal("grouped entries have no derived planning_id")
	}
	for i, entry := range env.Spaces {
		if entry.Role != wantRoles[i] || entry.PlanningID != planningID {
			t.Errorf("entry %d = %+v, want role %q and planning id %q", i, entry, wantRoles[i], planningID)
		}
	}
	// JSON remains in registry order. Human output alone promotes the direct checkout
	// to the group's anchor, with the pointers visibly nested beneath it.
	if env.Spaces[0].ID != "desirelines" || env.Spaces[1].ID != "desirelines-planning" {
		t.Errorf("flat JSON lost registry order: %+v", env.Spaces)
	}

	human, errOut, err := runIn(t, t.TempDir(), "space", "list", "--color=never")
	if err != nil {
		t.Fatalf("human list: %v\n%s%s", err, human, errOut)
	}
	planningAt := strings.Index(human, "desirelines-planning")
	implementationAt := strings.Index(human, "desirelines ")
	deploymentAt := strings.Index(human, "desirelines-deploy")
	if planningAt < 0 || implementationAt < planningAt || deploymentAt < implementationAt {
		t.Errorf("direct checkout did not anchor the grouped tree:\n%s", human)
	}
	if !strings.Contains(human, "  ├─ ") || !strings.Contains(human, "  └─ ") {
		t.Errorf("grouped entry points need tree connectors:\n%s", human)
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
	if _, _, err := userconfig.AddSpace(userconfig.Space{ID: "missing", Path: missing}, false); err != nil {
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
	if env.Spaces[0].Detail == "" || env.Spaces[0].Remedy == "" {
		t.Errorf("broken entry must explain itself and its repair: %+v", env.Spaces[0])
	}
	if spaces, err := userconfig.Spaces(); err != nil || len(spaces) != 1 {
		t.Errorf("listing auto-removed the broken entry: %v / %v", spaces, err)
	}
}

// TestSpaceList_ReportsDurableIDMismatch is the mechanical wrong-repo diagnosis. The
// path still resolves, but the registry assertion names a different durable repo; list
// must remain successful while making the conflict and repair explicit.
func TestSpaceList_ReportsDurableIDMismatch(t *testing.T) {
	spaceConfigHome(t)
	repo := initializedSpaceRepo(t)
	if _, _, err := userconfig.AddSpace(userconfig.Space{
		ID: "drifted", Path: repo, VerifyID: "6gwrongid000",
	}, false); err != nil {
		t.Fatal(err)
	}
	out, errOut, err := runIn(t, t.TempDir(), "space", "list", "--json")
	if err != nil {
		t.Fatalf("mismatched entry must remain listable: %v\n%s%s", err, out, errOut)
	}
	env := decodeSpaces(t, out)
	if len(env.Spaces) != 1 {
		t.Fatalf("spaces = %+v", env.Spaces)
	}
	entry := env.Spaces[0]
	if entry.State != wire.SpaceStateMismatch || entry.Root == "" ||
		!strings.Contains(entry.Detail, "does not match") || entry.Remedy == "" {
		t.Errorf("mismatch diagnosis = %+v", entry)
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
	if _, _, err := userconfig.AddSpace(userconfig.Space{ID: "other", Path: other}, false); err != nil {
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

// TestBrokenSpace_DoesNotBlockOrdinaryCommands pins the advisory boundary at the command
// surface. Only list/doctor inspect registry health; a dead entry cannot prevent work in
// the healthy cwd-anchored planning repo.
func TestBrokenSpace_DoesNotBlockOrdinaryCommands(t *testing.T) {
	spaceConfigHome(t)
	local := initializedSpaceRepo(t)
	missing := filepath.Join(t.TempDir(), "gone")
	if _, _, err := userconfig.AddSpace(userconfig.Space{ID: "missing", Path: missing}, false); err != nil {
		t.Fatal(err)
	}
	out, errOut, err := runIn(t, local, "task", "list", "--json")
	if err != nil {
		t.Fatalf("broken home entry blocked an ordinary command: %v\n%s%s", err, out, errOut)
	}
	if spaces, err := userconfig.Spaces(); err != nil || len(spaces) != 1 {
		t.Errorf("ordinary command mutated registry: spaces=%v err=%v", spaces, err)
	}
}

// TestSpaceMutations_DryRunUsesRegistryRules regresses the review finding at the
// command boundary. A preview is a plan of the real mutation: an existing physical path
// is unchanged, a label collision is still a conflict, and a new entry carries the date
// it would persist. Add and forget previews leave the registry byte-for-byte untouched.
func TestSpaceMutations_DryRunUsesRegistryRules(t *testing.T) {
	home := spaceConfigHome(t)
	registered := initializedSpaceRepo(t)
	conflicting := initializedSpaceRepo(t)
	fresh := initializedSpaceRepo(t)
	if out, errOut, err := runIn(t, registered, "space", "add", "--id", "registered"); err != nil {
		t.Fatalf("register fixture: %v\n%s%s", err, out, errOut)
	}
	registryPath := filepath.Join(home, userconfig.SpacesFile)
	before, err := os.ReadFile(registryPath)
	if err != nil {
		t.Fatal(err)
	}

	out, errOut, err := runIn(t, registered, "space", "add", "--id", "alias", "--dry-run", "--json")
	if err != nil {
		t.Fatalf("same-path preview: %v\n%s%s", err, out, errOut)
	}
	var unchanged wire.SpaceMutationEnvelope
	if err := json.Unmarshal([]byte(out), &unchanged); err != nil {
		t.Fatal(err)
	}
	if !unchanged.DryRun || unchanged.Changed || unchanged.Space.ID != "registered" {
		t.Errorf("same-path preview must report the real no-op: %+v", unchanged)
	}

	out, errOut, err = runIn(t, t.TempDir(), "space", "add", conflicting, "--id", "registered", "--dry-run", "--json")
	if err == nil || ExitCode(err) != 14 {
		t.Fatalf("collision preview must exit 14, got err=%v code=%d\n%s%s", err, ExitCode(err), out, errOut)
	}

	out, errOut, err = runIn(t, t.TempDir(), "space", "add", fresh, "--id", "fresh", "--dry-run", "--json")
	if err != nil {
		t.Fatalf("new-entry preview: %v\n%s%s", err, out, errOut)
	}
	var added wire.SpaceMutationEnvelope
	if err := json.Unmarshal([]byte(out), &added); err != nil {
		t.Fatal(err)
	}
	if !added.DryRun || !added.Changed || added.Space.ID != "fresh" || added.Space.Added == "" {
		t.Errorf("new-entry preview must report the exact would-be entry: %+v", added)
	}

	out, errOut, err = runIn(t, t.TempDir(), "space", "forget", "registered", "--dry-run", "--json")
	if err != nil {
		t.Fatalf("forget preview: %v\n%s%s", err, out, errOut)
	}
	var forgotten wire.SpaceMutationEnvelope
	if err := json.Unmarshal([]byte(out), &forgotten); err != nil {
		t.Fatal(err)
	}
	if !forgotten.DryRun || !forgotten.Changed || forgotten.Space.ID != "registered" {
		t.Errorf("forget preview must report its exact target: %+v", forgotten)
	}

	after, err := os.ReadFile(registryPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(after) != string(before) {
		t.Errorf("dry-run changed spaces.toml:\nbefore:\n%s\nafter:\n%s", before, after)
	}
}

// TestSpaceRegistryErrors_AreClassifiedByCause pins the adapter boundary: malformed
// hand-edited registry content is validation, label reuse is conflict (covered above),
// and an operational error keeps its identity and generic exit code.
func TestSpaceRegistryErrors_AreClassifiedByCause(t *testing.T) {
	home := spaceConfigHome(t)
	if err := os.WriteFile(filepath.Join(home, userconfig.SpacesFile), []byte("[[space]\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	out, errOut, err := runIn(t, t.TempDir(), "space", "list")
	if err == nil || ExitCode(err) != 11 {
		t.Fatalf("malformed registry must exit 11, got err=%v code=%d\n%s%s", err, ExitCode(err), out, errOut)
	}

	operational := errors.New("disk unavailable")
	if got := classifySpaceRegistryError(operational); got != operational || ExitCode(got) != 1 {
		t.Errorf("operational error was reclassified: got=%v code=%d", got, ExitCode(got))
	}
}
