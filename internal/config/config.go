// Package config locates the planning data within a single repo.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/BurntSushi/toml"

	"github.com/andy-esch/taskflow/internal/domain"
)

// Config records where the planning data lives (the "taskflow root": the dir
// holding tasks/, epics/, ...). One planning repo per product; no cross-product
// registry.
//
// PlanningRepo, when set, points at an EXTERNAL planning repo (the sanctioned
// out-of-tree escape): Discover resolves and validates it to choose Root, and it
// wins over taskflow_root. The raw (unresolved) value is also carried here for
// the linkback checks. TrackedRepos stays RAW and UNRESOLVED — CheckLinks reads
// it. Dir is the (physical) directory the .tskflwctl.toml was found in — the
// anchor planning_repo/tracked_repos resolve against, and this repo's identity
// for linkback. It is empty when discovery fell back to a bare tasks/ dir.
type Config struct {
	Root         string
	Dir          string
	PlanningRepo string
	TrackedRepos []string
	Pager        PagerConfig
	Theme        ThemeConfig
}

// PagerConfig is the `[pager]` table: whether to page long human output and which
// program to use. Enabled is a pointer so "unset" (nil → default on) is distinct
// from an explicit `enabled = false`. A local-terminal concern, so it rides on
// whichever config Discover lands on — not resolved across a planning_repo pointer.
type PagerConfig struct {
	Enabled *bool
	Command string
}

// ThemeConfig is the `[theme]` table: the color theme by name (empty → the
// built-in default). A local-terminal concern like Pager, so it rides on whichever
// config Discover lands on — not resolved across a planning_repo pointer. Plain
// data (a name), no color types, so config stays dependency-light.
type ThemeConfig struct {
	Name string
}

// configFile mirrors the on-disk .tskflwctl.toml schema for a real TOML decode.
// Defaults (taskflow_root ".", the rest empty) are applied by readConfigFile.
type configFile struct {
	TaskflowRoot string        `toml:"taskflow_root"`
	PlanningRepo string        `toml:"planning_repo"`
	TrackedRepos []string      `toml:"tracked_repos"`
	Pager        pagerFileTOML `toml:"pager"`
	Theme        themeFileTOML `toml:"theme"`
}

// pagerFileTOML is the `[pager]` table as decoded from disk.
type pagerFileTOML struct {
	Enabled *bool  `toml:"enabled"`
	Command string `toml:"command"`
}

// themeFileTOML is the `[theme]` table as decoded from disk.
type themeFileTOML struct {
	Name string `toml:"name"`
}

// Discover walks up from start to find the planning root. At each level it
// prefers an explicit `.tskflwctl.toml` marker (honoring its planning_repo if
// set, else taskflow_root), then falls back to a tasks/ directory or a
// planning/tasks/ subdir. It terminates at a .git boundary or the filesystem
// root — never an infinite climb.
func Discover(start string) (*Config, error) {
	dir, err := filepath.Abs(start)
	if err != nil {
		return nil, fmt.Errorf("resolve %q: %w", start, err)
	}
	// Resolve symlinks so the .git-boundary walk-up and the discovered root use
	// PHYSICAL ancestry — a symlinked worktree must climb its real parents, not its
	// logical ones. evalOr falls back to the lexical path when start doesn't exist
	// yet (a fresh dir), so discovery still errs sensibly there.
	dir = evalOr(dir)
	for {
		if fileExists(filepath.Join(dir, ConfigFile)) {
			cf, err := readConfigFile(filepath.Join(dir, ConfigFile))
			if err != nil {
				return nil, err // malformed TOML is loud, never a silent default
			}
			root, err := resolveRoot(dir, cf)
			if err != nil {
				return nil, err // loud, never a silently empty/forked tree
			}
			// resolveRoot already followed/validated planning_repo into Root; the
			// raw planning_repo + tracked_repos ride along for the linkback checks.
			return &Config{
				Root:         root,
				Dir:          dir,
				PlanningRepo: cf.PlanningRepo,
				TrackedRepos: cf.TrackedRepos,
				Pager:        PagerConfig{Enabled: cf.Pager.Enabled, Command: cf.Pager.Command},
				Theme:        ThemeConfig{Name: cf.Theme.Name},
			}, nil
		}
		if isDir(filepath.Join(dir, domain.TasksDir)) {
			return &Config{Root: dir}, nil
		}
		if isDir(filepath.Join(dir, "planning", domain.TasksDir)) {
			return &Config{Root: filepath.Join(dir, "planning")}, nil
		}
		parent := filepath.Dir(dir)
		// .git is the climb boundary whether it's a directory OR a file — in a
		// git worktree/submodule it's a file pointing elsewhere, and missing it
		// would over-climb into a parent's planning tree.
		if parent == dir || exists(filepath.Join(dir, ".git")) {
			return nil, fmt.Errorf(
				"not a taskflow planning repo (no %s or tasks/ found from %s up) — run `tskflwctl init`",
				ConfigFile, start)
		}
		dir = parent
	}
}

