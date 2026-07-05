// Command flatmigrate performs the one-time, ADR-0003 §6 cutover of a planning
// tree to the flat, id-led layout:
//
//   - tasks/<status>/<slug>.md   -> tasks/<id>-<slug>.md
//   - audits/<bucket>/<slug>.md  -> audits/<id>-<slug>.md
//
// It is SELF-CONTAINED (it must be — the now-flat store can no longer read a
// dir-based tree, so a `lint --fix` pre-pass is impossible): it mints a missing
// frontmatter id from the file's own date (id.NewAt, ADR-0003 §3 policy; errors if
// there is no date to mint from) and injects a missing status/bucket from the
// subdirectory it is leaving — the last moment that dir carries authority. Epics
// (NN-<slug>, already flat) are untouched. It rewrites entity↔entity relative-path
// body links to the new flat paths (wikilinks are slug-based and stable, so they
// are left alone — converting them is the separate scheme-2 step), sweeps loose
// non-entity .md files at a scanned-dir root into meta/, and removes the emptied
// status/bucket subdirs.
//
// Throwaway by design (NOT a `tskflwctl` command): run on a COPY, verify the diff,
// then commit as one churn commit — git is the only undo. DRY-RUN by default; pass
// -apply to write, and it refuses a dirty git tree unless -force.
//
//	go run ./internal/tools/flatmigrate -root <planning-dir>          # preview
//	go run ./internal/tools/flatmigrate -root <planning-dir> -apply   # execute
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
	"time"

	"github.com/andy-esch/taskflow/internal/id"
)

func main() {
	root := flag.String("root", ".", "planning root (the dir holding tasks/)")
	apply := flag.Bool("apply", false, "write changes (default: dry-run preview)")
	force := flag.Bool("force", false, "skip the clean-git-tree safety check")
	flag.Parse()
	if err := run(*root, *apply, *force); err != nil {
		fmt.Fprintln(os.Stderr, "flatmigrate:", err)
		os.Exit(1)
	}
}

// flatDirs are the entity kinds that flatten (epics already are); each holds files
// one level down in a status/bucket subdir today.
var flatDirs = []struct{ dir, field string }{{"tasks", "status"}, {"audits", "bucket"}}

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

	moves := map[string]string{} // OLD slash-rel path -> NEW slash-rel path (renames + sweeps)
	pre := map[string][]byte{}   // OLD path -> content with fields injected (before link rewrite)
	seenID := map[string]bool{}
	var renames, sweeps []move
	minted, backfilled := 0, 0

	// 1. Entity renames + field injection.
	for _, kind := range flatDirs {
		kindDir := filepath.Join(root, kind.dir)
		subdirs, err := os.ReadDir(kindDir)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return err
		}
		for _, sd := range subdirs {
			if !sd.IsDir() { // a loose .md at the kind root (e.g. audits/HOWTO-execute.md)
				if isMarkdown(sd.Name()) && sd.Name() != "README.md" {
					old := path.Join(kind.dir, sd.Name())
					dst := path.Join("meta", sd.Name())
					moves[old] = dst
					sweeps = append(sweeps, move{old, dst, "loose"})
				}
				continue
			}
			files, err := os.ReadDir(filepath.Join(kindDir, sd.Name()))
			if err != nil {
				return err
			}
			for _, f := range files {
				if !isMarkdown(f.Name()) {
					continue
				}
				oldRel := path.Join(kind.dir, sd.Name(), f.Name())
				content, err := os.ReadFile(filepath.Join(kindDir, sd.Name(), f.Name()))
				if err != nil {
					return err
				}
				id, out, m, b, err := ensureFields(content, kind.field, sd.Name(), seenID)
				if err != nil {
					return fmt.Errorf("%s: %w", oldRel, err)
				}
				minted += m
				backfilled += b
				if out != nil {
					pre[oldRel] = out
				}
				newRel := path.Join(kind.dir, id+"-"+f.Name())
				moves[oldRel] = newRel
				renames = append(renames, move{oldRel, newRel, sd.Name()})
			}
		}
	}

	// 2. Uniqueness guard (external review #5): two files mapping to the same target,
	// or the same id on different slugs (ambiguous by id-prefix), is a hazard the
	// operator resolves — never a silent clobber.
	if err := checkCollisions(renames); err != nil {
		return err
	}

	// 3. Rewrite entity↔entity relative-path links across every .md (a link's target
	// and/or its own file may have moved).
	edits, err := rewriteLinks(root, moves, pre)
	if err != nil {
		return err
	}

	report(renames, sweeps, edits, minted, backfilled, apply)
	if !apply {
		fmt.Printf("\nDRY RUN — nothing written. Re-run with -apply (on a COPY, then commit).\n")
		return nil
	}
	if err := applyPlan(root, edits); err != nil {
		return err
	}
	for _, kind := range flatDirs {
		removeEmptySubdirs(filepath.Join(root, kind.dir))
	}
	fmt.Printf("\nAPPLIED. Review `git status` / `git diff`, then commit as one churn commit.\n")
	return nil
}

