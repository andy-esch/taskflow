package tui

import (
	"strings"
	"testing"

	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
)

func epicSum(id, status string, total, done int) core.EpicSummary {
	return core.EpicSummary{Epic: domain.Epic{ID: id, Status: status}, Total: total, Done: done}
}

// TestFilterEpicsByView pins the epics-tab status axis: the default view keeps only
// active buckets, an exact status filters to it, and "all" keeps everything — the
// epic echo of loadTaskList's view switch, but on the stored status field.
func TestFilterEpicsByView(t *testing.T) {
	in := []core.EpicSummary{
		epicSum("a-active", domain.EpicStatusActive, 1, 0),
		epicSum("b-retired", domain.EpicStatusRetired, 1, 1),
		epicSum("c-deprecated", domain.EpicStatusDeprecated, 1, 0),
		epicSum("d-foreign", "planning", 1, 0), // non-canonical: must FAIL OPEN into the live view
	}
	cases := []struct {
		view string
		want []string
	}{
		// Default "live" view fails open: keeps active AND the foreign status, hides
		// only the known terminals.
		{"", []string{"a-active", "d-foreign"}},
		{"retired", []string{"b-retired"}},
		{"deprecated", []string{"c-deprecated"}},
		{"all", []string{"a-active", "b-retired", "c-deprecated", "d-foreign"}},
	}
	for _, c := range cases {
		got := filterEpicsByView(in, c.view)
		if len(got) != len(c.want) {
			t.Fatalf("view %q: ids len = %d, want %d", c.view, len(got), len(c.want))
		}
		for i, w := range c.want {
			if got[i].Epic.ID != w {
				t.Fatalf("view %q: ids = %v, want %v", c.view, ids(got), c.want)
			}
		}
	}
}

// TestSortEpicsForView pins the default-view recede: dormant (drained) epics float
// to the bottom while live ones keep their incoming (store) order; a non-default
// view is left untouched.
func TestSortEpicsForView(t *testing.T) {
	mk := func() []core.EpicSummary {
		return []core.EpicSummary{
			epicSum("dormant", domain.EpicStatusActive, 2, 2), // drained → dormant
			epicSum("working", domain.EpicStatusActive, 2, 0), // open → live
			epicSum("fresh", domain.EpicStatusActive, 0, 0),   // no tasks → live
		}
	}

	def := mk()
	sortEpicsForView(def, "")
	if got := ids(def); got[len(got)-1] != "dormant" {
		t.Errorf("default view: dormant must sink to the bottom, got %v", got)
	}
	if got := ids(def); got[0] != "working" || got[1] != "fresh" {
		t.Errorf("default view: live epics must keep store order, got %v", got)
	}

	all := mk()
	sortEpicsForView(all, "all")
	if got := ids(all); got[0] != "dormant" {
		t.Errorf("non-default view must be left untouched, got %v", got)
	}
}

// TestEpicGlyphAndNote pins the warning-triangle affordance: a non-conforming epic
// leads with ⚠ and shows its offending status; a conforming one shows neither (just
// its liveness glyph, no status note).
func TestEpicGlyphAndNote(t *testing.T) {
	foreign := core.EpicSummary{Epic: domain.Epic{ID: "x", Status: "planning"}, Total: 1}
	if g := epicGlyph(foreign); !strings.Contains(g, "⚠") {
		t.Errorf("non-conforming epic glyph = %q, want a ⚠", g)
	}
	if n := epicStatusNote(foreign); !strings.Contains(n, "planning") {
		t.Errorf("non-conforming status note = %q, want it to name the bad status", n)
	}

	ok := core.EpicSummary{Epic: domain.Epic{ID: "y", Status: domain.EpicStatusActive}, Total: 1}
	if g := epicGlyph(ok); strings.Contains(g, "⚠") {
		t.Errorf("conforming epic glyph = %q, must not warn", g)
	}
	if n := epicStatusNote(ok); n != "" {
		t.Errorf("conforming status note = %q, want empty", n)
	}

	// Empty status is non-conforming too, and renders the em-dash placeholder.
	if n := epicStatusNote(core.EpicSummary{Epic: domain.Epic{ID: "z", Status: ""}}); !strings.Contains(n, "—") {
		t.Errorf("empty status note = %q, want the — placeholder", n)
	}
}

func ids(es []core.EpicSummary) []string {
	out := make([]string, len(es))
	for i, e := range es {
		out[i] = e.Epic.ID
	}
	return out
}