// configuredRoot resolves the planning root from dir, honoring the already-parsed
// taskflow_root rel value (default "" → "."). The result must stay within dir's
// tree and must look like a planning root (contain tasks/): a typo'd or
// escaping value previously presented a clean EMPTY project — and `task new`
// would then fork the data into a second tree — so both are loud errors now.
func configuredRoot(dir, rel string) (string, error) {
	if rel == "" {
		rel = "."
	}
	root := filepath.Join(dir, filepath.FromSlash(rel))
	// Containment must be PHYSICAL: filepath.Rel is lexical and can't see that a
	// `planning -> /etc` symlink escapes dir's real tree. Resolve both ends before
	// the no-`..` check (evalOr leaves a not-yet-existing path lexical, which then
	// fails the tasks/-exists check below with a clear message).
	if r, err := filepath.Rel(evalOr(dir), evalOr(root)); err != nil ||
		r == ".." || strings.HasPrefix(r, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("%w: taskflow_root %q escapes %s's directory (%s)", domain.ErrValidation, rel, ConfigFile, dir)
	}
	if !isDir(filepath.Join(root, domain.TasksDir)) {
		return "", fmt.Errorf(
			"%w: taskflow_root %q points at %s, which has no tasks/ — fix %s or run `tskflwctl init`",
			domain.ErrValidation, rel, root, ConfigFile)
	}
	return root, nil
}

// resolveRoot picks the planning root from a parsed config. planning_repo, when
// set, is the sanctioned out-of-tree escape and wins; taskflow_root is the
// in-tree (containment-checked) default. Setting both to different roots is a
// conflict, not a silent override — taskflow_root "" / "." is the only value
// that coexists with planning_repo (it just means "the default in-tree root",
// which planning_repo overrides).
func resolveRoot(dir string, cf configFile) (string, error) {
	if cf.PlanningRepo == "" {
		return configuredRoot(dir, cf.TaskflowRoot)
	}
	// Only a default-equivalent taskflow_root coexists with planning_repo; anything
	// else names a SECOND root, which is a conflict, not a silent override.
	// Normalize first so "./", " . ", "./." — all the in-tree root — aren't
	// mistaken for a different root than ".".
	if tr := strings.TrimSpace(cf.TaskflowRoot); tr != "" && filepath.Clean(filepath.FromSlash(tr)) != "." {
		return "", fmt.Errorf(
			"%w: %s sets both planning_repo (%q) and taskflow_root (%q), which name different roots — keep one",
			domain.ErrConflict, ConfigFile, cf.PlanningRepo, cf.TaskflowRoot)
	}
	return resolvePlanningRepo(dir, cf.PlanningRepo)
}

// resolvePlanningRepo resolves planning_repo relative to the config dir (an
// absolute value is used as-is) and validates it is a real planning root. Unlike
// taskflow_root, it is ALLOWED to escape dir's tree — that is the entire point
// of pointing an impl repo at an external planning repo. A target without tasks/
// is a loud error (the "require + validate" contract), never a clean empty tree.
func resolvePlanningRepo(dir, planningRepo string) (string, error) {
	root := filepath.FromSlash(planningRepo)
	if !filepath.IsAbs(root) {
		root = filepath.Join(dir, root)
	}
	root = filepath.Clean(root)
	if !isDir(filepath.Join(root, domain.TasksDir)) {
		return "", fmt.Errorf(
			"%w: planning_repo %q points at %s, which has no tasks/ — run `tskflwctl init` there first",
			domain.ErrValidation, planningRepo, root)
	}
	// Resolve symlinks so Root is PHYSICAL, matching the in-tree branch (Discover
	// evalOr's dir, which configuredRoot inherits). The linkback work compares
	// physical paths, so a symlinked external planning repo must not leave Root
	// logical here. evalOr falls back to the lexical path if it can't resolve.
	return evalOr(root), nil
}

