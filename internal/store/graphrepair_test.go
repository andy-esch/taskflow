package store

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/testutil"
)

func TestMutateTaskGraphRepairDryRunThenAppliesExactSourceEdits(t *testing.T) {
	root := t.TempDir()
	prerequisiteID := testutil.TaskID("repair-store-prerequisite")
	writeGraphMutationTask(t, root, "repair-store-prerequisite", domain.StatusCompleted, nil, "")
	ownerPath := writeGraphMutationTask(t, root, "repair-store-owner", domain.StatusReadyToStart,
		[]string{prerequisiteID, prerequisiteID, "human-invalid-token"},
		"custom_key: keep-me # source comment\nblocked_by: [] # remove only when selected\n")
	before, _ := os.ReadFile(ownerPath)
	service := core.MustNewService(testutil.Must(NewFS(root, core.UnrestrictedMutations())), core.WithClock(func() time.Time { return graphMutationNow }))

	dry, err := service.RepairTaskGraph(core.TaskGraphRepairRequest{Auto: true}, true)
	if err != nil {
		t.Fatal(err)
	}
	afterDry, _ := os.ReadFile(ownerPath)
	if !dry.Changed || dry.Committed || len(dry.PlannedFiles) != 1 || !slices.Equal(before, afterDry) {
		t.Fatalf("dry-run receipt=%+v changed file=%v", dry, !slices.Equal(before, afterDry))
	}
	if !hasStoreRepairReason(dry.Residual, core.RepairInvalidID) {
		t.Fatalf("explicit defect missing from dry-run residual: %+v", dry.Residual)
	}

	receipt, err := service.RepairTaskGraph(core.TaskGraphRepairRequest{Auto: true}, false)
	if err != nil {
		t.Fatal(err)
	}
	if !receipt.Changed || !receipt.Committed || len(receipt.AppliedFiles) != 1 || receipt.FinalHealth != core.GraphBroken {
		t.Fatalf("repair receipt = %+v", receipt)
	}
	content, _ := os.ReadFile(ownerPath)
	for _, preserved := range []string{"human-invalid-token", "custom_key: keep-me # source comment", "Body stays intact."} {
		if !strings.Contains(string(content), preserved) {
			t.Fatalf("repair lost %q:\n%s", preserved, content)
		}
	}
	for _, removed := range []string{"blocked_by:", prerequisiteID + ", " + prerequisiteID} {
		if strings.Contains(string(content), removed) {
			t.Fatalf("repair retained %q:\n%s", removed, content)
		}
	}
	if !strings.Contains(string(content), `updated_at: "2026-08-27"`) {
		t.Fatalf("repair did not stamp source:\n%s", content)
	}
}

