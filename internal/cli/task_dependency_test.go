package cli

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/testutil"
	"github.com/andy-esch/taskflow/internal/wire"
)

type dependencyCLITask struct {
	slug       string
	status     domain.Status
	dependsOn  []string
	legacyYAML string
}

func dependencyCLIRepo(t *testing.T, tasks ...dependencyCLITask) string {
	t.Helper()
	repo := testutil.NewRepo(t)
	for _, task := range tasks {
		dependsOn := ""
		if len(task.dependsOn) > 0 {
			dependsOn = "depends_on: [" + strings.Join(task.dependsOn, ", ") + "]\n"
		}
		content := "---\n" +
			"id: " + testutil.TaskID(task.slug) + "\n" +
			"status: " + string(task.status) + "\n" +
			"description: " + task.slug + "\n" +
			"tags: [graph]\n" + dependsOn + task.legacyYAML +
			"---\n# " + task.slug + "\n\nBody stays intact.\n"
		repo.Task(string(task.status), task.slug+".md", content)
	}
	return repo.Root
}

func TestTaskDependAddRemoveJSONDryRunAndNoop(t *testing.T) {
	alphaID := testutil.TaskID("alpha-prerequisite")
	dependentID := testutil.TaskID("dependent")
	root := dependencyCLIRepo(t,
		dependencyCLITask{slug: "alpha-prerequisite", status: domain.StatusCompleted},
		dependencyCLITask{slug: "dependent", status: domain.StatusReadyToStart},
	)
	dependentPath := filepath.Join(root, domain.TasksDir, dependentID+"-dependent.md")

	out, errOut, err := runIn(t, root, "task", "depend", "add", "DEPEN", "--on", "ALPHA", "--dry-run", "--json")
	if err != nil || errOut != "" {
		t.Fatalf("dry add: %v\nstdout=%s\nstderr=%s", err, out, errOut)
	}
	var dry wire.DependencyMutationEnvelope
	if err := json.Unmarshal([]byte(out), &dry); err != nil {
		t.Fatalf("decode dry receipt: %v\n%s", err, out)
	}
	if !dry.DryRun || !dry.Changed || len(dry.PlannedTaskIDs) != 1 || len(dry.AppliedTaskIDs) != 0 || dry.Workspace.PlanningRoot == "" {
		t.Fatalf("dry receipt = %+v", dry)
	}
	content, _ := os.ReadFile(dependentPath)
	if strings.Contains(string(content), "depends_on:") {
		t.Fatalf("dry run wrote dependency:\n%s", content)
	}

	out, errOut, err = runIn(t, root, "task", "depend", "add", "dependent", "--on", "alpha-prerequisite", "--json")
	if err != nil || errOut != "" {
		t.Fatalf("add: %v\nstdout=%s\nstderr=%s", err, out, errOut)
	}
	var added wire.DependencyMutationEnvelope
	if err := json.Unmarshal([]byte(out), &added); err != nil {
		t.Fatal(err)
	}
	if !added.Changed || added.Edges[0].Outcome != "added" ||
		added.Edges[0].DependentID != dependentID || added.Edges[0].PrerequisiteID != alphaID ||
		!slices.Equal(added.AppliedTaskIDs, []string{dependentID}) {
		t.Fatalf("add receipt = %+v", added)
	}
	afterAdd, _ := os.ReadFile(dependentPath)

	out, _, err = runIn(t, root, "task", "depend", "add", dependentID, "--on", alphaID, "--json")
	if err != nil {
		t.Fatal(err)
	}
	var noop wire.DependencyMutationEnvelope
	if err := json.Unmarshal([]byte(out), &noop); err != nil {
		t.Fatal(err)
	}
	if noop.Changed || noop.Edges[0].Outcome != "skipped" || len(noop.AppliedTaskIDs) != 0 {
		t.Fatalf("idempotent receipt = %+v", noop)
	}
	afterNoopAdd, _ := os.ReadFile(dependentPath)
	if !bytes.Equal(afterAdd, afterNoopAdd) {
		t.Fatalf("idempotent add changed task bytes:\n--- before ---\n%s\n--- after ---\n%s", afterAdd, afterNoopAdd)
	}

	out, _, err = runIn(t, root, "task", "depend", "remove", "dependent", "--on", "alpha", "--json")
	if err != nil {
		t.Fatal(err)
	}
	var removed wire.DependencyMutationEnvelope
	if err := json.Unmarshal([]byte(out), &removed); err != nil {
		t.Fatal(err)
	}
	if !removed.Changed || removed.Edges[0].Outcome != "removed" {
		t.Fatalf("remove receipt = %+v", removed)
	}
	content, _ = os.ReadFile(dependentPath)
	if strings.Contains(string(content), "depends_on:") {
		t.Fatalf("remove did not clear empty dependency field:\n%s", content)
	}
	afterRemove := append([]byte(nil), content...)
	out, _, err = runIn(t, root, "task", "depend", "remove", "dependent", "--on", "alpha", "--json")
	if err != nil {
		t.Fatal(err)
	}
	var noopRemove wire.DependencyMutationEnvelope
	if err := json.Unmarshal([]byte(out), &noopRemove); err != nil || noopRemove.Changed || noopRemove.Edges[0].Outcome != "skipped" {
		t.Fatalf("idempotent remove=%+v decode=%v", noopRemove, err)
	}
	afterNoopRemove, _ := os.ReadFile(dependentPath)
	if !bytes.Equal(afterRemove, afterNoopRemove) {
		t.Fatalf("idempotent remove changed task bytes:\n--- before ---\n%s\n--- after ---\n%s", afterRemove, afterNoopRemove)
	}
}

