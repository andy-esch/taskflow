package cli

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"os"
	"reflect"
	"strings"
	"syscall"
	"testing"

	"github.com/spf13/cobra"

	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/wire"
)

type relocationPlanningStub struct {
	cliPlanningStub
	fixes []domain.FixResult
}

func (f *relocationPlanningStub) FixFrontmatter(dryRun bool) ([]domain.FixResult, error) {
	f.fixCalls++
	f.dryRun = dryRun
	return f.fixes, f.fixErr
}

func partialPlanningFixture(prefix bool) *relocationPlanningStub {
	stub := &relocationPlanningStub{cliPlanningStub: cliPlanningStub{witness: core.NewSourceSetID()}}
	stub.fixErr = &core.FrontmatterRepairFailure{
		Cause: &os.PathError{Op: "remove", Path: "/planning/tasks/bad.md", Err: syscall.EBUSY},
		Relocation: core.FrontmatterRepairRelocation{
			SourcePath: "/planning/tasks/bad.md", DestinationPath: "/planning/tasks/corrected.md", DestinationWritten: true,
		},
	}
	if prefix {
		stub.fixes = []domain.FixResult{{Path: "/planning/tasks/earlier.md", Changes: []string{"tags normalized"}}}
	}
	return stub
}

func TestCLILintFixRetainsPartialRelocationWithoutCountingItFixed(t *testing.T) {
	if details := filesystemDetails(partialPlanningFixture(false).fixErr); details == nil || !details.Retryable {
		t.Fatal("fixture must need the committed-destination retry override")
	}
	for _, prefix := range []bool{false, true} {
		for _, asJSON := range []bool{false, true} {
			t.Run(strings.Join([]string{map[bool]string{false: "no-prefix", true: "with-prefix"}[prefix], map[bool]string{false: "human", true: "json"}[asJSON]}, "/"), func(t *testing.T) {
				fake := partialPlanningFixture(prefix)
				var stdout, stderr bytes.Buffer
				cmd, app := newTestRootCmdWithApp(strings.NewReader(""), &stdout, &stderr)
				app.Svc = core.MustNewService(fake)
				cmd.PersistentPreRunE = func(command *cobra.Command, _ []string) error { return app.bindCommandSafety(command) }
				args := []string{"lint", "--fix"}
				if asJSON {
					args = append(args, "--json")
				}
				cmd.SetArgs(args)
				err := cmd.Execute()
				if !errors.Is(err, syscall.EBUSY) || fake.fixCalls != 1 || fake.auditCalls+fake.lintCalls != 0 {
					t.Fatalf("error=%v fake=%+v", err, fake)
				}
				if !asJSON {
					if !strings.Contains(stdout.String(), "partial ID relocation") || !strings.Contains(stdout.String(), "/planning/tasks/corrected.md") || strings.Contains(stdout.String(), "nothing to fix") {
						t.Fatalf("misleading partial output: %s", stdout.String())
					}
					if !prefix && strings.Contains(stdout.String(), "file(s) fixed") {
						t.Fatalf("partial move counted fixed: %s", stdout.String())
					}
					for _, want := range []string{"compare and merge both copies", "refuses while both paths exist"} {
						if !strings.Contains(stdout.String(), want) {
							t.Errorf("partial output omits safe recovery %q: %s", want, stdout.String())
						}
					}
					return
				}
				var receipt wire.FixEnvelope
				if jsonErr := json.Unmarshal(stdout.Bytes(), &receipt); jsonErr != nil || receipt.Relocation == nil || !receipt.Relocation.DestinationWritten || receipt.Relocation.SourceRemoved || len(receipt.Fixed) != len(fake.fixes) {
					t.Fatalf("lost/incorrect relocation: %s err=%v", stdout.String(), jsonErr)
				}
				WriteError(&stderr, err, true)
				var failure wire.ErrorEnvelope
				if jsonErr := json.Unmarshal(stderr.Bytes(), &failure); jsonErr != nil || failure.Error.PlanningRepair == nil || !reflect.DeepEqual(receipt, *failure.Error.PlanningRepair) {
					t.Fatalf("stderr lost stdout receipt: %s err=%v", stderr.String(), jsonErr)
				}
				if failure.Error.Filesystem == nil || failure.Error.Filesystem.Retryable || failure.Error.Filesystem.Operation != "remove" || failure.Error.Code != "error" {
					t.Fatalf("committed move advertised blind retry or changed exit policy: %+v", failure.Error)
				}
			})
		}
	}
}