func TestMutateTaskGraphRepairRetainsSourceIdentityWithBrokenDeclaredID(t *testing.T) {
	for _, tc := range []struct {
		name       string
		declaredID string
		problem    core.GraphProblemCode
	}{
		{name: "missing", problem: core.ProblemMissingTaskID},
		{name: "drifted", declaredID: testutil.TaskID("repair-drifted-declaration"), problem: core.ProblemTaskIDDrift},
	} {
		for _, dryRun := range []bool{true, false} {
			t.Run(fmt.Sprintf("%s/dry-run=%t", tc.name, dryRun), func(t *testing.T) {
				root := t.TempDir()
				prerequisiteID := testutil.TaskID("repair-source-id-prerequisite")
				ownerID := testutil.TaskID("repair-source-id-owner")
				writeGraphMutationTask(t, root, "repair-source-id-prerequisite", domain.StatusCompleted, nil, "")
				path := writeGraphMutationTask(t, root, "repair-source-id-owner", domain.StatusNextUp,
					[]string{prerequisiteID, prerequisiteID}, "custom_key: keep-me # source comment\n")
				content, err := os.ReadFile(path)
				if err != nil {
					t.Fatal(err)
				}
				declaration := ""
				if tc.declaredID != "" {
					declaration = "id: " + tc.declaredID + "\n"
				}
				before := strings.Replace(string(content), "id: "+ownerID+"\n", declaration, 1)
				testutil.Write(t, path, before)
				fs := testutil.Must(NewFS(root, core.UnrestrictedMutations()))
				service := core.MustNewService(fs, core.WithClock(func() time.Time { return graphMutationNow }))

				receipt, err := service.RepairTaskGraph(core.TaskGraphRepairRequest{Auto: true}, dryRun)
				if err != nil {
					t.Fatal(err)
				}
				if !receipt.Changed || receipt.Committed == dryRun || receipt.InitialHealth != core.GraphBroken || receipt.FinalHealth != core.GraphBroken {
					t.Fatalf("repair receipt = %+v", receipt)
				}
				foundIdentityProblem := false
				for _, problem := range receipt.Problems {
					if problem.Code == tc.problem && problem.TaskID == ownerID && problem.Path == path {
						foundIdentityProblem = true
					}
					if problem.Code == core.ProblemDuplicateDependency {
						t.Fatalf("dependency defect survived prospective repair: %+v", receipt.Problems)
					}
				}
				if !foundIdentityProblem {
					t.Fatalf("residual identity problem lost source attribution: %+v", receipt.Problems)
				}
				if len(receipt.Operations) != 1 || receipt.Operations[0].Edit.Source.TaskID != ownerID || receipt.Operations[0].Edit.Source.LocalPath != path {
					t.Fatalf("repair lost source identity: %+v", receipt.Operations)
				}
				reloaded, err := fs.ReadTask(ownerID)
				if err != nil {
					t.Fatal(err)
				}
				wantDependencies := []string{prerequisiteID}
				if dryRun {
					wantDependencies = append(wantDependencies, prerequisiteID)
				}
				if reloaded.Source.ID != ownerID || reloaded.Value.Task.ID != tc.declaredID || !slices.Equal(reloaded.Value.Task.DependsOn, wantDependencies) {
					t.Fatalf("repair changed declaration or failed to reload by source ID: %+v", reloaded)
				}
				after, err := os.ReadFile(path)
				if err != nil {
					t.Fatal(err)
				}
				if dryRun && string(after) != before {
					t.Fatal("dry-run changed the source")
				}
				for _, preserved := range []string{"custom_key: keep-me # source comment", "Body stays intact."} {
					if !strings.Contains(string(after), preserved) {
						t.Fatalf("repair lost %q:\n%s", preserved, after)
					}
				}
			})
		}
	}
}

func TestMutateTaskGraphRepairAllowsMalformedThreadsButCASProtectsTheirBytes(t *testing.T) {
	root := t.TempDir()
	ownerID := testutil.TaskID("repair-thread-evidence-owner")
	ownerPath := writeGraphMutationTask(t, root, "repair-thread-evidence-owner", domain.StatusReadyToStart,
		[]string{"invalid-token"}, "")
	threadID := testutil.TaskID("repair-thread-evidence")
	threadPath := filepath.Join(root, domain.ThreadsDir, threadID+"-repair-thread-evidence.md")
	testutil.Write(t, threadPath, "---\nid: [unterminated\n---\n# Broken Thread\n")
	fs := testutil.Must(NewFS(root, core.UnrestrictedMutations()))
	request := core.TaskGraphRepairRequest{Edits: []core.TaskGraphSourceEdit{{
		Action: core.TaskGraphSourceDropDeclaration, Source: core.TaskGraphSourceRef{TaskID: ownerID},
		Field: core.TaskDependencyDependsOn, Value: "invalid-token",
	}}}

	dry, err := core.MustNewService(fs).RepairTaskGraph(request, true)
	if err != nil || len(dry.IncompleteThreads) != 1 {
		t.Fatalf("malformed Thread dry run = %+v, %v", dry, err)
	}
	original := testHookBeforeGraphRepairWrite
	t.Cleanup(func() { testHookBeforeGraphRepairWrite = original })
	testHookBeforeGraphRepairWrite = func(core.TaskGraphSourceRef) {
		testHookBeforeGraphRepairWrite = nil
		testutil.Write(t, threadPath, "---\nid: [unterminated\n---\n# Concurrent bytes\n")
	}
	before, _ := os.ReadFile(ownerPath)
	receipt, err := core.MustNewService(fs, core.WithRetry(0, func(int) {})).RepairTaskGraph(request, false)
	if !errors.Is(err, domain.ErrConflict) || receipt.Changed || receipt.Committed || len(receipt.AppliedFiles) != 0 {
		t.Fatalf("Thread CAS receipt=%+v err=%v", receipt, err)
	}
	after, _ := os.ReadFile(ownerPath)
	if !slices.Equal(before, after) {
		t.Fatal("Thread evidence conflict still changed task source")
	}
}