func TestTaskDependAddCallsOutNewlyBlockedDependent(t *testing.T) {
	prerequisiteID := testutil.TaskID("unfinished-prerequisite")
	dependentID := testutil.TaskID("newly-blocked")
	root := dependencyCLIRepo(t,
		dependencyCLITask{slug: "unfinished-prerequisite", status: domain.StatusNextUp},
		dependencyCLITask{slug: "newly-blocked", status: domain.StatusReadyToStart},
	)
	out, errOut, err := runIn(t, root, "task", "depend", "add", "newly-blocked", "--on", "unfinished-prerequisite")
	if err != nil || errOut != "" {
		t.Fatalf("dependency add: %v\nstdout=%s\nstderr=%s", err, out, errOut)
	}
	for _, want := range []string{dependentID, "candidate/clear -> candidate/blocked", "task blockers " + dependentID} {
		if !strings.Contains(out, want) {
			t.Errorf("human consequence missing %q:\n%s", want, out)
		}
	}

	root = dependencyCLIRepo(t,
		dependencyCLITask{slug: "unfinished-prerequisite", status: domain.StatusNextUp},
		dependencyCLITask{slug: "newly-blocked", status: domain.StatusReadyToStart},
	)
	out, _, err = runIn(t, root, "task", "depend", "add", "newly-blocked", "--on", prerequisiteID, "--json")
	if err != nil {
		t.Fatal(err)
	}
	var receipt wire.DependencyMutationEnvelope
	if err := json.Unmarshal([]byte(out), &receipt); err != nil || len(receipt.Impacts) != 1 ||
		receipt.Impacts[0].TaskID != dependentID || receipt.Impacts[0].After.Gate != "blocked" {
		t.Fatalf("machine consequence=%+v decode=%v", receipt, err)
	}
}

func TestTaskDependAddRejectsCycleWithValidationExit(t *testing.T) {
	alphaID := testutil.TaskID("cycle-alpha")
	betaID := testutil.TaskID("cycle-beta")
	root := dependencyCLIRepo(t,
		dependencyCLITask{slug: "cycle-alpha", status: domain.StatusReadyToStart},
		dependencyCLITask{slug: "cycle-beta", status: domain.StatusReadyToStart, dependsOn: []string{alphaID}},
	)
	out, errOut, err := runIn(t, root, "task", "depend", "add", alphaID, "--on", betaID, "--json")
	if err == nil || ExitCode(err) != 11 || out != "" {
		t.Fatalf("cycle stdout=%q stderr=%q err=%v exit=%d", out, errOut, err, ExitCode(err))
	}
	var errorOut bytes.Buffer
	WriteError(&errorOut, err, true)
	var envelope wire.ErrorEnvelope
	if decodeErr := json.Unmarshal(errorOut.Bytes(), &envelope); decodeErr != nil {
		t.Fatalf("decode error: %v\n%s", decodeErr, errorOut.String())
	}
	if envelope.Error.Code != "validation" || !strings.Contains(envelope.Error.Message, "cycle") {
		t.Fatalf("cycle error envelope = %+v", envelope)
	}
	if envelope.Error.DependencyMutation != nil {
		t.Fatalf("pre-write cycle failure must not imply an applied mutation: %+v", envelope.Error.DependencyMutation)
	}
}

