package tui

import (
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/theme"
)

// transition is one lifecycle action — the verb a user knows from the CLI mapped
// to the destination state it moves the selection to. `to` is a string (a task
// status OR an audit bucket); the tab's applyMove closure interprets it for the
// right entity, so the menu/`:`-verb machinery is entity-agnostic. Each entity
// declares its own table in the registry (entity.go) rather than this being
// task-only.
//
// The task/audit verb→state mappings now live in domain.TaskTransitions()/
// domain.AuditTransitions() (the shared registry the CLI also reads); this local
// shape is a thin TUI-facing view of domain.Transition, so the rest of action.go
// (the menu render, the y/n confirm gate) is unchanged. Epics aren't in the
// registry — they have no CLI verb vocabulary to share — so epicTransitions stays
// declared inline below.
type transition struct {
	verb        string
	to          string
	destructive bool // requires a y/n confirm (an archiving move)
}

// fromDomain maps the shared domain.Transition table to the TUI's local shape.
func fromDomain(ts []domain.Transition) []transition {
	out := make([]transition, len(ts))
	for i, t := range ts {
		out[i] = transition{verb: t.Verb, to: t.To, destructive: t.Destructive}
	}
	return out
}

// taskTransitions are the task status moves (the working-set lifecycle), sourced
// from the shared domain registry so the CLI and TUI can't drift.
var taskTransitions = fromDomain(domain.TaskTransitions())

// auditTransitions are the audit bucket moves, mirroring `audit close/reopen/defer`.
// close/defer to a non-open bucket are the ones the store guards on still-open
// findings (M4) — that rejection surfaces as an actionErrMsg, which is correct.
// Sourced from the shared domain registry (see taskTransitions).
var auditTransitions = fromDomain(domain.AuditTransitions())

// epicTransitions are the epic status moves, mirroring `epic move <id> <status>`.
// Unlike task/audit, epic status is a frontmatter FIELD not a directory, so the
// move (svc.MoveEpic) rewrites the field in place — no file relocates. None of the
// three is destructive (no archive-style y/n gate): retiring/deprecating an epic
// is a reversible status flip, not a file move.
var epicTransitions = []transition{
	{"activate", "active", false},
	{"retire", "retired", false},
	{"deprecate", "deprecated", false},
}

// transitionFor resolves a `:`-command verb to its transition within a given
// table (the active tab's), so verbs are scoped to the entity in view.
func transitionFor(transitions []transition, verb string) (transition, bool) {
	for _, tr := range transitions {
		if tr.verb == verb {
			return tr, true
		}
	}
	return transition{}, false
}

// validTransitions are the moves offered from state cur — every transition except
// the one that lands on cur (a no-op, never worth a menu row). cur is the current
// task status or audit bucket as a string.
func validTransitions(transitions []transition, cur string) []transition {
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
	// A task defer opens a revisit ("snooze until") sub-state instead of applying
	// at once — the TUI face of the CLI's `task defer` date prompt. dateInput takes
	// an absolute YYYY-MM-DD or a relative offset (2w/10d); dateErr shows a parse
	// error in place (keeping what was typed) until the next keystroke.
	revisit   bool
	dateInput textinput.Model
	dateErr   string
}

// open shows the transition menu for slug from state cur (its current status or
// bucket), offering the given entity's transition table minus the no-op row.
func (a *actionMenu) open(slug string, transitions []transition, cur string) {
	opts := validTransitions(transitions, cur)
	if len(opts) == 0 {
		return // nothing to offer — don't open an empty menu, so selected() never indexes nil
	}
	*a = actionMenu{active: true, slug: slug, options: opts}
}

// openConfirm jumps straight to the y/n gate for one verb — used when a `:`
// command names a destructive verb explicitly (no menu to pick from).
func (a *actionMenu) openConfirm(slug string, tr transition) {
	*a = actionMenu{active: true, slug: slug, options: []transition{tr}, confirm: true}
}

// beginRevisit switches the menu into the revisit-date sub-state for slug (also
// usable cold, from a `:defer`/palette verb with no menu open) and focuses the
// date input. Esc from here returns to the menu when one is open (options set),
// else closes — see handleActionKey.
func (a *actionMenu) beginRevisit(slug string) tea.Cmd {
	a.active = true
	a.slug = slug
	a.confirm = false
	a.revisit = true
	a.dateErr = ""
	ti := textinput.New()
	ti.Prompt = ""
	ti.Placeholder = "YYYY-MM-DD, or 2w / 10d · blank to skip"
	ti.CharLimit = 32
	ti.SetWidth(40)
	a.dateInput = ti
	// Focus the STORED field, not the local copy — textinput.Focus has a pointer
	// receiver, so focusing `ti` here would leave a.dateInput unfocused and silently
	// reject every keystroke.
	return a.dateInput.Focus()
}

func (a *actionMenu) close() {
	a.active = false
	a.revisit = false
	a.dateErr = ""
	a.dateInput.Blur()
}

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
	actionBorder  = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(pal.BorderActive.Color()).Padding(0, 2)
	dangerBorder  = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(pal.Danger.Color()).Padding(0, 2)
	actionHeading = lipgloss.NewStyle().Bold(true).Foreground(pal.Heading.Color())
)

// view renders the menu (or confirm prompt) as a centered box + hint line, ready
// to composite over the body with overlay(). Clamped to (maxW, maxH).
func (a actionMenu) view(maxW, maxH int) string {
	slug := truncate(a.slug, max(maxW-8, 12))
	if a.revisit {
		var b strings.Builder
		b.WriteString(actionHeading.Render("defer " + slug))
		b.WriteString("\n\n")
		b.WriteString("revisit date " + dim("(snooze until — optional)") + "\n")
		b.WriteString(a.dateInput.View())
		if a.dateErr != "" {
			b.WriteString("\n" + fg(theme.ColorRed, a.dateErr))
		}
		box := actionBorder.Render(b.String())
		hint := dim("⏎ apply · esc back · blank = no date")
		return clampBox(lipgloss.JoinVertical(lipgloss.Center, box, hint), maxW, maxH)
	}
	if a.confirm {
		tr := a.selected()
		q := fg(theme.ColorRed, tr.verb+"?") + " " + slug + dim(" → "+tr.to)
		// Put the y/n prompt INSIDE the danger box and bold the keys — a dim hint
		// below the box was easy to miss ("is it waiting on me?").
		prompt := helpKeyStyle.Render("y") + " confirm   " + helpKeyStyle.Render("n") + "/" + helpKeyStyle.Render("esc") + " cancel"
		box := dangerBorder.Render(q + "\n\n" + prompt)
		return clampBox(box, maxW, maxH)
	}
	var b strings.Builder
	b.WriteString(actionHeading.Render("move " + slug))
	b.WriteString("\n\n")
	for i, tr := range a.options {
		label := tr.verb + dim(" → "+tr.to)
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
