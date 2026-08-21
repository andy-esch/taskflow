package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"

	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/tomledit"
)

type PresentationField string

const (
	PresentationThemeName    PresentationField = "theme.name"
	PresentationPagerEnabled PresentationField = "pager.enabled"
	PresentationPagerCommand PresentationField = "pager.command"
)

// SetPresentation edits one repository-scoped presentation preference. value is already
// TOML-encoded; nil removes the override. Only the three typed presentation keys are
// accepted. The final document is decoded before an atomic write.
func SetPresentation(start string, field PresentationField, value *string, dryRun bool) (string, bool, error) {
	cfg, err := Discover(start)
	if err != nil {
		return "", false, err
	}
	if cfg.Dir == "" {
		return "", false, fmt.Errorf("%w: no %s governs %s", domain.ErrValidation, ConfigFile, start)
	}
	path := filepath.Join(cfg.Dir, ConfigFile)
	table, key, err := presentationFieldParts(field)
	if err != nil {
		return "", false, err
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return "", false, err
	}
	updated, changed, err := tomledit.SetTableKey(string(b), table, key, value)
	if err != nil || !changed {
		return path, changed, err
	}
	var check configFile
	if _, err := toml.Decode(updated, &check); err != nil {
		return "", false, fmt.Errorf("%w: edit %s: %v", domain.ErrValidation, path, err)
	}
	if dryRun {
		return path, true, nil
	}
	unlock, err := writeLock(cfg.Dir)
	if err != nil {
		return "", false, err
	}
	defer unlock()
	// Re-read inside the critical section: another taskflow process may have changed
	// a different config key after the optimistic dry-run calculation above.
	b, err = os.ReadFile(path)
	if err != nil {
		return "", false, err
	}
	updated, changed, err = tomledit.SetTableKey(string(b), table, key, value)
	if err != nil || !changed {
		return path, changed, err
	}
	if _, err := toml.Decode(updated, &check); err != nil {
		return "", false, fmt.Errorf("%w: edit %s: %v", domain.ErrValidation, path, err)
	}
	if err := writeFileAtomic(path, []byte(updated), 0o644); err != nil {
		return "", false, fmt.Errorf("write %s: %w", path, err)
	}
	return path, true, nil
}

func presentationFieldParts(field PresentationField) (string, string, error) {
	switch field {
	case PresentationThemeName:
		return "theme", "name", nil
	case PresentationPagerEnabled:
		return "pager", "enabled", nil
	case PresentationPagerCommand:
		return "pager", "command", nil
	default:
		return "", "", fmt.Errorf("%w: unsupported repository preference %q", domain.ErrValidation, field)
	}
}
