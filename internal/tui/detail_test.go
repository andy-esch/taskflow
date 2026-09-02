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

func TestDetailPaneUsesLoadedIdentityInsteadOfDisplayTitle(t *testing.T) {
	d := newDetailPane(&testStyles)
	d.SetSize(40, 5)
	content := taskDetail{
		t:    domain.Task{Slug: "duplicate-display-slug"},
		body: strings.Repeat("line\n", 40),
	}
	d.SetContent("stable-a", content)
	d.vp.GotoBottom()
	offset := d.vp.YOffset()
	if offset == 0 {
		t.Fatal("setup: detail content should scroll")
	}

	d.SetContent("stable-a", content)
	if d.vp.YOffset() != offset {
		t.Errorf("same canonical record lost its scroll offset: %d -> %d", offset, d.vp.YOffset())
	}
	if !d.showing("stable-a") || d.showing("duplicate-display-slug") {
		t.Error("loaded identity was conflated with the display title")
	}

	d.SetContent("stable-b", content)
	if d.vp.YOffset() != 0 {
		t.Errorf("a different canonical record with the same title kept scroll offset %d", d.vp.YOffset())
	}
}
