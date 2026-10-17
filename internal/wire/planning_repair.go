package wire

import "github.com/andy-esch/taskflow/internal/core"

// FrontmatterRepairRelocationJSON describes durable actions inside an interrupted
// lint ID spelling move. These paths are recovery evidence, not entity identity
// or filesystem authority. Flags record actions, not current path existence.
type FrontmatterRepairRelocationJSON struct {
	SourcePath         string `json:"source_path" jsonschema:"description=exact local source path whose removal did not complete"`
	DestinationPath    string `json:"destination_path" jsonschema:"description=exact local path where corrected content was created"`
	DestinationWritten bool   `json:"destination_written" jsonschema:"description=true after this repair durably created the destination; not a claim it still exists after raw edits"`
	SourceRemoved      bool   `json:"source_removed" jsonschema:"description=true only if this repair removed the source; false does not certify that another writer retained it"`
}

// ToPlanningRepairEnvelope preserves the ordinary fixed/skipped row vocabulary
// while separately exposing progress inside a failed relocation. The same value
// backs stdout and error recovery so clients do not infer a commit from prose.
func ToPlanningRepairEnvelope(result core.PlanningRepairResult, dryRun bool, ws WorkspaceJSON) FixEnvelope {
	envelope := ToFixEnvelope(result.Fixes, result.Problems, result.LintResults, dryRun, ws)
	if result.Relocation != nil {
		outcome := result.Relocation
		envelope.Relocation = &FrontmatterRepairRelocationJSON{
			SourcePath: outcome.SourcePath, DestinationPath: outcome.DestinationPath,
			DestinationWritten: outcome.DestinationWritten, SourceRemoved: outcome.SourceRemoved,
		}
	}
	return envelope
}
