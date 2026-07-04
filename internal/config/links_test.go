package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestCheckLinks_SubdirPlanningLayout regresses the review finding: a planning
// repo whose config sits ABOVE its root (taskflow_root subdir) must still link
// cleanly in BOTH directions — no false-positive ⚠ on the impl side — and
// link-back must land in the config dir, not the root.
func TestCheckLinks_SubdirPlanningLayout(t *testing.T) {
	parent := t.TempDir()
	plan := filepath.Join(parent, "plan")                  // config lives here
	mustMkdir(t, filepath.Join(plan, "planning", "tasks")) // root is plan/planning
	writeConfig(t, plan, "taskflow_root = \"./planning\"\n")
	impl := filepath.Join(parent, "impl")
	mustMkdir(t, impl)

	if _, err := InitPointer(impl, "../plan/planning", false); err != nil {
		t.Fatal(err)
	}
	back, err := LinkBack(impl, "../plan/planning", false)
	if err != nil || back == "" {
		t.Fatalf("link-back should write into the subdir planning's config dir, got %q / %v", back, err)
	}
	if b, _ := os.ReadFile(filepath.Join(plan, ConfigFile)); !strings.Contains(string(b), "../impl") {
		t.Errorf("back-link must land in the config dir (plan/), not the root:\n%s", b)
	}
	if p := linksAt(t, impl); len(p) != 0 {
		t.Errorf("impl side must be clean for a subdir planning layout, got %v", p)
	}
	if p := linksAt(t, plan); len(p) != 0 {
		t.Errorf("planning side must be clean for a subdir planning layout, got %v", p)
	}
}

// TestCheckLinks_SubdirPlanning_ImplPointsAtRepoRoot regresses the desirelines case: a
// planning repo with taskflow_root="planning" whose tracked impl points its planning_repo at
// the REPO ROOT (not the taskflow_root subdir) must link cleanly. The planning-side check must
// accept the impl naming either the repo (Dir) or the subdir (Root) — comparing only against
// Root falsely flagged "points its planning_repo elsewhere".
func TestCheckLinks_SubdirPlanning_ImplPointsAtRepoRoot(t *testing.T) {
	parent := t.TempDir()
	plan := filepath.Join(parent, "plan")
	mustMkdir(t, filepath.Join(plan, "planning", "tasks")) // entities under plan/planning
	writeConfig(t, plan, "taskflow_root = \"planning\"\ntracked_repos = [\"../impl\"]\n")
	impl := filepath.Join(parent, "impl")
	mustMkdir(t, impl)
	// Write the impl config directly: its planning_repo names the planning REPO ROOT
	// (../plan). This is a valid LEGACY pointer — created by InitPointer back when tasks/
	// sat at the root, before an isolation moved them under taskflow_root. Discover still
	// resolves it (it reads plan's config + taskflow_root); the link check must accept it.
	// (InitPointer can't recreate it now — it rejects a target with no tasks/ directly under
	// it — so the config is written by hand here to mirror the real desirelines state.)
	writeConfig(t, impl, "planning_repo = \"../plan\"\n")
	if p := linksAt(t, plan); len(p) != 0 {
		t.Errorf("planning side must accept an impl that points at the repo root, got %v", p)
	}
}

// linksAt discovers cfg at dir and returns CheckLinks(cfg).
func linksAt(t *testing.T, dir string) []LinkProblem {
	t.Helper()
	cfg, err := Discover(dir)
	if err != nil {
		t.Fatalf("Discover(%s): %v", dir, err)
	}
	return CheckLinks(cfg)
}

// TestCheckLinks_Consistent: a fully linked impl <-> planning pair has no
// problems from EITHER side, and absolute/relative spellings don't false-positive.
func TestCheckLinks_Consistent(t *testing.T) {
	parent := t.TempDir()
	planning := filepath.Join(parent, "planning")
	impl := filepath.Join(parent, "impl")
	mustMkdir(t, filepath.Join(planning, "tasks"))
	mustMkdir(t, impl)
	if _, err := Init(planning, false); err != nil {
		t.Fatal(err)
	}
	// Point via an ABSOLUTE path; link-back stores the relative form — physical
	// comparison must treat them as the same repo.
	if _, err := InitPointer(impl, planning, false); err != nil {
		t.Fatal(err)
	}
	if _, err := LinkBack(impl, planning, false); err != nil {
		t.Fatal(err)
	}
	if p := linksAt(t, impl); len(p) != 0 {
		t.Errorf("consistent impl should be clean, got %v", p)
	}
	if p := linksAt(t, planning); len(p) != 0 {
		t.Errorf("consistent planning should be clean, got %v", p)
	}
}

