package userconfig

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
)

// The space registry: which local entry points into planning spaces this machine knows
// about. Several rows may share one durable planning identity; grouping and direct/pointer
// roles are derived by spacehealth rather than persisted here.
//
// It lives in its OWN file, separate from the user-preference config.toml, because registry
// mutations should never risk a person's presentation settings. The registry is still
// inspectable and hand-editable: `space add` appends one table and `space forget` removes
// one table, leaving comments, key order, unknown keys, and every other entry byte-for-byte
// intact. Writes are atomic.
//
// The registry is ADVISORY: ordinary cwd discovery never consults it. Only an explicit
// primary-adapter selection such as CLI --space opts into a registered entry point. A
// machine with no spaces.toml therefore behaves exactly as it did before one existed,
// and deleting the file costs convenience, never data. (internal/config does not import
// this package, so ambient discovery independence is enforced rather than remembered.)

// SpacesFile is the registry filename, beside config.toml.
const SpacesFile = "spaces.toml"

// spacesSchemaVersion is the on-disk shape of spaces.toml, independent of the --json
// contract's schema_version. Bumped only if the entry shape changes incompatibly.
const spacesSchemaVersion = 1

// Registry errors are typed so the CLI can distinguish user-correctable registry
// content/collisions from filesystem failures. This package stays independent of the CLI's
// domain exit-code vocabulary; the adapter maps these sentinels at its boundary.
var (
	ErrInvalidRegistry = errors.New("invalid space registry")
	ErrSpaceIDConflict = errors.New("space label conflict")
)

// Space is one registered entry point into a planning repo.
//
// The two identities do different jobs and neither substitutes for the other:
//
//	ID       a LOCAL label — the address. Unique per checkout, typable, completable,
//	         and what `--space` takes.
//	VerifyID the target repo's DURABLE id (config.Describe → ID) — the assertion and
//	         natural logical-space grouping key. Shared by every entry point and
//	         worktree of a planning repo, so it can never be the address.
//
// Path is the repo directory (where .tskflwctl.toml lives), not the resolved planning
// root: registration then means one thing whether the repo scaffolds a tree or points at
// another, and discovery does the rest.
type Space struct {
	ID       string `toml:"id"`
	Path     string `toml:"path"`
	VerifyID string `toml:"verify_id,omitempty"`
	Label    string `toml:"label,omitempty"`
	Accent   string `toml:"accent,omitempty"`
	Added    string `toml:"added,omitempty"`
}

// spacesFileTOML mirrors spaces.toml on disk.
type spacesFileTOML struct {
	SchemaVersion int     `toml:"schema_version"`
	Space         []Space `toml:"space"`
}

// SpacesPath is the registry's location. Shares Dir() with config.toml, so the same
// TSKFLW_CONFIG_HOME override covers both and no test can reach a real $HOME.
func SpacesPath() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, SpacesFile), nil
}

// Spaces reads the registry. A missing file is the normal case and yields no spaces with
// no error; a malformed one errors so a typo is not silently read as "you have none".
//
// Entries are returned in file order. Registration uses insertion order: it is the only
// ordering that lets later edits preserve every existing table byte-for-byte.
func Spaces() ([]Space, error) {
	path, err := SpacesPath()
	if err != nil {
		return nil, nil // no resolvable home: the same silent degrade as Load
	}
	_, spaces, err := readRegistry(path)
	return spaces, err
}

// readRegistry returns the source text and decoded entries from ONE filesystem snapshot.
// Mutations use both values together so a second read can never make a decoded table index
// refer to different text.
func readRegistry(path string) (string, []Space, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return initialSpacesText, nil, nil
		}
		return "", nil, fmt.Errorf("read %s: %w", path, err)
	}
	var f spacesFileTOML
	if _, err := toml.Decode(string(data), &f); err != nil {
		return "", nil, fmt.Errorf("%w: read %s: %v", ErrInvalidRegistry, path, err)
	}
	if f.SchemaVersion != 0 && f.SchemaVersion != spacesSchemaVersion {
		return "", nil, fmt.Errorf(
			"%w: read %s: unsupported schema_version %d (want %d)",
			ErrInvalidRegistry, path, f.SchemaVersion, spacesSchemaVersion,
		)
	}
	return string(data), f.Space, nil
}

