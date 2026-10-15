package core

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/testutil"
)

type criterionIdentityStore struct {
	nopStore
	records      map[string]LoadedRecord[TaskWithBody]
	readErrors   map[string]error
	transformRef string
	transforms   int
	body         string
}

func (s *criterionIdentityStore) ReadTask(ref string) (LoadedRecord[TaskWithBody], error) {
	if err := s.readErrors[ref]; err != nil {
		return LoadedRecord[TaskWithBody]{}, err
	}
	record, ok := s.records[ref]
	if !ok {
		return LoadedRecord[TaskWithBody]{}, domain.ErrNotFound
	}
	return record, nil
}

func (s *criterionIdentityStore) TransformTaskBody(target TaskMutationTarget, _ time.Time, dryRun bool, transform func(string) (string, error)) (domain.Task, string, bool, error) {
	s.transforms++
	s.transformRef = target.ID
	next, err := transform(s.body)
	if err != nil {
		return domain.Task{}, "", false, err
	}
	changed := next != s.body
	if !dryRun {
		s.body = next
	}
	return domain.Task{}, next, changed, nil
}

func TestCriteriaTrackingUsesSourceIdentity(t *testing.T) {
	sourceID, destinationID := testutil.TaskID("source"), testutil.TaskID("destination")
	for _, tc := range []struct {
		name, sourceIdentity, destinationIdentity, sourceSlug, destinationSlug string
		wantRefusal                                                            bool
	}{
		{"same source different declarations", sourceID, sourceID, "renamed-source", "renamed-destination", true},
		{"distinct sources same display slug", sourceID, destinationID, "source-alias", "source-alias", false},
		{"missing source identity", "", destinationID, "source", "destination", true},
		{"blank source identity", "  ", destinationID, "source", "destination", true},
		{"missing destination identity", sourceID, "", "source", "destination", true},
		{"blank destination identity", sourceID, "  ", "source", "destination", true},
	} {
		for _, dryRun := range []bool{false, true} {
			t.Run(tc.name+"/"+map[bool]string{false: "write", true: "preview"}[dryRun], func(t *testing.T) {
				initial := "# Task\n\n## Acceptance criteria\n\n- [ ] works\n"
				st := &criterionIdentityStore{body: initial, records: map[string]LoadedRecord[TaskWithBody]{
					"source-alias":      {Source: RecordSource{ID: tc.sourceIdentity, Location: "remote:source"}, Value: TaskWithBody{Task: domain.Task{ID: testutil.TaskID("declared-source"), Slug: tc.sourceSlug}}},
					"destination-alias": {Source: RecordSource{ID: tc.destinationIdentity, Location: "remote:destination"}, Value: TaskWithBody{Task: domain.Task{ID: testutil.TaskID("declared-destination"), Slug: tc.destinationSlug}}},
				}}
				st.records[sourceID] = st.records["source-alias"]
				_, body, _, err := MustNewService(st).SetCriteriaTracked("source-alias", []int{1}, "destination-alias", "follow up", dryRun)
				if tc.wantRefusal {
					if !errors.Is(err, domain.ErrValidation) || st.transforms != 0 || st.body != initial {
						t.Fatalf("invalid tracking reached effects: err=%v transforms=%d body=%q", err, st.transforms, st.body)
					}
					return
				}
				if err != nil || st.transforms != 1 || st.transformRef != sourceID || !strings.Contains(body, "by "+destinationID) {
					t.Fatalf("tracking did not bind canonical identities: ref=%q body=%q transforms=%d err=%v", st.transformRef, body, st.transforms, err)
				}
				if dryRun && st.body != initial {
					t.Fatal("preview changed the stored body")
				}
			})
		}
	}
}

