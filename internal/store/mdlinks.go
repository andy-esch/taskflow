package store

import (
	"bytes"
	"regexp"
	"sort"
)

// mdLinkRe matches an inline markdown link `[display](target)` — the Scheme-2 body-link form.
var mdLinkRe = regexp.MustCompile(`\[([^\]]*)\]\(([^)]+)\)`)

// mdRefLinkRe matches a reference-style link definition `[label]: target` at the start of a
// line (`target` is the first whitespace-delimited token, so a trailing "title" is left out).
var mdRefLinkRe = regexp.MustCompile(`(?m)^[ \t]*\[[^\]]+\]:[ \t]+(\S+)`)

// inlineCodeRe matches an inline code span — a run of backticks, non-backtick content on the
// same line, then a closing backtick run (`[..](x.md)` shown as prose). Links inside one are
// examples, not real references, so scanLinks skips them (as it does fenced blocks).
var inlineCodeRe = regexp.MustCompile("`+[^`\n]*`+")

// linkRef is one markdown link occurrence found by scanLinks: the byte span of its target
// (what a rename rewrites / a dangler checks) and, for inline links, its display span so a
// rename can freshen stale display text. Reference-style definitions carry no display.
type linkRef struct {
	tStart, tEnd int // target span within the scanned content
	target       string
	dStart, dEnd int // display span (inline only)
	display      string
	inline       bool
}

// scanLinks returns every markdown link — inline `[d](t)` and reference-style `[label]: t` —
// that falls OUTSIDE a code region (fenced block or inline code span), in ascending
// target-position order. Skipping code keeps example links shown in ``` ``` / ~~~ blocks or
// inline `[..](x.md)` spans from being cascaded or dangler-flagged.
func scanLinks(content []byte) []linkRef {
	code := codeSpans(content)
	var refs []linkRef
	for _, m := range mdLinkRe.FindAllSubmatchIndex(content, -1) {
		if inSpans(code, m[0]) {
			continue
		}
		refs = append(refs, linkRef{
			tStart: m[4], tEnd: m[5], target: string(content[m[4]:m[5]]),
			dStart: m[2], dEnd: m[3], display: string(content[m[2]:m[3]]),
			inline: true,
		})
	}
	for _, m := range mdRefLinkRe.FindAllSubmatchIndex(content, -1) {
		if inSpans(code, m[0]) {
			continue
		}
		refs = append(refs, linkRef{
			tStart: m[2], tEnd: m[3], target: string(content[m[2]:m[3]]),
			inline: false,
		})
	}
	sort.Slice(refs, func(i, j int) bool { return refs[i].tStart < refs[j].tStart })
	return refs
}

// codeSpans returns the byte ranges to ignore when scanning for links: fenced code blocks
// unioned with inline code spans. Membership (inSpans) is order-independent, so overlap
// between a fence and a backtick run inside it is harmless.
func codeSpans(content []byte) [][2]int {
	spans := fenceSpans(content)
	for _, m := range inlineCodeRe.FindAllIndex(content, -1) {
		spans = append(spans, [2]int{m[0], m[1]})
	}
	return spans
}

// fenceSpans returns the byte `[start,end)` ranges of fenced code blocks so link scanning can
// skip example links inside them. A fence opens on a line whose first non-blank run is >=3
// backticks or tildes and closes on the next line with an equal-or-longer run of the SAME
// char and no trailing content (an unterminated fence runs to EOF). Backtick opens whose info
// string contains a backtick are ignored (not a real fence per CommonMark).
func fenceSpans(content []byte) [][2]int {
	var spans [][2]int
	var open bool
	var fenceChar byte
	var fenceLen, start, pos int
	for pos <= len(content) {
		lineEnd := len(content)
		if nl := bytes.IndexByte(content[pos:], '\n'); nl >= 0 {
			lineEnd = pos + nl
		}
		line := content[pos:lineEnd]
		i := 0
		for i < len(line) && (line[i] == ' ' || line[i] == '\t') {
			i++
		}
		if i < len(line) && (line[i] == '`' || line[i] == '~') {
			c := line[i]
			j := i
			for j < len(line) && line[j] == c {
				j++
			}
			if runLen := j - i; runLen >= 3 {
				rest := bytes.TrimRight(line[j:], " \t")
				switch {
				case open && c == fenceChar && runLen >= fenceLen && len(rest) == 0:
					end := lineEnd
					if lineEnd < len(content) {
						end = lineEnd + 1 // consume the closing fence's newline
					}
					spans = append(spans, [2]int{start, end})
					open = false
				case !open && (c != '`' || !bytes.ContainsRune(rest, '`')):
					// a backtick fence whose info string holds a backtick is not a fence
					open, fenceChar, fenceLen, start = true, c, runLen, pos
				}
			}
		}
		if lineEnd == len(content) {
			break
		}
		pos = lineEnd + 1
	}
	if open {
		spans = append(spans, [2]int{start, len(content)})
	}
	return spans
}

// inSpans reports whether byte offset pos lies within one of the given spans.
func inSpans(spans [][2]int, pos int) bool {
	for _, s := range spans {
		if pos >= s[0] && pos < s[1] {
			return true
		}
	}
	return false
}