// AddSpace registers a repo. The caller is expected to have validated that path IS a
// planning repo first — this layer owns the file, not the meaning.
//
// Dedup is on PHYSICAL paths (symlinks and spellings collapse), never on VerifyID: two
// checkouts of one repo legitimately share a durable id and must stay separately
// addressable. A path that is already registered returns its existing entry with
// added=false rather than erroring, so re-adding is idempotent. dryRun performs the
// identical snapshot validation and returns the exact would-be result without locking,
// creating a directory, or writing.
func AddSpace(s Space, dryRun bool) (added bool, existing Space, err error) {
	path, err := SpacesPath()
	if err != nil {
		return false, Space{}, err
	}
	if dryRun {
		text, spaces, err := readRegistry(path)
		if err != nil {
			return false, Space{}, err
		}
		return addSpaceFromSnapshot(path, text, spaces, s, true)
	}
	unlock, err := lockRegistryForWrite(path)
	if err != nil {
		return false, Space{}, err
	}
	defer unlock()
	text, spaces, err := readRegistry(path)
	if err != nil {
		return false, Space{}, err
	}
	return addSpaceFromSnapshot(path, text, spaces, s, false)
}

func addSpaceFromSnapshot(path, text string, spaces []Space, s Space, dryRun bool) (bool, Space, error) {
	want := PhysicalPath(s.Path)
	for _, e := range spaces {
		if PhysicalPath(e.Path) == want {
			return false, e, nil
		}
		if e.ID == s.ID {
			return false, e, fmt.Errorf("%w: "+
				"space %q is already registered for %s — pass --id to choose a different label",
				ErrSpaceIDConflict, s.ID, e.Path)
		}
	}
	if s.Added == "" {
		s.Added = time.Now().Format("2006-01-02")
	}
	if dryRun {
		return true, s, nil
	}
	updated, err := appendSpaceToText(text, s)
	if err != nil {
		return false, Space{}, err
	}
	if err := writeRegistryText(path, updated); err != nil {
		return false, Space{}, err
	}
	return true, s, nil
}

// ForgetSpace drops an entry by its label. It never touches the repo on disk — the whole
// point is that forgetting is a registry edit, not a deletion. It returns the exact entry
// selected from the same snapshot used for the edit, so mutation receipts cannot race with
// a second read. With dryRun, removed means "would remove" and nothing is written.
func ForgetSpace(id string, dryRun bool) (removed bool, existing Space, err error) {
	path, err := SpacesPath()
	if err != nil {
		return false, Space{}, err
	}
	var text string
	var spaces []Space
	var unlock func()
	if dryRun {
		text, spaces, err = readRegistry(path)
	} else {
		unlock, err = lockRegistryForWrite(path)
		if err == nil {
			defer unlock()
			text, spaces, err = readRegistry(path)
		}
	}
	if err != nil {
		return false, Space{}, err
	}
	index := -1
	for i, e := range spaces {
		if e.ID == id {
			if index >= 0 {
				return false, Space{}, fmt.Errorf(
					"%w: space %q appears more than once in %s", ErrInvalidRegistry, id, path,
				)
			}
			index = i
		}
	}
	if index < 0 {
		return false, Space{}, nil
	}
	existing = spaces[index]
	if dryRun {
		return true, existing, nil
	}
	updated, err := removeSpaceFromText(text, index, len(spaces))
	if err != nil {
		return false, Space{}, fmt.Errorf("%w: edit %s: %v", ErrInvalidRegistry, path, err)
	}
	if err := writeRegistryText(path, updated); err != nil {
		return false, Space{}, err
	}
	return true, existing, nil
}

var initialSpacesText = fmt.Sprintf(
	"# tskflwctl space registry.\n"+
		"# Entries are edited surgically by `space add` / `space forget`; comments and unknown keys survive.\n\n"+
		"schema_version = %d\n",
	spacesSchemaVersion,
)

func lockRegistryForWrite(path string) (func(), error) {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("mkdir %s: %w", dir, err)
	}
	return userConfigWriteLock(dir)
}

// appendSpaceToText adds exactly one canonical [[space]] table at EOF. Existing text is
// never re-encoded, so comments, unknown fields, key order, whitespace, and entry order
// remain byte-for-byte unchanged.
func appendSpaceToText(text string, s Space) (string, error) {
	var b strings.Builder
	if err := toml.NewEncoder(&b).Encode(struct {
		Space []Space `toml:"space"`
	}{Space: []Space{s}}); err != nil {
		return "", fmt.Errorf("encode space %q: %w", s.ID, err)
	}
	if text == "" {
		text = initialSpacesText
	}
	if !strings.HasSuffix(text, "\n") {
		text += "\n"
	}
	if !strings.HasSuffix(text, "\n\n") {
		text += "\n"
	}
	return text + b.String(), nil
}