// TestCheckLinks_OneSided: an impl points at a planning repo that doesn't track
// it back — flagged from the impl side.
func TestCheckLinks_OneSided(t *testing.T) {
	parent := t.TempDir()
	planning := filepath.Join(parent, "planning")
	impl := filepath.Join(parent, "impl")
	mustMkdir(t, filepath.Join(planning, "tasks"))
	mustMkdir(t, impl)
	if _, err := Init(planning, false); err != nil {
		t.Fatal(err)
	}
	if _, err := InitPointer(impl, "../planning", false); err != nil { // no link-back
		t.Fatal(err)
	}
	p := linksAt(t, impl)
	if len(p) != 1 || !strings.Contains(p[0].Message, "one-sided") {
		t.Errorf("expected a one-sided link problem, got %v", p)
	}
	// The planning side doesn't track impl, so it has nothing to complain about.
	if p := linksAt(t, planning); len(p) != 0 {
		t.Errorf("planning that tracks nothing should be clean, got %v", p)
	}
}

// TestCheckLinks_PlanningSide covers the tracked-side failure modes: dangling,
// no-config, no-planning_repo, and points-elsewhere.
func TestCheckLinks_PlanningSide(t *testing.T) {
	parent := t.TempDir()
	planning := filepath.Join(parent, "planning")
	mustMkdir(t, filepath.Join(planning, "tasks"))
	if _, err := Init(planning, false); err != nil {
		t.Fatal(err)
	}

	// dangling: tracked repo doesn't exist.
	if _, err := AddTrackedRepo(planning, "../ghost", false); err != nil {
		t.Fatal(err)
	}
	// no-config: tracked dir exists but has no .tskflwctl.toml.
	mustMkdir(t, filepath.Join(parent, "bare"))
	if _, err := AddTrackedRepo(planning, "../bare", false); err != nil {
		t.Fatal(err)
	}
	// no-planning_repo: tracked repo is itself a scaffold (no pointer back).
	other := filepath.Join(parent, "other")
	mustMkdir(t, filepath.Join(other, "tasks"))
	if _, err := Init(other, false); err != nil {
		t.Fatal(err)
	}
	if _, err := AddTrackedRepo(planning, "../other", false); err != nil {
		t.Fatal(err)
	}
	// points-elsewhere: tracked impl points at a DIFFERENT planning repo.
	elsewhere := filepath.Join(parent, "elsewhere")
	mustMkdir(t, filepath.Join(elsewhere, "tasks"))
	if _, err := Init(elsewhere, false); err != nil {
		t.Fatal(err)
	}
	impl := filepath.Join(parent, "impl")
	mustMkdir(t, impl)
	if _, err := InitPointer(impl, "../elsewhere", false); err != nil {
		t.Fatal(err)
	}
	if _, err := AddTrackedRepo(planning, "../impl", false); err != nil {
		t.Fatal(err)
	}

	got := make(map[string]string)
	for _, p := range linksAt(t, planning) {
		got[p.Repo] = p.Message
	}
	for _, tc := range []struct{ repo, want string }{
		{"../ghost", "does not exist"},
		{"../bare", "has no"},
		{"../other", "does not point back"},
		{"../impl", "points its planning_repo elsewhere"},
	} {
		if msg, ok := got[tc.repo]; !ok || !strings.Contains(msg, tc.want) {
			t.Errorf("%s: want a problem containing %q, got %q (ok=%v)", tc.repo, tc.want, msg, ok)
		}
	}
}

// TestCheckLinks_NoLinks: a repo with no planning_repo and no tracked_repos (or
// no config at all) has nothing to check.
func TestCheckLinks_NoLinks(t *testing.T) {
	repo := t.TempDir()
	mustMkdir(t, filepath.Join(repo, "tasks"))
	if _, err := Init(repo, false); err != nil {
		t.Fatal(err)
	}
	if p := linksAt(t, repo); len(p) != 0 {
		t.Errorf("a plain planning repo should have no link problems, got %v", p)
	}
	// A bare tasks/ dir with no config → no Dir → nil.
	bare := t.TempDir()
	mustMkdir(t, filepath.Join(bare, "tasks"))
	if p := linksAt(t, bare); p != nil {
		t.Errorf("a config-less repo should yield nil link problems, got %v", p)
	}
}
