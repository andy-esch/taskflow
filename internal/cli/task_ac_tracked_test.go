package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/andy-esch/taskflow/internal/testutil"
)

const trackedTestBody = "# Alpha\n\n## Acceptance criteria\n\n- [ ] one\n- [ ] two\n"

func setupRepoForTracked(t *testing.T) string {
	t.Helper()
	root := setupRepo(t)
	runRoot(t, "-C", root, "task", "set", "alpha", "--body", trackedTestBody)
	return root
}

func alphaBody(t *testing.T, root string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(root, "tasks", testutil.TaskID("alpha")+"-alpha.md"))
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// The destination is recorded as the target's STABLE ID, not the reference typed.
// A slug held as prose is a pointer nothing resolves and nothing maintains: `task
// rename` cascades body links, not criterion reasons.
func TestTaskAcTracked_RecordsTheResolvedID(t *testing.T) {
	root := setupRepoForTracked(t)

	runRoot(t, "-C", root, "task", "ac", "alpha", "--tracked", "1",
		"--to", "beta", "--reason", "the split artifact is beta's deliverable")

	body := alphaBody(t, root)
	wantID := testutil.TaskID("beta")
	if !strings.Contains(body, "by "+wantID) {
		t.Errorf("the criterion should record beta's stable id %q:\n%s", wantID, body)
	}
	if !strings.Contains(body, "the split artifact is beta's deliverable") {
		t.Errorf("the prose reason should survive beside the id:\n%s", body)
	}
	if !strings.Contains(body, "**tracked:**") {
		t.Errorf("the state suffix should still be tracked:\n%s", body)
	}
}

// The id, not the typed reference: passing a slug must not store the slug.
func TestTaskAcTracked_StoresTheIDEvenWhenGivenASlug(t *testing.T) {
	root := setupRepoForTracked(t)

	runRoot(t, "-C", root, "task", "ac", "alpha", "--tracked", "1", "--to", "beta", "--reason", "why")

	body := alphaBody(t, root)
	if strings.Contains(body, "by beta") {
		t.Errorf("the slug must not be what is stored — it is not stable:\n%s", body)
	}
}

// `tracked` without a destination is the improvisation the word replaces: it says
// the work left this task while leaving no way to find where it went.
func TestTaskAcTracked_RequiresADestination(t *testing.T) {
	root := setupRepoForTracked(t)
	before := alphaBody(t, root)

	_, err := runRootRC(t, "-C", root, "task", "ac", "alpha", "--tracked", "1", "--reason", "moved elsewhere")
	if err == nil {
		t.Fatal("--tracked without --to should be refused")
	}
	if got := ExitCode(err); got != 11 {
		t.Errorf("exit = %d, want 11", got)
	}
	if !strings.Contains(err.Error(), "--to") {
		t.Errorf("the error should name the flag that fixes it: %v", err)
	}
	if alphaBody(t, root) != before {
		t.Error("a refused write must not touch the file")
	}
}

// Resolution happens BEFORE the write, so an unresolvable destination cannot be
// recorded at all.
func TestTaskAcTracked_UnresolvableDestinationIsRefused(t *testing.T) {
	root := setupRepoForTracked(t)
	before := alphaBody(t, root)

	_, err := runRootRC(t, "-C", root, "task", "ac", "alpha", "--tracked", "1",
		"--to", "no-such-task", "--reason", "why")
	if err == nil {
		t.Fatal("an unresolvable --to should be refused")
	}
	if !strings.Contains(err.Error(), "no-such-task") {
		t.Errorf("the error should name what could not be resolved: %v", err)
	}
	if alphaBody(t, root) != before {
		t.Error("a refused write must not touch the file")
	}
}

// Tracking a criterion to its own task is a no-op pointer that reads as a handoff.
func TestTaskAcTracked_RefusesSelfReference(t *testing.T) {
	root := setupRepoForTracked(t)

	_, err := runRootRC(t, "-C", root, "task", "ac", "alpha", "--tracked", "1", "--to", "alpha", "--reason", "why")
	if err == nil {
		t.Fatal("tracking a criterion to its own task should be refused")
	}
	if !strings.Contains(err.Error(), "its own task") {
		t.Errorf("the error should say why: %v", err)
	}
}

// --to belongs to --tracked alone; accepting it elsewhere would silently drop it.
func TestTaskAcTracked_ToIsRejectedForOtherStates(t *testing.T) {
	root := setupRepoForTracked(t)

	_, err := runRootRC(t, "-C", root, "task", "ac", "alpha", "--defer", "1",
		"--to", "beta", "--reason", "waiting")
	if err == nil {
		t.Fatal("--to with --defer should be refused")
	}
	if !strings.Contains(err.Error(), "--tracked") {
		t.Errorf("the error should point at the flag --to belongs to: %v", err)
	}
}

// The destination is exposed as its own field, which is what makes it queryable
// rather than prose — the whole point of recording it.
func TestTaskAcTracked_DestinationIsAFieldInJSON(t *testing.T) {
	root := setupRepoForTracked(t)
	runRoot(t, "-C", root, "task", "ac", "alpha", "--tracked", "1", "--to", "beta", "--reason", "handed over")

	out := runRoot(t, "-C", root, "--json", "task", "ac", "alpha")

	var env struct {
		Acceptance []struct {
			Index     int    `json:"index"`
			State     string `json:"state"`
			Reason    string `json:"reason"`
			TrackedBy string `json:"tracked_by"`
		} `json:"acceptance"`
	}
	if err := json.Unmarshal([]byte(out), &env); err != nil {
		t.Fatalf("invalid json: %v\n%s", err, out)
	}
	if len(env.Acceptance) == 0 {
		t.Fatalf("no criteria in the envelope:\n%s", out)
	}
	first := env.Acceptance[0]
	if first.TrackedBy != testutil.TaskID("beta") {
		t.Errorf("tracked_by = %q, want beta's id:\n%s", first.TrackedBy, out)
	}
	if first.Reason != "handed over" {
		t.Errorf("reason should be the prose alone, got %q", first.Reason)
	}
}
