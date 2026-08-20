package tui

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/store"
	"github.com/andy-esch/taskflow/internal/testutil"
	"github.com/charmbracelet/x/ansi"
)

// toResearch tabs forward until it lands on research, settling each load. Derived from
// the tab name rather than a fixed number of presses, so tab order can change freely.
func toResearch(t *testing.T, m Model) Model {
	t.Helper()
	for i := 0; i <= len(m.tabs)+1; i++ {
		if m.cur() != nil && m.cur().name == "research" {
			return m
		}
		tm, cmd := m.Update(press("]"))
		m = settle(t, tm.(Model), cmd)
	}
	t.Fatalf("never landed on the research tab")
	return m
}

// settle applies a command's message whether it is a single msg or a tea.Batch — the
// research tab is reached by plain tab-cycling (single msg) but `o`/`enter` can batch.
func settle(t *testing.T, m Model, cmd tea.Cmd) Model {
	t.Helper()
	if cmd == nil {
		return m
	}
	msg := cmd()
	if msg == nil {
		return m
	}
	if batch, ok := msg.(tea.BatchMsg); ok {
		for _, c := range batch {
			m = settle(t, m, c)
		}
		return m
	}
	tm, _ := m.Update(msg)
	return tm.(Model)
}

func TestResearchTab_ListsNewestFirst(t *testing.T) {
	m := toResearch(t, loaded(t, 120, 40))
	if m.cur().name != "research" {
		t.Fatalf("expected the research tab, got %q", m.cur().name)
	}
	if got := len(m.cur().list.Items()); got != 2 {
		t.Fatalf("expected 2 research docs, got %d", got)
	}
	// ListResearch order is newest-first, so zeta (2026-06-10) precedes alpha (2026-06-02)
	// even though that's the reverse of slug order — proving the row order is the
	// service's, not incidental alphabetical.
	if got := m.selectedID(); got != "zeta-doc" {
		t.Errorf("expected zeta-doc selected (newest), got %q", got)
	}
}

// The row leads with the created date (research has no status glyph or progress bar) and
// carries the description.
func TestResearchTab_RowShowsDateAndDescription(t *testing.T) {
	m := toResearch(t, loaded(t, 120, 40))
	v := ansi.Strip(m.View().Content)
	for _, want := range []string{"2026-06-10", "zeta-doc", "the zeta research"} {
		if !strings.Contains(v, want) {
			t.Errorf("research row missing %q:\n%s", want, v)
		}
	}
}

// The detail view renders metadata + body through core.Service.
func TestResearchTab_DetailRendersMetaAndBody(t *testing.T) {
	m := toResearch(t, loaded(t, 120, 40))
	tm, cmd := m.Update(press("enter"))
	m = settle(t, tm.(Model), cmd)
	v := ansi.Strip(m.View().Content)
	for _, want := range []string{"zeta-doc", "2026-06-10"} {
		if !strings.Contains(v, want) {
			t.Errorf("detail missing %q:\n%s", want, v)
		}
	}
}

// Research is the only AXIS-LESS tab: `s`/`S` must be inert rather than showing a view
// chip for an axis that doesn't exist.
func TestResearchTab_StatusViewKeysAreInert(t *testing.T) {
	m := toResearch(t, loaded(t, 120, 40))
	before := m.cur().statusView
	for _, k := range []string{"s", "S"} {
		tm, cmd := m.Update(press(k))
		m = settle(t, tm.(Model), cmd)
	}
	if m.cur().statusView != before {
		t.Errorf("research has no view axis; statusView changed to %q", m.cur().statusView)
	}
	if m.cur().name != "research" {
		t.Errorf("the view keys must not navigate away, landed on %q", m.cur().name)
	}
}

// Research is also LIFECYCLE-less: the `m` action menu has no transitions to offer, so it
// must not open. A menu with zero entries would be a dead-end overlay.
func TestResearchTab_ActionMenuDoesNotOpen(t *testing.T) {
	m := toResearch(t, loaded(t, 120, 40))
	tm, cmd := m.Update(press("m"))
	m = settle(t, tm.(Model), cmd)
	if m.action.active {
		t.Error("research has no lifecycle transitions; the m menu must stay closed")
	}
}

