// Package architecturecheck verifies the bounded architecture navigation surface.
// It is not a Markdown linter or proof of the prose's semantic compatibility.
package architecturecheck

import (
	"fmt"
	"go/parser"
	"go/token"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"
	"unicode"
)

var (
	// Canonical guides use ordinary inline links, ATX headings and backtick/tilde
	// fences. Keep the test deliberately scoped; general docs drift is separate work.
	localLink  = regexp.MustCompile(`\[[^\]]*\]\(([^()\s]+)\)`)
	guideRoute = regexp.MustCompile(`docs/architecture/[a-z-]+\.md(?:#[a-z0-9-]+)?`)
)

func TestArchitectureNavigation(t *testing.T) {
	root := filepath.Join("..", "..", "..")
	guides, err := filepath.Glob(filepath.Join(root, "docs", "architecture", "*.md"))
	if err != nil {
		t.Fatal(err)
	}
	if len(guides) == 0 {
		t.Fatal("no focused architecture guides found")
	}
	documents := append([]string{
		filepath.Join(root, "docs", "ARCHITECTURE.md"),
		filepath.Join(root, "AGENTS.md"),
		filepath.Join(root, "README.md"),
	}, guides...)
	for _, document := range documents {
		content := readFile(t, document)
		for _, target := range inlineLinks(content) {
			if err := checkTarget(document, target); err != nil {
				t.Errorf("%s: %v", document, err)
			}
		}
	}
	index := readFile(t, filepath.Join(root, "docs", "ARCHITECTURE.md"))
	for _, guide := range guides {
		t.Run(filepath.Base(guide), func(t *testing.T) {
			if !slices.Contains(inlineLinks(index), "architecture/"+filepath.Base(guide)) {
				t.Error("architecture index must link directly to this guide")
			}
			content := readFile(t, guide)
			if !slices.Contains(inlineLinks(content), "../ARCHITECTURE.md") {
				t.Error("guide must link back to the architecture index")
			}
			// Derive sources from the guide's links rather than maintaining another
			// package/file inventory. At least one linked package overview must route
			// back; every route found in linked overviews must remain navigable.
			hasReturnRoute := false
			for _, raw := range inlineLinks(content) {
				target, err := url.Parse(raw)
				if err != nil || target.IsAbs() || target.Host != "" || !strings.HasSuffix(target.Path, ".go") {
					continue
				}
				source := filepath.Join(filepath.Dir(guide), filepath.FromSlash(target.Path))
				file, err := parser.ParseFile(token.NewFileSet(), source, nil, parser.PackageClauseOnly|parser.ParseComments)
				if err != nil {
					t.Errorf("package comment %s: %v", source, err)
					continue
				}
				if file.Doc == nil {
					continue
				}
				for _, route := range guideRoute.FindAllString(file.Doc.Text(), -1) {
					if err := checkTarget(filepath.Join(root, "AGENTS.md"), route); err != nil {
						t.Errorf("package comment %s: %v", source, err)
					}
					if strings.SplitN(route, "#", 2)[0] == "docs/architecture/"+filepath.Base(guide) {
						hasReturnRoute = true
					}
				}
			}
			if !hasReturnRoute {
				t.Error("guide needs a linked owning package comment that routes back to it")
			}
		})
	}
}

// Tasks and Threads are the current consumers of retained forwarding anchors.
// Derive their incoming links, not a second inventory of historical headings.
// This checks only their routes to the architecture index, not all planning
// prose or historical audits' source-line citations.
func TestArchitectureForwardingAnchors(t *testing.T) {
	root := filepath.Join("..", "..", "..")
	index := filepath.Join(root, "docs", "ARCHITECTURE.md")
	for _, kind := range []string{"tasks", "threads"} {
		documents, err := filepath.Glob(filepath.Join(root, "planning", kind, "*.md"))
		if err != nil {
			t.Fatal(err)
		}
		for _, document := range documents {
			for _, target := range linksTo(document, readFile(t, document), index) {
				if err := checkTarget(document, target); err != nil {
					t.Errorf("%s: %v", document, err)
				}
			}
		}
	}
}

