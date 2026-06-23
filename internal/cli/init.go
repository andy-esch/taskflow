package cli

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/andy-esch/taskflow/internal/cli/prompt"
	"github.com/andy-esch/taskflow/internal/cli/render"
	"github.com/andy-esch/taskflow/internal/config"
	"github.com/andy-esch/taskflow/internal/domain"
)

func newInitCmd(app *App) *cobra.Command {
	var (
		path         string
		planningRepo string
		tracks       []string
		noLinkBack   bool
	)
	cmd := &cobra.Command{
		Use:         "init",
		Short:       "Scaffold a planning tree here, or point at an external planning repo",
		Args:        cobra.NoArgs,
		Annotations: map[string]string{"safety": "mutating"},
		Example: "  tskflwctl init\n" +
			"  tskflwctl init --path ./planning\n" +
			"  tskflwctl init --planning-repo ../desirelines-planning",
		// init may scaffold a NEW planning repo, so it must NOT require an existing
		// one — its own PersistentPreRunE overrides the root's resolve() (skips
		// discovery) and just sets up styling + the Gate/Prompter. The here-vs-
		// elsewhere prompt is TTY-gated (Gate.On()), so a headless agent/pipe never
		// hangs: with --planning-repo it goes straight to pointer mode, otherwise it
		// falls back to the full scaffold (today's non-interactive behavior).
		PersistentPreRunE: func(*cobra.Command, []string) error { app.setStyle(); return nil },
		RunE: func(cmd *cobra.Command, _ []string) error {
			abs, err := filepath.Abs(path)
			if err != nil {
				return err
			}
			pointer, repo, err := app.resolveInitTarget(planningRepo, cmd.Flags().Changed("planning-repo"))
			if err != nil {
				return err
			}
			if pointer {
				// tracked_repos lives in a PLANNING repo; a pointer repo only points.
				if len(tracks) > 0 {
					return fmt.Errorf("%w: --track records repos a PLANNING repo tracks; it can't combine with --planning-repo (pointer mode)", domain.ErrValidation)
				}
				return runInitPointer(app, abs, repo, !noLinkBack)
			}
			// --no-link-back is pointer-only; reject it in scaffold mode for symmetry
			// with the --track guard above (don't silently ignore a misused flag).
			if cmd.Flags().Changed("no-link-back") {
				return fmt.Errorf("%w: --no-link-back only applies with --planning-repo (pointer mode)", domain.ErrValidation)
			}
			return runInitScaffold(app, abs, tracks)
		},
	}
	cmd.Flags().StringVar(&path, "path", ".", "directory to initialize")
	cmd.Flags().StringVar(&planningRepo, "planning-repo", "",
		"point this repo at an external planning repo (relative to --path, or absolute): writes a pointer config, no tree")
	cmd.Flags().StringSliceVar(&tracks, "track", nil,
		"record an impl repo this planning repo tracks (repeatable; scaffold mode only)")
	cmd.Flags().BoolVar(&noLinkBack, "no-link-back", false,
		"pointer mode: don't add this repo to the planning repo's tracked_repos")
	return cmd
}

// resolveInitTarget decides init's mode (the flag-twin pattern). --planning-repo
// (flagSet) means pointer mode outright. Otherwise, on a TTY (gate open) ask
// here-vs-elsewhere; off a TTY default to scaffold — the headless contract, so an
// agent/pipe never blocks. Extracted so the interactive branch is unit-testable
// with a Fake prompter (PersistentPreRunE's setStyle would otherwise reset Gate).
func (a *App) resolveInitTarget(planningRepo string, flagSet bool) (pointer bool, repo string, err error) {
	if flagSet {
		return true, planningRepo, nil
	}
	if !a.Gate.On() {
		return false, "", nil
	}
	where, err := a.Prompt.SelectOne("Where does this repo's planning live?", []prompt.Option{
		{Label: "Here — scaffold a planning tree in this repo", Value: "here"},
		{Label: "Another repo — point at an external planning repo", Value: "elsewhere"},
	})
	if err != nil {
		return false, "", err
	}
	if where != "elsewhere" {
		return false, "", nil
	}
	repo, err = a.Prompt.Text("Path to the planning repo (relative or absolute)", "../planning")
	if err != nil {
		return false, "", err
	}
	return true, repo, nil
}