type failedRepairOutput struct{ err error }

func (w failedRepairOutput) Write([]byte) (int, error) { return 0, w.err }

func TestCLILintFixKeepsRecoveryWhenStdoutWriterReturnsError(t *testing.T) {
	fake := partialPlanningFixture(false)
	outputErr := errors.New("output unavailable")
	var stderr bytes.Buffer
	cmd, app := newTestRootCmdWithApp(strings.NewReader(""), failedRepairOutput{err: outputErr}, &stderr)
	app.Svc = core.MustNewService(fake)
	cmd.PersistentPreRunE = func(command *cobra.Command, _ []string) error { return app.bindCommandSafety(command) }
	cmd.SetArgs([]string{"lint", "--fix", "--json"})
	err := cmd.Execute()
	if !errors.Is(err, outputErr) || !errors.Is(err, syscall.EBUSY) {
		t.Fatalf("writer or mutation error was discarded: %v", err)
	}
	WriteError(&stderr, err, true)
	var failure wire.ErrorEnvelope
	if jsonErr := json.Unmarshal(stderr.Bytes(), &failure); jsonErr != nil || failure.Error.PlanningRepair == nil || failure.Error.PlanningRepair.Relocation == nil {
		t.Fatalf("stdout failure hid durable progress: %s err=%v", stderr.String(), jsonErr)
	}
}

func TestCLILintFixKeepsCommittedReceiptWhenTransientStdoutErrorIsJoined(t *testing.T) {
	fake := partialPlanningFixture(false)
	fake.fixErr.(*core.FrontmatterRepairFailure).Cause = domain.ErrConflict
	outputErr := &os.PathError{Op: "write", Path: "/dev/stdout", Err: syscall.EAGAIN}
	if details := filesystemDetails(outputErr); details == nil || !details.Retryable {
		t.Fatal("fixture must otherwise advertise a transient filesystem retry")
	}
	var stderr bytes.Buffer
	cmd, app := newTestRootCmdWithApp(strings.NewReader(""), failedRepairOutput{err: outputErr}, &stderr)
	app.Svc = core.MustNewService(fake)
	cmd.PersistentPreRunE = func(command *cobra.Command, _ []string) error { return app.bindCommandSafety(command) }
	cmd.SetArgs([]string{"lint", "--fix", "--json"})
	err := cmd.Execute()
	if !errors.Is(err, domain.ErrConflict) || !errors.Is(err, outputErr) {
		t.Fatalf("joined errors lost classification: %v", err)
	}
	WriteError(&stderr, err, true)
	var failure wire.ErrorEnvelope
	if jsonErr := json.Unmarshal(stderr.Bytes(), &failure); jsonErr != nil || failure.Error.PlanningRepair == nil || failure.Error.PlanningRepair.Relocation == nil || !failure.Error.PlanningRepair.Relocation.DestinationWritten {
		t.Fatalf("returned writer error hid committed receipt: %s err=%v", stderr.String(), jsonErr)
	}
	if failure.Error.Code != "conflict" || failure.Error.Filesystem == nil || failure.Error.Filesystem.Operation != "write" || failure.Error.Filesystem.Retryable {
		t.Fatalf("joined stdout failure encouraged replay or changed domain classification: %+v", failure.Error)
	}
}

