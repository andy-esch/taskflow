package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

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
