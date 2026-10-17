package cli

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"

	"github.com/spf13/cobra"

	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/testutil"
	"github.com/andy-esch/taskflow/internal/wire"
)

func taskRenameCLIRepo(t *testing.T) (root, oldPath, newPath, inboundPath string) {
	t.Helper()
	oldID := testutil.TaskID("old")
	inboundID := testutil.TaskID("inbound")
	repo := testutil.NewRepo(t).
		Task("ready-to-start", "old.md", "---\nid: "+oldID+"\nstatus: ready-to-start\ndescription: old task\ntags: [rename]\n---\n# Old Title\n").
		Task("ready-to-start", "inbound.md", "---\nid: "+inboundID+"\nstatus: ready-to-start\ndescription: inbound link\ntags: [rename]\n---\n# Inbound\n\nSee [old]("+oldID+"-old.md).\n")
	root = repo.Root
	oldPath = filepath.Join(root, "tasks", oldID+"-old.md")
	newPath = filepath.Join(root, "tasks", oldID+"-new-title.md")
	inboundPath = filepath.Join(root, "tasks", inboundID+"-inbound.md")
	return root, oldPath, newPath, inboundPath
}

func decodeTaskRenameEnvelope(t *testing.T, output string) wire.TaskRenameEnvelope {
	t.Helper()
	var envelope wire.TaskRenameEnvelope
	decoder := json.NewDecoder(strings.NewReader(output))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&envelope); err != nil {
		t.Fatalf("decode task rename envelope: %v\n%s", err, output)
	}
	return envelope
}

func TestTaskRenameJSONReportsPreviewAndDurableOutcome(t *testing.T) {
	root, oldPath, newPath, inboundPath := taskRenameCLIRepo(t)

	previewOut, previewErr, err := runIn(t, root, "--dry-run", "task", "rename", "old", "New Title", "--json")
	if err != nil {
		t.Fatalf("preview task rename: %v\n%s%s", err, previewOut, previewErr)
	}
	preview := decodeTaskRenameEnvelope(t, previewOut)
	wantSource := filepath.Join(preview.Workspace.PlanningRoot, "tasks", filepath.Base(oldPath))
	wantDestination := filepath.Join(preview.Workspace.PlanningRoot, "tasks", filepath.Base(newPath))
	if preview.SchemaVersion != wire.SchemaVersion || preview.TaskID != testutil.TaskID("old") ||
		preview.FromSlug != "old" || preview.ToSlug != "new-title" || preview.Task.Slug != "new-title" ||
		preview.SourcePath != wantSource || preview.DestinationPath != wantDestination ||
		preview.PlannedDocuments != 2 || preview.AppliedDocuments != 0 ||
		preview.PlannedLinks != 1 || preview.AppliedLinks != 0 || !preview.Changed || !preview.DryRun ||
		preview.Committed || preview.Complete || preview.DestinationWritten || preview.SourceRemoved ||
		preview.Workspace.PlanningRoot == "" {
		t.Fatalf("preview receipt = %+v", preview)
	}
	if strings.Contains(previewOut, `"remedy"`) {
		t.Fatalf("successful preview must not carry failure recovery prose:\n%s", previewOut)
	}
	if _, err := os.Stat(oldPath); err != nil {
		t.Fatalf("preview removed source: %v", err)
	}
	if _, err := os.Stat(newPath); !os.IsNotExist(err) {
		t.Fatalf("preview created destination: %v", err)
	}

	resultOut, resultErr, err := runIn(t, root, "task", "rename", "old", "New Title", "--json")
	if err != nil {
		t.Fatalf("task rename: %v\n%s%s", err, resultOut, resultErr)
	}
	result := decodeTaskRenameEnvelope(t, resultOut)
	if result.TaskID != preview.TaskID || result.FromSlug != preview.FromSlug || result.ToSlug != preview.ToSlug ||
		result.SourcePath != wantSource || result.DestinationPath != wantDestination ||
		result.PlannedDocuments != preview.PlannedDocuments || result.PlannedLinks != preview.PlannedLinks ||
		result.AppliedDocuments != 2 || result.AppliedLinks != 1 || result.DryRun ||
		!result.Changed || !result.Committed || !result.Complete || !result.DestinationWritten || !result.SourceRemoved {
		t.Fatalf("committed receipt = %+v", result)
	}
	if _, err := os.Stat(oldPath); !os.IsNotExist(err) {
		t.Fatalf("committed rename retained source: %v", err)
	}
	if _, err := os.Stat(newPath); err != nil {
		t.Fatalf("committed rename omitted destination: %v", err)
	}
	inbound, err := os.ReadFile(inboundPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(inbound), testutil.TaskID("old")+"-new-title.md") {
		t.Fatalf("committed rename did not rewrite inbound link:\n%s", inbound)
	}
}

