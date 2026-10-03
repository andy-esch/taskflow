package core

import (
	"errors"
	"strings"
	"testing"

	"github.com/andy-esch/taskflow/internal/domain"
)

// These embedded ports are deliberately never invoked: this probe tests the
// construction contract for every independently injectable capability.
type sourceSetProbe struct {
	id SourceSetID
	TaskGraphSource
	LintSource
	AuditSnapshotSource
	TaskGraphMutationStore
	TaskGraphRepairStore
	TaskLifecycleMutationStore
	ThreadStore
	ThreadPathSource
	TaskPathSource
	EpicPathSource
	AuditPathSource
	ResearchPathSource
	ThreadCreationMutationStore
	ThreadMutationStore
	ThreadApplyMutationStore
	Fixer
	Linter
	CompletionSource
}

func (p *sourceSetProbe) SourceSetID() SourceSetID { return p.id }

func (p *sourceSetProbe) ReadAuditSnapshot(string) (AuditSnapshot, error) {
	panic("construction probe must never be read")
}

type noSourceSetGraph struct{ TaskGraphSource }
type noSourceSetMutation struct{ TaskGraphMutationStore }
type noSourceSetCompletion struct{ CompletionSource }
type noSourceSetFixer struct{ Fixer }
type noSourceSetLinter struct{ Linter }

type unstableSourceSetGraph struct{ TaskGraphSource }

func (unstableSourceSetGraph) SourceSetID() SourceSetID { return NewSourceSetID() }

func TestSourceSetIDIsOpaqueAndPerAdapterInstance(t *testing.T) {
	first, second := NewSourceSetID(), NewSourceSetID()
	if first.IsZero() || second.IsZero() || first == second {
		t.Fatalf("new source-set IDs = %v, %v", first, second)
	}
	if !(SourceSetID{}).IsZero() {
		t.Fatal("zero source-set ID must be invalid")
	}
}

func TestNewServiceRejectsEveryMismatchedSplitCapability(t *testing.T) {
	foreign := &sourceSetProbe{id: NewSourceSetID()}
	for _, tc := range []struct {
		name string
		opt  Option
	}{
		{"task graph read", WithTaskGraphSource(foreign)},
		{"lint read", WithLintSource(foreign)},
		{"audit snapshot", WithAuditSnapshotSource(foreign)},
		{"graph mutation", WithTaskGraphMutationStore(foreign)},
		{"graph repair", WithTaskGraphRepairStore(foreign)},
		{"lifecycle mutation", WithTaskLifecycleMutationStore(foreign)},
		{"Thread read", WithThreadStore(foreign)},
		{"Thread path", WithThreadPathSource(foreign)},
		{"task path", WithTaskPathSource(foreign)},
		{"epic path", WithEpicPathSource(foreign)},
		{"audit path", WithAuditPathSource(foreign)},
		{"research path", WithResearchPathSource(foreign)},
		{"Thread creation", WithThreadCreationMutationStore(foreign)},
		{"Thread mutation", WithThreadMutationStore(foreign)},
		{"Thread apply", WithThreadApplyMutationStore(foreign)},
		{"frontmatter repair", WithFrontmatterRepairer(foreign)},
		{"body link checks", WithBodyLinkSource(foreign)},
		{"completion", WithCompletionSource(foreign)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			svc, err := NewService(&fakeStore{}, tc.opt)
			if svc != nil || !errors.Is(err, ErrIncompatibleCapabilities) || !errors.Is(err, domain.ErrValidation) ||
				!strings.Contains(err.Error(), "different source sets") {
				t.Fatalf("service = %v, error = %v; want typed mismatch before use", svc, err)
			}
		})
	}
}

