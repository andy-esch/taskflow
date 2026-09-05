package core

import (
	"fmt"
	"slices"
	"sort"
	"strings"

	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/id"
)

// TaskDependencyField is the complete task-owned dependency vocabulary. Keeping
// it typed prevents repair-oriented code from treating arbitrary frontmatter as
// graph state.
type TaskDependencyField string

const (
	TaskDependencyDependsOn    TaskDependencyField = "depends_on"
	TaskDependencyBlockedBy    TaskDependencyField = "blocked_by"
	TaskDependencyDependencies TaskDependencyField = "dependencies"
	TaskDependencyBlocks       TaskDependencyField = "blocks"
)

// TaskGraphSourceRef identifies the readable source record that owns a
// declaration. Location is optional adapter context; duplicate IDs remain
// distinct when an adapter can provide distinct locations.
type TaskGraphSourceRef struct {
	TaskID   string
	TaskSlug string
	Location string
}

// TaskGraphSourceField is one present graph-owned field and its raw values in
// source order. Present-but-empty legacy fields therefore remain observable.
type TaskGraphSourceField struct {
	Field  TaskDependencyField
	Values []string
}

// TaskGraphSourceRecord is the adapter-neutral graph-owned portion of one
// readable physical task record. Source revision tokens remain private to the
// graph snapshot and are never returned through this query view.
type TaskGraphSourceRecord struct {
	Source TaskGraphSourceRef
	Fields []TaskGraphSourceField
}

// TaskGraphSourceDeclaration identifies one raw value occurrence. Occurrence is
// zero-based among equal values in the same source field, not the absolute list
// index. ProjectedEdge is zero when an invalid or unresolved value cannot name an
// edge actually admitted to the semantic graph; for legacy blocks, its owner is
// the edge prerequisite rather than the dependent.
type TaskGraphSourceDeclaration struct {
	Source           TaskGraphSourceRef
	Field            TaskDependencyField
	Value            string
	Occurrence       int
	ProjectedEdge    DependencyEdge
	HasProjectedEdge bool
}

// TaskGraphSourceEditAction is deliberately removal-only. The downstream repair
// policy decides which edits are authorized; this layer only models their exact
// source effect without offering a replacement-state escape hatch.
type TaskGraphSourceEditAction string

const (
	TaskGraphSourceDropDeclaration TaskGraphSourceEditAction = "drop"
	TaskGraphSourceDedupe          TaskGraphSourceEditAction = "dedupe"
	TaskGraphSourceDropEmptyField  TaskGraphSourceEditAction = "drop-empty-field"
)

// TaskGraphSourceEdit is the narrow materialization vocabulary for future graph
// repair. Drop targets one occurrence; dedupe ignores Occurrence and retains at
// most one matching value, making retries independent of occurrence renumbering.
type TaskGraphSourceEdit struct {
	Action     TaskGraphSourceEditAction
	Source     TaskGraphSourceRef
	Field      TaskDependencyField
	Value      string
	Occurrence int
}

// SourceRecords returns every readable physical record, including duplicate-ID
// shadows, in deterministic order. Returned fields and value slices are copies.
// It fails closed for behavior-only graphs reconstructed from representative
// task maps because those projections are not complete repository evidence.
func (g *TaskGraph) SourceRecords() ([]TaskGraphSourceRecord, error) {
	if err := g.requireCompleteSource(); err != nil {
		return nil, err
	}
	return g.sourceRecords(), nil
}

func (g *TaskGraph) sourceRecords() []TaskGraphSourceRecord {
	tasks := cloneTasks(g.sourceTasks)
	sort.SliceStable(tasks, func(i, j int) bool { return taskSourceSortKey(tasks[i]) < taskSourceSortKey(tasks[j]) })
	records := make([]TaskGraphSourceRecord, 0, len(tasks))
	for _, task := range tasks {
		records = append(records, TaskGraphSourceRecord{
			Source: sourceRefForTask(task),
			Fields: sourceFieldsForTask(task),
		})
	}
	return records
}