func TestMutateTaskGraphRepairReportsReadableThreadProjectionImpacts(t *testing.T) {
	root := t.TempDir()
	prerequisiteID := testutil.TaskID("repair-thread-impact-prerequisite")
	ownerID := testutil.TaskID("repair-thread-impact-owner")
	writeGraphMutationTask(t, root, "repair-thread-impact-prerequisite", domain.StatusCompleted, nil, "")
	writeGraphMutationTask(t, root, "repair-thread-impact-owner", domain.StatusReadyToStart,
		[]string{prerequisiteID, prerequisiteID}, "")
	threadID := testutil.TaskID("repair-thread-impact")
	threadPath := filepath.Join(root, domain.ThreadsDir, threadID+"-repair-thread-impact.md")
	testutil.Write(t, threadPath, fmt.Sprintf("---\nschema: 1\nid: %s\nstatus: unstarted\ndescription: Observe graph repair\ngoal: Report derived Thread changes\ncreated: \"2026-09-05\"\ntasks: [%s]\n---\n# Repair impact\n", threadID, ownerID))

	receipt, err := core.MustNewService(testutil.Must(NewFS(root, core.UnrestrictedMutations()))).RepairTaskGraph(core.TaskGraphRepairRequest{Auto: true}, false)
	if err != nil {
		t.Fatal(err)
	}
	if receipt.InitialHealth != core.GraphBroken || receipt.FinalHealth != core.GraphHealthy || len(receipt.ThreadImpacts) != 1 {
		t.Fatalf("repair receipt = %+v", receipt)
	}
	impact := receipt.ThreadImpacts[0]
	if impact.ThreadID != threadID || !impact.Direct || !slices.Contains(impact.ChangedTaskIDs, ownerID) ||
		impact.Before.GraphHealth != core.GraphBroken || impact.After.GraphHealth != core.GraphHealthy {
		t.Fatalf("Thread impact = %+v", impact)
	}
}

func TestMutateTaskGraphRepairRejectsLateUnreadableTaskByteChange(t *testing.T) {
	root := t.TempDir()
	ownerID := testutil.TaskID("repair-unreadable-owner")
	ownerPath := writeGraphMutationTask(t, root, "repair-unreadable-owner", domain.StatusReadyToStart, []string{"invalid-token"}, "")
	unreadableID := testutil.TaskID("repair-unreadable-peer")
	unreadablePath := filepath.Join(root, domain.TasksDir, unreadableID+"-repair-unreadable-peer.md")
	first := "---\nid: [unterminated\n---\n# First bytes\n"
	testutil.Write(t, unreadablePath, first)
	request := core.TaskGraphRepairRequest{Edits: []core.TaskGraphSourceEdit{{
		Action: core.TaskGraphSourceDropDeclaration, Source: core.TaskGraphSourceRef{TaskID: ownerID},
		Field: core.TaskDependencyDependsOn, Value: "invalid-token",
	}}}
	original := testHookBeforeGraphRepairWrite
	t.Cleanup(func() { testHookBeforeGraphRepairWrite = original })
	testHookBeforeGraphRepairWrite = func(core.TaskGraphSourceRef) {
		testHookBeforeGraphRepairWrite = nil
		testutil.Write(t, unreadablePath, "---\nid: [unterminated\n---\n# Different bytes\n")
	}
	before, _ := os.ReadFile(ownerPath)
	receipt, err := core.MustNewService(testutil.Must(NewFS(root, core.UnrestrictedMutations())), core.WithRetry(0, func(int) {})).RepairTaskGraph(request, false)
	if !errors.Is(err, domain.ErrConflict) || receipt.Committed || len(receipt.RemainingFiles) != 1 {
		t.Fatalf("unreadable task CAS receipt=%+v err=%v", receipt, err)
	}
	after, _ := os.ReadFile(ownerPath)
	if !slices.Equal(before, after) {
		t.Fatal("unreadable task evidence conflict still changed target")
	}
}

