// Package theme is the single source of truth for the *semantic* presentation
// of domain values — which glyph and which color represent a status, bucket,
// priority, or completion level. It imports only domain (no ANSI, no lipgloss),
// so both the CLI render layer and the TUI consume the same decisions and each
// maps Color to its own rendering tech.
package theme

import (
	"strings"

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

// MarkdownStyleFor picks the markdown style for the terminal background. Each
// caller resolves darkBG with its own background detection (the TUI once at
// startup; the CLI per `show`) and feeds it here, so the mapping lives in one
// place.
func MarkdownStyleFor(darkBG bool) string {
	if darkBG {
		return MarkdownStyleDark
	}
	return MarkdownStyleLight
}

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
// (open · in-progress · fixed · landed · deferred · superseded · wontfix);
// matching is case-insensitive. An empty/unknown status falls to the neutral dot
// (audit lint flags those separately).
func FindingStatus(s string) Token {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "open":
		return Token{"○", ColorYellow}
	case "in-progress":
		return Token{"●", ColorYellow}
	case "fixed", "landed":
		return Token{"✔", ColorGreen}
	case "deferred", "superseded":
		return Token{"◌", ColorGray}
	case "wontfix":
		return Token{"✘", ColorRed}
	default:
		return Token{"•", ColorGray}
	}
}

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
