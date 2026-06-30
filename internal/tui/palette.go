package tui

import (
	"strings"

	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// The command palette (ctrl+p): a floating fuzzy launcher over one flat index of
// everything — every loaded task/epic/audit (jump to it) plus the `:` command
// words (switch tab/view, run a lifecycle verb). You open it, type, and Enter
// either jumps or dispatches. It's the discovery surface: instead of remembering
// keys or scanning the help, type what you want. `:` stays as the terse path.

type paletteKind int

const (
	palJump    paletteKind = iota // jump to an entity (ek, id)
	palCommand                    // run a `:` command word (word)
)

// paletteItem is one row in the palette index. title is shown; filter is what the
// fuzzy matcher sees (slug + entity kind, or the bare command word).
type paletteItem struct {
	kind   paletteKind
	ek     entityKind // palJump: which tab to land on
	id     string     // palJump: the entity id/slug
	word   string     // palCommand: the `:` word to dispatch
	title  string
	filter string
}

func (i paletteItem) Title() string       { return i.title }
func (i paletteItem) Description() string { return "" }
func (i paletteItem) FilterValue() string { return i.filter }

// palette is the modal state: a query input over a candidate list. We drive the
// filtering ourselves (input → DefaultFilter → SetItems) so typing narrows
// immediately, rather than going through bubbles/list's `/` filter mode.
type palette struct {
	input  textinput.Model
	list   list.Model
	all    []paletteItem
	active bool
}

func newPalette() palette {
	ti := textinput.New()
	ti.Prompt = "› "
	ti.Placeholder = "jump to anything / run a command…"
	ti.CharLimit = 64

	d := list.NewDefaultDelegate()
	d.ShowDescription = false // one compact line per candidate
	d.SetSpacing(0)
	l := list.New(nil, d, 0, 0)
	l.SetShowTitle(false)
	l.SetShowStatusBar(false)
	l.SetShowHelp(false)
	l.SetFilteringEnabled(false) // we own filtering via the input
	return palette{input: ti, list: l}
}

// open seeds the index, sizes the list, and focuses the query.
func (p *palette) open(all []paletteItem, w, h int) tea.Cmd {
	p.active = true
	p.all = all
	p.input.SetValue("")
	p.list.SetSize(w, h)
	p.refilter()
	return p.input.Focus()
}

func (p *palette) close() {
	p.active = false
	p.input.Blur()
}

// reindex swaps in a fresh candidate set (a tab finished loading while the palette
// was open), preserving the current query.
func (p *palette) reindex(all []paletteItem) {
	p.all = all
	p.refilter()
}

// refilter ranks the index against the current query (fuzzy, via bubbles/list's
// DefaultFilter) and feeds the ordered subset to the list. Empty query = all.
func (p *palette) refilter() {
	q := strings.TrimSpace(p.input.Value())
	var items []list.Item
	if q == "" {
		items = make([]list.Item, len(p.all))
		for i := range p.all {
			items[i] = p.all[i]
		}
	} else {
		targets := make([]string, len(p.all))
		for i, it := range p.all {
			targets[i] = it.FilterValue()
		}
		for _, r := range list.DefaultFilter(q, targets) {
			items = append(items, p.all[r.Index])
		}
	}
	p.list.SetItems(items)
}

// view stacks the query over the candidate list in a bordered box, with a hint
// line — clamped to the overlay budget (mirrors the action/follow modals).
func (p palette) view(s styles, maxW, maxH int) string {
	box := s.actionBorder.Render(lipgloss.JoinVertical(lipgloss.Left, p.input.View(), p.list.View()))
	hint := s.dim("type to filter · ↑↓ select · ⏎ go · esc cancel")
	return clampBox(lipgloss.JoinVertical(lipgloss.Left, box, hint), maxW, maxH)
}
