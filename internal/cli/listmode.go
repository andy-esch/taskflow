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
)

// listMode binds the output flags shared by the list commands (task/epic/audit)
// and resolves the chosen format. The format axis is ONE flag —
// `-o/--output {human,json,name,table}` — with `--json` (persistent, universal)
// and `-q` kept as documented aliases for `-o json` / `-o name`. The projection
// axis is a second, orthogonal flag: `-c/--columns` selects table columns and
// implies `-o table`. See the consolidate-output-flags task for the design and
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
	cmd.Flags().StringVarP(&m.output, "output", "o", "", "output format: human|json|name|table")
	cmd.Flags().StringSliceVarP(&m.columns, "columns", "c", nil,
		"select table columns, comma-separated (implies -o table); available: "+specNames(columnSpecs))
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

	// -c implies table. If the user ALSO pinned a non-table format, that's the
	// clearer error to raise than the generic conflict below.
	if len(m.columns) > 0 {
		for mode, flag := range want {
			if mode != modeTable {
				return 0, fmt.Errorf("%w: --columns applies to -o table, not %s", domain.ErrValidation, flag)
			}
		}
		want[modeTable] = "--columns"
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
	default:
		return 0, fmt.Errorf("%w: unknown output format %q (valid: human, json, name, table)", domain.ErrValidation, s)
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
	cols []render.Column[T],
	jsonFn func(io.Writer, []T, []domain.FileProblem) error,
	humanFn func(io.Writer, render.Style, []T) error,
) error {
	if mode == modeJSON {
		return jsonFn(app.Out, items, problems)
	}
	switch mode {
	case modeName:
		ids := make([]string, len(items))
		for i, it := range items {
			ids[i] = cols[0].Extract(it) // first column is the id (slug / epic id)
		}
		render.IDsQuiet(app.Out, ids)
	case modeTable:
		sel, err := render.SelectColumns(cols, columns)
		if err != nil {
			return err
		}
		render.WriteTablePlain(app.Out, sel, items)
	default: // modeHuman
		if err := humanFn(app.Out, app.Style, items); err != nil {
			return err
		}
	}
	render.ProblemsHuman(app.ErrOut, app.Style, problems)
	return nil
}

// completeOutputFormats offers the four output formats with descriptions
// (KeepOrder so the shell shows them in this deliberate order, not sorted).
func completeOutputFormats(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
	return []string{
			"human\tcolored, aligned table (default)",
			"json\tstable JSON envelope",
			"name\tids only, one per line",
			"table\theadered tab-separated table",
		},
		cobra.ShellCompDirectiveNoFileComp | cobra.ShellCompDirectiveKeepOrder
}

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
		var out []string
		for _, s := range specs {
			if used[s.Name] || !strings.HasPrefix(s.Name, last) {
				continue
			}
			cand := s.Name
			if prefix != "" {
				cand = prefix + "," + s.Name
			}
			out = append(out, cand+"\t"+s.Desc)
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
