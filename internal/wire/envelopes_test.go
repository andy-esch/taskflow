package wire

import (
	"bytes"
	"encoding/json"
	"io"
	"reflect"
	"testing"

	"github.com/santhosh-tekuri/jsonschema/v6"

	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
)

// emit encodes the envelope value a constructor returns, so each case proves the
// constructor's output (the same value render's *JSON funcs encode, and the value a
// web handler would wrap) validates against the schema.
func emit(w io.Writer, v any) error { return EncodeJSON(w, v) }

func loadedTask(t domain.Task) core.LoadedRecord[domain.Task] {
	return core.LoadedRecord[domain.Task]{Value: t, Source: core.RecordSource{ID: t.ID}}
}

func loadedTaskBody(t domain.Task, body string) core.LoadedRecord[core.TaskWithBody] {
	return core.LoadedRecord[core.TaskWithBody]{Value: core.TaskWithBody{Task: t, Body: body}, Source: core.RecordSource{ID: t.ID}}
}

func loadedAuditBody(a domain.Audit, body string) core.LoadedRecord[core.AuditWithBody] {
	return core.LoadedRecord[core.AuditWithBody]{Value: core.AuditWithBody{Audit: a, Body: body}, Source: core.RecordSource{ID: a.ID}}
}

func loadedResearchBody(r domain.Research, body string) core.LoadedRecord[core.ResearchWithBody] {
	return core.LoadedRecord[core.ResearchWithBody]{Value: core.ResearchWithBody{Research: r, Body: body}, Source: core.RecordSource{ID: r.ID}}
}

func loadedResearchRecord(r domain.Research) core.LoadedRecord[domain.Research] {
	return core.LoadedRecord[domain.Research]{Value: r, Source: core.RecordSource{ID: r.ID}}
}

func TestToSchemaEnvelopeStampsRevisionPolicy(t *testing.T) {
	for _, input := range []SchemaRevisionPolicy{
		{},
		{Scheme: "stale", Scope: "contradictory"},
	} {
		envelope := ToSchemaEnvelope(SchemaContract{RevisionPolicy: input})
		if envelope.RevisionPolicy != CurrentSchemaRevisionPolicy() {
			t.Fatalf("schema envelope policy = %+v, want current wire policy", envelope.RevisionPolicy)
		}
	}
}

func TestToLintLoadProblemsJSONKeepsOpaqueLocationsOutOfPath(t *testing.T) {
	got := ToLintLoadProblemsJSON([]core.LoadProblem{{
		EntityKind: core.EntityAudit, EntityID: "6g0000000001", EntitySlug: "broken-audit",
		Location: "db://audits/6g0000000001", Message: "remote decode failed",
	}, {
		EntityKind: core.EntityAudit, EntityID: "6g0000000002",
		Location: "/repo/planning/audits/broken.md", LocalPath: "/repo/planning/audits/broken.md", Message: "local decode failed",
	}, {
		EntityKind: core.EntityTask, EntityID: "6g0000000003",
		Location: "db://tasks/3", LocalPath: "/repo/planning/tasks/repair.md",
		Message: "remote decode failed with a local repair copy",
	}})
	if len(got) != 3 {
		t.Fatalf("problems = %+v", got)
	}
	if got[0].EntityID != "6g0000000001" || got[0].EntitySlug != "broken-audit" ||
		got[0].Location != "db://audits/6g0000000001" || got[0].Path != "" {
		t.Fatalf("opaque problem = %+v", got[0])
	}
	if got[1].Location != "/repo/planning/audits/broken.md" || got[1].Path != got[1].Location {
		t.Fatalf("local problem = %+v", got[1])
	}
	if got[2].Location != "db://tasks/3" || got[2].Path != "/repo/planning/tasks/repair.md" {
		t.Fatalf("dual-location problem = %+v", got[2])
	}
}

func TestPathlessRepairDefectsExposeDiagnosticDeclarationsNotActions(t *testing.T) {
	source := core.TaskGraphSourceRef{TaskID: "6g0000000001", Location: "db://tasks/owner"}
	defects := []core.TaskGraphRepairDefect{
		{Reason: core.RepairInvalidID, Target: core.TaskGraphSourceEdit{
			Action: core.TaskGraphSourceDropDeclaration, Source: source,
			Field: core.TaskDependencyDependsOn, Value: "bad-one", Occurrence: 0,
		}},
		{Reason: core.RepairInvalidID, Target: core.TaskGraphSourceEdit{
			Action: core.TaskGraphSourceDropDeclaration, Source: source,
			Field: core.TaskDependencyDependsOn, Value: "bad-two", Occurrence: 0,
		}},
	}
	got := toTaskGraphRepairDefectsJSON(defects)
	if len(got) != 2 || got[0].Target != nil || got[1].Target != nil ||
		got[0].Declaration == nil || got[1].Declaration == nil ||
		got[0].Declaration.Value != "bad-one" || got[1].Declaration.Value != "bad-two" ||
		got[0].Declaration.Source.Location != source.Location || got[0].Declaration.Source.Path != "" {
		t.Fatalf("pathless declarations = %+v", got)
	}
}

func TestOrdinaryReadEnvelopesPreferSourceIdentityForEveryEntity(t *testing.T) {
	source := core.RecordSource{ID: "source-id", Location: "db://records/misleading-name"}
	task := domain.Task{ID: "declared-id", Slug: "task"}
	epic := domain.Epic{ID: "stale-epic", Description: "epic"}
	audit := domain.Audit{ID: "declared-id", Slug: "audit"}
	research := domain.Research{ID: "declared-id", Slug: "research"}

	if got := ToTasksEnvelope([]core.LoadedRecord[domain.Task]{{Value: task, Source: source}}, nil).Tasks[0].ID; got != source.ID {
		t.Fatalf("task list id = %q", got)
	}
	if got := ToTaskShowEnvelope(core.LoadedRecord[core.TaskWithBody]{Value: core.TaskWithBody{Task: task}, Source: source}).Task.ID; got != source.ID {
		t.Fatalf("task show id = %q", got)
	}
	summary := core.EpicSummary{Epic: epic, Source: source}
	if got := ToEpicsEnvelope([]core.EpicSummary{summary}, nil).Epics[0].ID; got != source.ID {
		t.Fatalf("epic list id = %q", got)
	}
	if got := ToEpicShowEnvelope(core.EpicDetail{Summary: summary}).Epic.ID; got != source.ID {
		t.Fatalf("epic show id = %q", got)
	}
	if got := ToAuditsEnvelope([]core.LoadedRecord[domain.Audit]{{Value: audit, Source: source}}, nil).Audits[0].ID; got != source.ID {
		t.Fatalf("audit list id = %q", got)
	}
	if got := ToAuditShowEnvelope(core.LoadedRecord[core.AuditWithBody]{Value: core.AuditWithBody{Audit: audit}, Source: source}).Audit.ID; got != source.ID {
		t.Fatalf("audit show id = %q", got)
	}
	if got := ToResearchListEnvelope([]core.LoadedRecord[domain.Research]{{Value: research, Source: source}}, nil).Research[0].ID; got != source.ID {
		t.Fatalf("research list id = %q", got)
	}
	if got := ToResearchShowEnvelope(core.LoadedRecord[core.ResearchWithBody]{Value: core.ResearchWithBody{Research: research}, Source: source}).Research.ID; got != source.ID {
		t.Fatalf("research show id = %q", got)
	}
}

