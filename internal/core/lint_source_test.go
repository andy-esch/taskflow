package core

import (
	"strings"
	"testing"

	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/testutil"
)

type locationLintSource struct {
	lintSourceFake
	tasks    []LoadedRecord[TaskWithBody]
	epics    []LoadedRecord[domain.Epic]
	audits   []LoadedRecord[AuditWithFindings]
	research []LoadedRecord[domain.Research]
}

func (f *locationLintSource) ReadLintTasks() ([]LoadedRecord[TaskWithBody], []LoadProblem, error) {
	return f.tasks, nil, nil
}

func (f *locationLintSource) ReadLintEpics() ([]LoadedRecord[domain.Epic], []LoadProblem, error) {
	return f.epics, nil, nil
}

func (f *locationLintSource) ReadAuditSnapshot(string) (AuditSnapshot, error) {
	return AuditSnapshot{Audits: f.audits}, nil
}

func (f *locationLintSource) ReadLintResearch() ([]LoadedRecord[domain.Research], []LoadProblem, error) {
	return f.research, nil, nil
}

func TestLintAttributesTaskIDDriftToAdapterSourceForActiveAndArchivedTasks(t *testing.T) {
	source := &locationLintSource{}
	for _, status := range []domain.Status{domain.StatusReadyToStart, domain.StatusCompleted} {
		slug := string(status)
		source.tasks = append(source.tasks, LoadedRecord[TaskWithBody]{
			Value:  TaskWithBody{Task: domain.Task{ID: testutil.TaskID("declared-" + slug), Slug: slug, Status: status}},
			Source: RecordSource{ID: testutil.TaskID("source-" + slug), Location: "db://tasks/" + slug},
		})
	}
	results, _, err := MustNewService(nil, WithLintSource(source)).Lint()
	if err != nil {
		t.Fatal(err)
	}
	for _, record := range source.tasks {
		assertLintIssue(t, results, record.Value.Task.Slug, "id", record.Source.ID)
	}
}

type threadLintSource struct {
	threadReadFake
	records []VersionedRecord[domain.Thread]
}

func (f *threadLintSource) ReadThreads() (ThreadRead, error) {
	return ThreadRead{Records: f.records}, nil
}

func TestLintAttributesThreadIdentityAndLocationToAdapterSource(t *testing.T) {
	declaredID := testutil.TaskID("declared-thread")
	firstID := testutil.TaskID("source-thread-first")
	secondID := testutil.TaskID("source-thread-second")
	threads := &threadLintSource{records: []VersionedRecord[domain.Thread]{
		{Record: LoadedRecord[domain.Thread]{
			Value: domain.Thread{ID: declaredID, Slug: "first", Status: domain.ThreadStatusUnstarted,
				Description: "First Thread", Goal: "Track first", Created: "2026-10-02"},
			Source: RecordSource{ID: firstID, Location: "db://threads/first"},
		}},
		{Record: LoadedRecord[domain.Thread]{
			Value: domain.Thread{ID: declaredID, Slug: "second", Status: domain.ThreadStatusUnstarted,
				Description: "Second Thread", Goal: "Track second", Created: "2026-10-02"},
			Source: RecordSource{ID: secondID, Location: "threads/path-shaped.md"},
		}},
	}}
	results, _, err := MustNewService(nil, WithLintSource(&lintSourceFake{}), WithThreadStore(threads)).Lint()
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []struct{ slug, id, location string }{
		{"first", firstID, "db://threads/first"},
		{"second", secondID, "threads/path-shaped.md"},
	} {
		assertLintIssue(t, results, want.slug, "id", want.id)
		found := false
		for _, result := range results {
			if result.Slug == want.slug && result.Location == want.location {
				found = true
				for _, issue := range result.Issues {
					if strings.Contains(issue.Message, "duplicate stable id") {
						t.Fatalf("distinct source IDs reported duplicate: %+v", results)
					}
				}
			}
		}
		if !found {
			t.Fatalf("missing Thread source location %q: %+v", want.location, results)
		}
	}
}

