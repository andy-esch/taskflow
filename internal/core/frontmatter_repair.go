package core

import "fmt"

// FrontmatterRepairRelocation records adapter-owned local operation evidence,
// not entity identity or permission to open a path. Flags describe this repair's
// durable actions, not current file existence after an out-of-band edit.
// A failing source removal must not turn destination creation into a fixed row.
// See docs/architecture/persistence.md#ordinary-frontmatter-id-repair for recovery
// boundaries and the distinction between action evidence and current existence.
type FrontmatterRepairRelocation struct {
	SourcePath         string
	DestinationPath    string
	DestinationWritten bool
	SourceRemoved      bool
}

// FrontmatterRepairFailure supplements Fixer's completed/proposed prefix with
// progress inside one relocation. The cause retains domain/OS classification;
// callers must inspect the paths rather than replay the whole batch.
type FrontmatterRepairFailure struct {
	Cause      error
	Relocation FrontmatterRepairRelocation
}

func (e *FrontmatterRepairFailure) Error() string {
	return fmt.Sprintf("%v; repair destination %q was written but source removal for %q did not complete; compare and merge both copies without overwriting either, then deliberately remove or archive the copy not kept outside scanned entity directories; lint --fix refuses while both paths exist", e.Cause, e.Relocation.DestinationPath, e.Relocation.SourcePath)
}

func (e *FrontmatterRepairFailure) Unwrap() error { return e.Cause }
