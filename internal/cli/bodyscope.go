package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/andy-esch/taskflow/internal/domain"
)

// addBodyScopeFlags wires the shared read-narrowing flags onto a `show` command:
// --section (one named body section) and --frontmatter-only (drop the body). They
// are mutually exclusive. Task/epic/audit show all use these so the surface can't
// drift between entities.
func addBodyScopeFlags(cmd *cobra.Command, section *string, fmOnly *bool) {
	cmd.Flags().StringVar(section, "section", "",
		"show only the body section whose heading matches this name (e.g. acceptance, progress)")
	cmd.Flags().BoolVar(fmOnly, "frontmatter-only", false,
		"show only the metadata, skipping the body")
	cmd.MarkFlagsMutuallyExclusive("section", "frontmatter-only")
}

// narrowBody applies --section / --frontmatter-only to a body: "" for
// frontmatter-only, the named section slice for --section (ErrNotFound when no
// heading matches), else the body unchanged. These flags narrow only the free-text
// markdown body — the metadata block (and, for epics/audits, the roster/finding
// tree) is always shown. kind/id shape the not-found message.
func narrowBody(kind, id, body, section string, fmOnly bool) (string, error) {
	switch {
	case fmOnly:
		return "", nil
	case section != "":
		sec, ok := domain.Section(body, section)
		if !ok {
			return "", fmt.Errorf("%w: %s %q has no section matching %q", domain.ErrNotFound, kind, id, section)
		}
		return sec, nil
	}
	return body, nil
}
