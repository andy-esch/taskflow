package store

import (
	"path/filepath"
	"strings"

	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
)

// ReadLintTasks adapts the body-carrying local task scan to lint's portable
// failed-record contract without adding another filesystem pass.
func (s *FS) ReadLintTasks() ([]core.TaskWithBody, []core.LintLoadProblem, error) {
	records, problems, err := s.ListTasksWithBodies()
	return records, lintLoadProblems(core.LintEntityTask, problems), err
}

func (s *FS) ReadLintEpics() ([]domain.Epic, []core.LintLoadProblem, error) {
	records, problems, err := s.ListEpics()
	return records, lintLoadProblems(core.LintEntityEpic, problems), err
}

func (s *FS) ReadLintAudits() ([]core.AuditWithFindings, []core.LintLoadProblem, error) {
	records, problems, err := s.ListAuditsWithFindings()
	return records, lintLoadProblems(core.LintEntityAudit, problems), err
}

func (s *FS) ReadLintResearch() ([]domain.Research, []core.LintLoadProblem, error) {
	records, problems, err := s.ListResearch()
	return records, lintLoadProblems(core.LintEntityResearch, problems), err
}

func lintLoadProblems(kind core.LintEntityKind, problems []domain.FileProblem) []core.LintLoadProblem {
	out := make([]core.LintLoadProblem, 0, len(problems))
	for _, problem := range problems {
		entityID, entitySlug := problem.EntityID, problem.EntitySlug
		// Epic identity is its whole filename stem rather than a 12-character
		// stable-id prefix, so the generic scanner cannot recover it. This adapter
		// owns the filename convention and performs the conversion here once.
		if kind == core.LintEntityEpic && entityID == "" && problem.Path != "" {
			entityID = strings.TrimSuffix(filepath.Base(problem.Path), ".md")
		}
		out = append(out, core.LintLoadProblem{
			EntityKind: kind, EntityID: entityID, EntitySlug: entitySlug,
			Location: problem.Path, LocationIsPath: problem.Path != "", Message: problem.Message,
		})
	}
	return out
}
