package tui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"

	"github.com/andy-esch/taskflow/internal/domain"
)

// TestDetailPane_GlamourRendererCachedByWidth pins the fast-follow fix: the pane
// reuses its compiled renderer across selections at the same width, and only
// rebuilds it when the width changes.
func TestDetailPane_GlamourRendererCachedByWidth(t *testing.T) {
	var d detailPane
	d.width = 60
	d.prettyBody("# A")
	r1 := d.glam
	if r1 == nil {
		t.Fatal("a renderer should be cached after the first render")
	}
	d.prettyBody("# B") // same width → reuse, no recompile
	if d.glam != r1 {
		t.Error("a same-width render must reuse the cached renderer")
	}
	d.width = 80
	d.prettyBody("# C") // width changed → rebuild
	if d.glam == r1 {
		t.Error("a width change must rebuild the renderer")
	}
}

func TestGlamourBody(t *testing.T) {
	out := glamourBody("# Title\n\nsome text", 60, "dark")
	if !strings.Contains(ansi.Strip(out), "Title") {
		t.Errorf("glamour should render the heading text, got %q", ansi.Strip(out))
	}
	if glamourBody("", 60, "dark") != "" {
		t.Error("empty markdown should render to empty")
	}
}

// TestDetailPane_GlamourRendererRebuildsOnStyle pins that the cached renderer is
// keyed by style as well as width, so a background-driven style applies.
func TestDetailPane_GlamourRendererRebuildsOnStyle(t *testing.T) {
	d := newDetailPane(&testStyles, "dark")
	d.width = 60
	d.prettyBody("# A")
	r1 := d.glam
	if r1 == nil {
		t.Fatal("a renderer should be cached after the first render")
	}
	d.glamStyle = "light"
	d.prettyBody("# B") // style changed → rebuild
	if d.glam == r1 {
		t.Error("a style change must rebuild the renderer")
	}
}

// seedDetail puts a markdown body on the detail pane of a loaded model.
func seedDetail(t *testing.T, m Model, body string) Model {
	t.Helper()
	id := m.selectedID()
	tm, _ := m.Update(detailMsg{kind: entityTasks, id: id, gen: m.detailGen,
		content: taskDetail{t: domain.Task{Slug: id}, body: body}})
	return tm.(Model)
}

// TestModel_GlamourToggleAndIndicator: pretty (glamour) is the default and renders
// markdown (the ** bold markers are consumed); R flips to raw (markers return) and
// flags it in the title; R again returns to pretty.
func TestModel_GlamourToggleAndIndicator(t *testing.T) {
	m := loaded(t, 120, 30)
	m = seedDetail(t, m, "## Heading\n\nsome **bold** text")

	if !m.detail.pretty {
		t.Fatal("detail should default to pretty (glamour)")
	}
	if strings.Contains(ansi.Strip(m.detail.styled), "**bold**") {
		t.Error("pretty mode should render markdown, not show the ** markers")
	}

	tm, _ := m.Update(press("R"))
	m = tm.(Model)
	if m.detail.pretty {
		t.Fatal("R should switch to raw")
	}
	if !strings.Contains(ansi.Strip(m.detail.styled), "**bold**") {
		t.Error("raw mode should show the literal markdown")
	}
	if !strings.Contains(m.detailTitle(), "raw") {
		t.Errorf("the title should flag raw mode, got %q", m.detailTitle())
	}

	tm, _ = m.Update(press("R"))
	m = tm.(Model)
	if !m.detail.pretty {
		t.Error("R should toggle back to pretty")
	}
}

// TestModel_GlamourPrefPersistsAcrossSelection: the raw/pretty choice is a pane
// preference, so it survives a selection change.
func TestModel_GlamourPrefPersistsAcrossSelection(t *testing.T) {
	m := loaded(t, 120, 30)
	tm, _ := m.Update(press("R")) // → raw
	m = tm.(Model)
	if m.detail.pretty {
		t.Fatal("R should set raw mode")
	}
	tm, _ = m.Update(press("j")) // move selection (fires a new detail load)
	m = tm.(Model)
	m = seedDetail(t, m, "**still raw**")
	if m.detail.pretty {
		t.Error("the raw preference should persist across selection changes")
	}
	if !strings.Contains(ansi.Strip(m.detail.styled), "**still raw**") {
		t.Error("the newly selected item should also render raw")
	}
}

// TestModel_FindWorksInBothModes: find runs over the rendered body in either mode,
// and toggling recomputes the matches (the ANSI-aware highlighter handles glamour).
func TestModel_FindWorksInBothModes(t *testing.T) {
	m := loaded(t, 120, 30)
	m = seedDetail(t, m, "alpha has a needle here\n\nand another needle below")
	tm, _ := m.Update(press("l")) // focus detail
	m = tm.(Model)

	apply := func(m Model, q string) Model {
		tm, _ := m.Update(press("/"))
		m = tm.(Model)
		for _, r := range q {
			tm, _ = m.Update(press(string(r)))
			m = tm.(Model)
		}
		tm, _ = m.Update(press("enter"))
		return tm.(Model)
	}

	m = apply(m, "needle")
	if n := len(m.detail.find.matches); n != 2 {
		t.Errorf("pretty: expected 2 'needle' matches, got %d", n)
	}
	tm, _ = m.Update(press("R")) // → raw; matches recomputed over the raw body
	m = tm.(Model)
	if n := len(m.detail.find.matches); n != 2 {
		t.Errorf("raw: find should recompute and still find 2, got %d", n)
	}
}
