package store

import (
	"fmt"

	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
)

// ReadLintTasks adapts the body-carrying local task scan to lint's portable
// failed-record contract without adding another filesystem pass.
func (s *FS) ReadLintTasks() ([]core.LoadedRecord[core.TaskWithBody], []core.LoadProblem, error) {
	records, sourceProblems, err := s.scanTaskDocuments()
	loaded := make([]core.LoadedRecord[core.TaskWithBody], 0, len(records))
	for _, record := range records {
		loaded = append(loaded, record.record)
	}
	return loaded, loadedProblems(core.EntityTask, fileProblems(sourceProblems)), err
}

func (s *FS) ReadLintEpics() ([]core.LoadedRecord[domain.Epic], []core.LoadProblem, error) {
	records, problems, err := s.scanEpics()
	return records, loadedProblems(core.EntityEpic, problems), err
}

func (s *FS) ReadAuditSnapshot(selector string) (core.AuditSnapshot, error) {
	if err := s.rejectRepositoryPlannerCall(); err != nil {
		return core.AuditSnapshot{}, err
	}
	if selector != "" {
		path, err := s.resolveAudit(selector)
		if err != nil {
			return core.AuditSnapshot{}, err
		}
		content, err := s.auditReadFile(path)
		if err != nil {
			return core.AuditSnapshot{}, fmt.Errorf("read audit %s: %w", path, err)
		}
		a, findings, nearMisses, candidateIssues, err := parseAuditWithFindings(content, path)
		if err != nil {
			return core.AuditSnapshot{}, fmt.Errorf("%s: %w", path, err)
		}
		record := core.AuditWithFindings{
			Audit: a, Findings: findings, NearMisses: nearMisses, CandidateIssues: candidateIssues,
		}
		return core.AuditSnapshot{Audits: []core.LoadedRecord[core.AuditWithFindings]{{
			Value: record, Source: auditSource(path),
		}}}, nil
	}
	loaded, problems, err := s.scanAuditsWithFindings()
	return core.AuditSnapshot{
		Audits: loaded, Problems: loadedProblems(core.EntityAudit, problems),
	}, err
}

func (s *FS) ReadLintResearch() ([]core.LoadedRecord[domain.Research], []core.LoadProblem, error) {
	records, problems, err := s.scanResearch()
	return records, loadedProblems(core.EntityResearch, problems), err
}
