package store

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/testutil"
)

func TestThreadMutationMaterializerRequiresExplicitSourceEvidence(t *testing.T) {
	root := t.TempDir()
	created, _ := createThreadForMutation(t, root, "source-target")
	fs := testutil.Must(NewFS(root, core.ReadOnlyMutations()))
	read, err := fs.ReadThreads()
	if err != nil || len(read.Records) != 1 {
		t.Fatalf("read=%+v err=%v", read, err)
	}
	graph, err := core.LoadTaskGraph(fs)
	if err != nil {
		t.Fatal(err)
	}
	plan, analysis, err := core.ValidateThreadMutationPlan(
		core.ThreadMutationSnapshot{Graph: graph, Threads: read.SemanticThreads()},
		core.ThreadMutationPlan{ThreadID: created.Thread.ID, Operation: core.ThreadMutationCancel}, threadMutationStoreNow,
	)
	if err != nil {
		t.Fatal(err)
	}
	before, err := os.ReadFile(created.Local.CommittedPath)
	if err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct {
		name string
		edit func(*core.VersionedRecord[domain.Thread])
		want error
	}{
		{"opaque location with explicit handle", func(source *core.VersionedRecord[domain.Thread]) {
			source.Record.Source.Location = "db://threads/source-target"
			source.Record.Source.LocationIsPath = false
		}, nil},
		{"path hint without handle", func(source *core.VersionedRecord[domain.Thread]) { source.LocalPath = "" }, domain.ErrValidation},
		{"missing revision", func(source *core.VersionedRecord[domain.Thread]) { source.SourceVersion = "" }, domain.ErrValidation},
		{"different source id", func(source *core.VersionedRecord[domain.Thread]) {
			source.Record.Source.ID = testutil.TaskID("other-source")
		}, domain.ErrValidation},
		{"different declared id", func(source *core.VersionedRecord[domain.Thread]) {
			source.Record.Value.ID = testutil.TaskID("other-declaration")
		}, domain.ErrValidation},
		{"different local handle", func(source *core.VersionedRecord[domain.Thread]) { source.LocalPath += ".other" }, domain.ErrConflict},
		{"stale revision", func(source *core.VersionedRecord[domain.Thread]) { source.SourceVersion = "stale-byte-revision" }, domain.ErrConflict},
	} {
		t.Run(test.name, func(t *testing.T) {
			source := read.Records[0]
			test.edit(&source)
			materialized, err := fs.materializeThreadMutation(source, plan, analysis)
			if test.want == nil {
				if err != nil || !materialized.changed || materialized.path != created.Local.CommittedPath || materialized.ifVersion != source.SourceVersion || materialized.thread.Status != domain.ThreadStatusCancelled {
					t.Fatalf("materialized=%+v err=%v", materialized, err)
				}
			} else if !errors.Is(err, test.want) {
				t.Fatalf("materialization=%v, want %v", err, test.want)
			}
			after, err := os.ReadFile(created.Local.CommittedPath)
			if err != nil || !slices.Equal(before, after) {
				t.Fatalf("materialization wrote the source: err=%v", err)
			}
		})
	}
}

func TestThreadSourceSnapshotRetainsReadableHandleAndRevision(t *testing.T) {
	root := t.TempDir()
	created, _ := createThreadForMutation(t, root, "source-snapshot")
	fs := testutil.Must(NewFS(root, core.ReadOnlyMutations()))
	read, err := fs.ReadThreads()
	if err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct {
		name string
		edit func(*core.VersionedRecord[domain.Thread])
	}{
		{"removed local handle", func(source *core.VersionedRecord[domain.Thread]) { source.LocalPath = "" }},
		{"changed local handle", func(source *core.VersionedRecord[domain.Thread]) { source.LocalPath += ".other" }},
		{"changed source id", func(source *core.VersionedRecord[domain.Thread]) {
			source.Record.Source.ID = testutil.TaskID("other-source")
		}},
		{"changed location", func(source *core.VersionedRecord[domain.Thread]) {
			source.Record.Source.Location = "db://threads/other"
		}},
		{"changed revision", func(source *core.VersionedRecord[domain.Thread]) { source.SourceVersion = "different-bytes" }},
		{"missing revision", func(source *core.VersionedRecord[domain.Thread]) { source.SourceVersion = "" }},
	} {
		t.Run(test.name, func(t *testing.T) {
			changed := core.ThreadRead{Records: append([]core.VersionedRecord[domain.Thread](nil), read.Records...)}
			test.edit(&changed.Records[0])
			if err := verifyThreadSourceSnapshot(read, changed); !errors.Is(err, domain.ErrConflict) {
				t.Fatalf("changed source = %v, want conflict", err)
			}
			if test.name == "missing revision" {
				if err := verifyThreadSourceSnapshot(changed, changed); !errors.Is(err, domain.ErrConflict) {
					t.Fatalf("two unversioned readable snapshots = %v, want conflict", err)
				}
			}
		})
	}
	if read.Records[0].LocalPath != created.Local.CommittedPath {
		t.Fatalf("filesystem scan did not supply a local mutation handle: %+v", read.Records[0])
	}
}

