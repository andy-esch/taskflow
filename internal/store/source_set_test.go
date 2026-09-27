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
	first := NewFS(firstRepo.Root)
	second := NewFS(secondRepo.Root)
	sameRootNewInstance := NewFS(firstRepo.Root)
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
	if len(firstGraph.Records) != 1 || len(sameRootGraph.Records) != 1 ||
		firstGraph.Records[0].Source != sameRootGraph.Records[0].Source ||
		firstGraph.Records[0].Value.SourceVersion != sameRootGraph.Records[0].Value.SourceVersion {
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
	bound := NewFS(firstRepo.Root, WithSourceSetID(first.SourceSetID()))
	if svc, err := core.NewService(first, core.WithThreadPathSource(bound)); err != nil || svc == nil {
		t.Fatalf("explicitly bound same-corpus pairing = %v, %v", svc, err)
	}
	if svc, err := core.NewService(NewFS(firstRepo.Root, WithSourceSetID(core.SourceSetID{}))); svc != nil ||
		!errors.Is(err, core.ErrIncompatibleCapabilities) {
		t.Fatalf("empty explicitly supplied token = %v, %v", svc, err)
	}
}
