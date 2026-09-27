package cli

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"

	"github.com/spf13/cobra"

	"github.com/andy-esch/taskflow/internal/cli/render"
	"github.com/andy-esch/taskflow/internal/config"
	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/wire"
)

func TestCreateCommandsKeepPlannedAndCommittedLocalPaths(t *testing.T) {
	for _, tc := range []struct {
		name string
		args []string
		dir  string
	}{
		{"task", []string{"task", "new", "Created", "--epic", "01-test", "--tags", "test"}, "tasks"},
		{"task-start", []string{"task", "new", "Created", "--epic", "01-test", "--tags", "test", "--description", "started", "--start"}, "tasks"},
		{"epic", []string{"epic", "new", "Created", "--description", "goal"}, "epics"},
		{"audit", []string{"audit", "new", "Created", "--date", "2026-09-27"}, "audits"},
		{"research", []string{"research", "new", "Created", "--created", "2026-09-27"}, "research"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := freshRepo(t)
			if tc.dir == "tasks" {
				mustWrite(t, filepath.Join(root, "epics", "01-test.md"), "---\nstatus: active\npriority: medium\ndescription: test\n---\n# Test\n")
			}
			decode := func(dry bool) wire.CreatedEnvelope {
				args := append([]string(nil), tc.args...)
				args = append(args, "--json")
				if dry {
					args = append([]string{"--dry-run"}, args...)
				}
				stdout, stderr, err := runIn(t, root, args...)
				if err != nil || stderr != "" {
					t.Fatalf("create dry=%v: err=%v stderr=%s", dry, err, stderr)
				}
				var envelope wire.CreatedEnvelope
				if err := json.Unmarshal([]byte(stdout), &envelope); err != nil {
					t.Fatalf("decode created envelope: %v\n%s", err, stdout)
				}
				if envelope.SchemaVersion != wire.SchemaVersion || envelope.DryRun != dry ||
					!strings.HasPrefix(envelope.Created.Path, tc.dir+"/") {
					t.Fatalf("created envelope dry=%v: %+v", dry, envelope)
				}
				return envelope
			}
			preview := decode(true)
			if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(preview.Created.Path))); !os.IsNotExist(err) {
				t.Fatalf("preview wrote destination: %v", err)
			}
			committed := decode(false)
			if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(committed.Created.Path))); err != nil {
				t.Fatalf("committed destination missing: %v", err)
			}
		})
	}
}

type committedCreateStore struct {
	core.Store
	root      string
	cause     error
	sourceSet core.SourceSetID
}

func (s *committedCreateStore) SourceSetID() core.SourceSetID { return s.sourceSet }

func (s *committedCreateStore) ReadEpics() (core.EpicRead, error) {
	return core.EpicRead{Records: []core.LoadedRecord[domain.Epic]{{Value: domain.Epic{ID: "01-test"}, Source: core.RecordSource{ID: "01-test"}}}}, nil
}

func (s *committedCreateStore) CreateTask(task domain.Task, _ string, _ bool) (core.TaskCreationReceipt, error) {
	path := filepath.Join(s.root, "tasks", task.ID+"-"+task.Slug+".md")
	return core.TaskCreationReceipt{Task: task, Local: core.LocalCreateOutcome{CommittedPath: path}, Committed: true}, s.cause
}

func (s *committedCreateStore) CreateEpic(_ string, epic domain.Epic, _ string, _ bool) (core.EpicCreationReceipt, error) {
	epic.ID = "01-created"
	path := filepath.Join(s.root, "epics", epic.ID+".md")
	return core.EpicCreationReceipt{Epic: epic, Local: core.LocalCreateOutcome{CommittedPath: path}, Committed: true}, s.cause
}

func (s *committedCreateStore) CreateAudit(audit domain.Audit, _ string, _ bool) (core.AuditCreationReceipt, error) {
	path := filepath.Join(s.root, "audits", audit.ID+"-"+audit.Slug+".md")
	return core.AuditCreationReceipt{Audit: audit, Local: core.LocalCreateOutcome{CommittedPath: path}, Committed: true}, s.cause
}

