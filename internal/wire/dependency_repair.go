package wire

import "github.com/andy-esch/taskflow/internal/core"

// TaskGraphRepairSourceJSON is the exact owner of a graph declaration.
type TaskGraphRepairSourceJSON struct {
	TaskID   string `json:"task_id,omitempty"`
	TaskSlug string `json:"task_slug,omitempty"`
	Location string `json:"location,omitempty"`
}

// TaskGraphRepairEditJSON is a stable, replayable source-declaration selector.
type TaskGraphRepairEditJSON struct {
	Action     string                    `json:"action"`
	Source     TaskGraphRepairSourceJSON `json:"source"`
	Field      string                    `json:"field"`
	Value      string                    `json:"value"`
	Occurrence int                       `json:"occurrence"`
}

func toTaskGraphRepairEditJSON(edit core.TaskGraphSourceEdit) TaskGraphRepairEditJSON {
	return TaskGraphRepairEditJSON{
		Action: string(edit.Action), Source: TaskGraphRepairSourceJSON{
			TaskID: edit.Source.TaskID, TaskSlug: edit.Source.TaskSlug, Location: edit.Source.Location,
		},
		Field: string(edit.Field), Value: edit.Value, Occurrence: edit.Occurrence,
	}
}

// TaskGraphRepairOperationJSON is one selected and reauthorized repair action.
type TaskGraphRepairOperationJSON struct {
	Edit      TaskGraphRepairEditJSON `json:"edit"`
	Reason    string                  `json:"reason"`
	Automatic bool                    `json:"automatic"`
}

// TaskGraphRepairDeclarationJSON records an exact raw declaration removed by
// the prospective plan or durable prefix.
type TaskGraphRepairDeclarationJSON struct {
	Source           TaskGraphRepairSourceJSON `json:"source"`
	Field            string                    `json:"field"`
	Value            string                    `json:"value"`
	Occurrence       int                       `json:"occurrence"`
	HasProjectedEdge bool                      `json:"has_projected_edge"`
	PrerequisiteID   string                    `json:"prerequisite_id,omitempty"`
	DependentID      string                    `json:"dependent_id,omitempty"`
}

// TaskGraphRepairDefectJSON is one repairable declaration defect or residual
// direct-edit problem. Raw values are retained verbatim.
type TaskGraphRepairDefectJSON struct {
	Reason           string                   `json:"reason"`
	Repairable       bool                     `json:"repairable"`
	Automatic        bool                     `json:"automatic"`
	Target           *TaskGraphRepairEditJSON `json:"target,omitempty"`
	HasProjectedEdge bool                     `json:"has_projected_edge"`
	PrerequisiteID   string                   `json:"prerequisite_id,omitempty"`
	DependentID      string                   `json:"dependent_id,omitempty"`
	CandidateIDs     []string                 `json:"candidate_ids"`
	Problem          *GraphProblemJSON        `json:"problem,omitempty"`
}

// TaskGraphRepairJSON is the reusable guarded repair receipt payload.
type TaskGraphRepairJSON struct {
	Changed                  bool                             `json:"changed"`
	DryRun                   bool                             `json:"dry_run"`
	Committed                bool                             `json:"committed"`
	InitialHealth            string                           `json:"initial_health"`
	FinalHealth              string                           `json:"final_health"`
	Selected                 []TaskGraphRepairEditJSON        `json:"selected"`
	Operations               []TaskGraphRepairOperationJSON   `json:"operations"`
	Removed                  []TaskGraphRepairDeclarationJSON `json:"removed"`
	Addressed                []TaskGraphRepairDefectJSON      `json:"addressed"`
	Residual                 []TaskGraphRepairDefectJSON      `json:"residual"`
	Problems                 []GraphProblemJSON               `json:"problems"`
	PlannedFiles             []string                         `json:"planned_files"`
	AppliedFiles             []string                         `json:"applied_files"`
	RemainingFiles           []string                         `json:"remaining_files"`
	Impacts                  []TaskGraphStateImpactJSON       `json:"impacts"`
	ThreadImpacts            []ThreadProjectionImpactJSON     `json:"thread_impacts"`
	IncompleteThreadEvidence []ThreadReadProblemJSON          `json:"incomplete_thread_evidence"`
	Workspace                WorkspaceJSON                    `json:"workspace"`
}