// SourceDeclarations returns the raw declaration multiset across every readable
// physical record. Unlike CanonicalDependencies and Prerequisites, duplicates,
// invalid tokens, dangling IDs, legacy ownership, and occurrences are retained.
func (g *TaskGraph) SourceDeclarations() ([]TaskGraphSourceDeclaration, error) {
	if err := g.requireCompleteSource(); err != nil {
		return nil, err
	}
	declarations := make([]TaskGraphSourceDeclaration, 0)
	for _, record := range g.sourceRecords() {
		for _, field := range record.Fields {
			occurrences := make(map[string]int, len(field.Values))
			for _, value := range field.Values {
				declaration := TaskGraphSourceDeclaration{
					Source: record.Source, Field: field.Field, Value: value,
					Occurrence: occurrences[value],
				}
				occurrences[value]++
				declaration.ProjectedEdge, declaration.HasProjectedEdge = g.projectedSourceEdge(declaration)
				declarations = append(declarations, declaration)
			}
		}
	}
	sort.SliceStable(declarations, func(i, j int) bool {
		return sourceDeclarationSortKey(declarations[i]) < sourceDeclarationSortKey(declarations[j])
	})
	return declarations, nil
}

// SimulateSourceEdits applies removal-only edits to the graph's complete source
// snapshot and rebuilds the semantic projection. Duplicate-ID shadows and
// unreadable records/revisions are retained; no representative task map is used.
// A last-value removal retains an empty legacy field until drop-empty-field is
// also selected. Missing targets are convergent no-ops; this simulator proves
// source effects but does not replace downstream planner/materializer receipts.
func (g *TaskGraph) SimulateSourceEdits(edits []TaskGraphSourceEdit) (*TaskGraph, error) {
	if g == nil {
		return nil, fmt.Errorf("%w: authoritative task graph is required", domain.ErrValidation)
	}
	if err := g.requireCompleteSource(); err != nil {
		return nil, err
	}
	tasks := cloneTasks(g.sourceTasks)
	groups := make(map[sourceEditGroupKey]*sourceEditGroup)
	for _, edit := range edits {
		if !edit.Field.valid() {
			return nil, fmt.Errorf("%w: unsupported graph-owned field %q", domain.ErrValidation, edit.Field)
		}
		switch edit.Action {
		case TaskGraphSourceDropDeclaration:
			if edit.Occurrence < 0 {
				return nil, fmt.Errorf("%w: declaration occurrence must be non-negative", domain.ErrValidation)
			}
		case TaskGraphSourceDedupe:
		case TaskGraphSourceDropEmptyField:
			if edit.Field == TaskDependencyDependsOn {
				return nil, fmt.Errorf("%w: empty-field removal is supported only for legacy dependency fields", domain.ErrValidation)
			}
		default:
			return nil, fmt.Errorf("%w: unsupported source edit action %q", domain.ErrValidation, edit.Action)
		}
		index, found, err := resolveSourceTask(tasks, edit.Source)
		if err != nil {
			return nil, err
		}
		if !found {
			continue // an already-absent source is an idempotent no-op
		}
		key := sourceEditGroupKey{taskIndex: index, field: edit.Field}
		group := groups[key]
		if group == nil {
			group = &sourceEditGroup{drops: make(map[string]map[int]bool), dedupe: make(map[string]bool)}
			groups[key] = group
		}
		switch edit.Action {
		case TaskGraphSourceDropDeclaration:
			if group.drops[edit.Value] == nil {
				group.drops[edit.Value] = make(map[int]bool)
			}
			group.drops[edit.Value][edit.Occurrence] = true
		case TaskGraphSourceDedupe:
			group.dedupe[edit.Value] = true
		case TaskGraphSourceDropEmptyField:
			group.dropEmpty = true
		}
	}

	keys := make([]sourceEditGroupKey, 0, len(groups))
	for key := range groups {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool {
		if keys[i].taskIndex != keys[j].taskIndex {
			return keys[i].taskIndex < keys[j].taskIndex
		}
		return dependencyFieldOrder(keys[i].field) < dependencyFieldOrder(keys[j].field)
	})
	for _, key := range keys {
		values, present := taskSourceFieldValues(tasks[key.taskIndex], key.field)
		group := groups[key]
		remaining := applySourceDeclarationEdits(values, group)
		keepPresent := present
		if group.dropEmpty && len(remaining) != 0 {
			return nil, fmt.Errorf("%w: graph-owned field %s is not empty after selected declaration edits", domain.ErrValidation, key.field)
		}
		if group.dropEmpty {
			keepPresent = false
		}
		if slices.Equal(values, remaining) && present == keepPresent {
			continue
		}
		setTaskSourceField(&tasks[key.taskIndex], key.field, remaining, keepPresent)
		tasks[key.taskIndex].SourceVersion = "" // prospective values are never authoritative CAS evidence
	}
	return NewTaskGraphRead(TaskGraphRead{
		Tasks: tasks, Problems: cloneTaskGraphLoadProblems(g.loadProblems),
	}), nil
}