type lintSourceFake struct {
	testSourceSetProvider
	taskRecords      []TaskWithBody
	taskProblems     []LoadProblem
	epicProblems     []LoadProblem
	auditProblems    []LoadProblem
	researchProblems []LoadProblem
	taskReads        int
	epicReads        int
	auditReads       int
	researchReads    int
}

func (f *lintSourceFake) ReadLintTasks() ([]LoadedRecord[TaskWithBody], []LoadProblem, error) {
	f.taskReads++
	out := make([]LoadedRecord[TaskWithBody], 0, len(f.taskRecords))
	for _, record := range f.taskRecords {
		out = append(out, LoadedRecord[TaskWithBody]{Value: record, Source: RecordSource{ID: record.Task.CanonicalID(), Location: record.Task.Path}})
	}
	return out, f.taskProblems, nil
}

func (f *lintSourceFake) ReadLintEpics() ([]LoadedRecord[domain.Epic], []LoadProblem, error) {
	f.epicReads++
	return nil, f.epicProblems, nil
}

func (f *lintSourceFake) ReadAuditSnapshot(string) (AuditSnapshot, error) {
	f.auditReads++
	return AuditSnapshot{Problems: f.auditProblems}, nil
}

func (f *lintSourceFake) ReadLintResearch() ([]LoadedRecord[domain.Research], []LoadProblem, error) {
	f.researchReads++
	return nil, f.researchProblems, nil
}

func TestLintPreservesPortableLoadProblemIdentityWithoutLocations(t *testing.T) {
	source := &lintSourceFake{
		taskProblems: []LoadProblem{{
			EntityKind: EntityTask, EntityID: "6g0000000001", EntitySlug: "broken-task", Message: "bad task",
		}},
		epicProblems: []LoadProblem{{
			EntityKind: EntityEpic, EntityID: "21-broken-epic", Message: "bad epic",
		}},
		auditProblems: []LoadProblem{{
			EntityKind: EntityAudit, EntityID: "6g0000000002", EntitySlug: "2026-09-23-broken-audit", Message: "bad audit",
		}},
		researchProblems: []LoadProblem{{
			EntityKind: EntityResearch, EntityID: "6g0000000003", EntitySlug: "broken-research", Message: "bad research",
		}},
	}
	threads := &threadReadFake{problems: []ThreadReadProblem{{
		ThreadID: "6g0000000004", ThreadSlug: "broken-thread", Message: "bad Thread",
	}}}

	_, problems, err := MustNewService(nil, WithLintSource(source), WithThreadStore(threads)).Lint()
	if err != nil {
		t.Fatal(err)
	}
	if source.taskReads != 1 || source.epicReads != 1 || source.auditReads != 1 || source.researchReads != 1 {
		t.Fatalf("lint source reads = task:%d epic:%d audit:%d research:%d; want one each",
			source.taskReads, source.epicReads, source.auditReads, source.researchReads)
	}
	if len(problems) != 5 {
		t.Fatalf("problems = %+v; want all five entity kinds", problems)
	}
	byKind := make(map[EntityKind]LoadProblem, len(problems))
	for _, problem := range problems {
		byKind[problem.EntityKind] = problem
		if problem.Location != "" || problem.LocalPath != "" {
			t.Errorf("pathless problem acquired filesystem semantics: %+v", problem)
		}
	}
	for kind, wantID := range map[EntityKind]string{
		EntityTask: "6g0000000001", EntityEpic: "21-broken-epic",
		EntityAudit: "6g0000000002", EntityResearch: "6g0000000003",
		EntityThread: "6g0000000004",
	} {
		if got := byKind[kind]; got.EntityID != wantID || got.Message == "" {
			t.Errorf("%s problem = %+v; want id %q and a message", kind, got, wantID)
		}
	}
}

