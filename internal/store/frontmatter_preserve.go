package store

import (
	"bytes"
	"fmt"
	"reflect"
	"strings"

	"github.com/andy-esch/taskflow/internal/domain"
	yaml "go.yaml.in/yaml/v3"
)

// assembleEditedFile retains the source of untouched top-level entries when a
// document contains block scalars. yaml.Node retains the value/style, but its
// encoder refolds text and can add decoded newlines to more-indented folded
// scalars. Keeping whole entries also preserves nested indentation and anchors;
// replacing just a scalar's lines after encoding is not safe.
//
// Existing-document writers share this boundary, including body and graph
// repair edits. Fresh creation still uses assembleFile. If an unusual YAML
// layout cannot be preserved safely, refuse before the atomic write rather than
// silently changing its values.
func assembleEditedFile(original []byte, mapping *yaml.Node, body []byte, eol string) ([]byte, error) {
	var before yaml.Node
	if err := yaml.Unmarshal(original, &before); err != nil {
		return nil, fmt.Errorf("%w: parse frontmatter: %v", errBadFrontmatter, err)
	}
	if !containsBlockScalar(&before) {
		return assembleFile(mapping, body, eol)
	}
	out, err := assembleFile(mapping, body, "\n")
	if err != nil {
		return nil, err
	}
	encoded, _ := splitFrontmatter(out)
	var after yaml.Node
	if err := yaml.Unmarshal(encoded, &after); err != nil {
		return nil, frontmatterFidelityError(err)
	}
	beforeMapping, err := documentMapping(&before)
	if err != nil {
		return nil, err
	}
	afterMapping, err := documentMapping(&after)
	if err != nil {
		return nil, err
	}
	original = bytes.ReplaceAll(original, []byte("\r\n"), []byte("\n"))
	oldEntries, prefix, indent, err := frontmatterEntrySources(original, beforeMapping)
	if err != nil {
		return nil, frontmatterFidelityError(err)
	}
	newEntries, newPrefix, _, err := frontmatterEntrySources(encoded, afterMapping)
	if err != nil {
		return nil, frontmatterFidelityError(err)
	}
	var preserved bytes.Buffer
	if len(mapping.Content) > 0 {
		preserved.Write(prefix)
	}
	for i := 0; i+1 < len(mapping.Content); i += 2 {
		key, value := mapping.Content[i], mapping.Content[i+1]
		old, exists := oldEntries[key.Value]
		if exists && sameYAMLNode(old.key, key) && sameYAMLNode(old.value, value) {
			preserved.Write(old.source)
		} else {
			entry, ok := newEntries[key.Value]
			if !ok {
				return nil, frontmatterFidelityError(fmt.Errorf("cannot locate field %q", key.Value))
			}
			source := entry.source
			// Encoding moves a newly first key's head comment into the prefix.
			// Retain it as well as the original document preamble, but do not
			// duplicate that preamble when the original first key is retained.
			if i == 0 && key.Value != beforeMapping.Content[0].Value {
				source = append(bytes.Clone(newPrefix), source...)
			}
			for _, line := range strings.SplitAfter(string(source), "\n") {
				if line != "" {
					preserved.WriteString(strings.Repeat(" ", indent))
					preserved.WriteString(line)
				}
			}
		}
	}
	if len(mapping.Content) == 0 {
		preserved.Write(encoded)
	}
	var want, got any
	if err := mapping.Decode(&want); err != nil {
		return nil, frontmatterFidelityError(err)
	}
	if err := yaml.Unmarshal(preserved.Bytes(), &got); err != nil {
		return nil, frontmatterFidelityError(err)
	}
	if !reflect.DeepEqual(want, got) {
		return nil, frontmatterFidelityError(fmt.Errorf("decoded values differ from the requested edit"))
	}
	fm := preserved.Bytes()
	if eol != "\n" {
		fm = bytes.ReplaceAll(fm, []byte("\n"), []byte(eol))
	}
	result := []byte("---" + eol)
	result = append(result, fm...)
	result = append(result, []byte("---"+eol)...)
	return append(result, body...), nil
}

