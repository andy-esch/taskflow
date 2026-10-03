package core

import (
	"errors"
	"testing"

	"github.com/andy-esch/taskflow/internal/domain"
)

// This complete semantic fake deliberately has no Resolve*Path methods. An
// adapter may support every ordinary use case without local navigation.
var _ Store = nopStore{}

type entityPathFake struct {
	testSourceSetProvider
	path  string
	calls []string
}

func (f *entityPathFake) ResolveTaskPath(ref string) (string, error) {
	f.calls = append(f.calls, "task:"+ref)
	return f.path, nil
}
func (f *entityPathFake) ResolveEpicPath(ref string) (string, error) {
	f.calls = append(f.calls, "epic:"+ref)
	return f.path, nil
}
func (f *entityPathFake) ResolveAuditPath(ref string) (string, error) {
	f.calls = append(f.calls, "audit:"+ref)
	return f.path, nil
}
func (f *entityPathFake) ResolveResearchPath(ref string) (string, error) {
	f.calls = append(f.calls, "research:"+ref)
	return f.path, nil
}

type aggregateEntityPaths struct {
	nopStore
	*entityPathFake
}

func (*aggregateEntityPaths) SourceSetID() SourceSetID { return testSourceSetID }

func TestEntityPathCapabilitiesAreOptionalAndEntitySpecific(t *testing.T) {
	pathless := MustNewService(nopStore{})
	for _, kind := range []EntityKind{EntityTask, EntityEpic, EntityAudit, EntityResearch, EntityThread} {
		if pathless.HasLocalPath(kind) {
			t.Fatalf("pathless semantic adapter advertises %s local navigation", kind)
		}
	}
	for _, path := range []func(string) (string, error){
		pathless.TaskPath, pathless.EpicPath, pathless.AuditPath, pathless.ResearchPath,
	} {
		if _, err := path("any"); !errors.Is(err, domain.ErrValidation) {
			t.Fatalf("pathless semantic adapter unexpectedly resolved a local path: %v", err)
		}
	}

	paths := &entityPathFake{path: "/planning/local.md"}
	svc := MustNewService(nopStore{}, WithTaskPathSource(paths), WithEpicPathSource(paths),
		WithAuditPathSource(paths), WithResearchPathSource(paths))
	for _, kind := range []EntityKind{EntityTask, EntityEpic, EntityAudit, EntityResearch} {
		if !svc.HasLocalPath(kind) {
			t.Fatalf("explicit %s local path capability was not advertised", kind)
		}
	}
	for _, tc := range []struct {
		resolve func(string) (string, error)
		want    string
	}{
		{svc.TaskPath, "task:one"}, {svc.EpicPath, "epic:one"},
		{svc.AuditPath, "audit:one"}, {svc.ResearchPath, "research:one"},
	} {
		got, err := tc.resolve("one")
		if err != nil || got != paths.path || paths.calls[len(paths.calls)-1] != tc.want {
			t.Fatalf("%s path = %q, %v; calls = %v", tc.want, got, err, paths.calls)
		}
	}
	paths.path = ""
	if _, err := svc.TaskPath("one"); !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("empty resolver result was accepted as a local path: %v", err)
	}
}

func TestEntityPathDefaultsDetachOnExplicitReadReplacement(t *testing.T) {
	aggregatePaths := &entityPathFake{path: "/planning/aggregate.md"}
	aggregate := &aggregateEntityPaths{entityPathFake: aggregatePaths}
	if got, err := MustNewService(aggregate).TaskPath("one"); err != nil || got != aggregatePaths.path {
		t.Fatalf("aggregate path default = %q, %v", got, err)
	}
	reads := &sourceSetProbe{id: testSourceSetID}
	if _, err := MustNewService(aggregate, WithTaskGraphSource(reads)).TaskPath("one"); !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("task graph override borrowed aggregate path: %v", err)
	}
	if _, err := MustNewService(aggregate, WithAuditSnapshotSource(reads)).AuditPath("one"); !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("audit snapshot override borrowed aggregate path: %v", err)
	}
	if _, err := MustNewService(aggregate, WithLintSource(reads)).AuditPath("one"); !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("lint-source audit snapshot override borrowed aggregate path: %v", err)
	}
	for _, opts := range [][]Option{
		{WithTaskPathSource(aggregatePaths), WithTaskGraphSource(reads)},
		{WithTaskGraphSource(reads), WithTaskPathSource(aggregatePaths)},
	} {
		got, err := MustNewService(aggregate, opts...).TaskPath("one")
		if err != nil || got != aggregatePaths.path {
			t.Fatalf("explicit task path lost across option ordering: %q, %v", got, err)
		}
	}
	for _, opts := range [][]Option{
		{WithAuditPathSource(aggregatePaths), WithAuditSnapshotSource(reads)},
		{WithAuditSnapshotSource(reads), WithAuditPathSource(aggregatePaths)},
	} {
		got, err := MustNewService(aggregate, opts...).AuditPath("one")
		if err != nil || got != aggregatePaths.path {
			t.Fatalf("explicit audit path lost across option ordering: %q, %v", got, err)
		}
	}
}

func TestEntityPathCapabilitiesIgnoreTypedNilAndRejectForeignSourceSets(t *testing.T) {
	var typedNil *entityPathFake
	svc := MustNewService(nopStore{}, WithTaskPathSource(typedNil), WithEpicPathSource(typedNil),
		WithAuditPathSource(typedNil), WithResearchPathSource(typedNil))
	if _, err := svc.TaskPath("one"); !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("typed-nil task path became available: %v", err)
	}
	for _, kind := range []EntityKind{EntityTask, EntityEpic, EntityAudit, EntityResearch} {
		if svc.HasLocalPath(kind) {
			t.Fatalf("typed-nil %s path was advertised", kind)
		}
	}
	foreign := &entityPathFake{testSourceSetProvider: testSourceSetProvider{sourceSet: NewSourceSetID()}}
	for _, opt := range []Option{WithTaskPathSource(foreign), WithEpicPathSource(foreign),
		WithAuditPathSource(foreign), WithResearchPathSource(foreign)} {
		if _, err := NewService(nopStore{}, opt); !errors.Is(err, ErrIncompatibleCapabilities) {
			t.Fatalf("foreign path source was paired with aggregate reads: %v", err)
		}
	}
}
