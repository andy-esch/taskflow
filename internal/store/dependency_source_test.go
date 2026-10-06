package store

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/testutil"
)

func TestFilesystemTaskGraphKeepsGuardedEvidenceOutsideSemanticTask(t *testing.T) {
	root := t.TempDir()
	path := writeGraphMutationTask(t, root, "guarded-source-evidence", domain.StatusReadyToStart, nil, "")
	fs := testutil.Must(NewFS(root, core.ReadOnlyMutations()))
	beforeRead, err := fs.ReadTaskGraph()
	if err != nil || len(beforeRead.GuardedRecords) != 1 {
		t.Fatalf("guarded read = %+v, %v", beforeRead, err)
	}
	guarded := beforeRead.GuardedRecords[0]
	if guarded.SourceVersion == "" || guarded.LocalPath != path {
		t.Fatalf("source evidence escaped guarded wrapper: %+v", guarded)
	}
	before := core.NewTaskGraphRead(beforeRead)
	source, ok := before.TaskSource(guarded.Record.Source.ID)
	if !ok || source.LocalPath != path {
		t.Fatalf("graph source = %+v, %v", source, ok)
	}
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, append(content, []byte("\nA non-graph edit.\n")...), 0o644); err != nil {
		t.Fatal(err)
	}
	afterRead, err := fs.ReadTaskGraph()
	if err != nil {
		t.Fatal(err)
	}
	if afterRead.GuardedRecords[0].SourceVersion == guarded.SourceVersion || before.SameSourceSnapshot(core.NewTaskGraphRead(afterRead)) {
		t.Fatal("a body-only source edit escaped whole-snapshot comparison")
	}
	ordinary, err := fs.ReadTasks()
	if err != nil || len(ordinary.Records) != 1 {
		t.Fatalf("ordinary task read leaked revision evidence: %+v, %v", ordinary, err)
	}
	encoded, err := json.Marshal(ordinary)
	if err != nil || strings.Contains(string(encoded), guarded.SourceVersion) {
		t.Fatalf("ordinary task JSON leaked guarded revision: %s, %v", encoded, err)
	}
}

func TestFilesystemTaskReadsKeepSourceIdentityDespiteFrontmatterDrift(t *testing.T) {
	root := t.TempDir()
	path := writeGraphMutationTask(t, root, "drifted-task-source", domain.StatusReadyToStart, nil, "")
	const declaredID = "6g0000000002"
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	filenameID := testutil.TaskID("drifted-task-source")
	from := "id: " + filenameID
	if !strings.Contains(string(content), from) {
		t.Fatalf("task fixture does not contain %q", from)
	}
	content = []byte(strings.Replace(string(content), from, "id: "+declaredID, 1))
	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatal(err)
	}
	fs := testutil.Must(NewFS(root, core.ReadOnlyMutations()))
	graph, err := fs.ReadTaskGraph()
	if err != nil || len(graph.GuardedRecords) != 1 {
		t.Fatalf("graph read = %+v, err=%v", graph, err)
	}
	guarded := graph.GuardedRecords[0]
	if guarded.Record.Value.ID != declaredID || guarded.Record.Source.ID != filenameID ||
		guarded.Record.Source.Location != path || guarded.LocalPath != path || guarded.SourceVersion == "" {
		t.Fatalf("graph read confused declaration and source: %+v", guarded)
	}
	for _, read := range []func() (core.LoadedRecord[core.TaskWithBody], error){
		func() (core.LoadedRecord[core.TaskWithBody], error) { return fs.ReadTask(filenameID) },
		func() (core.LoadedRecord[core.TaskWithBody], error) {
			records, _, err := fs.ReadLintTasks()
			if err != nil || len(records) != 1 {
				return core.LoadedRecord[core.TaskWithBody]{}, fmt.Errorf("lint records=%+v, err=%v", records, err)
			}
			return records[0], nil
		},
	} {
		record, err := read()
		if err != nil || record.Source != guarded.Record.Source || record.Value.Task.ID != declaredID {
			t.Fatalf("ordinary task read confused declaration and source: %+v, err=%v", record, err)
		}
	}
}

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

	read, err := testutil.Must(NewFS(root, core.ReadOnlyMutations())).ReadTaskGraph()
	if err != nil || len(read.Problems) != 1 || read.Problems[0].SourceVersion == "" {
		t.Fatalf("read=%+v err=%v", read, err)
	}
	if read.Problems[0].TaskID != unreadableID || read.Problems[0].TaskSlug != "source-adapter-unreadable" ||
		read.Problems[0].Path != unreadablePath {
		t.Fatalf("filesystem adapter did not recover unreadable record identity: %+v", read.Problems[0])
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
