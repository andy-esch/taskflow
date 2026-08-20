package userconfig

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
)

// The space registry: which planning repos this machine knows about.
//
// It lives in its OWN file, separate from the hand-edited config.toml, because the two
// have opposite ownership. config.toml is written by a person and only ever read by the
// tool, so its comments and key order must survive. spaces.toml is written by the TOOL —
// `space add`/`forget` rewrite it wholesale — which is what makes writes a plain marshal
// instead of the surgical array-of-tables editor a single shared file would have needed.
//
// The registry is ADVISORY: nothing here may change what config.Discover resolves from a
// given cwd. A machine with no spaces.toml behaves exactly as it did before one existed,
// and deleting the file costs convenience, never data or addressability. (internal/config
// does not import this package, so that is enforced rather than remembered.)

// SpacesFile is the tool-owned registry filename, beside config.toml.
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
// Entries are returned in file order, which is sorted by ID (see writeSpaces).
func Spaces() ([]Space, error) {
	path, err := SpacesPath()
	if err != nil {
		return nil, nil // no resolvable home: the same silent degrade as Load
	}
	var f spacesFileTOML
	if _, err := toml.DecodeFile(path, &f); err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read %s: %w", path, err)
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
	if err := writeSpaces(append(spaces, s)); err != nil {
		return false, Space{}, err
	}
	return true, s, nil
}

// ForgetSpace drops an entry by its label. It never touches the repo on disk — the whole
// point is that forgetting is a registry edit, not a deletion.
func ForgetSpace(id string) (bool, error) {
	spaces, err := Spaces()
	if err != nil {
		return false, err
	}
	kept := make([]Space, 0, len(spaces))
	for _, e := range spaces {
		if e.ID != id {
			kept = append(kept, e)
		}
	}
	if len(kept) == len(spaces) {
		return false, nil
	}
	return true, writeSpaces(kept)
}

// writeSpaces rewrites the registry wholesale, sorted by ID.
//
// Sorted rather than insertion-ordered because the file is tool-owned: there is no human
// ordering intent to preserve, and a stable order keeps diffs (and any golden test) from
// churning on unrelated adds.
func writeSpaces(spaces []Space) error {
	path, err := SpacesPath()
	if err != nil {
		return err
	}
	sort.Slice(spaces, func(i, j int) bool { return spaces[i].ID < spaces[j].ID })
	var b strings.Builder
	b.WriteString("# tskflwctl space registry — MANAGED BY THE TOOL.\n" +
		"# Rewritten wholesale by `space add` / `space forget`; comments and unknown keys\n" +
		"# added by hand are not preserved. Personal settings belong in config.toml beside it.\n\n")
	if err := toml.NewEncoder(&b).Encode(spacesFileTOML{SchemaVersion: spacesSchemaVersion, Space: spaces}); err != nil {
		return fmt.Errorf("encode %s: %w", path, err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("mkdir %s: %w", filepath.Dir(path), err)
	}
	return writeFileAtomic(path, []byte(b.String()), 0o644)
}