func TestTaskRenameHumanOutputRemainsConcise(t *testing.T) {
	root, _, _, _ := taskRenameCLIRepo(t)
	out, errOut, err := runIn(t, root, "task", "rename", "old", "New Title")
	if err != nil {
		t.Fatalf("task rename: %v\n%s%s", err, out, errOut)
	}
	if got, want := out, "✔ renamed to new-title (1 inbound link(s) repointed)\n"; got != want {
		t.Fatalf("human rename output = %q, want %q", got, want)
	}
}

func TestWriteErrorCarriesStructuredTaskRenameRecovery(t *testing.T) {
	receipt := core.TaskRenameReceipt{
		Task:               domain.Task{ID: "6g7wxs43g7nh", Slug: "new-title"},
		FromSlug:           "old-title",
		SourcePath:         "/repo/planning/tasks/6g7wxs43g7nh-old-title.md",
		DestinationPath:    "/repo/planning/tasks/6g7wxs43g7nh-new-title.md",
		PlannedDocuments:   3,
		AppliedDocuments:   2,
		PlannedLinks:       2,
		AppliedLinks:       2,
		Changed:            true,
		Committed:          true,
		DestinationWritten: true,
		Remedy:             "inspect both task paths before retrying",
	}
	cause := &core.TaskRenameFailure{Cause: domain.ErrConflict, Receipt: receipt}
	err := &taskRenameCommandFailure{
		cause: cause, receipt: receipt,
		workspace: wire.WorkspaceJSON{PlanningRoot: "/repo/planning", Source: wire.WorkspaceSourceConfig},
	}

	var out bytes.Buffer
	WriteError(&out, err, true)
	var envelope wire.ErrorEnvelope
	if decodeErr := json.Unmarshal(out.Bytes(), &envelope); decodeErr != nil {
		t.Fatalf("decode task rename recovery: %v\n%s", decodeErr, out.String())
	}
	got := envelope.Error.TaskRename
	if envelope.Error.Code != "conflict" || got == nil || got.TaskID != receipt.Task.ID ||
		got.FromSlug != receipt.FromSlug || got.ToSlug != receipt.Task.Slug ||
		got.SourcePath != receipt.SourcePath || got.DestinationPath != receipt.DestinationPath ||
		got.PlannedDocuments != 3 || got.AppliedDocuments != 2 ||
		!got.Committed || got.Complete || !got.DestinationWritten || got.SourceRemoved ||
		got.Remedy != receipt.Remedy || got.Workspace.PlanningRoot != "/repo/planning" ||
		!errors.Is(err, domain.ErrConflict) {
		t.Fatalf("task rename recovery envelope = %+v", envelope)
	}
}

type partialDestinationRenameStore struct {
	cliPlanningStub
	result core.TaskRenameMutationResult
	cause  error
	calls  int
}

func (s *partialDestinationRenameStore) RenameTask(string, string, bool) (core.TaskRenameMutationResult, error) {
	s.calls++
	return s.result, s.cause
}

