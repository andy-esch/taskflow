package store

import (
	"fmt"
	"os"
	"strings"

	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
)

var _ core.CompletionSource = (*FS)(nil)

// ReadCompletionCandidates owns the local naming convention. Plain completion
// reads names only, so malformed frontmatter remains navigable. State-aware
// completion additionally parses each source, leaving failed reads as unknown
// state rather than losing a candidate or excluding a same-slug sibling. Candidate
// enumeration shares the actual resolver's filename rules; references are bare
// canonical IDs, not unresolvable id-slug stems. Alias safety uses the complete
// corpus before state filtering, including case folding and exact-ID precedence.
func (s *FS) ReadCompletionCandidates(kind core.EntityKind, includeState bool) ([]core.CompletionCandidate, error) {
	if err := s.rejectRepositoryPlannerCall(); err != nil {
		return nil, err
	}
	var dir string
	switch kind {
	case core.EntityTask:
		dir = s.tasksDir
	case core.EntityThread:
		dir = s.threadsDir
	case core.EntityEpic:
		dir = s.epicsDir
	case core.EntityAudit:
		dir = s.auditsDir
	case core.EntityResearch:
		dir = s.researchDir
	default:
		return nil, fmt.Errorf("%w: unknown completion entity kind %q", domain.ErrValidation, kind)
	}
	var records []candidate
	var err error
	if kind == core.EntityEpic {
		records, err = epicCandidates(dir)
	} else {
		records, err = flatCandidates(dir)
	}
	if err != nil {
		return nil, err
	}
	ids, aliases := make(map[string]int, len(records)), make(map[string]int, len(records))
	for _, record := range records {
		ids[strings.ToLower(record.id)]++
		if record.slug != "" {
			aliases[strings.ToLower(record.slug)]++
		}
	}
	var out []core.CompletionCandidate
	for _, record := range records {
		// Duplicate canonical IDs cannot honestly publish a canonical selector.
		// Do not let a unique-looking slug hide a broken source identity instead.
		if ids[strings.ToLower(record.id)] != 1 || validQueryName(string(kind), record.id) != nil {
			continue
		}
		slug := record.slug
		aliasSafe := false
		if kind == core.EntityEpic {
			slug, aliasSafe = record.id, true
		} else if validQueryName(string(kind), slug) == nil {
			key := strings.ToLower(slug)
			aliasSafe = aliases[key] == 1 && (ids[key] == 0 || strings.EqualFold(slug, record.id))
		}
		candidate := core.CompletionCandidate{
			ID: record.id, Slug: slug, Reference: record.id, SlugIsReference: aliasSafe,
		}
		if includeState {
			candidate.State = completionState(kind, record.path)
		}
		out = append(out, candidate)
	}
	return out, nil
}

func completionState(kind core.EntityKind, path string) string {
	content, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	switch kind {
	case core.EntityTask:
		value, err := parseTask(content, path)
		if err == nil && !value.StatusFellBack {
			return string(value.Status)
		}
	case core.EntityAudit:
		value, err := parseAudit(content, path)
		if err == nil {
			return string(value.Bucket)
		}
	case core.EntityThread:
		value, err := parseThread(content, path)
		if err == nil {
			return string(value.Status)
		}
	case core.EntityEpic:
		value, err := parseEpic(content, path)
		if err == nil {
			return string(value.Status)
		}
	}
	return ""
}