func TestMutateTaskGraphRepairRejectsLateReadableNonTargetTaskByteChange(t *testing.T) {
	root := t.TempDir()
	ownerID := testutil.TaskID("repair-readable-owner")
	ownerPath := writeGraphMutationTask(t, root, "repair-readable-owner", domain.StatusReadyToStart, []string{"invalid-token"}, "")
	peerPath := writeGraphMutationTask(t, root, "repair-readable-peer", domain.StatusNextUp, nil, "")
	request := core.TaskGraphRepairRequest{Edits: []core.TaskGraphSourceEdit{{
		Action: core.TaskGraphSourceDropDeclaration, Source: core.TaskGraphSourceRef{TaskID: ownerID},
		Field: core.TaskDependencyDependsOn, Value: "invalid-token",
	}}}
	original := testHookBeforeGraphRepairWrite
	t.Cleanup(func() { testHookBeforeGraphRepairWrite = original })
	testHookBeforeGraphRepairWrite = func(core.TaskGraphSourceRef) {
		testHookBeforeGraphRepairWrite = nil
		content, err := os.ReadFile(peerPath)
		if err != nil {
			t.Fatal(err)
		}
		testutil.Write(t, peerPath, string(content)+"\n<!-- concurrent readable edit -->\n")
	}
	before, _ := os.ReadFile(ownerPath)
	receipt, err := core.MustNewService(testutil.Must(NewFS(root, core.UnrestrictedMutations())), core.WithRetry(0, func(int) {})).RepairTaskGraph(request, false)
	if !errors.Is(err, domain.ErrConflict) || receipt.Committed || len(receipt.AppliedFiles) != 0 {
		t.Fatalf("readable task CAS receipt=%+v err=%v", receipt, err)
	}
	after, _ := os.ReadFile(ownerPath)
	if !slices.Equal(before, after) {
		t.Fatal("readable task evidence conflict still changed target")
	}
}

func TestMutateTaskGraphRepairReportsAndConvergesDurablePrefix(t *testing.T) {
	root := t.TempDir()
	firstID := testutil.TaskID("repair-prefix-first")
	secondID := testutil.TaskID("repair-prefix-second")
	writeGraphMutationTask(t, root, "repair-prefix-first", domain.StatusReadyToStart, []string{"first-invalid"}, "")
	writeGraphMutationTask(t, root, "repair-prefix-second", domain.StatusReadyToStart, []string{"second-invalid"}, "")
	request := core.TaskGraphRepairRequest{Edits: []core.TaskGraphSourceEdit{
		{Action: core.TaskGraphSourceDropDeclaration, Source: core.TaskGraphSourceRef{TaskID: firstID}, Field: core.TaskDependencyDependsOn, Value: "first-invalid"},
		{Action: core.TaskGraphSourceDropDeclaration, Source: core.TaskGraphSourceRef{TaskID: secondID}, Field: core.TaskDependencyDependsOn, Value: "second-invalid"},
	}}
	original := testHookAfterGraphRepairWrite
	t.Cleanup(func() { testHookAfterGraphRepairWrite = original })
	testHookAfterGraphRepairWrite = func(_ core.TaskGraphSourceRef, index int) error {
		if index == 0 {
			testHookAfterGraphRepairWrite = nil
			return fmt.Errorf("injected stop")
		}
		return nil
	}

	service := core.MustNewService(testutil.Must(NewFS(root, core.UnrestrictedMutations())), core.WithRetry(0, func(int) {}))
	partial, err := service.RepairTaskGraph(request, false)
	if err == nil || !partial.Committed || len(partial.AppliedFiles) != 1 || len(partial.RemainingFiles) != 1 || partial.FinalHealth != core.GraphBroken {
		t.Fatalf("partial receipt=%+v err=%v", partial, err)
	}
	completed, err := service.RepairTaskGraph(request, false)
	if err != nil {
		t.Fatal(err)
	}
	if !completed.Committed || len(completed.AppliedFiles) != 1 || completed.FinalHealth != core.GraphHealthy {
		t.Fatalf("retry receipt = %+v", completed)
	}
	noop, err := service.RepairTaskGraph(request, false)
	if err != nil || noop.Changed || noop.Committed || len(noop.AppliedFiles) != 0 ||
		len(noop.Selected) != len(request.Edits) || len(noop.Operations) != 0 {
		t.Fatalf("converged retry receipt=%+v err=%v", noop, err)
	}
}

