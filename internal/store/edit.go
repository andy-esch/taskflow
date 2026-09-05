package store

import (
	"errors"
	"fmt"
	"os"
	"slices"
	"sort"
	"time"

	yaml "go.yaml.in/yaml/v3"

	"github.com/andy-esch/taskflow/internal/domain"
)

// editFile is the shared editor-loop behind EditTask / EditEpic / EditAudit: it
// hands the current file content to edit — which runs the caller's $EDITOR and
// returns the new content — and accepts the result only if it still parses
// (parse-before-accept), so a frontmatter break never lands on disk. On an invalid
// edit it reopens the editor on the broken content (passing the parse error) and
// loops until: the edit is valid (atomic write), the user re-saves the same broken
// content unchanged (gives up → ErrValidation), or the content matches the original
// (no write — but a pre-existing on-disk break is still surfaced, never a phantom
// success). Just before the write it calls recheck, the caller's compare-and-swap
// against a concurrent source change during the (long) editor window; a nil
// recheck skips that guard. Returns the reloaded entity and whether it changed.
//
// An accepted change has its updated_at stamped to `now` (surgically, preserving the
// rest of the user's edit), so any editor edit advances the activity date the way the
// set/append/move paths do — uniform across task/epic/audit, which all carry the
// field. An unchanged save stamps nothing (no write).
//
// The fs and the editor stay decoupled: the store orchestrates resolve/parse/write;
// the caller's edit callback owns the editor (a cli human-face concern).
// acceptEdited wraps a parse function for the EDITOR path so an edit that introduces an
// illegal entity id is rejected the way a frontmatter break is — the editor reopens with
// the reason instead of the id reaching disk.
//
// Deliberately NOT folded into parseTask/parseAudit/parseResearch, which the reviewer
// suggested: reads must stay tolerant so `lint` can REPORT a bad id already on disk, while
// writes refuse to create one. Tightening the parser would turn every such file from
// "flagged and fixable" into "unreadable". Same asymmetry the filename-id guard already
// has.
func acceptEdited[T any](parse func([]byte) (T, error), entityID func(T) string) func([]byte) (T, error) {
	return func(content []byte) (T, error) {
		v, err := parse(content)
		if err != nil {
			return v, err
		}
		if got := entityID(v); got != "" {
			if err := validEntityID(got); err != nil {
				return v, err
			}
		}
		return v, nil
	}
}

// asValidation wraps a rejected edit as a validation failure without stuttering: the id
// guard already classifies its own error, and wrapping unconditionally produced
// "validation failed: validation failed: …".
func asValidation(err error) error {
	if errors.Is(err, domain.ErrValidation) {
		return err
	}
	return fmt.Errorf("%w: %v", domain.ErrValidation, err)
}

func editFile[T any](
	noun, path string,
	orig []byte,
	now time.Time,
	parse func(content []byte) (T, error),
	lock func() (func(), error),
	recheck func() error,
	edit func(current string, prevErr error) (string, error),
) (T, bool, error) {
	var zero T
	current := string(orig)
	var prevErr error
	for {
		edited, err := edit(current, prevErr)
		if err != nil {
			return zero, false, err
		}
		if edited == string(orig) {
			// No net change. Surface a parse error if the file was already broken on
			// disk (opened to inspect, saved unchanged) rather than report a phantom
			// success with an empty entity.
			v, perr := parse(orig)
			if perr != nil {
				return zero, false, asValidation(perr)
			}
			return v, false, nil
		}
		// Parse-before-accept on the user's own content: a frontmatter break reopens
		// the editor (or, re-saved unchanged, is a give-up) rather than landing.
		if _, perr := parse([]byte(edited)); perr != nil {
			if edited == current {
				// re-saved the same broken content → the user gave up
				return zero, false, asValidation(perr)
			}
			current, prevErr = edited, perr // reopen on the broken content
			continue
		}
		// Accepted. Stamp updated_at so any edit advances the activity date (uniform
		// with set/append/move). The frontmatter just parsed, so the surgical stamp
		// can't hit a structural break; re-parse the stamped form for the return value.
		stamped, err := updateFrontmatter([]byte(edited), map[string]any{"updated_at": now.Format("2006-01-02")})
		if err != nil {
			return zero, false, err
		}
		v, perr := parse(stamped)
		if perr != nil {
			return zero, false, asValidation(perr)
		}
		// Take the write lock only NOW — after the (long, interactive) editor returned and
		// the edit parsed — so the flock covers just the brief verify→write, never the whole
		// editor session. It makes the CAS atomic against cooperating writers.
		unlock, lockErr := lock()
		if lockErr != nil {
			return zero, false, lockErr
		}
		defer unlock()
		// Compare-and-swap before the write: the editor window is long, so any
		// concurrent lifecycle, rename, or content write must defeat this stale save.
		// Atomic replacement guards torn writes; this recheck guards lost updates.
		if recheck != nil {
			if err := recheck(); err != nil {
				return zero, false, err
			}
		}
		if err := writeFileAtomic(path, stamped, 0o644); err != nil {
			return zero, false, fmt.Errorf("write %s %s: %w", noun, path, err)
		}
		return v, true, nil
	}
}