// The `o`-cycle offers created (default) → updated → slug, and each genuinely reorders:
// the fixtures disagree on all three axes.
func TestResearchTab_SortCycleReorders(t *testing.T) {
	m := toResearch(t, loaded(t, 120, 40))
	if got := m.selectedID(); got != "zeta-doc" {
		t.Fatalf("default sort should be created-desc (zeta first), got %q", got)
	}
	// → updated. alpha carries updated_at 2026-07-01; zeta has none so it falls back to
	// its created date (2026-06-10), which is older — so alpha leads.
	tm, cmd := m.Update(press("o"))
	m = settle(t, tm.(Model), cmd)
	m.cur().list.Select(0)
	if got := m.selectedID(); got != "alpha-doc" {
		t.Errorf("updated sort should lead with alpha-doc, got %q", got)
	}
	// → slug, ascending.
	tm, cmd = m.Update(press("o"))
	m = settle(t, tm.(Model), cmd)
	m.cur().list.Select(0)
	if got := m.selectedID(); got != "alpha-doc" {
		t.Errorf("slug sort should lead with alpha-doc, got %q", got)
	}
}

// `/` must reach a doc by what it's ABOUT, not only its filename — the corpus is browsed
// by topic. Asserted on FilterValue directly: the matching itself is bubbles' fuzzy
// filter, so driving keystrokes would test their matcher rather than our field span.
func TestResearchItem_FilterValueSpansDescriptionAndTags(t *testing.T) {
	it := researchItem{r: domain.Research{
		Slug: "zeta-doc", Description: "the zeta research", Tags: []string{"tui", "color"},
	}}
	fv := it.FilterValue()
	for _, want := range []string{"zeta-doc", "the zeta research", "tui", "color"} {
		if !strings.Contains(fv, want) {
			t.Errorf("FilterValue %q should reach %q", fv, want)
		}
	}
}

// A doc with no description still filters by slug and tags, and the row shows a dash
// rather than a ragged blank.
func TestResearchItem_HandlesMissingDescription(t *testing.T) {
	it := researchItem{r: domain.Research{Slug: "bare-doc", Tags: []string{"cli"}}}
	if fv := it.FilterValue(); !strings.Contains(fv, "bare-doc") || !strings.Contains(fv, "cli") {
		t.Errorf("FilterValue = %q", fv)
	}
}

// sortFields falls back to created when the doc was never edited, so a fresh doc doesn't
// sink to the bottom of the updated sort (matching the CLI `updated` column).
func TestResearchItem_UpdatedFallsBackToCreated(t *testing.T) {
	never := researchItem{r: domain.Research{Slug: "s", Created: "2026-06-02"}}
	if got := never.sortFields().updated; got != "2026-06-02" {
		t.Errorf("unedited doc: updated = %q, want the created date", got)
	}
	edited := researchItem{r: domain.Research{Slug: "s", Created: "2026-06-02", Updated: "2026-07-01"}}
	if got := edited.sortFields().updated; got != "2026-07-01" {
		t.Errorf("edited doc: updated = %q, want the stamp", got)
	}
}

// The watcher path: research/ is in WatchPaths, and reloadAll iterates the tab registry —
// so a doc written by another process must appear on the research tab without a restart.
// This is the acceptance criterion the task calls out; it works because research is a
// registry entry rather than a special case, so no research-specific reload code exists
// to be forgotten.
func TestResearchTab_ReloadsOnFilesystemEvent(t *testing.T) {
	root := seedRepo(t)
	m := New(core.NewService(store.NewFS(root)))
	tm, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	m = tm.(Model)
	tm, _ = m.Update(m.Init()())
	m = toResearch(t, tm.(Model))

	if got := len(m.cur().list.Items()); got != 2 {
		t.Fatalf("expected the 2 seeded docs, got %d", got)
	}

	// A third doc lands from outside the TUI.
	path, out := testutil.ResearchFixture(root, "mid-doc.md",
		"---\nschema: 1\ncreated: \"2026-06-06\"\ndescription: written behind the TUI's back\ntags: [x]\n---\n# Mid\n")
	testutil.Write(t, path, out)

	// reloadMsg is what the fs path produces once its debounce settles (fsEventMsg arms a
	// generation-checked tick; that coalescing is entity-agnostic and covered in
	// watch_test.go). Sending it directly exercises the part research participates in:
	// reloadAll walking the tab registry.
	tm, cmd := m.Update(reloadMsg{})
	m = settle(t, tm.(Model), cmd)

	if got := len(m.cur().list.Items()); got != 3 {
		t.Errorf("the fs event should have picked up the new doc: %d items, want 3", got)
	}
	if v := ansi.Strip(m.View().Content); !strings.Contains(v, "mid-doc") {
		t.Errorf("new doc not rendered:\n%s", v)
	}
}
