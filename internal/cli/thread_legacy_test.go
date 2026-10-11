package cli

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/andy-esch/taskflow/internal/testutil"
	"github.com/andy-esch/taskflow/internal/wire"
)

func TestThreadLegacyDiagnosticsAcrossRealAdapterAndMigration(t *testing.T) {
	r := testutil.NewRepo(t)
	resolvedRoot, err := filepath.EvalSymlinks(r.Root)
	if err != nil {
		t.Fatal(err)
	}
	member, owner, peer := testutil.TaskID("legacy-member"), testutil.TaskID("legacy-outside"), testutil.TaskID("legacy-peer")
	r.Task("next-up", "legacy-member.md", "---\nid: "+member+"\ndescription: Independent member\n---\n# Member\n")
	r.Task("completed", "legacy-peer.md", "---\nid: "+peer+"\ndescription: Outside prerequisite\n---\n# Peer\n")
	r.Task("next-up", "legacy-outside.md", "---\nid: "+owner+"\ndescription: Legacy field outside the Thread\nblocked_by: [legacy-peer]\n---\n# Outside\n")
	for _, slug := range []string{"delivery", "other"} {
		r.File("threads/"+testutil.TaskID("legacy-"+slug)+"-"+slug+".md", fmt.Sprintf(
			"---\nid: %s\nstatus: in-progress\ndescription: Independent delivery\ngoal: Ship\ncreated: '2026-10-11'\ntasks: [%s]\n---\n# %s\n", testutil.TaskID("legacy-"+slug), member, slug))
	}
	before := testutil.SnapshotTree(t, r.Root)
	for _, surface := range []string{"show", "frontier", "graph", "plan"} {
		out, stderr, err := runIn(t, r.Root, "thread", surface, "delivery", "--json")
		if err != nil || stderr != "" {
			t.Fatalf("%s failed: %v %s", surface, err, stderr)
		}
		var payload struct {
			View       wire.ThreadViewJSON            `json:"view"`
			Projection wire.ThreadGraphProjectionJSON `json:"projection"`
		}
		if err := json.Unmarshal([]byte(out), &payload); err != nil {
			t.Fatal(err)
		}
		view := payload.View
		if surface == "graph" || surface == "plan" {
			view = payload.Projection.View
		}
		if view.GraphHealth != "degraded" || view.ProjectionHealth != "degraded" || len(view.Frontier) != 0 || len(view.GraphProblems) != 0 || len(view.GraphLegacy) != 1 || view.GraphLegacy[0].TaskID != owner || view.GraphLegacy[0].References[0].PrerequisiteID != peer || len(view.ExternalGates) != 0 {
			t.Fatalf("%s lost outside cause or changed eligibility: %+v", surface, view)
		}
		if surface == "frontier" {
			redacted := strings.ReplaceAll(out, `\\`, "/")
			assertGolden(t, "thread_frontier_legacy_json", strings.ReplaceAll(redacted, filepath.ToSlash(resolvedRoot), "<ROOT>"))
		}
		if surface != "graph" {
			human := runRoot(t, "-C", r.Root, "thread", surface, "delivery")
			if strings.Count(human, "legacy blocked_by on "+owner) != 1 || !strings.Contains(human, "run task depend migrate") {
				t.Fatalf("%s lacks unique diagnosis/remedy:\n%s", surface, human)
			}
			if surface == "frontier" {
				assertGolden(t, "thread_frontier_legacy_human", human)
			}
		}
	}
	for _, flags := range [][]string{nil, {"--search", "not-present"}, {"--limit", "1"}, {"--limit", "1", "-c", "id"}, {"--search", "not-present", "-c", "id"}} {
		args := append([]string{"thread", "list", "--json"}, flags...)
		out, stderr, err := runIn(t, r.Root, args...)
		if err != nil || stderr != "" {
			t.Fatalf("list %v: %v %s", flags, err, stderr)
		}
		var payload struct {
			GraphLegacy []wire.LegacyDependencyJSON  `json:"graph_legacy"`
			Threads     []map[string]json.RawMessage `json:"threads"`
		}
		if err := json.Unmarshal([]byte(out), &payload); err != nil {
			t.Fatal(err)
		}
		if len(payload.GraphLegacy) != 1 || payload.GraphLegacy[0].TaskID != owner {
			t.Fatalf("list %v lost repository diagnosis: %s", flags, out)
		}
		for _, row := range payload.Threads {
			if raw, ok := row["graph_legacy"]; ok && string(raw) != "[]" {
				t.Fatalf("list repeated repository cause per row: %s", out)
			}
		}
	}
	human := runRoot(t, "-C", r.Root, "thread", "list")
	if strings.Count(human, "legacy blocked_by on "+owner) != 1 {
		t.Fatalf("human list duplicated repository diagnosis:\n%s", human)
	}
	for _, mode := range []string{"name", "table", "csv"} {
		res, err := runRootStreams(t, "-C", r.Root, "thread", "list", "--search", "not-present", "-o", mode)
		if err != nil || !strings.Contains(res.Err, "run task depend migrate") || strings.Contains(res.Out, "legacy blocked_by") {
			t.Fatalf("empty %s stream lost stderr diagnosis: %+v %v", mode, res, err)
		}
	}
	if !reflect.DeepEqual(before, testutil.SnapshotTree(t, r.Root)) {
		t.Fatal("diagnostic reads wrote to storage")
	}
	preview := runRoot(t, "-C", r.Root, "task", "depend", "migrate", "--dry-run", "--json")
	if !strings.Contains(preview, "blocked_by") || !reflect.DeepEqual(before, testutil.SnapshotTree(t, r.Root)) {
		t.Fatalf("migration preview lost intent or wrote: %s", preview)
	}
	runRoot(t, "-C", r.Root, "task", "depend", "migrate")
	var frontier wire.ThreadFrontierEnvelope
	if err := json.Unmarshal([]byte(runRoot(t, "-C", r.Root, "thread", "frontier", "delivery", "--json")), &frontier); err != nil {
		t.Fatal(err)
	}
	if frontier.View.GraphHealth != "healthy" || len(frontier.View.GraphLegacy) != 0 || len(frontier.View.Frontier) != 1 || frontier.View.Frontier[0].Task.TaskID != member {
		t.Fatalf("advertised migration did not restore frontier: %+v", frontier.View)
	}
}

