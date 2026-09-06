package cli

import (
	"bytes"
	"encoding/json"
	"errors"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/testutil"
	"github.com/andy-esch/taskflow/internal/wire"
)

func TestTaskDependRepairDiagnosesThenAppliesAutoAndExplicitIntent(t *testing.T) {
	ownerID := testutil.TaskID("repair-cli-owner")
	root := dependencyCLIRepo(t, dependencyCLITask{
		slug: "repair-cli-owner", status: domain.StatusReadyToStart,
		dependsOn:  []string{ownerID, ownerID, "invalid-human-token"},
		legacyYAML: "blocked_by: []\n",
	})
	out, errOut, err := runIn(t, root, "task", "depend", "repair")
	if err != nil || errOut != "" || !strings.Contains(out, "inferable:") ||
		!strings.Contains(out, "task depend repair --drop") || !strings.Contains(out, "invalid-human-token#0") {
		t.Fatalf("diagnosis err=%v\nstdout=%s\nstderr=%s", err, out, errOut)
	}

	out, errOut, err = runIn(t, root, "task", "depend", "repair", "--auto", "--json")
	if err != nil || errOut != "" {
		t.Fatalf("auto repair err=%v\nstdout=%s\nstderr=%s", err, out, errOut)
	}
	var automatic wire.TaskGraphRepairEnvelope
	if err := json.Unmarshal([]byte(out), &automatic); err != nil {
		t.Fatal(err)
	}
	if !automatic.Changed || !automatic.Committed || automatic.FinalHealth != "broken" ||
		len(automatic.Removed) != 2 || len(automatic.Residual) != 1 || automatic.Workspace.PlanningRoot == "" ||
		len(automatic.Operations) == 0 || !automatic.Operations[0].Automatic {
		t.Fatalf("auto receipt = %+v", automatic)
	}

	out, errOut, err = runIn(t, root, "task", "depend", "repair", "--drop", "repair-cli-owner:depends_on=invalid-human-token#0", "--json")
	if err != nil || errOut != "" {
		t.Fatalf("explicit repair err=%v\nstdout=%s\nstderr=%s", err, out, errOut)
	}
	var explicit wire.TaskGraphRepairEnvelope
	if err := json.Unmarshal([]byte(out), &explicit); err != nil {
		t.Fatal(err)
	}
	if !explicit.Committed || explicit.FinalHealth != "healthy" || explicit.Removed[0].Value != "invalid-human-token" {
		t.Fatalf("explicit receipt = %+v", explicit)
	}
}

func TestTaskDependRepairManifestAndSelectorParsing(t *testing.T) {
	root := dependencyCLIRepo(t, dependencyCLITask{
		slug: "repair-cli-manifest", status: domain.StatusReadyToStart,
		dependsOn: []string{"raw#12"},
	})
	manifest := filepath.Join(t.TempDir(), "repair.yaml")
	testutil.Write(t, manifest, "schema: 1\noperations:\n  - action: drop\n    task: repair-cli-manifest\n    field: depends_on\n    value: raw#12\n    occurrence: 0\n")
	out, errOut, err := runIn(t, root, "task", "depend", "repair", "--plan", manifest, "--dry-run", "--json")
	if err != nil || errOut != "" {
		t.Fatalf("manifest err=%v\nstdout=%s\nstderr=%s", err, out, errOut)
	}
	var receipt wire.TaskGraphRepairEnvelope
	if err := json.Unmarshal([]byte(out), &receipt); err != nil || !receipt.DryRun || receipt.Removed[0].Value != "raw#12" {
		t.Fatalf("manifest receipt=%+v decode=%v", receipt, err)
	}
	selector, err := parseGraphRepairSelector("repair-cli-manifest:depends_on=raw#12#0", core.TaskGraphSourceDropDeclaration)
	if err != nil || selector.Value != "raw#12" || selector.Occurrence != 0 {
		t.Fatalf("selector=%+v err=%v", selector, err)
	}
	selector, err = parseGraphRepairSelector("repair-cli-manifest:depends_on=not:a:stable:id#0", core.TaskGraphSourceDropDeclaration)
	if err != nil || selector.Value != "not:a:stable:id" || selector.Occurrence != 0 {
		t.Fatalf("colon-bearing selector=%+v err=%v", selector, err)
	}
	selector, err = parseGraphRepairSelector("tasks/a.md:depends_on=a:depends_on=b#0", core.TaskGraphSourceDropDeclaration)
	if err != nil || selector.Source.Location != "tasks/a.md" || selector.Value != "a:depends_on=b" || selector.Occurrence != 0 {
		t.Fatalf("embedded-field-marker selector=%+v err=%v", selector, err)
	}
}

func TestTaskDependRepairDocumentedRelativePathIsPlanningRootRelative(t *testing.T) {
	slug := "repair-cli-relative-path"
	taskID := testutil.TaskID(slug)
	root := dependencyCLIRepo(t, dependencyCLITask{
		slug: slug, status: domain.StatusReadyToStart, dependsOn: []string{"invalid-token"},
	})
	selector := filepath.Join("tasks", taskID+"-"+slug+".md") + ":depends_on=invalid-token#0"
	out, errOut, err := runIn(t, root, "task", "depend", "repair", "--drop", selector, "--dry-run", "--json")
	if err != nil || errOut != "" {
		t.Fatalf("relative selector err=%v\nstdout=%s\nstderr=%s", err, out, errOut)
	}
	var receipt wire.TaskGraphRepairEnvelope
	if err := json.Unmarshal([]byte(out), &receipt); err != nil || !receipt.DryRun || !receipt.Changed || len(receipt.Removed) != 1 {
		t.Fatalf("relative selector receipt=%+v decode=%v", receipt, err)
	}
}

func TestWriteErrorCarriesStructuredGraphRepairRecovery(t *testing.T) {
	receipt := core.TaskGraphRepairReceipt{
		Changed: true, Committed: true, InitialHealth: core.GraphBroken, FinalHealth: core.GraphBroken,
		PlannedFiles: []string{"tasks/a.md", "tasks/b.md"}, AppliedFiles: []string{"tasks/a.md"}, RemainingFiles: []string{"tasks/b.md"},
	}
	err := &graphRepairCommandFailure{
		cause: &core.TaskGraphRepairFailure{Cause: domain.ErrConflict, Receipt: receipt}, receipt: receipt,
		workspace: wire.WorkspaceJSON{PlanningRoot: "/repo/planning", Source: wire.WorkspaceSourceConfig},
	}
	var out bytes.Buffer
	WriteError(&out, err, true)
	var envelope wire.ErrorEnvelope
	if decodeErr := json.Unmarshal(out.Bytes(), &envelope); decodeErr != nil {
		t.Fatalf("decode recovery: %v\n%s", decodeErr, out.String())
	}
	if envelope.Error.Code != "conflict" || envelope.Error.GraphRepair == nil ||
		!slices.Equal(envelope.Error.GraphRepair.AppliedFiles, []string{"tasks/a.md"}) ||
		envelope.Error.GraphRepair.Workspace.PlanningRoot != "/repo/planning" || !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("recovery envelope = %+v", envelope)
	}
}
