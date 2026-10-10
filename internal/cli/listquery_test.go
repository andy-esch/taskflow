package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/testutil"
	"github.com/andy-esch/taskflow/internal/wire"
)

func listQueryFixture(t *testing.T) string {
	t.Helper()
	root := freshRepo(t)
	for i, name := range []string{"alpha", "beta", "gamma"} {
		taskID := writeThreadApplyCLITask(t, root, name, domain.StatusNextUp)
		mustWrite(t, filepath.Join(root, "epics", fmt.Sprintf("%02d-%s.md", i+1, name)),
			"---\nstatus: active\ndescription: bike "+name+"\ncreated: \"2026-10-08\"\nupdated_at: \"2026-10-10\"\n---\n# "+name+"\n")
		auditID := testutil.TaskID("audit-" + name)
		mustWrite(t, filepath.Join(root, "audits", auditID+"-"+name+".md"),
			"---\nschema: 1\nid: "+auditID+"\nbucket: open\narea: bike "+name+"\ndate: \"2026-10-08\"\n---\n# "+name+"\n\n#### M1. bike "+name+"\n\n**Status:** open\n\nEvidence.\n")
		researchID := testutil.TaskID("research-" + name)
		mustWrite(t, filepath.Join(root, "research", researchID+"-"+name+".md"),
			"---\nschema: 1\nid: "+researchID+"\ndescription: bike "+name+"\ncreated: \"2026-10-08\"\n---\n# "+name+"\n")
		threadID := testutil.TaskID("thread-" + name)
		mustWrite(t, filepath.Join(root, "threads", threadID+"-"+name+".md"),
			"---\nschema: 1\nid: "+threadID+"\nstatus: unstarted\ndescription: bike "+name+"\ngoal: ship the bike\ncreated: \"2026-10-08\"\ntasks: ["+taskID+"]\n---\n# "+name+"\n")
	}
	return root
}

func decodeListPage(t *testing.T, out, key string) ([]map[string]any, *wire.ListPageJSON) {
	t.Helper()
	var envelope map[string]json.RawMessage
	if err := json.Unmarshal([]byte(out), &envelope); err != nil {
		t.Fatalf("envelope: %v\n%s", err, out)
	}
	var items []map[string]any
	if err := json.Unmarshal(envelope[key], &items); err != nil {
		t.Fatalf("items: %v\n%s", err, out)
	}
	var page *wire.ListPageJSON
	if value, ok := envelope["page"]; ok {
		if err := json.Unmarshal(value, &page); err != nil {
			t.Fatal(err)
		}
	}
	return items, page
}

