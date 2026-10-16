package core

import (
	"fmt"

	"github.com/andy-esch/taskflow/internal/domain"
)

// TaskPathByID resolves only an already-selected source identity. It never
// restarts human-reference matching and remains optional and parse-free.
func (s *Service) TaskPathByID(id string) (string, error) {
	if err := requireSourceID(EntityTask, RecordSource{ID: id}); err != nil {
		return "", err
	}
	if s.taskPaths == nil {
		return "", fmt.Errorf("%w: task path resolution is unavailable from this service", domain.ErrValidation)
	}
	path, err := s.taskPaths.ResolveTaskPathByID(id)
	return requireResolvedLocalPath(EntityTask, path, err)
}

// EpicPathByID resolves only an already-selected source identity. It never
// restarts human-reference matching and remains optional and parse-free.
func (s *Service) EpicPathByID(id string) (string, error) {
	if err := requireSourceID(EntityEpic, RecordSource{ID: id}); err != nil {
		return "", err
	}
	if s.epicPaths == nil {
		return "", fmt.Errorf("%w: epic path resolution is unavailable from this service", domain.ErrValidation)
	}
	path, err := s.epicPaths.ResolveEpicPathByID(id)
	return requireResolvedLocalPath(EntityEpic, path, err)
}

// AuditPathByID resolves only an already-selected source identity. It never
// restarts human-reference matching and remains optional and parse-free.
func (s *Service) AuditPathByID(id string) (string, error) {
	if err := requireSourceID(EntityAudit, RecordSource{ID: id}); err != nil {
		return "", err
	}
	if s.auditPaths == nil {
		return "", fmt.Errorf("%w: audit path resolution is unavailable from this service", domain.ErrValidation)
	}
	path, err := s.auditPaths.ResolveAuditPathByID(id)
	return requireResolvedLocalPath(EntityAudit, path, err)
}

// ResearchPathByID resolves only an already-selected source identity. It never
// restarts human-reference matching and remains optional and parse-free.
func (s *Service) ResearchPathByID(id string) (string, error) {
	if err := requireSourceID(EntityResearch, RecordSource{ID: id}); err != nil {
		return "", err
	}
	if s.researchPaths == nil {
		return "", fmt.Errorf("%w: research path resolution is unavailable from this service", domain.ErrValidation)
	}
	path, err := s.researchPaths.ResolveResearchPathByID(id)
	return requireResolvedLocalPath(EntityResearch, path, err)
}

// ThreadPathByID resolves only an already-selected source identity. It never
// restarts human-reference matching and remains optional and parse-free.
func (s *Service) ThreadPathByID(id string) (string, error) {
	if err := requireSourceID(EntityThread, RecordSource{ID: id}); err != nil {
		return "", err
	}
	if s.threadPaths == nil {
		return "", fmt.Errorf("%w: thread path resolution is unavailable from this service", domain.ErrValidation)
	}
	path, err := s.threadPaths.ResolveThreadPathByID(id)
	return requireResolvedLocalPath(EntityThread, path, err)
}