func inlineLinks(content string) []string {
	content = withoutInlineCode(withoutFences(content))
	var targets []string
	for _, match := range localLink.FindAllStringSubmatchIndex(content, -1) {
		if match[0] > 0 && (content[match[0]-1] == '!' || content[match[0]-1] == '\\') {
			continue // Images and escaped examples are not navigation links.
		}
		targets = append(targets, content[match[2]:match[3]])
	}
	return targets
}

func linksTo(document, content, destination string) []string {
	var targets []string
	for _, raw := range inlineLinks(content) {
		target, err := url.Parse(raw)
		if err != nil || target.IsAbs() || target.Host != "" {
			continue
		}
		path := document
		if target.Path != "" {
			path = filepath.Join(filepath.Dir(document), filepath.FromSlash(target.Path))
		}
		if filepath.Clean(path) == filepath.Clean(destination) {
			targets = append(targets, raw)
		}
	}
	return targets
}

// The guides use ordinary Markdown code spans. Strip matching backtick runs
// before parsing links, but leave unmatched runs as prose. This is deliberately
// not a general Markdown parser (reference links, HTML, and nested labels are out
// of scope); multiline labels and code spans are supported.
func withoutInlineCode(content string) string {
	result := []byte(content)
	for i := 0; i < len(content); {
		if content[i] != '`' {
			i++
			continue
		}
		end := i
		for end < len(content) && content[end] == '`' {
			end++
		}
		closed := false
		for j := end; j < len(content); {
			if content[j] != '`' {
				j++
				continue
			}
			closeEnd := j
			for closeEnd < len(content) && content[closeEnd] == '`' {
				closeEnd++
			}
			if closeEnd-j == end-i {
				for k := i; k < closeEnd; k++ {
					if result[k] != '\n' {
						result[k] = ' '
					}
				}
				i, closed = closeEnd, true
				break
			}
			j = closeEnd
		}
		if !closed {
			i = end
		}
	}
	return string(result)
}

func checkTarget(document, raw string) error {
	target, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("invalid target %q: %w", raw, err)
	}
	if target.IsAbs() || target.Host != "" {
		return nil // Network/content verification is intentionally not this test's job.
	}
	path := document
	if target.Path != "" {
		path = filepath.Join(filepath.Dir(document), filepath.FromSlash(target.Path))
	}
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("target %q: %w", raw, err)
	}
	if target.Fragment == "" {
		return nil
	}
	if info.IsDir() || !strings.HasSuffix(path, ".md") {
		return fmt.Errorf("unsupported local fragment target %q", raw)
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("target %q: %w", raw, err)
	}
	if !headings(string(content))[target.Fragment] {
		return fmt.Errorf("target %q: missing heading anchor %q", raw, target.Fragment)
	}
	return nil
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(content)
}

func withoutFences(content string) string {
	lines := strings.Split(content, "\n")
	var fence byte
	var length int
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if len(trimmed) >= 3 && (trimmed[0] == '`' || trimmed[0] == '~') {
			n := 0
			for n < len(trimmed) && trimmed[n] == trimmed[0] {
				n++
			}
			if n >= 3 && fence == 0 {
				fence, length = trimmed[0], n
				lines[i] = ""
				continue
			}
			if trimmed[0] == fence && n >= length && strings.TrimSpace(trimmed[n:]) == "" {
				fence = 0
				lines[i] = ""
				continue
			}
		}
		if fence != 0 {
			lines[i] = ""
		}
	}
	return strings.Join(lines, "\n")
}

func headings(content string) map[string]bool {
	anchors := make(map[string]bool)
	counts := make(map[string]int)
	for _, line := range strings.Split(withoutFences(content), "\n") {
		title := strings.TrimLeft(line, "#")
		if title == line || !strings.HasPrefix(title, " ") {
			continue
		}
		title = strings.TrimSpace(strings.TrimRight(strings.TrimSpace(title), "#"))
		var slug strings.Builder
		for _, r := range strings.ToLower(title) {
			switch {
			case unicode.IsLetter(r), unicode.IsDigit(r), r == '-', r == '_':
				slug.WriteRune(r)
			case r == ' ':
				slug.WriteByte('-')
			}
		}
		key := slug.String()
		n := counts[key]
		counts[key]++
		if n != 0 {
			key = fmt.Sprintf("%s-%d", key, n)
		}
		anchors[key] = true
	}
	return anchors
}

