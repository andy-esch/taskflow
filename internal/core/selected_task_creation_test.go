package core

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/andy-esch/taskflow/internal/domain"
)

// Only this narrow, pathless capability is composed: ordinary reference reads
// and creation are not available to hide an accidental fallback.
type selectedEpicCreationFake struct {
	calls     int
	target    EpicMutationTarget
	task      domain.Task
	body      string
	dry       bool
	err       error
	committed bool
}

func (*selectedEpicCreationFake) SourceSetID() SourceSetID { return testSourceSetID }

func (f *selectedEpicCreationFake) CreateTaskForSelectedEpic(target EpicMutationTarget, task domain.Task, body string, dry bool) (TaskCreationReceipt, error) {
	f.calls++
	f.target, f.task, f.body, f.dry = target, task, body, dry
	return TaskCreationReceipt{Task: task, DryRun: dry, Committed: f.committed}, f.err
}

func TestSelectedEpicTaskCreationSharesDefaultsAndScaffolding(t *testing.T) {
	for _, dry := range []bool{false, true} {
		for _, p := range []NewTaskParams{
			{Title: "Created", Tags: []string{"test"}},
			{Title: "Created", Tags: []string{"test"}, Next: true, Description: "next task"},
			{Title: "Created", Tags: []string{"test"}, Body: "# Custom body\n", Priority: "high", Tier: 2, Autonomy: 4, Effort: "one day"},
		} {
			p.Epic, p.DryRun = "01-domain", dry
			options := []Option{WithClock(fixedClock("2026-10-19")), WithIDGen(func() string { return "6gmy4c0x6e31" })}
			ordinary := MustNewService(&fakeStore{epics: []domain.Epic{{ID: p.Epic}}}, options...)
			want, err := ordinary.NewTask(p)
			if err != nil {
				t.Fatal(err)
			}
			port := &selectedEpicCreationFake{committed: !dry}
			selected := MustNewService(nil, append(options, WithSelectedEpicTaskCreationStore(port))...)
			target := EpicMutationTarget{ID: p.Epic}
			// The selected target is the only required epic input.
			p.Epic = ""
			got, err := selected.NewTaskForSelectedEpic(target, p)
			if err != nil || !reflect.DeepEqual(got.Task, want.Task) || got.Local != (LocalCreateOutcome{}) ||
				got.DryRun != dry || got.Committed == dry || port.calls != 1 || port.target != target || port.dry != dry {
				t.Fatalf("selected=%+v ordinary=%+v port=%+v err=%v", got, want, port, err)
			}
			if p.Body != "" && port.body != p.Body {
				t.Fatalf("body override changed: %q", port.body)
			}
			if p.Body == "" && (!strings.Contains(port.body, "Created") || !strings.Contains(port.body, target.ID)) {
				t.Fatalf("default scaffold lost title/epic: %q", port.body)
			}
		}
	}
}

func TestSelectedEpicCreationPreservesOpaqueIdentityWithoutReferenceLookup(t *testing.T) {
	port := &selectedEpicCreationFake{}
	svc := MustNewService(nil, WithSelectedEpicTaskCreationStore(port))
	target := EpicMutationTarget{ID: "urn:epic:selected"}
	got, err := svc.NewTaskForSelectedEpic(target, NewTaskParams{Title: "Created", Epic: target.ID, Tags: []string{"test"}})
	if err != nil || port.target != target || got.Task.Epic != target.ID {
		t.Fatalf("opaque target changed: %+v %+v %v", got, port, err)
	}
}

