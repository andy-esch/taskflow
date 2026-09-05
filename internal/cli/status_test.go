package cli

import (
	"encoding/json"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/andy-esch/taskflow/internal/config"
	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/testutil"
	"github.com/andy-esch/taskflow/internal/userconfig"
	"github.com/andy-esch/taskflow/internal/wire"
)

// M3 (2026-06-22 audit): status must exit non-zero when files are unreadable,
// matching the list/lint contract — an agent gating on `status` must not get a
// success code on a broken tree. The dashboard still renders first.
func TestStatus_ExitsNonZeroOnUnreadableFiles(t *testing.T) {
	root := setupRepo(t)
	// tier as a quoted string fails the strict decode → a FileProblem.
	path, out := testutil.TaskFixture(root, "ready-to-start", "broken.md",
		"---\nstatus: ready-to-start\ntier: \"4\"\n---\n# Broken\n")
	mustWrite(t, path, out)
	out, err := runRootRC(t, "-C", root, "status")
	if err == nil {
		t.Fatal("status must exit non-zero when a file is unreadable")
	}
	if !strings.Contains(out, "Tasks") {
		t.Errorf("the dashboard should still render before the non-zero exit:\n%s", out)
	}
}

func TestStatus_Smoke(t *testing.T) {
	root := setupRepo(t) // alpha (ready-to-start), beta (in-progress)
	out := runRoot(t, "-C", root, "status")
	for _, want := range []string{"Tasks", "In progress", "beta", "ready-to-start"} {
		if !strings.Contains(out, want) {
			t.Errorf("status output missing %q:\n%s", want, out)
		}
	}
}

func TestStatus_JSON(t *testing.T) {
	root := setupRepo(t)
	out := runRoot(t, "-C", root, "status", "--json")
	var got struct {
		SchemaVersion string           `json:"schema_version"`
		Counts        []map[string]any `json:"counts"`
		InProgress    []map[string]any `json:"in_progress"`
	}
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("bad json: %v\n%s", err, out)
	}
	if got.SchemaVersion == "" || len(got.Counts) == 0 {
		t.Errorf("incomplete summary json: %+v", got)
	}
	if len(got.InProgress) != 1 {
		t.Errorf("expected 1 in-progress task (beta), got %+v", got.InProgress)
	}
}

func TestStatusReportsGraphDegradationWithoutTurningReadIntoFailure(t *testing.T) {
	for _, tc := range []struct {
		name       string
		dependency string
		health     string
		remedy     string
	}{
		{name: "resolved legacy", dependency: "blocked_by: [gate]\n", health: "degraded", remedy: "task depend migrate"},
		{name: "missing canonical", dependency: "depends_on: [6g0000000000]\n", health: "broken", remedy: "run `tskflwctl lint`"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := testutil.NewRepo(t)
			r.Task("completed", "gate.md", "---\nid: "+testutil.TaskID("gate")+"\nstatus: completed\n---\n# gate\n")
			r.Task("ready-to-start", "member.md", "---\nid: "+testutil.TaskID("member")+"\nstatus: ready-to-start\n"+tc.dependency+"---\n# member\n")

			human, err := runRootRC(t, "-C", r.Root, "status", "--color=never")
			if err != nil {
				t.Fatalf("graph health is an informational status read: %v\n%s", err, human)
			}
			for _, want := range []string{"task graph " + tc.health, tc.remedy} {
				if !strings.Contains(human, want) {
					t.Errorf("status missing %q:\n%s", want, human)
				}
			}

			jsonOut := runRoot(t, "-C", r.Root, "status", "--json")
			var envelope wire.SummaryEnvelope
			if err := json.Unmarshal([]byte(jsonOut), &envelope); err != nil {
				t.Fatal(err)
			}
			if envelope.Graph == nil || envelope.Graph.Health != tc.health || !strings.Contains(envelope.Graph.Detail, tc.remedy) {
				t.Fatalf("status graph JSON = %+v\n%s", envelope.Graph, jsonOut)
			}
		})
	}
}