func TestCLILintFixRetainsOrdinaryRepairRecoveryWithoutBlindReplay(t *testing.T) {
	for _, scenario := range []string{"repair failure", "post-lint failure", "stdout failure", "joined stdout failure", "dry-run", "skipped", "no-progress"} {
		t.Run(scenario, func(t *testing.T) {
			transient := &os.PathError{Op: "read", Path: "/planning/tasks/later.md", Err: syscall.EBUSY}
			outputErr := &os.PathError{Op: "write", Path: "/dev/stdout", Err: syscall.EAGAIN}
			fake := &relocationPlanningStub{
				cliPlanningStub: cliPlanningStub{witness: core.NewSourceSetID(), fixErr: transient},
				fixes:           []domain.FixResult{{Path: "/planning/tasks/earlier.md", Changes: []string{"tags normalized"}}},
			}
			committed := true
			switch scenario {
			case "post-lint failure":
				fake.fixErr, fake.lintErr = nil, transient
			case "stdout failure":
				fake.fixErr = nil
			case "dry-run":
				committed = false
			case "skipped":
				fake.fixes[0].Skipped = true
				committed = false
			case "no-progress":
				fake.fixes = nil
				committed = false
			}
			var stdout, stderr bytes.Buffer
			var output io.Writer = &stdout
			if strings.Contains(scenario, "stdout") {
				output = failedRepairOutput{err: outputErr}
			}
			cmd, app := newTestRootCmdWithApp(strings.NewReader(""), output, &stderr)
			app.Svc = core.MustNewService(fake)
			cmd.PersistentPreRunE = func(command *cobra.Command, _ []string) error { return app.bindCommandSafety(command) }
			args := []string{"lint", "--fix", "--json"}
			if scenario == "dry-run" {
				args = append(args, "--dry-run")
			}
			cmd.SetArgs(args)
			err := cmd.Execute()
			if err == nil || fake.fixCalls != 1 {
				t.Fatalf("error lost or batch replayed: %v calls=%d", err, fake.fixCalls)
			}
			if scenario != "stdout failure" && !errors.Is(err, transient) {
				t.Fatalf("primary filesystem cause lost: %v", err)
			}
			if strings.Contains(scenario, "stdout") && !errors.Is(err, outputErr) {
				t.Fatalf("stdout cause lost: %v", err)
			}
			WriteError(&stderr, err, true)
			var failure wire.ErrorEnvelope
			if jsonErr := json.Unmarshal(stderr.Bytes(), &failure); jsonErr != nil {
				t.Fatal(jsonErr)
			}
			if (failure.Error.PlanningRepair != nil) != committed || failure.Error.Filesystem == nil || failure.Error.Filesystem.Retryable == committed {
				t.Fatalf("commit=%t: recovery lost or false retry safety: %s", committed, stderr.String())
			}
			if committed {
				receipt := failure.Error.PlanningRepair
				if receipt.Relocation != nil || receipt.DryRun || len(receipt.Fixed) != 1 || receipt.Fixed[0].Path != fake.fixes[0].Path || receipt.Fixed[0].Skipped {
					t.Fatalf("ordinary repair became a partial move or lost prefix: %+v", receipt)
				}
			}
		})
	}
}

type residualPlanningStub struct {
	relocationPlanningStub
	problems []core.LoadProblem
	tasks    []core.LoadedRecord[core.TaskWithBody]
}

func (f *residualPlanningStub) ReadLintTasks() ([]core.LoadedRecord[core.TaskWithBody], []core.LoadProblem, error) {
	f.lintCalls++
	return f.tasks, f.problems, nil
}

