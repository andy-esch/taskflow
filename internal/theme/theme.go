// Package theme is the single source of truth for the *semantic* presentation
// of domain values — which glyph and which color represent a status, bucket,
// priority, or completion level. It imports only domain (no ANSI, no lipgloss),
// so both the CLI render layer and the TUI consume the same decisions and each
// maps Color to its own rendering tech.
package theme

import "github.com/andy-esch/taskflow/internal/domain"

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

// Bucket maps an audit bucket to its color.
func Bucket(b domain.AuditBucket) Color {
	switch b {
	case domain.AuditOpen:
		return ColorYellow
	case domain.AuditClosed:
		return ColorGreen
	case domain.AuditDeferred:
		return ColorGray
	default:
		return ColorNone
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