func TestTaskDependHumanReceiptMatchesMutationSemantics(t *testing.T) {
	alphaID := testutil.TaskID("human-alpha")
	dependentID := testutil.TaskID("human-dependent")
	root := dependencyCLIRepo(t,
		dependencyCLITask{slug: "human-alpha", status: domain.StatusCompleted},
		dependencyCLITask{slug: "human-dependent", status: domain.StatusReadyToStart},
	)
	preview, _, err := runIn(t, root, "task", "depend", "add", "human-dependent", "--on", "human-alpha", "--dry-run")
	if err != nil || !strings.Contains(preview, "would be added") ||
		!strings.Contains(preview, alphaID+" -> "+dependentID) || !strings.Contains(preview, "planned task files") {
		t.Fatalf("human preview: %v\n%s", err, preview)
	}
	applied, _, err := runIn(t, root, "task", "depend", "add", "human-dependent", "--on", "human-alpha")
	if err != nil || !strings.Contains(applied, "added") || !strings.Contains(applied, "applied task files: "+dependentID) {
		t.Fatalf("human apply: %v\n%s", err, applied)
	}
	noop, _, err := runIn(t, root, "task", "depend", "add", "human-dependent", "--on", "human-alpha")
	if err != nil || !strings.Contains(noop, "already satisfies") || !strings.Contains(noop, "already satisfied") {
		t.Fatalf("human no-op: %v\n%s", err, noop)
	}
}

func TestTaskDependMigratePreservesContentAndIsIdempotent(t *testing.T) {
	prerequisiteID := testutil.TaskID("legacy-prerequisite")
	dependentID := testutil.TaskID("legacy-dependent")
	root := dependencyCLIRepo(t,
		dependencyCLITask{slug: "legacy-prerequisite", status: domain.StatusCompleted, legacyYAML: "blocks: [legacy-dependent]\ncustom_key: keep-me # preserve comment\n"},
		dependencyCLITask{slug: "legacy-dependent", status: domain.StatusReadyToStart, legacyYAML: "blocked_by: [legacy-prerequisite]\n"},
	)

	out, _, err := runIn(t, root, "task", "depend", "migrate", "--dry-run", "--json")
	if err != nil {
		t.Fatal(err)
	}
	var dry wire.DependencyMutationEnvelope
	if err := json.Unmarshal([]byte(out), &dry); err != nil || !dry.DryRun || len(dry.ClearedLegacyFields) != 2 {
		t.Fatalf("dry migration=%+v decode=%v", dry, err)
	}
	out, _, err = runIn(t, root, "task", "depend", "migrate", "--json")
	if err != nil {
		t.Fatal(err)
	}
	var migrated wire.DependencyMutationEnvelope
	if err := json.Unmarshal([]byte(out), &migrated); err != nil || len(migrated.AppliedTaskIDs) != 2 {
		t.Fatalf("migration=%+v decode=%v", migrated, err)
	}
	for _, slug := range []string{"legacy-prerequisite", "legacy-dependent"} {
		path := filepath.Join(root, domain.TasksDir, testutil.TaskID(slug)+"-"+slug+".md")
		content, readErr := os.ReadFile(path)
		if readErr != nil {
			t.Fatal(readErr)
		}
		if strings.Contains(string(content), "blocked_by:") || strings.Contains(string(content), "blocks:") ||
			!strings.Contains(string(content), "Body stays intact.") {
			t.Fatalf("migration content for %s:\n%s", slug, content)
		}
	}
	prerequisiteContent, _ := os.ReadFile(filepath.Join(root, domain.TasksDir, prerequisiteID+"-legacy-prerequisite.md"))
	if !strings.Contains(string(prerequisiteContent), "custom_key: keep-me # preserve comment") {
		t.Fatalf("custom frontmatter was not preserved:\n%s", prerequisiteContent)
	}
	dependentContent, _ := os.ReadFile(filepath.Join(root, domain.TasksDir, dependentID+"-legacy-dependent.md"))
	if !strings.Contains(string(dependentContent), "depends_on: ["+prerequisiteID+"]") {
		t.Fatalf("canonical edge missing:\n%s", dependentContent)
	}

	out, _, err = runIn(t, root, "task", "depend", "migrate", "--json")
	if err != nil {
		t.Fatal(err)
	}
	var noop wire.DependencyMutationEnvelope
	if err := json.Unmarshal([]byte(out), &noop); err != nil || noop.Changed || len(noop.AppliedTaskIDs) != 0 {
		t.Fatalf("idempotent migration=%+v decode=%v", noop, err)
	}
}