type move struct{ old, new, from string }

// fileEdit is the final state of one file: its (possibly new) path and content.
type fileEdit struct {
	oldRel, newRel string
	content        []byte
	linkChanges    int
}

func isMarkdown(name string) bool {
	return strings.HasSuffix(name, ".md") && !strings.HasPrefix(name, ".")
}

// ensureFields returns the file's id (minting one from its date if the frontmatter
// has none) and, if the id or the status/bucket field had to be injected, the
// rewritten content (else out is nil). m/b count a minted id / a backfilled
// status-or-bucket for the report.
func ensureFields(content []byte, field, subdir string, seen map[string]bool) (eid string, out []byte, m, b int, err error) {
	inject := map[string]string{}
	eid = frontmatterField(content, "id")
	if eid == "" {
		millis, ok := firstDate(content)
		if !ok {
			return "", nil, 0, 0, fmt.Errorf("no frontmatter id and no date to mint one from — add a `created: YYYY-MM-DD`")
		}
		eid = mintUnique(millis, seen)
		inject["id"] = eid
		m = 1
	}
	if !id.Valid(eid) {
		return "", nil, 0, 0, fmt.Errorf("frontmatter id %q is not a valid 12-char id", eid)
	}
	seen[eid] = true
	if frontmatterField(content, field) == "" {
		inject[field] = subdir
		b = 1
	}
	if len(inject) == 0 {
		return eid, nil, 0, 0, nil
	}
	return eid, injectFields(content, inject), m, b, nil
}

// injectFields inserts the given keys (in a stable id,status,bucket order) right
// after the opening `---` fence, skipping keys the block already has.
func injectFields(content []byte, fields map[string]string) []byte {
	var b bytes.Buffer
	b.WriteString("---\n")
	for _, k := range []string{"id", "status", "bucket"} {
		if v, ok := fields[k]; ok {
			fmt.Fprintf(&b, "%s: %s\n", k, v)
		}
	}
	b.Write(content[len("---\n"):])
	return b.Bytes()
}

func frontmatterField(content []byte, key string) string {
	if !bytes.HasPrefix(content, []byte("---\n")) {
		return ""
	}
	body := content[4:]
	if end := bytes.Index(body, []byte("\n---")); end >= 0 {
		body = body[:end]
	}
	for _, line := range strings.Split(string(body), "\n") {
		if v, ok := strings.CutPrefix(strings.TrimSpace(line), key+":"); ok {
			return strings.Trim(strings.TrimSpace(v), `"'`)
		}
	}
	return ""
}

var dateRe = regexp.MustCompile(`\d{4}-\d{2}-\d{2}`)

func firstDate(content []byte) (int64, bool) {
	for _, key := range []string{"created", "date", "completed_at", "deprecated_at", "deferred_at", "started_at", "updated_at"} {
		if m := dateRe.FindString(frontmatterField(content, key)); m != "" {
			if t, err := time.Parse("2006-01-02", m); err == nil {
				return t.UnixMilli(), true
			}
		}
	}
	return 0, false
}

// mintUnique mints a date-stamped id, regenerating on the astronomically-rare
// same-run collision (id.NewAt's low bits are random).
func mintUnique(millis int64, seen map[string]bool) string {
	for {
		got := id.NewAt(millis)
		if !seen[got] {
			return got
		}
	}
}