func TestLintAttributesDuplicateUnreadableOccurrencesByOpaqueLocation(t *testing.T) {
	const sameID = "6g0000000001"
	source := &lintSourceFake{researchProblems: []LoadProblem{
		{EntityKind: EntityResearch, EntityID: sameID, EntitySlug: "same-research", Location: "db://research/a", Message: "unreadable"},
		{EntityKind: EntityResearch, EntityID: sameID, EntitySlug: "same-research", Location: "db://research/b", Message: "unreadable"},
	}}
	results, problems, err := MustNewService(nil, WithLintSource(source)).Lint()
	if err != nil || len(problems) != 2 {
		t.Fatalf("lint problems = %+v, %v", problems, err)
	}
	seen := make(map[string]bool)
	for _, result := range results {
		if result.Slug == "same-research" {
			seen[result.Location] = true
		}
	}
	if len(seen) != 2 || !seen["db://research/a"] || !seen["db://research/b"] {
		t.Fatalf("unreadable duplicate occurrences = %v; all results: %+v", seen, results)
	}
}

func TestLintRequiresDedicatedReadCapability(t *testing.T) {
	_, _, err := MustNewService(nopStore{}).Lint()
	if err == nil || !strings.Contains(err.Error(), "lint reads are unavailable") {
		t.Fatalf("Lint error = %v; want missing lint capability", err)
	}
}

func TestLintRejectsMissingAuditSnapshotCapability(t *testing.T) {
	source := &lintSourceFake{}
	dropAuditCapability := func(s *Service) { s.auditReads = nil }

	_, _, err := MustNewService(nil, WithLintSource(source), dropAuditCapability).Lint()
	if err == nil || !strings.Contains(err.Error(), "audit snapshot reads are unavailable") {
		t.Fatalf("Lint error = %v; want missing audit snapshot capability", err)
	}
	if source.taskReads != 0 || source.epicReads != 0 || source.researchReads != 0 || source.auditReads != 0 {
		t.Fatalf("lint read before capability validation: %+v", source)
	}
}

func TestExplicitAuditSnapshotSourceWinsRegardlessOfOptionOrder(t *testing.T) {
	for _, tc := range []struct {
		name string
		opts func(*lintSourceFake, *auditSnapshotStub) []Option
	}{
		{"audit then lint", func(lint *lintSourceFake, audit *auditSnapshotStub) []Option {
			return []Option{WithAuditSnapshotSource(audit), WithLintSource(lint)}
		}},
		{"lint then audit", func(lint *lintSourceFake, audit *auditSnapshotStub) []Option {
			return []Option{WithLintSource(lint), WithAuditSnapshotSource(audit)}
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			broad := &lintSourceFake{}
			dedicated := &auditSnapshotStub{all: AuditSnapshot{Problems: []LoadProblem{{
				EntityKind: EntityAudit, EntityID: "6g0000000009", Message: "dedicated source",
			}}}}
			svc := MustNewService(nil, tc.opts(broad, dedicated)...)

			_, problems, err := svc.QueryFindings(FindingFilter{})
			if err != nil || len(problems) != 1 || problems[0].Message != "dedicated source" {
				t.Fatalf("QueryFindings problems = %+v, err = %v", problems, err)
			}
			if len(dedicated.calls) != 1 || dedicated.calls[0] != "" || broad.auditReads != 0 {
				t.Fatalf("audit reads = dedicated:%q broad:%d", dedicated.calls, broad.auditReads)
			}
		})
	}
}

func TestLintSourceSuppliesDefaultAuditSnapshotSource(t *testing.T) {
	source := &lintSourceFake{}
	if _, _, err := MustNewService(nil, WithLintSource(source)).QueryFindings(FindingFilter{}); err != nil {
		t.Fatal(err)
	}
	if source.auditReads != 1 {
		t.Fatalf("audit reads = %d, want one embedded default read", source.auditReads)
	}
}

