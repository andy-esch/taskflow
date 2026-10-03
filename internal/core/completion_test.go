package core

import (
	"errors"
	"reflect"
	"slices"
	"testing"

	"github.com/andy-esch/taskflow/internal/domain"
)

type completionStub struct {
	candidates []CompletionCandidate
	err        error
	calls      int
	kind       EntityKind
	withState  bool
}

func (*completionStub) SourceSetID() SourceSetID { return testSourceSetID }

func (f *completionStub) ReadCompletionCandidates(kind EntityKind, withState bool) ([]CompletionCandidate, error) {
	f.calls++
	f.kind, f.withState = kind, withState
	return f.candidates, f.err
}

func TestCompleteEntitiesPreservesResolutionAndMalformedCandidates(t *testing.T) {
	candidates := []CompletionCandidate{
		{ID: "6fjangd7kva1", Slug: "alpha", Reference: "6fjangd7kva1", SlugIsReference: true, State: "next-up"},
		{ID: "6fjangd7kvb2", Slug: "beta", Reference: "6fjangd7kvb2", SlugIsReference: true, State: "in-progress"},
		// Defend against a source accidentally certifying duplicate exact labels.
		{ID: "6fjangd7kvc3", Slug: "dup", Reference: "6fjangd7kvc3", SlugIsReference: true, State: "in-progress"},
		{ID: "6fjangd7kvd4", Slug: "dup", Reference: "6fjangd7kvd4", SlugIsReference: true}, // unreadable sibling
	}
	for _, test := range []struct {
		name    string
		request CompletionRequest
		want    []string
	}{
		{"unique slug", CompletionRequest{Prefix: "al"}, []string{"alpha"}},
		{"ID prefix", CompletionRequest{Prefix: "6fjangd7kvb"}, []string{"6fjangd7kvb2"}},
		{"duplicate alias", CompletionRequest{Prefix: "dup"}, []string{"6fjangd7kvc3", "6fjangd7kvd4"}},
		{"typed slug", CompletionRequest{Prefix: "al", Args: []string{"alpha"}}, nil},
		{"typed reference", CompletionRequest{Prefix: "al", Args: []string{"6fjangd7kva1"}}, nil},
		{"typed ID", CompletionRequest{Prefix: "al", Args: []string{"6fjangd7kva1"}}, nil},
		{"exclude only observed record", CompletionRequest{Prefix: "dup", ExcludeState: "in-progress"}, []string{"6fjangd7kvd4"}},
	} {
		t.Run(test.name, func(t *testing.T) {
			fake := &completionStub{candidates: append([]CompletionCandidate(nil), candidates...)}
			svc := MustNewService(nil, WithCompletionSource(fake))
			test.request.Kind = EntityTask
			got, err := svc.CompleteEntities(test.request)
			if err != nil || !slices.Equal(got, test.want) || fake.calls != 1 || fake.kind != EntityTask || fake.withState != (test.request.ExcludeState != "") {
				t.Fatalf("completion = %v, %v; source=%+v", got, err, fake)
			}
			if !reflect.DeepEqual(candidates, fake.candidates) {
				t.Fatal("completion mutated adapter candidates")
			}
		})
	}
}

func TestCompleteEntitiesAcceptsOpaqueSelectorsForEveryKind(t *testing.T) {
	for _, kind := range []EntityKind{EntityTask, EntityThread, EntityEpic, EntityAudit, EntityResearch} {
		t.Run(string(kind), func(t *testing.T) {
			fake := &completionStub{candidates: []CompletionCandidate{
				{ID: "db-key:alpha", Slug: "alpha", Reference: "record:alpha", SlugIsReference: true},
				{ID: "db-key:first", Slug: "dup", Reference: "record:first"},
				{ID: "db-key:second", Slug: "dup", Reference: "record:second"},
			}}
			svc := MustNewService(nil, WithCompletionSource(fake))
			got, err := svc.CompleteEntities(CompletionRequest{Kind: kind})
			if err != nil || !slices.Equal(got, []string{"alpha", "record:first", "record:second"}) || fake.calls != 1 || fake.kind != kind || fake.withState {
				t.Fatalf("opaque completion = %v, %v; source=%+v", got, err, fake)
			}
		})
	}
}

func TestCompleteEntitiesSelectedRecordDoesNotHideSibling(t *testing.T) {
	for _, arg := range []string{"first", "record:first"} {
		t.Run(arg, func(t *testing.T) {
			fake := &completionStub{candidates: []CompletionCandidate{
				{ID: "first", Slug: "dup", Reference: "record:first"},
				{ID: "second", Slug: "dup", Reference: "record:second"},
			}}
			got, err := MustNewService(nil, WithCompletionSource(fake)).CompleteEntities(CompletionRequest{
				Kind: EntityTask, Prefix: "du", Args: []string{arg},
			})
			if err != nil || !slices.Equal(got, []string{"record:second"}) || fake.calls != 1 {
				t.Fatalf("typed=%q suggestions=%v err=%v calls=%d", arg, got, err, fake.calls)
			}
		})
	}
}

func TestCompleteEntitiesDoesNotInferAliasSafety(t *testing.T) {
	fake := &completionStub{candidates: []CompletionCandidate{
		{ID: "db-key:alpha", Slug: "human-label", Reference: "record:alpha"},
	}}
	got, err := MustNewService(nil, WithCompletionSource(fake)).CompleteEntities(CompletionRequest{Kind: EntityTask, Prefix: "human"})
	if err != nil || !slices.Equal(got, []string{"record:alpha"}) {
		t.Fatalf("uncertified alias was emitted: suggestions=%v err=%v", got, err)
	}
}

func TestCompleteEntitiesDoesNotFallbackOnMissingOrFailedCapability(t *testing.T) {
	if _, err := MustNewService(nil).CompleteEntities(CompletionRequest{Kind: EntityTask}); !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("missing capability = %v", err)
	}
	sentinel := errors.New("completion snapshot unavailable")
	fake := &completionStub{candidates: []CompletionCandidate{{ID: "key", Slug: "partial", Reference: "key"}}, err: sentinel}
	svc := MustNewService(nil, WithCompletionSource(fake))
	if got, err := svc.CompleteEntities(CompletionRequest{Kind: EntityTask}); !errors.Is(err, sentinel) || len(got) != 0 || fake.calls != 1 {
		t.Fatalf("failed capability = %v, %v; calls=%d", got, err, fake.calls)
	}
	if _, err := svc.CompleteEntities(CompletionRequest{Kind: "invalid"}); !errors.Is(err, domain.ErrValidation) || fake.calls != 1 {
		t.Fatalf("invalid kind reached source: err=%v calls=%d", err, fake.calls)
	}
}