func checkCollisions(renames []move) error {
	byNew, byID := map[string]string{}, map[string]string{}
	for _, r := range renames {
		if prev, ok := byNew[r.new]; ok {
			return fmt.Errorf("collision: %q and %q both map to %s", prev, r.old, r.new)
		}
		byNew[r.new] = r.old
		id := strings.SplitN(path.Base(r.new), "-", 2)[0]
		if prev, ok := byID[id]; ok {
			return fmt.Errorf("duplicate id %s on %q and %q — resolve before migrating", id, prev, r.old)
		}
		byID[id] = r.old
	}
	return nil
}

var linkRe = regexp.MustCompile(`\]\(([^)]+)\)`)

func rewriteLinks(root string, moves map[string]string, pre map[string][]byte) ([]fileEdit, error) {
	var edits []fileEdit
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
		newRel := rel
		if mv, ok := moves[rel]; ok {
			newRel = mv
		}
		content := pre[rel]
		if content == nil {
			if content, err = os.ReadFile(p); err != nil {
				return err
			}
		}
		changes := 0
		out := linkRe.ReplaceAllFunc(content, func(match []byte) []byte {
			target := string(linkRe.FindSubmatch(match)[1])
			link, anchor := target, ""
			if i := strings.IndexByte(link, '#'); i >= 0 {
				link, anchor = link[:i], link[i:]
			}
			if link == "" || !strings.HasSuffix(link, ".md") || isExternal(link) || path.IsAbs(link) {
				return match
			}
			targetOld := path.Clean(path.Join(path.Dir(rel), link))
			targetNew, moved := moves[targetOld]
			if !moved {
				return match
			}
			changes++
			return []byte("](" + relLink(path.Dir(newRel), targetNew) + anchor + ")")
		})
		if changes > 0 || newRel != rel {
			edits = append(edits, fileEdit{oldRel: rel, newRel: newRel, content: out, linkChanges: changes})
		}
		return nil
	})
	return edits, err
}

func isExternal(link string) bool {
	return strings.Contains(link, "://") || strings.HasPrefix(link, "mailto:")
}

// relLink is the slash relative path from dir fromDir to file to (both
// slash-relative to the same root) — the form a markdown link uses.
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

func applyPlan(root string, edits []fileEdit) error {
	for _, e := range edits {
		to := filepath.Join(root, filepath.FromSlash(e.newRel))
		if err := os.MkdirAll(filepath.Dir(to), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(to, e.content, 0o644); err != nil {
			return err
		}
		if e.newRel != e.oldRel {
			if err := os.Remove(filepath.Join(root, filepath.FromSlash(e.oldRel))); err != nil {
				return fmt.Errorf("remove old %s: %w", e.oldRel, err)
			}
		}
	}
	return nil
}

func removeEmptySubdirs(dir string) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		sub := filepath.Join(dir, e.Name())
		_ = os.Remove(filepath.Join(sub, ".gitkeep"))
		_ = os.Remove(sub) // no-op unless now empty
	}
}

func report(renames, sweeps []move, edits []fileEdit, minted, backfilled int, apply bool) {
	verb := map[bool]string{true: "will", false: "would"}[apply]
	fmt.Printf("Plan (%s):\n", map[bool]string{true: "APPLY", false: "dry run"}[apply])
	fmt.Printf("  %d task/audit files %s flatten to <id>-<slug>.md\n", len(renames), verb)
	byFrom := map[string]int{}
	for _, r := range renames {
		byFrom[r.from]++
	}
	for _, k := range sortedKeys(byFrom) {
		fmt.Printf("      %-16s %d\n", k+"/", byFrom[k])
	}
	if minted > 0 || backfilled > 0 {
		fmt.Printf("  %d missing id(s) minted from date; %d status/bucket backfilled from the subdir\n", minted, backfilled)
	}
	linkFiles, linkCount := 0, 0
	for _, e := range edits {
		if e.linkChanges > 0 {
			linkFiles++
			linkCount += e.linkChanges
		}
	}
	fmt.Printf("  %d relative-path link(s) in %d file(s) %s be rewritten to flat paths\n", linkCount, linkFiles, verb)
	fmt.Printf("  %d loose file(s) %s be swept into meta/\n", len(sweeps), verb)
	for _, s := range sweeps {
		fmt.Printf("      %s -> %s\n", s.old, s.new)
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
