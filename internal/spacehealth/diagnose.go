// Package spacehealth owns the read-only health projection of the home space registry.
// Keeping diagnosis here gives the CLI doctor/list commands and a future TUI one answer
// for the same path instead of letting each adapter reinterpret discovery failures.
package spacehealth

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/andy-esch/taskflow/internal/config"
	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/userconfig"
)

// Kind is the stable diagnosis vocabulary for one registered space.
type Kind string

const (
	KindOK         Kind = "ok"
	KindEmpty      Kind = "empty"
	KindMissing    Kind = "missing"
	KindNotARepo   Kind = "not-a-repo"
	KindUnreadable Kind = "unreadable"
	KindMismatch   Kind = "mismatch"
)

// Role describes how a registered path reaches its planning tree. It is derived from
// repo config, never persisted in spaces.toml: a direct entry owns the planning tree,
// while a pointer entry reaches one through planning_repo. Unknown is honest for an
// entry whose missing or unreadable config cannot be inspected.
type Role string

const (
	RoleDirect  Role = "direct"
	RolePointer Role = "pointer"
	RoleUnknown Role = "unknown"
)

// SpaceProblem is the complete diagnosis of one registry entry. Despite the name, OK and
// empty entries are included so every consumer receives one typed record per space.
// Message explains the state; Remedy is populated only when a person can repair it.
type SpaceProblem struct {
	Space userconfig.Space
	Kind  Kind
	Role  Role
	// PlanningID is the durable identity this entry belongs to. The registry's
	// verify_id wins because it is the intended identity even when the path has drifted;
	// a successfully discovered id supplies it for legacy registry entries.
	PlanningID string
	Root       string
	Message    string
	Remedy     string
}

// SpaceGroup is one logical planning space and all registered paths that enter it.
// Entries retain registry order. PlanningID is empty only for legacy id-less trees and
// broken entries that have no retained identity assertion.
type SpaceGroup struct {
	PlanningID string
	Entries    []SpaceProblem
}

// Broken reports whether this diagnosis should make doctor exit non-zero. An empty
// planning repo is healthy and addressable; it is information, not a failure.
func (p SpaceProblem) Broken() bool {
	return p.Kind != KindOK && p.Kind != KindEmpty
}

// DiagnoseRegistry reads and diagnoses the home registry. Entry failures are returned as
// data and never stop the sweep; err is reserved for a registry file that cannot itself be
// read or decoded. The function is read-only and never forgets or repairs anything.
func DiagnoseRegistry() ([]SpaceProblem, error) {
	spaces, err := userconfig.Spaces()
	if err != nil {
		return nil, err
	}
	out := make([]SpaceProblem, 0, len(spaces))
	for _, space := range spaces {
		out = append(out, DiagnoseSpace(space))
	}
	return out, nil
}

// Group projects registry diagnoses into logical planning spaces. A durable planning id
// is the preferred key. Legacy healthy entries fall back to their physical resolved root;
// broken id-less entries remain isolated so unrelated failures can never collapse into one
// apparent space. Both group order and entry order follow registry insertion order.
func Group(problems []SpaceProblem) []SpaceGroup {
	groups := make([]SpaceGroup, 0, len(problems))
	byKey := make(map[string]int, len(problems))
	for i, problem := range problems {
		key := groupKey(problem, i)
		groupIndex, exists := byKey[key]
		if !exists {
			groupIndex = len(groups)
			byKey[key] = groupIndex
			groups = append(groups, SpaceGroup{PlanningID: problem.PlanningID})
		}
		groups[groupIndex].Entries = append(groups[groupIndex].Entries, problem)
	}
	return groups
}