// configForRoot finds the .tskflwctl.toml that GOVERNS the planning root at
// `root` — i.e. the config whose taskflow_root resolves back to `root` — and
// returns its directory + parsed contents. Usually that's `root` itself
// (taskflow_root "."); for a taskflow_root-subdir layout the config sits in an
// ancestor (e.g. config at repo/, root at repo/planning), so a plain
// root/.tskflwctl.toml lookup would miss it and falsely report a broken link.
// ok is false when no config governs the root (a config-less or unrelated tree).
func configForRoot(root string) (dir string, cf configFile, ok bool) {
	root = evalOr(root)
	for d := root; ; {
		if p := filepath.Join(d, ConfigFile); fileExists(p) {
			c, err := readConfigFile(p)
			if err != nil {
				return "", configFile{}, false
			}
			if r, err := resolveRoot(d, c); err == nil && evalOr(r) == root {
				return d, c, true
			}
			return "", configFile{}, false // a config here governs a different root
		}
		parent := filepath.Dir(d)
		if parent == d || exists(filepath.Join(d, ".git")) {
			return "", configFile{}, false
		}
		d = parent
	}
}

// readConfigFile decodes ConfigFile into a configFile with a real TOML parser.
// A missing file decodes to zero values (the caller already gated on existence,
// but this stays lenient on a vanished file). Malformed TOML is a LOUD error
// wrapping domain.ErrValidation — never a silent default that would present a
// clean empty project. Absent keys keep their zero values: taskflow_root is
// defaulted to "." downstream in configuredRoot; planning_repo/tracked_repos
// stay "" / nil.
func readConfigFile(configPath string) (configFile, error) {
	var cf configFile
	if _, err := toml.DecodeFile(configPath, &cf); err != nil {
		if os.IsNotExist(err) {
			return configFile{}, nil
		}
		return configFile{}, fmt.Errorf("%w: parse %s: %w", domain.ErrValidation, ConfigFile, err)
	}
	return cf, nil
}

// evalOr resolves symlinks in p, falling back to p itself when it can't (e.g. p
// doesn't exist yet) — so containment and boundary checks operate on physical
// paths without breaking on a not-yet-created directory.
func evalOr(p string) string {
	if resolved, err := filepath.EvalSymlinks(p); err == nil {
		return resolved
	}
	return p
}

// exists reports whether path exists as anything (file or directory).
func exists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}

func isDir(p string) bool {
	info, err := os.Stat(p)
	return err == nil && info.IsDir()
}

func fileExists(p string) bool {
	info, err := os.Stat(p)
	return err == nil && !info.IsDir()
}

// ConfigFile is the per-repo config filename written by Init.
const ConfigFile = ".tskflwctl.toml"

// gitKeep is the placeholder Init drops in each scaffolded dir so an empty
// planning tree is still git-committable (git won't track empty directories).
const gitKeep = ".gitkeep"

const defaultConfigTOML = `# tskflwctl planning config — also the marker that anchors discovery.
# taskflow_root: planning dir relative to this file (default ".").
taskflow_root = "."
# tracked_repos: impl repos this planning repo tracks (managed by ` + "`init --track`" + ` /
# the auto-link-back from ` + "`init --planning-repo`" + `).
tracked_repos = []

# [pager]: page long human output (show/schema) through $PAGER on a TTY, like git.
# Never affects piped/--json output. Override the program or turn it off here.
# [pager]
# enabled = true
# command = "less -FRX"

# [theme]: the color theme — a registered name, or "auto" for the default. Override
# here, or with the --theme flag / TSKFLW_THEME env (precedence: flag > env > config).
# An unrecognized name warns (to stderr) and falls back to the default.
# [theme]
# name = "neon"
`

