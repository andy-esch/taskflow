package core

import (
	"fmt"
	"sort"
	"strings"

	"github.com/andy-esch/taskflow/internal/domain"
)

// CompletionCandidate is an adapter-observed resolution candidate, including
// unreadable records. ID is the canonical source identity, not a declaration.
// Reference is an unambiguous selector understood by that adapter; it need not
// be a filename. Slug is a human search label, not automatically a valid selector.
// SlugIsReference explicitly grants use of that label as an unambiguous selector
// under the source's resolver rules (including case folding and ID precedence).
// Sources omit records without an unambiguous reference. State is optional and
// must remain empty when the record cannot be read, keeping damaged records addressable.
type CompletionCandidate struct {
	ID              string
	Slug            string
	Reference       string
	SlugIsReference bool
	State           string
}

// CompletionSource enumerates resolution metadata without requiring semantic
// parsing. includeState permits optional status/bucket reads only for a
// state-aware request. No directory, diagnostic location, or local handle is
// required. Like every planning capability it must publish a SourceSetProvider.
type CompletionSource interface {
	ReadCompletionCandidates(kind EntityKind, includeState bool) ([]CompletionCandidate, error)
}

type CompletionRequest struct {
	Kind         EntityKind
	Prefix       string
	Args         []string
	ExcludeState string
}

// CompleteEntities supplies deterministic, resolution-compatible suggestions.
// Primary adapters decide how to suppress errors for their completion protocol;
// the core never opens a fallback adapter when this capability is unavailable.
func (s *Service) CompleteEntities(request CompletionRequest) ([]string, error) {
	switch request.Kind {
	case EntityTask, EntityThread, EntityEpic, EntityAudit, EntityResearch:
	default:
		return nil, fmt.Errorf("%w: unknown completion entity kind %q", domain.ErrValidation, request.Kind)
	}
	if s.completions == nil {
		return nil, fmt.Errorf("%w: entity completion is unavailable from this service", domain.ErrValidation)
	}
	candidates, err := s.completions.ReadCompletionCandidates(request.Kind, request.ExcludeState != "")
	if err != nil {
		return nil, err
	}
	// Count aliases across the full set, including excluded states: a suggestion
	// must not become ambiguous just because its sibling was omitted from the UI.
	counts := make(map[string]int, len(candidates))
	typed := make(map[string]bool, len(request.Args))
	for _, arg := range request.Args {
		typed[arg] = true
	}
	for _, candidate := range candidates {
		if candidate.ID == "" || candidate.Reference == "" || candidate.Slug == "" {
			continue
		}
		counts[candidate.Slug]++
	}
	// Selecting one record must not suppress siblings sharing its search label.
	// Only a source-certified alias can select a record; an unsafe label is not
	// interpreted as a selector by generic application policy.
	selected := make(map[string]bool, len(request.Args))
	for _, candidate := range candidates {
		if candidate.ID == "" || candidate.Reference == "" || candidate.Slug == "" {
			continue
		}
		if typed[candidate.ID] || typed[candidate.Reference] ||
			(candidate.SlugIsReference && counts[candidate.Slug] == 1 && typed[candidate.Slug]) {
			selected[candidate.ID] = true
		}
	}
	suggested := make(map[string]bool, len(candidates))
	var out []string
	for _, candidate := range candidates {
		if candidate.ID == "" || candidate.Reference == "" || candidate.Slug == "" ||
			(request.ExcludeState != "" && candidate.State == request.ExcludeState) {
			continue
		}
		if selected[candidate.ID] {
			continue
		}
		var suggestion string
		switch {
		case strings.HasPrefix(candidate.Slug, request.Prefix):
			suggestion = candidate.Reference
			if candidate.SlugIsReference && counts[candidate.Slug] == 1 {
				suggestion = candidate.Slug
			}
		case strings.HasPrefix(candidate.Reference, request.Prefix), strings.HasPrefix(candidate.ID, request.Prefix):
			suggestion = candidate.Reference
		}
		if suggestion != "" && !suggested[suggestion] {
			suggested[suggestion] = true
			out = append(out, suggestion)
		}
	}
	sort.Strings(out)
	return out, nil
}
