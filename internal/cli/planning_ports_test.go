package cli

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
)

// Unimplemented embedded methods panic if reached: these probes must use only
// the explicitly injected application capabilities, never a filesystem fallback.
type cliPlanningStub struct {
	core.Store
	witness         core.SourceSetID
	fixCalls        int
	dryRun          bool
	fixErr          error
	auditCalls      int
	lintCalls       int
	lintErr         error
	linkCalls       int
	completionCalls int
	completionKind  core.EntityKind
	completionErr   error
	withState       bool
}

func (f *cliPlanningStub) SourceSetID() core.SourceSetID { return f.witness }

func (f *cliPlanningStub) ReadLintTasks() ([]core.LoadedRecord[core.TaskWithBody], []core.LoadProblem, error) {
	f.lintCalls++
	return nil, nil, f.lintErr
}

func (*cliPlanningStub) ReadLintEpics() ([]core.LoadedRecord[domain.Epic], []core.LoadProblem, error) {
	return nil, nil, nil
}

func (*cliPlanningStub) ReadLintResearch() ([]core.LoadedRecord[domain.Research], []core.LoadProblem, error) {
	return nil, nil, nil
}

func (f *cliPlanningStub) ReadAuditSnapshot(string) (core.AuditSnapshot, error) {
	f.auditCalls++
	return core.AuditSnapshot{}, nil
}

func (f *cliPlanningStub) FixFrontmatter(dryRun bool) ([]domain.FixResult, error) {
	f.fixCalls++
	f.dryRun = dryRun
	return []domain.FixResult{{Changes: []string{"fixed portable metadata"}}}, f.fixErr
}

func (f *cliPlanningStub) DanglingLinks() ([]core.LoadProblem, error) {
	f.linkCalls++
	return []core.LoadProblem{{EntityKind: core.EntityTask, EntityID: "opaque-key", Location: "db://records/problem", Message: "unresolved body reference"}}, nil
}

func (f *cliPlanningStub) ReadCompletionCandidates(kind core.EntityKind, state bool) ([]core.CompletionCandidate, error) {
	f.completionCalls++
	f.completionKind, f.withState = kind, state
	return []core.CompletionCandidate{{ID: "opaque-key", Slug: "portable", Reference: "record:portable", SlugIsReference: true}}, f.completionErr
}

func TestCLICompletionUsesInjectedApplicationForEveryEntity(t *testing.T) {
	for _, kind := range []core.EntityKind{core.EntityTask, core.EntityThread, core.EntityEpic, core.EntityAudit, core.EntityResearch} {
		t.Run(string(kind), func(t *testing.T) {
			fake := &cliPlanningStub{witness: core.NewSourceSetID()}
			svc := core.MustNewService(nil, core.WithCompletionSource(fake))
			var stdout, stderr bytes.Buffer
			cmd, app := newTestRootCmdWithApp(strings.NewReader(""), &stdout, &stderr)
			compositions := 0
			app.CompletionService = func() (*core.Service, error) {
				compositions++
				return svc, nil
			}
			// An unusable local anchor cannot affect a non-filesystem injection.
			cmd.SetArgs([]string{"__complete", "-C", "/no/local/planning/corpus", string(kind), "show", "por"})
			if err := cmd.Execute(); err != nil || !strings.Contains(stdout.String(), "portable\n") || compositions != 1 || fake.completionCalls != 1 || fake.completionKind != kind || fake.withState {
				t.Fatalf("completion out=%q err=%v compositions=%d source=%+v", stdout.String(), err, compositions, fake)
			}
			if fake.lintCalls+fake.auditCalls+fake.fixCalls+fake.linkCalls != 0 {
				t.Fatalf("completion scanned unrelated records: %+v", fake)
			}
		})
	}
}