// Init scaffolds the planning directory tree and writes the config file under
// root. It is idempotent: existing dirs/config are left untouched. Returns the
// relative paths created (empty if nothing was needed).
func Init(root string, dryRun bool) ([]string, error) {
	// Refuse to scaffold a local tree over an existing POINTER config: discovery
	// would follow the pointer and the new tree would be orphaned/forked data (the
	// "don't fork the data" non-negotiable). Re-initializing a SCAFFOLD repo is
	// fine — it repairs .gitkeeps — so only a planning_repo config is refused.
	if cfgPath := filepath.Join(root, ConfigFile); fileExists(cfgPath) {
		cf, err := readConfigFile(cfgPath)
		if err != nil {
			return nil, err
		}
		if cf.PlanningRepo != "" {
			return nil, fmt.Errorf(
				"%w: %s points at an external planning repo (planning_repo=%q) — remove it to scaffold a local tree",
				domain.ErrConflict, ConfigFile, cf.PlanningRepo)
		}
	}
	// Flat layout (ADR-0003 §4): scaffold only the entity parents — no per-status or
	// per-bucket subdirs. The flat store never reads them, and a `.md` dropped into one
	// would be invisible to the scan (a silent data-loss trap).
	dirs := []string{domain.TasksDir, domain.EpicsDir, domain.AuditsDir, domain.ProjectsDir}
	var created []string
	for _, d := range dirs {
		p := filepath.Join(root, filepath.FromSlash(d))
		if !isDir(p) {
			if !dryRun {
				if err := os.MkdirAll(p, 0o755); err != nil {
					return created, fmt.Errorf("mkdir %s: %w", p, err)
				}
			}
			created = append(created, d)
		}
		// A .gitkeep makes the (otherwise empty) dir git-committable. Written if
		// absent even when the dir already exists, so re-running init also repairs
		// a tree scaffolded before this existed.
		keep := filepath.Join(p, gitKeep)
		if !fileExists(keep) {
			if !dryRun {
				if err := os.WriteFile(keep, nil, 0o644); err != nil {
					return created, fmt.Errorf("write %s: %w", keep, err)
				}
			}
			created = append(created, d+"/"+gitKeep)
		}
	}
	cfg := filepath.Join(root, ConfigFile)
	// Exclusive create instead of exists-then-write: a concurrent init must not
	// clobber a config the other process just wrote (idempotency falls out of
	// O_EXCL rather than a racy stat).
	if dryRun {
		if !fileExists(cfg) {
			created = append(created, ConfigFile)
		}
		return created, nil
	}
	switch err := writeFileExclusive(cfg, []byte(defaultConfigTOML), 0o644); {
	case err == nil:
		created = append(created, ConfigFile)
	case !os.IsExist(err):
		return created, fmt.Errorf("write config: %w", err)
	}
	return created, nil
}

// pointerConfigTOML renders the POINTER config InitPointer writes: the marker
// that anchors discovery, pointing at an external planning repo instead of
// scaffolding a tree here. The path is stored as the user gave it (relative or
// absolute) so it stays portable; discovery resolves it relative to this file.
// %q yields a valid TOML basic string for the common path cases.
func pointerConfigTOML(planningRepo string) string {
	return fmt.Sprintf("# tskflwctl planning config — this repo's planning lives in another repo.\n"+
		"# planning_repo: the external planning repo, relative to this file (or absolute).\n"+
		"planning_repo = %q\n", planningRepo)
}

