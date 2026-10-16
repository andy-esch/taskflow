package store

import (
	"errors"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/testutil"
)

func TestDependencyRetriesPinBothEndpointsAndRecomputeCurrentEdges(t *testing.T) {
	for _, remove := range []bool{false, true} {
		for _, endpoint := range []string{"dependent", "prerequisite"} {
			for _, missing := range []bool{false, true} {
				t.Run(map[bool]string{false: "add", true: "remove"}[remove]+"/"+endpoint+map[bool]string{false: "/rename", true: "/missing"}[missing], func(t *testing.T) {
					root := t.TempDir()
					depID, preID, extraID := testutil.TaskID("dependent"), testutil.TaskID("prerequisite"), testutil.TaskID("extra")
					deps := []string(nil)
					if remove {
						deps = []string{preID}
					}
					depPath := writeGraphMutationTask(t, root, "dependent", domain.StatusReadyToStart, deps, "")
					prePath := writeGraphMutationTask(t, root, "prerequisite", domain.StatusCompleted, nil, "")
					writeGraphMutationTask(t, root, "extra", domain.StatusCompleted, nil, "")
					var afterDrift map[string]testutil.TreeEntry
					old := testHookBeforeGraphVerify
					t.Cleanup(func() { testHookBeforeGraphVerify = old })
					testHookBeforeGraphVerify = func() {
						testHookBeforeGraphVerify = nil
						path, selectedID := depPath, depID
						if endpoint == "prerequisite" {
							path, selectedID = prePath, preID
						}
						if missing {
							if err := os.Remove(path); err != nil {
								t.Fatal(err)
							}
						} else {
							renamed := filepath.Join(filepath.Dir(path), selectedID+"-renamed.md")
							if err := os.Rename(path, renamed); err != nil {
								t.Fatal(err)
							}
							if endpoint == "dependent" {
								depPath = renamed
							}
						}
						impostorPath := writeGraphMutationTask(t, root, "impostor", domain.StatusReadyToStart, nil, "")
						// Give the replacement a different source ID but the old alias.
						impostorAlias := filepath.Join(filepath.Dir(path), testutil.TaskID("impostor")+"-"+endpoint+".md")
						if err := os.Rename(impostorPath, impostorAlias); err != nil {
							t.Fatal(err)
						}
						if missing {
							idImpostor := writeGraphMutationTask(t, root, "id-impostor", domain.StatusReadyToStart, nil, "")
							if err := os.Rename(idImpostor, filepath.Join(filepath.Dir(path), testutil.TaskID("id-impostor")+"-"+selectedID+".md")); err != nil {
								t.Fatal(err)
							}
						}
						if !missing {
							// Concurrent legitimate dependency addition must survive the retry.
							content, err := os.ReadFile(depPath)
							if err != nil {
								t.Fatal(err)
							}
							changed, err := updateFrontmatter(content, map[string]any{"depends_on": append(slices.Clone(deps), extraID)})
							if err != nil {
								t.Fatal(err)
							}
							testutil.Write(t, depPath, string(changed))
						}
						afterDrift = testutil.SnapshotTree(t, root)
					}
					svc := core.MustNewService(testutil.Must(NewFS(root, core.UnrestrictedMutations())), core.WithRetry(2, func(int) {}))
					var receipt core.DependencyMutationReceipt
					var err error
					if remove {
						receipt, err = svc.RemoveTaskDependencies("dependent", []string{"prerequisite"}, false)
					} else {
						receipt, err = svc.AddTaskDependencies("dependent", []string{"prerequisite"}, false)
					}
					if afterDrift == nil {
						t.Fatalf("pre-write drift seam was not reached: %v", err)
					}
					after := testutil.SnapshotTree(t, root)
					if missing {
						if err == nil || !maps.Equal(afterDrift, after) {
							t.Fatalf("missing endpoint was retargeted: receipt=%+v err=%v", receipt, err)
						}
						return
					}
					if err != nil || len(receipt.Edges) != 1 || receipt.Edges[0].DependentID != depID || receipt.Edges[0].PrerequisiteID != preID {
						t.Fatalf("endpoint identity drift: receipt=%+v err=%v", receipt, err)
					}
					read, err := svc.ShowTask(depID)
					if err != nil || !slices.Contains(read.Value.Task.DependsOn, extraID) || slices.Contains(read.Value.Task.DependsOn, preID) == remove {
						t.Fatalf("retry replayed stale edges: task=%+v err=%v", read, err)
					}
					key := filepath.Join(domain.TasksDir, filepath.Base(depPath))
					delete(afterDrift, key)
					delete(after, key)
					if !maps.Equal(afterDrift, after) {
						t.Fatal("dependency retry changed a sibling")
					}
				})
			}
		}
	}
}

