package cli

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/andy-esch/taskflow/internal/cli/render"
	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/testutil"
	"github.com/andy-esch/taskflow/internal/wire"
)

func TestTaskStart_ChangesStatusInPlace(t *testing.T) {
	root := setupRepo(t)
	path := filepath.Join(root, "tasks", testutil.TaskID("alpha")+"-alpha.md")
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("alpha fixture missing: %v", err)
	}
	out := runRoot(t, "-C", root, "task", "start", "alpha")
	if !strings.Contains(out, "alpha -> in-progress") {
		t.Errorf("unexpected output: %q", out)
	}
	// Flat layout: the file path never changes; status is an in-place frontmatter edit.
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("alpha no longer at original path: %v", err)
	}
	if !strings.Contains(string(b), "status: in-progress") {
		t.Errorf("alpha frontmatter status not updated:\n%s", b)
	}
}

func TestTaskStartEligibilityForceAndMachineReceipt(t *testing.T) {
	prerequisiteID := testutil.TaskID("prerequisite")
	root := dependencyCLIRepo(t,
		dependencyCLITask{slug: "prerequisite", status: domain.StatusNextUp},
		dependencyCLITask{slug: "target", status: domain.StatusReadyToStart, dependsOn: []string{prerequisiteID}},
	)

	out, errOut, err := runIn(t, root, "task", "start", "target")
	if err == nil || ExitCode(err) != 11 || out != "" || !strings.Contains(errOut, "outstanding blockers") {
		t.Fatalf("default start stdout=%q stderr=%q err=%v", out, errOut, err)
	}
	out, errOut, err = runIn(t, root, "task", "start", "target", "--json")
	if err == nil || ExitCode(err) != 11 || errOut != "" {
		t.Fatalf("default JSON start stdout=%q stderr=%q err=%v", out, errOut, err)
	}
	var refused wire.MovesEnvelope
	if decodeErr := json.Unmarshal([]byte(out), &refused); decodeErr != nil {
		t.Fatalf("decode default refusal: %v\n%s", decodeErr, out)
	}
	if len(refused.Moves) != 1 || refused.Moves[0].LifecycleFailure == nil ||
		refused.Moves[0].LifecycleFailure.State.Gate != "blocked" ||
		!refused.Moves[0].LifecycleFailure.OverrideAllowed ||
		len(refused.Moves[0].LifecycleFailure.Blockers) != 1 ||
		refused.Moves[0].LifecycleFailure.Remedy == "" {
		t.Fatalf("structured refusal = %+v", refused.Moves)
	}
	out, errOut, err = runIn(t, root, "task", "start", "target", "--force", "--json")
	if err != nil || errOut != "" {
		t.Fatalf("forced start: %v\nstdout=%s\nstderr=%s", err, out, errOut)
	}
	var envelope wire.MovesEnvelope
	if err := json.Unmarshal([]byte(out), &envelope); err != nil {
		t.Fatalf("decode forced receipt: %v\n%s", err, out)
	}
	if len(envelope.Moves) != 1 || envelope.Moves[0].Lifecycle == nil {
		t.Fatalf("forced receipt = %+v", envelope)
	}
	lifecycle := envelope.Moves[0].Lifecycle
	if !lifecycle.Forced || lifecycle.Override != "dependency-gate" ||
		len(lifecycle.OutstandingBlockers) != 1 || lifecycle.OutstandingBlockers[0].TaskID != prerequisiteID ||
		lifecycle.From != "ready-to-start" || lifecycle.Remedy == "" {
		t.Fatalf("forced lifecycle detail = %+v", lifecycle)
	}
}