func TestCLILintFixWriterFailureKeepsResidualValidationAndRecovery(t *testing.T) {
	for _, scenario := range []string{"clean", "unreadable", "remaining"} {
		t.Run(scenario, func(t *testing.T) {
			residual := scenario != "clean"
			fake := &residualPlanningStub{relocationPlanningStub: relocationPlanningStub{
				cliPlanningStub: cliPlanningStub{witness: core.NewSourceSetID()},
				fixes:           []domain.FixResult{{Path: "/planning/earlier.md", Changes: []string{"tags normalized"}}},
			}}
			switch scenario {
			case "unreadable":
				fake.problems = []core.LoadProblem{{EntityKind: core.EntityTask, EntityID: "000000000002", LocalPath: "/planning/broken.md", Message: "unterminated frontmatter"}}
			case "remaining":
				fake.tasks = []core.LoadedRecord[core.TaskWithBody]{{
					Source: core.RecordSource{ID: "000000000002"},
					Value:  core.TaskWithBody{Task: domain.Task{ID: "000000000002", Slug: "needs-metadata", Status: domain.StatusNextUp}},
				}}
			}
			outputErr := &os.PathError{Op: "write", Path: "/dev/stdout", Err: syscall.EAGAIN}
			var stderr bytes.Buffer
			cmd, app := newTestRootCmdWithApp(strings.NewReader(""), failedRepairOutput{err: outputErr}, &stderr)
			app.Svc = core.MustNewService(fake)
			cmd.PersistentPreRunE = func(command *cobra.Command, _ []string) error { return app.bindCommandSafety(command) }
			cmd.SetArgs([]string{"lint", "--fix", "--json"})
			err := cmd.Execute()
			wantExit, wantCode, wantUnreadable := 1, "error", 0
			if residual {
				wantExit, wantCode = 11, "validation"
				if scenario == "unreadable" {
					wantUnreadable = 1
				}
			}
			var committed *core.PlanningRepairFailure
			if !errors.Is(err, outputErr) || errors.Is(err, domain.ErrValidation) != residual || !errors.As(err, &committed) || ExitCode(err) != wantExit || fake.fixCalls != 1 {
				t.Fatalf("writer failure lost validation, recovery or original cause: %v exit=%d calls=%d", err, ExitCode(err), fake.fixCalls)
			}
			WriteError(&stderr, err, true)
			var failure wire.ErrorEnvelope
			if jsonErr := json.Unmarshal(stderr.Bytes(), &failure); jsonErr != nil {
				t.Fatal(jsonErr)
			}
			receipt := failure.Error.PlanningRepair
			if failure.Error.Code != wantCode || !strings.Contains(failure.Error.Message, "do not blindly replay") || receipt == nil || len(receipt.Fixed) != 1 || len(receipt.Unreadable) != wantUnreadable || receipt.Relocation != nil || failure.Error.Filesystem == nil || failure.Error.Filesystem.Retryable {
				t.Fatalf("inconsistent recovery envelope: %s", stderr.String())
			}
			if (len(receipt.Remaining) > 0) != (scenario == "remaining") {
				t.Fatalf("lost residual lint issues: %+v", receipt)
			}
		})
	}
}

func TestCLILintFixJoinedErrorsKeepFilesystemEvidenceCoherent(t *testing.T) {
	for _, scenario := range []string{"skipped", "preview", "committed no-space", "committed permission"} {
		t.Run(scenario, func(t *testing.T) {
			committed := strings.HasPrefix(scenario, "committed")
			primary := &os.PathError{Op: "lstat", Path: "/planning/later.md", Err: syscall.ENOTDIR}
			secondary := &os.PathError{Op: "write", Path: "/dev/stdout", Err: syscall.EAGAIN}
			if committed {
				primary.Op, primary.Err = "read", syscall.EBUSY
				secondary.Err = syscall.ENOSPC
				if scenario == "committed permission" {
					secondary.Err = syscall.EACCES
				}
			}
			fake := &relocationPlanningStub{
				cliPlanningStub: cliPlanningStub{witness: core.NewSourceSetID(), fixErr: primary},
				fixes:           []domain.FixResult{{Path: "/planning/earlier.md", Changes: []string{"tags normalized"}, Skipped: scenario == "skipped"}},
			}
			var stderr bytes.Buffer
			cmd, app := newTestRootCmdWithApp(strings.NewReader(""), failedRepairOutput{err: secondary}, &stderr)
			app.Svc = core.MustNewService(fake)
			cmd.PersistentPreRunE = func(command *cobra.Command, _ []string) error { return app.bindCommandSafety(command) }
			args := []string{"lint", "--fix", "--json"}
			if scenario == "preview" {
				args = append(args, "--dry-run")
			}
			cmd.SetArgs(args)
			err := cmd.Execute()
			if !errors.Is(err, primary) || !errors.Is(err, secondary) || fake.fixCalls != 1 || ExitCode(err) != 1 {
				t.Fatalf("joined causes lost or operation replayed: %v calls=%d", err, fake.fixCalls)
			}
			WriteError(&stderr, err, true)
			var failure wire.ErrorEnvelope
			if jsonErr := json.Unmarshal(stderr.Bytes(), &failure); jsonErr != nil {
				t.Fatal(jsonErr)
			}
			got := failure.Error.Filesystem
			if got == nil || got.Path != primary.Path || got.Operation != primary.Op || got.Class != "io" || got.Retryable || (failure.Error.PlanningRepair != nil) != committed {
				t.Fatalf("joined failures produced contradictory recovery: %s", stderr.String())
			}
		})
	}
}
