package cli

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/andy-esch/taskflow/internal/cli/render"
	"github.com/andy-esch/taskflow/internal/domain"
)

// outputMode is the resolved output format for a list command.
type outputMode int

const (
	modeHuman outputMode = iota // colored, aligned table (default)
	modeJSON                    // stable JSON envelope
	modeName                    // ids only, one per line (the -q alias)
	modeTable                   // headered, byte-stable, tab-separated table
	modeCSV                     // headered, RFC 4180 comma-separated (for spreadsheets)
)

// isProjectable reports whether a format honors `-c/--columns` field projection:
// the table/csv views and the JSON envelope (where the selection becomes
// column-named string fields). modeName/modeHuman are not projectable.
func isProjectable(m outputMode) bool { return m == modeTable || m == modeCSV || m == modeJSON }

// listMode binds the output flags shared by the list commands (task/epic/audit)
// and resolves the chosen format. The format axis is ONE flag —
// `-o/--output {human,json,name,table,csv}` — with `--json` (persistent,
// universal) and `-q` kept as documented aliases for `-o json` / `-o name`. The
// projection axis is a second, orthogonal flag: `-c/--columns` selects columns
// for the projectable formats (table/csv and the json envelope) and implies
// `-o table` when no format is pinned. See the
// consolidate-output-flags task for the design and
// the research behind the split (no value-internal DSL, so columns stay
// shell-completable).
type listMode struct {
	output  string
	columns []string
	quiet   bool
}

// bind registers the output flags and their completion. columnSpecs is the
// entity's column set — it drives `-c` completion and the help text — so each
// list command passes its own (e.g. render.Specs(render.TaskColumns())).
func (m *listMode) bind(cmd *cobra.Command, columnSpecs []render.ColumnSpec) {
	cmd.Flags().StringVarP(&m.output, "output", "o", "", "output format: human|json|name|table|csv")
	cmd.Flags().StringSliceVarP(&m.columns, "columns", "c", nil,
		"select columns for -o table/csv/json, comma-separated (implies -o table); available: "+specNames(columnSpecs))
	cmd.Flags().BoolVarP(&m.quiet, "quiet", "q", false, "ids only, one per line (alias for -o name)")
	_ = cmd.RegisterFlagCompletionFunc("output", completeOutputFormats)
	_ = cmd.RegisterFlagCompletionFunc("columns", columnCompleter(columnSpecs))
}

// resolve reduces the format surfaces — `-o`, the `--json`/`-q` aliases, and the
// table-implying `-c` — to a single outputMode, erroring (exit 11) on any
// conflict or an unknown `-o` value. It needs cmd to tell an explicit `-o` from
// the default. Centralizing every conflict here (rather than cobra's
// presence-based MarkFlagsMutuallyExclusive) keeps them all on exit 11.
func (m listMode) resolve(cmd *cobra.Command, app *App) (outputMode, error) {
	// want maps each requested format to the flag that asked for it, so a
	// conflict can name the culprits. Aliases that AGREE collapse to one key.
	want := map[outputMode]string{}
	if app.JSON {
		want[modeJSON] = "--json"
	}
	if m.quiet {
		want[modeName] = "-q/--quiet"
	}
	if cmd.Flags().Changed("output") {
		mode, err := parseFormat(m.output)
		if err != nil {
			return 0, err
		}
		want[mode] = "--output " + m.output
	}

	// -c needs a columnar format (table/csv); it implies table when none is
	// pinned. If the user pinned a NON-columnar format, that's the clearer error
	// to raise than the generic conflict below — name every offender in a stable
	// order (map iteration is randomized).
	if len(m.columns) > 0 {
		var bad []string
		projectable := false
		for mode, flag := range want {
			if isProjectable(mode) {
				projectable = true
			} else {
				bad = append(bad, flag)
			}
		}
		if len(bad) > 0 {
			sort.Strings(bad)
			return 0, fmt.Errorf("%w: --columns applies to -o table, -o csv, or --json, not %s",
				domain.ErrValidation, strings.Join(bad, ", "))
		}
		// A pinned json/csv/table is projected in place; bare `-c` (no format)
		// still implies table — the human-facing default projection target.
		if !projectable {
			want[modeTable] = "--columns"
		}
	}

	switch len(want) {
	case 0:
		return modeHuman, nil
	case 1:
		for mode := range want {
			return mode, nil
		}
	}
	return 0, fmt.Errorf("%w: choose at most one output format (%s)", domain.ErrValidation, conflictList(want))
}

// parseFormat maps an `-o/--output` value to a mode.
func parseFormat(s string) (outputMode, error) {
	switch s {
	case "human":
		return modeHuman, nil
	case "json":
		return modeJSON, nil
	case "name":
		return modeName, nil
	case "table":
		return modeTable, nil
	case "csv":
		return modeCSV, nil
	default:
		return 0, fmt.Errorf("%w: unknown output format %q (valid: human, json, name, table, csv)", domain.ErrValidation, s)
	}
}

