// Package store is the secondary adapter: planning tasks as markdown+frontmatter
// files on disk, behind the core's TaskStore port.
package store

import (
	"bytes"
	"fmt"
	"strconv"

	yaml "go.yaml.in/yaml/v3"
)

// splitFrontmatter separates a leading `---`-fenced YAML block from the markdown
// body. Returns (nil, content) when there is no frontmatter. Zero-dependency
// byte scan — deliberately not pulling in a markdown AST parser for a ~15-line
// job.
func splitFrontmatter(content []byte) (frontmatter, body []byte) {
	if !bytes.HasPrefix(content, []byte("---\n")) && !bytes.HasPrefix(content, []byte("---\r\n")) {
		return nil, content
	}
	nl := bytes.IndexByte(content, '\n')
	rest := content[nl+1:]

	offset := 0
	for offset <= len(rest) {
		lineEnd := bytes.IndexByte(rest[offset:], '\n')
		var line []byte
		if lineEnd < 0 {
			line = rest[offset:]
		} else {
			line = rest[offset : offset+lineEnd]
		}
		if string(bytes.TrimRight(line, "\r")) == "---" { // closing fence
			frontmatter = rest[:offset]
			if lineEnd < 0 {
				return frontmatter, nil
			}
			return frontmatter, rest[offset+lineEnd+1:]
		}
		if lineEnd < 0 {
			break
		}
		offset += lineEnd + 1
	}
	return nil, content // no closing fence → treat as no frontmatter
}

// updateFrontmatter applies key=value updates to a file's frontmatter, then
// reassembles the file. It edits a yaml.Node surgically, so unknown/custom
// fields, comments, and key order survive; the body is preserved verbatim.
// Values may be string, int, or []string.
func updateFrontmatter(content []byte, updates map[string]any) ([]byte, error) {
	fm, body := splitFrontmatter(content)

	var doc yaml.Node
	if len(bytes.TrimSpace(fm)) > 0 {
		if err := yaml.Unmarshal(fm, &doc); err != nil {
			return nil, fmt.Errorf("%w: parse frontmatter: %v", errBadFrontmatter, err)
		}
	}
	mapping, err := documentMapping(&doc)
	if err != nil {
		return nil, err
	}
	for k, v := range updates {
		node, err := valueNode(v)
		if err != nil {
			return nil, fmt.Errorf("field %q: %w", k, err)
		}
		setMapNode(mapping, k, node)
	}

	var fmBuf bytes.Buffer
	enc := yaml.NewEncoder(&fmBuf)
	enc.SetIndent(2)
	if err := enc.Encode(mapping); err != nil {
		return nil, fmt.Errorf("encode frontmatter: %w", err)
	}
	_ = enc.Close()

	var out bytes.Buffer
	out.WriteString("---\n")
	out.Write(fmBuf.Bytes())
	out.WriteString("---\n")
	out.Write(body)
	return out.Bytes(), nil
}

// documentMapping returns the top-level mapping node, creating an empty one if
// the document has no frontmatter yet. If the frontmatter is valid YAML but not
// a mapping (a bare scalar or a sequence), it returns an error rather than
// silently discarding the existing content — overwriting it would lose data.
func documentMapping(doc *yaml.Node) (*yaml.Node, error) {
	if doc.Kind == 0 { // empty document — no frontmatter yet
		m := &yaml.Node{Kind: yaml.MappingNode}
		doc.Kind = yaml.DocumentNode
		doc.Content = []*yaml.Node{m}
		return m, nil
	}
	if doc.Kind == yaml.DocumentNode && len(doc.Content) == 1 && doc.Content[0].Kind == yaml.MappingNode {
		return doc.Content[0], nil
	}
	return nil, fmt.Errorf("%w: frontmatter is not a key/value mapping", errBadFrontmatter)
}

// setMapNode replaces key's value node in place (preserving position) or
// appends the key. Replacing the whole node lets a scalar become a list, etc.;
// any comments attached to the old value node are carried onto the new one.
func setMapNode(mapping *yaml.Node, key string, val *yaml.Node) {
	for i := 0; i+1 < len(mapping.Content); i += 2 {
		if mapping.Content[i].Value == key {
			old := mapping.Content[i+1]
			val.HeadComment, val.LineComment, val.FootComment = old.HeadComment, old.LineComment, old.FootComment
			mapping.Content[i+1] = val
			return
		}
	}
	mapping.Content = append(mapping.Content,
		&yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: key}, val)
}

func valueNode(v any) (*yaml.Node, error) {
	switch x := v.(type) {
	case string:
		return &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: x}, nil
	case int:
		return &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!int", Value: strconv.Itoa(x)}, nil
	case []string:
		seq := &yaml.Node{Kind: yaml.SequenceNode, Tag: "!!seq", Style: yaml.FlowStyle}
		for _, s := range x {
			seq.Content = append(seq.Content, &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: s})
		}
		return seq, nil
	default:
		return nil, fmt.Errorf("unsupported value type %T", v)
	}
}