func TestHeadingAnchorsIgnoreFencedExamples(t *testing.T) {
	content := "# Overview\n## The TUI (`internal/tui`)\n~~~markdown\n## Hidden\n~~~\n## Overview\n"
	got := headings(content)
	for _, want := range []string{"overview", "the-tui-internaltui", "overview-1"} {
		if !got[want] {
			t.Errorf("missing anchor %q", want)
		}
	}
	if got["hidden"] {
		t.Error("fenced example is not a document heading")
	}
}

func TestInlineLinksRequireNavigableProse(t *testing.T) {
	for _, target := range []string{"architecture/tui.md", "../ARCHITECTURE.md"} {
		t.Run(target, func(t *testing.T) {
			link := "[Guide](" + target + ")"
			for name, content := range map[string]string{
				"live":            link,
				"multiline label": "[Guide\nwith context](" + target + ")",
				"code label":      "[`Guide`](" + target + ")",
				"unmatched tick":  "` unmatched " + link,
			} {
				t.Run(name, func(t *testing.T) {
					if got := inlineLinks(content); !slices.Equal(got, []string{target}) {
						t.Errorf("links = %q, want live target %q", got, target)
					}
				})
			}
			for name, content := range map[string]string{
				"bare path":          "(" + target + ")",
				"backtick fence":     "```text\n" + link + "\n```\n",
				"tilde fence":        "~~~markdown\n" + link + "\n~~~\n",
				"inline code":        "`" + link + "`",
				"multi-tick code":    "`` ` " + link + " ` ``",
				"multiline code":     "`example\n" + link + "`",
				"image":              "!" + link,
				"escaped link label": "\\" + link,
			} {
				t.Run(name, func(t *testing.T) {
					if got := inlineLinks(content); len(got) != 0 {
						t.Errorf("non-navigable mention produced links %q", got)
					}
				})
			}
		})
	}
}

func TestIncomingForwardingAnchorRequiresExistingHeading(t *testing.T) {
	dir := t.TempDir()
	index := filepath.Join(dir, "ARCHITECTURE.md")
	if err := os.WriteFile(index, []byte("# Architecture\n## Planning-data change checklist\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	document := filepath.Join(dir, "planning", "task.md")
	want := "../ARCHITECTURE.md#planning-data-change-checklist"
	content := "[Checklist](" + want + ")\n" +
		"[Elsewhere](elsewhere.md#ignored)\n" +
		"[Hosted](https://example.org/ARCHITECTURE.md#ignored)\n" +
		"[Host-relative](//example.org/ARCHITECTURE.md#ignored)\n" +
		"```markdown\n[Historical example](../ARCHITECTURE.md#ignored)\n```\n"
	targets := linksTo(document, content, index)
	if !slices.Equal(targets, []string{want}) {
		t.Fatalf("incoming targets = %q, want only %q", targets, want)
	}
	if err := checkTarget(document, targets[0]); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(index, []byte("# Architecture\n## Planning-data routing\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := checkTarget(document, targets[0]); err == nil || !strings.Contains(err.Error(), "missing heading anchor") {
		t.Fatalf("removed forwarding anchor = %v, want missing-heading failure", err)
	}
}

func TestCheckTargetDetectsBrokenRoutes(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "guide.md"), []byte("# Stable contract\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	document := filepath.Join(dir, "guide.md")
	if err := checkTarget(document, "#stable-contract"); err != nil {
		t.Fatal(err)
	}
	for _, target := range []string{"missing.md", "guide.md#removed-contract"} {
		if err := checkTarget(document, target); err == nil {
			t.Errorf("broken route %q passed", target)
		}
	}
}