func (g *TaskGraph) requireCompleteSource() error {
	if g == nil {
		return fmt.Errorf("%w: authoritative task graph is required", domain.ErrValidation)
	}
	if !g.sourceComplete {
		return fmt.Errorf("%w: task graph was reconstructed from representative records and is not a complete source snapshot", domain.ErrValidation)
	}
	return nil
}

type sourceEditGroupKey struct {
	taskIndex int
	field     TaskDependencyField
}

type sourceEditGroup struct {
	drops     map[string]map[int]bool
	dedupe    map[string]bool
	dropEmpty bool
}

func applySourceDeclarationEdits(values []string, group *sourceEditGroup) []string {
	remaining := make([]string, 0, len(values))
	occurrences := make(map[string]int, len(values))
	for _, value := range values {
		occurrence := occurrences[value]
		occurrences[value]++
		if group.drops[value][occurrence] {
			continue
		}
		remaining = append(remaining, value)
	}
	seen := make(map[string]bool, len(group.dedupe))
	return slices.DeleteFunc(remaining, func(value string) bool {
		if !group.dedupe[value] {
			return false
		}
		if seen[value] {
			return true
		}
		seen[value] = true
		return false
	})
}

func resolveSourceTask(tasks []domain.Task, source TaskGraphSourceRef) (int, bool, error) {
	if source.TaskID == "" && source.TaskSlug == "" && source.Location == "" {
		return 0, false, fmt.Errorf("%w: source identity is required", domain.ErrValidation)
	}
	matches := make([]int, 0, 1)
	for index, task := range tasks {
		candidate := sourceRefForTask(task)
		if (source.TaskID != "" && candidate.TaskID != source.TaskID) ||
			(source.TaskSlug != "" && candidate.TaskSlug != source.TaskSlug) ||
			(source.Location != "" && candidate.Location != source.Location) {
			continue
		}
		matches = append(matches, index)
	}
	switch len(matches) {
	case 0:
		return 0, false, nil
	case 1:
		return matches[0], true, nil
	default:
		return 0, false, fmt.Errorf("%w: source identity %+v matches %d readable task records", domain.ErrAmbiguous, source, len(matches))
	}
}

func sourceRefForTask(task domain.Task) TaskGraphSourceRef {
	return TaskGraphSourceRef{TaskID: canonicalTaskID(task), TaskSlug: task.Slug, Location: task.Path}
}

func sourceFieldsForTask(task domain.Task) []TaskGraphSourceField {
	fields := make([]TaskGraphSourceField, 0, 4)
	if len(task.DependsOn) > 0 {
		fields = append(fields, TaskGraphSourceField{Field: TaskDependencyDependsOn, Values: append([]string(nil), task.DependsOn...)})
	}
	legacy := []struct {
		field  TaskDependencyField
		values []string
	}{
		{field: TaskDependencyBlockedBy, values: task.LegacyBlockedBy},
		{field: TaskDependencyDependencies, values: task.LegacyDependencies},
		{field: TaskDependencyBlocks, values: task.LegacyBlocks},
	}
	for _, item := range legacy {
		if len(item.values) == 0 && !slices.Contains(task.LegacyDependencyFields, string(item.field)) {
			continue
		}
		fields = append(fields, TaskGraphSourceField{Field: item.field, Values: append([]string{}, item.values...)})
	}
	return fields
}

