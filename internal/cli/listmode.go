package cli

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"github.com/andy-esch/taskflow/internal/cli/render"
	"github.com/andy-esch/taskflow/internal/domain"
)

// outputMode is the resolved output format for a list command.
type outputMode int

const (
	modeHuman outputMode = iota
	modeJSON
	modeQuiet // -q: ids only, one per line
	modePlain // --plain: stable tab-separated table
)

// listMode binds the pipeline output-mode flags shared by the list commands
// (task/epic/audit) and resolves the chosen mode against the persistent --json.
type listMode struct{ quiet, plain bool }

func (m *listMode) bind(cmd *cobra.Command) {
	cmd.Flags().BoolVarP(&m.quiet, "quiet", "q", false, "ids only, one per line (for `| xargs`)")
	cmd.Flags().BoolVar(&m.plain, "plain", false,
		"stable tab-separated table with a header row (no color/truncation, absolute dates)")
	cmd.MarkFlagsMutuallyExclusive("quiet", "plain")
}

// resolve picks the output mode; --json, --quiet, and --plain are mutually
// exclusive (cobra guards quiet⇄plain; --json is persistent so it's checked here).
func (m listMode) resolve(app *App) (outputMode, error) {
	switch {
	case app.JSON && (m.quiet || m.plain):
		return modeHuman, fmt.Errorf("%w: choose at most one of --json, --quiet, --plain", domain.ErrValidation)
	case app.JSON:
		return modeJSON, nil
	case m.quiet:
		return modeQuiet, nil
	case m.plain:
		return modePlain, nil
	default:
		return modeHuman, nil
	}
}

// renderList writes a list result in the resolved mode — JSON envelope, ids-only
// (-q), plain table (--plain), or the human table — so the four-mode switch
// lives in ONE place instead of being copy-pasted across the task/epic/audit
// list commands. Problems go to stderr except in JSON mode, where they're
// embedded in the envelope. The caller still owns problemsError() for the exit
// code, since it knows whether a problem is fatal for that command.
func renderList[T any](
	app *App, mode outputMode, items []T, problems []domain.FileProblem,
	jsonFn func(io.Writer, []T, []domain.FileProblem) error,
	plainFn func(io.Writer, []T),
	humanFn func(io.Writer, render.Style, []T) error,
	idOf func(T) string,
) error {
	if mode == modeJSON {
		return jsonFn(app.Out, items, problems)
	}
	switch mode {
	case modeQuiet:
		ids := make([]string, len(items))
		for i, it := range items {
			ids[i] = idOf(it)
		}
		render.IDsQuiet(app.Out, ids)
	case modePlain:
		plainFn(app.Out, items)
	default:
		if err := humanFn(app.Out, app.Style, items); err != nil {
			return err
		}
	}
	render.ProblemsHuman(app.ErrOut, app.Style, problems)
	return nil
}
