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
	"github.com/andy-esch/taskflow/internal/id"
)

// FixFrontmatter walks every task, epic, and audit file and applies safe repairs:
// text-level frontmatter normalization (quote unquoted-colon values, normalize
// list fields), realigning a drifted task status to its folder, and — for tasks
// and audits — backfilling a missing stable id (ADR-0003), minted from the
// entity's own date so it sorts near its real age. Epics keep their NN-slug
// identity and are text-normalized only. When dryRun is true nothing is written.
func (s *FS) FixFrontmatter(dryRun bool) ([]domain.FixResult, error) {
	var results []domain.FixResult
	// Every id already on disk, so a backfill never re-mints one; mintUniqueID adds
	// each id it assigns, so same-date entities in this run stay distinct too.
	seen := s.knownIDs()
	fixDir := func(dir string, dirStatus domain.Status, backfill bool) error {
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
			if dirStatus != "" {
				if realigned, ok := realignStatus(fixed, dirStatus); ok {
					fixed = realigned
					changes = append(changes, fmt.Sprintf("status: realigned to folder %q", dirStatus))
				}
			}
			if backfill {
				withID, ok, err := backfillMissingID(fixed, seen)
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

	// On a mid-run write failure, return the results accumulated so far ALONGSIDE
	// the error: files in earlier dirs may already be repaired, and the caller must
	// be able to report that partial progress rather than discard it (atomic writes
	// mean each file is whole; only the run is incomplete).
	for _, st := range domain.AllStatuses() {
		if err := fixDir(filepath.Join(s.tasksDir, st.Dir()), st, true); err != nil {
			return results, err
		}
	}
	if err := fixDir(s.epicsDir, "", false); err != nil { // epics: text-level only, keep NN-slug identity
		return results, err
	}
	for _, b := range domain.AllAuditBuckets() {
		if err := fixDir(filepath.Join(s.auditsDir, b.Dir()), "", true); err != nil { // audits: text + id backfill
			return results, err
		}
	}
	// Sweep the tool's own crash-orphaned temp files (housekeeping). Only on a real
	// run — a dry-run previews repairs and must write/remove nothing. The age +
	// prefix guards in sweepStaleTemps keep it from touching a live write or a user
	// file; the .md scan filter already hides these, so this just keeps the tree tidy.
	if !dryRun {
		now := time.Now()
		dirs := []string{s.epicsDir}
		for _, st := range domain.AllStatuses() {
			dirs = append(dirs, filepath.Join(s.tasksDir, st.Dir()))
		}
		for _, b := range domain.AllAuditBuckets() {
			dirs = append(dirs, filepath.Join(s.auditsDir, b.Dir()))
		}
		for _, dir := range dirs {
			for _, p := range sweepStaleTemps(dir, now) {
				results = append(results, domain.FixResult{Path: p, Changes: []string{"removed stale temp orphan"}})
			}
		}
	}
	return results, nil
}

// knownIDs gathers every stable id already assigned across tasks and audits so a
// backfill never mints a duplicate. Best-effort: a listing error yields a partial
// set — mintUniqueID's per-run tracking still prevents new-vs-new collisions.
func (s *FS) knownIDs() map[string]bool {
	seen := map[string]bool{}
	if tasks, _, err := s.ListTasks(); err == nil {
		for _, t := range tasks {
			if t.ID != "" {
				seen[t.ID] = true
			}
		}
	}
	if audits, _, err := s.ListAudits(); err == nil {
		for _, a := range audits {
			if a.ID != "" {
				seen[a.ID] = true
			}
		}
	}
	return seen
}

// backfillMissingID appends a stable id to a file whose frontmatter lacks one,
// timestamping it from the entity's own date — created, else the audit slug date,
// else any activity/lifecycle stamp (updated_at, then started/completed/deferred/
// deprecated_at) — so a backfilled id sorts near the entity's real age. It's a
// no-op (ok=false) when the file already has an id, has no usable date at all, or
// won't parse — the re-lint after `--fix` re-flags any id still missing. seen
// dedups against every id already present or assigned this run.
func backfillMissingID(content []byte, seen map[string]bool) ([]byte, bool, error) {
	fm, _ := splitFrontmatter(content)
	if len(fm) == 0 {
		return content, false, nil
	}
	var meta struct {
		ID         string `yaml:"id"`
		Created    string `yaml:"created"`
		Date       string `yaml:"date"`
		Updated    string `yaml:"updated_at"`
		Started    string `yaml:"started_at"`
		Completed  string `yaml:"completed_at"`
		Deferred   string `yaml:"deferred_at"`
		Deprecated string `yaml:"deprecated_at"`
	}
	if yaml.Unmarshal(fm, &meta) != nil {
		return content, false, nil // unparseable — the text fixer / re-lint handles it
	}
	if strings.TrimSpace(meta.ID) != "" {
		return content, false, nil // already has one
	}
	// Preference: birth date first, then the last-activity/lifecycle stamps — an
	// archived task may carry only a completed_at/deprecated_at, which still dates
	// the id better than nothing.
	millis, ok := firstDateMillis(meta.Created, meta.Date, meta.Updated,
		meta.Started, meta.Completed, meta.Deferred, meta.Deprecated)
	if !ok {
		return content, false, nil // no usable date to derive the id's timestamp from
	}
	newID, ok := mintUniqueID(millis, seen, id.NewAt)
	if !ok {
		return content, false, nil
	}
	out, err := updateFrontmatter(content, map[string]any{"id": newID})
	if err != nil {
		return nil, false, err
	}
	return out, true, nil
}

// firstDateMillis returns the UTC-midnight Unix millis of the first parseable
// YYYY-MM-DD among candidates (in preference order), false when none parse.
func firstDateMillis(candidates ...string) (int64, bool) {
	for _, d := range candidates {
		if d = strings.TrimSpace(d); d == "" {
			continue
		}
		if t, err := time.Parse("2006-01-02", d); err == nil {
			return t.UnixMilli(), true
		}
	}
	return 0, false
}

// mintUniqueID mints an id stamped at millis, re-rolling until it avoids every id
// in seen. id.NewAt is stateless-random, so each retry re-rolls the 17-bit tail;
// the cap guards a pathological generator (unreachable with real entropy). The
// returned id is recorded in seen.
func mintUniqueID(millis int64, seen map[string]bool, gen func(int64) string) (string, bool) {
	for i := 0; i < 64; i++ {
		v := gen(millis)
		if !seen[v] {
			seen[v] = true
			return v, true
		}
	}
	return "", false
}

// realignStatus rewrites a task's frontmatter status to dirStatus when the
// declared status is a *valid* but different status (the misfiled case). A
// foreign/invalid status word is left untouched — the folder governs anyway —
// and an unparseable file is skipped (the text fixer handles those).
func realignStatus(content []byte, dirStatus domain.Status) ([]byte, bool) {
	fm, _ := splitFrontmatter(content)
	var t domain.Task
	if len(fm) == 0 || yaml.Unmarshal(fm, &t) != nil {
		return content, false
	}
	if !t.Status.Valid() || t.Status == dirStatus {
		return content, false
	}
	out, err := updateFrontmatter(content, map[string]any{"status": string(dirStatus)})
	if err != nil {
		return content, false
	}
	return out, true
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