// DiagnoseSpace diagnoses one already-loaded entry. It is useful for mutation receipts
// and explicit --space selection, which need the shared rules without re-reading the
// registry or requiring that a dry-run preview already exist there.
func DiagnoseSpace(space userconfig.Space) SpaceProblem {
	p := SpaceProblem{Space: space, Role: RoleUnknown, PlanningID: space.VerifyID}
	dir := userconfig.ExpandTilde(space.Path)
	info, err := os.Stat(dir)
	if err != nil {
		if os.IsNotExist(err) {
			p.Kind = KindMissing
			p.Message = "not found at " + space.Path
			p.Remedy = forgetAndReadd(space.ID)
			return p
		}
		p.Kind = KindUnreadable
		p.Message = firstLine(fmt.Sprintf("cannot inspect %s: %v", space.Path, err))
		p.Remedy = "check the path permissions, or `space forget " + space.ID + "`"
		return p
	}
	if !info.IsDir() {
		p.Kind = KindNotARepo
		p.Message = "path is not a directory"
		p.Remedy = forgetAndReadd(space.ID)
		return p
	}
	// Discover's existence helpers intentionally collapse stat failures to false. Probe the
	// registered directory directly first so a permission failure is not mislabeled as an
	// ordinary uninitialized directory.
	if _, err := os.ReadDir(dir); err != nil {
		p.Kind = KindUnreadable
		p.Message = firstLine(fmt.Sprintf("cannot read %s: %v", space.Path, err))
		p.Remedy = "check the path permissions, or `space forget " + space.ID + "`"
		return p
	}

	cfg, err := config.Discover(dir)
	if err != nil {
		p.Kind = KindNotARepo
		p.Message = "no planning repo here"
		p.Remedy = "run `tskflwctl init` there, or `space forget " + space.ID + "`"
		if errorsIsConfigFailure(err) {
			p.Kind = KindUnreadable
			p.Message = firstLine(err.Error())
			p.Remedy = "repair the repo config, or `space forget " + space.ID + "`"
		}
		return p
	}
	p.Root = cfg.Root
	if cfg.PlanningRepo == "" {
		p.Role = RoleDirect
	} else {
		p.Role = RolePointer
	}
	if p.PlanningID == "" {
		p.PlanningID = cfg.ID
	}
	if space.VerifyID != "" && cfg.ID != space.VerifyID {
		p.Kind = KindMismatch
		if cfg.ID == "" {
			p.Message = fmt.Sprintf("resolved repo carries no id; registry expects %q", space.VerifyID)
		} else {
			p.Message = fmt.Sprintf("resolved repo id %q does not match registry verify_id %q", cfg.ID, space.VerifyID)
		}
		p.Remedy = "restore the intended checkout at this path, or " + forgetAndReadd(space.ID)
		return p
	}

	empty, err := planningRootEmpty(cfg.Root)
	if err != nil {
		p.Kind = KindUnreadable
		p.Message = firstLine(err.Error())
		p.Remedy = "check the planning-tree permissions, or `space forget " + space.ID + "`"
		return p
	}
	if empty {
		p.Kind = KindEmpty
		p.Message = "no planning entities yet"
		return p
	}
	p.Kind = KindOK
	return p
}

func groupKey(problem SpaceProblem, index int) string {
	if problem.PlanningID != "" {
		return "id:" + problem.PlanningID
	}
	if problem.Root != "" {
		return "root:" + problem.Root
	}
	// Include the index as well as the local label: Group also accepts synthetic slices
	// in tests and future adapters, where duplicate labels are possible even though a
	// valid persisted registry rejects them.
	return fmt.Sprintf("entry:%d:%s", index, problem.Space.ID)
}

// errorsIsConfigFailure distinguishes a present-but-invalid planning config/pointer from
// the ordinary discovery miss, whose error intentionally has no domain classification.
func errorsIsConfigFailure(err error) bool {
	class := domain.Classify(err)
	return class == domain.ClassValidation || class == domain.ClassConflict
}

func planningRootEmpty(root string) (bool, error) {
	for _, dir := range []string{
		domain.TasksDir,
		domain.EpicsDir,
		domain.AuditsDir,
		domain.ResearchDir,
		domain.ProjectsDir,
	} {
		path := filepath.Join(root, dir)
		entries, err := os.ReadDir(path)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return false, fmt.Errorf("read %s: %w", path, err)
		}
		for _, entry := range entries {
			if !entry.IsDir() && strings.EqualFold(filepath.Ext(entry.Name()), ".md") {
				return false, nil
			}
		}
	}
	return true, nil
}

func forgetAndReadd(id string) string {
	return "`space forget " + id + "`, then `space add <new-path> --id " + id + "`"
}

func firstLine(s string) string {
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		return s[:i]
	}
	return s
}
