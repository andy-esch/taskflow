package core

import (
	"reflect"
	"strings"
	"testing"

	"github.com/andy-esch/taskflow/internal/domain"
)

func TestLintCrossKindIdentityIncludesUnreadableRecords(t *testing.T) {
	const sharedID = "6g0000000001"
	for _, taskReadable := range []bool{false, true} {
		for _, threadReadable := range []bool{false, true} {
			name := "task-unreadable/thread-unreadable"
			if taskReadable {
				name = strings.Replace(name, "task-unreadable", "task-readable", 1)
			}
			if threadReadable {
				name = strings.Replace(name, "thread-unreadable", "thread-readable", 1)
			}
			t.Run(name, func(t *testing.T) {
				source := &lintSourceFake{}
				threads := &threadReadFake{}
				if taskReadable {
					source.taskRecords = []LoadedRecord[TaskWithBody]{{
						Value:  TaskWithBody{Task: domain.Task{ID: sharedID, Slug: "task", Status: domain.StatusCompleted}},
						Source: RecordSource{ID: sharedID},
					}}
				} else {
					source.taskProblems = []LoadProblem{{EntityKind: EntityTask, EntityID: sharedID, EntitySlug: "task", Message: "bad task"}}
				}
				if threadReadable {
					threads.threads = []domain.Thread{{ID: sharedID, Slug: "thread", Status: domain.ThreadStatusUnstarted,
						Description: "A Thread", Goal: "Ship it", Created: "2026-10-04", Tasks: []string{sharedID}}}
				} else {
					threads.problems = []ThreadReadProblem{{ThreadID: sharedID, ThreadSlug: "thread", Message: "bad Thread"}}
				}
				results, problems, err := MustNewService(nil, WithLintSource(source), WithThreadStore(threads)).Lint()
				if err != nil {
					t.Fatal(err)
				}
				wantProblems := 0
				if !taskReadable {
					wantProblems++
				}
				if !threadReadable {
					wantProblems++
				}
				if len(problems) != wantProblems {
					t.Fatalf("load problems = %+v; want %d preserved", problems, wantProblems)
				}
				if source.taskReads != 1 || source.epicReads != 1 || source.auditReads != 1 || source.researchReads != 1 {
					t.Fatalf("collision detection rescanned lint sources: %+v", source)
				}
				for _, result := range results {
					if result.Location != "" {
						t.Fatalf("pathless identity acquired a location: %+v", result)
					}
				}
				assertCrossKindLintCount(t, results, "task", sharedID, "Thread", 1)
				assertCrossKindLintCount(t, results, "thread", sharedID, "task", 1)
				if threadReadable && !taskReadable {
					// Identity ownership is not proof that an unreadable member is valid.
					assertLintIssue(t, results, "thread", "tasks", "unknown member")
				}
			})
		}
	}
}

func TestLintCrossKindIdentityDoesNotInferOrAcceptMalformedRecoveredIDs(t *testing.T) {
	for _, recoveredID := range []string{"", " ", "bad-id", "6G0000000001", "6g000000000i", " 6g0000000001 "} {
		t.Run("id="+recoveredID, func(t *testing.T) {
			// Even a matching declared ID cannot make a malformed recovered ID authoritative.
			readableID := recoveredID
			if strings.TrimSpace(readableID) == "" {
				readableID = "6g0000000001"
			}
			for _, taskReadable := range []bool{false, true} {
				source := &lintSourceFake{}
				threads := &threadReadFake{}
				if taskReadable {
					source.taskRecords = []LoadedRecord[TaskWithBody]{{
						Value:  TaskWithBody{Task: domain.Task{ID: readableID, Slug: "task", Status: domain.StatusCompleted}},
						Source: RecordSource{ID: "6g0000000002"},
					}}
					threads.problems = []ThreadReadProblem{{ThreadID: recoveredID, ThreadSlug: "thread",
						Location: "db://threads/6g0000000001-broken.md", Message: "declared id: 6g0000000001"}}
				} else {
					source.taskProblems = []LoadProblem{{EntityKind: EntityTask, EntityID: recoveredID,
						EntitySlug: "task", Location: "db://tasks/6g0000000001-broken.md",
						LocalPath: "tasks/6g0000000001-broken.md", Message: "declared id: 6g0000000001"}}
					threads.threads = []domain.Thread{{ID: readableID, Slug: "thread"}}
					threads.recordID = "6g0000000003"
				}
				results, _, err := MustNewService(nil, WithLintSource(source), WithThreadStore(threads)).Lint()
				if err != nil {
					t.Fatal(err)
				}
				for _, result := range results {
					for _, issue := range result.Issues {
						if strings.Contains(issue.Message, "is also used by") {
							t.Fatalf("inferred collision for recovered ID %q: %+v", recoveredID, results)
						}
					}
				}
			}
		})
	}
}

