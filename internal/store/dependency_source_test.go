package store

import (
	"path/filepath"
	"slices"
	"testing"

	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/testutil"
)

func TestReadTaskGraphPopulatesLosslessDependencySourceProjection(t *testing.T) {
	root := t.TempDir()
	prerequisiteID := testutil.TaskID("source-adapter-prerequisite")
	writeGraphMutationTask(t, root, "source-adapter-prerequisite", domain.StatusCompleted, nil, "")
	ownerPath := writeGraphMutationTask(t, root, "source-adapter-owner", domain.StatusReadyToStart,
		[]string{prerequisiteID, prerequisiteID},
		"blocked_by: [source-adapter-prerequisite]\n"+
			"dependencies: ["+prerequisiteID+"]\n"+
			"blocks: []\n")
	unreadableID := testutil.TaskID("source-adapter-unreadable")
	unreadablePath := filepath.Join(root, domain.TasksDir, unreadableID+"-source-adapter-unreadable.md")
	testutil.Write(t, unreadablePath, "---\nid: [unterminated\n---\n# Broken\n")

	read, err := NewFS(root).ReadTaskGraph()
	if err != nil || len(read.Problems) != 1 || read.Problems[0].SourceVersion == "" {
		t.Fatalf("read=%+v err=%v", read, err)
	}
	graph := core.NewTaskGraphRead(read)
	records, err := graph.SourceRecords()
	if err != nil {
		t.Fatal(err)
	}
	var owner core.TaskGraphSourceRecord
	for _, record := range records {
		if record.Source.Location == ownerPath {
			owner = record
		}
	}
	if owner.Source.TaskID == "" {
		t.Fatalf("owner source record missing from %+v", records)
	}
	wantFields := []core.TaskGraphSourceField{
		{Field: core.TaskDependencyDependsOn, Values: []string{prerequisiteID, prerequisiteID}},
		{Field: core.TaskDependencyBlockedBy, Values: []string{"source-adapter-prerequisite"}},
		{Field: core.TaskDependencyDependencies, Values: []string{prerequisiteID}},
		{Field: core.TaskDependencyBlocks, Values: []string{}},
	}
	if !slices.EqualFunc(owner.Fields, wantFields, func(left, right core.TaskGraphSourceField) bool {
		return left.Field == right.Field && slices.Equal(left.Values, right.Values)
	}) {
		t.Fatalf("source fields=%+v, want %+v", owner.Fields, wantFields)
	}
	if owner.Fields[3].Values == nil {
		t.Fatal("present-but-empty legacy field used a nil values slice")
	}

	declarations, err := graph.SourceDeclarations()
	if err != nil {
		t.Fatal(err)
	}
	duplicateOccurrences := make([]int, 0, 2)
	for _, declaration := range declarations {
		if declaration.Source.Location == ownerPath && declaration.Field == core.TaskDependencyDependsOn && declaration.Value == prerequisiteID {
			duplicateOccurrences = append(duplicateOccurrences, declaration.Occurrence)
		}
	}
	if !slices.Equal(duplicateOccurrences, []int{0, 1}) {
		t.Fatalf("canonical duplicate occurrences = %v in %+v", duplicateOccurrences, declarations)
	}
}
