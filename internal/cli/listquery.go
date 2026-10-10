package cli

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"github.com/andy-esch/taskflow/internal/cli/render"
	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/wire"
)

type listQueryFlags struct{ core.ListQuery }

func (q *listQueryFlags) bind(cmd *cobra.Command, dated bool) {
	cmd.Flags().StringVar(&q.Sort, "sort", "", "sort by id|name|created|updated; prefix with - for descending (findings: id|name only)")
	cmd.Flags().IntVar(&q.Limit, "limit", 0, "maximum returned rows (1-1000; omitted means unbounded)")
	cmd.Flags().StringVar(&q.Cursor, "cursor", "", "resume page.next_cursor with the same filters, sort, and limit; restart if results changed")
	cmd.Flags().StringVar(&q.Search, "search", "", "case-insensitive substring in list metadata, not full bodies (maximum 256 UTF-8 bytes)")
	dateHelp := "own updated date >= YYYY-MM-DD, falling back to creation; not a complete change feed"
	if !dated {
		dateHelp = "unsupported for findings (no own update date); requesting it is a validation error"
	}
	cmd.Flags().StringVar(&q.UpdatedSince, "updated-since", "", dateHelp)
	_ = cmd.RegisterFlagCompletionFunc("sort", func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
		values := []string{"id", "-id", "name", "-name"}
		if dated {
			values = append(values, "created", "-created", "updated", "-updated")
		}
		return values, cobra.ShellCompDirectiveNoFileComp
	})
}

func (q listQueryFlags) resolve(cmd *cobra.Command) (core.ListQuery, error) {
	if cmd.Flags().Changed("limit") && q.Limit == 0 {
		return core.ListQuery{}, fmt.Errorf("%w: explicit limit must be between 1 and %d", domain.ErrValidation, core.MaxListLimit)
	}
	for _, name := range []string{"sort", "cursor", "search", "updated-since"} {
		if cmd.Flags().Changed(name) && cmd.Flags().Lookup(name).Value.String() == "" {
			return core.ListQuery{}, fmt.Errorf("%w: --%s must not be empty", domain.ErrValidation, name)
		}
	}
	return q.ListQuery, nil
}

// Non-JSON output stays a normal row stream; continuation is only machine
// metadata, not a prose token clients would have to parse out of a table.
func listPageNotice(w io.Writer, page *core.ListPage) {
	if page != nil && page.NextCursor != "" {
		fmt.Fprintf(w, "list: returned %d of %d matching rows; use --json for page.next_cursor\n", page.Returned, page.Matched)
	}
}

func renderListPage[T any](app *App, mode outputMode, columns []string, items []T, problems []core.LoadProblem,
	page *core.ListPage, listKey string, cols []render.Column[T],
	jsonFn func(io.Writer, []T, []core.LoadProblem, *core.ListPage) error,
	humanFn func(io.Writer, render.Style, []T) error,
) error {
	jsonResult := func(w io.Writer, items []T, problems []core.LoadProblem) error {
		return jsonFn(w, items, problems, page)
	}
	if page == nil {
		return renderList(app, mode, columns, items, problems, listKey, cols, jsonResult, humanFn)
	}
	return renderListWithProblemsAndPage(app, mode, columns, items, problems, page, listKey, cols,
		wire.ToLintLoadProblemsJSON, jsonResult, humanFn, render.LintProblemsHuman)
}
