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
// list fields), relocating a misfiled task to the folder its frontmatter status
// names (ADR-0003 Phase A: frontmatter is authoritative, so the fix MOVES the file
// rather than rewriting the status), and — for tasks and audits — backfilling a
// missing stable id, minted from the entity's own date so it sorts near its real
// age. Epics keep their NN-slug identity and are text-normalized only. Misfiled
// moves are collected during the walk and applied after it, so the fix never
// mutates a directory it is still iterating. When dryRun is true nothing is written.
func (s *FS) FixFrontmatter(dryRun bool) ([]domain.FixResult, error) {
	var results []domain.FixResult
	// Misfiled tasks are relocated AFTER the walk (moving a file mid-iteration would
	// disturb the ReadDir loop it lives in); collected here, applied below.
	var moves []plannedMove
	// Every id already on disk, so a backfill never re-mints one; mintUniqueID adds
	// each id it assigns, so same-date entities in this run stay distinct too.
	seen := s.knownIDs()
	fixDir := func(dir string, dirStatus domain.Status, auditBucket domain.AuditBucket, backfill bool) error {
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
				withID, ok, err := backfillMissingID(fixed, e.Name(), seen)
				if err != nil {
					return err
				}
				if ok {
					fixed = withID
					changes = append(changes, "id: assigned (was missing)")
				}
			}
			// A misfiled task (frontmatter names a valid status that isn't its folder)
			// is repaired by MOVING the file to match the authority — the fixed content
			// (text + id repairs) rides along. Collected now, relocated after the walk.
			if dirStatus != "" {
				if target, ok := misfiledTarget(fixed, dirStatus); ok {
					moves = append(moves, plannedMove{
						from:    path,
						to:      filepath.Join(s.tasksDir, target.Dir(), e.Name()),
						content: fixed,
						changes: append(changes, fmt.Sprintf("status: moved to %s/ to match frontmatter", target)),
					})
					continue
				}
			}
			// Audits: backfill a missing `bucket:` frontmatter (the pre-Phase-A state,
			// where bucket was dir-only), then relocate a misfiled audit (frontmatter
			// bucket ≠ its folder) to match — same collect-then-apply as tasks.
			if auditBucket != "" {
				withBucket, ok, err := backfillMissingBucket(fixed, auditBucket)
				if err != nil {
					return err
				}
				if ok {
					fixed = withBucket
					changes = append(changes, "bucket: assigned (was missing)")
				}
				if target, ok := auditMisfiledTarget(fixed, auditBucket); ok {
					// Only relocate when the target bucket would accept it: a non-open
					// bucket must have no open findings (the invariant MoveAudit enforces).
					// Relocating into a gate-violating state would leave a tree the linter
					// rejects and can't repair, so leave it misfiled for the re-lint to flag.
					_, body := splitFrontmatter(fixed)
					if target == domain.AuditOpen || domain.CountOpenFindings(domain.ParseFindings(string(body))) == 0 {
						moves = append(moves, plannedMove{
							from:    path,
							to:      filepath.Join(s.auditsDir, target.Dir(), e.Name()),
							content: fixed,
							changes: append(changes, fmt.Sprintf("bucket: moved to %s/ to match frontmatter", target)),
						})
						continue
					}
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
		if err := fixDir(filepath.Join(s.tasksDir, st.Dir()), st, "", true); err != nil {
			return results, err
		}
	}
	if err := fixDir(s.epicsDir, "", "", false); err != nil { // epics: text-level only, keep NN-slug identity
		return results, err
	}
	for _, b := range domain.AllAuditBuckets() {
		if err := fixDir(filepath.Join(s.auditsDir, b.Dir()), "", b, true); err != nil { // audits: text + id + bucket backfill + misfiled relocation
			return results, err
		}
	}
	// Apply the misfiled-task relocations now the walk is done. A file already at the
	// target is a real slug collision — skip it (the file stays put, and the re-lint's
	// duplicate-slug check surfaces the pair) rather than clobber the occupant.
	taken := map[string]bool{} // targets claimed this run, so a dry-run preview matches the real one
	for _, mv := range moves {
		// Skip if the target is occupied — on disk (a pre-existing file) OR by an earlier
		// pending move this run (two same-slug misfiles racing for one dir). The loser
		// stays misfiled for the re-lint / dup-slug check; because a dry-run touches
		// neither disk nor the loser's source, `taken` keeps the preview honest.
		if taken[mv.to] {
			continue
		}
		if _, err := os.Stat(mv.to); err == nil {
			continue // target occupied — leave misfiled for the re-lint to flag
		} else if !os.IsNotExist(err) {
			return results, fmt.Errorf("stat %s: %w", mv.to, err)
		}
		taken[mv.to] = true
		if !dryRun {
			if err := os.MkdirAll(filepath.Dir(mv.to), 0o755); err != nil {
				return results, fmt.Errorf("mkdir %s: %w", filepath.Dir(mv.to), err)
			}
			// Write the target first, then remove the source — a crash between the two
			// leaves a recoverable duplicate (dup-slug lint), never a lost file.
			if err := writeFileAtomic(mv.to, mv.content, 0o644); err != nil {
				return results, err
			}
			if err := os.Remove(mv.from); err != nil {
				return results, fmt.Errorf("remove old file %s: %w", mv.from, err)
			}
		}
		results = append(results, domain.FixResult{Path: mv.from, Changes: mv.changes})
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
// timestamping it from the entity's own date so a backfilled id sorts near its
// real age. Date sources, in preference order: the created field, then date, then
// the activity/lifecycle stamps (updated_at, then started/completed/deferred/
// deprecated_at), and finally a YYYY-MM-DD prefix on the file name (the historical
// task/audit naming convention) when the frontmatter carries no date at all. It's
// a no-op (ok=false) when the file already has an id, has no usable date anywhere,
// or won't parse — the re-lint after `--fix` re-flags any id still missing. seen
// dedups against every id already present or assigned this run.
func backfillMissingID(content []byte, filename string, seen map[string]bool) ([]byte, bool, error) {
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
		// No frontmatter date — fall back to a YYYY-MM-DD prefix on the file name,
		// the historical task/audit naming convention, before giving up.
		millis, ok = firstDateMillis(dateFromFilename(filename))
	}
	if !ok {
		return content, false, nil // no usable date anywhere to derive the id's timestamp from
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

// dateFromFilename returns the leading YYYY-MM-DD of a date-prefixed file name
// (e.g. "2025-10-19-slug.md"), or "" when the name doesn't begin with one — the
// filename-date fallback backfillMissingID uses when the frontmatter has no date.
func dateFromFilename(name string) string {
	if len(name) < 10 {
		return ""
	}
	if _, err := time.Parse("2006-01-02", name[:10]); err != nil {
		return ""
	}
	return name[:10]
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

// plannedMove is a misfiled task's relocation, deferred until the walk finishes.
// content carries any text/id repairs so they ride along with the move.
type plannedMove struct {
	from, to string
	content  []byte
	changes  []string
}

// misfiledTarget returns the status directory a task belongs in when its frontmatter
// names a *valid* status that disagrees with its current folder (dirStatus) — i.e.
// the file is misfiled and should be relocated to match the authority (ADR-0003
// Phase A). It's a no-op (ok=false) when the frontmatter status is missing/foreign
// (the folder governs as a fallback), already agrees with the folder, or won't parse.
func misfiledTarget(content []byte, dirStatus domain.Status) (domain.Status, bool) {
	fm, _ := splitFrontmatter(content)
	var t domain.Task
	if len(fm) == 0 || yaml.Unmarshal(fm, &t) != nil {
		return "", false
	}
	if !t.Status.Valid() || t.Status == dirStatus {
		return "", false
	}
	return t.Status, true
}

// backfillMissingBucket adds `bucket: <dir>` to an audit whose frontmatter lacks a
// recognized bucket — the pre-Phase-A state, where bucket was dir-only. A no-op
// (ok=false) when the frontmatter already names a valid bucket, or the file won't parse.
func backfillMissingBucket(content []byte, dirBucket domain.AuditBucket) ([]byte, bool, error) {
	fm, _ := splitFrontmatter(content)
	if len(fm) == 0 {
		return content, false, nil
	}
	var meta struct {
		Bucket domain.AuditBucket `yaml:"bucket"`
	}
	if yaml.Unmarshal(fm, &meta) != nil {
		return content, false, nil
	}
	if strings.TrimSpace(string(meta.Bucket)) != "" {
		// Present already — valid, or a foreign/legacy word we must NOT clobber (backfill
		// is for a truly ABSENT bucket; a bad value is the re-lint's / replace-misfiled's
		// job to surface, not this repair's to silently overwrite).
		return content, false, nil
	}
	out, err := updateFrontmatter(content, map[string]any{"bucket": string(dirBucket)})
	if err != nil {
		return nil, false, err
	}
	return out, true, nil
}

// auditMisfiledTarget returns the bucket dir an audit belongs in when its frontmatter
// names a valid bucket that disagrees with its folder (dirBucket) — a misfiled audit to
// relocate. A no-op when the frontmatter bucket is missing/foreign (the folder governs)
// or already agrees.
func auditMisfiledTarget(content []byte, dirBucket domain.AuditBucket) (domain.AuditBucket, bool) {
	fm, _ := splitFrontmatter(content)
	var meta struct {
		Bucket domain.AuditBucket `yaml:"bucket"`
	}
	if len(fm) == 0 || yaml.Unmarshal(fm, &meta) != nil {
		return "", false
	}
	if !meta.Bucket.Valid() || meta.Bucket == dirBucket {
		return "", false
	}
	return meta.Bucket, true
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