func TestSummaryOpenAuditUsesSourceIdentity(t *testing.T) {
	source := core.RecordSource{ID: "canonical-audit", Location: "db://audits/one"}
	got := ToSummaryEnvelope(core.Summary{OpenAudits: []core.LoadedRecord[domain.Audit]{{
		Value:  domain.Audit{ID: "stale-declaration", Slug: "review", Bucket: domain.AuditOpen},
		Source: source,
	}}})
	if len(got.OpenAudits) != 1 || got.OpenAudits[0].ID != source.ID || got.OpenAudits[0].Location != source.Location {
		t.Fatalf("status open audit lost source identity: %+v", got.OpenAudits)
	}
}

func TestReadableLocationsAreOptionalAndNeverBecomePathsOrIdentity(t *testing.T) {
	source := core.RecordSource{ID: "canonical", Location: "db://records/one"}
	local := "/planning/records/local.md"
	task := domain.Task{ID: "declared", Slug: "same"}
	epic := domain.Epic{ID: "declared"}
	audit := domain.Audit{ID: "declared", Slug: "same"}
	research := domain.Research{ID: "declared", Slug: "same"}
	check := func(name string, payload any) {
		t.Helper()
		encoded, err := json.Marshal(payload)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Contains(encoded, []byte(`"id":"canonical"`)) ||
			!bytes.Contains(encoded, []byte(`"location":"db://records/one"`)) ||
			bytes.Contains(encoded, []byte(`"path":`)) || bytes.Contains(encoded, []byte(`"id":"declared"`)) {
			t.Fatalf("%s confused source ID, location, and local path: %s", name, encoded)
		}
	}
	check("task", ToLoadedTaskJSON(core.LoadedRecord[domain.Task]{Value: task, Source: source}))
	check("epic", ToEpicJSON(core.EpicSummary{Epic: epic, Source: source}))
	check("audit", ToLoadedAuditJSON(core.LoadedRecord[domain.Audit]{Value: audit, Source: source}))
	check("research", ToLoadedResearchJSON(core.LoadedRecord[domain.Research]{Value: research, Source: source}))
	for _, tc := range []struct {
		name    string
		payload func(core.RecordSource) any
	}{
		{"task", func(src core.RecordSource) any {
			return ToLoadedTaskJSON(core.LoadedRecord[domain.Task]{Value: task, Source: src})
		}},
		{"epic", func(src core.RecordSource) any { return ToEpicJSON(core.EpicSummary{Epic: epic, Source: src}) }},
		{"audit", func(src core.RecordSource) any {
			return ToLoadedAuditJSON(core.LoadedRecord[domain.Audit]{Value: audit, Source: src})
		}},
		{"research", func(src core.RecordSource) any {
			return ToLoadedResearchJSON(core.LoadedRecord[domain.Research]{Value: research, Source: src})
		}},
	} {
		for _, location := range []string{"", local} {
			encoded, err := json.Marshal(tc.payload(core.RecordSource{ID: "canonical", Location: location, LocationIsPath: location != ""}))
			if err != nil {
				t.Fatal(err)
			}
			if bytes.Contains(encoded, []byte(`"location":`)) {
				t.Fatalf("%s emitted absent/redundant location %q: %s", tc.name, location, encoded)
			}
		}
		encoded, err := json.Marshal(tc.payload(core.RecordSource{ID: "canonical", Location: local}))
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Contains(encoded, []byte(`"location":"`+local+`"`)) {
			t.Fatalf("%s inferred opaque location to be a local path: %s", tc.name, encoded)
		}
	}
	sourceLess := core.RecordSource{Location: "db://records/one"}
	for name, got := range map[string]string{
		"task":     ToLoadedTaskJSON(core.LoadedRecord[domain.Task]{Value: task, Source: sourceLess}).ID,
		"epic":     ToEpicJSON(core.EpicSummary{Epic: epic, Source: sourceLess}).ID,
		"audit":    ToLoadedAuditJSON(core.LoadedRecord[domain.Audit]{Value: audit, Source: sourceLess}).ID,
		"research": ToLoadedResearchJSON(core.LoadedRecord[domain.Research]{Value: research, Source: sourceLess}).ID,
	} {
		if got != "" {
			t.Fatalf("source-less %s inferred identity %q from location or declaration", name, got)
		}
	}
}

func TestEqualReadableRecordsRemainDistinctByLocationInEveryList(t *testing.T) {
	first := core.RecordSource{ID: "canonical", Location: "db://records/a"}
	second := core.RecordSource{ID: "canonical", Location: "db://records/b"}
	task := domain.Task{ID: "declared", Slug: "same"}
	epic := domain.Epic{ID: "declared"}
	audit := domain.Audit{ID: "declared", Slug: "same"}
	research := domain.Research{ID: "declared", Slug: "same"}
	for name, locations := range map[string][]string{
		"task": func() []string {
			rows := ToTasksEnvelope([]core.LoadedRecord[domain.Task]{{Value: task, Source: first}, {Value: task, Source: second}}, nil).Tasks
			return []string{rows[0].Location, rows[1].Location}
		}(),
		"epic": func() []string {
			rows := ToEpicsEnvelope([]core.EpicSummary{{Epic: epic, Source: first}, {Epic: epic, Source: second}}, nil).Epics
			return []string{rows[0].Location, rows[1].Location}
		}(),
		"audit": func() []string {
			rows := ToAuditsEnvelope([]core.LoadedRecord[domain.Audit]{{Value: audit, Source: first}, {Value: audit, Source: second}}, nil).Audits
			return []string{rows[0].Location, rows[1].Location}
		}(),
		"research": func() []string {
			rows := ToResearchListEnvelope([]core.LoadedRecord[domain.Research]{{Value: research, Source: first}, {Value: research, Source: second}}, nil).Research
			return []string{rows[0].Location, rows[1].Location}
		}(),
	} {
		if len(locations) != 2 || locations[0] != first.Location || locations[1] != second.Location {
			t.Fatalf("%s equal-value list locations = %v", name, locations)
		}
	}
}

