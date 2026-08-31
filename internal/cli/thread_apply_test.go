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

	yaml "go.yaml.in/yaml/v3"

	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/store"
	"github.com/andy-esch/taskflow/internal/testutil"
	"github.com/andy-esch/taskflow/internal/wire"
)

func writeThreadApplyCLITask(t *testing.T, root, seed string, status domain.Status) string {
	t.Helper()
	taskID := testutil.TaskID(seed)
	mustWrite(t, filepath.Join(root, domain.TasksDir, taskID+"-"+seed+".md"), "---\n"+
		"schema: 1\nid: "+taskID+"\nstatus: "+string(status)+"\nepic: 30\n"+
		"description: "+seed+"\neffort: 1h\ntier: 1\npriority: high\n"+
		"autonomy_level: 2\ntags: [threads]\ncreated: \"2026-08-30\"\n---\n# "+seed+"\n")
	return taskID
}

func TestThreadComposeAndApplyCLIConvergesExistingTasks(t *testing.T) {
	root := freshRepo(t)
	gateID := writeThreadApplyCLITask(t, root, "bulk-gate", domain.StatusCompleted)
	firstID := writeThreadApplyCLITask(t, root, "bulk-first", domain.StatusNextUp)
	secondID := writeThreadApplyCLITask(t, root, "bulk-second", domain.StatusReadyToStart)
	manifestPath := filepath.Join(root, "thread-manifest.yml")
	planPath := filepath.Join(root, "thread-plan.yml")
	mustWrite(t, manifestPath, "schema: 1\nthread:\n  title: Bulk delivery\n  description: Link existing work safely\n  goal: Create a resumable Thread\n  tags: [threads, bulk]\nnodes:\n"+
		"  - key: gate\n    task_id: "+gateID+"\n    member: false\n"+
		"  - key: first\n    task_id: "+firstID+"\n"+
		"  - key: second\n    task_id: "+secondID+"\n"+
		"dependencies:\n  - from: gate\n    to: first\n  - from: first\n    to: second\n")

	composed := runRoot(t, "-C", root, "thread", "compose", "--from", manifestPath, "--out", planPath, "--json")
	var compose wire.ThreadApplyComposeEnvelope
	if err := json.Unmarshal([]byte(composed), &compose); err != nil {
		t.Fatalf("compose JSON: %v\n%s", err, composed)
	}
	if compose.SchemaVersion != wire.SchemaVersion || !compose.Written || compose.DryRun || compose.PlanPath != planPath {
		t.Fatalf("compose = %+v", compose)
	}
	if info, err := os.Stat(planPath); err != nil || info.Mode().Perm() != 0o600 {
		t.Fatalf("durable plan mode=%v err=%v", info.Mode(), err)
	}
	content, err := os.ReadFile(planPath)
	if err != nil {
		t.Fatal(err)
	}
	var plan core.ThreadApplyPlan
	if err := yaml.Unmarshal(content, &plan); err != nil || plan.PlanningRepoID == "" || plan.Thread.Body == "" {
		t.Fatalf("plan=%+v err=%v", plan, err)
	}

	previewText := runRoot(t, "-C", root, "thread", "apply", planPath, "--dry-run", "--json")
	var preview wire.ThreadApplyEnvelope
	if err := json.Unmarshal([]byte(previewText), &preview); err != nil {
		t.Fatalf("preview JSON: %v\n%s", err, previewText)
	}
	if !preview.DryRun || !preview.Changed || preview.Complete || preview.Committed || len(preview.Operations) != 3 {
		t.Fatalf("preview = %+v", preview)
	}

	appliedText := runRoot(t, "-C", root, "thread", "apply", planPath, "--json")
	var applied wire.ThreadApplyEnvelope
	if err := json.Unmarshal([]byte(appliedText), &applied); err != nil {
		t.Fatalf("apply JSON: %v\n%s", err, appliedText)
	}
	if !applied.Changed || !applied.Complete || !applied.Committed {
		t.Fatalf("applied = %+v", applied)
	}
	for _, operation := range applied.Operations {
		if operation.State != "applied" {
			t.Fatalf("applied operation = %+v", operation)
		}
	}
	graph, err := core.LoadTaskGraph(store.NewFS(root))
	if err != nil || graph.Health() != core.GraphHealthy {
		t.Fatalf("graph health=%s err=%v", graph.Health(), err)
	}
	first, _ := graph.Task(firstID)
	second, _ := graph.Task(secondID)
	if !slices.Contains(first.DependsOn, gateID) || !slices.Contains(second.DependsOn, firstID) {
		t.Fatalf("first=%v second=%v", first.DependsOn, second.DependsOn)
	}
	thread, body, err := store.NewFS(root).GetThread(plan.Thread.ID)
	if err != nil || !slices.Equal(thread.Tasks, plan.Thread.Tasks) || body != plan.Thread.Body {
		t.Fatalf("thread=%+v body=%q err=%v", thread, body, err)
	}

	convergedText := runRoot(t, "-C", root, "thread", "apply", planPath, "--json")
	var converged wire.ThreadApplyEnvelope
	if err := json.Unmarshal([]byte(convergedText), &converged); err != nil {
		t.Fatal(err)
	}
	if converged.Changed || !converged.Complete || converged.Committed {
		t.Fatalf("converged = %+v", converged)
	}
}