func TestTaskStartAcceptsQueuedWorkWithClearOrForcedBlockedGate(t *testing.T) {
	root := dependencyCLIRepo(t, dependencyCLITask{slug: "queued-clear", status: domain.StatusNextUp})
	out, errOut, err := runIn(t, root, "task", "start", "queued-clear", "--json")
	if err != nil || errOut != "" {
		t.Fatalf("clear queued start: %v\nstdout=%s\nstderr=%s", err, out, errOut)
	}
	var started wire.MovesEnvelope
	if err := json.Unmarshal([]byte(out), &started); err != nil || len(started.Moves) != 1 ||
		started.Moves[0].Lifecycle == nil || started.Moves[0].Lifecycle.From != "next-up" ||
		started.Moves[0].Lifecycle.Forced {
		t.Fatalf("clear queued receipt=%+v decode=%v", started, err)
	}

	prerequisiteID := testutil.TaskID("queued-blocker")
	root = dependencyCLIRepo(t,
		dependencyCLITask{slug: "queued-blocker", status: domain.StatusNextUp},
		dependencyCLITask{slug: "queued-blocked", status: domain.StatusNextUp, dependsOn: []string{prerequisiteID}},
	)
	out, _, err = runIn(t, root, "task", "start", "queued-blocked", "--json")
	var refused wire.MovesEnvelope
	if err == nil || json.Unmarshal([]byte(out), &refused) != nil || len(refused.Moves) != 1 ||
		refused.Moves[0].LifecycleFailure == nil || !refused.Moves[0].LifecycleFailure.OverrideAllowed {
		t.Fatalf("blocked queued refusal=%+v err=%v", refused, err)
	}

	out, errOut, err = runIn(t, root, "task", "start", "queued-blocked", "--force", "--json")
	var forced wire.MovesEnvelope
	if err != nil || errOut != "" || json.Unmarshal([]byte(out), &forced) != nil || len(forced.Moves) != 1 ||
		forced.Moves[0].Lifecycle == nil || forced.Moves[0].Lifecycle.From != "next-up" ||
		!forced.Moves[0].Lifecycle.Forced {
		t.Fatalf("forced queued receipt=%+v err=%v stderr=%s", forced, err, errOut)
	}
}

func TestTaskStartMixedBatchRetainsSuccessAndTypedFailure(t *testing.T) {
	prerequisiteID := testutil.TaskID("mixed-prerequisite")
	root := dependencyCLIRepo(t,
		dependencyCLITask{slug: "mixed-prerequisite", status: domain.StatusNextUp},
		dependencyCLITask{slug: "clear-target", status: domain.StatusReadyToStart},
		dependencyCLITask{slug: "blocked-target", status: domain.StatusReadyToStart, dependsOn: []string{prerequisiteID}},
	)
	out, _, err := runIn(t, root, "task", "start", "clear-target", "blocked-target", "--json")
	if err == nil || ExitCode(err) != 11 {
		t.Fatalf("mixed batch error = %v\n%s", err, out)
	}
	var envelope wire.MovesEnvelope
	if decodeErr := json.Unmarshal([]byte(out), &envelope); decodeErr != nil {
		t.Fatalf("decode mixed batch: %v\n%s", decodeErr, out)
	}
	if len(envelope.Moves) != 2 || envelope.Moves[0].Lifecycle == nil || !envelope.Moves[0].Lifecycle.Committed ||
		envelope.Moves[1].LifecycleFailure == nil || envelope.Moves[1].LifecycleFailure.Blockers[0].TaskID != prerequisiteID {
		t.Fatalf("mixed batch receipt = %+v", envelope.Moves)
	}
}

func TestRunMovesPreservesCommittedFailureForHumanJSONAndFatalRecovery(t *testing.T) {
	receipt := core.TaskLifecycleReceipt{
		Task: domain.Task{ID: testutil.TaskID("committed"), Slug: "committed", Status: domain.StatusInProgress},
		From: domain.StatusReadyToStart, To: domain.StatusInProgress,
		Changed: true, Committed: true, Remedy: "inspect blockers before retrying",
	}
	cause := &core.TaskLifecycleMutationFailure{Cause: fmt.Errorf("unlock: %w", domain.ErrConflict), Receipt: receipt}

	var stdout, stderr bytes.Buffer
	app := &App{Out: &stdout, ErrOut: &stderr, JSON: true, Style: render.NewStyle(false)}
	err := runMoves(app, []string{"committed"}, string(domain.StatusInProgress),
		func(string) (core.TaskLifecycleReceipt, error) { return receipt, cause },
		func(got core.TaskLifecycleReceipt) string { return got.Task.Slug })
	var commandFailure *taskLifecycleCommandFailure
	if !errors.As(err, &commandFailure) {
		t.Fatalf("runMoves error = %T %v", err, err)
	}
	var moves wire.MovesEnvelope
	if decodeErr := json.Unmarshal(stdout.Bytes(), &moves); decodeErr != nil || len(moves.Moves) != 1 ||
		moves.Moves[0].Lifecycle == nil || !moves.Moves[0].Lifecycle.Committed || moves.Moves[0].Error == "" {
		t.Fatalf("committed move JSON = %+v decode=%v\n%s", moves, decodeErr, stdout.String())
	}

	var fatal bytes.Buffer
	WriteError(&fatal, err, true)
	var envelope wire.ErrorEnvelope
	if decodeErr := json.Unmarshal(fatal.Bytes(), &envelope); decodeErr != nil || envelope.Error.TaskLifecycle == nil ||
		!envelope.Error.TaskLifecycle.Lifecycle.Committed || envelope.Error.TaskLifecycle.Slug != "committed" {
		t.Fatalf("committed fatal JSON = %+v decode=%v\n%s", envelope, decodeErr, fatal.String())
	}

	stdout.Reset()
	stderr.Reset()
	app.JSON = false
	_ = runMoves(app, []string{"committed"}, string(domain.StatusInProgress),
		func(string) (core.TaskLifecycleReceipt, error) { return receipt, cause },
		func(got core.TaskLifecycleReceipt) string { return got.Task.Slug })
	if !strings.Contains(stderr.String(), "committed") || strings.Contains(stderr.String(), "✘") {
		t.Fatalf("human committed warning = %q", stderr.String())
	}
}

