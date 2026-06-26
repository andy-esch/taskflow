package cli

import (
	"fmt"

	"github.com/andy-esch/taskflow/internal/cli/render"
)

// runMoves applies a per-slug transition to a target status/bucket, prints a
// per-item report (JSON or human), and returns a summary error wrapping the
// first failure for the exit code — every slug is attempted, no silent
// partials. The summary (not the raw error) goes back to main so the failure
// isn't printed twice: the per-item ✘ line already carried the detail. Shared
// by task and audit transitions so the loop + reporting policy live in exactly
// one place.
// An optional `decorate` callback fills extra MoveResult fields from the moved
// item (e.g. defer surfaces the would-be/recorded revisit_at) without coupling
// the shared loop to any one verb's payload; transitions that have nothing extra
// to report omit it.
func runMoves[T any](app *App, slugs []string, status string, move func(slug string) (T, error), slugOf func(T) string, decorate ...func(T, *render.MoveResult)) error {
	var chosenErr error
	failed := 0
	results := make([]render.MoveResult, 0, len(slugs))
	for _, slug := range slugs {
		res := render.MoveResult{Slug: slug, To: status}
		if got, err := move(slug); err != nil {
			res.Error = err.Error()
			failed++
			// Prefer a sentinel-bearing error (a meaningful exit code: 10/11/13/14)
			// over a generic exit-1 one, so the batch's summary code reports the most
			// actionable cause rather than whichever failure happened to be first in
			// argv. (The first sentinel wins; per-item ✘ lines carry the full detail.)
			if chosenErr == nil || (ExitCode(chosenErr) == 1 && ExitCode(err) != 1) {
				chosenErr = err
			}
		} else {
			res.Slug = slugOf(got)
			for _, d := range decorate {
				d(got, &res)
			}
		}
		results = append(results, res)
	}
	if app.JSON {
		if err := render.MovesJSON(app.Out, results, app.DryRun); err != nil {
			return err
		}
	} else {
		render.MovesHuman(app.Out, app.ErrOut, app.Style, results, app.DryRun)
	}
	if chosenErr != nil {
		// %w keeps the sentinel (exit-code mapping); the text is a count, not a
		// repeat of the already-printed detail.
		return fmt.Errorf("%d of %d transitions failed: %w", failed, len(slugs), chosenErr)
	}
	return nil
}
