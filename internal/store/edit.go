package store

import (
	"fmt"
	"os"

	"github.com/andy-esch/taskflow/internal/domain"
)

// EditTask resolves slug, hands the current file content to edit — which runs the
// caller's $EDITOR and returns the new content — and accepts the result only if it
// still parses as a task (parse-before-accept), so a frontmatter break never lands
// on disk. On an invalid edit it reopens the editor on the broken content (passing
// the parse error) and loops until: the edit is valid (atomic write), the user
// re-saves the same broken content unchanged (gives up → ErrValidation), or the
// content matches the original (no write). Returns the reloaded task and whether
// it changed.
//
// The fs and the editor stay decoupled: the store orchestrates resolve/parse/write
// here; the caller's edit callback owns the editor (a cli human-face concern).
func (s *FS) EditTask(slug string, edit func(current string, prevErr error) (string, error)) (domain.Task, bool, error) {
	path, st, err := s.resolve(slug)
	if err != nil {
		return domain.Task{}, false, err
	}
	orig, err := os.ReadFile(path)
	if err != nil {
		return domain.Task{}, false, fmt.Errorf("read task %s: %w", path, err)
	}

	current := string(orig)
	var prevErr error
	for {
		edited, err := edit(current, prevErr)
		if err != nil {
			return domain.Task{}, false, err
		}
		if edited == string(orig) {
			// No net change. Surface a parse error if the file was already broken on
			// disk (opened to inspect, saved unchanged) rather than report success
			// with an empty task.
			t, perr := parseTask(orig, path, st)
			if perr != nil {
				return domain.Task{}, false, fmt.Errorf("%w: %v", domain.ErrValidation, perr)
			}
			return t, false, nil
		}
		t, perr := parseTask([]byte(edited), path, st)
		if perr != nil {
			if edited == current {
				// re-saved the same broken content → the user gave up
				return domain.Task{}, false, fmt.Errorf("%w: %v", domain.ErrValidation, perr)
			}
			current, prevErr = edited, perr // reopen on the broken content
			continue
		}
		// Compare-and-swap before the write (mirrors SetFields/Move): the editor
		// window is long, so a concurrent `task move` may have relocated the file —
		// writing to the original path would resurrect the slug in its old status
		// directory (a permanent ErrAmbiguous). Atomicity guards torn writes, not
		// lost updates.
		if curPath, _, rerr := s.resolve(slug); rerr != nil || curPath != path {
			return domain.Task{}, false, fmt.Errorf("task %q changed on disk during edit; retry: %w", slug, domain.ErrConflict)
		}
		if err := writeFileAtomic(path, []byte(edited), 0o644); err != nil {
			return domain.Task{}, false, fmt.Errorf("write task %s: %w", path, err)
		}
		return t, true, nil
	}
}

// EditAudit is the audit twin of EditTask: resolve the audit, hand its file content
// to edit (the caller's $EDITOR), and accept the result only if it still parses as
// an audit (parse-before-accept), looping on a broken edit. The compare-and-swap
// before the write guards against a concurrent `audit close`/`reopen`/`defer`
// relocating the file across buckets during the (long) editor window. Returns the
// reloaded audit and whether it changed. Finding-level lint (status vocab,
// bucket↔state) is left to the caller, mirroring how task edit leaves field lint to
// `lint` — the store only guarantees the file still parses.
func (s *FS) EditAudit(slug string, edit func(current string, prevErr error) (string, error)) (domain.Audit, bool, error) {
	path, bucket, err := s.resolveAudit(slug)
	if err != nil {
		return domain.Audit{}, false, err
	}
	orig, err := os.ReadFile(path)
	if err != nil {
		return domain.Audit{}, false, fmt.Errorf("read audit %s: %w", path, err)
	}

	current := string(orig)
	var prevErr error
	for {
		edited, err := edit(current, prevErr)
		if err != nil {
			return domain.Audit{}, false, err
		}
		if edited == string(orig) {
			a, perr := parseAudit(orig, path, bucket)
			if perr != nil {
				return domain.Audit{}, false, fmt.Errorf("%w: %v", domain.ErrValidation, perr)
			}
			return a, false, nil
		}
		a, perr := parseAudit([]byte(edited), path, bucket)
		if perr != nil {
			if edited == current {
				return domain.Audit{}, false, fmt.Errorf("%w: %v", domain.ErrValidation, perr)
			}
			current, prevErr = edited, perr // reopen on the broken content
			continue
		}
		if curPath, _, rerr := s.resolveAudit(slug); rerr != nil || curPath != path {
			return domain.Audit{}, false, fmt.Errorf("audit %q changed on disk during edit; retry: %w", slug, domain.ErrConflict)
		}
		if err := writeFileAtomic(path, []byte(edited), 0o644); err != nil {
			return domain.Audit{}, false, fmt.Errorf("write audit %s: %w", path, err)
		}
		return a, true, nil
	}
}