func TestThreadLegacyBrokenReferencesRemainExplanatoryAndFailClosed(t *testing.T) {
	r := testutil.NewRepo(t)
	member, owner := testutil.TaskID("broken-member"), testutil.TaskID("broken-owner")
	r.Task("next-up", "broken-member.md", "---\nid: "+member+"\ndescription: Member\n---\n")
	r.Task("next-up", "broken-owner.md", "---\nid: "+owner+"\ndescription: Outside broken references\nblocked_by: [gone, shared]\ndepends_on: [invalid-token]\n---\n")
	for _, seed := range []string{"first", "second"} {
		r.File("tasks/"+testutil.TaskID(seed)+"-shared.md", "---\nid: "+testutil.TaskID(seed)+"\nstatus: completed\ndescription: Ambiguous label\n---\n")
	}
	r.File("threads/"+testutil.TaskID("broken-thread")+"-delivery.md", fmt.Sprintf("---\nid: %s\nstatus: unstarted\ndescription: Delivery\ngoal: Ship\ncreated: '2026-10-11'\ntasks: [%s]\n---\n", testutil.TaskID("broken-thread"), member))
	out := runRoot(t, "-C", r.Root, "thread", "frontier", "delivery")
	for _, want := range []string{`"gone": missing`, `"shared": ambiguous`, "invalid-dependency-id", "run task depend repair, then task depend migrate"} {
		if !strings.Contains(out, want) {
			t.Fatalf("broken frontier lacks %q:\n%s", want, out)
		}
	}
	if strings.Count(out, "legacy blocked_by on "+owner) != 1 || strings.Contains(out, "legacy-reference-missing:") || strings.Contains(out, "legacy-reference-ambiguous:") || strings.Contains(out, "✔") {
		t.Fatalf("duplicated causes or dispatched broken work:\n%s", out)
	}
	var payload wire.ThreadFrontierEnvelope
	if err := json.Unmarshal([]byte(runRoot(t, "-C", r.Root, "thread", "frontier", "delivery", "--json")), &payload); err != nil {
		t.Fatal(err)
	}
	codes := make([]string, 0, len(payload.View.GraphProblems))
	for _, problem := range payload.View.GraphProblems {
		codes = append(codes, problem.Code)
	}
	if payload.View.GraphHealth != "broken" || len(payload.View.Frontier) != 0 || len(payload.View.GraphLegacy) != 1 || !slices.Contains(codes, "legacy-reference-missing") || !slices.Contains(codes, "legacy-reference-ambiguous") || !slices.Contains(codes, "invalid-dependency-id") {
		t.Fatalf("machine projection lost distinct structural evidence: %+v", payload.View)
	}
	before := testutil.SnapshotTree(t, r.Root)
	if _, err := runRootRC(t, "-C", r.Root, "task", "depend", "migrate"); err == nil {
		t.Fatal("ambiguous/missing intent migrated without selection")
	}
	if !reflect.DeepEqual(before, testutil.SnapshotTree(t, r.Root)) {
		t.Fatal("refused migration wrote to storage")
	}
}