// TaskGraphRepairEnvelope is `task depend repair --json`.
type TaskGraphRepairEnvelope struct {
	SchemaVersion string `json:"schema_version"`
	TaskGraphRepairJSON
}

func ToTaskGraphRepairJSON(receipt core.TaskGraphRepairReceipt, workspace WorkspaceJSON) TaskGraphRepairJSON {
	payload := TaskGraphRepairJSON{
		Changed: receipt.Changed, DryRun: receipt.DryRun, Committed: receipt.Committed,
		InitialHealth: string(receipt.InitialHealth), FinalHealth: string(receipt.FinalHealth),
		Selected:   make([]TaskGraphRepairEditJSON, 0, len(receipt.Selected)),
		Operations: make([]TaskGraphRepairOperationJSON, 0, len(receipt.Operations)),
		Removed:    make([]TaskGraphRepairDeclarationJSON, 0, len(receipt.Removed)),
		Addressed:  toTaskGraphRepairDefectsJSON(receipt.Addressed), Residual: toTaskGraphRepairDefectsJSON(receipt.Residual),
		Problems:     toGraphProblemsJSON(receipt.Problems),
		PlannedFiles: append([]string{}, receipt.PlannedFiles...), AppliedFiles: append([]string{}, receipt.AppliedFiles...),
		RemainingFiles: append([]string{}, receipt.RemainingFiles...),
		Impacts:        toTaskGraphStateImpactsJSON(receipt.Impacts), ThreadImpacts: toThreadProjectionImpactsJSON(receipt.ThreadImpacts),
		IncompleteThreadEvidence: make([]ThreadReadProblemJSON, 0, len(receipt.IncompleteThreads)), Workspace: workspace,
	}
	for _, selection := range receipt.Selected {
		payload.Selected = append(payload.Selected, toTaskGraphRepairEditJSON(selection))
	}
	for _, operation := range receipt.Operations {
		payload.Operations = append(payload.Operations, TaskGraphRepairOperationJSON{
			Edit: toTaskGraphRepairEditJSON(operation.Edit), Reason: string(operation.Reason), Automatic: operation.Automatic,
		})
	}
	for _, declaration := range receipt.Removed {
		payload.Removed = append(payload.Removed, TaskGraphRepairDeclarationJSON{
			Source: toTaskGraphRepairEditJSON(core.TaskGraphSourceEdit{Source: declaration.Source}).Source,
			Field:  string(declaration.Field), Value: declaration.Value, Occurrence: declaration.Occurrence,
			HasProjectedEdge: declaration.HasProjectedEdge,
			PrerequisiteID:   declaration.ProjectedEdge.From, DependentID: declaration.ProjectedEdge.To,
		})
	}
	for _, problem := range receipt.IncompleteThreads {
		payload.IncompleteThreadEvidence = append(payload.IncompleteThreadEvidence, ThreadReadProblemJSON{
			ThreadID: problem.ThreadID, ThreadSlug: problem.ThreadSlug, Location: problem.Location, Message: problem.Message,
		})
	}
	return payload
}

func ToTaskGraphRepairEnvelope(receipt core.TaskGraphRepairReceipt, workspace WorkspaceJSON) TaskGraphRepairEnvelope {
	return TaskGraphRepairEnvelope{SchemaVersion: SchemaVersion, TaskGraphRepairJSON: ToTaskGraphRepairJSON(receipt, workspace)}
}

func toTaskGraphRepairDefectsJSON(defects []core.TaskGraphRepairDefect) []TaskGraphRepairDefectJSON {
	out := make([]TaskGraphRepairDefectJSON, 0, len(defects))
	for _, defect := range defects {
		item := TaskGraphRepairDefectJSON{
			Reason: string(defect.Reason), Repairable: defect.Repairable, Automatic: defect.Automatic,
			HasProjectedEdge: defect.HasProjectedEdge, PrerequisiteID: defect.ProjectedEdge.From,
			DependentID: defect.ProjectedEdge.To, CandidateIDs: append([]string{}, defect.CandidateIDs...),
		}
		if defect.Repairable {
			target := toTaskGraphRepairEditJSON(defect.Target)
			item.Target = &target
		}
		if defect.Problem.Code != "" {
			problem := toGraphProblemsJSON([]core.GraphProblem{defect.Problem})[0]
			item.Problem = &problem
		}
		out = append(out, item)
	}
	return out
}
