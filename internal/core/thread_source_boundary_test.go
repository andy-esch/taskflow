package core

import (
	"errors"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/testutil"
)

func TestThreadMutationSourceValidatorsRetainIndependentIdentity(t *testing.T) {
	thread := threadRecord(domain.ThreadStatusUnstarted)
	for _, test := range []struct {
		name string
		edit func(*ThreadRead)
		want string
	}{
		{"matching pathless", func(*ThreadRead) {}, ""},
		{"missing source", func(read *ThreadRead) { read.Records[0].Record.Source.ID = "" }, "canonical source ID"},
		{"missing declaration", func(read *ThreadRead) { read.Records[0].Record.Value.ID = "" }, "disagrees"},
		{"drifting declaration", func(read *ThreadRead) { read.Records[0].Record.Value.ID = testutil.TaskID("drifting-thread") }, "disagrees"},
		{"duplicate source", func(read *ThreadRead) {
			shadow := read.Records[0]
			shadow.Record.Source.Location = "db://threads/shadow"
			read.Records = append(read.Records, shadow)
		}, "duplicate canonical Thread source ID"},
		{"two sources one declaration", func(read *ThreadRead) {
			shadow := read.Records[0]
			shadow.Record.Source.ID = testutil.TaskID("second-source")
			read.Records = append(read.Records, shadow)
		}, "disagrees"},
	} {
		t.Run(test.name, func(t *testing.T) {
			read := semanticThreadRead(thread)
			read.Records[0].Record.Source.Location = "/looks/local/but-is-context.md"
			test.edit(&read)
			original := append([]VersionedRecord[domain.Thread](nil), read.Records...)
			for name, validate := range map[string]func(*TaskGraph, ThreadRead) error{
				"creation": ValidateThreadCreationSource, "mutation": ValidateThreadMutationSource,
			} {
				err := validate(NewTaskGraph(nil, nil), read)
				if test.want == "" {
					if err != nil {
						t.Fatalf("%s pathless source = %v", name, err)
					}
				} else if !errors.Is(err, domain.ErrValidation) || !strings.Contains(err.Error(), test.want) {
					t.Fatalf("%s = %v, want validation containing %q", name, err, test.want)
				}
			}
			if !reflect.DeepEqual(read.Records, original) {
				t.Fatal("source validation reordered or changed the adapter snapshot")
			}
		})
	}
}

func TestLoadedThreadsOwnProjectionValuesWithoutGuardedEvidence(t *testing.T) {
	thread := threadRecord(domain.ThreadStatusUnstarted, testutil.TaskID("member"))
	thread.Tags = []string{"original"}
	read := semanticThreadRead(thread)
	read.Records[0].SourceVersion = "private-revision"
	read.Records[0].LocalPath = "/private/local/target.md"
	read.Records[0].Record.Source.Location = "db://threads/original"
	loaded := read.LoadedThreads()
	if len(loaded) != 1 || loaded[0].Source != read.Records[0].Record.Source || !reflect.DeepEqual(loaded[0].Value, thread) {
		t.Fatalf("loaded Thread = %+v", loaded)
	}
	loaded[0].Value.Tags[0] = "changed"
	loaded[0].Value.Tasks[0] = "changed"
	if read.Records[0].Record.Value.Tags[0] != "original" || !slices.Equal(read.Records[0].Record.Value.Tasks, thread.Tasks) {
		t.Fatal("projection values alias the guarded snapshot")
	}
}

