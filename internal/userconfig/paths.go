package userconfig

import (
	"os"
	"path/filepath"
	"strings"
)

// Path handling for registry entries.
//
// Two different jobs, deliberately not conflated: entries are STORED with `~` intact so
// the file stays portable and committable to a dotfiles repo, and COMPARED on physical
// paths so `../x`, an absolute spelling, and a symlinked checkout collapse to one entry.

// TildePath re-abbreviates an absolute path under the user's home to `~/…` — the form
// stored in spaces.toml. Returns p unchanged when it is not under $HOME, or when the home
// dir cannot be resolved.
func TildePath(p string) string {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return p
	}
	rel, err := filepath.Rel(home, p)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return p
	}
	if rel == "." {
		return "~"
	}
	return "~/" + filepath.ToSlash(rel)
}

// ExpandTilde is TildePath's inverse: `~` and `~/x` become absolute. Anything else is
// returned unchanged, including `~user` forms, which are deliberately NOT supported —
// guessing another user's home would be worse than leaving the path alone.
func ExpandTilde(p string) string {
	if p != "~" && !strings.HasPrefix(p, "~/") {
		return p
	}
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return p
	}
	if p == "~" {
		return home
	}
	return filepath.Join(home, filepath.FromSlash(strings.TrimPrefix(p, "~/")))
}

// PhysicalPath resolves p to an absolute, symlink-free path for COMPARISON only — the
// same Abs+EvalSymlinks discipline internal/config uses, so the two agree about when two
// spellings name one directory. A path that does not exist degrades to its lexical
// absolute form rather than erroring: a registered repo that has been deleted must still
// compare equal to itself.
func PhysicalPath(p string) string {
	p = ExpandTilde(p)
	abs, err := filepath.Abs(p)
	if err != nil {
		return p
	}
	if resolved, err := filepath.EvalSymlinks(abs); err == nil {
		return resolved
	}
	return abs
}

// writeFileAtomic writes via a temp file + rename, so a crash mid-write cannot leave a
// half-written config. Existing permissions survive, and a symlink is followed so a
// dotfiles-managed config keeps being a symlink instead of being replaced by the temp
// file. Mirrors store/atomic.go's contract for the planning tree.
func writeFileAtomic(path string, data []byte, perm os.FileMode) error {
	destination, perm, err := atomicWriteDestination(path, perm)
	if err != nil {
		return err
	}
	dir := filepath.Dir(destination)
	tmp, err := os.CreateTemp(dir, ".tmp-*")
	if err != nil {
		return err
	}
	name := tmp.Name()
	defer func() { _ = os.Remove(name) }()
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Chmod(name, perm); err != nil {
		return err
	}
	return os.Rename(name, destination)
}

func atomicWriteDestination(path string, fallback os.FileMode) (string, os.FileMode, error) {
	destination := path
	if info, err := os.Lstat(path); err == nil {
		if info.Mode()&os.ModeSymlink != 0 {
			resolved, err := filepath.EvalSymlinks(path)
			if err != nil {
				return "", 0, err
			}
			destination = resolved
		}
	} else if !os.IsNotExist(err) {
		return "", 0, err
	}
	if info, err := os.Stat(destination); err == nil {
		fallback = info.Mode().Perm()
	} else if !os.IsNotExist(err) {
		return "", 0, err
	}
	return destination, fallback, nil
}
