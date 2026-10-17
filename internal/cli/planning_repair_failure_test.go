package cli

import (
	"bytes"
	"encoding/json"
	"errors"
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
