// Command researchmigrate performs the one-time cutover of `research/` to the flat,
// id-led layout research docs now use as a first-class entity (epic 28):
//
//	research/2026-06-28-color-palette.md -> research/<id>-color-palette.md
//	research/hybrid-search-architecture.md -> research/<id>-hybrid-search-architecture.md
//
// The id is BACKDATED from each doc's own date via id.NewAt (ADR-0003 §3, the same
// policy flatmigrate used), so lexical id order stays authorship order. The date is
// recovered from, in precedence: the frontmatter `created:`, the `YYYY-MM-DD-` filename
// prefix, or a prose `**Created**: YYYY-MM-DD` header. A doc with no date at all is an
// error, not a guess — the operator adds one and re-runs.
//
// It also backfills the frontmatter contract (schema, id, created, and empty
// description/tags slots to fill in). Docs with NO frontmatter at all get a block
// created; docs that have one are edited surgically, so unknown keys survive —
// including the vestigial `status: reference` the legacy corpus carries, which is
// deliberately NOT part of the research contract and is left to ride along. Prose
// headers (`**Status**: Proposal`, `**Created**: …`) are left in the body untouched
// rather than forced into fields.
//
// Inbound relative-path markdown links across the whole tree are repointed to the new
// filenames. Already-id-led docs are left alone, so re-running is a no-op.
//
// Throwaway by design (NOT a `tskflwctl` command): run on a COPY or a clean tree,
// verify the diff, then commit as one churn commit — git is the only undo. DRY-RUN by
// default; pass -apply to write, and it refuses a dirty git tree unless -force.
//
//	go run ./internal/tools/researchmigrate -root <planning-dir>          # preview
//	go run ./internal/tools/researchmigrate -root <planning-dir> -apply   # execute
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

	"github.com/andy-esch/taskflow/internal/domain"
	idpkg "github.com/andy-esch/taskflow/internal/id"
)

func main() {
	root := flag.String("root", ".", "planning root (the dir holding research/)")
	apply := flag.Bool("apply", false, "write changes (default: dry-run preview)")
	force := flag.Bool("force", false, "skip the clean-git-tree safety check")
	flag.Parse()
	if err := run(*root, *apply, *force); err != nil {
		fmt.Fprintln(os.Stderr, "researchmigrate:", err)
		os.Exit(1)
	}
}

type move struct{ old, new string }

// fileEdit is the final state of one file: its (possibly new) path and content.
type fileEdit struct {
	oldRel, newRel string
	content        []byte
	linkChanges    int
}

func run(root string, apply, force bool) error {
	root = filepath.Clean(root)
	researchDir := filepath.Join(root, domain.ResearchDir)
	if fi, err := os.Stat(researchDir); err != nil || !fi.IsDir() {
		return fmt.Errorf("%s has no %s/ dir", root, domain.ResearchDir)
	}
	if apply && !force {
		if err := requireCleanTree(root); err != nil {
			return err
		}
	}

	entries, err := os.ReadDir(researchDir)
	if err != nil {
		return err
	}
	moves := map[string]string{} // OLD slash-rel path -> NEW slash-rel path
	pre := map[string][]byte{}   // OLD path -> content with frontmatter backfilled
	seenID := map[string]bool{}
	var renames []move
	var skipped []string
	created, backfilled := 0, 0

	// Sorted for a deterministic plan (ReadDir order is already sorted, but the
	// same-millisecond id dedupe below makes the ORDER matter for reproducibility).
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() || !isMarkdown(e.Name()) || strings.EqualFold(e.Name(), "README.md") {
			continue
		}
		names = append(names, e.Name())
	}
	sort.Strings(names)

	for _, name := range names {
		stem := strings.TrimSuffix(name, ".md")
		if isIDLed(stem) { // already migrated — re-running is a no-op
			skipped = append(skipped, name)
			continue
		}
		content, err := os.ReadFile(filepath.Join(researchDir, name))
		if err != nil {
			return err
		}
		date, ok := resolveDate(content, stem)
		if !ok {
			return fmt.Errorf("%s: no date found — add a `created: YYYY-MM-DD` (checked frontmatter created:, the filename prefix, and a prose **Created**: header)", name)
		}
		millis, err := dateMillis(date)
		if err != nil {
			return fmt.Errorf("%s: %w", name, err)
		}
		eid := mintUnique(millis, seenID)
		out, c, b := ensureFrontmatter(content, eid, date)
		if c {
			created++
		}
		if b {
			backfilled++
		}
		oldRel := path.Join(domain.ResearchDir, name)
		newRel := path.Join(domain.ResearchDir, eid+"-"+slugOf(stem)+".md")
		pre[oldRel] = out
		moves[oldRel] = newRel
		renames = append(renames, move{oldRel, newRel})
	}

	// Two docs mapping to the same target, or a duplicate id, is a hazard the operator
	// resolves — never a silent clobber.
	if err := checkCollisions(renames); err != nil {
		return err
	}

	edits, err := rewriteLinks(root, moves, pre)
	if err != nil {
		return err
	}

	report(renames, skipped, edits, created, backfilled, apply)
	if !apply {
		fmt.Printf("\nDRY RUN — nothing written. Re-run with -apply (on a clean tree, then commit).\n")
		return nil
	}
	if err := applyPlan(root, edits); err != nil {
		return err
	}
	fmt.Printf("\nAPPLIED. Review `git status` / `git diff`, then commit as one churn commit.\n")
	return nil
}

