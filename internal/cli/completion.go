package cli

import (
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/andy-esch/taskflow/internal/config"
	"github.com/andy-esch/taskflow/internal/domain"
)

// completeFunc is cobra's ValidArgsFunction shape.
type completeFunc = func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective)

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

// planningRoot resolves the planning root for completion, tolerant of being
// outside a repo (returns ok=false). It does its own discovery rather than
// relying on the lazily-built service, so completion works even when
// PersistentPreRunE found no repo.
func (a *App) planningRoot() (string, bool) {
	start, err := a.startDir()
	if err != nil {
		return "", false
	}
	cfg, err := config.Discover(start)
	if err != nil {
		return "", false
	}
	return cfg.Root, true
}

// slugsFromGlobs returns the .md filename stems matching any pattern, keeping
// only those with the typed prefix and dropping any already on the command
// line. It parses no YAML — the slug *is* the filename — so completion is fast
// and works even when a file's frontmatter is malformed (the case you most want
// to complete while fixing it). Because status/bucket *is* the directory, the
// caller selects which dirs to glob to filter by state — still without parsing.
func slugsFromGlobs(patterns []string, prefix string, taken []string) []string {
	seen := make(map[string]bool, len(taken))
	for _, t := range taken {
		seen[t] = true
	}
	var out []string
	for _, pat := range patterns {
		matches, err := filepath.Glob(pat)
		if err != nil {
			continue
		}
		for _, m := range matches {
			slug := strings.TrimSuffix(filepath.Base(m), ".md")
			if slug == "" || seen[slug] || !strings.HasPrefix(slug, prefix) {
				continue
			}
			seen[slug] = true
			out = append(out, slug)
		}
	}
	sort.Strings(out)
	return out
}

// taskCompleter completes task slugs whose status (== their directory) is not
// `exclude`, so `task start` won't offer already-in-progress tasks. An empty
// exclude offers every task.
func (a *App) taskCompleter(exclude domain.Status) completeFunc {
	return func(_ *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		root, ok := a.planningRoot()
		if !ok {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		var pats []string
		for _, st := range domain.AllStatuses() {
			if st == exclude {
				continue
			}
			pats = append(pats, filepath.Join(root, domain.TasksDir, st.Dir(), "*.md"))
		}
		return slugsFromGlobs(pats, toComplete, args), cobra.ShellCompDirectiveNoFileComp
	}
}

// auditCompleter completes audit slugs whose bucket (== their directory) is not
// `exclude`, so `audit reopen` won't offer already-open audits. An empty
// exclude offers every audit.
func (a *App) auditCompleter(exclude domain.AuditBucket) completeFunc {
	return func(_ *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		root, ok := a.planningRoot()
		if !ok {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		var pats []string
		for _, b := range domain.AllAuditBuckets() {
			if b == exclude {
				continue
			}
			pats = append(pats, filepath.Join(root, domain.AuditsDir, b.Dir(), "*.md"))
		}
		return slugsFromGlobs(pats, toComplete, args), cobra.ShellCompDirectiveNoFileComp
	}
}

// completeTaskSlugs completes every task slug (any status). Used by show/set/
// move, where the current status doesn't constrain the choice.
func (a *App) completeTaskSlugs(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return a.taskCompleter("")(cmd, args, toComplete)
}

// completeAuditSlugs completes every audit slug (any bucket). Used by audit show.
func (a *App) completeAuditSlugs(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return a.auditCompleter("")(cmd, args, toComplete)
}

// completeEpicIDs completes epic ids (epics live flat in epics/).
func (a *App) completeEpicIDs(_ *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	root, ok := a.planningRoot()
	if !ok {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	ids := slugsFromGlobs([]string{filepath.Join(root, domain.EpicsDir, "*.md")}, toComplete, args)
	return ids, cobra.ShellCompDirectiveNoFileComp
}