func TestGraphProblemLocationNeverBecomesPath(t *testing.T) {
	got := toGraphProblemsJSON([]core.GraphProblem{{
		Code: core.ProblemDuplicateTaskID, TaskID: "canonical", Location: "db://tasks/first", Message: "duplicate",
	}, {
		Code: core.ProblemUnreadable, TaskID: "canonical", Path: "/planning/tasks/local.md",
		Location: "/planning/tasks/local.md", Message: "unreadable",
	}})
	if got[0].Location != "db://tasks/first" || got[0].Path != "" {
		t.Fatalf("opaque graph diagnostic confused location with path: %+v", got[0])
	}
	if got[1].Location != "" || got[1].Path != "/planning/tasks/local.md" {
		t.Fatalf("local graph diagnostic duplicated path as location: %+v", got[1])
	}
}

func TestGraphRepairSourceSeparatesDiagnosticLocationFromLocalTarget(t *testing.T) {
	source := core.TaskGraphSourceRef{
		TaskID: "canonical", Location: "db://tasks/one", LocalPath: "/planning/tasks/local.md",
	}
	got := ToTaskGraphRepairJSON(core.TaskGraphRepairReceipt{Selected: []core.TaskGraphSourceEdit{{
		Action: core.TaskGraphSourceDropDeclaration, Source: source,
		Field: core.TaskDependencyDependsOn, Value: "invalid",
	}}}, WorkspaceJSON{})
	if len(got.Selected) != 1 || got.Selected[0].Source.Location != source.Location ||
		got.Selected[0].Source.Path != source.LocalPath {
		t.Fatalf("repair source projection = %+v", got.Selected)
	}
}

func TestJSONSchemaRejectsEnvelopeFromAnotherRevision(t *testing.T) {
	schemaBytes, err := JSONSchema()
	if err != nil {
		t.Fatalf("JSONSchema: %v", err)
	}
	doc, err := jsonschema.UnmarshalJSON(bytes.NewReader(schemaBytes))
	if err != nil {
		t.Fatalf("unmarshal schema: %v", err)
	}
	id := doc.(map[string]any)["$id"].(string)
	c := jsonschema.NewCompiler()
	if err := c.AddResource(id, doc); err != nil {
		t.Fatalf("add resource: %v", err)
	}
	sch, err := c.Compile(id + "#/$defs/TaskShowEnvelope")
	if err != nil {
		t.Fatalf("compile task-show definition: %v", err)
	}

	payload, err := json.Marshal(ToTaskShowEnvelope(loadedTaskBody(domain.Task{ID: "6g0000000001", Slug: "alpha"}, "# Alpha\n")))
	if err != nil {
		t.Fatalf("marshal task-show envelope: %v", err)
	}
	var instance map[string]any
	if err := json.Unmarshal(payload, &instance); err != nil {
		t.Fatalf("decode task-show envelope: %v", err)
	}
	if err := sch.Validate(instance); err != nil {
		t.Fatalf("current revision should validate: %v", err)
	}
	instance["schema_version"] = "0.0"
	if err := sch.Validate(instance); err == nil {
		t.Fatal("exact-revision schema accepted an envelope declaring schema_version 0.0")
	}
}

