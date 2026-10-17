package cli

import (
	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/wire"
)

// planningRepairCommandFailure adds invocation context to the core completed
// repair prefix or partial relocation, including returned stdout errors. Process
// termination (for example SIGPIPE on stdout) can prevent stderr reporting.
type planningRepairCommandFailure struct {
	cause     error
	result    core.PlanningRepairResult
	dryRun    bool
	workspace wire.WorkspaceJSON
}

func (e *planningRepairCommandFailure) Error() string { return e.cause.Error() }
func (e *planningRepairCommandFailure) Unwrap() error { return e.cause }
