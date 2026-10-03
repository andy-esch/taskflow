package core

import (
	"errors"
	"fmt"
	"slices"
	"testing"
	"time"

	"github.com/andy-esch/taskflow/internal/domain"
)

type maintenanceStore struct {
	fakeStore
	events    []string
	fixes     []domain.FixResult
	fixErr    error
	headerErr error
	lintErr   error
	links     []LoadProblem
	linkErr   error
}

type headerMaintenanceStore struct {
	maintenanceStore
	bodies  map[string]string
	failRef string
}

func (f *headerMaintenanceStore) ReadAuditSnapshot(string) (AuditSnapshot, error) {
	f.events = append(f.events, "audit snapshot")
	var snapshot AuditSnapshot
	// Explicit ordering makes the failure-prefix assertion independent of maps.
	for _, ref := range []string{"first", "second"} {
		body := f.bodies[ref]
		snapshot.Audits = append(snapshot.Audits, LoadedRecord[AuditWithFindings]{
			Source: RecordSource{ID: "opaque-" + ref},
			Value: AuditWithFindings{
				Audit: domain.Audit{Slug: ref}, NearMisses: domain.NearMissFindingHeaders(body),
			},
		})
	}
	return snapshot, nil
}

func (f *headerMaintenanceStore) TransformAuditBody(ref string, _ time.Time, dryRun bool, transform func(domain.Audit, string) (string, error)) (domain.Audit, string, bool, error) {
	f.events = append(f.events, "body/"+ref)
	audit := domain.Audit{Slug: ref}
	if ref == f.failRef {
		return audit, "", false, f.headerErr
	}
	current := f.bodies[ref]
	next, err := transform(audit, current)
	if err == nil && !dryRun {
		f.bodies[ref] = next
	}
	return audit, next, current != next, err
}

func TestRepairPlanningRetainsActualBodyRepairPrefix(t *testing.T) {
	for _, dryRun := range []bool{false, true} {
		t.Run(fmt.Sprintf("dry-run=%t", dryRun), func(t *testing.T) {
			body := "## Findings\n\n#### H-1. fix me · **Status:** open\n"
			sentinel := errors.New("second body unavailable")
			fake := &headerMaintenanceStore{
				maintenanceStore: maintenanceStore{
					fixes: []domain.FixResult{{Changes: []string{"ordinary field repaired"}}}, headerErr: sentinel,
				},
				bodies: map[string]string{"first": body, "second": body}, failRef: "second",
			}
			receipt, err := MustNewService(fake).RepairPlanning(dryRun)
			want := []string{fmt.Sprintf("frontmatter/%t", dryRun), "audit snapshot", "body/first", "body/second"}
			if !errors.Is(err, sentinel) || len(receipt.Fixes) != 2 || len(receipt.Fixes[1].Changes) != 1 || !slices.Equal(fake.events, want) {
				t.Fatalf("repair=%+v err=%v events=%v want=%v", receipt, err, fake.events, want)
			}
			if receipt.Fixes[1].Path != "" || fake.bodies["second"] != body {
				t.Fatal("repair required a path or touched the failed source")
			}
			if dryRun {
				if fake.bodies["first"] != body {
					t.Fatal("dry-run changed the source")
				}
			} else if len(domain.ParseFindings(fake.bodies["first"])) != 1 {
				t.Fatal("completed prefix did not canonicalize the first body")
			}
		})
	}
}

func (f *maintenanceStore) FixFrontmatter(dryRun bool) ([]domain.FixResult, error) {
	f.events = append(f.events, fmt.Sprintf("frontmatter/%t", dryRun))
	return f.fixes, f.fixErr
}

func (f *maintenanceStore) ReadAuditSnapshot(selector string) (AuditSnapshot, error) {
	f.events = append(f.events, "audit snapshot")
	if f.headerErr != nil {
		return AuditSnapshot{}, f.headerErr
	}
	return f.fakeStore.ReadAuditSnapshot(selector)
}

func (f *maintenanceStore) ReadLintTasks() ([]LoadedRecord[TaskWithBody], []LoadProblem, error) {
	f.events = append(f.events, "lint tasks")
	if f.lintErr != nil {
		return nil, nil, f.lintErr
	}
	return f.fakeStore.ReadLintTasks()
}

func (f *maintenanceStore) DanglingLinks() ([]LoadProblem, error) {
	f.events = append(f.events, "body links")
	return f.links, f.linkErr
}

func TestRepairPlanningOwnsOrderingAndDryRun(t *testing.T) {
	for _, dry := range []bool{true, false} {
		t.Run(fmt.Sprintf("dry-run=%t", dry), func(t *testing.T) {
			fake := &maintenanceStore{fixes: []domain.FixResult{{Changes: []string{"repaired ordinary field"}}}}
			receipt, err := MustNewService(fake).RepairPlanning(dry)
			if err != nil || len(receipt.Fixes) != 1 || len(receipt.LintResults)+len(receipt.Problems) != 0 {
				t.Fatalf("repair=%+v err=%v", receipt, err)
			}
			want := []string{fmt.Sprintf("frontmatter/%t", dry), "audit snapshot"}
			if !dry {
				want = append(want, "lint tasks", "audit snapshot")
			}
			if !slices.Equal(fake.events, want) {
				t.Fatalf("events=%v want=%v", fake.events, want)
			}
		})
	}
}

