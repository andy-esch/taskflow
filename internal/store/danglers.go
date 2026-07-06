package store

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/andy-esch/taskflow/internal/domain"
)

// DanglingLinks walks the planning tree and returns every body markdown link
// `[…](<rel-path>.md)` whose target file does not exist — a broken cross-reference (e.g.
// one a rename cascade or a hand-edit missed). Skipped: external links (`://`, `mailto:`)
// and links whose target carries a template placeholder (`…`, `<`, `>`, `{`, `}`), which are
// examples, not real paths. Anchors (`#…`) are stripped before the existence check. It is
// the Scheme-2 dangler check `lint --links` surfaces.
func (s *FS) DanglingLinks() ([]domain.FileProblem, error) {
	var out []domain.FileProblem
	err := filepath.WalkDir(s.root, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if d.Name() == ".git" {
				return filepath.SkipDir
			}
			return nil
		}
		if !markdownDoc(d) {
			return nil
		}
		content, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		dir := filepath.Dir(p)
		for _, m := range mdLinkRe.FindAllSubmatch(content, -1) {
			target := string(m[2])
			link := target
			if i := strings.IndexAny(link, "#?"); i >= 0 { // strip a #fragment or ?query
				link = link[:i]
			}
			if !strings.HasSuffix(link, ".md") { // only local .md cross-references
				continue
			}
			if strings.Contains(link, "://") || strings.HasPrefix(link, "mailto:") {
				continue
			}
			if strings.ContainsAny(link, "…<>{}") { // a template placeholder, not a real path
				continue
			}
			resolved := filepath.Clean(filepath.Join(dir, filepath.FromSlash(link)))
			// Skip a link that escapes the planning root — not a planning cross-reference,
			// and no business stat-ing files outside the tree.
			if rel, err := filepath.Rel(s.root, resolved); err != nil || strings.HasPrefix(rel, "..") {
				continue
			}
			if _, err := os.Stat(resolved); os.IsNotExist(err) {
				out = append(out, domain.FileProblem{Path: p, Message: "body link to missing file: " + target})
			}
		}
		return nil
	})
	return out, err
}
