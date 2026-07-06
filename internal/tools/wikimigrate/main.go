// Command wikimigrate converts Obsidian-style `[[slug]]` wikilinks in a planning tree
// to GitHub-clickable relative-path markdown links — `[slug](<relative path to the
// target .md>)` — the Scheme-2 reference form (ADR-0003 / epic 24). The flatten
// migration (flatmigrate) deliberately LEFT wikilinks (slug-based, stable); this is the
// separate Scheme-2 step.
//
// A wikilink resolves against an index of every `.md` under root, keyed by its "logical
// name": the human slug for an id-led entity (`tasks`/`audits` `<id>-<slug>.md`) or the
// whole stem for the rest (`epics` `NN-<slug>`, `adrs`, `research`). An unresolved or
// AMBIGUOUS name is left untouched — so `[[…]]` placeholders and dup slugs stay put — and
// reported. Obsidian aliases (`[[target|display]]`) and anchors (`[[target#h]]`) are honored.
//
// Throwaway by design: run on a COPY, verify the diff, commit as one churn commit — git
// is the only undo. DRY-RUN by default; -apply writes and refuses a dirty git tree unless
// -force.
//
//	go run ./internal/tools/wikimigrate -root <planning-dir>          # preview
//	go run ./internal/tools/wikimigrate -root <planning-dir> -apply   # execute
package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	idpkg "github.com/andy-esch/taskflow/internal/id"
)

func main() {
	root := flag.String("root", ".", "planning root (the dir holding tasks/)")
	apply := flag.Bool("apply", false, "write changes (default: dry-run preview)")
	force := flag.Bool("force", false, "skip the clean-git-tree safety check")
	flag.Parse()
	if err := run(*root, *apply, *force); err != nil {
		fmt.Fprintln(os.Stderr, "wikimigrate:", err)
		os.Exit(1)
	}
}

func run(root string, apply, force bool) error {
	root = filepath.Clean(root)
	if fi, err := os.Stat(filepath.Join(root, "tasks")); err != nil || !fi.IsDir() {
		return fmt.Errorf("%s is not a planning root (no tasks/ dir)", root)
	}
	if apply && !force {
		if err := requireCleanTree(root); err != nil {
			return err
		}
	}
	index, ambiguous, err := buildIndex(root)
	if err != nil {
		return err
	}
	edits, unresolved, err := rewrite(root, index, ambiguous)
	if err != nil {
		return err
	}
	report(edits, unresolved, apply)
	if !apply {
		fmt.Printf("\nDRY RUN — nothing written. Re-run with -apply (on a COPY, then commit).\n")
		return nil
	}
	for _, e := range edits {
		if err := os.WriteFile(filepath.Join(root, filepath.FromSlash(e.rel)), e.content, 0o644); err != nil {
			return err
		}
	}
	fmt.Printf("\nAPPLIED. Review `git status` / `git diff`, then commit as one churn commit.\n")
	return nil
}

type fileEdit struct {
	rel     string
	content []byte
	changes int
}

// buildIndex maps every .md file's logical name(s) → its root-relative slash path. An
// id-led entity (tasks/audits) is keyed by BOTH its slug and its full `<id>-<slug>` stem;
// every other file by its stem. A name that maps to two different files is ambiguous
// (recorded so rewrite leaves it alone rather than picking one).
func buildIndex(root string) (index map[string]string, ambiguous map[string]bool, err error) {
	index = map[string]string{}
	ambiguous = map[string]bool{}
	add := func(name, rel string) {
		if name == "" {
			return
		}
		if prev, ok := index[name]; ok && prev != rel {
			ambiguous[name] = true
			return
		}
		index[name] = rel
	}
	err = filepath.WalkDir(root, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if d.Name() == ".git" {
				return filepath.SkipDir
			}
			return nil
		}
		if !isMarkdown(d.Name()) {
			return nil
		}
		rel := filepath.ToSlash(mustRel(root, p))
		stem := strings.TrimSuffix(d.Name(), ".md")
		add(stem, rel)
		if _, slug, ok := splitFlatName(stem); ok {
			add(slug, rel)
		}
		return nil
	})
	return index, ambiguous, err
}

var wikilinkRe = regexp.MustCompile(`\[\[([^\]]+)\]\]`)