func TestThreadCreationBindsMembersBeforeCollisionValidation(t *testing.T) {
	root := t.TempDir()
	memberID := testutil.TaskID("member")
	memberPath := writeGraphMutationTask(t, root, "member", domain.StatusReadyToStart, nil, "")
	collision, _ := createThreadForMutation(t, root, "colliding-thread")
	var afterDrift map[string]testutil.TreeEntry
	svc := core.MustNewService(testutil.Must(NewFS(root, core.UnrestrictedMutations())),
		core.WithIDGen(func() string { return collision.Thread.ID }),
		core.WithRetry(1, func(int) {
			// Validation rejected the new Thread's colliding ID after member
			// selection. A raw editor removes that collision before the retry.
			if err := os.Remove(collision.Local.CommittedPath); err != nil {
				t.Fatal(err)
			}
			if err := os.Rename(memberPath, filepath.Join(filepath.Dir(memberPath), memberID+"-renamed.md")); err != nil {
				t.Fatal(err)
			}
			impostor := writeGraphMutationTask(t, root, "member-impostor", domain.StatusReadyToStart, nil, "")
			if err := os.Rename(impostor, filepath.Join(filepath.Dir(impostor), testutil.TaskID("member-impostor")+"-member.md")); err != nil {
				t.Fatal(err)
			}
			afterDrift = testutil.SnapshotTree(t, root)
		}))
	receipt, err := svc.NewThread(core.NewThreadParams{Title: "new-thread", Description: "creation after collision", Goal: "retain members", Tasks: []string{"member"}})
	if err != nil || !receipt.Committed || !slices.Equal(receipt.Thread.Tasks, []string{memberID}) || afterDrift == nil {
		t.Fatalf("collision retry changed member selection: receipt=%+v err=%v", receipt, err)
	}
	after := testutil.SnapshotTree(t, root)
	delete(after, filepath.Join(domain.ThreadsDir, filepath.Base(receipt.Local.CommittedPath)))
	if !maps.Equal(afterDrift, after) {
		t.Fatal("creation retry changed an existing source")
	}
}

