package store

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/andy-esch/taskflow/internal/domain"
)

// EditBody replaces (appendMode=false) or appends to (true) a task's markdown
// body, in one atomic, validated write — the agent face of body editing, beside
// the human `task edit`. The frontmatter is preserved surgically (unknown keys,
// comments, and key order survive) and updated_at is stamped. Like the other
// mutators it parses before committing and compare-and-swaps against a concurrent
// relocation; dryRun runs every check but skips the write. Returns the reloaded
// task and the resulting (LF) body, so a --json caller can echo what it wrote.
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

	var newBody string
	if appendMode {
		newBody = appendSection(string(body), text)
	} else {
		newBody = normalizeBody(text)
	}
	newContent, err := replaceBodyStamped(content, newBody, now.Format("2006-01-02"))
	if err != nil {
		return domain.Task{}, "", err
	}
	// Echo the body exactly as it lands on disk (the file's line ending), so a
	// --json caller's echoed body matches what `task show --json` later returns.
	_, storedBody := splitFrontmatter(newContent)
	// Parse before committing: never leave an unreloadable file on disk.
	t, err := parseTask(newContent, path, st)
	if err != nil {
		return domain.Task{}, "", fmt.Errorf("%w: %v", domain.ErrValidation, err)
	}
	if testHookBeforeBodyWrite != nil {
		testHookBeforeBodyWrite()
	}
	// Compare-and-swap before the write (mirrors SetFields/Move): a concurrent
	// move may have relocated the file; writing the original path would resurrect
	// the slug in its old status directory.
	if curPath, _, rerr := s.resolve(slug); rerr != nil || curPath != path {
		return domain.Task{}, "", fmt.Errorf("task %q changed on disk during edit; retry: %w", slug, domain.ErrConflict)
	}
	if dryRun {
		return t, string(storedBody), nil
	}
	if err := writeFileAtomic(path, newContent, 0o644); err != nil {
		return domain.Task{}, "", fmt.Errorf("write task %s: %w", path, err)
	}
	return t, string(storedBody), nil
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