// rewrite converts resolvable `[[target]]` / `[[target|display]]` wikilinks to markdown;
// an unresolved or ambiguous target is left as-is and tallied for the report.
func rewrite(root string, index map[string]string, ambiguous map[string]bool) ([]fileEdit, map[string]int, error) {
	var edits []fileEdit
	unresolved := map[string]int{}
	err := filepath.WalkDir(root, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if d.Name() == ".git" {
				return filepath.SkipDir
			}
			return nil
		}
		if !isMarkdown(d.Name()) {
			return nil
		}
		rel := filepath.ToSlash(mustRel(root, p))
		content, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		changes := 0
		out := wikilinkRe.ReplaceAllFunc(content, func(m []byte) []byte {
			inner := string(wikilinkRe.FindSubmatch(m)[1])
			target, display := inner, inner
			if i := strings.IndexByte(inner, '|'); i >= 0 { // Obsidian alias [[target|display]]
				target, display = inner[:i], inner[i+1:]
			}
			anchor := ""
			if i := strings.IndexByte(target, '#'); i >= 0 {
				target, anchor = target[:i], target[i:]
			}
			targetRel, ok := index[target]
			if !ok || ambiguous[target] {
				unresolved[target]++
				return m // leave [[…]] as-is — a placeholder, a dup slug, or a dangling ref
			}
			changes++
			return []byte("[" + display + "](" + relLink(path.Dir(rel), targetRel) + anchor + ")")
		})
		if changes > 0 {
			edits = append(edits, fileEdit{rel: rel, content: out, changes: changes})
		}
		return nil
	})
	return edits, unresolved, err
}

func isMarkdown(name string) bool {
	return strings.HasSuffix(name, ".md") && !strings.HasPrefix(name, ".")
}

// splitFlatName is wikimigrate's copy of the store's parser: an id-led stem is the
// fixed-width id, a `-`, then a non-empty slug. ok=false for anything else (epic/adr/
// research stems, or a stray), which the index then keys by its whole stem.
func splitFlatName(stem string) (id, slug string, ok bool) {
	if len(stem) < idpkg.Length+2 || stem[idpkg.Length] != '-' {
		return "", "", false
	}
	cand := stem[:idpkg.Length]
	if !idpkg.Valid(cand) {
		return "", "", false
	}
	return cand, stem[idpkg.Length+1:], true
}

// relLink is the slash relative path from dir fromDir to file to (both slash-relative to
// the same root) — the form a markdown link uses.
func relLink(fromDir, to string) string {
	var from []string
	if fromDir != "." && fromDir != "" {
		from = strings.Split(fromDir, "/")
	}
	toParts := strings.Split(to, "/")
	i := 0
	for i < len(from) && i < len(toParts)-1 && from[i] == toParts[i] {
		i++
	}
	var rel []string
	for range from[i:] {
		rel = append(rel, "..")
	}
	rel = append(rel, toParts[i:]...)
	return strings.Join(rel, "/")
}

func mustRel(base, p string) string {
	if r, err := filepath.Rel(base, p); err == nil {
		return r
	}
	return p
}

func requireCleanTree(root string) error {
	out, err := exec.Command("git", "-C", root, "status", "--porcelain").Output()
	if err != nil {
		return fmt.Errorf("clean-tree check failed (not a git repo? use -force): %w", err)
	}
	if len(bytes.TrimSpace(out)) != 0 {
		return fmt.Errorf("git tree is dirty — commit or stash first (git is the undo), or pass -force")
	}
	return nil
}

func report(edits []fileEdit, unresolved map[string]int, apply bool) {
	verb := map[bool]string{true: "will", false: "would"}[apply]
	total := 0
	for _, e := range edits {
		total += e.changes
	}
	fmt.Printf("Plan (%s):\n", map[bool]string{true: "APPLY", false: "dry run"}[apply])
	fmt.Printf("  %d wikilink(s) in %d file(s) %s become relative-path markdown links\n", total, len(edits), verb)
	if len(unresolved) > 0 {
		left := 0
		for _, n := range unresolved {
			left += n
		}
		fmt.Printf("  %d wikilink(s) left untouched (no unique target — placeholder, dup slug, or dangler):\n", left)
		for _, name := range sortedKeys(unresolved) {
			fmt.Printf("      [[%s]]  ×%d\n", name, unresolved[name])
		}
	}
}

func sortedKeys(m map[string]int) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
