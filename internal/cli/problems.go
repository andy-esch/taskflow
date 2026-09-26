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

// problemsError returns a validation error (non-zero exit) when any per-file
// load problems exist, else nil. It does not print: human commands render the
// problems to stderr themselves, and JSON commands carry them in the payload.
//
// The command still exits non-zero (11) on a single bad file, deliberately: the listing
// above it is best-effort and complete, but a caller that got a partial result must be able
// to tell. This mirrors `status --all`, which renders every available space and then exits
// non-zero when one could not be read.
func problemsError(problems []domain.FileProblem) error {
	if len(problems) == 0 {
		return nil
	}
	names := make([]string, 0, problemNamesInError)
	for _, p := range problems {
		if len(names) == problemNamesInError {
			break
		}
		names = append(names, filepath.Base(p.Path))
	}
	listed := strings.Join(names, ", ")
	if extra := len(problems) - len(names); extra > 0 {
		listed += fmt.Sprintf(", +%d more", extra)
	}
	return fmt.Errorf("%w: %d file(s) with unreadable frontmatter: %s",
		domain.ErrValidation, len(problems), listed)
}

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
func portableProblemsError(kind string, problems []core.LintLoadProblem) error {
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

func portableProblemName(kind string, problem core.LintLoadProblem) string {
	name := problem.EntitySlug
	if name == "" {
		name = problem.EntityID
	}
	path := problem.Path
	if path == "" && problem.LocationIsPath {
		path = problem.Location
	}
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
