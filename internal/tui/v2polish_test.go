package tui

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/andy-esch/taskflow/internal/domain"
)

// TestView_WindowTitle: the terminal window title reflects the current selection.
func TestView_WindowTitle(t *testing.T) {
	m := loaded(t, 80, 24)
	id := m.selectedID()
	if id == "" {
		t.Fatal("setup: expected a selected task")
	}
	if got := m.View().WindowTitle; got != "tskflwctl · "+id {
		t.Errorf("WindowTitle = %q, want %q", got, "tskflwctl · "+id)
	}
}

// TestView_WindowTitleFallsBackToTab: with no selection, the title shows the tab.
func TestView_WindowTitleFallsBackToTab(t *testing.T) {
	m := newModel(t) // not loaded → nothing selected
	tm, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m = tm.(Model)
	m.onDash = false // off the landing dashboard, onto a tab with no selection
	if got := m.View().WindowTitle; got != "tskflwctl · "+m.cur().name {
		t.Errorf("WindowTitle = %q, want %q", got, "tskflwctl · "+m.cur().name)
	}
	m.onDash = true
	if got := m.View().WindowTitle; got != "tskflwctl · overview" {
		t.Errorf("on the dashboard the title should be the dashboard, got %q", got)
	}
}

// TestDetailTitle_ClickableLink: the detail title is an OSC 8 hyperlink to the
// loaded entity's file path (click-to-open).
func TestDetailTitle_ClickableLink(t *testing.T) {
	m := loaded(t, 120, 40) // two-pane, so the detail pane renders
	m.detail.SetContent(taskDetail{t: domain.Task{Slug: "x", Path: "/abs/x.md"}, body: "hi"})
	out := m.View().Content
	if !strings.Contains(out, "\x1b]8;;file:///abs/x.md") {
		t.Errorf("detail title should be an OSC 8 link to the file path; got:\n%q", out)
	}
}

// TestDetailTitle_NoLinkWithoutContent: with no loaded content (the "Detail"
// placeholder), the title is not linkified — no dangling file:// link.
func TestDetailTitle_NoLinkWithoutContent(t *testing.T) {
	m := loaded(t, 120, 40)
	m.detail.clear() // no content
	if out := m.View().Content; strings.Contains(out, "]8;;file://") {
		t.Errorf("empty detail must not emit a file link:\n%q", out)
	}
}
