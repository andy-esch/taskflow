package cli

import (
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
)

// completeFunc is cobra's ValidArgsFunction shape.
type completeFunc = func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective)

// completeSpaceIDs offers the registry's local labels without consulting cwd discovery.
// The registry exists independently of a planning repo, so this stays useful from any
// directory. A missing or malformed registry degrades to no candidates: completion must
// never turn a configuration problem into shell noise.
func completeSpaceIDs(registry *core.SpaceRegistryService) completeFunc {
	return func(_ *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		catalog, err := registry.Catalog()
		if err != nil {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		taken := make(map[string]bool, len(args))
		for _, arg := range args {
			taken[arg] = true
		}
		var out []string
		for _, space := range catalog.Entries {
			if !taken[space.ID] && strings.HasPrefix(space.ID, toComplete) {
				out = append(out, space.ID)
			}
		}
		sort.Strings(out)
		return out, cobra.ShellCompDirectiveNoFileComp
	}
}

// activeHelpArg is a ValidArgsFunction for a free-form positional (a title/area):
// it offers no candidates and suppresses file completion (which only misleads
// here), surfacing a one-line ActiveHelp hint instead. ActiveHelp shows only on
// shells that support it (bash V2) and respects the user's on/off config; it
// degrades to silence elsewhere.
func activeHelpArg(hint string) completeFunc {
	return func(cmd *cobra.Command, args []string, _ string) ([]string, cobra.ShellCompDirective) {
		var comps []string
		if len(args) == 0 && cobra.GetActiveHelpConfig(cmd) != "off" {
			comps = cobra.AppendActiveHelp(comps, hint)
		}
		return comps, cobra.ShellCompDirectiveNoFileComp
	}
}

// completeKinds offers the document kinds (task|epic|audit) — for `template
// list --kind` and the `template show` first positional.
func completeKinds(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
	return domain.SchemaKinds(), cobra.ShellCompDirectiveNoFileComp
}

// completeTemplateShowArgs completes `template show <kind> [name]`: the kind first,
// then that kind's template names.
func completeTemplateShowArgs(_ *cobra.Command, args []string, _ string) ([]string, cobra.ShellCompDirective) {
	switch len(args) {
	case 0:
		return domain.SchemaKinds(), cobra.ShellCompDirectiveNoFileComp
	case 1:
		return domain.TemplateNames(args[0]), cobra.ShellCompDirectiveNoFileComp
	}
	return nil, cobra.ShellCompDirectiveNoFileComp
}

// completeTemplateNames offers a kind's body-template names for `--template`
// (default first), with file completion suppressed. The set is registry-driven, so
// a new built-in (or, later, repo-local) template shows up here automatically.
func completeTemplateNames(kind string) completeFunc {
	return func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
		return domain.TemplateNames(kind), cobra.ShellCompDirectiveNoFileComp
	}
}

// isCompletionCommand reports whether cmd is cobra's hidden completion driver
// (`__complete`/`__completeNoDesc`), so PersistentPreRunE can stay non-fatal
// during shell completion.
func isCompletionCommand(cmd *cobra.Command) bool {
	switch cmd.Name() {
	case cobra.ShellCompRequestCmd, cobra.ShellCompNoDescRequestCmd:
		return true
	default:
		return false
	}
}

// entityCompleter delegates planning data and candidate policy to the application.
// Composition is deferred until Cobra has parsed the completed command's flags.
// Injected services, missing capabilities, and failed discovery never cause a
// controller to open a filesystem fallback.
func (a *App) entityCompleter(kind core.EntityKind, excludedState string) completeFunc {
	return func(_ *cobra.Command, args []string, prefix string) ([]string, cobra.ShellCompDirective) {
		svc := a.Svc
		if a.CompletionService != nil {
			var err error
			svc, err = a.CompletionService()
			if err != nil {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
		}
		if svc == nil {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		candidates, err := svc.CompleteEntities(core.CompletionRequest{
			Kind: kind, Prefix: prefix, Args: args, ExcludeState: excludedState,
		})
		if err != nil {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		return candidates, cobra.ShellCompDirectiveNoFileComp
	}
}

// taskCompleter excludes only records observed in the destination status.
// Malformed records with unknown state stay addressable for repair.
func (a *App) taskCompleter(exclude domain.Status) completeFunc {
	return a.entityCompleter(core.EntityTask, string(exclude))
}

func (a *App) completeThreadSlugs(cmd *cobra.Command, args []string, prefix string) ([]string, cobra.ShellCompDirective) {
	return a.entityCompleter(core.EntityThread, "")(cmd, args, prefix)
}

// auditCompleter excludes records observed in the destination bucket.
func (a *App) auditCompleter(exclude domain.AuditBucket) completeFunc {
	return a.entityCompleter(core.EntityAudit, string(exclude))
}

func (a *App) completeTaskSlugs(cmd *cobra.Command, args []string, prefix string) ([]string, cobra.ShellCompDirective) {
	return a.taskCompleter("")(cmd, args, prefix)
}

func (a *App) completeAuditSlugs(cmd *cobra.Command, args []string, prefix string) ([]string, cobra.ShellCompDirective) {
	return a.auditCompleter("")(cmd, args, prefix)
}

func (a *App) completeEpicIDs(cmd *cobra.Command, args []string, prefix string) ([]string, cobra.ShellCompDirective) {
	return a.entityCompleter(core.EntityEpic, "")(cmd, args, prefix)
}
