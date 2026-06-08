package cli

import "github.com/andy-esch/taskflow/internal/cli/render"

// runMoves applies a per-slug transition to a target status/bucket, prints a
// per-item report (JSON or human), and returns the first error for the exit
// code — every slug is attempted, no silent partials. Shared by task and audit
// transitions so the loop + reporting policy live in exactly one place.
func runMoves[T any](app *App, slugs []string, status string, move func(slug string) (T, error), slugOf func(T) string) error {
	var firstErr error
	results := make([]render.MoveResult, 0, len(slugs))
	for _, slug := range slugs {
		res := render.MoveResult{Slug: slug, Status: status}
		if got, err := move(slug); err != nil {
			res.Error = err.Error()
			if firstErr == nil {
				firstErr = err
			}
		} else {
			res.Slug = slugOf(got)
		}
		results = append(results, res)
	}
	if app.JSON {
		if err := render.MovesJSON(app.Out, results); err != nil {
			return err
		}
	} else {
		render.MovesHuman(app.Out, results)
	}
	return firstErr
}
