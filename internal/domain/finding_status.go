package domain

import (
	"regexp"
	"strings"
)

var findingParagraphBreakRe = regexp.MustCompile(`\r?\n[ \t]*\r?\n`)

// findingStatusSource exposes metadata only on the finding's heading or a
// standalone Status line in its section. It preserves byte offsets for guarded
// status writes. A middle-dot example in prose/quotes is not heading metadata;
// neither inline code nor an indented code example grants repair authority.
// Both canonical parsing and candidate classification use this same boundary.
func findingStatusSource(section string) string {
	visible := blankFindingInlineCode(section)
	var out strings.Builder
	for i, line := range strings.SplitAfter(visible, "\n") {
		standalone := strings.TrimLeft(line, " ")
		indent := len(line) - len(standalone)
		if i == 0 || (indent < 4 && strings.HasPrefix(strings.ToLower(standalone), "**status:**")) {
			out.WriteString(line)
			continue
		}
		for _, b := range []byte(line) {
			if b == '\n' || b == '\r' {
				out.WriteByte(b)
			} else {
				out.WriteByte(' ')
			}
		}
	}
	return out.String()
}

// Code spans close with an equal-length backtick run, including across lines.
// Unmatched/escaped opening delimiters remain literal Markdown. Masking bytes
// rather than deleting text keeps finding status spans tied to the original body.
func blankFindingInlineCode(source string) string {
	out := []byte(source)
	for pos := 0; pos < len(source); {
		if source[pos] != '`' {
			pos++
			continue
		}
		start := pos
		for pos < len(source) && source[pos] == '`' {
			pos++
		}
		escapes := 0
		for j := start - 1; j >= 0 && source[j] == '\\'; j-- {
			escapes++
		}
		if escapes%2 != 0 {
			continue
		}
		for next := pos; next < len(source); {
			at := strings.IndexByte(source[next:], '`')
			if at < 0 {
				break
			}
			begin := next + at
			if findingParagraphBreakRe.MatchString(source[pos:begin]) {
				break // inline code cannot consume a later paragraph's metadata
			}
			next = begin
			for next < len(source) && source[next] == '`' {
				next++
			}
			if next-begin != pos-start {
				continue
			}
			for j := start; j < next; j++ {
				if out[j] != '\n' && out[j] != '\r' {
					out[j] = ' '
				}
			}
			pos = next
			break
		}
	}
	return string(out)
}