// EditTask resolves slug, reads the file, and runs the shared editor-loop
// (parse-before-accept), accepting a save only if it still parses as a task. The
// recheck is a compare-and-swap across the editor window, rejecting a concurrent
// lifecycle transition, rename, or content edit. Returns the reloaded task and
// whether it changed.
func (s *FS) EditTask(slug string, now time.Time, edit func(current string, prevErr error) (string, error)) (domain.Task, bool, error) {
	if err := s.rejectRepositoryPlannerCall(); err != nil {
		return domain.Task{}, false, err
	}
	path, err := s.resolve(slug)
	if err != nil {
		return domain.Task{}, false, err
	}
	orig, err := os.ReadFile(path)
	if err != nil {
		return domain.Task{}, false, fmt.Errorf("read task %s: %w", path, err)
	}
	ifVersion := hashContent(orig)
	originalDependencies, dependenciesReadable := dependencyValues(orig)
	originalStatus, statusReadable := taskStatusValue(orig)
	parseAcceptedTask := func(content []byte) (domain.Task, error) {
		t, err := acceptEdited(
			func(content []byte) (domain.Task, error) { return parseTask(content, path) },
			func(t domain.Task) string { return t.ID })(content)
		if err != nil {
			return t, err
		}
		candidate := dependencyFieldsFromTask(t)
		if !dependenciesReadable {
			// Parser failure is not an empty dependency set. Reject every candidate so
			// deleting the malformed field cannot sneak through as a "repair".
			return t, fmt.Errorf("%w: cannot verify the original graph-owned fields while repairing malformed frontmatter; repair them directly, run lint, then use guarded dependency operations", domain.ErrValidation)
		} else if !candidate.equal(originalDependencies) {
			return t, fmt.Errorf("%w: task edit cannot change depends_on or legacy dependency fields; use guarded dependency operations", domain.ErrValidation)
		}
		if !statusReadable {
			return t, fmt.Errorf("%w: cannot verify the original lifecycle status while repairing malformed frontmatter; repair it directly, run lint, then use a lifecycle verb", domain.ErrValidation)
		}
		if string(t.Status) != originalStatus {
			return t, fmt.Errorf("%w: task edit cannot change status from %s to %s; save the content edit without that change, then use `task start`, `task complete`, `task defer`, or `task move`", domain.ErrValidation, originalStatus, t.Status)
		}
		return t, nil
	}
	return editFile("task", path, orig, now,
		parseAcceptedTask,
		s.writeLock,
		// Version-CAS across the long editor window: conflict if the path or content
		// changed under us.
		func() error { return verifyUnchanged(s.resolvePath, slug, path, ifVersion, "task", "edit") },
		edit)
}

func taskStatusValue(content []byte) (string, bool) {
	fm, _, err := splitFrontmatterStrict(content)
	if err != nil || fm == nil {
		return "", false
	}
	var fields struct {
		Status string `yaml:"status"`
	}
	if err := yaml.Unmarshal(fm, &fields); err != nil || fields.Status == "" {
		return "", false
	}
	return fields.Status, true
}

type taskDependencyFields struct {
	dependsOn           []string
	blockedBy           []string
	dependencies        []string
	blocks              []string
	blockedByPresent    bool
	dependenciesPresent bool
	blocksPresent       bool
}

func sortedCopy(values []string) []string {
	out := append([]string(nil), values...)
	sort.Strings(out)
	return out
}

