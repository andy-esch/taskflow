// Package spacehealth diagnoses the filesystem health of home-registry entries. The
// spacestore adapter translates this storage/discovery vocabulary into core values for
// application consumers.
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
	// Dir is the directory discovery actually resolved this entry to — the same
	// symlink-evaluated value a runtime workspace carries as its checkout. The registry
	// spelling stays on Space.Path; consumers that COMPARE an entry against an open
	// workspace must use this, or a symlinked (or differently-cased) registry row never
	// matches the tree the user is standing in. Empty when discovery failed, or when it
	// fell back to a bare tasks/ dir with no config anchoring it.
	Dir     string
	Message string
	Remedy  string
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
	p.Dir = cfg.Dir
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
		domain.ThreadsDir,
		// Legacy Projects content remains part of the planning space until the
		// operator explicitly migrates it; never diagnose that repository empty.
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
