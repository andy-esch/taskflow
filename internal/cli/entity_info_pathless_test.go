package cli

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/spf13/cobra"

	"github.com/andy-esch/taskflow/internal/cli/render"
	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/testutil"
)

// An ordinary read-only adapter need not pretend its records live on disk.
// Embedding Store supplies only unused semantic methods, not any path port.
type pathlessInfoStore struct {
	core.Store
	sourceSet core.SourceSetID
}

func (s *pathlessInfoStore) SourceSetID() core.SourceSetID { return s.sourceSet }

func (*pathlessInfoStore) ReadTask(string) (core.LoadedRecord[core.TaskWithBody], error) {
	id := testutil.TaskID("pathless-task-info")
	return core.LoadedRecord[core.TaskWithBody]{
		Value: core.TaskWithBody{Task: domain.Task{ID: id, Slug: "pathless-task", Status: domain.StatusReadyToStart},
			Body: "## Acceptance criteria\n\n- [x] done\n"},
		Source: core.RecordSource{ID: id, Location: "db:task-row"},
	}, nil
}

func (*pathlessInfoStore) ReadAudit(string) (core.LoadedRecord[core.AuditWithBody], error) {
	id := testutil.TaskID("pathless-audit-info")
	return core.LoadedRecord[core.AuditWithBody]{
		Value: core.AuditWithBody{Audit: domain.Audit{ID: "declared-other", Slug: "pathless-audit", Bucket: domain.AuditOpen,
			Findings: 1, OpenFindings: 1, UnparsedFindings: 2}},
		Source: core.RecordSource{ID: id, Location: "db:audit-row"},
	}, nil
}

func TestInfoCommandsRemainSemanticWithoutLocalPath(t *testing.T) {
	svc := core.MustNewService(&pathlessInfoStore{sourceSet: core.NewSourceSetID()})
	for _, tc := range []struct {
		name string
		cmd  func(*App) *cobra.Command
		arg  string
		key  string
	}{
		{name: "task", cmd: newTaskInfoCmd, arg: "pathless-task", key: "task_info"},
		{name: "audit", cmd: newAuditInfoCmd, arg: "pathless-audit", key: "audit_info"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var out bytes.Buffer
			app := &App{Svc: svc, JSON: true, Out: &out, ErrOut: &bytes.Buffer{}, Style: render.NewStyle(false)}
			cmd := tc.cmd(app)
			cmd.SetArgs([]string{tc.arg})
			cmd.SilenceUsage, cmd.SilenceErrors = true, true
			if err := cmd.Execute(); err != nil {
				t.Fatalf("pathless %s info: %v", tc.name, err)
			}
			var envelope map[string]json.RawMessage
			if err := json.Unmarshal(out.Bytes(), &envelope); err != nil {
				t.Fatalf("decode info: %v\n%s", err, out.String())
			}
			var info struct {
				Path             string `json:"path"`
				UnparsedFindings int    `json:"unparsed_findings"`
			}
			if err := json.Unmarshal(envelope[tc.key], &info); err != nil || info.Path != "" {
				t.Fatalf("pathless %s info invented local path %q: %v", tc.name, info.Path, err)
			}
			if tc.name == "audit" && info.UnparsedFindings != 2 {
				t.Fatalf("pathless audit info lost evidence: %+v", info)
			}
			out.Reset()
			app.JSON = false
			cmd = tc.cmd(app)
			cmd.SetArgs([]string{tc.arg})
			cmd.SilenceUsage, cmd.SilenceErrors = true, true
			if err := cmd.Execute(); err != nil {
				t.Fatalf("pathless %s human info: %v", tc.name, err)
			}
			if !bytes.Contains(out.Bytes(), []byte("unavailable (no local path capability)")) {
				t.Fatalf("pathless %s human info did not explain missing path:\n%s", tc.name, out.String())
			}
			if tc.name == "audit" && (!bytes.Contains(out.Bytes(), []byte("2 finding-like header(s)")) || !bytes.Contains(out.Bytes(), []byte("audit lint "+testutil.TaskID("pathless-audit-info"))) || bytes.Contains(out.Bytes(), []byte("audit lint declared-other"))) {
				t.Fatalf("pathless audit info omitted incomplete evidence: %s", out.String())
			}
		})
	}
}
