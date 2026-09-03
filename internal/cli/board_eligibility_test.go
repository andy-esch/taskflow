package cli

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/andy-esch/taskflow/internal/testutil"
)

// blockedBoardRepo gives two next-up tasks where the first alphabetically is
// blocked by the second — the shape that put an unstartable task at the top of the
// board with nothing to distinguish it.
func blockedBoardRepo(t *testing.T) string {
	t.Helper()
	r := testutil.NewRepo(t)
	r.Epic("01-e.md", "---\nstatus: active\npriority: high\ndescription: e\n---\n# E\n")
	// The fixture writer derives each filename's id from its slug, so the frontmatter
	// id must match or the graph is broken rather than merely blocked.
	gateID, workID := testutil.TaskID("zzz-gate"), testutil.TaskID("aaa-blocked")
	r.Task("next-up", "aaa-blocked.md", "---\nid: "+workID+
		"\nstatus: next-up\nepic: 01-e\ndescription: cannot start yet\ntags: [t]\ndepends_on: ["+gateID+"]\n---\n# blocked\n")
	r.Task("next-up", "zzz-gate.md", "---\nid: "+gateID+
		"\nstatus: next-up\nepic: 01-e\ndescription: the prerequisite\ntags: [t]\n---\n# gate\n")
	return r.Root
}

// The board is the tool's answer to "what should I do next". Answering with work
// `task start` will refuse is the one thing it must not do — and an agent re-derives
// the board every session, so it makes the same wrong choice every time.
func TestBoard_MarksBlockedTasks(t *testing.T) {
	root := blockedBoardRepo(t)

	out := runRoot(t, "-C", root, "board")

	if !strings.Contains(out, "⛔") {
		t.Errorf("a blocked task should be marked:\n%s", out)
	}
	blockedLine, gateLine := -1, -1
	for i, line := range strings.Split(out, "\n") {
		switch {
		case strings.Contains(line, "aaa-blocked"):
			blockedLine = i
			if !strings.Contains(line, "⛔") {
				t.Errorf("the blocked task's own row should carry the marker: %q", line)
			}
		case strings.Contains(line, "zzz-gate"):
			gateLine = i
			if strings.Contains(line, "⛔") {
				t.Errorf("the startable task must not be marked: %q", line)
			}
		}
	}
	if blockedLine < 0 || gateLine < 0 {
		t.Fatalf("both tasks should be listed:\n%s", out)
	}
	// Eligible work first: marking alone still leaves the top row unstartable.
	if blockedLine < gateLine {
		t.Errorf("blocked work should be parked below startable work:\n%s", out)
	}
}

func TestBoard_JSONCarriesTheBlockedFlag(t *testing.T) {
	root := blockedBoardRepo(t)

	out := runRoot(t, "-C", root, "--json", "board")

	var env struct {
		Columns []struct {
			Status string `json:"status"`
			Tasks  []struct {
				Slug    string `json:"slug"`
				Blocked bool   `json:"blocked"`
			} `json:"tasks"`
		} `json:"columns"`
	}
	if err := json.Unmarshal([]byte(out), &env); err != nil {
		t.Fatalf("invalid json: %v\n%s", err, out)
	}
	seen := map[string]bool{}
	for _, col := range env.Columns {
		for _, task := range col.Tasks {
			seen[task.Slug] = task.Blocked
		}
	}
	if !seen["aaa-blocked"] {
		t.Errorf("the blocked task should be flagged in JSON:\n%s", out)
	}
	if seen["zzz-gate"] {
		t.Errorf("the startable task must not be flagged:\n%s", out)
	}
}

// An in-progress task has already started, so reporting its gate would be advice
// about a decision already taken.
func TestBoard_InProgressIsNotMarkedBlocked(t *testing.T) {
	r := testutil.NewRepo(t)
	r.Epic("01-e.md", "---\nstatus: active\npriority: high\ndescription: e\n---\n# E\n")
	gateID := testutil.TaskID("gate")
	r.Task("in-progress", "running.md", "---\nid: "+testutil.TaskID("running")+
		"\nstatus: in-progress\nepic: 01-e\ndescription: already going\ntags: [t]\ndepends_on: ["+gateID+"]\n---\n# running\n")
	r.Task("next-up", "gate.md", "---\nid: "+gateID+
		"\nstatus: next-up\nepic: 01-e\ndescription: prereq\ntags: [t]\n---\n# gate\n")

	out := runRoot(t, "-C", r.Root, "board")

	for _, line := range strings.Split(out, "\n") {
		if strings.Contains(line, "running") && strings.Contains(line, "⛔") {
			t.Errorf("an in-progress task should not be marked blocked: %q", line)
		}
	}
}

// A repository with no dependencies at all must render exactly as before — the
// marking is additive, not a new column or a reordering.
func TestBoard_UnblockedRepoIsUnchanged(t *testing.T) {
	root := setupRepo(t)

	out := runRoot(t, "-C", root, "board")

	if strings.Contains(out, "⛔") {
		t.Errorf("nothing is blocked here:\n%s", out)
	}
	if !strings.Contains(out, "alpha") || !strings.Contains(out, "beta") {
		t.Errorf("both tasks should still be listed:\n%s", out)
	}
}