func TestThreadGraphImpactsRetainSourceDefectsAfterRepair(t *testing.T) {
	member := graphRecord("source-impact-member", domain.StatusNextUp, "invalid-token")
	before := NewTaskGraph([]domain.Task{member}, nil)
	member.DependsOn = nil
	after := NewTaskGraph([]domain.Task{member}, nil)
	thread := threadRecord(domain.ThreadStatusUnstarted, member.ID)
	for _, test := range []struct {
		name   string
		source RecordSource
		code   ThreadProblemCode
		path   string
	}{
		{"opaque drift", RecordSource{ID: testutil.TaskID("actual-source"), Location: "db://threads/actual"}, ThreadProblemIDDrift, ""},
		{"local drift", RecordSource{ID: testutil.TaskID("actual-source"), Location: "/planning/threads/actual.md", LocationIsPath: true}, ThreadProblemIDDrift, "/planning/threads/actual.md"},
		{"missing source", RecordSource{Location: "db://threads/unknown"}, ThreadProblemInvalidDocument, ""},
	} {
		t.Run(test.name, func(t *testing.T) {
			impacts := TaskGraphThreadImpacts([]LoadedRecord[domain.Thread]{{Value: thread, Source: test.source}}, before, after, []string{member.ID})
			if len(impacts) != 1 || impacts[0].ThreadID != test.source.ID || !impacts[0].Direct || impacts[0].After.ProjectionHealth != GraphBroken || len(impacts[0].After.Frontier) != 0 {
				t.Fatalf("repair impact lost the source defect: %+v", impacts)
			}
			for _, view := range []ThreadView{impacts[0].Before, impacts[0].After} {
				if view.Source != test.source {
					t.Fatalf("impact source = %+v, want %+v", view.Source, test.source)
				}
				found := false
				for _, problem := range view.Problems {
					if problem.Code == test.code && problem.Path == test.path {
						found = true
					}
				}
				if !found {
					t.Fatalf("missing attributable source problem: %+v", view.Problems)
				}
			}
		})
	}
}

func TestThreadImpactsQualifyDuplicateSourcesBeforeFiltering(t *testing.T) {
	member := graphRecord("duplicate-changing-member", domain.StatusNextUp)
	done := graphRecord("duplicate-unchanged-member", domain.StatusCompleted)
	before := NewTaskGraph([]domain.Task{member, done}, nil)
	changed := member
	changed.Status = domain.StatusInProgress
	after := NewTaskGraph([]domain.Task{changed, done}, nil)
	sourceID := testutil.TaskID("duplicate-impact-source")
	active := threadRecord(domain.ThreadStatusUnstarted, member.ID)
	active.ID, active.Slug = sourceID, "alpha"
	completed := threadRecord(domain.ThreadStatusCompleted, done.ID)
	completed.ID, completed.Slug = sourceID, "beta"
	records := []LoadedRecord[domain.Thread]{
		{Value: completed, Source: RecordSource{ID: sourceID, Location: "db://threads/beta"}},
		{Value: active, Source: RecordSource{ID: sourceID, Location: "db://threads/alpha"}},
	}
	for name, impacts := range map[string][]ThreadProjectionImpact{
		"graph": TaskGraphThreadImpacts(records, before, after, []string{member.ID}),
		"lifecycle": TaskLifecycleThreadImpacts(records, before, TaskLifecyclePlan{
			TaskID: member.ID, To: domain.StatusInProgress,
		}),
	} {
		t.Run(name, func(t *testing.T) {
			// The unchanged duplicate is filtered out of the receipt, but must
			// still qualify the changed occurrence as broken on both sides.
			if len(impacts) != 1 || impacts[0].ThreadID != sourceID || impacts[0].Slug != "alpha" {
				t.Fatalf("impacts = %+v", impacts)
			}
			for _, view := range []ThreadView{impacts[0].Before, impacts[0].After} {
				if view.ProjectionHealth != GraphBroken || len(view.Frontier) != 0 || !hasThreadProblem(view.Problems, ThreadProblemDuplicateID) {
					t.Fatalf("filtering erased the duplicate source defect: %+v", view)
				}
				for _, problem := range view.Problems {
					if problem.Code == ThreadProblemDuplicateID && (problem.ThreadID != sourceID || problem.Path != "") {
						t.Fatalf("opaque duplicate diagnostic lost source attribution: %+v", problem)
					}
				}
			}
		})
	}
	// Once task evidence is repaired, completion must not clear a remaining
	// duplicate source defect, even for an otherwise fully drained Thread.
	member.DependsOn = []string{"invalid-token"}
	broken := NewTaskGraph([]domain.Task{member, done}, nil)
	impacts := TaskGraphThreadImpacts(records, broken, after, []string{member.ID})
	if len(impacts) != 2 || impacts[1].Slug != "beta" || !impacts[1].After.Inconsistent ||
		!hasThreadProblem(impacts[1].After.Problems, ThreadProblemDuplicateID) ||
		!hasThreadProblem(impacts[1].After.Problems, ThreadProblemCompletedUnhealthyEvidence) {
		t.Fatalf("completed duplicate became consistent after graph repair: %+v", impacts)
	}
	if records[0].Value.Slug != "beta" || records[1].Value.Slug != "alpha" {
		t.Fatal("impact computation reordered the adapter records")
	}
}