// Parity alone cannot catch both entry points discarding the same input through
// shared preparation. Pin caller intent and named-template bytes independently.
func TestTaskCreationPreservesIndependentExpectedValues(t *testing.T) {
	for _, tc := range []struct {
		name string
		p    NewTaskParams
		want domain.Task
		body string
	}{
		{
			name: "defaults with body",
			p:    NewTaskParams{Title: "Created", Tags: []string{"test"}, Body: "# Exact body\n100%\n"},
			want: domain.Task{ID: "6gmy4c0x6e31", Slug: "created", Epic: "01-domain", Status: domain.StatusReadyToStart,
				Priority: "medium", Tier: 3, Autonomy: 3, Effort: "Unknown", Tags: []string{"test"}, Created: "2026-10-19"},
			body: "# Exact body\n100%\n",
		},
		{
			name: "explicit metadata with body",
			p: NewTaskParams{Title: "Created", Tags: []string{"test", "authoring"}, Description: "caller intent",
				Priority: "high", Tier: 2, Autonomy: 4, Effort: "one day", Body: "# Caller body\n"},
			want: domain.Task{ID: "6gmy4c0x6e31", Slug: "created", Epic: "01-domain", Status: domain.StatusReadyToStart,
				Description: "caller intent", Priority: "high", Tier: 2, Autonomy: 4, Effort: "one day",
				Tags: []string{"test", "authoring"}, Created: "2026-10-19"},
			body: "# Caller body\n",
		},
		{
			name: "next with named template",
			p: NewTaskParams{Title: "Created", Tags: []string{"test", "authoring"}, Description: "caller intent",
				Priority: "high", Tier: 2, Autonomy: 4, Effort: "one day", Next: true, Template: "selected"},
			want: domain.Task{ID: "6gmy4c0x6e31", Slug: "created", Epic: "01-domain", Status: domain.StatusNextUp,
				Description: "caller intent", Priority: "high", Tier: 2, Autonomy: 4, Effort: "one day",
				Tags: []string{"test", "authoring"}, Created: "2026-10-19"},
			body: "# Created\nEpic 01-domain\n100% {{unknown}}\n",
		},
	} {
		for _, dry := range []bool{false, true} {
			for _, selected := range []bool{false, true} {
				mode := map[bool]string{false: "/ordinary", true: "/selected"}[selected] + map[bool]string{false: "/write", true: "/preview"}[dry]
				t.Run(tc.name+mode, func(t *testing.T) {
					options := []Option{WithClock(fixedClock("2026-10-19")), WithIDGen(func() string { return "6gmy4c0x6e31" }),
						WithTemplateSource(fakeTemplates{byKind: map[string][]domain.NamedTemplate{
							"task": {{Name: "selected", Body: "# {{title}}\nEpic {{epic}}\n100% {{unknown}}\n"}},
						}})}
					p := tc.p
					p.DryRun = dry
					var got TaskCreationReceipt
					var body string
					var err error
					if selected {
						port := &selectedEpicCreationFake{}
						svc := MustNewService(nil, append(options, WithSelectedEpicTaskCreationStore(port))...)
						got, err = svc.NewTaskForSelectedEpic(EpicMutationTarget{ID: "01-domain"}, p)
						body = port.body
					} else {
						store := &fakeStore{epics: []domain.Epic{{ID: "01-domain"}}}
						p.Epic = "01-domain"
						got, err = MustNewService(store, options...).NewTask(p)
						if len(store.createdBodies) == 1 {
							body = store.createdBodies[0]
						}
					}
					if err != nil || !reflect.DeepEqual(got.Task, tc.want) || body != tc.body {
						t.Fatalf("got task=%+v body=%q err=%v; want task=%+v body=%q", got.Task, body, err, tc.want, tc.body)
					}
				})
			}
		}
	}
}

