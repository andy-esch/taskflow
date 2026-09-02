package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/andy-esch/taskflow/internal/testutil"
	"github.com/andy-esch/taskflow/internal/wire"
)

// acMultiBody gives alpha four criteria, all unchecked, so a multi-index flip has
// something to leave alone as well as something to change.
const acMultiBody = "# Alpha\n\n## Acceptance criteria\n\n- [ ] one\n- [ ] two\n- [ ] three\n- [ ] four\n"

func setupRepoWithFourAC(t *testing.T) string {
	t.Helper()
	root := setupRepo(t)
	runRoot(t, "-C", root, "task", "set", "alpha", "--body", acMultiBody)
	return root
}

// acStates reads back the checkbox states in order.
func acStates(t *testing.T, root string) []bool {
	t.Helper()
	path := filepath.Join(root, "tasks", testutil.TaskID("alpha")+"-alpha.md")
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var got []bool
	for _, line := range strings.Split(string(b), "\n") {
		switch {
		case strings.HasPrefix(line, "- [x]"), strings.HasPrefix(line, "- [X]"):
			got = append(got, true)
		case strings.HasPrefix(line, "- [ ]"):
			got = append(got, false)
		}
	}
	return got
}

func updatedStamp(t *testing.T, root string) string {
	t.Helper()
	path := filepath.Join(root, "tasks", testutil.TaskID("alpha")+"-alpha.md")
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	for _, line := range strings.Split(string(b), "\n") {
		if strings.HasPrefix(line, "updated_at:") {
			return line
		}
	}
	return ""
}

func TestTaskAc_CheckCommaSeparatedList(t *testing.T) {
	root := setupRepoWithFourAC(t)

	runRoot(t, "-C", root, "task", "ac", "alpha", "--check", "1,2,4")

	want := []bool{true, true, false, true}
	got := acStates(t, root)
	if len(got) != 4 {
		t.Fatalf("expected 4 criteria, got %d (%v)", len(got), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("criterion %d = %v, want %v (all: %v)", i+1, got[i], want[i], got)
		}
	}
}

func TestTaskAc_CheckRepeatedFlagMatchesCommaList(t *testing.T) {
	comma := setupRepoWithFourAC(t)
	repeated := setupRepoWithFourAC(t)

	runRoot(t, "-C", comma, "task", "ac", "alpha", "--check", "1,2")
	runRoot(t, "-C", repeated, "task", "ac", "alpha", "--check", "1", "--check", "2")

	if a, b := acStates(t, comma), acStates(t, repeated); !equalBools(a, b) {
		t.Errorf("repeated flags = %v, comma-separated = %v; the two spellings must agree", b, a)
	}
}

func TestTaskAc_UncheckList(t *testing.T) {
	root := setupRepoWithFourAC(t)
	runRoot(t, "-C", root, "task", "ac", "alpha", "--check", "1,2,3")

	runRoot(t, "-C", root, "task", "ac", "alpha", "--uncheck", "1,3")

	want := []bool{false, true, false, false}
	if got := acStates(t, root); !equalBools(got, want) {
		t.Errorf("after --uncheck 1,3 = %v, want %v", got, want)
	}
}

// A single index is the old spelling and must keep working unchanged — the flag
// changed type, which is exactly the kind of change that silently breaks callers.
func TestTaskAc_SingleIndexStillWorks(t *testing.T) {
	root := setupRepoWithFourAC(t)

	runRoot(t, "-C", root, "task", "ac", "alpha", "--check", "3")

	want := []bool{false, false, true, false}
	if got := acStates(t, root); !equalBools(got, want) {
		t.Errorf("after --check 3 = %v, want %v", got, want)
	}
}

func TestTaskAc_DuplicateIndicesFlipOnce(t *testing.T) {
	root := setupRepoWithFourAC(t)

	runRoot(t, "-C", root, "task", "ac", "alpha", "--check", "1,1,2")

	want := []bool{true, true, false, false}
	if got := acStates(t, root); !equalBools(got, want) {
		t.Errorf("after --check 1,1,2 = %v, want %v", got, want)
	}
}

