package store

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	yaml "go.yaml.in/yaml/v3"

	"github.com/andy-esch/taskflow/internal/domain"
)

// FixFrontmatter walks every task, epic, and audit file and applies safe repairs:
// text-level frontmatter normalization (quote unquoted-colon values, normalize
// list fields) and — for tasks and audits — backfilling a missing stable id from
// the id that already leads the flat filename (ADR-0003 §4), so the frontmatter
// copy can never drift from the name. Epics keep their NN-slug identity and are
// text-normalized only. Under the flat layout there is no relocation: a bad
// status/bucket is lint-flagged, not moved. When dryRun is true nothing is written.
func (s *FS) FixFrontmatter(dryRun bool) ([]domain.FixResult, error) {
	var results []domain.FixResult
	fixDir := func(dir string, backfill bool) error {
		entries, err := os.ReadDir(dir)
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return fmt.Errorf("read dir %s: %w", dir, err)
		}
		for _, e := range entries {
			if !markdownDoc(e) {
				continue
			}
			path := filepath.Join(dir, e.Name())
			content, err := os.ReadFile(path)
			if err != nil {
				return fmt.Errorf("read %s: %w", path, err)
			}
			fixed, changes := fixFrontmatterText(content)
			if backfill {
				withID, ok, err := backfillMissingID(fixed, e.Name())
				if err != nil {
					return err
				}
				if ok {
					fixed = withID
					changes = append(changes, "id: assigned (was missing)")
				}
			}
			if len(changes) == 0 {
				continue
			}
			if !dryRun {
				if err := writeFileAtomic(path, fixed, 0o644); err != nil {
					return err
				}
			}
			results = append(results, domain.FixResult{Path: path, Changes: changes})
		}
		return nil
	}

	// Tasks and audits are both flat (ADR-0003 §4): one dir each, no misfiled relocation
	// (a bad status/bucket is lint-flagged, not moved). Text normalization + id-backfill
	// apply to both; epics are text-level only (they keep their NN-slug identity). On a
	// mid-run write failure the results so far are returned ALONGSIDE the error, since
	// earlier files may already be repaired (atomic writes mean each file is whole).
	if err := fixDir(s.tasksDir, true); err != nil {
		return results, err
	}
	if err := fixDir(s.epicsDir, false); err != nil {
		return results, err
	}
	if err := fixDir(s.auditsDir, true); err != nil {
		return results, err
	}
	// Sweep the tool's own crash-orphaned temp files (housekeeping). Only on a real
	// run — a dry-run previews repairs and must write/remove nothing. The age +
	// prefix guards in sweepStaleTemps keep it from touching a live write or a user
	// file; the .md scan filter already hides these, so this just keeps the tree tidy.
	if !dryRun {
		now := time.Now()
		// Flat entity dirs (ADR-0003 §4): temps land directly in tasks/, epics/, audits/,
		// where writeFileAtomic/createFileAtomic stage them (os.CreateTemp uses the target
		// file's dir) — there are no per-status/bucket subdirs to sweep anymore.
		for _, dir := range []string{s.tasksDir, s.epicsDir, s.auditsDir} {
			for _, p := range sweepStaleTemps(dir, now) {
				results = append(results, domain.FixResult{Path: p, Changes: []string{"removed stale temp orphan"}})
			}
		}
	}
	return results, nil
}

// backfillMissingID fills a task/audit's frontmatter `id:` from the id that already
// leads its flat filename (<id>-<slug>.md) when the frontmatter carries none — the
// canonical key resolveID/CAS match on, so the two can never drift (IDDriftIssue).
// It's a no-op (ok=false) when the file already has an id, is unparseable, or has a
// name that is not id-led — a non-entity stray the scan gate flags for the operator
// to move to meta/ (minting an id into its frontmatter wouldn't make it an entity).
func backfillMissingID(content []byte, filename string) ([]byte, bool, error) {
	fm, _ := splitFrontmatter(content)
	if len(fm) == 0 {
		return content, false, nil
	}
	var meta struct {
		ID string `yaml:"id"`
	}
	if yaml.Unmarshal(fm, &meta) != nil {
		return content, false, nil // unparseable — the text fixer / re-lint handles it
	}
	if strings.TrimSpace(meta.ID) != "" {
		return content, false, nil // already has one
	}
	fnID, _, ok := splitFlatName(strings.TrimSuffix(filename, ".md"))
	if !ok {
		return content, false, nil // not id-led — a stray, not an entity to backfill
	}
	out, err := updateFrontmatter(content, map[string]any{"id": fnID})
	if err != nil {
		return nil, false, err
	}
	return out, true, nil
}