func (s *committedCreateStore) CreateResearch(research domain.Research, _ string, _ bool) (core.ResearchCreationReceipt, error) {
	path := filepath.Join(s.root, "research", research.ID+"-"+research.Slug+".md")
	return core.ResearchCreationReceipt{Research: research, Local: core.LocalCreateOutcome{CommittedPath: path}, Committed: true}, s.cause
}

func TestCreateCommandsProjectCommittedFailuresFromKindSpecificReceipts(t *testing.T) {
	for _, tc := range []struct {
		kind, dir string
		args      []string
		build     func(*App) *cobra.Command
	}{
		{"task", "tasks", []string{"Created", "--epic", "01-test", "--tags", "test"}, newTaskNewCmd},
		{"epic", "epics", []string{"Created", "--description", "goal"}, newEpicNewCmd},
		{"audit", "audits", []string{"Created", "--date", "2026-09-27"}, newAuditNewCmd},
		{"research", "research", []string{"Created", "--created", "2026-09-27"}, newResearchNewCmd},
	} {
		t.Run(tc.kind, func(t *testing.T) {
			root := t.TempDir()
			adapter := &committedCreateStore{root: root, cause: fmt.Errorf("release failed: %w", domain.ErrConflict), sourceSet: core.NewSourceSetID()}
			stdout := &bytes.Buffer{}
			app := &App{Svc: core.MustNewService(adapter), Cfg: &config.Config{Root: root}, JSON: true,
				Out: stdout, ErrOut: &bytes.Buffer{}}
			cmd := tc.build(app)
			cmd.SetArgs(tc.args)
			cmd.SetOut(app.Out)
			cmd.SetErr(app.ErrOut)
			cmd.SilenceUsage = true
			cmd.SilenceErrors = true
			err := cmd.Execute()
			if err == nil || !errors.Is(err, domain.ErrConflict) || !strings.Contains(err.Error(), "inspect the committed document before retrying") {
				t.Fatalf("create error = %v", err)
			}
			if stdout.Len() != 0 {
				t.Fatalf("failed create wrote success output: %q", stdout.String())
			}
			var out bytes.Buffer
			WriteError(&out, err, true)
			var got wire.ErrorEnvelope
			if err := json.Unmarshal(out.Bytes(), &got); err != nil || got.Error.Created == nil ||
				got.Error.Created.Kind != tc.kind || !got.Error.Created.Committed ||
				!strings.HasPrefix(got.Error.Created.Path, tc.dir+"/") ||
				got.Error.Created.Workspace.PlanningRoot != physicalPath(root) {
				t.Fatalf("committed %s recovery = %s decode=%v", tc.kind, out.String(), err)
			}
		})
	}
}

func TestPathlessCreationProjectionDoesNotInventALocation(t *testing.T) {
	app := &App{Cfg: &config.Config{Root: "/planning"}, Style: render.NewStyle(false)}
	if got := app.rel(""); got != "" {
		t.Fatalf("empty local outcome became relative path %q", got)
	}
	if got := app.linkPath(""); got != "" {
		t.Fatalf("empty local outcome became link %q", got)
	}
	var out bytes.Buffer
	render.CreatedHuman(&out, app.Style, app.linkPath(""), false)
	if got := out.String(); got != "created\n" {
		t.Fatalf("pathless human creation = %q", got)
	}
	out.Reset()
	render.ThreadCreatedHuman(&out, app.Style, core.ThreadCreationReceipt{Thread: domain.Thread{ID: "6gdx7mn9f0a4"}}, app.linkPath(""))
	if got := out.String(); !strings.Contains(got, "Thread 6gdx7mn9f0a4") || strings.Contains(got, " at ") {
		t.Fatalf("pathless human Thread creation = %q", got)
	}
}