// Order must not matter: a criterion is named by its index, not by its position in
// the argument list.
func TestTaskAc_UnsortedIndices(t *testing.T) {
	root := setupRepoWithFourAC(t)

	runRoot(t, "-C", root, "task", "ac", "alpha", "--check", "4,1")

	want := []bool{true, false, false, true}
	if got := acStates(t, root); !equalBools(got, want) {
		t.Errorf("after --check 4,1 = %v, want %v", got, want)
	}
}

// An out-of-range index anywhere in the list rejects the WHOLE request: a partially
// applied multi-flip would leave a state the caller never asked for and cannot
// distinguish from success.
func TestTaskAc_OutOfBoundsRejectsBeforeWriting(t *testing.T) {
	for _, bad := range []string{"0", "99", "1,99", "99,1"} {
		root := setupRepoWithFourAC(t)
		before := acStates(t, root)

		_, err := runRootRC(t, "-C", root, "task", "ac", "alpha", "--check", bad)
		if err == nil {
			t.Fatalf("--check %s should fail", bad)
		}
		if got := ExitCode(err); got != 11 {
			t.Errorf("--check %s exit = %d, want 11", bad, got)
		}
		if after := acStates(t, root); !equalBools(before, after) {
			t.Errorf("--check %s wrote before validating: %v -> %v", bad, before, after)
		}
	}
}

// The three exclusive combinations are rejected by cobra's flag groups before the
// command body runs. Pinned as-is (exit 1, cobra's usage error) because that is
// pre-existing behaviour the list-valued flags must not change.
func TestTaskAc_MutuallyExclusiveCombinations(t *testing.T) {
	cases := [][]string{
		{"--check", "1", "--uncheck", "2"},
		{"--list", "--check", "1"},
		{"--list", "--uncheck", "1"},
	}
	for _, flags := range cases {
		root := setupRepoWithFourAC(t)
		before := acStates(t, root)

		args := append([]string{"-C", root, "task", "ac", "alpha"}, flags...)
		_, err := runRootRC(t, args...)
		if err == nil {
			t.Fatalf("%v should be rejected", flags)
		}
		if after := acStates(t, root); !equalBools(before, after) {
			t.Errorf("%v wrote despite being rejected: %v -> %v", flags, before, after)
		}
	}
}

// A partial flip changes only what needs changing, and says how many.
func TestTaskAc_PartialFlipReportsAccurately(t *testing.T) {
	root := setupRepoWithFourAC(t)
	runRoot(t, "-C", root, "task", "ac", "alpha", "--check", "1")

	out := runRoot(t, "-C", root, "task", "ac", "alpha", "--check", "1,2,3")

	want := []bool{true, true, true, false}
	if got := acStates(t, root); !equalBools(got, want) {
		t.Errorf("partial flip = %v, want %v", got, want)
	}
	if !strings.Contains(out, "1, 2 and 3") {
		t.Errorf("the status line should name the criteria it set:\n%s", out)
	}
}

// Every requested criterion already in the target state is a no-op: no write, and
// crucially no updated_at bump, so an idempotent re-run doesn't churn the file.
func TestTaskAc_FullNoOpDoesNotWrite(t *testing.T) {
	root := setupRepoWithFourAC(t)
	runRoot(t, "-C", root, "task", "ac", "alpha", "--check", "1,2")
	stampBefore := updatedStamp(t, root)

	out := runRoot(t, "-C", root, "task", "ac", "alpha", "--check", "1,2")

	if !strings.Contains(out, "already") {
		t.Errorf("a no-op should say so:\n%s", out)
	}
	if stampAfter := updatedStamp(t, root); stampAfter != stampBefore {
		t.Errorf("a no-op must not bump updated_at: %q -> %q", stampBefore, stampAfter)
	}
}

