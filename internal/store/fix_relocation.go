package store

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
)

// relocateFixedFrontmatter runs under FixFrontmatter's repository guard. Raw
// editors do not join that guard: bound their windows with content checks and
// exclusive creation, retaining both copies rather than deleting on conflict.
// This is not a crash-atomic move or a guarantee against the final check/unlink race.
func (s *FS) relocateFixedFrontmatter(source, destination string, before, fixed []byte) error {
	if testHookBeforeFixIDDestinationCreate != nil {
		testHookBeforeFixIDDestinationCreate(source, destination)
	}
	if err := verifyFixPath(source, hashContent(before)); err != nil {
		return err
	}
	// Source CAS cannot notice a newly added canonical owner or inbound
	// reference in another file. Recheck the existing spelling-repair policy
	// immediately before creation as well; this is not whole-repository CAS.
	_, target, note, err := s.repairInvalidID(filepath.Dir(source), filepath.Base(source), fixed)
	if err != nil {
		return err
	}
	if target != destination {
		return fmt.Errorf("%w: ID spelling repair authorization changed: %s", domain.ErrConflict, note)
	}
	info, err := os.Stat(source)
	if err != nil {
		return fmt.Errorf("stat repair source %s: %w", source, err)
	}
	if testHookBeforeFixIDExclusiveCreate != nil {
		testHookBeforeFixIDExclusiveCreate(destination)
	}
	if err := createFileAtomicExactMode(destination, fixed, info.Mode().Perm()); err != nil {
		if os.IsExist(err) {
			return fmt.Errorf("%w: repair destination already exists: %s", domain.ErrConflict, destination)
		}
		return fmt.Errorf("create repair destination %s: %w", destination, err)
	}
	fail := func(cause error) error {
		return &core.FrontmatterRepairFailure{Cause: cause, Relocation: core.FrontmatterRepairRelocation{
			SourcePath: source, DestinationPath: destination, DestinationWritten: true,
		}}
	}
	if testHookBeforeFixIDSourceRemove != nil {
		if err := testHookBeforeFixIDSourceRemove(source, destination); err != nil {
			return fail(fmt.Errorf("before removing repair source %s: %w", source, err))
		}
	}
	// The destination must still contain our repaired bytes before the old copy
	// is removed; a raw deletion/replacement of it cannot justify dropping source.
	if err := verifyFixPath(destination, hashContent(fixed)); err != nil {
		return fail(err)
	}
	if err := verifyFixPath(source, hashContent(before)); err != nil {
		return fail(err)
	}
	removeSource := os.Remove
	if testHookFixIDRemoveSource != nil {
		removeSource = testHookFixIDRemoveSource
	}
	if err := removeSource(source); err != nil {
		return fail(fmt.Errorf("remove repair source %s: %w", source, err))
	}
	return nil
}

// Separate the final create window from the earlier authorization/source seam:
// a fresh availability scan must not mask loss of the exclusive-create invariant.
var testHookBeforeFixIDExclusiveCreate func(destination string)

// Inject a syscall failure at the removal itself, after both content checks.
// The earlier pre-remove hook models raw edits, not a failed unlink.
var testHookFixIDRemoveSource func(source string) error

func verifyFixPath(path, version string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("%w: frontmatter repair file %s disappeared", domain.ErrConflict, path)
		}
		return fmt.Errorf("re-read frontmatter repair file %s: %w", path, err)
	}
	if hashContent(content) != version {
		return fmt.Errorf("%w: frontmatter repair file %s changed on disk", domain.ErrConflict, path)
	}
	return nil
}