func TestTaskGraphQueryCommandsAndUnblockedSelector(t *testing.T) {
	rootID := testutil.TaskID("query-root")
	middleID := testutil.TaskID("query-middle")
	targetID := testutil.TaskID("query-target")
	root := dependencyCLIRepo(t,
		dependencyCLITask{slug: "query-root", status: domain.StatusReadyToStart},
		dependencyCLITask{slug: "query-middle", status: domain.StatusCompleted, dependsOn: []string{rootID}},
		dependencyCLITask{slug: "query-target", status: domain.StatusReadyToStart, dependsOn: []string{middleID}},
	)

	out, _, err := runIn(t, root, "task", "blockers", "query-target", "--json")
	if err != nil {
		t.Fatal(err)
	}
	var frontier wire.TaskBlockersEnvelope
	if err := json.Unmarshal([]byte(out), &frontier); err != nil || frontier.Projection != "frontier" ||
		len(frontier.Blockers) != 1 || frontier.Blockers[0].Task.TaskID != rootID || frontier.Blockers[0].Reason != "not-started" ||
		frontier.State.Gate != "blocked" || frontier.State.Eligible {
		t.Fatalf("frontier=%+v decode=%v", frontier, err)
	}
	out, _, err = runIn(t, root, "task", "blockers", targetID, "--causal", "--json")
	if err != nil {
		t.Fatal(err)
	}
	var causal wire.TaskBlockersEnvelope
	if err := json.Unmarshal([]byte(out), &causal); err != nil || causal.Projection != "causal" || len(causal.Blockers) != 2 {
		t.Fatalf("causal=%+v decode=%v", causal, err)
	}
	out, _, err = runIn(t, root, "task", "unblocks", "query-root", "--json")
	if err != nil {
		t.Fatal(err)
	}
	var unblocks wire.TaskUnblocksEnvelope
	if err := json.Unmarshal([]byte(out), &unblocks); err != nil || len(unblocks.Unblocks) != 2 || unblocks.State.Role != "candidate" {
		t.Fatalf("unblocks=%+v decode=%v", unblocks, err)
	}
	for _, dependent := range unblocks.Unblocks {
		if dependent.Task.TaskID == targetID && !slices.Equal(dependent.Path, []string{rootID, middleID, targetID}) {
			t.Fatalf("target path = %v", dependent.Path)
		}
	}

	out, _, err = runIn(t, root, "task", "list", "--unblocked", "--json")
	if err != nil {
		t.Fatal(err)
	}
	var tasks wire.TasksEnvelope
	if err := json.Unmarshal([]byte(out), &tasks); err != nil || len(tasks.Tasks) != 1 || tasks.Tasks[0].ID != rootID {
		t.Fatalf("unblocked list=%+v decode=%v", tasks, err)
	}

	human, _, err := runIn(t, root, "task", "blockers", "query-target")
	if err != nil || !strings.Contains(human, "query-root") || !strings.Contains(human, "not-started") || !strings.Contains(human, rootID) {
		t.Fatalf("human blocker parity: %v\n%s", err, human)
	}
	human, _, err = runIn(t, root, "task", "unblocks", "query-root")
	if err != nil || !strings.Contains(human, "query-target") || !strings.Contains(human, "transitive") ||
		!strings.Contains(human, rootID+" -> "+middleID+" -> "+targetID) {
		t.Fatalf("human downstream parity: %v\n%s", err, human)
	}
}

