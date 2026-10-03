package core

import (
	"errors"
	"testing"

	"github.com/andy-esch/taskflow/internal/domain"
)

type sourceLessEntityStore struct{ fakeStore }

func (*sourceLessEntityStore) ReadTask(string) (LoadedRecord[TaskWithBody], error) {
	return LoadedRecord[TaskWithBody]{Value: TaskWithBody{Task: domain.Task{ID: "declared", Slug: "task"}}}, nil
}

func (*sourceLessEntityStore) ReadEpics() (EpicRead, error) {
	return EpicRead{Records: []LoadedRecord[domain.Epic]{{
		Value: domain.Epic{ID: "21-declared"}, Source: RecordSource{Location: "db://epics/misleading"},
	}}}, nil
}

func (*sourceLessEntityStore) ReadEpic(string) (LoadedRecord[EpicWithBody], error) {
	return LoadedRecord[EpicWithBody]{Value: EpicWithBody{Epic: domain.Epic{ID: "21-declared"}}}, nil
}

func (*sourceLessEntityStore) ReadAudits() (AuditRead, error) {
	return AuditRead{Records: []LoadedRecord[domain.Audit]{{
		Value:  domain.Audit{ID: "declared", Slug: "audit", Bucket: domain.AuditOpen},
		Source: RecordSource{Location: "db://audits/misleading"},
	}}}, nil
}

func (*sourceLessEntityStore) ReadAudit(string) (LoadedRecord[AuditWithBody], error) {
	return LoadedRecord[AuditWithBody]{Value: AuditWithBody{Audit: domain.Audit{ID: "declared", Slug: "audit"}}}, nil
}

func (*sourceLessEntityStore) ReadResearch() (ResearchRead, error) {
	return ResearchRead{Records: []LoadedRecord[domain.Research]{{
		Value:  domain.Research{ID: "declared", Slug: "research"},
		Source: RecordSource{Location: "db://research/misleading"},
	}}}, nil
}

func (*sourceLessEntityStore) ReadResearchDocument(string) (LoadedRecord[ResearchWithBody], error) {
	return LoadedRecord[ResearchWithBody]{Value: ResearchWithBody{Research: domain.Research{ID: "declared", Slug: "research"}}}, nil
}

func TestOrdinaryReadServicesDoNotPublishSourceLessRecords(t *testing.T) {
	svc := MustNewService(&sourceLessEntityStore{})
	epics, epicProblems, err := svc.ListEpics()
	if err != nil || len(epics) != 0 || len(epicProblems) != 1 || epicProblems[0].EntityKind != EntityEpic ||
		epicProblems[0].EntityID != "" || epicProblems[0].EntitySlug != "21-declared" || epicProblems[0].Location != "db://epics/misleading" {
		t.Fatalf("epic list = %+v, %+v, %v", epics, epicProblems, err)
	}
	audits, auditProblems, err := svc.ListAudits("", false)
	if err != nil || len(audits) != 0 || len(auditProblems) != 1 || auditProblems[0].EntityKind != EntityAudit ||
		auditProblems[0].EntityID != "" || auditProblems[0].EntitySlug != "audit" {
		t.Fatalf("audit list = %+v, %+v, %v", audits, auditProblems, err)
	}
	research, researchProblems, err := svc.ListResearch("")
	if err != nil || len(research) != 0 || len(researchProblems) != 1 || researchProblems[0].EntityKind != EntityResearch ||
		researchProblems[0].EntityID != "" || researchProblems[0].EntitySlug != "research" {
		t.Fatalf("research list = %+v, %+v, %v", research, researchProblems, err)
	}
	for name, show := range map[string]func() error{
		"task":     func() error { _, err := svc.ShowTask("task"); return err },
		"epic":     func() error { _, err := svc.ShowEpic("epic"); return err },
		"audit":    func() error { _, err := svc.ShowAudit("audit"); return err },
		"research": func() error { _, err := svc.ShowResearch("research"); return err },
	} {
		if err := show(); !errors.Is(err, domain.ErrValidation) {
			t.Errorf("%s show with missing source ID = %v, want validation error", name, err)
		}
	}
}

func TestThreadReadValidateSourcesFailsClosedWithoutCanonicalIdentity(t *testing.T) {
	read := ThreadRead{Records: []VersionedRecord[domain.Thread]{
		{Record: LoadedRecord[domain.Thread]{Value: domain.Thread{ID: "canonical-a"}, Source: RecordSource{ID: "canonical-a"}}},
		{Record: LoadedRecord[domain.Thread]{Value: domain.Thread{ID: "canonical-b"}, Source: RecordSource{ID: "canonical-b"}}},
	}}
	if err := read.ValidateSources(); err != nil {
		t.Fatalf("unique canonical sources = %v", err)
	}
	read.Records[0].SourceVersion = "guarded-revision"
	if threads := read.SemanticThreads(); len(threads) != 2 || threads[0].ID != "canonical-a" {
		t.Fatalf("semantic Thread projection changed records: %+v", threads)
	}
	read.Records[1].Record.Value.ID = "drifted-declaration"
	if err := read.ValidateSources(); !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("declared/source identity drift = %v, want validation error", err)
	}
	read.Records[1].Record.Value.ID = "canonical-b"
	read.Records[1].Record.Source.ID = ""
	read.Records[1].Record.Source.Location = "remote://thread/declared-b"
	if err := read.ValidateSources(); !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("missing canonical source ID = %v, want validation error", err)
	}
	read.Records[1].Record.Source.ID = "canonical-a"
	if err := read.ValidateSources(); !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("duplicate canonical source ID = %v, want validation error", err)
	}
}