func TestCriteriaTrackingPreservesCanonicalResolutionErrors(t *testing.T) {
	sourceID := testutil.TaskID("source")
	for _, want := range []error{domain.ErrAmbiguous, domain.ErrNotFound, domain.ErrConflict} {
		for _, dryRun := range []bool{false, true} {
			t.Run(want.Error()+"/"+map[bool]string{false: "write", true: "preview"}[dryRun], func(t *testing.T) {
				initial := "# Task\n\n## Acceptance criteria\n\n- [ ] works\n"
				record := LoadedRecord[TaskWithBody]{Source: RecordSource{ID: sourceID, Location: "remote:source"}}
				st := &criterionIdentityStore{
					body:       initial,
					records:    map[string]LoadedRecord[TaskWithBody]{"source-alias": record, "destination-alias": record},
					readErrors: map[string]error{sourceID: want},
				}
				_, _, changed, err := MustNewService(st).SetCriteriaTracked("source-alias", []int{1}, "destination-alias", "follow up", dryRun)
				if !errors.Is(err, want) || strings.Contains(err.Error(), "its own task") || changed || st.transforms != 0 || st.body != initial {
					t.Fatalf("canonical resolution error was masked or reached effects: err=%v changed=%v transforms=%d body=%q", err, changed, st.transforms, st.body)
				}
			})
		}
	}
}

func TestCriteriaTrackingDoesNotMisreportCanonicalFallbackAsSelf(t *testing.T) {
	sourceID := testutil.TaskID("source")
	for _, dry := range []bool{false, true} {
		record := LoadedRecord[TaskWithBody]{Source: RecordSource{ID: sourceID}}
		st := &criterionIdentityStore{records: map[string]LoadedRecord[TaskWithBody]{
			"source": record, "same-source-alias": record,
			sourceID: {Source: RecordSource{ID: testutil.TaskID("impostor")}},
		}}
		_, _, _, err := MustNewService(st).SetCriteriaTracked("source", []int{1}, "same-source-alias", "follow up", dry)
		if !errors.Is(err, domain.ErrNotFound) || strings.Contains(err.Error(), "its own task") || st.transforms != 0 {
			t.Fatalf("canonical fallback was misreported or reached mutation: err=%v transforms=%d", err, st.transforms)
		}
	}
}

type taskCreationProbeStore struct {
	fakeStore
	epicReads int
}

func (s *taskCreationProbeStore) ReadEpics() (EpicRead, error) {
	s.epicReads++
	return s.fakeStore.ReadEpics()
}

func TestNewTaskRejectsConflictingLifecycleOptionsBeforeEffects(t *testing.T) {
	for _, dryRun := range []bool{false, true} {
		st := &taskCreationProbeStore{fakeStore: fakeStore{epics: []domain.Epic{{ID: "01-e"}}}}
		mints := 0
		svc := MustNewService(st, WithIDGen(func() string { mints++; return testutil.TaskID("created") }))
		_, err := svc.NewTask(NewTaskParams{Title: "Conflicting", Epic: "01-e", Tags: []string{"go"}, Description: "work", Next: true, Start: true, DryRun: dryRun})
		if !errors.Is(err, domain.ErrValidation) || !strings.Contains(err.Error(), "mutually exclusive") || st.epicReads != 0 || mints != 0 || len(st.created) != 0 {
			t.Fatalf("dryRun=%v conflict did not fail before effects: err=%v reads=%d mints=%d creates=%d", dryRun, err, st.epicReads, mints, len(st.created))
		}
	}
}

func TestSetFieldsValidationIsDeterministic(t *testing.T) {
	for _, tc := range []struct {
		name    string
		updates map[string]any
		force   bool
		want    string
	}{
		{"unknown fields", map[string]any{"zzz": "1", "aaa": "2"}, false, `unknown field "aaa"`},
		{"unset fields", map[string]any{"zzz": domain.UnsetField{}, "aaa": domain.UnsetField{}}, false, `unknown field "aaa"`},
		{"typed values", map[string]any{"tier": "bad", "autonomy_level": "bad"}, false, "autonomy_level must be an integer"},
		{"owned fields with force", map[string]any{"status": "completed", "depends_on": []string{"other"}}, true, "depends_on is graph-owned"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			st := &conflictStore{}
			svc := MustNewService(st)
			for range 100 {
				_, err := svc.SetFields("task", tc.updates, tc.force, false)
				if !errors.Is(err, domain.ErrValidation) || !strings.Contains(err.Error(), tc.want) || st.setCalls != 0 {
					t.Fatalf("validation changed or reached writes: err=%v writes=%d want=%q", err, st.setCalls, tc.want)
				}
			}
		})
	}
}
