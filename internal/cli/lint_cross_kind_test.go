package cli

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/andy-esch/taskflow/internal/testutil"
	"github.com/andy-esch/taskflow/internal/wire"
)

func TestLintCrossKindUnreadableHumanAndJSON(t *testing.T) {
	const sharedID = "6g0000000001"
	for _, taskReadable := range []bool{false, true} {
		for _, threadReadable := range []bool{false, true} {
			t.Run(fmt.Sprintf("task-readable=%t/thread-readable=%t", taskReadable, threadReadable), func(t *testing.T) {
				r := testutil.NewRepo(t)
				taskContent := "---\nid: [unterminated\n---\n# Task\n"
				threadContent := "---\nid: [unterminated\n---\n# Thread\n"
				if taskReadable {
					taskContent = "---\nid: " + sharedID + "\nstatus: completed\n---\n# Task\n"
				}
				if threadReadable {
					threadContent = "---\nid: " + sharedID + "\nstatus: unstarted\ndescription: A Thread\ngoal: Ship it\ncreated: 2026-10-04\ntasks: []\n---\n# Thread\n"
				}
				r.File("tasks/"+sharedID+"-task.md", taskContent)
				r.File("threads/"+sharedID+"-thread.md", threadContent)
				human, err := runRootRC(t, "-C", r.Root, "lint", "--color=never")
				if ExitCode(err) != 11 {
					t.Fatalf("human exit = %v; want validation failure:\n%s", err, human)
				}
				for _, otherKind := range []string{"task", "Thread"} {
					message := "stable id " + sharedID + " is also used by a " + otherKind
					if count := strings.Count(human, message); count != 1 {
						t.Fatalf("human %q count = %d, want one:\n%s", message, count, human)
					}
				}
				jsonOut, err := runRootRC(t, "-C", r.Root, "lint", "--json")
				if ExitCode(err) != 11 {
					t.Fatalf("JSON exit = %v; want validation failure:\n%s", err, jsonOut)
				}
				var envelope wire.LintEnvelope
				if err := json.Unmarshal([]byte(jsonOut), &envelope); err != nil {
					t.Fatalf("decode lint JSON: %v\n%s", err, jsonOut)
				}
				wantProblems := 0
				if !taskReadable {
					wantProblems++
				}
				if !threadReadable {
					wantProblems++
				}
				if envelope.SchemaVersion != wire.SchemaVersion || len(envelope.Unreadable) != wantProblems || len(envelope.Issues) != 2 {
					t.Fatalf("lint envelope = %+v; want two collision owners and %d load problems", envelope, wantProblems)
				}
				for i, slug := range []string{"task", "thread"} {
					result := envelope.Issues[i]
					if result.Slug != slug || result.Location != "" || len(result.Issues) != 1 || result.Issues[0].Field != "id" {
						t.Fatalf("lint issue %d = %+v; want one collision for %q", i, result, slug)
					}
				}
				for _, problem := range envelope.Unreadable {
					if problem.EntityID != sharedID || problem.Path == "" || problem.Message == "" {
						t.Fatalf("original load problem lost identity, repair path, or error: %+v", problem)
					}
				}
				for range 3 {
					againHuman, err := runRootRC(t, "-C", r.Root, "lint", "--color=never")
					if ExitCode(err) != 11 || human != againHuman {
						t.Fatalf("human lint changed between runs: %v\n%s", err, againHuman)
					}
					againJSON, err := runRootRC(t, "-C", r.Root, "lint", "--json")
					if ExitCode(err) != 11 || jsonOut != againJSON {
						t.Fatalf("JSON lint changed between runs: %v\n%s", err, againJSON)
					}
				}
			})
		}
	}
}

