package tui

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/store"
	"github.com/andy-esch/taskflow/internal/testutil"
)

// auditsTab switches to the audits tab and settles its load.
func auditsTab(t *testing.T, m Model) Model {
	t.Helper()
	for i := 0; i < len(m.tabs) && m.cur().name != "audits"; i++ {
		tm, cmd := m.Update(press("]"))
		m = drain(t, tm.(Model), cmd)
	}
	if m.cur().name != "audits" {
		t.Fatalf("could not reach the audits tab, on %q", m.cur().name)
	}
	return m
}

// TestModel_ActionMenuMovesAudit pins the M10 win: the registry-driven `m` menu
// now drives audit lifecycle (close/reopen/defer), not just tasks. Closing an
// open audit with no findings relocates it to the closed bucket on disk.
func TestModel_ActionMenuMovesAudit(t *testing.T) {
	m := loaded(t, 120, 40)
	m = auditsTab(t, m)
	if m.selectedID() != "2026-06-01-thing" {
		t.Fatalf("setup: want the seeded audit selected, got %q", m.selectedID())
	}
	tm, _ := m.Update(press("m"))
	m = tm.(Model)
	if !m.action.active {
		t.Fatal("a should open the action menu on an audit")
	}
	// Open audit: reopen (→open) is the no-op row and dropped; close + defer remain.
	offered := map[string]bool{}
	for _, tr := range m.action.options {
		offered[tr.verb] = true
	}
	if !offered["close"] || !offered["defer"] {
		t.Errorf("audit menu should offer close + defer, got %v", m.action.options)
	}
	if offered["reopen"] {
		t.Errorf("reopen must be dropped for an already-open audit, got %v", m.action.options)
	}
	m = cursorTo(t, m, "close")
	tm, cmd := m.Update(press("enter"))
	m = tm.(Model)
	if m.action.active {
		t.Error("a non-destructive apply should close the menu")
	}
	if cmd == nil {
		t.Fatal("apply should return a MoveAudit command")
	}
	tm, _ = m.Update(cmd())
	m = tm.(Model)
	if m.flash == "" || m.flashErr {
		t.Errorf("expected a success flash, got %q (err=%v)", m.flash, m.flashErr)
	}
	a, _, err := m.svc.ShowAudit("2026-06-01-thing")
	if err != nil || a.Bucket != domain.AuditClosed {
		t.Errorf("audit should be closed after the action: bucket=%s err=%v", a.Bucket, err)
	}
}

// TestModel_CommandVerbMovesAudit pins the `:`-verb half of M10: a verb is resolved
// against the ACTIVE tab's transition table, so `:close` moves an audit.
func TestModel_CommandVerbMovesAudit(t *testing.T) {
	m := loaded(t, 120, 40)
	m = auditsTab(t, m)
	tm, _ := m.Update(press(":"))
	m = tm.(Model)
	for _, r := range "close" {
		tm, _ = m.Update(press(string(r)))
		m = tm.(Model)
	}
	_, cmd := m.Update(press("enter"))
	if cmd == nil {
		t.Fatal(":close should fire a MoveAudit")
	}
	msg, ok := cmd().(movedMsg)
	if !ok || msg.to != string(domain.AuditClosed) {
		t.Fatalf(":close should yield a movedMsg → closed, got %T %+v", cmd(), cmd())
	}
}

// TestModel_AuditCloseBlockedByOpenFindings pins the M4 guard surfacing in the TUI:
// closing an audit that still has open findings flashes red (actionErrMsg) and the
// audit does not move — the same rule `audit lint`/the CLI enforce.
func TestModel_AuditCloseBlockedByOpenFindings(t *testing.T) {
	r := testutil.NewRepo(t)
	r.Audit("open", "2026-06-02-open.md",
		"---\narea: store\ndate: 2026-06-02\n---\n# Audit\n\n#### H1. thing  · **Status:** open\n")
	m := New(core.NewService(store.NewFS(r.Root)))
	tm, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	m = tm.(Model)
	tm, _ = m.Update(m.Init()())
	m = auditsTab(t, tm.(Model))
	if m.selectedID() != "2026-06-02-open" {
		t.Fatalf("setup: want the open-findings audit selected, got %q", m.selectedID())
	}

	tm, _ = m.Update(press("m"))
	m = cursorTo(t, tm.(Model), "close")
	tm, cmd := m.Update(press("enter"))
	m = tm.(Model)
	if cmd == nil {
		t.Fatal("close should still fire a (failing) MoveAudit command")
	}
	tm, _ = m.Update(cmd())
	m = tm.(Model)
	if !m.flashErr || m.flash == "" {
		t.Errorf("closing an audit with open findings should flash red, got %q (err=%v)", m.flash, m.flashErr)
	}
	// The audit stayed open on disk — the guard refused the move.
	a, _, err := m.svc.ShowAudit("2026-06-02-open")
	if err != nil || a.Bucket != domain.AuditOpen {
		t.Errorf("audit must remain open after a blocked close: bucket=%s err=%v", a.Bucket, err)
	}
	if !strings.Contains(m.flash, "open finding") {
		t.Errorf("the flash should explain the open-findings block, got %q", m.flash)
	}
}
