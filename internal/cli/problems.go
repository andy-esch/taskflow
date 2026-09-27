package cli

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
)

// problemNamesInError is how many offending files the summary names before it stops. The
// per-file detail is already on stderr above; this line exists so a reader who scrolled
// past — or an agent that captured only the error — still learns WHICH files, not just how
// many. Beyond a handful the list stops being useful and the detail above is the answer.
const problemNamesInError = 3

func threadProblemsError(problems []core.ThreadReadProblem) error {
	if len(problems) == 0 {
		return nil
	}
	names := make([]string, 0, problemNamesInError)
	for _, problem := range problems {
		if len(names) == problemNamesInError {
			break
		}
		name := problem.ThreadSlug
		if name == "" {
			name = problem.ThreadID
		}
		if name == "" && problem.Location != "" {
			name = filepath.Base(problem.Location)
		}
		if name == "" {
			name = "unidentified Thread record"
		}
		names = append(names, name)
	}
	listed := strings.Join(names, ", ")
	if extra := len(problems) - len(names); extra > 0 {
		listed += fmt.Sprintf(", +%d more", extra)
	}
	return fmt.Errorf("%w: %d unreadable Thread record(s): %s",
		domain.ErrValidation, len(problems), listed)
}

// portableProblemsError reports a partial adapter-neutral read without assuming
// a filesystem path exists. Identity leads; location is only the final fallback.
func portableProblemsError(kind string, problems []core.LoadProblem) error {
	if len(problems) == 0 {
		return nil
	}
	names := make([]string, 0, problemNamesInError)
	for _, problem := range problems {
		if len(names) == problemNamesInError {
			break
		}
		names = append(names, portableProblemName(kind, problem))
	}
	listed := strings.Join(names, ", ")
	if extra := len(problems) - len(names); extra > 0 {
		listed += fmt.Sprintf(", +%d more", extra)
	}
	return fmt.Errorf("%w: %d unreadable %s record(s): %s",
		domain.ErrValidation, len(problems), kind, listed)
}

func portableProblemName(kind string, problem core.LoadProblem) string {
	name := problem.EntitySlug
	if name == "" {
		name = problem.EntityID
	}
	path := problem.LocalPath
	if path != "" {
		base := filepath.Base(path)
		if name == "" {
			name = base
		} else {
			name += " (" + base + ")"
		}
	} else if name == "" && problem.Location != "" {
		name = problem.Location
	}
	if name == "" {
		name = "unidentified " + kind + " record"
	}
	return name
}