// TestJSONSchema_ValidatesRealOutput is the round-trip proof: the emitted schema
// actually validates real --json output across a representative spread of
// envelopes (list, show, mutation, nested item, lint, and the nil-slice fix path).
func TestJSONSchema_ValidatesRealOutput(t *testing.T) {
	schemaBytes, err := JSONSchema()
	if err != nil {
		t.Fatalf("JSONSchema: %v", err)
	}
	doc, err := jsonschema.UnmarshalJSON(bytes.NewReader(schemaBytes))
	if err != nil {
		t.Fatalf("unmarshal schema: %v", err)
	}
	id := doc.(map[string]any)["$id"].(string)
	c := jsonschema.NewCompiler()
	if err := c.AddResource(id, doc); err != nil {
		t.Fatalf("add resource: %v", err)
	}

	task := domain.Task{Slug: "alpha", Status: domain.StatusInProgress, Tier: 2, Tags: []string{"x"}}
	// A second task so a multi-item TasksEnvelope is validated against the schema.
	beta := domain.Task{Slug: "beta", Status: domain.StatusReadyToStart, Tier: 3, Tags: []string{"y"}}
	epic := domain.Epic{ID: "e1", Status: "active", Description: "d"}
	epicSum := core.EpicSummary{Epic: epic, Total: 2, Done: 1}
	thread := domain.Thread{
		ID: "6g0000000003", Slug: "initiative", Status: domain.ThreadStatusUnstarted,
		Description: "Initiative", Goal: "Ship it", Created: "2026-08-29", Tasks: []string{},
	}
	threadView := core.ProjectThread(thread, core.NewTaskGraph(nil, nil))
	threadGraph := core.ProjectThreadGraph(thread, core.NewTaskGraph(nil, nil))

	// Every envelope, validated against its own $defs entry — the whole --json
	// contract, not a sample. The embedded-struct envelopes (schema/schema_kind,
	// epic rollup) are here precisely because reflection of embedded fields is the
	// likeliest place the schema and the real output drift.
	cases := []struct {
		def  string
		emit func(io.Writer) error
	}{
		{"TasksEnvelope", func(w io.Writer) error {
			return emit(w, ToTasksEnvelope([]core.LoadedRecord[domain.Task]{loadedTask(task), loadedTask(beta)}, nil))
		}},
		{"BoardEnvelope", func(w io.Writer) error {
			return emit(w, ToBoardEnvelope(core.Board{
				Columns: []core.BoardColumn{{Status: domain.StatusInProgress, Tasks: []core.LoadedRecord[domain.Task]{loadedTask(task)}}},
				Problems: []core.LoadProblem{{
					EntityKind: core.EntityTask, EntityID: "6g0000000007",
					Location: "remote:tasks/7", Message: "decode failed",
				}},
			}))
		}},
		{"TaskShowEnvelope", func(w io.Writer) error { return emit(w, ToTaskShowEnvelope(loadedTaskBody(task, "# body"))) }},
		{"TaskInfoEnvelope", func(w io.Writer) error {
			return emit(w, ToTaskInfoEnvelope(loadedTaskBody(task, "# body"), domain.ACCount{Checked: 1, Total: 3}, "/root/tasks/alpha.md"))
		}},
		{"PathEnvelope", func(w io.Writer) error { return emit(w, ToPathEnvelope("/root/tasks/alpha.md")) }},
		{"AcceptanceEnvelope", func(w io.Writer) error {
			return emit(w, ToAcceptanceEnvelope("alpha", []domain.Criterion{{Index: 1, Checked: true, Text: "done"}, {Index: 2, Checked: false, Text: "todo"}}))
		}},
		{"AuditInfoEnvelope", func(w io.Writer) error {
			return emit(w, ToAuditInfoEnvelope(loadedAuditBody(domain.Audit{Slug: "x", Bucket: domain.AuditOpen, Findings: 3, OpenFindings: 1, ActiveFindings: 1, DoneFindings: 1}, ""), "/root/audits/x.md"))
		}},
		{"TaskMutationEnvelope", func(w io.Writer) error {
			return emit(w, ToTaskMutationEnvelope(task, "# new body", true, WorkspaceJSON{}))
		}},
		{"TaskRenameEnvelope", func(w io.Writer) error {
			return emit(w, ToTaskRenameEnvelope(core.TaskRenameReceipt{
				Task:     domain.Task{ID: "6g0000000001", Slug: "renamed", Status: domain.StatusInProgress},
				FromSlug: "alpha", PlannedDocuments: 3, AppliedDocuments: 3,
				PlannedLinks: 2, AppliedLinks: 2, Changed: true,
				Committed: true, Complete: true, DestinationWritten: true, SourceRemoved: true,
			}, WorkspaceJSON{PlanningRoot: "/repo/planning", Source: WorkspaceSourceConfig}))
		}},
		{"DependencyMutationEnvelope", func(w io.Writer) error {
			return emit(w, ToDependencyMutationEnvelope(core.DependencyMutationReceipt{
				Operation: core.DependencyAdd, Changed: true, DryRun: true,
				Edges: []core.DependencyEdgeOutcome{{
					DependentID: "6g0000000002", PrerequisiteID: "6g0000000001",
					Action: core.DependencyAdd, Outcome: "added",
				}},
				PlannedTaskIDs: []string{"6g0000000002"},
				Impacts: []core.TaskGraphStateImpact{{
					TaskID: "6g0000000002", Direct: true,
					Before: core.TaskGraphState{TaskID: "6g0000000002", Role: core.RoleCandidate, Gate: core.GateClear, Eligible: true},
					After:  core.TaskGraphState{TaskID: "6g0000000002", Role: core.RoleCandidate, Gate: core.GateBlocked},
				}},
				Remedy: "preview only: inspect affected tasks before applying",
			}, WorkspaceJSON{PlanningRoot: "/repo/planning", Source: WorkspaceSourceConfig}))
		}},
		{"TaskGraphRepairEnvelope", func(w io.Writer) error {
			return emit(w, ToTaskGraphRepairEnvelope(core.TaskGraphRepairReceipt{
				Changed: true, DryRun: true, InitialHealth: core.GraphBroken, FinalHealth: core.GraphBroken,
				Selected: []core.TaskGraphSourceEdit{{
					Action: core.TaskGraphSourceDropDeclaration,
					Source: core.TaskGraphSourceRef{TaskID: "6g0000000002", TaskSlug: "alpha", Location: "db://tasks/alpha", LocalPath: "/repo/planning/tasks/alpha.md"},
					Field:  core.TaskDependencyDependsOn, Value: "invalid raw value",
				}},
				Operations: []core.TaskGraphRepairOperation{{
					Edit:   core.TaskGraphSourceEdit{Action: core.TaskGraphSourceDropDeclaration, Source: core.TaskGraphSourceRef{TaskID: "6g0000000002"}, Field: core.TaskDependencyDependsOn, Value: "invalid raw value"},
					Reason: core.RepairInvalidID,
				}},
				Residual: []core.TaskGraphRepairDefect{{
					Reason: core.RepairDirectEdit, Problem: core.GraphProblem{Code: core.ProblemUnreadable, Path: "/repo/planning/tasks/bad.md", Message: "bad yaml"},
				}},
				PlannedFiles:      []string{"/repo/planning/tasks/alpha.md"},
				IncompleteThreads: []core.ThreadReadProblem{{ThreadID: "6g0000000004", ThreadSlug: "bad-thread", Location: "/repo/planning/threads/bad.md", Message: "bad yaml"}},
			}, WorkspaceJSON{PlanningRoot: "/repo/planning", Source: WorkspaceSourceConfig}))
		}},
		{"TaskBlockersEnvelope", func(w io.Writer) error {
			return emit(w, ToTaskBlockersEnvelope(core.TaskBlockersResult{
				TaskID: "6g0000000002", Task: task,
				State:      core.TaskGraphState{TaskID: "6g0000000002", Role: core.RoleInFlight, Gate: core.GateBroken, Inconsistent: true},
				Projection: "frontier", Health: core.GraphBroken,
				Problems: []core.GraphProblem{{Code: core.ProblemLegacyMissing, TaskID: "6g0000000002", Field: "blocked_by", Message: "missing legacy reference"}},
				Legacy: []core.LegacyDependencyDiagnostic{{
					TaskID: "6g0000000002", TaskSlug: "alpha", Field: "blocked_by",
					References: []core.LegacyReference{{Value: "gone", Resolution: core.LegacyMissing}},
				}},
				Blockers: []core.TaskBlockerDetail{{
					Blocker: core.Blocker{TaskID: "6g0000000001", Reason: core.BlockerNotStarted, Path: []string{"6g0000000002", "6g0000000001"}, Direct: true},
					Task:    beta, State: core.TaskGraphState{TaskID: "6g0000000001", Role: core.RoleCandidate, Gate: core.GateClear, Eligible: true},
				}},
			}))
		}},
		{"TaskUnblocksEnvelope", func(w io.Writer) error {
			return emit(w, ToTaskUnblocksEnvelope(core.TaskUnblocksResult{
				TaskID: "6g0000000001", Task: beta,
				State:  core.TaskGraphState{TaskID: "6g0000000001", Role: core.RoleCandidate, Gate: core.GateClear, Eligible: true},
				Health: core.GraphHealthy,
				Unblocks: []core.TaskDependentDetail{{
					Impact: core.DependentImpact{TaskID: "6g0000000002", Path: []string{"6g0000000001", "6g0000000002"}, Direct: true},
					Task:   task, State: core.TaskGraphState{TaskID: "6g0000000002", Role: core.RoleInFlight, Gate: core.GateClear},
				}},
			}))
		}},
		{"ThreadsEnvelope", func(w io.Writer) error {
			return emit(w, ToThreadsEnvelope(core.ThreadListView{
				Threads: []core.ThreadView{threadView}, GraphHealth: threadView.GraphHealth,
				GraphProblems: threadView.GraphProblems,
			}, []core.ThreadReadProblem{{
				ThreadID: "6g0000000004", ThreadSlug: "pathless-record", Message: "remote decode failed",
			}}))
		}},
		{"ThreadShowEnvelope", func(w io.Writer) error {
			return emit(w, ToThreadShowEnvelope(threadView, "# Initiative\n"))
		}},
		{"ThreadFrontierEnvelope", func(w io.Writer) error {
			return emit(w, ToThreadFrontierEnvelope(threadView))
		}},
		{"ThreadGraphEnvelope", func(w io.Writer) error {
			return emit(w, ToThreadGraphEnvelope(threadGraph))
		}},
		{"ThreadPlanEnvelope", func(w io.Writer) error {
			return emit(w, ToThreadPlanEnvelope(threadGraph))
		}},
		{"ThreadMutationEnvelope", func(w io.Writer) error {
			return emit(w, ToThreadMutationEnvelope(core.ThreadCreationReceipt{
				Thread: thread, Changed: true, DryRun: true,
			}, "threads/6g0000000003-initiative.md", WorkspaceJSON{PlanningRoot: "/repo/planning", Source: WorkspaceSourceConfig}))
		}},
		{"ThreadUpdateEnvelope", func(w io.Writer) error {
			return emit(w, ToThreadUpdateEnvelope(core.ThreadMutationReceipt{
				Operation: core.ThreadMutationAddMembers, Thread: thread,
				Before: threadView, After: threadView,
				MemberOutcomes: []core.ThreadMemberOutcome{{TaskID: task.ID, Action: "add", Outcome: "skipped"}},
				DryRun:         true,
			}, "threads/6g0000000003-initiative.md", WorkspaceJSON{PlanningRoot: "/repo/planning", Source: WorkspaceSourceConfig}))
		}},
		{"ThreadApplyComposeEnvelope", func(w io.Writer) error {
			plan := core.ThreadApplyPlan{
				Schema: core.ThreadApplyPlanSchema, PlanningRepoID: "6gplan", ComposedAt: "2026-08-30",
				Thread: core.ThreadApplyThread{
					ID: "6g0000000004", Slug: "bulk", Status: domain.ThreadStatusUnstarted,
					Description: "Bulk Thread", Goal: "Ship it", Created: "2026-08-30",
					Tasks: []string{"6g0000000002"}, Body: "# Bulk\n",
				},
			}
			return emit(w, ToThreadApplyComposeEnvelope(plan, "/tmp/plan.yml", true, WorkspaceJSON{}))
		}},
		{"ThreadApplyEnvelope", func(w io.Writer) error {
			return emit(w, ToThreadApplyEnvelope(core.ThreadApplyReceipt{
				Plan:    core.ThreadApplyPlan{Thread: core.ThreadApplyThread{ID: "6g0000000004", Slug: "bulk"}},
				Changed: true, DryRun: true,
				Operations: []core.ThreadApplyOperation{{
					Kind: "thread", Action: "create", State: core.ThreadApplyPending, ThreadID: "6g0000000004",
				}},
			}, "/tmp/plan.yml", WorkspaceJSON{}))
		}},
		{"EpicMutationEnvelope", func(w io.Writer) error { return emit(w, ToEpicMutationEnvelope(epic, true, WorkspaceJSON{})) }},
		{"CreatedEnvelope", func(w io.Writer) error {
			return emit(w, ToCreatedEnvelope("task", "6fsa428vc2mm", "alpha", "ready-to-start", "tasks/6fsa428vc2mm-alpha.md", false, WorkspaceJSON{}))
		}},
		{"MovesEnvelope", func(w io.Writer) error {
			afterThread := threadView
			afterThread.Inconsistent = true
			lifecycle := ToTaskLifecycleJSON(core.TaskLifecycleReceipt{
				Task: task, Changed: true, Committed: true,
				Impacts: []core.TaskGraphStateImpact{{
					TaskID: "6g0000000002", Direct: true,
					Before: core.TaskGraphState{TaskID: "6g0000000002", Role: core.RoleInFlight, Gate: core.GateClear},
					After:  core.TaskGraphState{TaskID: "6g0000000002", Role: core.RoleInFlight, Gate: core.GateBlocked, Inconsistent: true},
				}},
				ThreadImpacts: []core.ThreadProjectionImpact{{
					ThreadID: thread.ID, Slug: thread.Slug, Before: threadView, After: afterThread,
				}},
				Remedy: "inspect the affected task and Thread",
			})
			return emit(w, ToMovesEnvelope([]MoveResult{{Slug: "alpha", To: "in-progress", Lifecycle: &lifecycle}}, false, WorkspaceJSON{}))
		}},
		{"SummaryEnvelope", func(w io.Writer) error {
			return emit(w, ToSummaryEnvelope(core.Summary{
				Counts:            []core.StatusCount{{Status: domain.StatusInProgress, Count: 1}},
				InProgressRecords: []core.LoadedRecord[domain.Task]{loadedTask(task)},
				Epics:             []core.EpicSummary{epicSum},
				Problems: []core.LoadProblem{{
					EntityKind: core.EntityTask, EntityID: "6g0000000008",
					Location: "db://tasks/8", Message: "decode failed",
				}},
				GraphHealth: core.GraphDegraded,
				GraphDetail: "one safe legacy dependency field remains",
			}))
		}},
		{"StatusAllEnvelope", func(w io.Writer) error {
			summary := core.Summary{
				Counts:            []core.StatusCount{{Status: domain.StatusInProgress, Count: 1}},
				InProgressRecords: []core.LoadedRecord[domain.Task]{loadedTask(task)},
				Problems: []core.LoadProblem{{
					EntityKind: core.EntityAudit, EntitySlug: "broken-audit",
					Location: "/repo/planning/audits/broken.md", LocalPath: "/repo/planning/audits/broken.md",
					Message: "decode failed",
				}},
				GraphHealth: core.GraphBroken,
				GraphDetail: "canonical dependency target is missing",
			}
			return emit(w, ToStatusAllEnvelope(core.SpaceOverview{
				Spaces: []core.SpaceSummary{{
					ID: "planning", PlanningID: "6gplan",
					Selected: &core.SpaceEntryPoint{ID: "planning", Role: core.SpaceRoleDirect, State: core.SpaceStateOK},
					Entries: []core.SpaceEntryPoint{
						{ID: "implementation", Path: "/repo/impl", PlanningID: "6gplan", Role: core.SpaceRolePointer, State: core.SpaceStateMissing},
						{ID: "planning", Path: "/repo/planning", PlanningID: "6gplan", Role: core.SpaceRoleDirect, State: core.SpaceStateOK, Root: "/repo/planning"},
					},
					Summary: &summary,
				}},
				InProgress: []core.SpaceInProgress{{SpaceID: "planning", PlanningID: "6gplan", Task: task}},
			}))
		}},
		{"VersionEnvelope", func(w io.Writer) error { return emit(w, ToVersionEnvelope("v0.6.0")) }},
		{"ThemesEnvelope", func(w io.Writer) error {
			return emit(w, ToThemesEnvelope([]ThemeEntry{{Name: "neon", Active: true, Default: true}}))
		}},
		{"ThemePreviewEnvelope", func(w io.Writer) error {
			return emit(w, ToThemePreviewEnvelope("neon", "dark", []ThemeSwatch{{Token: "accent", Hex: "#ea5ce2", ANSI: 13}}))
		}},
		{"EpicsEnvelope", func(w io.Writer) error { return emit(w, ToEpicsEnvelope([]core.EpicSummary{epicSum}, nil)) }},
		{"EpicShowEnvelope", func(w io.Writer) error {
			return emit(w, ToEpicShowEnvelope(core.EpicDetail{Summary: core.EpicSummary{Epic: epic, Source: core.RecordSource{ID: epic.ID}}, Tasks: []core.LoadedRecord[domain.Task]{loadedTask(task)}, Body: "# body"}))
		}},
		{"ResearchListEnvelope", func(w io.Writer) error {
			return emit(w, ToResearchListEnvelope([]core.LoadedRecord[domain.Research]{loadedResearchRecord(domain.Research{
				ID: "6ff3hpm01p4a", Slug: "theming-libs", Created: "2026-06-23", Description: "Weighed three libs", Tags: []string{"tui"},
			}),
			}, nil))
		}},
		{"ResearchShowEnvelope", func(w io.Writer) error {
			return emit(w, ToResearchShowEnvelope(loadedResearchBody(
				domain.Research{ID: "6ff3hpm01p4a", Slug: "theming-libs", Created: "2026-06-23"}, "# body")))
		}},
		{"ResearchMutationEnvelope", func(w io.Writer) error {
			return emit(w, ToResearchMutationEnvelope(
				domain.Research{ID: "6ff3hpm01p4a", Slug: "theming-libs", Created: "2026-06-23", Updated: "2026-08-18"}, "# new body", true, WorkspaceJSON{}))
		}},
		{"AuditsEnvelope", func(w io.Writer) error {
			return emit(w, ToAuditsEnvelope([]core.LoadedRecord[domain.Audit]{{Value: domain.Audit{Slug: "x", Bucket: domain.AuditOpen, Findings: 1, OpenFindings: 1}, Source: core.RecordSource{ID: "6gaudit00001"}}}, nil))
		}},
		{"AuditShowEnvelope", func(w io.Writer) error {
			return emit(w, ToAuditShowEnvelope(loadedAuditBody(domain.Audit{Slug: "x", Bucket: domain.AuditOpen, Findings: 2, OpenFindings: 1}, "# body")))
		}},
		{"AuditMutationEnvelope", func(w io.Writer) error {
			return emit(w, ToAuditMutationEnvelope(domain.Audit{Slug: "x", Bucket: domain.AuditOpen, Findings: 2, OpenFindings: 1}, "# new body", true, WorkspaceJSON{}))
		}},
		{"FindingCreationEnvelope", func(w io.Writer) error {
			return emit(w, ToFindingCreationEnvelope(core.FindingCreationReceipt{
				Audit:   domain.Audit{Slug: "x", Bucket: domain.AuditOpen, Findings: 2, OpenFindings: 2},
				Finding: domain.Finding{Code: "H2", Title: "new issue", Status: "open", Effort: "S", Urgency: "soon"},
				DryRun:  true,
			}, WorkspaceJSON{}))
		}},
		{"FindingsEnvelope", func(w io.Writer) error {
			return emit(w, ToFindingsEnvelope([]core.AuditFinding{{
				Finding: domain.Finding{Code: "S1", Title: "tighten the gateway", Status: "open", Effort: "S", Urgency: "soon"},
				Audit:   "2026-01-01-area", Bucket: "open",
			}}, []core.LoadProblem{{
				EntityKind: core.EntityAudit, EntityID: "6g0000000004", EntitySlug: "broken-audit",
				Location: "db://audits/6g0000000004", Message: "remote decode failed",
			}}))
		}},
		{"LintEnvelope", func(w io.Writer) error {
			return emit(w, ToLintEnvelope([]core.LintResult{{Slug: "alpha", Issues: []domain.Issue{{Field: "epic", Message: "missing"}}}}, nil))
		}},
		{"FixEnvelope", func(w io.Writer) error {
			return emit(w, ToFixEnvelope(nil, nil, nil, false, WorkspaceJSON{})) // the nil-slice path: must emit [] and validate
		}},
		{"InitEnvelope", func(w io.Writer) error {
			return emit(w, NormalizeInitEnvelope(InitEnvelope{
				Mode: "scaffold", Root: "/root", Created: []string{"tasks"},
				Registration: ToInitRegistrationJSON(core.SpaceRegistrationReceipt{
					ID: "root", Path: "/root", VerifyID: "6gplan", Changed: true,
				}),
			}))
		}},
		{"DoctorEnvelope", func(w io.Writer) error {
			return emit(w, ToDoctorEnvelope(
				"/root",
				[]DoctorProblem{{Repo: "../impl", Message: "one-sided link"}},
				DoctorRegistry{Checked: 2, Problems: []DoctorSpaceProblem{{
					ID: "missing", Path: "~/git/missing", Kind: SpaceStateMissing,
					Message: "not found", Remedy: "forget or re-add",
				}}},
			))
		}},
		{"SchemaEnvelope", func(w io.Writer) error {
			return emit(w, ToSchemaEnvelope(SchemaContract{
				Statuses:        []SchemaStatus{{Value: "in-progress", Active: true}},
				EpicStatuses:    []string{"active"},
				ThreadStatuses:  []string{"unstarted"},
				AuditBuckets:    []string{"open"},
				FindingStatuses: []string{"open", "fixed"},
				CriterionStates: []string{"deferred", "wontfix"},
				TaskFields:      []SchemaField{{Name: "tier", Type: "int"}},
				EpicFields:      []string{"status", "description"},
				ResearchFields:  []SchemaField{{Name: "created", Type: "date"}},
				ExitCodes: []SchemaExitCode{{
					Code: 10, Name: "not-found", State: ExitCodeStateActive,
					Meaning: "a requested named entity or registered planning space does not exist",
				}},
				Commands: []SchemaCommand{{Path: "tskflwctl task list", Safety: "read-only"}},
				Kinds:    []string{"task"},
			}))
		}},
		{"SchemaKindEnvelope", func(w io.Writer) error {
			return emit(w, ToSchemaKindEnvelope(KindSchema{
				Kind:         "task",
				Sections:     []string{"Objective"},
				BodyTemplate: "## Objective\n",
				Fields:       []domain.FieldDoc{{Name: "tier", Type: "int", Required: true, Description: "d", Example: "3"}},
				Conventions:  []string{"c"},
				Templates:    []TemplateInfo{{Kind: "task", Name: "default", Description: "d"}},
			}))
		}},
		{"TemplatesEnvelope", func(w io.Writer) error {
			return emit(w, ToTemplatesEnvelope([]TemplateInfo{{Kind: "task", Name: "default", Description: "d"}}))
		}},
		{"TemplateShowEnvelope", func(w io.Writer) error {
			return emit(w, ToTemplateShowEnvelope(TemplateInfo{Kind: "task", Name: "default", Description: "d"}, "# body"))
		}},
		{"WorkspaceEnvelope", func(w io.Writer) error {
			return emit(w, ToWorkspaceEnvelope(WorkspaceJSON{
				PlanningRoot: "/repo/planning", ConfigPath: "/repo/.tskflwctl.toml", Source: WorkspaceSourcePointer,
			}))
		}},
		{"ConfigEnvelope", func(w io.Writer) error {
			enabled := false
			return emit(w, ToConfigEnvelope(core.ConfigurationSnapshot{
				Repository: core.RepositoryConfiguration{
					Path: "/repo/.tskflwctl.toml", PlanningRoot: "/repo/planning", Mode: core.ConfigModeScaffold,
					TrackedRepos: []string{}, PendingMigration: []core.ConfigurationMigrationKind{},
				},
				User: core.UserConfiguration{Path: "/home/me/.config/tskflwctl/config.toml", PagerEnabled: &enabled},
				Effective: core.EffectiveConfiguration{
					Theme:        core.EffectiveString{Value: "neon", Source: core.ConfigSourceDefault},
					PagerEnabled: core.EffectiveBool{Value: false, Source: core.ConfigSourceUser},
					PagerCommand: core.EffectiveString{Value: "less -FRX", Source: core.ConfigSourceDefault},
				},
			}))
		}},
		{"ConfigMigrationEnvelope", func(w io.Writer) error {
			return emit(w, ToConfigMigrationEnvelope(core.ConfigurationMigration{
				ConfigPath: "/repo/.tskflwctl.toml", Mode: core.ConfigModeScaffold, DryRun: true,
				Steps: []core.ConfigurationMigrationStep{{Kind: core.ConfigurationMigrationRepoID, Key: "id", Value: "6gid"}},
			}, WorkspaceJSON{PlanningRoot: "/repo/planning", Source: WorkspaceSourceConfig}))
		}},
		{"SpacesEnvelope", func(w io.Writer) error {
			return emit(w, ToSpacesEnvelope([]SpaceEntry{{
				ID: "taskflow", Path: "~/git/taskflow", PlanningID: "6gplan", Role: SpaceRoleDirect, State: SpaceStateMismatch,
				Root: "/repo/planning", Detail: "wrong repo", Remedy: "re-register",
			}}))
		}},
		{"SpaceMutationEnvelope", func(w io.Writer) error {
			return emit(w, ToSpaceMutationEnvelope(
				SpaceEntry{ID: "taskflow", Path: "~/git/taskflow", PlanningID: "6gplan", Role: SpaceRoleDirect, State: SpaceStateOK, Root: "/repo/planning"}, true, false))
		}},
		{"ErrorEnvelope", func(w io.Writer) error {
			// Built by cli.WriteError (not a constructor here) — marshal the named type
			// directly to prove its schema matches. Include post-commit recovery payloads
			// so their nested, schema-version-free shapes are covered too.
			created := ToCreatedRecoveryJSON(CreatedItem{
				Kind: "task", ID: "6g0000000002", Slug: "created", Status: "ready-to-start",
				Path: "tasks/6g0000000002-created.md",
			}, WorkspaceJSON{PlanningRoot: "/repo/planning", Source: WorkspaceSourceConfig})
			dependency := ToDependencyMutationJSON(core.DependencyMutationReceipt{
				Operation: core.DependencyMigrate, Changed: true,
				PlannedTaskIDs: []string{"6g0000000001", "6g0000000002"},
				AppliedTaskIDs: []string{"6g0000000001"}, RemainingTaskIDs: []string{"6g0000000002"},
				Remedy: "inspect current graph and durable progress before resuming",
			}, WorkspaceJSON{})
			return emit(w, ErrorEnvelope{SchemaVersion: SchemaVersion, Error: ErrorItem{
				Code: "conflict", Message: "Thread creation committed before cleanup failed",
				Created: &created, DependencyMutation: &dependency,
				TaskRename: &TaskRenameRecoveryJSON{
					TaskRenameJSON: TaskRenameJSON{
						TaskID: "6g0000000001", FromSlug: "old", ToSlug: "new",
						PlannedDocuments: 3, AppliedDocuments: 1, PlannedLinks: 2, AppliedLinks: 1,
						Changed: true, Committed: true,
						Workspace: WorkspaceJSON{PlanningRoot: "/repo/planning", Source: WorkspaceSourceConfig},
					},
					Remedy: "rerun by stable id",
				},
				ThreadMutation: &ThreadMutationJSON{
					Thread: ToThreadJSON(thread), Changed: true, Committed: true,
					Path:      "threads/6g0000000003-initiative.md",
					Workspace: WorkspaceJSON{PlanningRoot: "/repo/planning", Source: WorkspaceSourceConfig},
				},
			}})
		}},
	}
	// Registry-derived coverage guard (replaces a brittle literal count): every
	// envelope type the jsonEnvelopes registry pulls into the schema must have a
	// case here, so a newly-added envelope can't be silently left unvalidated. The
	// $defs key is the Go type name, which is also each case's `def`. ErrorEnvelope
	// is built by cli.WriteError (not a constructor here) but is still a registered
	// envelope with a case, so it's covered too.
	covered := make(map[string]bool, len(cases))
	for _, tc := range cases {
		covered[tc.def] = true
	}
	rt := reflect.TypeOf(Envelopes())
	for i := range rt.NumField() {
		def := rt.Field(i).Type.Name()
		if !covered[def] {
			t.Errorf("envelope %q is in the jsonEnvelopes registry but has no validation case", def)
		}
	}
	for _, tc := range cases {
		sch, err := c.Compile(id + "#/$defs/" + tc.def)
		if err != nil {
			t.Errorf("compile %s: %v", tc.def, err)
			continue
		}
		var buf bytes.Buffer
		if err := tc.emit(&buf); err != nil {
			t.Errorf("emit %s: %v", tc.def, err)
			continue
		}
		inst, err := jsonschema.UnmarshalJSON(&buf)
		if err != nil {
			t.Errorf("unmarshal %s output: %v", tc.def, err)
			continue
		}
		if err := sch.Validate(inst); err != nil {
			t.Errorf("%s output does NOT validate against its own schema:\n%v", tc.def, err)
		}
	}
}