func TestPlanningListsPageAcrossRealAdapterAndOutputModes(t *testing.T) {
	root := listQueryFixture(t)
	before := testutil.SnapshotTree(t, root)
	for _, surface := range []struct {
		args []string
		key  string
	}{
		{[]string{"task", "list"}, "tasks"}, {[]string{"epic", "list"}, "epics"},
		{[]string{"audit", "list"}, "audits"}, {[]string{"research", "list"}, "research"},
		{[]string{"thread", "list"}, "threads"}, {[]string{"audit", "findings"}, "findings"},
	} {
		t.Run(surface.key, func(t *testing.T) {
			base := append([]string{"-C", root}, surface.args...)
			legacyOut := runRoot(t, append(slices.Clone(base), "--json")...)
			legacy, page := decodeListPage(t, legacyOut, surface.key)
			if len(legacy) != 3 || page != nil {
				t.Fatalf("legacy=%v page=%+v", legacy, page)
			}
			for _, projection := range [][]string{nil, {"-c", "id"}} {
				if surface.key == "findings" && len(projection) > 0 {
					projection = []string{"-c", "audit,code"}
				}
				args := append(slices.Clone(base), "--json", "--limit", "2", "--sort", "name")
				args = append(args, projection...)
				res, err := runRootStreams(t, args...)
				if err != nil || res.Err != "" {
					t.Fatalf("page stdout/stderr: %+v, %v", res, err)
				}
				first, page := decodeListPage(t, res.Out, surface.key)
				if len(first) != 2 || page == nil || page.Returned != 2 || page.Matched != 3 || page.NextCursor == "" || page.Order[0] != "name" {
					t.Fatalf("first=%v page=%+v", first, page)
				}
				second, final := decodeListPage(t, runRoot(t, append(args, "--cursor", page.NextCursor)...), surface.key)
				if len(second) != 1 || final.NextCursor != "" || reflect.DeepEqual(second[0], first[0]) || reflect.DeepEqual(second[0], first[1]) {
					t.Fatalf("first=%v second=%v final=%+v", first, second, final)
				}
			}
			for _, mode := range []string{"name", "table", "csv", "human"} {
				res, err := runRootStreams(t, append(slices.Clone(base), "-o", mode, "--sort", "name", "--limit", "1", "--color=never")...)
				if err != nil || strings.Contains(res.Out, "cursor") || !strings.Contains(res.Err, "returned 1 of 3") {
					t.Fatalf("mode=%s result=%+v err=%v", mode, res, err)
				}
			}
			items, empty := decodeListPage(t, runRoot(t, append(slices.Clone(base), "--json", "--limit", "1", "--search", "not present")...), surface.key)
			if items == nil || len(items) != 0 || empty.Matched != 0 || empty.NextCursor != "" {
				t.Fatalf("empty=%v page=%+v", items, empty)
			}
		})
	}
	if !reflect.DeepEqual(before, testutil.SnapshotTree(t, root)) {
		t.Fatal("list queries wrote to the repository")
	}
}

func TestListQueryChangedFilesAndDiagnosticsRefuseContinuation(t *testing.T) {
	root := listQueryFixture(t)
	base := []string{"-C", root, "task", "list", "--json", "--limit", "1", "--sort", "name"}
	_, page := decodeListPage(t, runRoot(t, base...), "tasks")
	path := filepath.Join(root, "tasks", testutil.TaskID("alpha")+"-alpha.md")
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	mustWrite(t, path, strings.Replace(string(content), "description: alpha", "description: changed", 1))
	if _, err := runRootRC(t, append(slices.Clone(base), "--cursor", page.NextCursor)...); !errors.Is(err, domain.ErrConflict) || !strings.Contains(err.Error(), "restart") {
		t.Fatalf("changed record: %v", err)
	}
	mustWrite(t, path, string(content))
	mustWrite(t, filepath.Join(root, "tasks", "malformed.md"), "---\nstatus: [\n---\n")
	if _, err := runRootRC(t, append(slices.Clone(base), "--cursor", page.NextCursor)...); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("new load diagnostic: %v", err)
	}
	res, err := runRootStreams(t, base...)
	if err == nil || !strings.Contains(res.Out, "unreadable") || !strings.Contains(res.Out, "next_cursor") {
		t.Fatalf("partial page suppressed diagnostics: %+v, %v", res, err)
	}
	_, partial := decodeListPage(t, res.Out, "tasks")
	res, err = runRootStreams(t, append(slices.Clone(base), "--cursor", partial.NextCursor)...)
	if err == nil || errors.Is(err, domain.ErrConflict) || !strings.Contains(res.Out, "unreadable") {
		t.Fatalf("continuation lost existing read problems: %+v, %v", res, err)
	}
}

