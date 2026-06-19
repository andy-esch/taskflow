package render

import (
	"fmt"
	"io"
	"strings"
)

// Pipeline output primitives shared by the list commands. `-o name` (and its
// `-q` alias) is ids-only, one per line (for `| xargs`). `-o table` is a STABLE
// tab-separated table with a header row — no ANSI, no truncation/padding,
// absolute dates — a documented contract under the one global schema_version (a
// column add/reorder is a schema bump). The per-entity columns live in
// columns.go.

// IDsQuiet writes one id per line (the `-o name` / `-q` output).
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