// InitPointer writes a POINTER config under dir (`.tskflwctl.toml` with a
// planning_repo) and scaffolds NO tree — for an impl repo whose planning lives
// elsewhere. The target is validated as a real planning root BEFORE anything is
// written (the "require + error" contract via the same resolvePlanningRepo
// discovery uses), so a typo'd path fails loudly with nothing left behind.
// Idempotent via O_EXCL: an existing config is left untouched (returns empty).
func InitPointer(dir, planningRepo string, dryRun bool) ([]string, error) {
	if strings.TrimSpace(planningRepo) == "" {
		return nil, fmt.Errorf("%w: planning_repo path is required", domain.ErrValidation)
	}
	if _, err := resolvePlanningRepo(dir, planningRepo); err != nil {
		return nil, err // not a planning root → loud, nothing written
	}
	cfg := filepath.Join(dir, ConfigFile)
	// An existing config: same target → idempotent no-op; a DIFFERENT target (a
	// re-point) or a scaffold config (a mode switch) → refuse, so a corrected typo
	// or an intended switch isn't silently dropped as "already initialized".
	if fileExists(cfg) {
		existing, err := readConfigFile(cfg)
		if err != nil {
			return nil, err
		}
		if existing.PlanningRepo != planningRepo {
			return nil, fmt.Errorf(
				"%w: %s already exists here — remove it to re-init (or edit it to change the target)",
				domain.ErrConflict, ConfigFile)
		}
		return nil, nil // unchanged
	}
	if dryRun {
		return []string{ConfigFile}, nil
	}
	// Create the target dir if missing — parity with scaffold Init (which MkdirAll's
	// the tree); done only after the target validates, so a bad path leaves nothing.
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("mkdir %s: %w", dir, err)
	}
	if err := writeFileExclusive(cfg, []byte(pointerConfigTOML(planningRepo)), 0o644); err != nil {
		if os.IsExist(err) {
			return nil, nil // O_EXCL race: a config appeared just now — treat as no-op
		}
		return nil, fmt.Errorf("write config: %w", err)
	}
	return []string{ConfigFile}, nil
}

// AddTrackedRepo records repoPath in the tracked_repos of the planning config in
// dir (an impl repo this planning repo tracks). Deduped by PHYSICAL path, so
// "../x", an absolute path, and a symlinked checkout that resolve to the same
// place collapse to one entry. The edit is SURGICAL — only the tracked_repos
// array is rewritten; comments and other keys/order are preserved. A missing
// config is a no-op. Returns whether an entry was actually added.
func AddTrackedRepo(dir, repoPath string, dryRun bool) (bool, error) {
	return appendTrackedRepo(dir, repoPath, dryRun)
}

// LinkBack records the impl repo at implDir in the EXTERNAL planning repo's
// tracked_repos — the reverse of the impl's planning_repo pointer, so both sides
// know about each other. The stored value is the planning→impl relative path. A
// planning repo without a config yet is a SILENT no-op (returns ""), as is an
// already-recorded repo. On a fresh link it returns the back-link path written.
func LinkBack(implDir, planningRepo string, dryRun bool) (string, error) {
	planningRoot, err := resolvePlanningRepo(implDir, planningRepo)
	if err != nil {
		return "", err
	}
	// Write the back-link into the planning repo's CONFIG dir (which a subdir
	// taskflow_root layout puts above the root), so tracked_repos is anchored
	// where Discover/CheckLinks read it. No config there yet → silent no-op.
	pdir, _, ok := configForRoot(planningRoot)
	if !ok {
		return "", nil
	}
	rel, err := filepath.Rel(pdir, evalOr(implDir))
	if err != nil {
		return "", err
	}
	rel = filepath.ToSlash(filepath.Clean(rel))
	added, err := appendTrackedRepo(pdir, rel, dryRun)
	if err != nil || !added {
		return "", err
	}
	return rel, nil
}

// appendTrackedRepo is the shared core: append entry to dir's tracked_repos,
// physical-path deduped, with a surgical (comment-preserving) text edit and an
// atomic write. A missing config is a no-op (so LinkBack can skip an un-init'd
// target); a blank entry is a validation error.
func appendTrackedRepo(dir, entry string, dryRun bool) (bool, error) {
	if strings.TrimSpace(entry) == "" {
		return false, fmt.Errorf("%w: tracked repo path is required", domain.ErrValidation)
	}
	cfgPath := filepath.Join(dir, ConfigFile)
	if !fileExists(cfgPath) {
		return false, nil // nothing to track into
	}
	cf, err := readConfigFile(cfgPath)
	if err != nil {
		return false, err
	}
	target := resolveRepoPath(dir, entry)
	for _, e := range cf.TrackedRepos {
		if resolveRepoPath(dir, e) == target {
			return false, nil // already tracked (same physical path)
		}
	}
	if dryRun {
		return true, nil
	}
	data, err := os.ReadFile(cfgPath)
	if err != nil {
		return false, err
	}
	updated := setTrackedReposInText(string(data), append(cf.TrackedRepos, entry))
	if err := writeFileAtomic(cfgPath, []byte(updated), 0o644); err != nil {
		return false, err
	}
	return true, nil
}