func taskSourceFieldValues(task domain.Task, field TaskDependencyField) ([]string, bool) {
	switch field {
	case TaskDependencyDependsOn:
		return append([]string(nil), task.DependsOn...), len(task.DependsOn) > 0
	case TaskDependencyBlockedBy:
		return append([]string(nil), task.LegacyBlockedBy...), slices.Contains(task.LegacyDependencyFields, string(field)) || len(task.LegacyBlockedBy) > 0
	case TaskDependencyDependencies:
		return append([]string(nil), task.LegacyDependencies...), slices.Contains(task.LegacyDependencyFields, string(field)) || len(task.LegacyDependencies) > 0
	case TaskDependencyBlocks:
		return append([]string(nil), task.LegacyBlocks...), slices.Contains(task.LegacyDependencyFields, string(field)) || len(task.LegacyBlocks) > 0
	default:
		return nil, false
	}
}

func setTaskSourceField(task *domain.Task, field TaskDependencyField, values []string, present bool) {
	values = append([]string(nil), values...)
	switch field {
	case TaskDependencyDependsOn:
		task.DependsOn = values
	case TaskDependencyBlockedBy:
		task.LegacyBlockedBy = values
	case TaskDependencyDependencies:
		task.LegacyDependencies = values
	case TaskDependencyBlocks:
		task.LegacyBlocks = values
	}
	if field == TaskDependencyDependsOn {
		return
	}
	task.LegacyDependencyFields = slices.DeleteFunc(task.LegacyDependencyFields, func(value string) bool {
		return value == string(field)
	})
	if present {
		task.LegacyDependencyFields = append(task.LegacyDependencyFields, string(field))
		sort.Strings(task.LegacyDependencyFields)
	}
}

func (f TaskDependencyField) valid() bool {
	switch f {
	case TaskDependencyDependsOn, TaskDependencyBlockedBy, TaskDependencyDependencies, TaskDependencyBlocks:
		return true
	default:
		return false
	}
}

func dependencyFieldOrder(field TaskDependencyField) int {
	switch field {
	case TaskDependencyDependsOn:
		return 0
	case TaskDependencyBlockedBy:
		return 1
	case TaskDependencyDependencies:
		return 2
	case TaskDependencyBlocks:
		return 3
	default:
		return 4
	}
}

func taskSourceSortKey(task domain.Task) string {
	return strings.Join([]string{
		canonicalTaskID(task), task.Slug, task.Path, task.SourceVersion,
		strings.Join(task.DependsOn, "\x00"), strings.Join(task.LegacyBlockedBy, "\x00"),
		strings.Join(task.LegacyDependencies, "\x00"), strings.Join(task.LegacyBlocks, "\x00"),
	}, "\x01")
}

func taskSourceSnapshotKey(task domain.Task) string {
	return strings.Join([]string{
		canonicalTaskID(task), task.Slug, task.Path, task.SourceVersion,
	}, "\x00")
}

func sourceDeclarationSortKey(declaration TaskGraphSourceDeclaration) string {
	return fmt.Sprintf("%s\x00%s\x00%s\x00%d\x00%s\x00%09d",
		declaration.Source.TaskID, declaration.Source.TaskSlug, declaration.Source.Location,
		dependencyFieldOrder(declaration.Field), declaration.Value, declaration.Occurrence)
}

func (g *TaskGraph) projectedSourceEdge(declaration TaskGraphSourceDeclaration) (DependencyEdge, bool) {
	if !g.isRepresentativeSource(declaration.Source) {
		return DependencyEdge{}, false
	}
	if declaration.Field == TaskDependencyDependsOn {
		if id.Valid(declaration.Source.TaskID) && id.Valid(declaration.Value) && taskExists(g.tasks, declaration.Value) {
			return DependencyEdge{From: declaration.Value, To: declaration.Source.TaskID}, true
		}
		return DependencyEdge{}, false
	}
	for _, diagnostic := range g.legacy {
		if diagnostic.TaskID != declaration.Source.TaskID || diagnostic.TaskSlug != declaration.Source.TaskSlug ||
			diagnostic.TaskPath != declaration.Source.Location || diagnostic.Field != string(declaration.Field) {
			continue
		}
		for _, reference := range diagnostic.References {
			if reference.Value == declaration.Value && reference.Edge.From != "" && reference.Edge.To != "" {
				return reference.Edge, true
			}
		}
	}
	return DependencyEdge{}, false
}

func (g *TaskGraph) isRepresentativeSource(source TaskGraphSourceRef) bool {
	return g.sourceRefCounts[source] == 1 && g.representative[source.TaskID] == source
}