func TestStatusAllEnvelope_PreservesRetainedSummaryAndFailure(t *testing.T) {
	summary := core.Summary{InProgressRecords: []core.LoadedRecord[domain.Task]{{
		Value: domain.Task{Slug: "working", Status: domain.StatusInProgress}, Source: core.RecordSource{ID: "working-id"},
	}}}
	envelope := ToStatusAllEnvelope(core.SpaceOverview{Spaces: []core.SpaceSummary{{
		ID: "planning", PlanningID: "6gplan", Summary: &summary,
		Failure: &core.SpaceLoadFailure{
			Class: domain.ClassConflict, Message: "planner window is active",
		},
		Stale: true,
	}}})

	if len(envelope.Spaces) != 1 || envelope.Spaces[0].Summary == nil ||
		envelope.Spaces[0].Error != "planner window is active" {
		t.Fatalf("retained summary/failure mapping = %+v", envelope.Spaces)
	}
}

func TestStatusAllEnvelopeUsesSourceIdentityInCombinedWorkingSet(t *testing.T) {
	task := domain.Task{ID: "declared-id", Slug: "working", Status: domain.StatusInProgress}
	source := core.RecordSource{ID: "source-id", Location: "db://tasks/working"}
	summary := core.Summary{InProgressRecords: []core.LoadedRecord[domain.Task]{{Value: task, Source: source}}}
	envelope := ToStatusAllEnvelope(core.SpaceOverview{
		Spaces:     []core.SpaceSummary{{ID: "planning", Summary: &summary}},
		InProgress: []core.SpaceInProgress{{SpaceID: "planning", Task: task, Source: source}},
	})
	if len(envelope.InProgress) != 1 || len(envelope.Spaces) != 1 || envelope.Spaces[0].Summary == nil {
		t.Fatalf("status-all envelope = %+v", envelope)
	}
	got := envelope.InProgress[0].Task
	want := envelope.Spaces[0].Summary.InProgress[0]
	if got.ID != source.ID || got.Location != source.Location || !reflect.DeepEqual(got, want) {
		t.Fatalf("combined task = %+v, nested task = %+v", got, want)
	}
}