func TestNewServiceRejectsMissingOrEmptySourceSetIdentity(t *testing.T) {
	for _, tc := range []struct {
		name string
		opt  Option
		want string
	}{
		{"missing provider", WithTaskGraphSource(&noSourceSetGraph{}), "does not publish"},
		{"guarded mutation without provider", WithTaskGraphMutationStore(&noSourceSetMutation{}), "does not publish"},
		{"completion without provider", WithCompletionSource(&noSourceSetCompletion{}), "does not publish"},
		{"frontmatter repair without provider", WithFrontmatterRepairer(&noSourceSetFixer{}), "does not publish"},
		{"link checks without provider", WithBodyLinkSource(&noSourceSetLinter{}), "does not publish"},
		{"completion without witness", WithCompletionSource(&sourceSetProbe{}), "empty source-set"},
		{"frontmatter repair without witness", WithFrontmatterRepairer(&sourceSetProbe{}), "empty source-set"},
		{"link checks without witness", WithBodyLinkSource(&sourceSetProbe{}), "empty source-set"},
		{"empty provider", WithTaskGraphSource(&sourceSetProbe{}), "empty source-set"},
		{"unstable provider", WithTaskGraphSource(&unstableSourceSetGraph{}), "unstable source-set"},
		{"guarded mutation without witness", WithTaskGraphMutationStore(&sourceSetProbe{}), "empty source-set"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			svc, err := NewService(nil, tc.opt)
			if svc != nil || !errors.Is(err, ErrIncompatibleCapabilities) || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("service = %v, error = %v; want %q", svc, err, tc.want)
			}
		})
	}
}

func TestNewServiceRejectsAggregateStoreWithoutSourceSetIdentity(t *testing.T) {
	store := &struct{ Store }{}
	svc, err := NewService(store)
	if svc != nil || !errors.Is(err, ErrIncompatibleCapabilities) ||
		!strings.Contains(err.Error(), "aggregate store does not publish") {
		t.Fatalf("aggregate without source-set identity = %v, %v", svc, err)
	}
}

func TestNewServicePreservesTypedNilAndOrderIndependentThreadPairing(t *testing.T) {
	var typedNil *threadReadFake
	if _, err := NewService(&fakeStore{}, WithThreadStore(typedNil)); err != nil {
		t.Fatalf("typed-nil option should remain absent: %v", err)
	}
	compatible := &sourceSetProbe{id: testSourceSetID}
	for _, opts := range [][]Option{
		{WithThreadStore(compatible), WithThreadPathSource(compatible)},
		{WithThreadPathSource(compatible), WithThreadStore(compatible)},
	} {
		svc, err := NewService(&fakeStore{}, opts...)
		if err != nil || svc.threadPaths != compatible || svc.threads != compatible {
			t.Fatalf("compatible explicit pairing = %v, %v", svc, err)
		}
	}
	foreign := &sourceSetProbe{id: NewSourceSetID()}
	for _, opts := range [][]Option{
		{WithThreadStore(compatible), WithThreadPathSource(foreign)},
		{WithThreadPathSource(foreign), WithThreadStore(compatible)},
	} {
		if svc, err := NewService(&fakeStore{}, opts...); svc != nil || !errors.Is(err, ErrIncompatibleCapabilities) {
			t.Fatalf("incompatible explicit pairing = %v, %v", svc, err)
		}
	}
}

func TestNewServiceRetainsImplicitThreadPathDetach(t *testing.T) {
	aggregate := &aggregateThreadPathFake{fakeStore: &fakeStore{}}
	compatible := &sourceSetProbe{id: testSourceSetID}
	svc, err := NewService(aggregate, WithThreadStore(compatible))
	if err != nil || svc.threadPaths != nil {
		t.Fatalf("implicit aggregate paths after explicit Thread reads = %v, %v", svc, err)
	}
}

func TestNewServiceAcceptsPathlessReadOnlySource(t *testing.T) {
	svc, err := NewService(nil, WithTaskGraphSource(&taskGraphReadFake{}))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Board(); err != nil {
		t.Fatalf("pathless read-only board: %v", err)
	}
}