func TestListQueryValidationAndUpdatedSinceSemantics(t *testing.T) {
	root := listQueryFixture(t)
	for _, flags := range [][]string{
		{"--limit", "nope"}, {"--limit", "1.5"}, {"--limit", "99999999999999999999999999"},
		{"--limit", "0"}, {"--limit", "-1"}, {"--limit", "1001"}, {"--sort", "priority"}, {"--sort", ""},
		{"--cursor", "invalid"}, {"--limit", "1", "--cursor", "invalid"},
		{"--updated-since", "2026-02-30"}, {"--search", strings.Repeat("x", 257)},
	} {
		if _, err := runRootRC(t, append([]string{"-C", root, "task", "list", "--json"}, flags...)...); !errors.Is(err, domain.ErrValidation) {
			t.Fatalf("flags=%q err=%v", flags, err)
		}
	}
	for _, flags := range [][]string{{"--updated-since", "2026-10-10"}, {"--sort", "created"}} {
		if _, err := runRootRC(t, append([]string{"-C", root, "audit", "findings"}, flags...)...); !errors.Is(err, domain.ErrValidation) {
			t.Fatalf("finding flags=%q err=%v", flags, err)
		}
	}
	items, page := decodeListPage(t, runRoot(t, "-C", root, "epic", "list", "--json", "--updated-since", "2026-10-10"), "epics")
	if len(items) != 3 || page == nil {
		t.Fatalf("own updated date: %v, %+v", items, page)
	}
	items, _ = decodeListPage(t, runRoot(t, "-C", root, "research", "list", "--json", "--updated-since", "2026-10-09"), "research")
	if len(items) != 0 {
		t.Fatalf("creation fallback included older research: %v", items)
	}
}

func TestListQueryBoundedProjectionVolume(t *testing.T) {
	root := freshRepo(t)
	for i := range 100 {
		writeThreadApplyCLITask(t, root, fmt.Sprintf("bike-%03d", i), domain.StatusNextUp)
	}
	full := runRoot(t, "-C", root, "task", "list", "--json")
	bounded := runRoot(t, "-C", root, "task", "list", "--json", "--limit", "5", "-c", "id,slug")
	items, page := decodeListPage(t, bounded, "tasks")
	if len(items) != 5 || page.Matched != 100 || len(bounded)*10 >= len(full) {
		t.Fatalf("bound/projection failed: %d vs %d bytes; page=%+v", len(bounded), len(full), page)
	}
	t.Logf("100-task fixture: full JSON %d bytes; five projected rows with continuation %d bytes (%.1f%% smaller)", len(full), len(bounded), 100*(1-float64(len(bounded))/float64(len(full))))
}

