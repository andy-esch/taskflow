package theme

import "fmt"

// The progress composite — "<bar>  <percent>  <done>/<total>" — is assembled per
// surface (a CLI table cell, a labeled `progress:` field, a TUI list row), so its
// LAYOUT is genuinely context-specific and not worth forcing into one renderer. Its
// bar and percent COLOR are already shared (progressbar.Render, theme.Percent); the
// remaining shared pieces are the percent and done/total NUMBER formats, kept here
// in one place so they can't drift (the "%d%% here, %3d%% there" inconsistency).
// Each surface still applies its own color (ANSI Style vs lipgloss) and bar width.

// PercentLabel renders a percent compactly ("7%") — for inline / prose contexts.
func PercentLabel(pct int) string { return fmt.Sprintf("%d%%", pct) }

// PercentLabelPadded right-justifies the percent to 3 digits ("  7%", " 70%",
// "100%") so it aligns in hand-laid-out columns / rows.
func PercentLabelPadded(pct int) string { return fmt.Sprintf("%3d%%", pct) }

// AuditPercentLabel qualifies an audit's percent as the SETTLED share — "80% settled"
// — so a bare number can't be misread as overall progress. An epic's percent is
// unambiguous (done/total), but an audit bands four dispositions, and "settled" is the
// one word that covers all the terminal ones: a finding ruled `wontfix` is as resolved
// as one `fixed`. The word is the domain's own (Audit.Settled), so 100% settled and
// "ready to close" are the same fact stated twice rather than two numbers to reconcile.
func AuditPercentLabel(pct int) string { return fmt.Sprintf("%d%% settled", pct) }

// AuditPercentLabelPadded is AuditPercentLabel right-justified to 3 percent digits
// ("  0% settled", "100% settled") so audit list rows align.
func AuditPercentLabelPadded(pct int) string { return fmt.Sprintf("%3d%% settled", pct) }

// Counts renders a done/total rollup ("7/12"). Width-justification for aligned
// columns is the caller's concern (CLI tables pad cells; the TUI measures + pads).
func Counts(done, total int) string { return fmt.Sprintf("%d/%d", done, total) }

// The segmented audit bar's bands, in render order. Each groups SEVERAL finding statuses
// under one colour deliberately: a reader takes in the shape at a glance, and seven hues
// would be a decode task rather than a glance. The glyphs are distinct as well as the
// colours, so the stacking survives `--color=never` and a mono terminal — which is also
// what lets a legend name them.
//
// Defined here rather than inside the bar so the legend that NAMES a band and the bar that
// DRAWS it read from one place.
func BandDone() Token    { return Token{"█", ColorGreen} }  // fixed · tracked
func BandActive() Token  { return Token{"▓", ColorYellow} } // in-progress
func BandDropped() Token { return Token{"▒", ColorGray} }   // deferred · superseded · wontfix
// BandOpen is the unfilled remainder rather than a disposition, so the bar paints it in the
// palette's dimmer empty-track tone; the glyph is what tells it from BandDropped.
func BandOpen() Token { return Token{"░", ColorGray} }
