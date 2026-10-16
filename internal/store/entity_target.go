package store

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
)

// Source selection is parse-free so ordinary field writes can still repair typed
// frontmatter. Declarations and Location never substitute for filename ownership.
func (s *FS) ResolveAuditSource(ref string) (core.RecordSource, error) {
	path, err := s.ResolveAuditPath(ref)
	if err != nil {
		return core.RecordSource{}, err
	}
	return auditSource(path), nil
}

func (s *FS) ResolveResearchSource(ref string) (core.RecordSource, error) {
	path, err := s.ResolveResearchPath(ref)
	if err != nil {
		return core.RecordSource{}, err
	}
	return researchSource(path), nil
}

func (s *FS) ResolveEpicSource(ref string) (core.RecordSource, error) {
	path, err := s.ResolveEpicPath(ref)
	if err != nil {
		return core.RecordSource{}, err
	}
	return core.RecordSource{ID: strings.TrimSuffix(filepath.Base(path), ".md"), Location: path, LocationIsPath: true}, nil
}

func (s *FS) resolveAuditMutationTarget(target core.AuditMutationTarget) (string, error) {
	if err := validEntityID(target.ID); err != nil {
		return "", err
	}
	return s.resolveAuditPath(target.ID)
}

func (s *FS) resolveResearchMutationTarget(target core.ResearchMutationTarget) (string, error) {
	if err := validEntityID(target.ID); err != nil {
		return "", err
	}
	return s.resolveResearchPathExact(target.ID)
}

func (s *FS) resolveEpicMutationPath(id string) (string, error) {
	if strings.TrimSpace(id) == "" {
		return "", fmt.Errorf("%w: selected epic identity is required", domain.ErrValidation)
	}
	candidates, err := epicCandidates(s.epicsDir)
	if err != nil {
		return "", err
	}
	selected, err := resolveExactID(candidates, id)
	if err != nil {
		return "", err
	}
	return selected.path, nil
}