func TestTaskMoveForceUsesDestinationSpecificGate(t *testing.T) {
	prerequisiteID := testutil.TaskID("move-prerequisite")
	root := dependencyCLIRepo(t,
		dependencyCLITask{slug: "move-prerequisite", status: domain.StatusDeferred},
		dependencyCLITask{slug: "move-target", status: domain.StatusReadyToStart, dependsOn: []string{prerequisiteID}},
	)
	_, _, err := runIn(t, root, "task", "move", "move-target", "in-progress", "--force")
	if err != nil {
		t.Fatalf("generic move should apply dependency override: %v", err)
	}

	root = dependencyCLIRepo(t, dependencyCLITask{slug: "parked", status: domain.StatusDeferred})
	_, _, err = runIn(t, root, "task", "move", "parked", "in-progress", "--force")
	if err == nil || ExitCode(err) != 11 || !strings.Contains(err.Error(), "move it to next-up or ready-to-start") {
		t.Fatalf("force bypassed lifecycle role: %v", err)
	}
}

func TestTaskReopenReportsUnsoundDescendantCountAndRemedy(t *testing.T) {
	upstreamID := testutil.TaskID("upstream")
	directID := testutil.TaskID("direct")
	root := dependencyCLIRepo(t,
		dependencyCLITask{slug: "upstream", status: domain.StatusCompleted},
		dependencyCLITask{slug: "direct", status: domain.StatusCompleted, dependsOn: []string{upstreamID}},
		dependencyCLITask{slug: "transitive", status: domain.StatusCompleted, dependsOn: []string{directID}},
	)
	out, _, err := runIn(t, root, "task", "ready", "upstream", "--json")
	if err != nil {
		t.Fatal(err)
	}
	var envelope wire.MovesEnvelope
	if err := json.Unmarshal([]byte(out), &envelope); err != nil {
		t.Fatal(err)
	}
	lifecycle := envelope.Moves[0].Lifecycle
	if lifecycle == nil || lifecycle.ImpactCount != 2 || len(lifecycle.Impacts) != 2 || lifecycle.Remedy == "" {
		t.Fatalf("reopen lifecycle detail = %+v", lifecycle)
	}
	for _, impact := range lifecycle.Impacts {
		if !impact.After.Inconsistent || impact.After.SoundlyCompleted {
			t.Errorf("reopen did not expose unsound completion: %+v", impact)
		}
	}
}

func TestTaskShow(t *testing.T) {
	root := setupRepo(t)
	out := runRoot(t, "-C", root, "task", "show", "alpha")
	if !strings.Contains(out, "slug:") || !strings.Contains(out, "# Alpha") {
		t.Errorf("unexpected show output:\n%s", out)
	}
}

func TestTaskStart_NotFound_ExitCode(t *testing.T) {
	root := setupRepo(t)
	var out bytes.Buffer
	cmd := NewRootCmd(strings.NewReader(""), &out, &out)
	cmd.SetArgs([]string{"-C", root, "task", "start", "ghost"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing task")
	}
	if ExitCode(err) != 10 {
		t.Errorf("want exit code 10 (not-found), got %d", ExitCode(err))
	}
}

func TestExitCode(t *testing.T) {
	cases := map[error]int{
		nil:                                     0,
		fmt.Errorf("x: %w", domain.ErrNotFound): 10,
		fmt.Errorf("x: %w", domain.ErrValidation): 11,
		fmt.Errorf("x: %w", domain.ErrAmbiguous):  13,
		fmt.Errorf("plain"):                       1,
	}
	for err, want := range cases {
		if got := ExitCode(err); got != want {
			t.Errorf("ExitCode(%v) = %d, want %d", err, got, want)
		}
	}
}