func TestRepairPlanningRetainsPrefixAndDoesNotRetryFailures(t *testing.T) {
	sentinel := errors.New("later phase failed")
	for _, stage := range []string{"frontmatter", "headers", "post-lint"} {
		t.Run(stage, func(t *testing.T) {
			fake := &maintenanceStore{fixes: []domain.FixResult{{Changes: []string{"already repaired"}}}}
			var want []string
			switch stage {
			case "frontmatter":
				fake.fixErr = sentinel
				want = []string{"frontmatter/false"}
			case "headers":
				fake.headerErr = sentinel
				want = []string{"frontmatter/false", "audit snapshot"}
			case "post-lint":
				fake.lintErr = sentinel
				want = []string{"frontmatter/false", "audit snapshot", "lint tasks"}
			}
			receipt, err := MustNewService(fake).RepairPlanning(false)
			if !errors.Is(err, sentinel) || len(receipt.Fixes) != 1 || !slices.Equal(fake.events, want) {
				t.Fatalf("repair=%+v err=%v events=%v want=%v", receipt, err, fake.events, want)
			}
		})
	}
}

func TestLintWithLinksUsesOptionalCapabilityOnceAfterLint(t *testing.T) {
	problem := LoadProblem{EntityKind: EntityTask, EntityID: "opaque-key", Location: "db://records/problem", Message: "unresolved body reference"}
	fake := &maintenanceStore{links: []LoadProblem{problem}}
	svc := MustNewService(fake)
	_, plainProblems, err := svc.Lint()
	if err != nil || len(plainProblems) != 0 || slices.Contains(fake.events, "body links") {
		t.Fatalf("ordinary lint invoked opt-in links: problems=%v err=%v events=%v", plainProblems, err, fake.events)
	}
	fake.events = nil
	_, problems, err := svc.LintWithLinks()
	if err != nil || len(problems) != 1 || problems[0] != problem || !slices.Equal(fake.events, []string{"lint tasks", "audit snapshot", "body links"}) {
		t.Fatalf("linked lint problems=%v err=%v events=%v", problems, err, fake.events)
	}
	fake.events = nil
	fake.lintErr = errors.New("lint snapshot failed")
	if _, _, err := svc.LintWithLinks(); !errors.Is(err, fake.lintErr) || slices.Contains(fake.events, "body links") {
		t.Fatalf("links ran after failed lint: err=%v events=%v", err, fake.events)
	}
	fake.lintErr = nil
	fake.linkErr = errors.New("link snapshot failed")
	if _, _, err := svc.LintWithLinks(); !errors.Is(err, fake.linkErr) {
		t.Fatalf("link error lost: %v", err)
	}
}

func TestMaintenanceMissingCapabilitiesFailBeforeMutation(t *testing.T) {
	fake := &maintenanceStore{}
	if _, err := MustNewService(nil, WithFrontmatterRepairer(fake)).RepairPlanning(false); !errors.Is(err, domain.ErrValidation) || len(fake.events) != 0 {
		t.Fatalf("incomplete workflow mutated: err=%v events=%v", err, fake.events)
	}
	if _, err := MustNewService(&fakeStore{}).RepairPlanning(true); !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("missing repair = %v", err)
	}
	if _, _, err := MustNewService(&fakeStore{}).LintWithLinks(); !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("missing link checks = %v", err)
	}
	// This store has a write boundary but no body-aware or lint capability.
	if _, err := MustNewService(nopStore{}, WithFrontmatterRepairer(fake)).RepairPlanning(true); !errors.Is(err, domain.ErrValidation) || len(fake.events) != 0 {
		t.Fatalf("missing audit reads mutated: err=%v events=%v", err, fake.events)
	}
	svc := MustNewService(nopStore{}, WithFrontmatterRepairer(fake), WithAuditSnapshotSource(fake))
	if _, err := svc.RepairPlanning(false); !errors.Is(err, domain.ErrValidation) || len(fake.events) != 0 {
		t.Fatalf("missing post-lint mutated: err=%v events=%v", err, fake.events)
	}
	// A dry-run needs no prospective-state lint capability.
	if _, err := svc.RepairPlanning(true); err != nil || !slices.Equal(fake.events, []string{"frontmatter/true", "audit snapshot"}) {
		t.Fatalf("dry-run incorrectly required lint: err=%v events=%v", err, fake.events)
	}
}

func TestMaintenanceTypedNilOptionsStayAbsent(t *testing.T) {
	var fake *maintenanceStore
	var completion *completionStub
	svc := MustNewService(nil, WithFrontmatterRepairer(fake), WithBodyLinkSource(fake), WithCompletionSource(completion))
	if svc.frontmatterRepairs != nil || svc.bodyLinks != nil || svc.completions != nil {
		t.Fatal("typed-nil maintenance options were selected")
	}
}