// runInitScaffold writes a full planning tree + config under abs, then records
// any --track impl repos in its tracked_repos (deduped, surgical).
func runInitScaffold(app *App, abs string, tracks []string) error {
	created, err := config.Init(abs, app.DryRun)
	if err != nil {
		return err
	}
	var tracked []string
	for _, tr := range tracks {
		added, err := config.AddTrackedRepo(abs, tr, app.DryRun)
		if err != nil {
			return err
		}
		if added {
			tracked = append(tracked, tr)
		}
	}
	if app.JSON {
		return render.InitJSON(app.Out, render.InitEnvelope{
			DryRun: app.DryRun, Mode: "scaffold", Root: abs, Tracked: tracked, Created: created,
		})
	}
	if len(created) == 0 && len(tracked) == 0 {
		fmt.Fprintf(app.Out, "%s already initialized: %s\n", app.Style.Dim("·"), abs)
		return nil
	}
	verb := "initialized"
	switch {
	case app.DryRun && len(created) == 0:
		verb = "would update"
	case app.DryRun:
		verb = "would initialize"
	case len(created) == 0:
		verb = "updated"
	}
	fmt.Fprintf(app.Out, "%s %s %s\n", app.Style.Green("✔"), verb, app.Style.Bold(abs))
	for _, c := range created {
		fmt.Fprintf(app.Out, "  %s %s\n", app.Style.Dim("+"), c)
	}
	for _, tr := range tracked {
		fmt.Fprintf(app.Out, "  %s tracks %s\n", app.Style.Dim("+"), app.Style.Bold(tr))
	}
	if len(created) > 0 {
		fmt.Fprintf(app.Out, "\n%s\n", app.Style.Dim(`→ next: tskflwctl epic new "Title" --description "..."`))
	}
	return nil
}

// runInitPointer writes a pointer config under abs (no tree), validating the
// external planning repo first, then (unless opted out) links back by recording
// this repo in the planning repo's tracked_repos.
func runInitPointer(app *App, abs, planningRepo string, linkBack bool) error {
	created, err := config.InitPointer(abs, planningRepo, app.DryRun)
	if err != nil {
		return err
	}
	var back string
	var linkErr error
	if linkBack {
		back, linkErr = config.LinkBack(abs, planningRepo, app.DryRun)
	}
	// Link-back is best-effort: the pointer config is already written, so a hiccup
	// (e.g. the planning repo isn't writable) warns rather than fails the init. The
	// warning goes to stderr, after the success line, so it never corrupts --json
	// stdout and reads in order on a combined terminal.
	warn := func() {
		if linkErr != nil {
			fmt.Fprintf(app.ErrOut, "%s link-back skipped: %v\n", app.Style.Warn("⚠"), linkErr)
		}
	}
	if app.JSON {
		warn()
		return render.InitJSON(app.Out, render.InitEnvelope{
			DryRun: app.DryRun, Mode: "pointer", Root: abs, PlanningRepo: planningRepo, LinkedBack: back, Created: created,
		})
	}
	if len(created) > 0 {
		verb := "pointed"
		if app.DryRun {
			verb = "would point"
		}
		fmt.Fprintf(app.Out, "%s %s %s at planning repo %s\n",
			app.Style.Green("✔"), verb, app.Style.Bold(abs), app.Style.Bold(planningRepo))
	} else {
		fmt.Fprintf(app.Out, "%s already initialized: %s\n", app.Style.Dim("·"), abs)
	}
	if back != "" {
		verb := "linked back"
		if app.DryRun {
			verb = "would link back"
		}
		fmt.Fprintf(app.Out, "  %s %s — %s now tracks this repo as %s\n",
			app.Style.Dim("+"), verb, app.Style.Bold(planningRepo), app.Style.Bold(back))
	}
	warn()
	if len(created) > 0 {
		fmt.Fprintf(app.Out, "\n%s\n", app.Style.Dim("→ next: run tskflwctl from here — planning resolves to the pointed repo"))
	}
	return nil
}