func TestTaskGraphQueriesProjectResolvedLegacyConstraints(t *testing.T) {
	prerequisiteID := testutil.TaskID("legacy-query-prerequisite")
	dependentID := testutil.TaskID("legacy-query-dependent")
	root := dependencyCLIRepo(t,
		dependencyCLITask{slug: "legacy-query-prerequisite", status: domain.StatusReadyToStart},
		dependencyCLITask{slug: "legacy-query-dependent", status: domain.StatusReadyToStart, legacyYAML: "blocked_by: [legacy-query-prerequisite]\n"},
	)
	out, _, err := runIn(t, root, "task", "blockers", "legacy-query-dependent", "--json")
	if err != nil {
		t.Fatal(err)
	}
	var blockers wire.TaskBlockersEnvelope
	if err := json.Unmarshal([]byte(out), &blockers); err != nil || blockers.Health != "degraded" ||
		blockers.State.Gate != "blocked" || len(blockers.Blockers) != 1 || blockers.Blockers[0].Task.TaskID != prerequisiteID {
		t.Fatalf("legacy blockers=%+v decode=%v", blockers, err)
	}
	out, _, err = runIn(t, root, "task", "unblocks", "legacy-query-prerequisite", "--json")
	if err != nil {
		t.Fatal(err)
	}
	var unblocks wire.TaskUnblocksEnvelope
	if err := json.Unmarshal([]byte(out), &unblocks); err != nil || unblocks.Health != "degraded" ||
		len(unblocks.Unblocks) != 1 || unblocks.Unblocks[0].Task.TaskID != dependentID {
		t.Fatalf("legacy unblocks=%+v decode=%v", unblocks, err)
	}
}

func TestTaskListUnblockedFailsClosedOnBrokenGraph(t *testing.T) {
	root := dependencyCLIRepo(t, dependencyCLITask{
		slug: "broken", status: domain.StatusReadyToStart, dependsOn: []string{"not-a-stable-id"},
	})
	out, _, err := runIn(t, root, "task", "list", "--unblocked", "--json")
	if err == nil || ExitCode(err) != 11 || out != "" || !strings.Contains(err.Error(), "requires a healthy") {
		t.Fatalf("broken selector stdout=%q err=%v exit=%d", out, err, ExitCode(err))
	}
}

func TestWriteErrorCarriesStructuredDependencyMutationRecovery(t *testing.T) {
	receipt := core.DependencyMutationReceipt{
		Operation: core.DependencyMigrate, Changed: true,
		PlannedTaskIDs: []string{"6g0000000001", "6g0000000002"},
		AppliedTaskIDs: []string{"6g0000000001"}, RemainingTaskIDs: []string{"6g0000000002"},
	}
	err := &dependencyCommandFailure{
		cause:   &core.DependencyMutationFailure{Cause: domain.ErrConflict, Receipt: receipt},
		receipt: receipt, workspace: wire.WorkspaceJSON{PlanningRoot: "/repo/planning", Source: wire.WorkspaceSourceConfig},
	}
	var out bytes.Buffer
	WriteError(&out, err, true)
	var envelope wire.ErrorEnvelope
	if decodeErr := json.Unmarshal(out.Bytes(), &envelope); decodeErr != nil {
		t.Fatalf("decode error envelope: %v\n%s", decodeErr, out.String())
	}
	if envelope.Error.Code != "conflict" || envelope.Error.DependencyMutation == nil ||
		!slices.Equal(envelope.Error.DependencyMutation.AppliedTaskIDs, []string{"6g0000000001"}) ||
		!slices.Equal(envelope.Error.DependencyMutation.RemainingTaskIDs, []string{"6g0000000002"}) ||
		envelope.Error.DependencyMutation.Workspace.PlanningRoot != "/repo/planning" {
		t.Fatalf("structured recovery envelope = %+v", envelope)
	}
	if !errors.Is(err, domain.ErrConflict) {
		t.Fatal("dependency command wrapper did not preserve error classification")
	}
}