func TestStatusAll_GroupsEntryPointsAndIsolatesBrokenOnes(t *testing.T) {
	spaceConfigHome(t)
	planningA := setupRepo(t)
	if _, err := config.Init(planningA, "", false); err != nil {
		t.Fatal(err)
	}
	cfgA, err := config.Discover(planningA)
	if err != nil {
		t.Fatal(err)
	}
	pointerA := t.TempDir()
	if _, err := config.InitPointer(pointerA, planningA, false); err != nil {
		t.Fatal(err)
	}
	missingA := filepath.Join(t.TempDir(), "gone")

	planningB := testutil.NewRepo(t).
		Task("completed", "gate.md", "---\nid: "+testutil.TaskID("gate")+"\nstatus: completed\n---\n# Gate\n").
		Task("in-progress", "gamma.md",
			"---\nid: "+testutil.TaskID("gamma")+"\nstatus: in-progress\ndescription: gamma\ntags: [seed]\nblocked_by: [gate]\n---\n# Gamma\n").Root
	if _, err := config.Init(planningB, "", false); err != nil {
		t.Fatal(err)
	}
	cfgB, err := config.Discover(planningB)
	if err != nil {
		t.Fatal(err)
	}

	for _, space := range []userconfig.Space{
		{ID: "impl-a", Path: pointerA, VerifyID: cfgA.ID},
		{ID: "planning-a", Path: planningA, VerifyID: cfgA.ID},
		{ID: "missing-a", Path: missingA, VerifyID: cfgA.ID},
		{ID: "planning-b", Path: planningB, VerifyID: cfgB.ID},
	} {
		if added, _, err := userconfig.AddSpace(space, false); err != nil || !added {
			t.Fatalf("register %+v: added=%v err=%v", space, added, err)
		}
	}

	out, errOut, err := runIn(t, t.TempDir(), "status", "--all", "--json")
	if err != nil {
		t.Fatalf("status --all: %v\n%s%s", err, out, errOut)
	}
	var envelope wire.StatusAllEnvelope
	if err := json.Unmarshal([]byte(out), &envelope); err != nil {
		t.Fatalf("decode status --all: %v\n%s", err, out)
	}
	if got := strings.Count(out, `"schema_version"`); got != 1 {
		t.Fatalf("status --all should own one top-level schema version, got %d:\n%s", got, out)
	}
	if envelope.SchemaVersion != wire.SchemaVersion || len(envelope.Spaces) != 2 {
		t.Fatalf("logical spaces = %+v", envelope)
	}
	groupA := envelope.Spaces[0]
	if groupA.ID != "planning-a" || groupA.SelectedEntryPoint != "planning-a" || len(groupA.EntryPoints) != 3 || groupA.Summary == nil {
		t.Fatalf("direct entry did not anchor/load the grouped identity: %+v", groupA)
	}
	if groupA.EntryPoints[2].State != wire.SpaceStateMissing {
		t.Errorf("broken entry point disappeared: %+v", groupA.EntryPoints)
	}
	if len(envelope.InProgress) != 2 {
		t.Fatalf("combined in-progress should contain each planning task once: %+v", envelope.InProgress)
	}
	if envelope.InProgress[0].Space != "planning-a" || envelope.InProgress[1].Space != "planning-b" {
		t.Errorf("combined tasks lost their space badges: %+v", envelope.InProgress)
	}
	if envelope.Spaces[1].Summary.Graph == nil || envelope.Spaces[1].Summary.Graph.Health != "degraded" {
		t.Errorf("cross-space summary lost graph verdict: %+v", envelope.Spaces[1].Summary)
	}

	human, errOut, err := runIn(t, t.TempDir(), "status", "--all", "--color=never")
	if err != nil {
		t.Fatalf("human status --all: %v\n%s%s", err, human, errOut)
	}
	for _, want := range []string{"planning-a", "planning-b", "missing-a", "not found", "[planning-a]", "[planning-b]", "beta", "gamma", "task graph degraded", "task depend migrate"} {
		if !strings.Contains(human, want) {
			t.Errorf("status --all missing %q:\n%s", want, human)
		}
	}
}

func TestStatusAll_ExitsNonZeroAfterRenderingUnreadableFiles(t *testing.T) {
	spaceConfigHome(t)
	root := setupRepo(t)
	if _, err := config.Init(root, "", false); err != nil {
		t.Fatal(err)
	}
	cfg, err := config.Discover(root)
	if err != nil {
		t.Fatal(err)
	}
	path, body := testutil.TaskFixture(root, "ready-to-start", "broken.md",
		"---\nstatus: ready-to-start\ntier: \"4\"\n---\n# Broken\n")
	mustWrite(t, path, body)
	if added, _, err := userconfig.AddSpace(userconfig.Space{ID: "planning", Path: root, VerifyID: cfg.ID}, false); err != nil || !added {
		t.Fatalf("register planning space: added=%v err=%v", added, err)
	}

	out, errOut, err := runIn(t, t.TempDir(), "status", "--all", "--json")
	if err == nil || ExitCode(err) != 11 {
		t.Fatalf("unreadable entity should make cross-space status exit 11: err=%v stdout=%s stderr=%s", err, out, errOut)
	}
	var envelope wire.StatusAllEnvelope
	if decodeErr := json.Unmarshal([]byte(out), &envelope); decodeErr != nil {
		t.Fatalf("dashboard must render before the partial-failure exit: %v\n%s", decodeErr, out)
	}
	if len(envelope.Spaces) != 1 || envelope.Spaces[0].Summary == nil || len(envelope.Spaces[0].Summary.Unreadable) != 1 {
		t.Fatalf("rendered envelope lost unreadable-file details: %+v", envelope)
	}
}

