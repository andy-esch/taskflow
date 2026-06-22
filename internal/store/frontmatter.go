// Package store is the secondary adapter: planning tasks as markdown+frontmatter
// files on disk, behind the core's TaskStore port.
package store

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"

	yaml "go.yaml.in/yaml/v3"

	"github.com/andy-esch/taskflow/internal/domain"
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
		if string(bytes.TrimRight(line, " \t\r")) == "---" { // closing fence (tolerate trailing whitespace, a common editor artifact)
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
	return nil, content // no closing fence; splitFrontmatterStrict flags this
}

// splitFrontmatterStrict is splitFrontmatter plus the unterminated-fence check:
// an opening `---` with no closing fence is malformed (a FileProblem) — NOT "no
// frontmatter". Treating it as the latter would list the file as an empty task
// and let a surgical edit prepend a *second* frontmatter block, demoting the
// partial one into the body.
func splitFrontmatterStrict(content []byte) (frontmatter, body []byte, err error) {
	fm, body := splitFrontmatter(content)
	if fm == nil && (bytes.HasPrefix(content, []byte("---\n")) || bytes.HasPrefix(content, []byte("---\r\n"))) {
		return nil, nil, fmt.Errorf("%w: unterminated frontmatter (no closing ---)", errBadFrontmatter)
	}
	return fm, body, nil
}

// detectLineEnding returns the file's dominant line ending, so a surgical edit
// re-emits the frontmatter in the style the file already uses — a CRLF file
// must not come back with a mixed-ending (LF frontmatter / CRLF body) diff.
func detectLineEnding(content []byte) string {
	crlf := bytes.Count(content, []byte("\r\n"))
	if lf := bytes.Count(content, []byte("\n")) - crlf; crlf > lf {
		return "\r\n"
	}
	return "\n"
}

// updateFrontmatter applies key=value updates to a file's frontmatter, then
// reassembles the file. It edits a yaml.Node surgically, so unknown/custom
// fields, comments, and key order survive; the body is preserved verbatim.
// Values may be string, int, or []string.
func updateFrontmatter(content []byte, updates map[string]any) ([]byte, error) {
	fm, body, err := splitFrontmatterStrict(content)
	if err != nil {
		return nil, err
	}

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
		if _, unset := v.(domain.UnsetField); unset {
			deleteMapNode(mapping, k)
			continue
		}
		node, err := valueNode(v)
		if err != nil {
			return nil, fmt.Errorf("field %q: %w", k, err)
		}
		setMapNode(mapping, k, node)
	}

	return assembleFile(mapping, body, detectLineEnding(content))
}

// assembleFile encodes a frontmatter mapping node and reattaches the `---`
// fences and body, using eol for the fences and the (LF-encoded) YAML block so
// the whole file keeps one line-ending style. Shared by surgical updates
// (which pass the file's detected ending) and fresh-file creation (LF).
func assembleFile(mapping *yaml.Node, body []byte, eol string) ([]byte, error) {
	var fmBuf bytes.Buffer
	enc := yaml.NewEncoder(&fmBuf)
	enc.SetIndent(2)
	if err := enc.Encode(mapping); err != nil {
		return nil, fmt.Errorf("encode frontmatter: %w", err)
	}
	_ = enc.Close()

	fmBytes := fmBuf.Bytes()
	if eol != "\n" {
		fmBytes = bytes.ReplaceAll(fmBytes, []byte("\n"), []byte(eol))
	}
	var out bytes.Buffer
	out.WriteString("---" + eol)
	out.Write(fmBytes)
	out.WriteString("---" + eol)
	out.Write(body)
	return out.Bytes(), nil
}

// replaceBodyStamped swaps a file's markdown body for newBody and stamps
// updated_at, preserving the frontmatter surgically — unknown keys, comments, and
// key order survive (the same yaml.Node path as updateFrontmatter, except the body
// is replaced rather than kept verbatim).
func replaceBodyStamped(content []byte, newBody, updatedAt string) ([]byte, error) {
	fm, _, err := splitFrontmatterStrict(content)
	if err != nil {
		return nil, err
	}
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
	setMapNode(mapping, "updated_at", &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: updatedAt})
	// newBody is built in LF; re-emit it in the file's own ending so a CRLF file
	// doesn't come back with a mixed CRLF-frontmatter / LF-body diff (assembleFile
	// only converts the frontmatter block, writing the body verbatim).
	eol := detectLineEnding(content)
	body := newBody
	if eol != "\n" {
		body = strings.ReplaceAll(newBody, "\n", eol)
	}
	return assembleFile(mapping, []byte(body), eol)
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

// deleteMapNode removes key (and its value) from the mapping; absent keys are
// a no-op, so unset is idempotent.
func deleteMapNode(mapping *yaml.Node, key string) {
	for i := 0; i+1 < len(mapping.Content); i += 2 {
		if mapping.Content[i].Value == key {
			mapping.Content = append(mapping.Content[:i], mapping.Content[i+2:]...)
			return
		}
	}
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