func TestLintAttributesPathlessGraphDiagnosticsToReadableRecords(t *testing.T) {
	t.Run("dependency and lifecycle", func(t *testing.T) {
		prerequisite := graphRecord("pathless-prerequisite", domain.StatusNextUp)
		dependent := graphRecord("pathless-in-flight", domain.StatusInProgress, prerequisite.ID)
		invalid := graphRecord("pathless-invalid", domain.StatusReadyToStart, "not-a-stable-id")
		prerequisite.Path, dependent.Path, invalid.Path = "", "", ""

		results := lintTaskRecords(t, prerequisite, dependent, invalid)
		assertLintIssue(t, results, invalid.Slug, "depends_on", "not a stable task id")
		assertLintIssue(t, results, dependent.Slug, "status", "dependency gate")
	})

	t.Run("cycle", func(t *testing.T) {
		left := graphRecord("pathless-cycle-left", domain.StatusReadyToStart)
		right := graphRecord("pathless-cycle-right", domain.StatusReadyToStart, left.ID)
		left.DependsOn = []string{right.ID}
		left.Path, right.Path = "", ""

		results := lintTaskRecords(t, left, right)
		assertLintIssue(t, results, left.Slug, "depends_on", "dependency cycle")
		assertLintIssue(t, results, right.Slug, "depends_on", "dependency cycle")
	})

	t.Run("legacy declaration", func(t *testing.T) {
		prerequisite := graphRecord("pathless-legacy-prerequisite", domain.StatusCompleted)
		owner := graphRecord("pathless-legacy-owner", domain.StatusReadyToStart)
		owner.LegacyBlockedBy = []string{prerequisite.Slug}
		owner.LegacyDependencyFields = []string{"blocked_by"}
		prerequisite.Path, owner.Path = "", ""

		results := lintTaskRecords(t, prerequisite, owner)
		assertLintIssue(t, results, owner.Slug, "blocked_by", "legacy dependency field")
	})
}

func TestLintRecordAttributionDoesNotCollideOnIDOrLocation(t *testing.T) {
	first := graphRecord("portable-duplicate-first", domain.StatusReadyToStart)
	second := graphRecord("portable-duplicate-second", domain.StatusReadyToStart, "bad-reference")
	second.ID, second.FilenameID = first.ID, first.FilenameID
	// An opaque or contradictory location is context, not the record join key.
	first.Path, second.Path = "opaque://same", "opaque://same"

	results := lintTaskRecords(t, first, second)
	assertLintIssue(t, results, first.Slug, "id", "duplicate stable task id")
	assertLintIssue(t, results, second.Slug, "id", "duplicate stable task id")
	assertLintIssue(t, results, second.Slug, "depends_on", "bad-reference")
	if lintResultHas(results, first.Slug, "depends_on", "bad-reference") {
		t.Fatalf("second record dependency defect leaked onto first record: %+v", results)
	}
}

func TestLintDistinguishesEqualReadableRecordsByOpaqueLocation(t *testing.T) {
	taskID := testutil.TaskID("lint-opaque-task")
	auditID := testutil.TaskID("lint-opaque-audit")
	researchID := testutil.TaskID("lint-opaque-research")
	task := TaskWithBody{Task: domain.Task{ID: taskID, Slug: "same-task", Status: domain.StatusReadyToStart}}
	epic := domain.Epic{ID: "21-same-epic", Status: "active"}
	audit := AuditWithFindings{Audit: domain.Audit{ID: auditID, Slug: "same-audit", Bucket: domain.AuditOpen}}
	research := domain.Research{ID: researchID, Slug: "same-research"}
	source := &locationLintSource{}
	for _, suffix := range []string{"a", "b"} {
		source.tasks = append(source.tasks, LoadedRecord[TaskWithBody]{Value: task, Source: RecordSource{ID: taskID, Location: "db://tasks/" + suffix}})
		source.epics = append(source.epics, LoadedRecord[domain.Epic]{Value: epic, Source: RecordSource{ID: epic.ID, Location: "db://epics/" + suffix}})
		source.audits = append(source.audits, LoadedRecord[AuditWithFindings]{Value: audit, Source: RecordSource{ID: auditID, Location: "db://audits/" + suffix}})
		source.research = append(source.research, LoadedRecord[domain.Research]{Value: research, Source: RecordSource{ID: researchID, Location: "db://research/" + suffix}})
	}
	results, _, err := MustNewService(nil, WithLintSource(source)).Lint()
	if err != nil {
		t.Fatal(err)
	}
	for label, locations := range map[string][]string{
		"same-task":     {"db://tasks/a", "db://tasks/b"},
		"21-same-epic":  {"db://epics/a", "db://epics/b"},
		"same-audit":    {"db://audits/a", "db://audits/b"},
		"same-research": {"db://research/a", "db://research/b"},
	} {
		seen := make(map[string]bool)
		for _, result := range results {
			if result.Slug == label {
				seen[result.Location] = true
			}
		}
		if len(seen) != 2 || !seen[locations[0]] || !seen[locations[1]] {
			t.Errorf("%s lint occurrence locations = %v, want %v", label, seen, locations)
		}
	}
}