func TestStatusAll_AllBrokenRegistryGroupRemainsInformational(t *testing.T) {
	spaceConfigHome(t)
	missing := filepath.Join(t.TempDir(), "gone")
	if added, _, err := userconfig.AddSpace(userconfig.Space{ID: "missing", Path: missing, VerifyID: "planning-id"}, false); err != nil || !added {
		t.Fatalf("register missing space: added=%v err=%v", added, err)
	}

	out, errOut, err := runIn(t, t.TempDir(), "status", "--all", "--json")
	if err != nil {
		t.Fatalf("registry health should remain informational: %v\n%s%s", err, out, errOut)
	}
	var envelope wire.StatusAllEnvelope
	if err := json.Unmarshal([]byte(out), &envelope); err != nil {
		t.Fatal(err)
	}
	if len(envelope.Spaces) != 1 || envelope.Spaces[0].Summary != nil || envelope.Spaces[0].Error == "" {
		t.Fatalf("broken registry group should be retained with a diagnosis: %+v", envelope)
	}
}

func TestStatusAllProblemsError_SelectedTreeLoadFailure(t *testing.T) {
	selected := core.SpaceEntryPoint{ID: "planning", State: core.SpaceStateOK}
	for _, summary := range []*core.Summary{nil, {}} {
		err := statusAllProblemsError(core.SpaceOverview{Spaces: []core.SpaceSummary{{
			ID: "planning", Selected: &selected, Summary: summary,
			Failure: &core.SpaceLoadFailure{Message: "checkout disappeared"},
		}}})
		if err == nil || !strings.Contains(err.Error(), "1 planning space(s) failed to load") || !errors.Is(err, domain.ErrValidation) {
			t.Fatalf("selected tree load failure with summary=%v should be a partial validation failure, got %v", summary != nil, err)
		}
	}
}

func TestStatusAll_EmptyRegistryFallsBackExactly(t *testing.T) {
	spaceConfigHome(t)
	root := setupRepo(t)
	wantOut, wantErrOut, wantErr := runIn(t, root, "status", "--json")
	gotOut, gotErrOut, gotErr := runIn(t, root, "status", "--all", "--json")
	if (wantErr == nil) != (gotErr == nil) || wantOut != gotOut || wantErrOut != gotErrOut {
		t.Fatalf("empty-registry fallback drifted\nnormal: err=%v stdout=%s stderr=%s\n--all: err=%v stdout=%s stderr=%s",
			wantErr, wantOut, wantErrOut, gotErr, gotOut, gotErrOut)
	}
}

func TestStatusAll_RejectsExplicitSpaceSelection(t *testing.T) {
	spaceConfigHome(t)
	out, errOut, err := runIn(t, t.TempDir(), "status", "--all", "--space", "one")
	if err == nil || ExitCode(err) != 11 || !strings.Contains(err.Error(), "--all and --space") {
		t.Fatalf("conflicting scopes: err=%v stdout=%s stderr=%s", err, out, errOut)
	}
}

func TestGolden_StatusAllJSON(t *testing.T) {
	spaceConfigHome(t)
	planning, err := filepath.Abs(fixtureRepo)
	if err != nil {
		t.Fatal(err)
	}
	cfg, err := config.Discover(planning)
	if err != nil {
		t.Fatal(err)
	}
	pointer := t.TempDir()
	if _, err := config.InitPointer(pointer, planning, false); err != nil {
		t.Fatal(err)
	}
	missing := filepath.Join(t.TempDir(), "gone")
	for _, space := range []userconfig.Space{
		{ID: "implementation", Path: pointer, VerifyID: cfg.ID, Added: "2026-08-21"},
		{ID: "planning", Path: planning, VerifyID: cfg.ID, Added: "2026-08-21"},
		{ID: "missing", Path: missing, VerifyID: cfg.ID, Added: "2026-08-21"},
	} {
		if added, _, err := userconfig.AddSpace(space, false); err != nil || !added {
			t.Fatalf("register %+v: added=%v err=%v", space, added, err)
		}
	}
	out, errOut, err := runIn(t, t.TempDir(), "status", "--all", "--json")
	if err != nil {
		t.Fatalf("status --all golden: %v\n%s%s", err, out, errOut)
	}
	out = redactFixtureRoot(t)(out)
	for _, replacement := range []struct{ from, to string }{
		{filepath.ToSlash(planning), "<ROOT>"},
		{filepath.ToSlash(pointer), "<POINTER>"},
		{filepath.ToSlash(missing), "<MISSING>"},
	} {
		out = strings.ReplaceAll(filepath.ToSlash(out), replacement.from, replacement.to)
	}
	assertGolden(t, "status_all_json", out)
}