func TestThreadIdentityDefectsRemainReadableButBlockGuardedPlanners(t *testing.T) {
	for _, defect := range []string{"missing declaration", "drifting declaration", "duplicate source"} {
		t.Run(defect, func(t *testing.T) {
			root := t.TempDir()
			created, _ := createThreadForMutation(t, root, "broken-source")
			path := created.Local.CommittedPath
			content, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			switch defect {
			case "missing declaration":
				content = []byte(strings.Replace(string(content), "id: "+created.Thread.ID+"\n", "", 1))
			case "drifting declaration":
				content = []byte(strings.Replace(string(content), "id: "+created.Thread.ID, "id: "+testutil.TaskID("different-declared-thread"), 1))
			case "duplicate source":
				testutil.Write(t, filepath.Join(root, domain.ThreadsDir, created.Thread.ID+"-shadow.md"), string(content))
			}
			testutil.Write(t, path, string(content))
			fs := testutil.Must(NewFS(root, core.UnrestrictedMutations()))
			read, err := fs.ReadThreads()
			if err != nil || len(read.Records) == 0 || read.Records[0].Record.Source.ID != created.Thread.ID {
				t.Fatalf("source identity disappeared: read=%+v err=%v", read, err)
			}
			if defect != "duplicate source" {
				selected, err := fs.ReadThread(created.Thread.ID)
				if err != nil || selected.Source.ID != created.Thread.ID || selected.Source.Location != path {
					t.Fatalf("selected source=%+v err=%v", selected, err)
				}
				if resolved, err := fs.ResolveThreadPath(created.Thread.ID); err != nil || resolved != path {
					t.Fatalf("local recovery path=%q err=%v", resolved, err)
				}
			}
			for _, dry := range []bool{true, false} {
				called := false
				_, err := fs.MutateThread(threadMutationStoreNow, dry, func(core.ThreadMutationSnapshot) (core.ThreadMutationPlan, error) {
					called = true
					return core.ThreadMutationPlan{}, nil
				})
				if called || !errors.Is(err, domain.ErrValidation) {
					t.Fatalf("defective source reached mutation planner: called=%t err=%v", called, err)
				}
				_, err = fs.MutateThreadCreation(threadMutationStoreNow, dry, func(core.ThreadCreationSnapshot) (core.ThreadCreationPlan, error) {
					called = true
					return core.ThreadCreationPlan{}, nil
				})
				if called || !errors.Is(err, domain.ErrValidation) {
					t.Fatalf("defective source reached creation planner: called=%t err=%v", called, err)
				}
			}
			after, err := os.ReadFile(path)
			if err != nil || !slices.Equal(content, after) {
				t.Fatalf("invalid source was rewritten: err=%v", err)
			}
		})
	}
}

func TestGraphRepairReceiptRetainsReadableThreadIdentityDrift(t *testing.T) {
	for _, dry := range []bool{true, false} {
		t.Run(fmt.Sprintf("dry-run=%t", dry), func(t *testing.T) {
			root := t.TempDir()
			prerequisiteID := testutil.TaskID("drift-impact-prerequisite")
			memberID := testutil.TaskID("drift-impact-member")
			writeGraphMutationTask(t, root, "drift-impact-prerequisite", domain.StatusCompleted, nil, "")
			writeGraphMutationTask(t, root, "drift-impact-member", domain.StatusNextUp, []string{prerequisiteID, prerequisiteID}, "")
			sourceID := testutil.TaskID("drift-impact-source")
			declaredID := testutil.TaskID("drift-impact-declaration")
			path := filepath.Join(root, domain.ThreadsDir, sourceID+"-drift-impact.md")
			content := fmt.Sprintf("---\nid: %s\nstatus: unstarted\ndescription: Keep drift visible\ngoal: Repair only task dependencies\ncreated: \"2026-10-03\"\ntasks: [%s]\n---\n# Drift\n", declaredID, memberID)
			testutil.Write(t, path, content)
			receipt, err := core.MustNewService(testutil.Must(NewFS(root, core.UnrestrictedMutations()))).RepairTaskGraph(core.TaskGraphRepairRequest{Auto: true}, dry)
			if err != nil || !receipt.Changed || receipt.Committed == dry || receipt.FinalHealth != core.GraphHealthy || len(receipt.ThreadImpacts) != 1 {
				t.Fatalf("repair=%+v err=%v", receipt, err)
			}
			impact := receipt.ThreadImpacts[0]
			if impact.ThreadID != sourceID || impact.After.Source.ID != sourceID || impact.After.Thread.ID != declaredID || impact.After.ProjectionHealth != core.GraphBroken || len(impact.After.Frontier) != 0 {
				t.Fatalf("repair manufactured healthy Thread evidence: %+v", impact)
			}
			found := false
			for _, problem := range impact.After.Problems {
				found = found || problem.Code == core.ThreadProblemIDDrift && problem.Path == path
			}
			if !found {
				t.Fatalf("repair lost drift attribution: %+v", impact.After.Problems)
			}
			after, err := os.ReadFile(path)
			if err != nil || string(after) != content {
				t.Fatalf("task repair rewrote a Thread: err=%v", err)
			}
		})
	}
}

