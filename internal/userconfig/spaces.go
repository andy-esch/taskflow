package userconfig

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
)

// The space registry: which planning repos this machine knows about.
//
// It lives in its OWN file, separate from the hand-edited config.toml, because registry
// mutations should never risk a person's presentation settings. The registry is still
// inspectable and hand-editable: `space add` appends one table and `space forget` removes
// one table, leaving comments, key order, unknown keys, and every other entry byte-for-byte
// intact. Writes are atomic.
//
// The registry is ADVISORY: nothing here may change what config.Discover resolves from a
// given cwd. A machine with no spaces.toml behaves exactly as it did before one existed,
// and deleting the file costs convenience, never data or addressability. (internal/config
// does not import this package, so that is enforced rather than remembered.)

// SpacesFile is the registry filename, beside config.toml.
const SpacesFile = "spaces.toml"

// spacesSchemaVersion is the on-disk shape of spaces.toml, independent of the --json
// contract's schema_version. Bumped only if the entry shape changes incompatibly.
const spacesSchemaVersion = 1

// Space is one registered planning repo.
//
// The two identities do different jobs and neither substitutes for the other:
//
//	ID       a LOCAL label — the address. Unique per checkout, typable, completable,
//	         and what `--space` takes.
//	VerifyID the target repo's DURABLE id (config.Describe → ID) — the assertion.
//	         Shared by every worktree of a repo, so it can never be the address.
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
	return readSpaces(path)
}

func readSpaces(path string) ([]Space, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	var f spacesFileTOML
	if _, err := toml.Decode(string(data), &f); err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	if f.SchemaVersion != 0 && f.SchemaVersion != spacesSchemaVersion {
		return nil, fmt.Errorf("read %s: unsupported schema_version %d (want %d)", path, f.SchemaVersion, spacesSchemaVersion)
	}
	return f.Space, nil
}

// AddSpace registers a repo. The caller is expected to have validated that path IS a
// planning repo first — this layer owns the file, not the meaning.
//
// Dedup is on PHYSICAL paths (symlinks and spellings collapse), never on VerifyID: two
// checkouts of one repo legitimately share a durable id and must stay separately
// addressable. A path that is already registered returns its existing entry with
// added=false rather than erroring, so re-adding is idempotent.
func AddSpace(s Space) (added bool, existing Space, err error) {
	spaces, err := Spaces()
	if err != nil {
		return false, Space{}, err
	}
	want := PhysicalPath(s.Path)
	for _, e := range spaces {
		if PhysicalPath(e.Path) == want {
			return false, e, nil
		}
		if e.ID == s.ID {
			return false, e, fmt.Errorf(
				"space %q is already registered for %s — pass --id to choose a different label",
				s.ID, e.Path)
		}
	}
	if s.Added == "" {
		s.Added = time.Now().Format("2006-01-02")
	}
	path, err := SpacesPath()
	if err != nil {
		return false, Space{}, err
	}
	text, err := readRegistryText(path)
	if err != nil {
		return false, Space{}, err
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
// point is that forgetting is a registry edit, not a deletion.
func ForgetSpace(id string) (bool, error) {
	path, err := SpacesPath()
	if err != nil {
		return false, err
	}
	spaces, err := readSpaces(path)
	if err != nil {
		return false, err
	}
	index := -1
	for i, e := range spaces {
		if e.ID == id {
			if index >= 0 {
				return false, fmt.Errorf("space %q appears more than once in %s", id, path)
			}
			index = i
		}
	}
	if index < 0 {
		return false, nil
	}
	text, err := readRegistryText(path)
	if err != nil {
		return false, err
	}
	updated, err := removeSpaceFromText(text, index, len(spaces))
	if err != nil {
		return false, fmt.Errorf("edit %s: %w", path, err)
	}
	return true, writeRegistryText(path, updated)
}

var initialSpacesText = fmt.Sprintf(
	"# tskflwctl space registry.\n"+
		"# Entries are edited surgically by `space add` / `space forget`; comments and unknown keys survive.\n\n"+
		"schema_version = %d\n",
	spacesSchemaVersion,
)

func readRegistryText(path string) (string, error) {
	b, err := os.ReadFile(path)
	if err == nil {
		return string(b), nil
	}
	if os.IsNotExist(err) {
		return initialSpacesText, nil
	}
	return "", fmt.Errorf("read %s: %w", path, err)
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
			return false
		}
		return true
	})
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
	return ok && array && inner == "space"
}

func isSpaceChildHeader(line string) bool {
	inner, _, ok := tableHeader(line)
	return ok && strings.HasPrefix(inner, "space.")
}

func isTableHeader(line string) bool {
	_, _, ok := tableHeader(line)
	return ok
}

func tableHeader(line string) (inner string, array, ok bool) {
	line = strings.TrimSpace(line)
	if i := strings.IndexByte(line, '#'); i >= 0 {
		line = strings.TrimSpace(line[:i])
	}
	array = strings.HasPrefix(line, "[[") && strings.HasSuffix(line, "]]")
	if array {
		return strings.TrimSpace(line[2 : len(line)-2]), true, true
	}
	if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
		return strings.TrimSpace(line[1 : len(line)-1]), false, true
	}
	return "", false, false
}

func writeRegistryText(path, text string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("mkdir %s: %w", filepath.Dir(path), err)
	}
	// A dotfiles-managed registry is commonly a symlink. Follow it instead of atomically
	// replacing the link itself; the target still receives the temp-file + rename write.
	destination := path
	if fi, err := os.Lstat(path); err == nil && fi.Mode()&os.ModeSymlink != 0 {
		resolved, err := filepath.EvalSymlinks(path)
		if err != nil {
			return fmt.Errorf("resolve %s: %w", path, err)
		}
		destination = resolved
	}
	perm := os.FileMode(0o644)
	if fi, err := os.Stat(destination); err == nil {
		perm = fi.Mode().Perm()
	}
	return writeFileAtomic(destination, []byte(text), perm)
}