func TestCommittedCreateFailuresRetainKindIdentityAndWorkspace(t *testing.T) {
	root := "/planning"
	app := &App{Cfg: &config.Config{Root: root}}
	for _, tc := range []struct {
		kind, id, slug, status, dir string
	}{
		{"task", "6ge7qn9ptaa1", "created", "ready-to-start", "tasks"},
		{"epic", "01-created", "01-created", "active", "epics"},
		{"audit", "6ge7qn9ptaa2", "2026-09-27-created", "open", "audits"},
		{"research", "6ge7qn9ptaa3", "created", "", "research"},
	} {
		t.Run(tc.kind, func(t *testing.T) {
			localPath := filepath.Join(root, tc.dir, tc.id+"-created.md")
			err := committedCreateFailure(app, fmt.Errorf("release failed: %w", domain.ErrConflict),
				tc.kind, tc.id, tc.slug, tc.status, localPath)
			if !errors.Is(err, domain.ErrConflict) || ExitCode(err) != exitConflict ||
				!strings.Contains(err.Error(), localPath) || !strings.Contains(err.Error(), "inspect the committed document before retrying") {
				t.Fatalf("human recovery/classification = %v", err)
			}
			var out bytes.Buffer
			WriteError(&out, err, true)
			var got wire.ErrorEnvelope
			if decodeErr := json.Unmarshal(out.Bytes(), &got); decodeErr != nil || got.Error.Created == nil {
				t.Fatalf("machine recovery = %s decode=%v", out.String(), decodeErr)
			}
			recovery := got.Error.Created
			if got.SchemaVersion != wire.SchemaVersion || got.Error.Code != "conflict" || !recovery.Committed ||
				recovery.Kind != tc.kind || recovery.ID != tc.id || recovery.Slug != tc.slug ||
				recovery.Status != tc.status || recovery.Path != filepath.Join(tc.dir, tc.id+"-created.md") ||
				recovery.Workspace.PlanningRoot != root {
				t.Fatalf("machine recovery = %+v", got)
			}
		})
	}
	pathless := committedCreateFailure(app, domain.ErrConflict, "task", "6ge7qn9ptaa4", "remote", "ready-to-start", "")
	if strings.Contains(pathless.Error(), " at ") {
		t.Fatalf("pathless recovery invented a path: %v", pathless)
	}
	var out bytes.Buffer
	WriteError(&out, pathless, true)
	var got wire.ErrorEnvelope
	if err := json.Unmarshal(out.Bytes(), &got); err != nil || got.Error.Created == nil || got.Error.Created.Path != "" {
		t.Fatalf("pathless recovery = %s err=%v", out.String(), err)
	}
	out.Reset()
	WriteError(&out, domain.ErrConflict, true)
	got = wire.ErrorEnvelope{}
	if err := json.Unmarshal(out.Bytes(), &got); err != nil || got.Error.Created != nil {
		t.Fatalf("pre-commit error claimed a created document: %s err=%v", out.String(), err)
	}
}

func TestCommittedCreateFilesystemErrorDoesNotInviteCommandRetry(t *testing.T) {
	app := &App{Cfg: &config.Config{Root: "/planning"}}
	for _, cause := range []error{syscall.EAGAIN, syscall.EINTR} {
		t.Run(cause.Error(), func(t *testing.T) {
			osErr := &os.PathError{Op: "flock", Path: "/planning", Err: cause}
			var out bytes.Buffer
			WriteError(&out, committedCreateFailure(app, osErr, "task", "6ge7qn9ptaa1", "created", "ready-to-start", "/planning/tasks/created.md"), true)
			var got wire.ErrorEnvelope
			if err := json.Unmarshal(out.Bytes(), &got); err != nil {
				t.Fatalf("decode recovery: %v\n%s", err, out.String())
			}
			if got.Error.Created == nil || !got.Error.Created.Committed || got.Error.Filesystem == nil ||
				got.Error.Filesystem.Retryable || got.Error.Filesystem.Operation != "flock" || got.Error.Filesystem.Class != "io" {
				t.Fatalf("unsafe or incomplete committed-create recovery: %+v", got.Error)
			}
			// The generic filesystem classifier still describes a standalone
			// transient OS call accurately; only the committed command is unsafe.
			if generic := filesystemDetails(osErr); generic == nil || !generic.Retryable {
				t.Fatalf("transient OS error misclassified: %+v", generic)
			}
		})
	}
}