// resolveRepoPath resolves p (relative to dir, or absolute) to a physical path
// for dedup comparison — the evalOr/Abs discipline used throughout discovery.
func resolveRepoPath(dir, p string) string {
	p = filepath.FromSlash(p)
	if !filepath.IsAbs(p) {
		p = filepath.Join(dir, p)
	}
	return evalOr(filepath.Clean(p))
}

// LinkProblem is one linkback inconsistency between an impl repo's planning_repo
// pointer and the planning repo's tracked_repos back-link. Repo is the offending
// repo as spelled in the config; Message is human-readable.
type LinkProblem struct {
	Repo    string
	Message string
}

// CheckLinks audits the bidirectional planning_repo <-> tracked_repos links
// reachable from cfg and returns any inconsistencies (nil when consistent, or
// when there are no links / no config). All comparisons are on PHYSICAL paths, so
// relative, absolute, and symlinked spellings of the same dir never false-
// positive. It assumes a planning repo's config sits at its root (how `init`
// scaffolds them). Read-only; nothing is mutated.
func CheckLinks(cfg *Config) []LinkProblem {
	if cfg == nil || cfg.Dir == "" {
		return nil
	}
	var problems []LinkProblem
	if cfg.PlanningRepo != "" {
		problems = append(problems, checkBackLink(cfg)...)
	}
	for _, tr := range cfg.TrackedRepos {
		if p, bad := checkTrackedRepo(cfg, tr); bad {
			problems = append(problems, p)
		}
	}
	return problems
}

// checkBackLink (impl side): the planning repo this impl points at should list
// this impl in its tracked_repos.
func checkBackLink(cfg *Config) []LinkProblem {
	me := resolveRepoPath(cfg.Dir, ".")
	planningRoot := resolveRepoPath(cfg.Dir, cfg.PlanningRepo)
	// Find the planning repo's config wherever it lives (root or a taskflow_root
	// ancestor) — NOT just at the root, which a subdir layout would miss.
	pdir, pcf, ok := configForRoot(planningRoot)
	if !ok {
		return []LinkProblem{{cfg.PlanningRepo, fmt.Sprintf(
			"planning repo %q has no %s, so it can't track this repo back", cfg.PlanningRepo, ConfigFile)}}
	}
	for _, tr := range pcf.TrackedRepos {
		if resolveRepoPath(pdir, tr) == me {
			return nil // back-link present
		}
	}
	return []LinkProblem{{cfg.PlanningRepo, fmt.Sprintf(
		"one-sided link: planning repo %q does not track this repo back (run `init --track` there, or re-run `init --planning-repo`)", cfg.PlanningRepo)}}
}

// checkTrackedRepo (planning side): a tracked impl should exist and point its
// planning_repo back here.
func checkTrackedRepo(cfg *Config, tr string) (LinkProblem, bool) {
	implDir := resolveRepoPath(cfg.Dir, tr)
	if !isDir(implDir) {
		return LinkProblem{tr, fmt.Sprintf("tracked repo %q does not exist", tr)}, true
	}
	icfPath := filepath.Join(implDir, ConfigFile)
	if !fileExists(icfPath) {
		return LinkProblem{tr, fmt.Sprintf("tracked repo %q has no %s (it can't point back)", tr, ConfigFile)}, true
	}
	icf, err := readConfigFile(icfPath)
	if err != nil {
		return LinkProblem{tr, fmt.Sprintf("tracked repo %q has an unreadable %s: %v", tr, ConfigFile, err)}, true
	}
	if icf.PlanningRepo == "" {
		return LinkProblem{tr, fmt.Sprintf("tracked repo %q does not point back (no planning_repo)", tr)}, true
	}
	// The impl's planning_repo may name EITHER the planning repo (cfg.Dir — where the
	// .tskflwctl.toml lives) or the taskflow_root subdir it resolves to (cfg.Root); both
	// unambiguously identify this planning repo. Accept either. Checking only one falsely
	// reports "points elsewhere" for the other convention once taskflow_root is set (then
	// Root != Dir: only-Root missed a repo-root pointer — the desirelines case — and
	// only-Dir would break an impl that points at the subdir).
	target := resolveRepoPath(implDir, icf.PlanningRepo)
	if target != resolveRepoPath(cfg.Dir, ".") && target != resolveRepoPath(cfg.Root, ".") {
		return LinkProblem{tr, fmt.Sprintf("tracked repo %q points its planning_repo elsewhere, not here", tr)}, true
	}
	return LinkProblem{}, false
}

