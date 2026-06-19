package render

import (
	"fmt"
	"io"
	"strings"

	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
)

// Pipeline output modes (decided 2026-06-17). `-q/--quiet` is ids-only, one per
// line (for `| xargs`). `--plain` is a STABLE tab-separated table with a header
// row — no ANSI, no truncation/padding, absolute dates — a documented contract
// under the one global schema_version (a column add/reorder is a schema bump).

// IDsQuiet writes one id per line (the `-q` output).
func IDsQuiet(w io.Writer, ids []string) {
	for _, id := range ids {
		fmt.Fprintln(w, id)
	}
}

// writePlain writes a header row + tab-separated rows. Tabs/newlines inside a
// cell are flattened to spaces so the one-record-per-line contract always holds.
func writePlain(w io.Writer, header []string, rows [][]string) {
	flat := strings.NewReplacer("\t", " ", "\n", " ", "\r", " ")
	fmt.Fprintln(w, strings.Join(header, "\t"))
	for _, r := range rows {
		cells := make([]string, len(r))
		for i, c := range r {
			cells[i] = flat.Replace(c)
		}
		fmt.Fprintln(w, strings.Join(cells, "\t"))
	}
}

// TasksPlain writes the --plain table for tasks (absolute updated date).
func TasksPlain(w io.Writer, tasks []domain.Task) {
	rows := make([][]string, 0, len(tasks))
	for _, t := range tasks {
		updated := t.Updated
		if updated == "" {
			updated = t.Created
		}
		rows = append(rows, []string{
			t.Slug, string(t.Status), fmt.Sprintf("%d", t.Tier), t.Priority, t.Epic, updated, t.Description,
		})
	}
	writePlain(w, []string{"slug", "status", "tier", "priority", "epic", "updated", "description"}, rows)
}

// EpicsPlain writes the --plain table for epics (done/total as separate columns
// so scripts get numbers, not the "2/3 (66%)" human cell).
func EpicsPlain(w io.Writer, epics []core.EpicSummary) {
	rows := make([][]string, 0, len(epics))
	for _, e := range epics {
		rows = append(rows, []string{
			e.Epic.ID, e.Epic.Status, e.Epic.Priority, fmt.Sprintf("%d", e.Done), fmt.Sprintf("%d", e.Total), e.Epic.Description,
		})
	}
	writePlain(w, []string{"id", "status", "priority", "done", "total", "description"}, rows)
}

// AuditsPlain writes the --plain table for audits.
func AuditsPlain(w io.Writer, audits []domain.Audit) {
	rows := make([][]string, 0, len(audits))
	for _, a := range audits {
		rows = append(rows, []string{
			a.Slug, string(a.Bucket), a.Area, a.Date, fmt.Sprintf("%d", a.Findings), fmt.Sprintf("%d", a.OpenFindings),
		})
	}
	writePlain(w, []string{"slug", "bucket", "area", "date", "findings", "open"}, rows)
}
