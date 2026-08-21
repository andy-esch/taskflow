// Package tomledit provides small, comment-preserving edits for configuration files.
// It is deliberately not a TOML encoder: callers parse and validate the document before
// and after the edit, while this package preserves unrelated text byte-for-byte.
package tomledit

import (
	"fmt"
	"strings"
)

// SetTableKey sets one key in a top-level table. encodedValue must already be valid TOML
// for the value; nil removes the key. Existing key position, indentation, and inline
// comment are preserved. Unknown tables, keys, comments, ordering, and whitespace are
// otherwise untouched.
func SetTableKey(text, table, key string, encodedValue *string) (string, bool, error) {
	if strings.TrimSpace(table) == "" || strings.TrimSpace(key) == "" {
		return "", false, fmt.Errorf("table and key are required")
	}
	lines := splitLines(text)
	start, end := tableBounds(lines, table)
	if start >= 0 {
		for i := start + 1; i < end; i++ {
			if !isKeyLine(lines[i], key) {
				continue
			}
			if encodedValue == nil {
				return strings.Join(append(lines[:i:i], lines[i+1:]...), ""), true, nil
			}
			updated := replaceKeyValue(lines[i], key, *encodedValue)
			if updated == lines[i] {
				return text, false, nil
			}
			out := append([]string{}, lines...)
			out[i] = updated
			return strings.Join(out, ""), true, nil
		}
		if encodedValue == nil {
			return text, false, nil
		}
		line := key + " = " + *encodedValue + "\n"
		out := make([]string, 0, len(lines)+1)
		out = append(out, lines[:end]...)
		out = append(out, line)
		out = append(out, lines[end:]...)
		return strings.Join(out, ""), true, nil
	}
	if encodedValue == nil {
		return text, false, nil
	}
	out := text
	if out != "" && !strings.HasSuffix(out, "\n") {
		out += "\n"
	}
	if out != "" && !strings.HasSuffix(out, "\n\n") {
		out += "\n"
	}
	out += fmt.Sprintf("[%s]\n%s = %s\n", table, key, *encodedValue)
	return out, true, nil
}

func splitLines(text string) []string {
	if text == "" {
		return nil
	}
	return strings.SplitAfter(text, "\n")
}

func tableBounds(lines []string, want string) (int, int) {
	start := -1
	for i, line := range lines {
		header, ok := tableHeader(line)
		if !ok {
			continue
		}
		if start >= 0 {
			return start, i
		}
		if header == want {
			start = i
		}
	}
	return start, len(lines)
}

func tableHeader(line string) (string, bool) {
	line = strings.TrimSpace(line)
	if len(line) < 3 || line[0] != '[' || line[1] == '[' {
		return "", false
	}
	end := strings.IndexByte(line, ']')
	if end < 2 {
		return "", false
	}
	return strings.TrimSpace(line[1:end]), true
}

func isKeyLine(line, want string) bool {
	trimmed := strings.TrimSpace(strings.TrimSuffix(line, "\n"))
	if trimmed == "" || strings.HasPrefix(trimmed, "#") {
		return false
	}
	eq := strings.IndexByte(trimmed, '=')
	return eq > 0 && strings.TrimSpace(trimmed[:eq]) == want
}

func replaceKeyValue(line, key, encoded string) string {
	hasNewline := strings.HasSuffix(line, "\n")
	line = strings.TrimSuffix(line, "\n")
	indent := line[:len(line)-len(strings.TrimLeft(line, " \t"))]
	eq := strings.IndexByte(line, '=')
	comment := inlineComment(line[eq+1:])
	out := indent + key + " = " + encoded
	if comment != "" {
		out += " " + comment
	}
	if hasNewline {
		out += "\n"
	}
	return out
}

func inlineComment(value string) string {
	var quote byte
	escaped := false
	for i := 0; i < len(value); i++ {
		c := value[i]
		if escaped {
			escaped = false
			continue
		}
		if quote == '"' && c == '\\' {
			escaped = true
			continue
		}
		if quote != 0 {
			if c == quote {
				quote = 0
			}
			continue
		}
		if c == '"' || c == '\'' {
			quote = c
			continue
		}
		if c == '#' {
			return strings.TrimSpace(value[i:])
		}
	}
	return ""
}