func TestGraphRepairReceiptRetainsDuplicateThreadSources(t *testing.T) {
	for _, dry := range []bool{true, false} {
		t.Run(fmt.Sprintf("dry-run=%t", dry), func(t *testing.T) {
			root := t.TempDir()
			prerequisiteID := testutil.TaskID("duplicate-impact-prerequisite")
			memberID := testutil.TaskID("duplicate-impact-member")
			writeGraphMutationTask(t, root, "duplicate-impact-prerequisite", domain.StatusCompleted, nil, "")
			writeGraphMutationTask(t, root, "duplicate-impact-member", domain.StatusNextUp, []string{prerequisiteID, prerequisiteID}, "")
			sourceID := testutil.TaskID("duplicate-impact-source")
			content := fmt.Sprintf("---\nid: %s\nstatus: unstarted\ndescription: Keep duplicate sources visible\ngoal: Repair only task dependencies\ncreated: \"2026-10-03\"\ntasks: [%s]\n---\n# Duplicate\n", sourceID, memberID)
			for _, slug := range []string{"alpha", "beta"} {
				testutil.Write(t, filepath.Join(root, domain.ThreadsDir, sourceID+"-"+slug+".md"), content)
			}
			svc := core.MustNewService(testutil.Must(NewFS(root, core.UnrestrictedMutations())))
			receipt, err := svc.RepairTaskGraph(core.TaskGraphRepairRequest{Auto: true}, dry)
			if err != nil || !receipt.Changed || receipt.Committed == dry || receipt.FinalHealth != core.GraphHealthy || len(receipt.ThreadImpacts) != 2 {
				t.Fatalf("repair=%+v err=%v", receipt, err)
			}
			list, problems, err := svc.ListThreadViews()
			if err != nil || len(problems) != 0 || len(list.Threads) != 2 {
				t.Fatalf("list=%+v problems=%+v err=%v", list, problems, err)
			}
			for i, impact := range receipt.ThreadImpacts {
				path := filepath.Join(root, domain.ThreadsDir, sourceID+"-"+[]string{"alpha", "beta"}[i]+".md")
				if impact.ThreadID != sourceID || !impact.Direct {
					t.Fatalf("impact lost source identity: %+v", impact)
				}
				for _, view := range []core.ThreadView{impact.Before, impact.After, list.Threads[i]} {
					if view.Source.ID != sourceID || view.Source.Location != path || view.ProjectionHealth != core.GraphBroken || len(view.Frontier) != 0 {
						t.Fatalf("duplicate source became healthy or lost attribution: %+v", view)
					}
					found := false
					for _, problem := range view.Problems {
						found = found || problem.Code == core.ThreadProblemDuplicateID && problem.ThreadID == sourceID && problem.Path == path
					}
					if !found {
						t.Fatalf("duplicate diagnostic disappeared: %+v", view.Problems)
					}
				}
				// Compare projection qualification, not operation-owned task
				// timestamps or hoisted global diagnostics in ordinary list rows.
				listed := list.Threads[i]
				if !dry && (listed.GraphHealth != impact.After.GraphHealth ||
					listed.ProjectionHealth != impact.After.ProjectionHealth ||
					listed.Inconsistent != impact.After.Inconsistent ||
					!reflect.DeepEqual(listed.Problems, impact.After.Problems) ||
					!reflect.DeepEqual(listed.Frontier, impact.After.Frontier)) {
					t.Fatalf("repair and ordinary read qualification disagree: impact=%+v listed=%+v", impact.After, listed)
				}
				after, err := os.ReadFile(path)
				if err != nil || string(after) != content {
					t.Fatalf("task repair rewrote a duplicate Thread: err=%v", err)
				}
			}
		})
	}
}
