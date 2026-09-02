package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/andy-esch/taskflow/internal/domain"
)

// Every noun in this CLI takes its verb second — `task list`, `audit show`. Typing
// the verb alone is the natural slip, and cobra answers it with edit distance:
// `tskflwctl list` suggested `lint`, and nothing else. That is a validator being
// offered in place of a read, with a very different effect on a repository the
// caller may not own — and an agent re-derives the same wrong choice every session,
// because nothing about the suggestion says it is a poor one.
//
// These hidden commands intercept the bare verbs and name the noun-qualified forms
// instead. Hidden keeps them out of help and completion; they exist only to answer a
// question cobra was answering badly. Real typos (`boadr`) still reach cobra's
// distance matching untouched. Same shape as deprecatedTransitionCmd, which redirects
// verbs that moved rather than verbs that were never top-level.
var bareVerbRedirects = map[string][]string{
	"list":     {"board", "task list", "epic list", "audit list", "research list"},
	"show":     {"task show <task>", "epic show <epic>", "audit show <audit>"},
	"new":      {"task new <title>", "epic new <title>", "audit new <area>", "research new <title>"},
	"start":    {"task start <task>"},
	"complete": {"task complete <task>"},
	"edit":     {"task edit <task>", "audit edit <audit>"},
}

// newBareVerbCmds returns one hidden command per bare verb above.
func newBareVerbCmds(app *App) []*cobra.Command {
	cmds := make([]*cobra.Command, 0, len(bareVerbRedirects))
	for verb, forms := range bareVerbRedirects {
		cmds = append(cmds, newBareVerbCmd(app, verb, forms))
	}
	return cmds
}

func newBareVerbCmd(app *App, verb string, forms []string) *cobra.Command {
	return &cobra.Command{
		Use:                verb,
		Hidden:             true,
		DisableFlagParsing: true, // the flags belong to the real command; never interpret them here
		Args:               cobra.ArbitraryArgs,
		// This is a usage error, so it must not depend on repo discovery: outside a
		// planning tree the caller would otherwise be told to run `init`, which is
		// not what they got wrong. Same opt-out `schema` uses.
		PersistentPreRunE: app.styleOnlyPreRun,
		RunE: func(_ *cobra.Command, _ []string) error {
			qualified := make([]string, len(forms))
			for i, f := range forms {
				qualified[i] = "tskflwctl " + f
			}
			return fmt.Errorf("%w: %q needs a noun — every verb here takes one. Try: %s",
				domain.ErrValidation, verb, strings.Join(qualified, " · "))
		},
	}
}