func TestCLICompletionFailuresStaySilentWithoutLocalFallback(t *testing.T) {
	// A direct cwd fallback must have real candidates to reveal itself. The
	// package's own cwd has no tasks/ directory and would mask that mutation.
	t.Chdir(setupRepo(t))
	for _, mode := range []string{"missing capability", "failed source", "failed composition"} {
		t.Run(mode, func(t *testing.T) {
			sentinel := errors.New("do not print this completion failure")
			fake := &cliPlanningStub{witness: core.NewSourceSetID(), completionErr: sentinel}
			svc := core.MustNewService(nil)
			if mode == "failed source" {
				svc = core.MustNewService(nil, core.WithCompletionSource(fake))
			}
			var stdout, stderr bytes.Buffer
			cmd, app := newTestRootCmdWithApp(strings.NewReader(""), &stdout, &stderr)
			compositions := 0
			app.CompletionService = func() (*core.Service, error) {
				compositions++
				if mode == "failed composition" {
					return nil, sentinel
				}
				return svc, nil
			}
			cmd.SetArgs([]string{"__complete", "task", "show", ""})
			if err := cmd.Execute(); err != nil || stdout.String() != ":4\n" || strings.Contains(stderr.String(), sentinel.Error()) || compositions != 1 {
				t.Fatalf("failed completion out=%q stderr=%q err=%v compositions=%d", stdout.String(), stderr.String(), err, compositions)
			}
			wantCalls := 0
			if mode == "failed source" {
				wantCalls = 1
			}
			if fake.completionCalls != wantCalls {
				t.Fatalf("source calls=%d want=%d", fake.completionCalls, wantCalls)
			}
		})
	}
}

func TestCLILintMaintenanceUsesPortableApplicationPorts(t *testing.T) {
	for _, mode := range []string{"links", "fix", "dry-run", "partial frontmatter", "post-lint failure"} {
		t.Run(mode, func(t *testing.T) {
			fake := &cliPlanningStub{witness: core.NewSourceSetID()}
			sentinel := errors.New("later phase failed")
			if mode == "partial frontmatter" {
				fake.fixErr = sentinel
			}
			if mode == "post-lint failure" {
				fake.lintErr = sentinel
			}
			var stdout, stderr bytes.Buffer
			cmd, app := newTestRootCmdWithApp(strings.NewReader(""), &stdout, &stderr)
			app.Svc = core.MustNewService(fake)
			// Test composition supplies the complete application service; no local
			// discovery is needed or permitted by this command invocation.
			cmd.PersistentPreRunE = func(command *cobra.Command, _ []string) error {
				return app.bindCommandSafety(command)
			}
			args := []string{"lint", "--json"}
			if mode == "links" {
				args = append(args, "--links")
			} else {
				args = append(args, "--fix")
			}
			if mode == "dry-run" {
				args = append(args, "--dry-run")
			}
			cmd.SetArgs(args)
			err := cmd.Execute()
			if !json.Valid(stdout.Bytes()) || stderr.Len() != 0 {
				t.Fatalf("machine output=%q stderr=%q err=%v", stdout.String(), stderr.String(), err)
			}
			if mode == "links" {
				if ExitCode(err) != 11 || fake.linkCalls != 1 || fake.fixCalls != 0 || fake.lintCalls != 1 || !strings.Contains(stdout.String(), "db://records/problem") {
					t.Fatalf("link output=%q err=%v source=%+v", stdout.String(), err, fake)
				}
				return
			}
			if fake.fixCalls != 1 || fake.dryRun != (mode == "dry-run") || !strings.Contains(stdout.String(), "fixed portable metadata") {
				t.Fatalf("repair output=%q err=%v source=%+v", stdout.String(), err, fake)
			}
			if mode == "partial frontmatter" || mode == "post-lint failure" {
				if !errors.Is(err, sentinel) {
					t.Fatalf("repair failure lost: %v", err)
				}
			} else if err != nil {
				t.Fatal(err)
			}
			wantLint, wantAudit := 1, 2
			switch mode {
			case "dry-run":
				wantLint, wantAudit = 0, 1
			case "partial frontmatter":
				wantLint, wantAudit = 0, 0
			case "post-lint failure":
				wantAudit = 1
			}
			if fake.lintCalls != wantLint || fake.auditCalls != wantAudit || fake.linkCalls != 0 || fake.completionCalls != 0 {
				t.Fatalf("unexpected maintenance calls: %+v", fake)
			}
		})
	}
}