func TestSelectedEpicCreationRejectsInvalidInputsBeforePort(t *testing.T) {
	for _, tc := range []struct {
		name   string
		target EpicMutationTarget
		edit   func(*NewTaskParams)
	}{
		{"empty target", EpicMutationTarget{}, func(*NewTaskParams) {}},
		{"different epic input", EpicMutationTarget{ID: "01-selected"}, func(p *NewTaskParams) { p.Epic = "01-other" }},
		{"create and start", EpicMutationTarget{ID: "01-selected"}, func(p *NewTaskParams) { p.Start = true }},
		{"body and template", EpicMutationTarget{ID: "01-selected"}, func(p *NewTaskParams) { p.Body, p.Template = "body", "default" }},
		{"invalid priority", EpicMutationTarget{ID: "01-selected"}, func(p *NewTaskParams) { p.Priority = "urgent" }},
		{"invalid tier", EpicMutationTarget{ID: "01-selected"}, func(p *NewTaskParams) { p.Tier = 9 }},
		{"invalid autonomy", EpicMutationTarget{ID: "01-selected"}, func(p *NewTaskParams) { p.Autonomy = 9 }},
		{"invalid description", EpicMutationTarget{ID: "01-selected"}, func(p *NewTaskParams) { p.Description = strings.Repeat("x", domain.MaxDescriptionLen+1) }},
		{"empty title", EpicMutationTarget{ID: "01-selected"}, func(p *NewTaskParams) { p.Title = "" }},
		{"missing tags", EpicMutationTarget{ID: "01-selected"}, func(p *NewTaskParams) { p.Tags = nil }},
		{"next without description", EpicMutationTarget{ID: "01-selected"}, func(p *NewTaskParams) { p.Next = true }},
	} {
		for _, dry := range []bool{false, true} {
			t.Run(tc.name+map[bool]string{false: "/write", true: "/preview"}[dry], func(t *testing.T) {
				port := &selectedEpicCreationFake{}
				svc := MustNewService(nil, WithSelectedEpicTaskCreationStore(port))
				p := NewTaskParams{Title: "Created", Tags: []string{"test"}, DryRun: dry}
				tc.edit(&p)
				if _, err := svc.NewTaskForSelectedEpic(tc.target, p); !errors.Is(err, domain.ErrValidation) || port.calls != 0 {
					t.Fatalf("invalid request reached port: calls=%d err=%v", port.calls, err)
				}
			})
		}
	}
}

func TestSelectedEpicCreationDoesNotRetryOrDiscardFailureReceipts(t *testing.T) {
	cleanup := &CreateFinalizationError{Cause: domain.ErrConflict}
	for _, failure := range []error{domain.ErrNotFound, domain.ErrAmbiguous, domain.ErrConflict, ErrReadOnlyPersistence, cleanup} {
		for _, committed := range []bool{false, true} {
			port := &selectedEpicCreationFake{err: failure, committed: committed}
			minted, sleeps := 0, 0
			svc := MustNewService(nil, WithSelectedEpicTaskCreationStore(port), WithRetry(3, func(int) { sleeps++ }), WithIDGen(func() string {
				minted++
				return "6gmy4c0x6e31"
			}))
			got, err := svc.NewTaskForSelectedEpic(EpicMutationTarget{ID: "selected"}, NewTaskParams{Title: "Created", Tags: []string{"test"}})
			if err != failure || got.Committed != committed || got.Task.ID != "6gmy4c0x6e31" ||
				port.calls != 1 || minted != 1 || sleeps != 0 {
				t.Fatalf("failure receipt/retry drift: got=%+v err=%v calls=%d minted=%d sleeps=%d", got, err, port.calls, minted, sleeps)
			}
		}
	}
}

type aggregateSelectedEpicCreation struct {
	nopStore
	port *selectedEpicCreationFake
}

func (s *aggregateSelectedEpicCreation) CreateTaskForSelectedEpic(target EpicMutationTarget, task domain.Task, body string, dry bool) (TaskCreationReceipt, error) {
	return s.port.CreateTaskForSelectedEpic(target, task, body, dry)
}

func TestSelectedEpicCreationCapabilityDiscoveryAndOverride(t *testing.T) {
	auto, override := &selectedEpicCreationFake{}, &selectedEpicCreationFake{}
	store := &aggregateSelectedEpicCreation{port: auto}
	for _, options := range [][]Option{nil, {WithSelectedEpicTaskCreationStore(override)}} {
		svc := MustNewService(store, options...)
		if _, err := svc.NewTaskForSelectedEpic(EpicMutationTarget{ID: "selected"}, NewTaskParams{Title: "Created", Tags: []string{"test"}}); err != nil {
			t.Fatal(err)
		}
	}
	if auto.calls != 1 || override.calls != 1 {
		t.Fatalf("capability selection: auto=%d explicit=%d", auto.calls, override.calls)
	}
	var absent *selectedEpicCreationFake
	svc := MustNewService(&fakeStore{}, WithSelectedEpicTaskCreationStore(absent))
	if _, err := svc.NewTaskForSelectedEpic(EpicMutationTarget{ID: "selected"}, NewTaskParams{Title: "Created", Tags: []string{"test"}}); !errors.Is(err, domain.ErrValidation) || !strings.Contains(err.Error(), "unavailable") {
		t.Fatalf("missing capability did not fail explicitly: %v", err)
	}
}