// conflictList renders the conflicting flags in a stable (sorted) order.
func conflictList(want map[outputMode]string) string {
	flags := make([]string, 0, len(want))
	for _, f := range want {
		flags = append(flags, f)
	}
	sort.Strings(flags)
	return strings.Join(flags, ", ")
}

// renderList writes a list result in the resolved mode. The column registry is
// the single source of truth: modeName projects the first (id) column, modeTable
// projects the `-c` selection (or all columns), and the others defer to the
// supplied JSON/human renderers. Problems go to stderr except in JSON mode,
// where the envelope embeds them. The caller still owns problemsError() for the
// exit code, since it knows whether a problem is fatal for that command.
func renderList[T any](
	app *App, mode outputMode, columns []string, items []T, problems []domain.FileProblem,
	listKey string, cols []render.Column[T],
	jsonFn func(io.Writer, []T, []domain.FileProblem) error,
	humanFn func(io.Writer, render.Style, []T) error,
) error {
	if mode == modeJSON {
		// `--json -c …` narrows each row to the selected columns (as column-named
		// string fields) while keeping the schema_version + unreadable envelope;
		// bare `--json` emits the full typed envelope via jsonFn.
		if len(columns) > 0 {
			sel, err := render.SelectColumns(cols, columns)
			if err != nil {
				return err
			}
			return render.ProjectedListJSON(app.Out, listKey, sel, items, problems)
		}
		return jsonFn(app.Out, items, problems)
	}
	switch mode {
	case modeName:
		ids := make([]string, len(items))
		for i, it := range items {
			ids[i] = cols[0].Extract(it) // first column is the id (slug / epic id)
		}
		render.IDsQuiet(app.Out, ids)
	case modeTable, modeCSV:
		sel, err := render.SelectColumns(cols, columns)
		if err != nil {
			return err
		}
		if mode == modeCSV {
			if err := render.WriteCSV(app.Out, sel, items); err != nil {
				return err
			}
		} else {
			render.WriteTablePlain(app.Out, sel, items)
		}
	default: // modeHuman
		if err := humanFn(app.Out, app.Style, items); err != nil {
			return err
		}
	}
	render.ProblemsHuman(app.ErrOut, app.Style, problems)
	return nil
}

// completeOutputFormats offers the four output formats with descriptions
// (KeepOrder so the shell shows them in this deliberate order, not sorted). Built
// from cobra's typed helpers rather than hand-joined "name\tdesc" strings.
var completeOutputFormats = cobra.FixedCompletions([]cobra.Completion{
	cobra.CompletionWithDesc("human", "colored, aligned table (default)"),
	cobra.CompletionWithDesc("json", "stable JSON envelope"),
	cobra.CompletionWithDesc("name", "ids only, one per line"),
	cobra.CompletionWithDesc("table", "headered tab-separated table"),
	cobra.CompletionWithDesc("csv", "headered comma-separated (RFC 4180)"),
}, cobra.ShellCompDirectiveNoFileComp|cobra.ShellCompDirectiveKeepOrder)

// columnCompleter completes a comma-separated `-c` value over the known columns:
// it completes the column after the last comma, drops columns already chosen,
// and keeps the cursor mid-token (NoSpace) so the next `,column` can be typed.
// Because the value is its own token with no internal `=`/parens, this works
// where kubectl's `custom-columns=` and a `table(...)` DSL can't.
func columnCompleter(specs []render.ColumnSpec) completeFunc {
	return func(_ *cobra.Command, _ []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		parts := strings.Split(toComplete, ",")
		prefix := strings.Join(parts[:len(parts)-1], ",") // columns chosen so far
		last := parts[len(parts)-1]                       // the column being typed
		used := make(map[string]bool, len(parts))
		for _, p := range parts[:len(parts)-1] {
			used[p] = true
		}
		var out []cobra.Completion
		for _, s := range specs {
			if used[s.Name] || !strings.HasPrefix(s.Name, last) {
				continue
			}
			cand := s.Name
			if prefix != "" {
				cand = prefix + "," + s.Name
			}
			out = append(out, cobra.CompletionWithDesc(cand, s.Desc))
		}
		return out, cobra.ShellCompDirectiveNoSpace | cobra.ShellCompDirectiveNoFileComp | cobra.ShellCompDirectiveKeepOrder
	}
}

// specNames joins column names for the `--columns` help text.
func specNames(specs []render.ColumnSpec) string {
	names := make([]string, len(specs))
	for i, s := range specs {
		names[i] = s.Name
	}
	return strings.Join(names, ",")
}
