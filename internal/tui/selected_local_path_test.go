package tui

import (
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/store"
	"github.com/andy-esch/taskflow/internal/testutil"
)

func selectedLocalPathModel(t *testing.T, kind entityKind) (Model, string, string) {
	t.Helper()
	root := seedRepo(t)
	if kind == entityThreads {
		root = threadRepo(t)
	}
	m := loadedAt(t, root, 120, 40)
	if m.cur().kind != kind {
		cmd := m.switchTab(indexOfKind(m.tabs, kind))
		m = settleAtlasCmd(t, m, cmd)
	}
	ref := m.selectedRef()
	fs := testutil.Must(store.NewFS(root, core.ReadOnlyMutations()))
	var path string
	var err error
	switch kind {
	case entityTasks:
		path, err = fs.ResolveTaskPathByID(ref.key)
	case entityEpics:
		path, err = fs.ResolveEpicPathByID(ref.key)
	case entityAudits:
		path, err = fs.ResolveAuditPathByID(ref.key)
	case entityResearch:
		path, err = fs.ResolveResearchPathByID(ref.key)
	case entityThreads:
		path, err = fs.ResolveThreadPathByID(ref.key)
	}
	if err != nil || ref.empty() {
		t.Fatalf("missing selected local fixture: ref=%+v err=%v", ref, err)
	}
	return m, root, path
}

func TestSelectedLocalPathActionsRefuseImpostorsAtEitherLookup(t *testing.T) {
	for _, kind := range []entityKind{entityTasks, entityEpics, entityAudits, entityResearch, entityThreads} {
		for _, action := range []string{"E", "Y"} {
			for _, phase := range []string{"initial", "confirmation"} {
				for _, mode := range []string{"missing", "impostor", "case-folded impostor", "duplicate"} {
					if kind == entityEpics && mode == "duplicate" {
						// Whole stems cannot duplicate in one directory.
						continue
					}
					t.Run(fmt.Sprintf("kind=%d/%s/%s/%s", kind, action, phase, mode), func(t *testing.T) {
						m, root, path := selectedLocalPathModel(t, kind)
						ref := m.selectedRef()
						tm, lookup := m.Update(press(action))
						m = tm.(Model)
						if lookup == nil {
							t.Fatal("action did not request a path")
						}
						if phase == "confirmation" {
							tm, lookup = m.Update(lookup())
							m = tm.(Model)
							if lookup == nil {
								t.Fatal("action did not recheck its identity")
							}
						}
						if mode != "duplicate" {
							if err := os.Remove(path); err != nil {
								t.Fatal(err)
							}
						}
						if mode != "missing" {
							name := testutil.TaskID("path-impostor") + "-" + ref.key + ".md"
							if mode == "case-folded impostor" {
								name = testutil.TaskID("path-impostor") + "-" + strings.ToUpper(ref.key) + ".md"
							}
							if mode == "duplicate" {
								name = ref.key + "-duplicate.md"
							}
							if kind == entityEpics {
								name = ref.key + "-other.md"
								if mode == "case-folded impostor" {
									name = strings.ToUpper(ref.key) + ".md"
								}
							}
							// Malformed content is deliberate: local recovery is parse-free;
							// only source ownership may decide which path is executable.
							testutil.Write(t, filepath.Join(filepath.Dir(path), name), "---\nbroken: [\n---\n# Must not open\n")
						}
						before := testutil.SnapshotTree(t, root)
						tm, effect := m.Update(lookup())
						m = tm.(Model)
						if phase == "initial" {
							if effect == nil {
								t.Fatal("initial lookup did not request its confirmation")
							}
							tm, effect = m.Update(effect())
							m = tm.(Model)
						}
						// Do not execute the returned effect, even on a broken implementation.
						if effect != nil || !m.flashErr || !strings.Contains(m.flash, "local path unavailable") || !maps.Equal(before, testutil.SnapshotTree(t, root)) {
							t.Fatalf("selected local action accepted another source: effect=%v flash=%q", effect != nil, m.flash)
						}
					})
				}
			}
		}
	}
}

func TestSelectedLocalPathResultsCannotCrossSameIDWorkspaceSessions(t *testing.T) {
	for _, action := range []string{"E", "Y"} {
		for _, phase := range []string{"initial", "confirmation"} {
			t.Run(action+"/"+phase, func(t *testing.T) {
				m, _, _ := selectedLocalPathModel(t, entityTasks)
				m.sessionScope, m.sessionGen = true, 1
				tm, lookup := m.Update(press(action))
				m = tm.(Model)
				if phase == "confirmation" {
					tm, lookup = m.Update(lookup())
					m = tm.(Model)
				}
				// A separately loaded checkout has matching IDs/list generations.
				// Only the session stamp prevents accepting the old path there.
				next, _, _ := selectedLocalPathModel(t, entityTasks)
				next.sessionScope, next.sessionGen = true, 2
				if next.selectedKey() != m.selectedKey() || next.cur().coherentGen != m.cur().coherentGen {
					t.Fatal("fixture does not collide across workspace sessions")
				}
				tm, effect := next.Update(lookup())
				if effect != nil || tm.(Model).flash != "" {
					t.Fatal("old workspace path result escaped its session")
				}
			})
		}
	}
}

func TestSelectedLocalActionsKeepParseFreeRepairAndRenameSupport(t *testing.T) {
	for _, kind := range []entityKind{entityTasks, entityEpics, entityAudits, entityResearch, entityThreads} {
		for _, action := range []string{"E", "Y"} {
			t.Run(fmt.Sprintf("kind=%d/%s", kind, action), func(t *testing.T) {
				m, root, path := selectedLocalPathModel(t, kind)
				ref := m.selectedRef()
				tm, lookup := m.Update(press(action))
				m = tm.(Model)
				if kind != entityEpics {
					renamed := filepath.Join(filepath.Dir(path), ref.key+"-renamed.md")
					if err := os.Rename(path, renamed); err != nil {
						t.Fatal(err)
					}
					testutil.Write(t, filepath.Join(filepath.Dir(path), testutil.TaskID("old-alias")+"-"+ref.label+".md"), "# unrelated alias\n")
					path = renamed
				}
				testutil.Write(t, path, "---\nbroken: [\n---\n# Needs manual repair\n")
				before := testutil.SnapshotTree(t, root)
				first, ok := lookup().(localPathResultMsg)
				if !ok || first.err != nil || first.path != path {
					t.Fatalf("initial repair path: %+v", first)
				}
				tm, confirm := m.Update(first)
				m = tm.(Model)
				second, ok := confirm().(localPathResultMsg)
				if !ok || second.err != nil || second.path != path || !second.rechecked {
					t.Fatalf("confirmed repair path: %+v", second)
				}
				tm, effect := m.Update(second)
				if effect == nil || tm.(Model).flashErr || !maps.Equal(before, testutil.SnapshotTree(t, root)) {
					t.Fatal("selected repair lost optional path access or changed sources")
				}
				// Inspect only the effect's existence; never open an editor or clipboard.
			})
		}
	}
}
