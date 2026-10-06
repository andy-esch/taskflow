package store

import (
	"errors"
	"testing"

	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/testutil"
)

func TestFilesystemSourceSetIsInstanceOwnedNotDerivedFromRootOrRecordID(t *testing.T) {
	firstRepo := testutil.NewRepo(t).Task("ready-to-start", "same.md", "---\ndescription: first corpus\n---\n# same\n")
	secondRepo := testutil.NewRepo(t).Task("ready-to-start", "same.md", "---\ndescription: second corpus\n---\n# same\n")
	first := testutil.Must(NewFS(firstRepo.Root, core.ReadOnlyMutations()))
	second := testutil.Must(NewFS(secondRepo.Root, core.ReadOnlyMutations()))
	sameRootNewInstance := testutil.Must(NewFS(firstRepo.Root, core.ReadOnlyMutations()))
	if first.SourceSetID().IsZero() || first.SourceSetID() == second.SourceSetID() ||
		first.SourceSetID() == sameRootNewInstance.SourceSetID() {
		t.Fatal("filesystem source-set ID must be unique per adapter instance, not inferred from root")
	}
	firstGraph, err := first.ReadTaskGraph()
	if err != nil {
		t.Fatal(err)
	}
	sameRootGraph, err := sameRootNewInstance.ReadTaskGraph()
	if err != nil {
		t.Fatal(err)
	}
	if len(firstGraph.GuardedRecords) != 1 || len(sameRootGraph.GuardedRecords) != 1 ||
		firstGraph.GuardedRecords[0].Record.Source != sameRootGraph.GuardedRecords[0].Record.Source ||
		firstGraph.GuardedRecords[0].SourceVersion == "" ||
		firstGraph.GuardedRecords[0].SourceVersion != sameRootGraph.GuardedRecords[0].SourceVersion {
		t.Fatal("same-root fixtures must share record identity, location, and revision evidence")
	}
	if svc, err := core.NewService(first, core.WithTaskGraphSource(sameRootNewInstance)); svc != nil ||
		!errors.Is(err, core.ErrIncompatibleCapabilities) {
		t.Fatalf("matching record/location/revision cannot substitute for a source-set ID: %v, %v", svc, err)
	}
	firstRead, err := first.ReadTasks()
	if err != nil {
		t.Fatal(err)
	}
	secondRead, err := second.ReadTasks()
	if err != nil {
		t.Fatal(err)
	}
	if len(firstRead.Records) != 1 || len(secondRead.Records) != 1 ||
		firstRead.Records[0].Source.ID != secondRead.Records[0].Source.ID {
		t.Fatalf("fixture must share a record ID: %+v / %+v", firstRead, secondRead)
	}
	if svc, err := core.NewService(first, core.WithTaskGraphSource(second)); svc != nil ||
		!errors.Is(err, core.ErrIncompatibleCapabilities) {
		t.Fatalf("cross-corpus pairing = %v, %v", svc, err)
	}
	if svc, err := core.NewService(first, core.WithThreadPathSource(sameRootNewInstance)); svc != nil ||
		!errors.Is(err, core.ErrIncompatibleCapabilities) {
		t.Fatalf("same-root but unbound instance pairing = %v, %v", svc, err)
	}
	for _, opt := range []core.Option{
		core.WithTaskPathSource(sameRootNewInstance), core.WithEpicPathSource(sameRootNewInstance),
		core.WithAuditPathSource(sameRootNewInstance), core.WithResearchPathSource(sameRootNewInstance),
	} {
		if svc, err := core.NewService(first, opt); svc != nil || !errors.Is(err, core.ErrIncompatibleCapabilities) {
			t.Fatalf("same-root entity path without matching source set = %v, %v", svc, err)
		}
	}
	bound := testutil.Must(NewFS(firstRepo.Root, core.ReadOnlyMutations(), WithSourceSetID(first.SourceSetID())))
	if svc, err := core.NewService(first, core.WithThreadPathSource(bound)); err != nil || svc == nil {
		t.Fatalf("explicitly bound same-corpus pairing = %v, %v", svc, err)
	}
	if svc, err := core.NewService(testutil.Must(NewFS(firstRepo.Root, core.ReadOnlyMutations(), WithSourceSetID(core.SourceSetID{})))); svc != nil ||
		!errors.Is(err, core.ErrIncompatibleCapabilities) {
		t.Fatalf("empty explicitly supplied token = %v, %v", svc, err)
	}
}
