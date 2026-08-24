// Package theme is the single source of truth for the *semantic* presentation
// of domain values — which glyph and which color represent a status, bucket,
// priority, or completion level. It imports only domain (no ANSI, no lipgloss),
// so both the CLI render layer and the TUI consume the same decisions and each
// maps Color to its own rendering tech.
package theme

import (
	"strings"
	"time"

	"github.com/andy-esch/taskflow/internal/domain"
)

// Color is a semantic 16-color slot. Each presenter maps it to its tech: the
// CLI to an ANSI SGR code, the TUI to a lipgloss.Color.
type Color int

const (
	ColorNone Color = iota
	ColorRed
	ColorGreen
	ColorYellow
	ColorBlue
	ColorCyan
	ColorGray
)

// Markdown body styles are glamour standard-style names, shared by `show` (CLI)
// and the TUI detail pane so the rendered theme is identical in both: dracula on
// dark terminals, glamour's light style on light ones.
const (
	MarkdownStyleDark  = "dracula"
	MarkdownStyleLight = "light"
)

// Token is a glyph + color for an entity state.
type Token struct {
	Glyph string
	Color Color
}

// Status maps a task status to its glyph + color.
func Status(s domain.Status) Token {
	switch s {
	case domain.StatusInProgress:
		return Token{"●", ColorYellow}
	case domain.StatusNextUp:
		return Token{"●", ColorBlue}
	case domain.StatusReadyToStart:
		return Token{"○", ColorCyan}
	case domain.StatusCompleted:
		return Token{"✔", ColorGreen}
	case domain.StatusDeprecated:
		return Token{"✘", ColorRed}
	case domain.StatusDeferred:
		return Token{"◌", ColorGray}
	default:
		return Token{"•", ColorGray}
	}
}

// Bucket maps an audit bucket to its glyph + color. Like Status (and unlike the
// old color-only mapping), the bucket carries a distinct *shape* so its state
// survives a mono terminal / --color=never / colorblindness — and the glyphs are
// shared with the task vocabulary where the concepts line up: ✔ green = done
// (closed ≙ completed), ◌ gray = parked (deferred ≙ deferred).
func Bucket(b domain.AuditBucket) Token {
	switch b {
	case domain.AuditOpen:
		return Token{"◆", ColorYellow}
	case domain.AuditClosed:
		return Token{"✔", ColorGreen}
	case domain.AuditDeferred:
		return Token{"◌", ColorGray}
	default:
		return Token{"■", ColorNone}
	}
}

// FindingStatus maps an audit finding's status to its glyph + color — the audit
// analog of Status, drawing from the same vocabulary so a finding reads like a
// task: ● active, ✔ done, ◌ parked, ✘ killed. The status set is finding.go's
// (open · in-progress · fixed · tracked · deferred · superseded · wontfix);
// matching is case-insensitive. An empty/unknown status falls to the neutral dot
// (audit lint flags those separately).
func FindingStatus(s string) Token {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "open":
		return Token{"○", ColorYellow}
	case "in-progress":
		return Token{"●", ColorYellow}
	case "fixed":
		return Token{"✔", ColorGreen}
	case "tracked":
		// Resolved for this audit but not BY it: an arrow says the work moved on, where a
		// tick would claim it was done here.
		return Token{"→", ColorGreen}
	case "deferred", "superseded":
		return Token{"◌", ColorGray}
	case "wontfix":
		return Token{"✘", ColorRed}
	default:
		return Token{"•", ColorGray}
	}
}

// Liveness maps an epic's derived activity band (core.EpicSummary.Liveness, passed
// as its string value so theme stays domain-only) to a glyph + color. The shape
// carries the state through a mono terminal: ● working (live work, like an active
// task), ✦ fresh (a new bucket awaiting tasks), ○ dormant (drained and quiet). An
// unknown value falls to the neutral dot.
func Liveness(s string) Token {
	switch s {
	case "working":
		return Token{"●", ColorYellow}
	case "fresh":
		return Token{"✦", ColorBlue}
	case "dormant":
		return Token{"○", ColorGray}
	default:
		return Token{"•", ColorGray}
	}
}

// Markers are fixed glyph+color tokens for cross-surface annotations that aren't
// keyed by a domain value (unlike Status/Bucket/Liveness/FindingStatus): the
// non-conforming status attention ⚠, the revisit ↻, an audit's "ready to close"
// ✓, the all-clear ✔, and the unreadable-file !. Exposed so the row delegates, the
// dashboard, and the `?` legend draw ONE glyph+color per concept instead of re-typing
// them — and so the legend can't drift from the rows. ✓ (ready-to-close, U+2713) and
// ✔ (all-clear / done, U+2714) are deliberately DIFFERENT glyphs for different states.
var (
	MarkerWarn         = Token{"⚠", ColorYellow}
	MarkerRevisit      = Token{"↻", ColorYellow}
	MarkerReadyToClose = Token{"✓", ColorGreen}
	MarkerAllClear     = Token{"✔", ColorGreen}
	MarkerUnreadable   = Token{"!", ColorRed}
)

// Priority maps a priority label to its color.
func Priority(p string) Color {
	switch p {
	case "high":
		return ColorRed
	case "medium":
		return ColorYellow
	case "low":
		return ColorGray
	default:
		return ColorNone
	}
}

// Percent maps a completion percentage to its color: gray <34, yellow <100,
// green at 100.
// Staleness colours an elapsed duration by how much it should worry you. Built in the same
// shape as Percent so the palette owns the thresholds rather than each view inventing its
// own: a caller passes days and renders the result, and the boundaries move in one place.
//
// The thresholds are deliberately generous. Planning is not a ticketing system with an SLA,
// and a fortnight on a hard task is ordinary — so nothing is flagged until a month, which
// is roughly when "in progress" stops describing reality. A negative or unknown age is
// neutral, never alarming.
func Staleness(days int) Color {
	switch {
	case days < 0:
		return ColorGray
	case days >= 90:
		return ColorRed
	case days >= 30:
		return ColorYellow
	default:
		return ColorGray
	}
}

// DaysSince is the whole days between a YYYY-MM-DD date and now, or -1 when the date is
// missing or unparseable — the same "unknown is not alarming" convention Staleness expects.
func DaysSince(date string) int { return daysSinceFrom(date, time.Now()) }

func daysSinceFrom(date string, now time.Time) int {
	t, err := time.Parse(time.DateOnly, date)
	if err != nil {
		return -1
	}
	if d := int(now.Sub(t).Hours() / 24); d >= 0 {
		return d
	}
	return -1
}

func Percent(pct int) Color {
	switch {
	case pct >= 100:
		return ColorGreen
	case pct >= 34:
		return ColorYellow
	default:
		return ColorGray
	}
}