func TestLintCrossKindUnreadableOccurrencesCombineDuplicateIssues(t *testing.T) {
	const sharedID = "6g0000000001"
	source := &lintSourceFake{taskProblems: []LoadProblem{{EntityKind: EntityTask,
		EntityID: sharedID, EntitySlug: "task", Location: "db://tasks/a", LocalPath: "optional-repair.md", Message: "bad task"}}}
	threads := &threadReadFake{problems: []ThreadReadProblem{
		{ThreadID: sharedID, ThreadSlug: "same-thread", Location: "db://threads/z", Message: "bad Thread z"},
		{ThreadID: sharedID, ThreadSlug: "same-thread", Location: "db://threads/a", Message: "bad Thread a"},
	}}
	svc := MustNewService(nil, WithLintSource(source), WithThreadStore(threads))
	results, problems, err := svc.Lint()
	if err != nil || len(results) != 3 || len(problems) != 3 {
		t.Fatalf("lint = %+v, %+v, %v; want three distinct occurrences", results, problems, err)
	}
	for i, wantLocation := range []string{"db://tasks/a", "db://threads/z", "db://threads/a"} {
		result := results[i]
		if result.Location != wantLocation {
			t.Fatalf("occurrence %d location = %q, want %q", i, result.Location, wantLocation)
		}
		if i == 0 {
			if len(result.Issues) != 1 {
				t.Fatalf("task issues = %+v", result.Issues)
			}
		} else if len(result.Issues) != 2 || !strings.Contains(result.Issues[0].Message, "duplicate stable id") ||
			!strings.Contains(result.Issues[1].Message, "also used by a task") {
			t.Fatalf("Thread duplicate and collision must share one result: %+v", result)
		}
	}
	for range 10 {
		again, againProblems, err := svc.Lint()
		if err != nil || !reflect.DeepEqual(results, again) || !reflect.DeepEqual(problems, againProblems) {
			t.Fatalf("non-deterministic lint: %+v, %+v, %v", again, againProblems, err)
		}
	}
}

func TestLintCrossKindRecoveredIDsMatchReadableSourceAndDeclaredIdentity(t *testing.T) {
	const sharedID = "6g0000000001"
	const otherID = "6g0000000002"
	for _, taskReadable := range []bool{false, true} {
		for _, sourceMatches := range []bool{false, true} {
			sourceID, declaredID := otherID, sharedID
			if sourceMatches {
				sourceID, declaredID = sharedID, otherID
			}
			source := &lintSourceFake{}
			threads := &threadReadFake{}
			if taskReadable {
				source.taskRecords = []LoadedRecord[TaskWithBody]{{
					Value:  TaskWithBody{Task: domain.Task{ID: declaredID, Slug: "task", Status: domain.StatusCompleted}},
					Source: RecordSource{ID: sourceID},
				}}
				threads.problems = []ThreadReadProblem{{ThreadID: sharedID, ThreadSlug: "thread", Message: "bad Thread"}}
			} else {
				source.taskProblems = []LoadProblem{{EntityKind: EntityTask, EntityID: sharedID, EntitySlug: "task", Message: "bad task"}}
				threads.threads = []domain.Thread{{ID: declaredID, Slug: "thread"}}
				threads.recordID = sourceID
			}
			results, _, err := MustNewService(nil, WithLintSource(source), WithThreadStore(threads)).Lint()
			if err != nil {
				t.Fatal(err)
			}
			assertCrossKindLintCount(t, results, "task", sharedID, "Thread", 1)
			assertCrossKindLintCount(t, results, "thread", sharedID, "task", 1)
			readableSlug := "thread"
			if taskReadable {
				readableSlug = "task"
			}
			assertLintIssue(t, results, readableSlug, "id", "disagrees with the filename id")
		}
	}
}

func TestLintCrossKindDistinctRecoveredIDsDoNotCollide(t *testing.T) {
	source := &lintSourceFake{taskProblems: []LoadProblem{{EntityKind: EntityTask,
		EntityID: "6g0000000001", EntitySlug: "task", Message: "bad task"}}}
	threads := &threadReadFake{problems: []ThreadReadProblem{{
		ThreadID: "6g0000000002", ThreadSlug: "thread", Message: "bad Thread"}}}
	results, problems, err := MustNewService(nil, WithLintSource(source), WithThreadStore(threads)).Lint()
	if err != nil || len(problems) != 2 || len(results) != 0 {
		t.Fatalf("distinct unreadable identities produced lint issues: %+v, %+v, %v", results, problems, err)
	}
}

func TestLintCrossKindIdentityDoesNotPromoteUntrustedDeclaredID(t *testing.T) {
	const declaredID = "6g0000000001"
	source := &lintSourceFake{taskRecords: []LoadedRecord[TaskWithBody]{{
		Value:  TaskWithBody{Task: domain.Task{ID: declaredID, Slug: "id-less-source", Status: domain.StatusCompleted}},
		Source: RecordSource{Location: "db://tasks/" + declaredID},
	}}}
	threads := &threadReadFake{threads: []domain.Thread{{ID: declaredID, Slug: "thread"}}}
	results, problems, err := MustNewService(nil, WithLintSource(source), WithThreadStore(threads)).Lint()
	if err != nil || len(problems) != 1 || problems[0].EntityID != "" {
		t.Fatalf("lint problems = %+v, %v; want untrusted declared ID withheld", problems, err)
	}
	assertCrossKindLintCount(t, results, "id-less-source", declaredID, "Thread", 0)
	assertCrossKindLintCount(t, results, "thread", declaredID, "task", 0)
}

func assertCrossKindLintCount(t *testing.T, results []LintResult, slug, stableID, otherKind string, want int) {
	t.Helper()
	count := 0
	for _, result := range results {
		if result.Slug != slug {
			continue
		}
		for _, issue := range result.Issues {
			if issue.Field == "id" && strings.Contains(issue.Message, "stable id "+stableID+" is also used by a "+otherKind) {
				count++
			}
		}
	}
	if count != want {
		t.Fatalf("%s collision count = %d, want %d: %+v", slug, count, want, results)
	}
}
