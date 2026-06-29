package store

import (
	"fmt"
	"os"
	"time"

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
// against a concurrent relocation during the (long) editor window; a nil recheck
// skips that guard (epics never move directories). Returns the reloaded entity and
// whether it changed.
//
// An accepted change has its updated_at stamped to `now` (surgically, preserving the
// rest of the user's edit), so any editor edit advances the activity date the way the
// set/append/move paths do — uniform across task/epic/audit, which all carry the
// field. An unchanged save stamps nothing (no write).
//
// The fs and the editor stay decoupled: the store orchestrates resolve/parse/write;
// the caller's edit callback owns the editor (a cli human-face concern).
func editFile[T any](
	noun, path string,
	orig []byte,
	now time.Time,
	parse func(content []byte) (T, error),
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
				return zero, false, fmt.Errorf("%w: %v", domain.ErrValidation, perr)
			}
			return v, false, nil
		}
		// Parse-before-accept on the user's own content: a frontmatter break reopens
		// the editor (or, re-saved unchanged, is a give-up) rather than landing.
		if _, perr := parse([]byte(edited)); perr != nil {
			if edited == current {
				// re-saved the same broken content → the user gave up
				return zero, false, fmt.Errorf("%w: %v", domain.ErrValidation, perr)
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
			return zero, false, fmt.Errorf("%w: %v", domain.ErrValidation, perr)
		}
		// Compare-and-swap before the write (mirrors SetFields/Move): the editor
		// window is long, so a concurrent move may have relocated the file — writing
		// to the original path would resurrect the slug in its old directory.
		// Atomicity guards torn writes, not lost updates.
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
// recheck is a compare-and-swap against a concurrent `task move` relocating the
// file during the editor window (which would otherwise resurrect the slug in its
// old status directory — a permanent ErrAmbiguous). Returns the reloaded task and
// whether it changed.
func (s *FS) EditTask(slug string, now time.Time, edit func(current string, prevErr error) (string, error)) (domain.Task, bool, error) {
	path, st, err := s.resolve(slug)
	if err != nil {
		return domain.Task{}, false, err
	}
	orig, err := os.ReadFile(path)
	if err != nil {
		return domain.Task{}, false, fmt.Errorf("read task %s: %w", path, err)
	}
	return editFile("task", path, orig, now,
		func(content []byte) (domain.Task, error) { return parseTask(content, path, st) },
		func() error {
			if curPath, _, rerr := s.resolve(slug); rerr != nil || curPath != path {
				return fmt.Errorf("task %q changed on disk during edit; retry: %w", slug, domain.ErrConflict)
			}
			return nil
		},
		edit)
}

// EditAudit is the audit twin of EditTask: same parse-before-accept editor-loop,
// with the compare-and-swap guarding against a concurrent `audit close`/`reopen`/
// `defer` relocating the file across buckets during the editor window. Finding-level
// lint (status vocab, bucket↔state) is left to the caller, mirroring how task edit
// leaves field lint to `lint` — the store only guarantees the file still parses.
func (s *FS) EditAudit(slug string, now time.Time, edit func(current string, prevErr error) (string, error)) (domain.Audit, bool, error) {
	path, bucket, err := s.resolveAudit(slug)
	if err != nil {
		return domain.Audit{}, false, err
	}
	orig, err := os.ReadFile(path)
	if err != nil {
		return domain.Audit{}, false, fmt.Errorf("read audit %s: %w", path, err)
	}
	return editFile("audit", path, orig, now,
		func(content []byte) (domain.Audit, error) { return parseAudit(content, path, bucket) },
		func() error {
			if curPath, _, rerr := s.resolveAudit(slug); rerr != nil || curPath != path {
				return fmt.Errorf("audit %q changed on disk during edit; retry: %w", slug, domain.ErrConflict)
			}
			return nil
		},
		edit)
}