func TestTaskRenameCommandRetainsDestinationFailureAndRequiresInspection(t *testing.T) {
	for _, transient := range []bool{false, true} {
		for _, asJSON := range []bool{false, true} {
			t.Run(map[bool]string{false: "conflict/", true: "transient/"}[transient]+map[bool]string{false: "human", true: "json"}[asJSON], func(t *testing.T) {
				cause := error(domain.ErrConflict)
				if transient {
					cause = &os.PathError{Op: "read", Path: "/planning/tasks/new-title.md", Err: syscall.EBUSY}
					if details := filesystemDetails(cause); details == nil || !details.Retryable {
						t.Fatal("transient fixture must need the committed-destination retry override")
					}
				}
				fake := &partialDestinationRenameStore{
					cliPlanningStub: cliPlanningStub{witness: core.NewSourceSetID()},
					result: core.TaskRenameMutationResult{
						Task: domain.Task{ID: "6fjangd7kva1", Slug: "new-title"}, FromSlug: "old",
						SourcePath: "/planning/tasks/old.md", DestinationPath: "/planning/tasks/new-title.md",
						PlannedDocuments: 2, AppliedDocuments: 2, PlannedLinks: 1, AppliedLinks: 1,
						Changed: true, Committed: true, DestinationWritten: true,
					},
					cause: cause,
				}
				var stdout, stderr bytes.Buffer
				cmd, app := newTestRootCmdWithApp(strings.NewReader(""), &stdout, &stderr)
				app.Svc = core.MustNewService(fake)
				cmd.PersistentPreRunE = func(command *cobra.Command, _ []string) error { return app.bindCommandSafety(command) }
				args := []string{"task", "rename", "old", "New title"}
				if asJSON {
					args = append(args, "--json")
				}
				cmd.SetArgs(args)
				err := cmd.Execute()
				var partial *core.TaskRenameFailure
				if !errors.Is(err, cause) || !errors.As(err, &partial) || fake.calls != 1 || stdout.Len() != 0 {
					t.Fatalf("command lost failure, reported success, or replayed rename: %v calls=%d stdout=%s", err, fake.calls, stdout.String())
				}
				WriteError(&stderr, err, asJSON)
				for _, expected := range []string{fake.result.SourcePath, fake.result.DestinationPath, "before retrying", "only after verifying"} {
					if !strings.Contains(stderr.String(), expected) {
						t.Errorf("recovery omits %q: %s", expected, stderr.String())
					}
				}
				if !asJSON {
					return
				}
				var envelope wire.ErrorEnvelope
				if err := json.Unmarshal(stderr.Bytes(), &envelope); err != nil {
					t.Fatal(err)
				}
				got := envelope.Error.TaskRename
				if got == nil || !got.Committed || !got.DestinationWritten || got.Complete || got.SourceRemoved || got.AppliedDocuments != 2 || got.AppliedLinks != 1 || got.SourcePath != fake.result.SourcePath || got.DestinationPath != fake.result.DestinationPath || got.Remedy != partial.Receipt.Remedy {
					t.Fatalf("structured receipt lost historical actions: %+v", envelope)
				}
				if transient {
					if envelope.Error.Filesystem == nil || envelope.Error.Filesystem.Retryable || envelope.Error.Filesystem.Operation != "read" || envelope.Error.Code != "error" {
						t.Fatalf("post-destination failure advertised blind retry or lost cause: %+v", envelope)
					}
				} else if envelope.Error.Code != "conflict" || envelope.Error.Filesystem != nil {
					t.Fatalf("conflict classification changed: %+v", envelope)
				}
			})
		}
	}
}

func TestTaskRenameFilesystemRetryHintRespectsDestinationCommit(t *testing.T) {
	for _, written := range []bool{false, true} {
		receipt := core.TaskRenameReceipt{
			Task: domain.Task{ID: "6fjangd7kva1", Slug: "new-title"}, FromSlug: "old",
			PlannedDocuments: 2, AppliedDocuments: 1, PlannedLinks: 1, AppliedLinks: 1,
			Changed: true, Committed: true, DestinationWritten: written,
		}
		if written {
			receipt.AppliedDocuments = 2
		}
		cause := &os.PathError{Op: "read", Path: "/planning/tasks/task.md", Err: syscall.EBUSY}
		failure := &taskRenameCommandFailure{
			cause: &core.TaskRenameFailure{Cause: cause, Receipt: receipt}, receipt: receipt,
		}
		var output bytes.Buffer
		WriteError(&output, failure, true)
		var envelope wire.ErrorEnvelope
		if err := json.Unmarshal(output.Bytes(), &envelope); err != nil {
			t.Fatal(err)
		}
		if envelope.Error.Filesystem == nil || envelope.Error.Filesystem.Retryable != !written {
			t.Fatalf("destination_written=%t gave incorrect retry hint: %s", written, output.String())
		}
	}
}
