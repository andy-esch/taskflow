package theme

import (
	"fmt"
	"time"
)

// RelativeDate renders a YYYY-MM-DD date as a compact "today" / "3d ago" /
// "2w ago" / "5mo ago" / "1y ago". Empty or unparseable input yields "". It's a
// shared presentation primitive — both the CLI render layer and the TUI use it.
func RelativeDate(date string) string { return relativeDateFrom(date, time.Now()) }

func relativeDateFrom(date string, now time.Time) string {
	t, err := time.Parse(time.DateOnly, date)
	if err != nil {
		return ""
	}
	days := int(now.Sub(t).Hours() / 24)
	switch {
	case days < 0:
		return date // future date — show it verbatim rather than "−3d"
	case days == 0:
		return "today"
	case days == 1:
		return "yesterday"
	case days < 7:
		return fmt.Sprintf("%dd ago", days)
	case days < 30:
		return fmt.Sprintf("%dw ago", days/7)
	case days < 365:
		return fmt.Sprintf("%dmo ago", days/30)
	default:
		return fmt.Sprintf("%dy ago", days/365)
	}
}