func TestThreadRetriesKeepSelectedThreadAndMemberIdentities(t *testing.T) {
	for _, operation := range []string{"start", "add", "remove", "create"} {
		for _, missing := range []bool{false, true} {
			t.Run(operation+map[bool]string{false: "/rename", true: "/missing"}[missing], func(t *testing.T) {
				root := t.TempDir()
				memberID, extraID := testutil.TaskID("member"), testutil.TaskID("extra")
				memberPath := writeGraphMutationTask(t, root, "member", domain.StatusReadyToStart, nil, "")
				writeGraphMutationTask(t, root, "extra", domain.StatusCompleted, nil, "")
				members := []string{extraID}
				if operation == "remove" {
					members = append(members, memberID)
				}
				created, _ := createThreadForMutation(t, root, "selected-thread", members...)
				threadPath := created.Local.CommittedPath
				hook := &testHookBeforeThreadMutationVerify
				if operation == "create" {
					hook = &testHookBeforeThreadCreationVerify
				}
				old := *hook
				t.Cleanup(func() { *hook = old })
				var afterDrift map[string]testutil.TreeEntry
				*hook = func() {
					*hook = nil
					path, selectedID, alias := memberPath, memberID, "member"
					if operation == "start" {
						path, selectedID, alias = threadPath, created.Thread.ID, "selected-thread"
					}
					if missing {
						if err := os.Remove(path); err != nil {
							t.Fatal(err)
						}
					} else {
						renamed := filepath.Join(filepath.Dir(path), selectedID+"-renamed.md")
						if err := os.Rename(path, renamed); err != nil {
							t.Fatal(err)
						}
						if operation == "start" {
							threadPath = renamed
						}
					}
					if operation == "start" {
						// Copy the original intent to a distinct-ID Thread reusing its alias.
						content := "---\nschema: 1\nid: " + testutil.TaskID("thread-impostor") + "\nstatus: unstarted\ndescription: impostor\ngoal: remain untouched\ncreated: 2026-10-15\ntasks: [" + extraID + "]\n---\n# Impostor\n"
						testutil.Write(t, filepath.Join(filepath.Dir(path), testutil.TaskID("thread-impostor")+"-"+alias+".md"), content)
						if missing {
							content := "---\nschema: 1\nid: " + testutil.TaskID("thread-id-impostor") + "\nstatus: unstarted\ndescription: impostor\ngoal: remain untouched\ncreated: 2026-10-15\ntasks: [" + extraID + "]\n---\n# ID-shaped impostor\n"
							testutil.Write(t, filepath.Join(filepath.Dir(path), testutil.TaskID("thread-id-impostor")+"-"+selectedID+".md"), content)
						}
					} else {
						impostor := writeGraphMutationTask(t, root, "member-impostor", domain.StatusReadyToStart, nil, "")
						if err := os.Rename(impostor, filepath.Join(filepath.Dir(impostor), testutil.TaskID("member-impostor")+"-"+alias+".md")); err != nil {
							t.Fatal(err)
						}
						if missing {
							idImpostor := writeGraphMutationTask(t, root, "member-id-impostor", domain.StatusReadyToStart, nil, "")
							if err := os.Rename(idImpostor, filepath.Join(filepath.Dir(idImpostor), testutil.TaskID("member-id-impostor")+"-"+selectedID+".md")); err != nil {
								t.Fatal(err)
							}
						}
					}
					if !missing && (operation == "add" || operation == "remove") {
						// A fresh concurrent membership addition is not part of the old plan.
						otherID := testutil.TaskID("concurrent")
						writeGraphMutationTask(t, root, "concurrent", domain.StatusCompleted, nil, "")
						content, err := os.ReadFile(threadPath)
						if err != nil {
							t.Fatal(err)
						}
						currentMembers := append(slices.Clone(members), otherID)
						slices.Sort(currentMembers)
						changed, err := updateFrontmatter(content, map[string]any{"tasks": currentMembers})
						if err != nil {
							t.Fatal(err)
						}
						testutil.Write(t, threadPath, string(changed))
					}
					afterDrift = testutil.SnapshotTree(t, root)
				}
				svc := core.MustNewService(testutil.Must(NewFS(root, core.UnrestrictedMutations())), core.WithRetry(2, func(int) {}), core.WithIDGen(func() string { return testutil.TaskID("new-thread") }))
				var got domain.Thread
				var err error
				switch operation {
				case "start":
					r, e := svc.StartThread("selected-thread", false)
					got, err = r.Thread, e
				case "add":
					r, e := svc.AddThreadMembers("selected-thread", []string{"member"}, false)
					got, err = r.Thread, e
				case "remove":
					r, e := svc.RemoveThreadMembers("selected-thread", []string{"member"}, false)
					got, err = r.Thread, e
				case "create":
					r, e := svc.NewThread(core.NewThreadParams{Title: "new-thread", Description: "creation retry", Goal: "fixed members", Tasks: []string{"member"}})
					got, err = r.Thread, e
					threadPath = filepath.Join(root, domain.ThreadsDir, testutil.TaskID("new-thread")+"-new-thread.md")
				}
				if afterDrift == nil {
					t.Fatalf("drift seam not reached: %v", err)
				}
				after := testutil.SnapshotTree(t, root)
				if missing {
					if err == nil || (!errors.Is(err, domain.ErrNotFound) && !errors.Is(err, domain.ErrValidation)) || !maps.Equal(afterDrift, after) {
						t.Fatalf("missing endpoint was not refused: err=%v unchanged=%v", err, maps.Equal(afterDrift, after))
					}
					return
				}
				wantID := created.Thread.ID
				if operation == "create" {
					wantID = testutil.TaskID("new-thread")
				}
				if err != nil || got.ID != wantID {
					t.Fatalf("Thread identity drift: got=%+v err=%v", got, err)
				}
				if operation == "add" || operation == "create" {
					if !slices.Contains(got.Tasks, memberID) {
						t.Fatalf("member identity drift: %+v", got)
					}
				}
				if operation == "add" || operation == "remove" {
					if !slices.Contains(got.Tasks, testutil.TaskID("concurrent")) || slices.Contains(got.Tasks, memberID) == (operation == "remove") {
						t.Fatalf("stale membership replay: %+v", got)
					}
				}
				key := filepath.Join(domain.ThreadsDir, filepath.Base(threadPath))
				delete(afterDrift, key)
				delete(after, key)
				if !maps.Equal(afterDrift, after) {
					t.Fatal("Thread retry changed a sibling")
				}
			})
		}
	}
}