// TestMutationEnvelopes_CarryWorkspace pins the 1.31 contract structurally: a receipt
// for a WRITE must name the planning tree it wrote to, so a caller can prove which one
// it changed without a second read (audit 2026-07-24-ai-agent-cli-ergonomics, H1).
//
// `dry_run` is the marker for "this is a mutation receipt" — it appears on nothing
// else, precisely because a preview must be distinguishable from a real write.
//
// This is a registry-driven guard rather than one more per-entity assertion because
// the gap it catches is one of OMISSION. `research set` shipped without a workspace
// object purely because its verbs were authored before 1.31 existed; it compiled, its
// tests passed, its envelope was registered, and it had a validation case. Nothing
// else in the suite could have noticed.
func TestMutationEnvelopes_CarryWorkspace(t *testing.T) {
	// These commands deliberately mutate without an existing resolved workspace: init
	// creates one, while space add/forget edits the home-scoped advisory registry rather
	// than any planning tree. Both receipts report their actual target directly.
	exempt := map[string]string{
		"InitEnvelope":          "init creates the tree, so there is no resolved workspace; it reports Root itself",
		"SpaceMutationEnvelope": "space add/forget edits the home registry, not a planning tree; Space.Path names its target",
	}

	rt := reflect.TypeOf(Envelopes())
	checked := 0
	for i := range rt.NumField() {
		et := rt.Field(i).Type
		if _, isMutation := et.FieldByName("DryRun"); !isMutation {
			continue
		}
		if _, ok := exempt[et.Name()]; ok {
			continue
		}
		checked++
		if _, ok := et.FieldByName("Workspace"); !ok {
			t.Errorf("%s carries dry_run (so it is a mutation receipt) but has no Workspace field — "+
				"every write receipt must name the planning tree it changed; add it, or add a "+
				"documented exemption here", et.Name())
		}
	}
	if checked == 0 {
		t.Fatal("no mutation envelopes were checked — the dry_run heuristic has stopped working")
	}
}