func TestUpdateDependencySourceEditsPreservesUnselectedSequenceNodes(t *testing.T) {
	content := []byte("---\nid: 6g0000000000\nstatus: next-up\ndepends_on:\n  - 6g1111111111 # keep first\n  - 6g1111111111 # remove duplicate\n  - invalid-token # preserve intent\ncustom: yes # untouched\n---\n# Body\n")
	edited, changed, err := updateDependencySourceEdits(content, []core.TaskGraphSourceEdit{{
		Action: core.TaskGraphSourceDedupe, Field: core.TaskDependencyDependsOn, Value: "6g1111111111",
	}}, "2026-09-05")
	if err != nil || !changed {
		t.Fatalf("edit changed=%v err=%v", changed, err)
	}
	text := string(edited)
	for _, want := range []string{"6g1111111111 # keep first", "invalid-token # preserve intent", "custom: yes # untouched", "# Body"} {
		if !strings.Contains(text, want) {
			t.Fatalf("missing %q:\n%s", want, text)
		}
	}
	if strings.Contains(text, "remove duplicate") {
		t.Fatalf("duplicate node survived:\n%s", text)
	}
}

func TestUpdateDependencySourceEditsMatchesAliasSemanticValue(t *testing.T) {
	content := []byte("---\nid: 6g0000000000\nstatus: next-up\nrepair_ref: &repair_ref 6g1111111111\ndepends_on: [*repair_ref, 6g1111111111]\n---\n# Body\n")
	edited, changed, err := updateDependencySourceEdits(content, []core.TaskGraphSourceEdit{{
		Action: core.TaskGraphSourceDedupe, Field: core.TaskDependencyDependsOn, Value: "6g1111111111",
	}}, "2026-09-05")
	if err != nil || !changed {
		t.Fatalf("alias edit changed=%v err=%v", changed, err)
	}
	text := string(edited)
	if !strings.Contains(text, "repair_ref: &repair_ref 6g1111111111") || !strings.Contains(text, "depends_on: [*repair_ref]") {
		t.Fatalf("alias representation was not preserved surgically:\n%s", text)
	}
}

