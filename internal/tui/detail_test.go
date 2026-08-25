package tui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"

	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/theme"
)

// Criterion 9: the roll-up is in the detail HEADER, so a criterion carrying a decision is
// visible without scrolling into the body. Before this, a task could say "deferred: waiting
// on the schema ADR" and the reader would only find it by scrolling.
func TestRenderTaskMeta_CriterionRollup(t *testing.T) {
	body := "## Acceptance criteria\n\n" +
		"- [x] done\n- [ ] open\n- [ ] parked · **deferred:** waiting on the ADR\n"
	st := testStyles
	got := renderTaskMeta(domain.Task{Slug: "t", Status: domain.StatusInProgress}, body, 120, &st)
	if !strings.Contains(got, "acceptance") {
		t.Fatalf("the header must carry an acceptance roll-up:\n%s", got)
	}
	plain := ansi.Strip(got)
	for _, want := range []string{"1/3", theme.CriterionState("met").Glyph, theme.CriterionState("deferred").Glyph} {
		if !strings.Contains(plain, want) {
			t.Errorf("roll-up missing %q:\n%s", want, plain)
		}
	}
	// A task with no criteria — most of them — gets no row at all rather than an empty one.
	bare := renderTaskMeta(domain.Task{Slug: "t", Status: domain.StatusInProgress}, "# Title\n\nprose\n", 120, &st)
	if strings.Contains(ansi.Strip(bare), "acceptance") {
		t.Errorf("a task with no criteria should have no acceptance row:\n%s", bare)
	}
}
