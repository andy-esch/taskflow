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