func TestLintUsesPathlessUnreadableIdentityInLifecycleDiagnosis(t *testing.T) {
	unreadableID := "6g0000000005"
	dependent := graphRecord("depends-on-pathless-unreadable", domain.StatusInProgress, unreadableID)
	dependent.Path = ""
	source := &lintSourceFake{
		taskRecords: []TaskWithBody{{Task: dependent}},
		taskProblems: []LoadProblem{{
			EntityKind: EntityTask, EntityID: unreadableID,
			EntitySlug: "unreadable", Message: "remote decode failed",
		}},
	}

	results, problems, err := MustNewService(nil, WithLintSource(source)).Lint()
	if err != nil {
		t.Fatal(err)
	}
	assertLintIssue(t, results, dependent.Slug, "status", unreadableID)
	if len(problems) != 1 || problems[0].EntityID != unreadableID || problems[0].Location != "" {
		t.Fatalf("load problems = %+v", problems)
	}
}

func lintTaskRecords(t *testing.T, tasks ...domain.Task) []LintResult {
	t.Helper()
	records := make([]TaskWithBody, len(tasks))
	for index, task := range tasks {
		records[index] = TaskWithBody{Task: task}
	}
	results, _, err := MustNewService(nil, WithLintSource(&lintSourceFake{taskRecords: records})).Lint()
	if err != nil {
		t.Fatal(err)
	}
	return results
}

func assertLintIssue(t *testing.T, results []LintResult, slug, field, messagePart string) {
	t.Helper()
	if !lintResultHas(results, slug, field, messagePart) {
		t.Fatalf("missing %s issue containing %q for %s in %+v", field, messagePart, slug, results)
	}
}

func lintResultHas(results []LintResult, slug, field, messagePart string) bool {
	for _, result := range results {
		if result.Slug != slug {
			continue
		}
		for _, issue := range result.Issues {
			if issue.Field == field && strings.Contains(issue.Message, messagePart) {
				return true
			}
		}
	}
	return false
}

func TestLintResearchComparesDeclarationWithAdapterSourceID(t *testing.T) {
	const sourceID = "6g0000000001"
	source := &locationLintSource{research: []LoadedRecord[domain.Research]{{
		Value:  domain.Research{ID: "6g0000000002", Slug: "drifted", Created: "2026-09-01"},
		Source: RecordSource{ID: sourceID, Location: "db://research/one"},
	}}}
	results, problems, err := MustNewService(nil, WithLintSource(source)).Lint()
	if err != nil || len(problems) != 0 {
		t.Fatalf("lint err=%v problems=%+v", err, problems)
	}
	assertLintIssue(t, results, "drifted", "id", sourceID)
	if results[0].Location != "db://research/one" {
		t.Fatalf("lint location = %q", results[0].Location)
	}
}

func TestLintAuditComparesDeclarationWithAdapterSourceID(t *testing.T) {
	const sourceID = "6g0000000001"
	source := &locationLintSource{audits: []LoadedRecord[AuditWithFindings]{{
		Value:  AuditWithFindings{Audit: domain.Audit{ID: "6g0000000002", Slug: "drifted", Bucket: domain.AuditOpen}},
		Source: RecordSource{ID: sourceID, Location: "db://audits/one"},
	}}}
	results, problems, err := MustNewService(nil, WithLintSource(source)).Lint()
	if err != nil || len(problems) != 0 {
		t.Fatalf("lint err=%v problems=%+v", err, problems)
	}
	assertLintIssue(t, results, "drifted", "id", sourceID)
	if results[0].Location != "db://audits/one" {
		t.Fatalf("lint location = %q", results[0].Location)
	}
}