// removeSpaceFromText removes one [[space]] table by its decoded file-order index.
// A table extends until the next [[space]] or unrelated table header; nested
// [space.*] tables belong to it. Refuse to edit when the textual and decoded table counts
// disagree — preserving a surprising hand-written layout is safer than guessing.
func removeSpaceFromText(text string, index, decodedCount int) (string, error) {
	starts := spaceTableStarts(text)
	if len(starts) != decodedCount {
		return "", fmt.Errorf("found %d textual [[space]] tables but decoded %d", len(starts), decodedCount)
	}
	if index < 0 || index >= len(starts) {
		return "", fmt.Errorf("space table index %d is out of range", index)
	}
	start := starts[index]
	end := nextSpaceBoundary(text, start)
	return text[:start] + text[end:], nil
}

func spaceTableStarts(text string) []int {
	var starts []int
	forEachLine(text, func(pos int, line string) bool {
		if isSpaceTableHeader(line) {
			starts = append(starts, pos)
		}
		return true
	})
	return starts
}

func nextSpaceBoundary(text string, start int) int {
	boundary := len(text)
	seenStart := false
	triviaStart := -1
	forEachLine(text, func(pos int, line string) bool {
		if pos < start {
			return true
		}
		if !seenStart {
			seenStart = true
			return true
		}
		if isSpaceTableHeader(line) || (isTableHeader(line) && !isSpaceChildHeader(line)) {
			boundary = pos
			if triviaStart >= 0 {
				// Preserve the blank/comment prelude immediately before the next table.
				// It may document that table, and retaining an orphaned comment is safer
				// than deleting hand-written information whose ownership is ambiguous.
				boundary = triviaStart
			}
			return false
		}
		if isTOMLTrivia(line) {
			if triviaStart < 0 {
				triviaStart = pos
			}
		} else {
			triviaStart = -1
		}
		return true
	})
	if boundary == len(text) && triviaStart >= 0 {
		return triviaStart // preserve a trailing comment/file footer after the last entry
	}
	return boundary
}

func forEachLine(text string, visit func(pos int, line string) bool) {
	for pos := 0; pos < len(text); {
		end := strings.IndexByte(text[pos:], '\n')
		if end < 0 {
			end = len(text)
		} else {
			end += pos
		}
		if !visit(pos, strings.TrimSuffix(text[pos:end], "\r")) {
			return
		}
		if end == len(text) {
			return
		}
		pos = end + 1
	}
}

func isSpaceTableHeader(line string) bool {
	inner, array, ok := tableHeader(line)
	return ok && array && isSpaceTableName(inner)
}

func isSpaceChildHeader(line string) bool {
	inner, _, ok := tableHeader(line)
	return ok && (strings.HasPrefix(inner, "space.") ||
		strings.HasPrefix(inner, `"space".`) || strings.HasPrefix(inner, "'space'."))
}

func isTableHeader(line string) bool {
	_, _, ok := tableHeader(line)
	return ok
}

func tableHeader(line string) (inner string, array, ok bool) {
	line = strings.TrimSpace(stripTOMLComment(line))
	array = strings.HasPrefix(line, "[[") && strings.HasSuffix(line, "]]")
	if array {
		return strings.TrimSpace(line[2 : len(line)-2]), true, true
	}
	if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
		return strings.TrimSpace(line[1 : len(line)-1]), false, true
	}
	return "", false, false
}

func isSpaceTableName(inner string) bool {
	return inner == "space" || inner == `"space"` || inner == "'space'"
}

func isTOMLTrivia(line string) bool {
	line = strings.TrimSpace(line)
	return line == "" || strings.HasPrefix(line, "#")
}

// stripTOMLComment removes a table header's trailing comment while respecting quoted
// keys. A raw strings.IndexByte('#') mistakes valid headers such as ["notes#archive"] for
// comments, which can make forget run through the unrelated table and delete it.
func stripTOMLComment(line string) string {
	var quote byte
	for i := 0; i < len(line); i++ {
		c := line[i]
		if quote == 0 {
			switch c {
			case '#':
				return line[:i]
			case '"', '\'':
				quote = c
			}
			continue
		}
		if quote == '"' && c == '\\' {
			i++ // escaped byte in a TOML basic quoted key
			continue
		}
		if c == quote {
			quote = 0
		}
	}
	return line
}

func writeRegistryText(path, text string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("mkdir %s: %w", filepath.Dir(path), err)
	}
	return writeFileAtomic(path, []byte(text), 0o644)
}