func frontmatterFidelityError(err error) error {
	return fmt.Errorf("%w: cannot preserve block-scalar frontmatter safely: %v; edit the document explicitly instead", domain.ErrValidation, err)
}

func containsBlockScalar(node *yaml.Node) bool {
	if node.Kind == yaml.ScalarNode && node.Style&(yaml.LiteralStyle|yaml.FoldedStyle) != 0 {
		return true
	}
	for _, child := range node.Content {
		if containsBlockScalar(child) {
			return true
		}
	}
	return false
}

// sameYAMLNode ignores source coordinates but retains comments and presentation.
// Alias.Value identifies its anchor; do not recursively follow alias pointers.
func sameYAMLNode(a, b *yaml.Node) bool {
	if a.Kind != b.Kind || a.Style != b.Style || a.Tag != b.Tag || a.Value != b.Value || a.Anchor != b.Anchor ||
		a.HeadComment != b.HeadComment || a.LineComment != b.LineComment || a.FootComment != b.FootComment ||
		len(a.Content) != len(b.Content) {
		return false
	}
	for i := range a.Content {
		if !sameYAMLNode(a.Content[i], b.Content[i]) {
			return false
		}
	}
	return true
}

type frontmatterEntrySource struct {
	key, value *yaml.Node
	source     []byte
}

// YAML's parsed key positions delimit block-mapping entries, so a multiline
// value (including nested values) is never bounded by a guessed scalar regex.
func frontmatterEntrySources(source []byte, mapping *yaml.Node) (map[string]frontmatterEntrySource, []byte, int, error) {
	entries := make(map[string]frontmatterEntrySource)
	if len(mapping.Content) == 0 {
		return entries, nil, 0, nil
	}
	if mapping.Style&yaml.FlowStyle != 0 {
		return nil, nil, 0, fmt.Errorf("flow mapping cannot contain source-preserved block entries")
	}
	lines := []int{0}
	for i, b := range source {
		if b == '\n' {
			lines = append(lines, i+1)
		}
	}
	indent := mapping.Content[0].Column - 1
	starts := make([]int, len(mapping.Content)/2)
	for i := 0; i+1 < len(mapping.Content); i += 2 {
		key := mapping.Content[i]
		if key.Kind != yaml.ScalarNode || key.Column != indent+1 || key.Line < 1 || key.Line > len(lines) {
			return nil, nil, 0, fmt.Errorf("unsupported top-level key layout")
		}
		line := key.Line - 1
		// Comments on a following key belong to that entry, not to the previous
		// scalar. Consume exactly the parsed head comment, never the blank lines
		// that a preceding |+ scalar owns. The first entry's document preamble is
		// kept separately, even if that field is removed.
		if i > 0 && key.HeadComment != "" {
			comments := strings.Split(key.HeadComment, "\n")
			for j := len(comments) - 1; j >= 0; j-- {
				if line == 0 || strings.TrimSpace(string(source[lines[line-1]:lines[line]])) != strings.TrimSpace(comments[j]) {
					return nil, nil, 0, fmt.Errorf("cannot locate head comment for %q", key.Value)
				}
				line--
			}
		}
		starts[i/2] = lines[line]
		if i+2 < len(mapping.Content) {
			next := mapping.Content[i+2]
			if next.Line <= key.Line || next.Line > len(lines) {
				return nil, nil, 0, fmt.Errorf("overlapping top-level key positions")
			}
		}
	}
	for i := 0; i+1 < len(mapping.Content); i += 2 {
		key := mapping.Content[i]
		start, end := starts[i/2], len(source)
		if i+2 < len(mapping.Content) {
			end = starts[i/2+1]
		}
		if _, duplicate := entries[key.Value]; duplicate {
			return nil, nil, 0, fmt.Errorf("duplicate key %q", key.Value)
		}
		entries[key.Value] = frontmatterEntrySource{key, mapping.Content[i+1], source[start:end]}
	}
	return entries, source[:starts[0]], indent, nil
}