func TestLintCrossKindDoesNotRecoverIDsFromInvalidFilenames(t *testing.T) {
	const sharedID = "6g0000000001"
	for _, brokenKind := range []string{"tasks", "threads"} {
		for _, name := range []string{"not-id-led.md", "6g000000000i-broken.md", "6G0000000001-broken.md"} {
			t.Run(brokenKind+"/"+name, func(t *testing.T) {
				r := testutil.NewRepo(t)
				// The apparent frontmatter ID is not authoritative without a valid source name.
				r.File(brokenKind+"/"+name, "---\nid: "+sharedID+"\nstatus: [unterminated\n---\n# Broken\n")
				if brokenKind == "tasks" {
					r.File("threads/"+sharedID+"-thread.md", "---\nid: "+sharedID+"\nstatus: unstarted\ndescription: A Thread\ngoal: Ship it\ncreated: 2026-10-04\n---\n# Thread\n")
				} else {
					r.File("tasks/"+sharedID+"-task.md", "---\nid: "+sharedID+"\nstatus: completed\n---\n# Task\n")
				}
				out, err := runRootRC(t, "-C", r.Root, "lint", "--json")
				if ExitCode(err) != 11 {
					t.Fatalf("lint exit = %v; malformed source must remain visible:\n%s", err, out)
				}
				var envelope wire.LintEnvelope
				if err := json.Unmarshal([]byte(out), &envelope); err != nil {
					t.Fatal(err)
				}
				if len(envelope.Unreadable) != 1 || envelope.Unreadable[0].EntityID != "" || strings.Contains(out, "is also used by") {
					t.Fatalf("untrusted identity inferred from source or body:\n%s", out)
				}
			})
		}
	}
}

func TestLintCrossKindDuplicateAndCollisionOutputIsStable(t *testing.T) {
	r := testutil.NewRepo(t)
	// Reverse creation order deliberately; the real source scan orders occurrences.
	for _, id := range []string{"6g0000000002", "6g0000000001"} {
		r.File("tasks/"+id+"-task-"+id+".md", "---\nid: "+id+"\nstatus: completed\n---\n# Task\n")
		r.File("threads/"+id+"-thread-z.md", "---\nid: [unterminated\n---\n")
		r.File("threads/"+id+"-thread-a.md", "---\nid: [unterminated\n---\n")
	}
	out, err := runRootRC(t, "-C", r.Root, "lint", "--json")
	if ExitCode(err) != 11 {
		t.Fatalf("lint exit = %v, want validation failure:\n%s", err, out)
	}
	var envelope wire.LintEnvelope
	if err := json.Unmarshal([]byte(out), &envelope); err != nil {
		t.Fatal(err)
	}
	if len(envelope.Unreadable) != 4 || len(envelope.Issues) != 6 {
		t.Fatalf("duplicate/collision owners = %+v; want one result per occurrence", envelope)
	}
	for i, result := range envelope.Issues {
		if i < 2 {
			wantSlug := fmt.Sprintf("task-6g000000000%d", i+1)
			if result.Slug != wantSlug || len(result.Issues) != 1 {
				t.Fatalf("task result %d = %+v; want %q and one collision", i, result, wantSlug)
			}
			continue
		}
		wantSlug := "thread-a"
		if i%2 != 0 {
			wantSlug = "thread-z"
		}
		if result.Slug != wantSlug || len(result.Issues) != 2 ||
			!strings.Contains(result.Issues[0].Message, "duplicate stable id") ||
			!strings.Contains(result.Issues[1].Message, "also used by a task") {
			t.Fatalf("Thread result %d = %+v; want %q, duplicate then collision", i, result, wantSlug)
		}
	}
	human, err := runRootRC(t, "-C", r.Root, "lint", "--color=never")
	if ExitCode(err) != 11 || strings.Count(human, "also used by a Thread") != 2 ||
		strings.Count(human, "also used by a task") != 4 || strings.Count(human, "duplicate stable id") != 4 {
		t.Fatalf("human duplicate/collision output = %v:\n%s", err, human)
	}
	for range 5 {
		again, err := runRootRC(t, "-C", r.Root, "lint", "--json")
		if ExitCode(err) != 11 || out != again {
			t.Fatalf("JSON duplicate/collision ordering changed: %v\n%s", err, again)
		}
		againHuman, err := runRootRC(t, "-C", r.Root, "lint", "--color=never")
		if ExitCode(err) != 11 || human != againHuman {
			t.Fatalf("human duplicate/collision ordering changed: %v\n%s", err, againHuman)
		}
	}
}