func TestThreadComposeCLIIsStrictNoClobberAndDryRunSafe(t *testing.T) {
	root := freshRepo(t)
	taskID := writeThreadApplyCLITask(t, root, "strict-member", domain.StatusNextUp)
	manifestPath := filepath.Join(root, "strict.yml")
	planPath := filepath.Join(root, "strict-plan.yml")
	valid := "thread:\n  title: Strict Thread\n  description: Exercise strict parsing\n  goal: Reject ambiguous input\nnodes:\n  - key: member\n    task_id: " + taskID + "\n"
	mustWrite(t, manifestPath, valid)

	preview := runRoot(t, "-C", root, "thread", "compose", "--from", manifestPath, "--out", planPath, "--dry-run")
	if !strings.Contains(preview, "would compose") || !strings.Contains(preview, "planning_repo_id:") {
		t.Fatalf("dry-run output:\n%s", preview)
	}
	if _, err := os.Stat(planPath); !os.IsNotExist(err) {
		t.Fatalf("dry-run wrote plan: %v", err)
	}
	runRoot(t, "-C", root, "thread", "compose", "--from", manifestPath, "--out", planPath)
	if _, err := runRootRC(t, "-C", root, "thread", "compose", "--from", manifestPath, "--out", planPath); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("no-clobber error = %v", err)
	}
	if _, err := runRootRC(t, "-C", root, "thread", "compose", "--from", manifestPath, "--out", planPath, "--dry-run"); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("dry-run no-clobber error = %v", err)
	}

	mustWrite(t, manifestPath, strings.Replace(valid, "title: Strict Thread", "title: Strict Thread\n  surprise: nope", 1))
	if _, err := runRootRC(t, "-C", root, "thread", "compose", "--from", manifestPath, "--out", filepath.Join(root, "unknown.yml")); !errors.Is(err, domain.ErrValidation) || !strings.Contains(err.Error(), "field surprise") {
		t.Fatalf("unknown-field error = %v", err)
	}
	mustWrite(t, manifestPath, valid+"---\nthread: {}\n")
	if _, err := runRootRC(t, "-C", root, "thread", "compose", "--from", manifestPath, "--out", filepath.Join(root, "multi.yml")); !errors.Is(err, domain.ErrValidation) || !strings.Contains(err.Error(), "exactly one") {
		t.Fatalf("multi-document error = %v", err)
	}
	if _, err := runRootRC(t, "-C", root, "thread", "apply", "-"); !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("stdin apply error = %v", err)
	}
}

func TestThreadApplyFailureJSONRetainsDurablePrefix(t *testing.T) {
	receipt := core.ThreadApplyReceipt{
		Plan:    core.ThreadApplyPlan{Thread: core.ThreadApplyThread{ID: testutil.TaskID("error-thread"), Slug: "error-thread"}},
		Changed: true, Committed: true,
		Operations: []core.ThreadApplyOperation{
			{Kind: "dependency", Action: "add", State: core.ThreadApplyApplied, DependentID: testutil.TaskID("error-dependent"), PrerequisiteID: testutil.TaskID("error-prerequisite")},
			{Kind: "thread", Action: "create", State: core.ThreadApplyPending, ThreadID: testutil.TaskID("error-thread")},
		},
	}
	cause := &core.ThreadApplyFailure{Cause: domain.ErrConflict, Receipt: receipt}
	err := &threadApplyCommandFailure{
		cause: cause, receipt: receipt, planPath: "/tmp/thread.apply.yml",
		workspace: wire.WorkspaceJSON{PlanningRoot: "/repo/planning", Source: wire.WorkspaceSourceConfig},
	}
	var output bytes.Buffer
	WriteError(&output, err, true)
	var envelope wire.ErrorEnvelope
	if decodeErr := json.Unmarshal(output.Bytes(), &envelope); decodeErr != nil {
		t.Fatal(decodeErr)
	}
	if envelope.Error.Code != "conflict" || envelope.Error.ThreadApply == nil ||
		!envelope.Error.ThreadApply.Committed || envelope.Error.ThreadApply.Complete ||
		envelope.Error.ThreadApply.Operations[0].State != "applied" || envelope.Error.ThreadApply.Operations[1].State != "pending" {
		t.Fatalf("error envelope = %+v", envelope)
	}
}
