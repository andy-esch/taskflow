package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestModel_YankSlug: y copies the selected entity's slug and flashes a green
// confirmation. We assert the flash (which carries the exact text yank copies) and
// that a clipboard cmd is issued, but don't invoke it — the cmd may shell out to a
// native clipboard tool, a side effect we keep out of the test run.
func TestModel_YankSlug(t *testing.T) {
	m := loaded(t, 80, 24)
	want := m.selectedID()
	if want == "" {
		t.Fatal("setup: expected a selected task")
	}
	tm, cmd := m.Update(press("y"))
	m = tm.(Model)

	if m.flash != "copied slug: "+want || m.flashErr {
		t.Errorf("flash = %q (err=%v), want %q", m.flash, m.flashErr, "copied slug: "+want)
	}
	if cmd == nil {
		t.Fatal("y must return a clipboard cmd")
	}
}

// TestModel_YankPath: Y copies the selected entity's on-disk file path.
func TestModel_YankPath(t *testing.T) {
	m := loaded(t, 80, 24)
	wantPath := m.selectedPath()
	if wantPath == "" || !strings.HasSuffix(wantPath, ".md") {
		t.Fatalf("setup: expected a .md file path, got %q", wantPath)
	}
	tm, cmd := m.Update(press("Y"))
	m = tm.(Model)

	if m.flash != "copied path: "+wantPath || m.flashErr {
		t.Errorf("flash = %q (err=%v), want %q", m.flash, m.flashErr, "copied path: "+wantPath)
	}
	if cmd == nil {
		t.Fatal("Y must return a clipboard cmd")
	}
}

// TestModel_YankEmpty: with nothing to copy, yank flashes an error and issues no
// clipboard cmd (rather than copying an empty string).
func TestModel_YankEmpty(t *testing.T) {
	var m Model
	got, cmd := m.yank("", "slug")
	mm := got.(Model)
	if !mm.flashErr || mm.flash != "nothing to copy" {
		t.Errorf("flash = %q (err=%v), want the 'nothing to copy' error", mm.flash, mm.flashErr)
	}
	if cmd != nil {
		t.Error("an empty yank must not issue a clipboard cmd")
	}
}

// TestCopyToClipboard_FallsBackToOSC52: with no native clipboard tool on PATH
// (the SSH/remote shape), copyToClipboard falls back to OSC 52 — and its payload
// is the text. Empty PATH guarantees no native tool runs, so this is side-effect-free.
func TestCopyToClipboard_FallsBackToOSC52(t *testing.T) {
	t.Setenv("PATH", "")
	cmd := copyToClipboard("epic-7")
	if cmd == nil {
		t.Fatal("expected a clipboard cmd")
	}
	// The OSC 52 cmd's message is bubbletea's setClipboardMsg, a string type whose
	// rendered value is the payload.
	if got := fmt.Sprint(cmd()); got != "epic-7" {
		t.Errorf("OSC 52 fallback payload = %q, want %q", got, "epic-7")
	}
}

func TestClipboardArgv_NoneWhenPathEmpty(t *testing.T) {
	t.Setenv("PATH", "")
	if argv := clipboardArgv(); argv != nil {
		t.Errorf("no clipboard tool should be found with an empty PATH, got %v", argv)
	}
}

// TestClipboardArgv_FindsToolOnPath: a tool present on PATH is selected (here a
// stub wl-copy; pbcopy isn't on the isolated PATH, so detection falls to it).
func TestClipboardArgv_FindsToolOnPath(t *testing.T) {
	dir := t.TempDir()
	stub := filepath.Join(dir, "wl-copy")
	if err := os.WriteFile(stub, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir)
	argv := clipboardArgv()
	if len(argv) == 0 || argv[0] != "wl-copy" {
		t.Errorf("clipboardArgv = %v, want [wl-copy]", argv)
	}
}

// stubClipboardTool drops an executable `pbcopy` stub (first in clipboardArgv's
// order, so it wins the lookup) on a temp dir and prepends that dir to PATH — the
// real PATH stays so the stub can still use /bin tools like cat.
func stubClipboardTool(t *testing.T, script string) {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "pbcopy"), []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
}

// TestCopyToClipboard_NativeSuccess: a present, working native tool copies and
// emits NO OSC52 fallback (the cmd returns nil).
func TestCopyToClipboard_NativeSuccess(t *testing.T) {
	stubClipboardTool(t, "#!/bin/sh\ncat >/dev/null\n") // drain stdin, succeed
	if msg := copyToClipboard("hello")(); msg != nil {
		t.Errorf("a working native tool should not emit an OSC52 msg, got %v", msg)
	}
}

// TestCopyToClipboard_NativeFailFallsBack: a present-but-failing native tool falls
// back to OSC52 carrying the same payload.
func TestCopyToClipboard_NativeFailFallsBack(t *testing.T) {
	stubClipboardTool(t, "#!/bin/sh\nexit 1\n")
	if got := fmt.Sprint(copyToClipboard("hello")()); got != "hello" {
		t.Errorf("native failure should fall back to the OSC52 payload, got %q", got)
	}
}

// loadTab switches tabs with key and runs the resulting reload so the destination
// tab's list is populated (tabs load lazily on first visit).
func loadTab(t *testing.T, m Model, key string) Model {
	t.Helper()
	tm, cmd := m.Update(press(key))
	m = tm.(Model)
	if cmd != nil {
		tm, _ = m.Update(cmd())
		m = tm.(Model)
	}
	return m
}

// TestModel_YankAcrossEntities: yank works on epics and audits, not only tasks —
// the wiring is generic via entityItem.id()/path().
func TestModel_YankAcrossEntities(t *testing.T) {
	m := loadTab(t, loaded(t, 80, 24), "]") // tasks → epics
	if m.cur().kind != entityEpics {
		t.Fatalf("expected the epics tab, got %v", m.cur().kind)
	}
	wantID := m.selectedID()
	if wantID == "" {
		t.Fatal("no epic selected")
	}
	tm, _ := m.Update(press("y"))
	m = tm.(Model)
	if m.flash != "copied slug: "+wantID || m.flashErr {
		t.Errorf("epic yank flash = %q, want %q", m.flash, "copied slug: "+wantID)
	}

	m = loadTab(t, m, "]") // epics → audits
	if m.cur().kind != entityAudits {
		t.Fatalf("expected the audits tab, got %v", m.cur().kind)
	}
	wantPath := m.selectedPath()
	if wantPath == "" || !strings.HasSuffix(wantPath, ".md") {
		t.Fatalf("expected an audit .md path, got %q", wantPath)
	}
	tm, _ = m.Update(press("Y"))
	m = tm.(Model)
	if m.flash != "copied path: "+wantPath || m.flashErr {
		t.Errorf("audit yank flash = %q, want %q", m.flash, "copied path: "+wantPath)
	}
}
