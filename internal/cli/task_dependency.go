package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/andy-esch/taskflow/internal/cli/render"
	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/wire"
)

// dependencyCommandFailure adds adapter-owned workspace identity to the typed
// core failure so WriteError can preserve a resumable mutation receipt under
// --json. Error/Unwrap keep ordinary classification and human output intact.
type dependencyCommandFailure struct {
	cause     error
	receipt   core.DependencyMutationReceipt
	workspace wire.WorkspaceJSON
}

func (e *dependencyCommandFailure) Error() string { return e.cause.Error() }
func (e *dependencyCommandFailure) Unwrap() error { return e.cause }

func dependencyFailure(app *App, receipt core.DependencyMutationReceipt, err error) error {
	if err == nil {
		return nil
	}
	// Pre-write validation/conflict failures use the ordinary classified error
	// envelope. Edge outcomes describe the intended semantic delta, so embedding
	// them in a failed no-write receipt could falsely read as applied. Structured
	// mutation recovery is necessary only once a durable prefix exists.
	if len(receipt.AppliedTaskIDs) == 0 {
		return err
	}
	return &dependencyCommandFailure{cause: err, receipt: receipt, workspace: app.workspace()}
}

func newTaskDependCmd(app *App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "depend",
		Short: "Change repository-global task dependencies through the graph guard",
	}
	cmd.AddCommand(
		newTaskDependencyEdgeCmd(app, core.DependencyAdd),
		newTaskDependencyEdgeCmd(app, core.DependencyRemove),
		newTaskDependencyMigrateCmd(app),
	)
	return cmd
}

// dependencyEdgeArgs takes the one positional task and, when handed more, names the
// flag form instead of cobra's bare "accepts 1 arg(s), received 2". `depend add A B`
// is the shape the verb itself suggests, so two task-shaped arguments is a strong
// signal of exactly that mistake — and an arity count teaches nothing about where the
// second one was supposed to go.
func dependencyEdgeArgs(verb string) cobra.PositionalArgs {
	return func(_ *cobra.Command, args []string) error {
		switch {
		case len(args) == 1:
			return nil
		case len(args) == 0:
			return fmt.Errorf("%w: name the task whose prerequisites change — tskflwctl task depend %s <task> --on <prereq>",
				domain.ErrValidation, verb)
		default:
			return fmt.Errorf("%w: prerequisites go in --on, not positionally; did you mean: tskflwctl task depend %s %s --on %s?",
				domain.ErrValidation, verb, args[0], strings.Join(args[1:], " --on "))
		}
	}
}

func newTaskDependencyEdgeCmd(app *App, operation core.DependencyOperation) *cobra.Command {
	var prerequisites []string
	verb := string(operation)
	cmd := &cobra.Command{
		Use:               verb + " <task>",
		Short:             fmt.Sprintf("%s one or more hard prerequisites", map[core.DependencyOperation]string{core.DependencyAdd: "Add", core.DependencyRemove: "Remove"}[operation]),
		Example:           fmt.Sprintf("  tskflwctl task depend %s deploy --on build --on verify", verb),
		Args:              dependencyEdgeArgs(verb),
		Annotations:       map[string]string{"safety": "mutating"},
		ValidArgsFunction: app.completeTaskSlugs,
		RunE: func(_ *cobra.Command, args []string) error {
			var (
				receipt core.DependencyMutationReceipt
				err     error
			)
			if operation == core.DependencyAdd {
				receipt, err = app.Svc.AddTaskDependencies(args[0], prerequisites, app.DryRun)
			} else {
				receipt, err = app.Svc.RemoveTaskDependencies(args[0], prerequisites, app.DryRun)
			}
			if err != nil {
				return dependencyFailure(app, receipt, err)
			}
			if app.JSON {
				return render.DependencyMutationJSON(app.Out, receipt, app.workspace())
			}
			return render.DependencyMutationHuman(app.Out, app.Style, receipt)
		},
	}
	cmd.Flags().StringSliceVar(&prerequisites, "on", nil, "prerequisite task reference (repeat or comma-separate)")
	_ = cmd.RegisterFlagCompletionFunc("on", app.completeTaskSlugs)
	return cmd
}

func newTaskDependencyMigrateCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:         "migrate",
		Short:       "Convert all safe legacy dependency fields to canonical depends_on IDs",
		Long:        "Convert every legacy blocked_by, dependencies, and blocks field occurrence to canonical depends_on IDs and remove the legacy keys. Present-but-empty legacy keys are also removed. The repository-wide plan writes dependents before legacy blocks owners so every durable prefix remains conservative and a retry converges after interruption.",
		Example:     "  tskflwctl task depend migrate --dry-run --json\n  tskflwctl task depend migrate",
		Args:        cobra.NoArgs,
		Annotations: map[string]string{"safety": "mutating"},
		RunE: func(_ *cobra.Command, _ []string) error {
			receipt, err := app.Svc.MigrateTaskDependencies(app.DryRun)
			if err != nil {
				return dependencyFailure(app, receipt, err)
			}
			if app.JSON {
				return render.DependencyMutationJSON(app.Out, receipt, app.workspace())
			}
			return render.DependencyMutationHuman(app.Out, app.Style, receipt)
		},
	}
}

func newTaskBlockersCmd(app *App) *cobra.Command {
	var causal bool
	cmd := &cobra.Command{
		Use:               "blockers <task>",
		Short:             "Explain the actionable blockers for a task",
		Long:              "Explain a task's current derived role, gate, eligibility, and actionable blocker frontier. --causal selects the full forensic closure. Resolved legacy constraints participate in both projections, while graph health still reports degraded until they are migrated.",
		Example:           "  tskflwctl task blockers deploy\n  tskflwctl task blockers deploy --causal --json",
		Args:              cobra.ExactArgs(1),
		Annotations:       map[string]string{"safety": "read-only"},
		ValidArgsFunction: app.completeTaskSlugs,
		RunE: func(_ *cobra.Command, args []string) error {
			result, err := app.Svc.TaskBlockers(args[0], causal)
			if err != nil {
				return err
			}
			if app.JSON {
				return render.TaskBlockersJSON(app.Out, result)
			}
			return render.TaskBlockersHuman(app.Out, app.Style, result)
		},
	}
	cmd.Flags().BoolVar(&causal, "causal", false, "show the full causal blocker closure instead of the actionable frontier")
	return cmd
}

func newTaskUnblocksCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:               "unblocks <task>",
		Short:             "Show every task transitively downstream of this task",
		Long:              "Show the queried task's current derived state and every transitive downstream task with deterministic shortest paths. Resolved legacy constraints participate in the projection. This is current impact, not a promise that completing the source alone makes every result eligible.",
		Example:           "  tskflwctl task unblocks build\n  tskflwctl task unblocks build --json",
		Args:              cobra.ExactArgs(1),
		Annotations:       map[string]string{"safety": "read-only"},
		ValidArgsFunction: app.completeTaskSlugs,
		RunE: func(_ *cobra.Command, args []string) error {
			result, err := app.Svc.TaskUnblocks(args[0])
			if err != nil {
				return err
			}
			if app.JSON {
				return render.TaskUnblocksJSON(app.Out, result)
			}
			return render.TaskUnblocksHuman(app.Out, app.Style, result)
		},
	}
}
