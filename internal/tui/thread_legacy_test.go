package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/testutil"
	"github.com/charmbracelet/x/ansi"
)

func TestThreadLegacyDiagnosticsReachTUIRegistryAndDetail(t *testing.T) {
	m, root := threadModel(t)
	owner := testutil.TaskID("tui-legacy-outside")
	if err := os.WriteFile(filepath.Join(root, "tasks", owner+"-outside.md"), []byte(fmt.Sprintf("---\nid: %s\nstatus: next-up\ndescription: Outside legacy owner\nblocked_by: [first]\n---\n# Outside\n", owner)), 0o644); err != nil {
		t.Fatal(err)
	}
	m = openThreads(t, m)
	detail := selectedThreadDetail(t, m)
	diagnostics := threadTab(m).threadDiagnostics
	if diagnostics == nil || diagnostics.graphHealth != core.GraphDegraded || len(diagnostics.graphLegacy) != 1 || len(detail.projection.View.GraphLegacy) != 1 || len(detail.projection.View.Frontier) != 0 {
		t.Fatalf("TUI lost shared cause or dispatched: list=%+v detail=%+v", diagnostics, detail.projection.View)
	}
	for name, output := range map[string]string{
		"summary": renderThreadMeta(detail, 180, m.st),
		"ranks":   renderThreadTopology(detail.projection, "", "", 180, m.st),
	} {
		plain := ansi.Strip(output)
		if strings.Count(plain, "legacy blocked_by on "+owner) != 1 || !strings.Contains(plain, "run task depend migrate") {
			t.Fatalf("%s lost unique diagnosis/remedy:\n%s", name, plain)
		}
	}
	m.width = 180
	if footer := ansi.Strip(m.footer()); !strings.Contains(footer, "1 legacy field(s); run task depend migrate") {
		t.Fatalf("active graph view has no legacy/migration cue: %s", footer)
	}
	threadTab(m).list.SetItems(nil)
	if empty := ansi.Strip(m.listPaneContent()); !strings.Contains(empty, "legacy blocked_by on "+owner) || !strings.Contains(empty, "run task depend migrate") || strings.Contains(empty, "(0 diagnostic(s))") {
		t.Fatalf("empty list falsely has no diagnostic: %s", empty)
	}
}

func TestThreadLegacyTUIDeduplicatesGroupedProblemsAndSanitizesContext(t *testing.T) {
	m, _ := threadModel(t)
	legacy := core.LegacyDependencyDiagnostic{TaskID: "owner", Field: "blocked_by", Location: "db://tasks/outside\n\x1b[31m",
		References: []core.LegacyReference{{Value: "gone", Resolution: core.LegacyMissing}}}
	view := core.ThreadView{GraphHealth: core.GraphBroken, GraphLegacy: []core.LegacyDependencyDiagnostic{legacy}, GraphProblems: []core.GraphProblem{
		{Code: core.ProblemLegacyMissing, TaskID: "owner", Field: "blocked_by", Location: legacy.Location, Message: "same missing cause"},
		{Code: core.ProblemInvalidDependencyID, TaskID: "owner", Message: "independent structural defect"},
	}}
	lines := threadDiagnosticLines(view, m.st)
	plain := ansi.Strip(strings.Join(lines, "\n"))
	if len(lines) != 2 || strings.Contains(plain, "same missing cause") || !strings.Contains(plain, `"gone": missing`) || !strings.Contains(plain, "independent structural defect") || !strings.Contains(plain, "run task depend repair, then task depend migrate") || strings.Count(plain, "\n") != 1 {
		t.Fatalf("grouped TUI diagnosis lost severity/context or forged a line: %q", plain)
	}
	if legacy.Location != "db://tasks/outside\n\x1b[31m" {
		t.Fatal("presentation mutated core context")
	}
}