func TestTaskAc_DryRunPreviewsWithoutWriting(t *testing.T) {
	root := setupRepoWithFourAC(t)
	before := acStates(t, root)

	out := runRoot(t, "-C", root, "--dry-run", "task", "ac", "alpha", "--check", "1,2,4")

	if !strings.Contains(out, "would") {
		t.Errorf("a dry run should say what it would do:\n%s", out)
	}
	if after := acStates(t, root); !equalBools(before, after) {
		t.Errorf("a dry run wrote: %v -> %v", before, after)
	}
}

func TestTaskAc_MultiCheckJSONEnvelope(t *testing.T) {
	root := setupRepoWithFourAC(t)

	out := runRoot(t, "-C", root, "--json", "task", "ac", "alpha", "--check", "1,2")

	var env struct {
		SchemaVersion string `json:"schema_version"`
		Task          struct {
			Slug string `json:"slug"`
		} `json:"task"`
		Body string `json:"body"`
	}
	if err := json.Unmarshal([]byte(out), &env); err != nil {
		t.Fatalf("invalid json: %v\n%s", err, out)
	}
	// Compared against the constant, not a literal: the acceptance criterion names
	// 1.40, which was current when the task was written and would now be a
	// regression to assert. The invariant it meant is "the standard envelope at the
	// current schema version", and that is what cannot go stale.
	if env.SchemaVersion != wire.SchemaVersion {
		t.Errorf("schema_version = %q, want %q:\n%s", env.SchemaVersion, wire.SchemaVersion, out)
	}
	if env.Task.Slug != "alpha" {
		t.Errorf("envelope task = %q, want alpha", env.Task.Slug)
	}
	if strings.Count(env.Body, "- [x]") != 2 {
		t.Errorf("the returned body should show both flips:\n%s", env.Body)
	}
}

// A criterion carrying a state suffix is met by --check, which drops the suffix —
// the behaviour the single-index path has, now across a list.
func TestTaskAc_MultiCheckClearsStateSuffixes(t *testing.T) {
	root := setupRepoWithFourAC(t)
	runRoot(t, "-C", root, "task", "ac", "alpha", "--defer", "2", "--reason", "waiting on the ADR")

	runRoot(t, "-C", root, "task", "ac", "alpha", "--check", "1,2")

	path := filepath.Join(root, "tasks", testutil.TaskID("alpha")+"-alpha.md")
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(b), "deferred:") {
		t.Errorf("checking a deferred criterion must clear its suffix:\n%s", b)
	}
	want := []bool{true, true, false, false}
	if got := acStates(t, root); !equalBools(got, want) {
		t.Errorf("states = %v, want %v", got, want)
	}
}

// The four state flags each need their own reason, so they stay single-index. A
// list there must be rejected rather than silently applying one reason to several.
func TestTaskAc_StateFlagsRemainSingleIndex(t *testing.T) {
	root := setupRepoWithFourAC(t)
	before := acStates(t, root)

	_, err := runRootRC(t, "-C", root, "task", "ac", "alpha", "--defer", "1,2", "--reason", "x")
	if err == nil {
		t.Fatal("--defer should not accept a list")
	}
	if after := acStates(t, root); !equalBools(before, after) {
		t.Errorf("a rejected --defer wrote: %v -> %v", before, after)
	}
}

func TestTaskAc_NoAcceptanceSectionIsValidationError(t *testing.T) {
	root := setupRepo(t) // beta has no acceptance section

	_, err := runRootRC(t, "-C", root, "task", "ac", "beta", "--check", "1,2")
	if err == nil {
		t.Fatal("a task with no acceptance section must fail")
	}
	if got := ExitCode(err); got != 11 {
		t.Errorf("exit = %d, want 11", got)
	}
	if !strings.Contains(err.Error(), "acceptance criteria") {
		t.Errorf("the error should name the missing section: %v", err)
	}
}

func equalBools(a, b []bool) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