func TestThreadQueryKeepsUnhealthyGraphVisibleOnEmptyPage(t *testing.T) {
	root := listQueryFixture(t)
	path := filepath.Join(root, "tasks", testutil.TaskID("alpha")+"-alpha.md")
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	mustWrite(t, path, strings.Replace(string(b), "schema: 1", "schema: 1\ndepends_on: ["+testutil.TaskID("absent")+"]", 1))
	before := testutil.SnapshotTree(t, root)
	for _, search := range []string{"no matches", "beta"} {
		for _, columns := range [][]string{nil, {"-c", "id"}, {"-c", "id,graph_health,projection_health"}} {
			args := []string{"-C", root, "thread", "list", "--json", "--limit", "1", "--search", search}
			res, err := runRootStreams(t, append(args, columns...)...)
			var envelope struct {
				GraphHealth   string                  `json:"graph_health"`
				GraphProblems []wire.GraphProblemJSON `json:"graph_problems"`
				Threads       []map[string]any        `json:"threads"`
				Page          *wire.ListPageJSON      `json:"page"`
			}
			if err != nil || res.Err != "" {
				t.Fatalf("search=%q columns=%v streams=%+v err=%v", search, columns, res, err)
			}
			if err := json.Unmarshal([]byte(res.Out), &envelope); err != nil {
				t.Fatal(err)
			}
			wantRows := 0
			if search == "beta" {
				wantRows = 1
			}
			if envelope.GraphHealth != "broken" || len(envelope.GraphProblems) == 0 ||
				envelope.GraphProblems[0].Code != "missing-dependency" || envelope.Threads == nil ||
				len(envelope.Threads) != wantRows || envelope.Page == nil || envelope.Page.Matched != wantRows {
				t.Fatalf("search=%q columns=%v masked global graph problems: %+v", search, columns, envelope)
			}
		}
		for _, mode := range []string{"name", "table", "csv", "human"} {
			res, err := runRootStreams(t, "-C", root, "thread", "list", "--limit", "1", "--search", search, "-o", mode, "--color=never")
			diagnostics := res.Err
			if mode == "human" {
				diagnostics = res.Out // Human output already contains the graph context.
			} else if strings.Contains(res.Out, "missing-dependency") || strings.Contains(res.Out, "Diagnostics") {
				t.Fatalf("mode=%s polluted row stream: %+v", mode, res)
			}
			if err != nil || !strings.Contains(diagnostics, "repository graph") || !strings.Contains(diagnostics, "missing-dependency") {
				t.Fatalf("search=%q mode=%s masked global graph problems: %+v, %v", search, mode, res, err)
			}
		}
	}
	if !reflect.DeepEqual(before, testutil.SnapshotTree(t, root)) {
		t.Fatal("graph diagnostics changed the corpus")
	}
	// The opt-in correction must not change legacy projected envelopes/streams.
	legacy, err := runRootStreams(t, "-C", root, "thread", "list", "--json", "-c", "id")
	if err != nil || legacy.Err != "" || strings.Contains(legacy.Out, "graph_health") || strings.Contains(legacy.Out, "page") {
		t.Fatalf("legacy projection changed: %+v, %v", legacy, err)
	}
	for _, mode := range []string{"name", "table", "csv"} {
		legacy, err := runRootStreams(t, "-C", root, "thread", "list", "-o", mode)
		if err != nil || legacy.Err != "" {
			t.Fatalf("legacy %s stream changed: %+v, %v", mode, legacy, err)
		}
	}
	// Healthy bounded projections still publish an explicit, empty problem list.
	mustWrite(t, path, string(b))
	healthy, err := runRootStreams(t, "-C", root, "thread", "list", "--json", "--search", "no matches", "-c", "id")
	var context struct {
		GraphHealth   string                  `json:"graph_health"`
		GraphProblems []wire.GraphProblemJSON `json:"graph_problems"`
	}
	if err != nil || json.Unmarshal([]byte(healthy.Out), &context) != nil || context.GraphHealth != "healthy" || context.GraphProblems == nil || len(context.GraphProblems) != 0 {
		t.Fatalf("healthy query lost graph verdict: %+v, %v", healthy, err)
	}
	for _, mode := range []string{"name", "table", "csv"} {
		healthy, err := runRootStreams(t, "-C", root, "thread", "list", "--search", "no matches", "-o", mode)
		if err != nil || healthy.Err != "" {
			t.Fatalf("healthy %s query emitted diagnostics: %+v, %v", mode, healthy, err)
		}
	}
}

func TestListQueryCursorDoesNotBindOutputFormatButDoesBindEntityFilters(t *testing.T) {
	root := listQueryFixture(t)
	base := []string{"-C", root, "task", "list", "--json", "--limit", "1", "--sort", "name", "--tag", "threads"}
	_, page := decodeListPage(t, runRoot(t, base...), "tasks")
	items, _ := decodeListPage(t, runRoot(t, append(slices.Clone(base), "--cursor", page.NextCursor, "-c", "slug")...), "tasks")
	if len(items) != 1 || items[0]["slug"] != "beta" {
		t.Fatalf("projected continuation=%+v", items)
	}
	for _, args := range [][]string{
		{"-C", root, "task", "list", "--limit", "1", "--sort", "name", "--tag", "other", "--cursor", page.NextCursor},
		{"-C", root, "research", "list", "--limit", "1", "--sort", "name", "--cursor", page.NextCursor},
	} {
		if _, err := runRootRC(t, args...); !errors.Is(err, domain.ErrValidation) {
			t.Fatalf("foreign cursor accepted: %v", err)
		}
	}
}
