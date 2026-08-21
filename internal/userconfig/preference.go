package userconfig

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"

	"github.com/andy-esch/taskflow/internal/tomledit"
)

type PreferenceField string

const (
	PreferenceThemeName    PreferenceField = "theme.name"
	PreferencePagerEnabled PreferenceField = "pager.enabled"
	PreferencePagerCommand PreferenceField = "pager.command"
)

// SetPreference edits one user-scoped preference atomically and surgically. value is
// already TOML-encoded; nil removes the override. The config directory and file are
// created on the first real write, never during dry-run.
func SetPreference(field PreferenceField, value *string, dryRun bool) (string, bool, error) {
	dir, err := Dir()
	if err != nil {
		return "", false, err
	}
	path := filepath.Join(dir, FileName)
	table, key, err := preferenceFieldParts(field)
	if err != nil {
		return "", false, err
	}
	text := ""
	if b, readErr := os.ReadFile(path); readErr == nil {
		text = string(b)
	} else if !os.IsNotExist(readErr) {
		return "", false, fmt.Errorf("read %s: %w", path, readErr)
	}
	updated, changed, err := tomledit.SetTableKey(text, table, key, value)
	if err != nil || !changed {
		return path, changed, err
	}
	var check configFileTOML
	if _, err := toml.Decode(updated, &check); err != nil {
		return "", false, fmt.Errorf("edit %s: %w", path, err)
	}
	if dryRun {
		return path, true, nil
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", false, fmt.Errorf("mkdir %s: %w", dir, err)
	}
	unlock, err := userConfigWriteLock(dir)
	if err != nil {
		return "", false, err
	}
	defer unlock()
	// Re-read while holding the directory lock so another cooperating writer cannot
	// disappear between the optimistic edit above and the atomic rename.
	text = ""
	if b, readErr := os.ReadFile(path); readErr == nil {
		text = string(b)
	} else if !os.IsNotExist(readErr) {
		return "", false, fmt.Errorf("read %s: %w", path, readErr)
	}
	updated, changed, err = tomledit.SetTableKey(text, table, key, value)
	if err != nil || !changed {
		return path, changed, err
	}
	if _, err := toml.Decode(updated, &check); err != nil {
		return "", false, fmt.Errorf("edit %s: %w", path, err)
	}
	if err := writeFileAtomic(path, []byte(updated), 0o644); err != nil {
		return "", false, fmt.Errorf("write %s: %w", path, err)
	}
	return path, true, nil
}

func preferenceFieldParts(field PreferenceField) (string, string, error) {
	switch field {
	case PreferenceThemeName:
		return "theme", "name", nil
	case PreferencePagerEnabled:
		return "pager", "enabled", nil
	case PreferencePagerCommand:
		return "pager", "command", nil
	default:
		return "", "", fmt.Errorf("unsupported user preference %q", field)
	}
}