// trackedReposKeyRe matches the START of a top-level tracked_repos array
// assignment — at the BEGINNING of a line (so the same text inside a comment or
// another string is never matched), through the opening `[`. The closing `]` is
// found separately by a string-aware scan, because a `]` can legally appear
// inside a quoted path value.
var trackedReposKeyRe = regexp.MustCompile(`(?m)^[ \t]*tracked_repos[ \t]*=[ \t]*\[`)

// setTrackedReposInText surgically rewrites (or appends) the tracked_repos array,
// leaving every other line — comments, key order, other values/arrays — intact.
// It edits exactly ONE assignment (the first top-level one) and locates its
// closing `]` with a quote-aware scan, so neither a `]` inside a path value nor a
// bracketed comment can derail it.
func setTrackedReposInText(text string, repos []string) string {
	assignment := "tracked_repos = " + tomlStringArray(repos)
	if start, end, ok := trackedReposSpan(text); ok {
		return text[:start] + assignment + text[end:]
	}
	if text != "" && !strings.HasSuffix(text, "\n") {
		text += "\n"
	}
	return text + assignment + "\n"
}

// trackedReposSpan returns the byte span [start,end) of the first top-level
// tracked_repos array assignment — from the `tracked_repos` key through its
// matching `]`. ok is false when there's no such assignment, or the array is
// unterminated (caller then leaves the file untouched rather than risk corruption).
func trackedReposSpan(text string) (start, end int, ok bool) {
	loc := trackedReposKeyRe.FindStringIndex(text)
	if loc == nil {
		return 0, 0, false
	}
	// Advance past any leading whitespace so the rewrite preserves indentation.
	start = loc[0]
	for start < loc[1] && (text[start] == ' ' || text[start] == '\t') {
		start++
	}
	// loc[1] is just past the opening '['. Scan to the matching ']', skipping over
	// basic ("...") and literal ('...') strings so a bracket in a value is ignored.
	for i := loc[1]; i < len(text); {
		switch text[i] {
		case ']':
			return start, i + 1, true
		case '"':
			i = skipTOMLString(text, i+1, '"', true)
		case '\'':
			i = skipTOMLString(text, i+1, '\'', false)
		default:
			i++
		}
	}
	return 0, 0, false // unterminated array
}

// skipTOMLString returns the index just past the closing quote, given i is the
// first byte AFTER the opening quote. Basic strings ('escapes' true) honor a
// backslash escape; literal strings have none.
func skipTOMLString(text string, i int, quote byte, escapes bool) int {
	for i < len(text) {
		if escapes && text[i] == '\\' {
			i += 2
			continue
		}
		if text[i] == quote {
			return i + 1
		}
		i++
	}
	return i // unterminated
}

// tomlStringArray renders paths as a TOML inline array of basic strings.
func tomlStringArray(items []string) string {
	if len(items) == 0 {
		return "[]"
	}
	quoted := make([]string, len(items))
	for i, s := range items {
		quoted[i] = fmt.Sprintf("%q", s)
	}
	return "[" + strings.Join(quoted, ", ") + "]"
}

// writeFileAtomic overwrites path atomically (temp in the same dir + rename), so a
// crash mid-write never leaves a truncated config. (The store has the same idea;
// config can't import store, so the few lines are inlined — cf. writeFileExclusive.)
func writeFileAtomic(path string, data []byte, perm os.FileMode) error {
	tmp, err := os.CreateTemp(filepath.Dir(path), ".tskflwctl-*.tmp")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer func() { _ = os.Remove(tmpName) }() // no-op once renamed
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Chmod(perm); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpName, path)
}

// writeFileExclusive creates a new file with O_EXCL semantics. (The store's
// createFileAtomic is the same idea; config can't import store without
// inverting the dependency, so the few lines are inlined.)
func writeFileExclusive(path string, data []byte, perm os.FileMode) error {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, perm)
	if err != nil {
		return err // os.IsExist(err) when the config already exists
	}
	if _, err := f.Write(data); err != nil {
		_ = f.Close()
		_ = os.Remove(path)
		return err
	}
	return f.Close()
}