// isIDLed reports whether stem is already `<12-char-id>-<slug>` — mirrors the store's
// splitFlatName, so an already-migrated doc is skipped and re-running is idempotent.
func isIDLed(stem string) bool {
	if len(stem) < idpkg.Length+2 || stem[idpkg.Length] != '-' {
		return false
	}
	return idpkg.Valid(stem[:idpkg.Length])
}

var datePrefixRe = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}-`)

// slugOf strips a `YYYY-MM-DD-` filename prefix: the date moves into frontmatter
// (`created:`) and the id now carries the ordering, so repeating it in the name would
// be a third copy. A non-date-led stem is already the slug.
func slugOf(stem string) string {
	return datePrefixRe.ReplaceAllString(stem, "")
}

var (
	dateRe      = regexp.MustCompile(`\d{4}-\d{2}-\d{2}`)
	proseDateRe = regexp.MustCompile(`\*\*(?:Created|Date)\*\*:?[^0-9\n]*(\d{4}-\d{2}-\d{2})`)
)

// resolveDate recovers a doc's authorship date, in precedence order: the frontmatter
// `created:` (the declared value wins), the `YYYY-MM-DD-` filename prefix, then a prose
// `**Created**:`/`**Date**:` header (how the pre-frontmatter corpus recorded it).
//
// Deliberately NOT a fallback: git's add-date. For the legacy batch it is the date of
// the repo import, not of the work — it would silently claim a wrong chronology, which
// is worse than the error this returns.
func resolveDate(content []byte, stem string) (string, bool) {
	if v := frontmatterField(content, "created"); v != "" {
		if m := dateRe.FindString(v); m != "" {
			return m, true
		}
	}
	if m := datePrefixRe.FindString(stem); m != "" {
		return strings.TrimSuffix(m, "-"), true
	}
	if m := proseDateRe.FindSubmatch(content); m != nil {
		return string(m[1]), true
	}
	return "", false
}

// dateMillis is the id's timestamp: a date-only string parses to UTC midnight, which
// is exactly what flatmigrate minted from, so backdated ids stay consistent with the
// ids already in the tree.
func dateMillis(date string) (int64, error) {
	t, err := time.Parse("2006-01-02", date)
	if err != nil {
		return 0, fmt.Errorf("unparseable date %q: %w", date, err)
	}
	return t.UnixMilli(), nil
}

// mintUnique mints a date-stamped id, regenerating on a same-millisecond collision
// (id.NewAt's low bits are random). Collisions are COMMON here, not rare: dates are
// day-precision and the corpus clusters heavily (9 docs share 2026-01-03), so every
// same-day doc mints into the same millisecond slot. Their relative order is therefore
// arbitrary — accepted by decision (epic 28, 2026-08-14): order matters at the date
// level, not within a date.
func mintUnique(millis int64, seen map[string]bool) string {
	for {
		got := idpkg.NewAt(millis)
		if !seen[got] {
			seen[got] = true
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
		eid := strings.SplitN(path.Base(r.new), "-", 2)[0]
		if prev, ok := byID[eid]; ok {
			return fmt.Errorf("duplicate id %s on %q and %q — resolve before migrating", eid, prev, r.old)
		}
		byID[eid] = r.old
	}
	return nil
}

// ensureFrontmatter backfills the research contract. createdBlock reports that the doc
// had NO frontmatter and a block was created (the 10 bare legacy docs); backfilled
// reports that an existing block gained keys.
//
// An existing block is edited surgically — keys are inserted right after the opening
// fence and everything else (unknown keys like the vestigial `status: reference`,
// comments, key order) is preserved byte-for-byte.
func ensureFrontmatter(content []byte, eid, date string) (out []byte, createdBlock, backfilled bool) {
	if !bytes.HasPrefix(content, []byte("---\n")) {
		// No frontmatter: create the block. The body (including any prose **Status**: /
		// **Created**: headers) is kept verbatim below it — prose stays prose.
		var b bytes.Buffer
		b.WriteString("---\n")
		writeContractFields(&b, eid, date, true, true)
		b.WriteString("---\n\n")
		b.Write(bytes.TrimLeft(content, "\n"))
		return b.Bytes(), true, false
	}
	var b bytes.Buffer
	b.WriteString("---\n")
	writeContractFields(&b, eid, date,
		frontmatterField(content, "description") == "",
		frontmatterField(content, "tags") == "")
	rest := content[len("---\n"):]
	// Drop an existing `created:`/`schema:`/`id:` line: the value we just wrote is the
	// same one (created was read FROM it), so keeping both would duplicate the key.
	rest = dropFrontmatterKeys(rest, "schema", "id", "created")
	b.Write(rest)
	return b.Bytes(), false, true
}

// writeContractFields writes the contract keys in researchFields order. description and
// tags are written as empty slots only when the doc has none — a visible place to fill
// in, rather than a value invented for it.
func writeContractFields(b *bytes.Buffer, eid, date string, wantDesc, wantTags bool) {
	fmt.Fprintf(b, "schema: %d\n", domain.FileSchemaVersion)
	fmt.Fprintf(b, "id: %s\n", eid)
	fmt.Fprintf(b, "created: %q\n", date)
	if wantDesc {
		b.WriteString("description: \"\"\n")
	}
	if wantTags {
		b.WriteString("tags: []\n")
	}
}

// dropFrontmatterKeys removes top-level `key:` lines from the frontmatter portion of
// rest (everything up to the closing `---`), leaving the body untouched. Only exact
// top-level keys match, so an indented list item or body prose is never touched.
func dropFrontmatterKeys(rest []byte, keys ...string) []byte {
	end := bytes.Index(rest, []byte("\n---"))
	if end < 0 {
		return rest
	}
	fm, body := string(rest[:end]), rest[end:]
	drop := map[string]bool{}
	for _, k := range keys {
		drop[k] = true
	}
	var kept []string
	for _, line := range strings.Split(fm, "\n") {
		key, _, found := strings.Cut(line, ":")
		if found && drop[strings.TrimSpace(key)] && key == strings.TrimRight(key, " \t") && !strings.HasPrefix(line, " ") {
			continue
		}
		kept = append(kept, line)
	}
	return append([]byte(strings.Join(kept, "\n")), body...)
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

func isMarkdown(name string) bool {
	return strings.HasSuffix(name, ".md") && !strings.HasPrefix(name, ".")
}

var linkRe = regexp.MustCompile(`\]\(([^)]+)\)`)

// rewriteLinks repoints every inbound relative-path markdown link across the tree to
// the renamed research files, and writes out the frontmatter-backfilled content.
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

// relLink is the slash relative path from fromDir to file to (both slash-relative to
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

func mustRel(root, p string) string {
	rel, err := filepath.Rel(root, p)
	if err != nil {
		return p
	}
	return rel
}

func report(renames []move, skipped []string, edits []fileEdit, created, backfilled int, apply bool) {
	verb := map[bool]string{true: "will", false: "would"}[apply]
	fmt.Printf("Plan (%s):\n", map[bool]string{true: "APPLY", false: "dry run"}[apply])
	fmt.Printf("  %d research doc(s) %s become <id>-<slug>.md\n", len(renames), verb)
	if created > 0 {
		fmt.Printf("      %d had NO frontmatter — a block %s be created (prose headers left in the body)\n", created, verb)
	}
	if backfilled > 0 {
		fmt.Printf("      %d had frontmatter — contract keys %s be backfilled surgically (unknown keys preserved)\n", backfilled, verb)
	}
	if len(skipped) > 0 {
		fmt.Printf("  %d already id-led, skipped\n", len(skipped))
	}
	linkFiles, linkCount := 0, 0
	for _, e := range edits {
		if e.linkChanges > 0 {
			linkFiles++
			linkCount += e.linkChanges
		}
	}
	fmt.Printf("  %d inbound link(s) across %d file(s) %s be repointed\n", linkCount, linkFiles, verb)
	fmt.Println("\nRenames:")
	for _, r := range renames {
		fmt.Printf("  %s\n    -> %s\n", r.old, r.new)
	}
}

// requireCleanTree refuses to write into a dirty git tree: git is the only undo for
// this migration, so an operator must be able to `git checkout .` back.
func requireCleanTree(root string) error {
	out, err := exec.Command("git", "-C", root, "status", "--porcelain").Output()
	if err != nil {
		return fmt.Errorf("git status failed (not a git repo? use -force to skip): %w", err)
	}
	if len(bytes.TrimSpace(out)) != 0 {
		return fmt.Errorf("git tree is dirty — commit or stash first (git is the only undo), or pass -force:\n%s", out)
	}
	return nil
}
