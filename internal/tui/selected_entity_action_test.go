package tui

import (
	"errors"
	"maps"
	"os"
	"path/filepath"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/store"
	"github.com/andy-esch/taskflow/internal/testutil"
)

func TestQueuedSelectedNonTaskActionsDoNotReselect(t *testing.T) {
	for _, action := range []struct {
		name string
		epic bool
		cmd  func(*core.Service, entityRef) tea.Cmd
	}{
		{"audit move", false, func(s *core.Service, ref entityRef) tea.Cmd {
			return moveAudit(s, ref, transition{to: string(domain.AuditClosed)})
		}},
		{"epic move", true, func(s *core.Service, ref entityRef) tea.Cmd {
			return moveEpic(s, ref, transition{to: "retired"})
		}},
		{"epic field", true, func(s *core.Service, ref entityRef) tea.Cmd {
			return setEpicFieldCmd(s, ref, "priority", "high")
		}},
	} {
		for _, impostor := range []bool{false, true} {
			t.Run(action.name+map[bool]string{false: "/missing", true: "/impostor"}[impostor], func(t *testing.T) {
				root := seedRepo(t)
				svc := core.MustNewService(testutil.Must(store.NewFS(root, core.UnrestrictedMutations())))
				ref := entityRef{key: testutil.TaskID("2026-06-01-thing"), label: "2026-06-01-thing"}
				path := filepath.Join(root, domain.AuditsDir, ref.key+"-"+ref.label+".md")
				if action.epic {
					ref = entityRef{key: "01-test", label: "01-test"}
					path = filepath.Join(root, domain.EpicsDir, ref.key+".md")
				}
				cmd := action.cmd(svc, ref) // selected intent predates the first Service call
				if err := os.Remove(path); err != nil {
					t.Fatal(err)
				}
				if impostor {
					name := testutil.TaskID("impostor") + "-" + ref.key + ".md"
					if action.epic {
						name = ref.key + "-two.md"
					}
					testutil.Write(t, filepath.Join(filepath.Dir(path), name), "---\nbucket: open\narea: impostor\ndate: 2026-10-15\nstatus: active\npriority: medium\ndescription: impostor\n---\n# Impostor\n")
				}
				before := testutil.SnapshotTree(t, root)
				result := cmd()
				failure, ok := result.(actionErrMsg)
				if !ok || !errors.Is(failure.err, domain.ErrNotFound) || !maps.Equal(before, testutil.SnapshotTree(t, root)) {
					t.Fatalf("selected action reinterpreted its target: %T %+v", result, result)
				}
			})
		}
	}
}
