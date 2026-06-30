package tui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

// TestFindStatus_HintsRawWhenPrettyMisses pins L16: a query that straddles a
// glamour wrap reads as 0 matches in pretty mode; when the raw render WOULD match,
// findStatus points at R so a real hit isn't mistaken for "not present".
func TestFindStatus_HintsRawWhenPrettyMisses(t *testing.T) {
	d := detailPane{pretty: true, rawStyled: "the needle is here", st: &testStyles}
	d.find.query = "needle" // no matches set → simulates the pretty-wrap miss

	if got := ansi.Strip(d.findStatus()); !strings.Contains(got, "R: raw") {
		t.Errorf("a pretty 0-match with a raw hit should hint R, got %q", got)
	}
	// No raw hit → no hint (R wouldn't help).
	d.rawStyled = "nothing relevant"
	if got := ansi.Strip(d.findStatus()); strings.Contains(got, "R: raw") {
		t.Errorf("no raw hit must not hint R, got %q", got)
	}
	// Already in raw mode → never hint.
	d.pretty = false
	d.rawStyled = "the needle is here"
	if got := ansi.Strip(d.findStatus()); strings.Contains(got, "R: raw") {
		t.Errorf("raw mode must not hint R, got %q", got)
	}
}
