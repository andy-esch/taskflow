package store

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/andy-esch/taskflow/internal/domain"
)

// writeBody is the shared body-mutation write path behind AppendAuditBody (audit)
// and EditBody (task). Given a file's content and the already-computed new body it:
// transforms the content (transform swaps in the new body and restores the file's
// EOL — and, for tasks, stamps updated_at), echoes the body exactly as it lands on
// disk (so a --json caller's echo matches a later `show --json`), parses before
// committing so a broken result never lands, compare-and-swaps against a concurrent
// relocation (recheck), then writes atomically. dryRun runs every check but skips
// the write. Returns the reloaded entity and the resulting (LF) body.
//
// The frontmatter-parse guard + body computation stay in each caller (they need the
// current body to append, and the per-entity stamp/parse/resolve differ); this folds
// the identical tail — the parse-before-accept + CAS contract — into one place.
func writeBody[T any](
	noun, path string,
	content []byte,
	newBody string,
	transform func(content []byte, newBody string) ([]byte, error),
	parse func(content []byte) (T, error),
	lock func() (func(), error),
	recheck func() error,
	dryRun bool,
) (T, string, error) {
	var zero T
	newContent, err := transform(content, newBody)
	if err != nil {
		return zero, "", err
	}
	// Echo the body exactly as it lands on disk (the file's line ending), so a
	// --json caller's echoed body matches what `show --json` later returns.
	_, storedBody := splitFrontmatter(newContent)
	// Parse before committing: never leave an unreloadable file on disk.
	v, err := parse(newContent)
	if err != nil {
		return zero, "", fmt.Errorf("%w: %v", domain.ErrValidation, err)
	}
	if testHookBeforeBodyWrite != nil {
		testHookBeforeBodyWrite()
	}
	// Serialize the verify→write critical section (flock) so the CAS is atomic — no
	// cooperating writer can slip a rename between our verify and ours.
	unlock, lockErr := lock()
	if lockErr != nil {
		return zero, "", lockErr
	}
	defer unlock()
	// Compare-and-swap before the write (mirrors SetFields/Move): a concurrent move
	// may have relocated the file during the read→write gap; writing the original
	// path would resurrect the slug in its old directory.
	if err := recheck(); err != nil {
		return zero, "", err
	}
	if dryRun {
		return v, string(storedBody), nil
	}
	if err := writeFileAtomic(path, newContent, 0o644); err != nil {
		return zero, "", fmt.Errorf("write %s %s: %w", noun, path, err)
	}
	return v, string(storedBody), nil
}

// AppendAuditBody appends markdown to an audit's body in one atomic, validated
// write — the audit twin of EditBody's append mode (`audit append`). It stamps
// updated_at like the task path (audits now carry that field); the audit's `date`
// stays untouched — that one is immutable, part of the slug. The shared write tail
// (parse-before-accept, compare-and-swap, dry-run, body echo) lives in writeBody.
// Returns the reloaded audit and the resulting (LF) body.
func (s *FS) AppendAuditBody(slug, text string, now time.Time, dryRun bool) (domain.Audit, string, error) {
	path, bucket, err := s.resolveAudit(slug)
	if err != nil {
		return domain.Audit{}, "", err
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return domain.Audit{}, "", fmt.Errorf("read audit %s: %w", path, err)
	}
	_, body, err := splitFrontmatterStrict(content)
	if err != nil {
		return domain.Audit{}, "", err // can't body-edit a file whose frontmatter won't parse
	}
	updatedAt := now.Format("2006-01-02")
	return writeBody(
		"audit", path, content, appendSection(string(body), text),
		func(c []byte, nb string) ([]byte, error) { return replaceBodyStamped(c, nb, updatedAt) },
		func(c []byte) (domain.Audit, error) { return parseAudit(c, path, bucket) },
		s.writeLock,
		// A concurrent bucket move (relocate) OR in-place edit during the read→write gap.
		func() error {
			return verifyUnchanged(s.resolveAuditPath, slug, path, hashContent(content), "audit", "edit")
		},
		dryRun,
	)
}

// EditBody replaces (appendMode=false) or appends to (true) a task's markdown
// body, in one atomic, validated write — the agent face of body editing, beside
// the human `task edit`. The frontmatter is preserved surgically (unknown keys,
// comments, and key order survive) and updated_at is stamped. The shared write tail
// lives in writeBody. Returns the reloaded task and the resulting (LF) body.
func (s *FS) EditBody(slug, text string, appendMode bool, now time.Time, dryRun bool) (domain.Task, string, error) {
	path, st, err := s.resolve(slug)
	if err != nil {
		return domain.Task{}, "", err
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return domain.Task{}, "", fmt.Errorf("read task %s: %w", path, err)
	}
	_, body, err := splitFrontmatterStrict(content)
	if err != nil {
		return domain.Task{}, "", err // can't body-edit a file whose frontmatter won't parse
	}
	newBody := normalizeBody(text)
	if appendMode {
		newBody = appendSection(string(body), text)
	}
	updatedAt := now.Format("2006-01-02")
	return writeBody(
		"task", path, content, newBody,
		func(c []byte, nb string) ([]byte, error) { return replaceBodyStamped(c, nb, updatedAt) },
		func(c []byte) (domain.Task, error) { return parseTask(c, path, st) },
		s.writeLock,
		func() error { return verifyUnchanged(s.resolvePath, slug, path, hashContent(content), "task", "edit") },
		dryRun,
	)
}

// normalizeBody works in LF internally (any CRLF/CR input is folded to LF), trims
// trailing blank lines, and guarantees a single trailing newline (the markdown-file
// convention), leaving an empty body empty. replaceBodyStamped re-emits the result
// in the file's actual line ending, so the whole file stays one EOL style.
func normalizeBody(text string) string {
	text = toLF(text)
	text = strings.TrimRight(text, "\n")
	if text == "" {
		return ""
	}
	return text + "\n"
}

// appendSection joins an addition to the end of an existing body, separated by one
// blank line, with a single trailing newline. Both sides are folded to LF first so
// a CRLF body can't leave a stray \r at the seam (replaceBodyStamped restores EOL).
func appendSection(old, addition string) string {
	add := strings.Trim(toLF(addition), "\n")
	if add == "" {
		return normalizeBody(old)
	}
	old = strings.TrimRight(toLF(old), "\n")
	if old == "" {
		return add + "\n"
	}
	return old + "\n\n" + add + "\n"
}

// toLF folds CRLF and bare CR to LF, so body manipulation happens in one ending.
func toLF(s string) string {
	return strings.ReplaceAll(strings.ReplaceAll(s, "\r\n", "\n"), "\r", "\n")
}

// testHookBeforeBodyWrite runs between EditBody's validation and its compare-and-
// swap, so a test can simulate a concurrent relocation in that window (mirrors
// testHookBeforeSetFieldsWrite). nil in production.
var testHookBeforeBodyWrite func()