func TestMutateTaskGraphRepairAutoDeduplicatesAliasDeclaration(t *testing.T) {
	root := t.TempDir()
	prerequisiteID := testutil.TaskID("repair-alias-prerequisite")
	ownerID := testutil.TaskID("repair-alias-owner")
	writeGraphMutationTask(t, root, "repair-alias-prerequisite", domain.StatusCompleted, nil, "")
	ownerPath := filepath.Join(root, domain.TasksDir, ownerID+"-repair-alias-owner.md")
	testutil.Write(t, ownerPath, fmt.Sprintf("---\nschema: 1\nid: %s\nstatus: next-up\ndescription: alias owner\nrepair_ref: &repair_ref %s\ndepends_on: [*repair_ref, %s]\n---\n# Alias owner\n", ownerID, prerequisiteID, prerequisiteID))

	receipt, err := core.MustNewService(testutil.Must(NewFS(root, core.UnrestrictedMutations()))).RepairTaskGraph(core.TaskGraphRepairRequest{Auto: true}, false)
	if err != nil || !receipt.Committed || receipt.FinalHealth != core.GraphHealthy {
		t.Fatalf("alias repair receipt=%+v err=%v", receipt, err)
	}
	content, err := os.ReadFile(ownerPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(content), "depends_on: [*repair_ref]") || strings.Contains(string(content), ", "+prerequisiteID+"]") {
		t.Fatalf("alias repair did not preserve the surviving alias:\n%s", content)
	}
	graph, err := core.LoadTaskGraph(testutil.Must(NewFS(root, core.UnrestrictedMutations())))
	if err != nil {
		t.Fatal(err)
	}
	task, ok := graph.Task(ownerID)
	if !ok || !slices.Equal(task.DependsOn, []string{prerequisiteID}) {
		t.Fatalf("alias repair semantic dependencies = %+v", task.DependsOn)
	}
}

func TestMutateTaskGraphRepairValidatesOneSourceGroupPerDurableWrite(t *testing.T) {
	root := t.TempDir()
	request := core.TaskGraphRepairRequest{}
	for _, seed := range []string{"repair-work-first", "repair-work-second", "repair-work-third"} {
		taskID := testutil.TaskID(seed)
		writeGraphMutationTask(t, root, seed, domain.StatusNextUp, []string{seed + "-invalid"}, "")
		request.Edits = append(request.Edits, core.TaskGraphSourceEdit{
			Action: core.TaskGraphSourceDropDeclaration, Source: core.TaskGraphSourceRef{TaskID: taskID},
			Field: core.TaskDependencyDependsOn, Value: seed + "-invalid",
		})
	}
	original := testHookGraphRepairStepValidation
	t.Cleanup(func() { testHookGraphRepairStepValidation = original })
	validated := 0
	testHookGraphRepairStepValidation = func(plan core.TaskGraphRepairPlan) {
		validated++
		if len(plan.Operations) != 1 || len(plan.Selections) != 1 || plan.Operations[0].Edit.Source != plan.Selections[0].Source {
			t.Fatalf("step plan was not source-bounded: %+v", plan)
		}
	}
	receipt, err := core.MustNewService(testutil.Must(NewFS(root, core.UnrestrictedMutations()))).RepairTaskGraph(request, false)
	if err != nil || !receipt.Committed || validated != len(request.Edits) {
		t.Fatalf("bounded validation count=%d receipt=%+v err=%v", validated, receipt, err)
	}
}

func BenchmarkMutateTaskGraphRepairManySources(b *testing.B) {
	for _, count := range []int{25, 50, 100, 200} {
		b.Run(fmt.Sprintf("tasks-%d", count), func(b *testing.B) {
			b.ReportMetric(float64(count), "files/op")
			for iteration := 0; iteration < b.N; iteration++ {
				b.StopTimer()
				root := b.TempDir()
				request := core.TaskGraphRepairRequest{}
				for index := 0; index < count; index++ {
					seed := fmt.Sprintf("repair-benchmark-%04d", index)
					taskID := testutil.TaskID(seed)
					writeGraphMutationTask(b, root, seed, domain.StatusNextUp, []string{seed + "-invalid"}, "")
					request.Edits = append(request.Edits, core.TaskGraphSourceEdit{
						Action: core.TaskGraphSourceDropDeclaration, Source: core.TaskGraphSourceRef{TaskID: taskID},
						Field: core.TaskDependencyDependsOn, Value: seed + "-invalid",
					})
				}
				service := core.MustNewService(testutil.Must(NewFS(root, core.UnrestrictedMutations())))
				b.StartTimer()
				if _, err := service.RepairTaskGraph(request, false); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func hasStoreRepairReason(defects []core.TaskGraphRepairDefect, reason core.TaskGraphRepairReason) bool {
	for _, defect := range defects {
		if defect.Reason == reason {
			return true
		}
	}
	return false
}