func TestThreadLegacyRepairReceiptsCarryDiagnosisAcrossRealAdapter(t *testing.T) {
	r := testutil.NewRepo(t)
	member, owner := testutil.TaskID("repair-legacy-member"), testutil.TaskID("repair-legacy-outside")
	r.Task("next-up", "repair-legacy-member.md", "---\nid: "+member+"\ndescription: Independent member\n---\n")
	r.Task("completed", "repair-legacy-outside.md", "---\nid: "+owner+"\ndescription: Empty field outside membership\nblocks: []\n---\n")
	r.File("threads/"+testutil.TaskID("repair-legacy-thread")+"-delivery.md", fmt.Sprintf(
		"---\nid: %s\nstatus: unstarted\ndescription: Delivery\ngoal: Ship\ncreated: '2026-10-11'\ntasks: [%s]\n---\n", testutil.TaskID("repair-legacy-thread"), member))
	before := testutil.SnapshotTree(t, r.Root)
	for _, dryRun := range []bool{true, false} {
		args := []string{"task", "depend", "repair", "--auto", "--json"}
		if dryRun {
			args = append(args, "--dry-run")
		}
		out, stderr, err := runIn(t, r.Root, args...)
		if err != nil || stderr != "" {
			t.Fatalf("repair dry-run=%t: %v %s", dryRun, err, stderr)
		}
		var receipt wire.TaskGraphRepairEnvelope
		if err := json.Unmarshal([]byte(out), &receipt); err != nil {
			t.Fatal(err)
		}
		if receipt.DryRun != dryRun || receipt.Committed == dryRun || receipt.InitialHealth != "degraded" || receipt.FinalHealth != "healthy" || len(receipt.ThreadImpacts) != 1 {
			t.Fatalf("repair lost global health impact: %+v", receipt)
		}
		impact := receipt.ThreadImpacts[0]
		if impact.Direct || len(impact.Before.GraphLegacy) != 1 || impact.Before.GraphLegacy[0].TaskID != owner || impact.Before.GraphLegacy[0].Field != "blocks" || len(impact.Before.Frontier) != 0 || len(impact.After.GraphLegacy) != 0 || len(impact.After.Frontier) != 1 || impact.After.Frontier[0].Task.TaskID != member {
			t.Fatalf("repair receipt lost diagnosis or restored eligibility: %+v", impact)
		}
		if dryRun && !reflect.DeepEqual(before, testutil.SnapshotTree(t, r.Root)) {
			t.Fatal("repair preview wrote to storage")
		}
	}
	var frontier wire.ThreadFrontierEnvelope
	if err := json.Unmarshal([]byte(runRoot(t, "-C", r.Root, "thread", "frontier", "delivery", "--json")), &frontier); err != nil {
		t.Fatal(err)
	}
	if frontier.View.GraphHealth != "healthy" || len(frontier.View.GraphLegacy) != 0 || len(frontier.View.Frontier) != 1 {
		t.Fatalf("repair receipt diverged from durable frontier: %+v", frontier.View)
	}
}