func dependencyFieldsFromTask(task domain.Task) taskDependencyFields {
	return taskDependencyFields{
		dependsOn:           sortedCopy(task.DependsOn),
		blockedBy:           sortedCopy(task.LegacyBlockedBy),
		dependencies:        sortedCopy(task.LegacyDependencies),
		blocks:              sortedCopy(task.LegacyBlocks),
		blockedByPresent:    slices.Contains(task.LegacyDependencyFields, "blocked_by") || len(task.LegacyBlockedBy) > 0,
		dependenciesPresent: slices.Contains(task.LegacyDependencyFields, "dependencies") || len(task.LegacyDependencies) > 0,
		blocksPresent:       slices.Contains(task.LegacyDependencyFields, "blocks") || len(task.LegacyBlocks) > 0,
	}
}

func (fields taskDependencyFields) equal(other taskDependencyFields) bool {
	return slices.Equal(fields.dependsOn, other.dependsOn) &&
		slices.Equal(fields.blockedBy, other.blockedBy) &&
		slices.Equal(fields.dependencies, other.dependencies) &&
		slices.Equal(fields.blocks, other.blocks) &&
		fields.blockedByPresent == other.blockedByPresent &&
		fields.dependenciesPresent == other.dependenciesPresent &&
		fields.blocksPresent == other.blocksPresent
}

// dependencyValues extracts every dependency-affecting field from an original
// task. A narrow decode can recover the graph baseline even when an unrelated
// typed field (for example tier) is malformed, preserving task edit's existing
// ability to repair such files without opening a dependency bypass. False means
// the YAML itself is not trustworthy.
func dependencyValues(content []byte) (taskDependencyFields, bool) {
	fm, _, err := splitFrontmatterStrict(content)
	if err != nil || fm == nil {
		return taskDependencyFields{}, false
	}
	var fields struct {
		DependsOn    []string `yaml:"depends_on"`
		BlockedBy    []string `yaml:"blocked_by"`
		Dependencies []string `yaml:"dependencies"`
		Blocks       []string `yaml:"blocks"`
	}
	if err := yaml.Unmarshal(fm, &fields); err != nil {
		return taskDependencyFields{}, false
	}
	var present map[string]yaml.Node
	if err := yaml.Unmarshal(fm, &present); err != nil {
		return taskDependencyFields{}, false
	}
	return taskDependencyFields{
		dependsOn:           sortedCopy(fields.DependsOn),
		blockedBy:           sortedCopy(fields.BlockedBy),
		dependencies:        sortedCopy(fields.Dependencies),
		blocks:              sortedCopy(fields.Blocks),
		blockedByPresent:    present["blocked_by"].Kind != 0,
		dependenciesPresent: present["dependencies"].Kind != 0,
		blocksPresent:       present["blocks"].Kind != 0,
	}, true
}

// EditAudit is the audit twin of EditTask: same parse-before-accept editor-loop,
// with the compare-and-swap guarding against a concurrent audit lifecycle or
// content change during the editor window. Finding-level lint (status vocab,
// bucket↔state) is left to the caller, mirroring how task edit leaves field lint
// to `lint` — the store only guarantees the file still parses.
func (s *FS) EditAudit(slug string, now time.Time, edit func(current string, prevErr error) (string, error)) (domain.Audit, bool, error) {
	if err := s.rejectRepositoryPlannerCall(); err != nil {
		return domain.Audit{}, false, err
	}
	path, err := s.resolveAudit(slug)
	if err != nil {
		return domain.Audit{}, false, err
	}
	orig, err := os.ReadFile(path)
	if err != nil {
		return domain.Audit{}, false, fmt.Errorf("read audit %s: %w", path, err)
	}
	ifVersion := hashContent(orig)
	_, origBody, _ := splitFrontmatterStrict(orig)
	return editFile("audit", path, orig, now,
		acceptEdited(
			func(content []byte) (domain.Audit, error) {
				a, err := parseAudit(content, path)
				if err != nil {
					return a, err
				}
				// The editor reopens on a rejected edit, so refusing here puts the exact
				// canonical replacement in front of the author while they still have the
				// text open — the same reason the unterminated-fence guard lives at the
				// write. Pre-existing drift elsewhere in the file is not this edit's fault.
				_, newBody, splitErr := splitFrontmatterStrict(content)
				if splitErr != nil {
					return a, splitErr
				}
				return a, domain.NearMissWriteError("audit",
					domain.IntroducedNearMissHeaders(string(origBody), string(newBody)))
			},
			func(a domain.Audit) string { return a.ID }),
		s.writeLock,
		func() error { return verifyUnchanged(s.resolveAuditPath, slug, path, ifVersion, "audit", "edit") },
		edit)
}
