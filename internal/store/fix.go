package store

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	yaml "go.yaml.in/yaml/v3"

	"github.com/andy-esch/taskflow/internal/domain"
)

// FixFrontmatter walks every task and epic file and applies safe repairs:
// text-level frontmatter normalization (quote unquoted-colon values, normalize
// list fields) and — for task files, where the folder is known — realigning a
// drifted status field to the folder. When dryRun is true nothing is written.
func (s *FS) FixFrontmatter(dryRun bool) ([]domain.FixResult, error) {
	var results []domain.FixResult
	fixDir := func(dir string, dirStatus domain.Status) error {
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

	for _, st := range domain.AllStatuses() {
		if err := fixDir(filepath.Join(s.tasksDir, st.Dir()), st); err != nil {
			return nil, err
		}
	}
	if err := fixDir(s.epicsDir, ""); err != nil { // epics: text-level only
		return nil, err
	}
	return results, nil
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
