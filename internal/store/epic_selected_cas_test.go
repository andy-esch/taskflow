package store

import (
	"maps"
	"path/filepath"
	"testing"

	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/testutil"
)

func TestEpicSelectedWritesKeepWholeStemAtCAS(t *testing.T) {
	for _, stem := range []string{"01-ordinary", "202601010000-launch"} {
		for _, selected := range []bool{false, true} {
			for _, move := range []bool{false, true} {
				for _, dry := range []bool{false, true} {
					name := stem + map[bool]string{false: "/reference", true: "/selected"}[selected] + map[bool]string{false: "/fields", true: "/move"}[move] + map[bool]string{false: "/write", true: "/preview"}[dry]
					t.Run(name, func(t *testing.T) {
						root := t.TempDir()
						content := "---\nstatus: active\npriority: medium\ndescription: source\ncustom: keep\n---\n# Source\nbody\n"
						writeEpic(t, root, stem+".md", content)
						writeEpic(t, root, stem+"-sibling.md", content)
						before := testutil.SnapshotTree(t, root)
						svc := core.MustNewService(testutil.Must(NewFS(root, core.UnrestrictedMutations())), core.WithRetry(0, func(int) {}))
						target := core.EpicMutationTarget{ID: stem}
						var got domain.Epic
						var err error
						switch {
						case move && selected:
							got, err = svc.MoveSelectedEpic(target, "retired", dry)
						case move:
							got, err = svc.MoveEpic(stem, "retired", dry)
						case selected:
							got, err = svc.SetSelectedEpicFields(target, map[string]any{"priority": "high"}, false, dry)
						default:
							got, err = svc.SetEpicFields(stem, map[string]any{"priority": "high"}, false, dry)
						}
						if err != nil || got.ID != stem || (move && got.Status != "retired") || (!move && got.Priority != "high") {
							t.Fatalf("whole-stem mutation failed: epic=%+v err=%v", got, err)
						}
						after := testutil.SnapshotTree(t, root)
						if dry {
							if !maps.Equal(before, after) {
								t.Fatal("preview changed source bytes or modes")
							}
							return
						}
						key := filepath.Join(domain.EpicsDir, stem+".md")
						if before[key] == after[key] {
							t.Fatal("write did not change selected epic")
						}
						delete(before, key)
						delete(after, key)
						if !maps.Equal(before, after) {
							t.Fatal("write changed a sibling")
						}
					})
				}
			}
		}
	}
}
