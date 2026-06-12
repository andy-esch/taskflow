package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/theme"
)

// transition is one lifecycle action — the verb a user knows from the CLI mapped
// to the status it moves a task to. This table is the single source of truth for
// both the action menu and the `:`-command verbs.
type transition struct {
	verb        string
	to          domain.Status
	destructive bool // requires a y/n confirm (an archiving move)
}

var transitions = []transition{
	{"start", domain.StatusInProgress, false},
	{"promote", domain.StatusNextUp, false},
	{"demote", domain.StatusReadyToStart, false},
	{"complete", domain.StatusCompleted, false},
	{"defer", domain.StatusDeferred, false},
	{"deprecate", domain.StatusDeprecated, true},
}

// transitionFor resolves a `:`-command verb to its transition.
func transitionFor(verb string) (transition, bool) {
	for _, tr := range transitions {
		if tr.verb == verb {
			return tr, true
		}
	}
	return transition{}, false
}

// transitionVerbs lists the lifecycle verbs for `:` Tab-completion.
func transitionVerbs() []string {
	v := make([]string, len(transitions))
	for i, tr := range transitions {
		v[i] = tr.verb
	}
	return v
}

// validTransitions are the moves offered for a task in status cur — every
// transition except the one that lands on cur (moving to the current status is a
// no-op, so it's never worth a menu row).
func validTransitions(cur domain.Status) []transition {
	out := make([]transition, 0, len(transitions))
	for _, tr := range transitions {
		if tr.to != cur {
			out = append(out, tr)
		}
	}
	return out
}

// actionMenu is the lifecycle action palette (S4): opened on a task, it lists the
// valid transitions, vim-selects one, and applies it — a destructive choice first
// gates on a y/n confirm. It's a modal like the `?` help and `:` command bar: the
// model routes every key to it while active and floats it over the body.
type actionMenu struct {
	active  bool
	slug    string       // the task being acted on
	options []transition // the rows (a single entry when a `:`-verb opened the confirm directly)
	cursor  int
	confirm bool // a destructive choice is awaiting y/n
}

// open shows the transition menu for slug (currently in status cur).
func (a *actionMenu) open(slug string, cur domain.Status) {
	*a = actionMenu{active: true, slug: slug, options: validTransitions(cur)}
}

// openConfirm jumps straight to the y/n gate for one verb — used when a `:`
// command names a destructive verb explicitly (no menu to pick from).
func (a *actionMenu) openConfirm(slug string, tr transition) {
	*a = actionMenu{active: true, slug: slug, options: []transition{tr}, confirm: true}
}

func (a *actionMenu) close() { a.active = false }

func (a *actionMenu) move(d int) {
	if n := len(a.options); n > 0 {
		a.cursor = ((a.cursor+d)%n + n) % n
	}
}

func (a actionMenu) selected() transition { return a.options[a.cursor] }

// confirmOnly reports whether the menu is a bare `:`-verb confirm (one row) with
// no list to fall back to — so `n`/Esc closes it rather than returning to a menu.
func (a actionMenu) confirmOnly() bool { return len(a.options) == 1 }

var (
	actionBorder  = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("6")).Padding(0, 2)
	dangerBorder  = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("1")).Padding(0, 2)
	actionHeading = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("6"))
)

// view renders the menu (or confirm prompt) as a centered box + hint line, ready
// to composite over the body with overlay(). Clamped to (maxW, maxH).
func (a actionMenu) view(maxW, maxH int) string {
	slug := truncate(a.slug, max(maxW-8, 12))
	if a.confirm {
		tr := a.selected()
		body := fg(theme.ColorRed, tr.verb+"?") + "\n\n" + slug + dim(" → "+string(tr.to))
		box := dangerBorder.Render(body)
		hint := dim("y confirm · n/esc cancel")
		return clampBox(lipgloss.JoinVertical(lipgloss.Center, box, hint), maxW, maxH)
	}
	var b strings.Builder
	b.WriteString(actionHeading.Render("move " + slug))
	b.WriteString("\n\n")
	for i, tr := range a.options {
		label := tr.verb + dim(" → "+string(tr.to))
		if tr.destructive {
			label += " " + fg(theme.ColorRed, "⚠")
		}
		if i == a.cursor {
			b.WriteString(selectedStyle.Render("› ") + label + "\n")
		} else {
			b.WriteString("  " + label + "\n")
		}
	}
	box := actionBorder.Render(strings.TrimRight(b.String(), "\n"))
	hint := dim("↑↓/jk select · ⏎ apply · esc cancel")
	return clampBox(lipgloss.JoinVertical(lipgloss.Center, box, hint), maxW, maxH)
}

// clampBox bounds an overlay box to the body so a tiny terminal can't overflow.
func clampBox(s string, maxW, maxH int) string {
	return lipgloss.NewStyle().MaxWidth(maxW).MaxHeight(maxH).Render(s)
}
