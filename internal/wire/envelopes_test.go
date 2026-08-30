package wire

import (
	"bytes"
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

	// Every envelope, validated against its own $defs entry — the whole --json
	// contract, not a sample. The embedded-struct envelopes (schema/schema_kind,
	// epic rollup) are here precisely because reflection of embedded fields is the
	// likeliest place the schema and the real output drift.
	cases := []struct {
		def  string
		emit func(io.Writer) error
	}{
		{"TasksEnvelope", func(w io.Writer) error { return emit(w, ToTasksEnvelope([]domain.Task{task, beta}, nil)) }},
		{"BoardEnvelope", func(w io.Writer) error {
			return emit(w, ToBoardEnvelope(core.Board{Columns: []core.BoardColumn{{Status: domain.StatusInProgress, Tasks: []domain.Task{task}}}}))
		}},
		{"TaskShowEnvelope", func(w io.Writer) error { return emit(w, ToTaskShowEnvelope(task, "# body")) }},
		{"TaskInfoEnvelope", func(w io.Writer) error {
			return emit(w, ToTaskInfoEnvelope(task, domain.ACCount{Checked: 1, Total: 3}, "/root/tasks/alpha.md"))
		}},
		{"PathEnvelope", func(w io.Writer) error { return emit(w, ToPathEnvelope("/root/tasks/alpha.md")) }},
		{"AcceptanceEnvelope", func(w io.Writer) error {
			return emit(w, ToAcceptanceEnvelope("alpha", []domain.Criterion{{Index: 1, Checked: true, Text: "done"}, {Index: 2, Checked: false, Text: "todo"}}))
		}},
		{"AuditInfoEnvelope", func(w io.Writer) error {
			return emit(w, ToAuditInfoEnvelope(domain.Audit{Slug: "x", Bucket: domain.AuditOpen, Findings: 3, OpenFindings: 1, ActiveFindings: 1, DoneFindings: 1}, "/root/audits/x.md"))
		}},
		{"TaskMutationEnvelope", func(w io.Writer) error {
			return emit(w, ToTaskMutationEnvelope(task, "# new body", true, WorkspaceJSON{}))
		}},
		{"DependencyMutationEnvelope", func(w io.Writer) error {
			return emit(w, ToDependencyMutationEnvelope(core.DependencyMutationReceipt{
				Operation: core.DependencyAdd, Changed: true, DryRun: true,
				Edges: []core.DependencyEdgeOutcome{{
					DependentID: "6g0000000002", PrerequisiteID: "6g0000000001",
					Action: core.DependencyAdd, Outcome: "added",
				}},
				PlannedTaskIDs: []string{"6g0000000002"},
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
			}, nil))
		}},
		{"ThreadShowEnvelope", func(w io.Writer) error {
			return emit(w, ToThreadShowEnvelope(threadView, "# Initiative\n"))
		}},
		{"ThreadFrontierEnvelope", func(w io.Writer) error {
			return emit(w, ToThreadFrontierEnvelope(threadView))
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
		{"EpicMutationEnvelope", func(w io.Writer) error { return emit(w, ToEpicMutationEnvelope(epic, true, WorkspaceJSON{})) }},
		{"CreatedEnvelope", func(w io.Writer) error {
			return emit(w, ToCreatedEnvelope("task", "6fsa428vc2mm", "alpha", "ready-to-start", "tasks/6fsa428vc2mm-alpha.md", false, WorkspaceJSON{}))
		}},
		{"MovesEnvelope", func(w io.Writer) error {
			return emit(w, ToMovesEnvelope([]MoveResult{{Slug: "alpha", To: "in-progress"}}, false, WorkspaceJSON{}))
		}},
		{"SummaryEnvelope", func(w io.Writer) error {
			return emit(w, ToSummaryEnvelope(core.Summary{
				Counts:     []core.StatusCount{{Status: domain.StatusInProgress, Count: 1}},
				InProgress: []domain.Task{task},
				Epics:      []core.EpicSummary{epicSum},
			}))
		}},
		{"StatusAllEnvelope", func(w io.Writer) error {
			summary := core.Summary{
				Counts:     []core.StatusCount{{Status: domain.StatusInProgress, Count: 1}},
				InProgress: []domain.Task{task},
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
			return emit(w, ToEpicShowEnvelope(epic, []domain.Task{task}, "# body"))
		}},
		{"ResearchListEnvelope", func(w io.Writer) error {
			return emit(w, ToResearchListEnvelope([]domain.Research{
				{ID: "6ff3hpm01p4a", Slug: "theming-libs", Created: "2026-06-23", Description: "Weighed three libs", Tags: []string{"tui"}},
			}, nil))
		}},
		{"ResearchShowEnvelope", func(w io.Writer) error {
			return emit(w, ToResearchShowEnvelope(
				domain.Research{ID: "6ff3hpm01p4a", Slug: "theming-libs", Created: "2026-06-23"}, "# body"))
		}},
		{"ResearchMutationEnvelope", func(w io.Writer) error {
			return emit(w, ToResearchMutationEnvelope(
				domain.Research{ID: "6ff3hpm01p4a", Slug: "theming-libs", Created: "2026-06-23", Updated: "2026-08-18"}, "# new body", true, WorkspaceJSON{}))
		}},
		{"AuditsEnvelope", func(w io.Writer) error {
			return emit(w, ToAuditsEnvelope([]domain.Audit{{Slug: "x", Bucket: domain.AuditOpen, Findings: 1, OpenFindings: 1}}, nil))
		}},
		{"AuditShowEnvelope", func(w io.Writer) error {
			return emit(w, ToAuditShowEnvelope(domain.Audit{Slug: "x", Bucket: domain.AuditOpen, Findings: 2, OpenFindings: 1}, "# body"))
		}},
		{"AuditMutationEnvelope", func(w io.Writer) error {
			return emit(w, ToAuditMutationEnvelope(domain.Audit{Slug: "x", Bucket: domain.AuditOpen, Findings: 2, OpenFindings: 1}, "# new body", true, WorkspaceJSON{}))
		}},
		{"FindingsEnvelope", func(w io.Writer) error {
			return emit(w, ToFindingsEnvelope([]core.AuditFinding{{
				Finding: domain.Finding{Code: "S1", Title: "tighten the gateway", Status: "open", Effort: "S", Urgency: "soon"},
				Audit:   "2026-01-01-area", Bucket: "open",
			}}, nil))
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
				ExitCodes:       []SchemaExitCode{{Code: 10, Name: "not-found"}},
				Kinds:           []string{"task"},
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
			// directly to prove its schema matches. Include the Thread post-commit
			// recovery payload so its nested, schema-version-free shape is covered too.
			return emit(w, ErrorEnvelope{SchemaVersion: SchemaVersion, Error: ErrorItem{
				Code: "conflict", Message: "Thread creation committed before cleanup failed",
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
