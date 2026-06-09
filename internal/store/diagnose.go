package store

import (
	"fmt"
	"strconv"
	"strings"

	yaml "go.yaml.in/yaml/v3"
)

// listFields are frontmatter fields tskflwctl expects to be YAML lists. The
// no-PyYAML Python pm sometimes wrote these as bare comma strings, which strict
// YAML rejects — diagnoseFrontmatter turns that into actionable guidance.
var listFields = map[string]bool{
	"tags": true, "related_tasks": true, "dependencies": true,
	"blocks": true, "blocked_by": true, "audit_sources": true, "projects": true,
}

// diagnoseFrontmatter returns an actionable, human-facing explanation of why fm
// won't decode into the typed struct (which field, what's wrong, how to fix),
// or "" if it can't pinpoint a specific field. Called only on a decode failure.
func diagnoseFrontmatter(fm []byte) string {
	var node yaml.Node
	if err := yaml.Unmarshal(fm, &node); err != nil {
		return diagnoseSyntaxError(fm, err)
	}
	mapping := mappingNode(&node)
	if mapping == nil {
		return ""
	}
	for i := 0; i+1 < len(mapping.Content); i += 2 {
		key, val := mapping.Content[i].Value, mapping.Content[i+1]
		switch {
		case listFields[key] && val.Kind != yaml.SequenceNode:
			return fmt.Sprintf("field %q must be a YAML list, but it is %s\n       fix: %s: [%s]",
				key, describeNode(val), key, splitCommaList(val.Value))
		case (key == "tier" || key == "autonomy_level") && isQuotedScalar(val):
			return fmt.Sprintf("field %q must be a whole number, but it is %s", key, describeNode(val))
		}
	}
	return ""
}

func mappingNode(n *yaml.Node) *yaml.Node {
	if n.Kind == yaml.DocumentNode && len(n.Content) > 0 {
		n = n.Content[0]
	}
	if n.Kind == yaml.MappingNode {
		return n
	}
	return nil
}

func describeNode(n *yaml.Node) string {
	switch n.Kind {
	case yaml.ScalarNode:
		return fmt.Sprintf("a string (%q)", n.Value)
	case yaml.MappingNode:
		return "a mapping"
	default:
		return "a scalar"
	}
}

func isQuotedScalar(n *yaml.Node) bool {
	return n.Kind == yaml.ScalarNode && n.Tag != "!!int" && n.Tag != "!!null"
}

func splitCommaList(s string) string {
	parts := strings.Split(s, ",")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	return strings.Join(parts, ", ")
}

// diagnoseSyntaxError turns a YAML parse failure into actionable guidance. The
// most common pm-written breakage is a value containing an unquoted ":" (e.g.
// a description), which YAML reads as a nested mapping.
func diagnoseSyntaxError(fm []byte, err error) string {
	emsg := err.Error()
	if strings.Contains(emsg, "mapping values are not allowed") {
		line := yamlErrorLine(emsg)
		if field := fieldOnLine(fm, line); field != "" {
			return fmt.Sprintf(
				"field %q (line %d) has a value containing an unquoted ':' — wrap the value in quotes, e.g. %s: \"...\"",
				field, line, field)
		}
		if line > 0 {
			return fmt.Sprintf("line %d has a value containing an unquoted ':' — wrap the value in quotes", line)
		}
	}
	return "invalid YAML — " + cleanYAMLError(err)
}

// yamlErrorLine extracts N from a "yaml: line N: ..." error (1-based, relative
// to the frontmatter block).
func yamlErrorLine(emsg string) int {
	idx := strings.Index(emsg, "line ")
	if idx < 0 {
		return 0
	}
	rest := emsg[idx+len("line "):]
	end := strings.IndexByte(rest, ':')
	if end < 0 {
		return 0
	}
	n, _ := strconv.Atoi(strings.TrimSpace(rest[:end]))
	return n
}

// fieldOnLine returns the key on the given 1-based line of the frontmatter.
func fieldOnLine(fm []byte, line int) string {
	if line < 1 {
		return ""
	}
	lines := strings.Split(string(fm), "\n")
	if line-1 >= len(lines) {
		return ""
	}
	if i := strings.IndexByte(lines[line-1], ':'); i > 0 {
		// Only report a plain top-level key. A list item or quoted value with an
		// inner colon (e.g. `- "issue: x"`) would otherwise yield a junk "key".
		if key := strings.TrimSpace(lines[line-1][:i]); isIdentifier(key) {
			return key
		}
	}
	return ""
}

func cleanYAMLError(err error) string {
	msg := strings.TrimPrefix(err.Error(), "yaml: ")
	msg = strings.TrimPrefix(msg, "unmarshal errors:\n  ")
	return strings.TrimSpace(msg)
}

// frontmatterError builds the best available message for a decode failure:
// a git-conflict notice if markers are present (common with synced repos —
// Dropbox/Syncthing/git), then a pinpointed field diagnosis, else a cleaned
// YAML error.
func frontmatterError(fm []byte, structErr error) string {
	if hasConflictMarkers(fm) {
		return "git merge conflict markers detected (<<<<<<< / >>>>>>>) — resolve the conflict before this file can be parsed"
	}
	if msg := diagnoseFrontmatter(fm); msg != "" {
		return msg
	}
	return cleanYAMLError(structErr)
}

// hasConflictMarkers reports whether fm contains git conflict markers. The
// angle-bracket markers are unambiguous (no valid YAML line starts with them).
func hasConflictMarkers(fm []byte) bool {
	for _, line := range strings.Split(string(fm), "\n") {
		if strings.HasPrefix(line, "<<<<<<<") || strings.HasPrefix(line, ">>>>>>>") {
			return true
		}
	}
	return false
}