// fixFrontmatterText normalizes a file's frontmatter at the TEXT level — it must
// work on files that don't even parse as YAML. It quotes scalar values
// containing an unquoted ": " and converts list-field values to YAML flow
// lists. Conservative: only touches top-level `key: value` lines.
func fixFrontmatterText(content []byte) ([]byte, []string) {
	fm, body := splitFrontmatter(content)
	if fm == nil {
		return content, nil
	}
	// Re-emit in the file's own line-ending style: a CRLF file must not come
	// back with LF frontmatter over a CRLF body.
	eol := detectLineEnding(content)
	lines := strings.Split(string(fm), "\n")
	var changes []string
	for i, line := range lines {
		line = strings.TrimSuffix(line, "\r")
		key, value, ok := splitKeyValue(line)
		if !ok {
			continue
		}
		fixed, change := fixValue(key, value)
		if change != "" {
			lines[i] = key + ": " + fixed
			changes = append(changes, change)
		} else {
			lines[i] = line
		}
	}
	if len(changes) == 0 {
		return content, nil
	}
	var out bytes.Buffer
	out.WriteString("---" + eol)
	out.WriteString(strings.Join(lines, eol))
	out.WriteString("---" + eol)
	out.Write(body)
	return out.Bytes(), changes
}

// splitKeyValue parses a top-level `key: value` line. It returns ok=false for
// continuation lines, list items, comments, and indented (nested) lines.
func splitKeyValue(line string) (key, value string, ok bool) {
	i := strings.IndexByte(line, ':')
	if i <= 0 {
		return "", "", false
	}
	key = line[:i]
	if key != strings.TrimSpace(key) || !isIdentifier(key) {
		return "", "", false // indented/nested or not a plain key
	}
	return key, strings.TrimSpace(line[i+1:]), true
}

func isIdentifier(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		switch {
		case r == '_' || r == '-':
		case r >= 'a' && r <= 'z':
		case r >= 'A' && r <= 'Z':
		case r >= '0' && r <= '9':
		default:
			return false
		}
	}
	return true
}

func fixValue(key, value string) (fixed, change string) {
	if value == "" {
		return value, "" // empty (e.g. a block list/map follows)
	}
	// Split off a trailing `# comment` first, so a colon *inside the comment*
	// can't drag the comment into the quoted value (it stays a real comment).
	val, comment := splitInlineComment(value)
	if val == "" {
		return value, "" // the value is only a comment
	}
	switch val[0] { // already quoted or a flow/complex value → leave it
	case '"', '\'', '[', '{', '|', '>', '&', '*', '#':
		return value, ""
	}
	suffix := ""
	if comment != "" {
		suffix = " " + comment
	}
	if domain.IsListField(key) {
		return "[" + splitCommaList(val) + "]" + suffix, fmt.Sprintf("%s: normalized to a YAML list", key)
	}
	if strings.Contains(val, ": ") {
		return quoteYAML(val) + suffix, fmt.Sprintf("%s: quoted value containing ':'", key)
	}
	return value, ""
}

// splitInlineComment separates an unquoted YAML scalar from a trailing
// `# comment`. A '#' begins a comment only when preceded by whitespace, so a
// value like `C# rocks` is left intact.
func splitInlineComment(value string) (val, comment string) {
	for i := 1; i < len(value); i++ {
		if value[i] == '#' && (value[i-1] == ' ' || value[i-1] == '\t') {
			return strings.TrimRight(value[:i], " \t"), value[i:]
		}
	}
	return value, ""
}

func quoteYAML(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	return `"` + s + `"`
}
