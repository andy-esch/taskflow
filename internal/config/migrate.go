package config

import (
	"fmt"
	"path/filepath"

	"github.com/andy-esch/taskflow/internal/domain"
	"github.com/andy-esch/taskflow/internal/id"
)

// MigrationKind identifies one safe, mechanical configuration upgrade.
type MigrationKind string

const (
	MigrationRepoID         MigrationKind = "repo-id"
	MigrationPlanningRepoID MigrationKind = "planning-repo-id"
)

// MigrationStep is one key that Migrate added (or would add under dry-run). Value is
// the exact post-migration value, so human and machine adapters can render one result
// rather than inspecting the unchanged file again after a preview.
type MigrationStep struct {
	Kind  MigrationKind
	Key   string
	Value string
}

// MigrationResult is the complete outcome for one repository configuration.
type MigrationResult struct {
	ConfigPath string
	Mode       string
	DryRun     bool
	Steps      []MigrationStep
}

// PendingMigrations reports which safe upgrades an existing configuration needs.
// It performs the same target validation as Migrate but does not mint ids or write.
func PendingMigrations(start string) ([]MigrationKind, error) {
	_, cf, err := migrationConfig(start)
	if err != nil {
		return nil, err
	}
	if cf.PlanningRepo == "" {
		if cf.ID == "" {
			return []MigrationKind{MigrationRepoID}, nil
		}
		return []MigrationKind{}, nil
	}
	if cf.PlanningRepoID != "" {
		return []MigrationKind{}, nil
	}
	// The missing expectation itself is enough to say a migration is pending. Target
	// validation (and the actionable "migrate the target first" error) belongs to the
	// explicit Migrate attempt; `init`/`config show` must still be able to describe a
	// legacy pointer whose target has not yet been modernized.
	return []MigrationKind{MigrationPlanningRepoID}, nil
}

// Migrate plans and applies the safe upgrades for the configuration governing start.
// It is idempotent. dryRun executes all parsing and target validation and returns the
// exact would-be values without touching disk.
func Migrate(start string, dryRun bool) (MigrationResult, error) {
	return migrateWithIDGen(start, dryRun, id.New)
}

func migrateWithIDGen(start string, dryRun bool, newID func() string) (MigrationResult, error) {
	if dryRun {
		return migrateUnlocked(start, true, newID)
	}
	// Discover once to locate the stable directory, then take the same config lock
	// used by preference and linkback writes. Re-read and revalidate everything under
	// the lock so two cooperating processes cannot overwrite each other's TOML edits.
	dir, _, err := migrationConfig(start)
	if err != nil {
		return MigrationResult{}, err
	}
	unlock, err := writeLock(dir)
	if err != nil {
		return MigrationResult{}, err
	}
	defer unlock()
	return migrateUnlocked(start, false, newID)
}

func migrateUnlocked(start string, dryRun bool, newID func() string) (MigrationResult, error) {
	dir, cf, err := migrationConfig(start)
	if err != nil {
		return MigrationResult{}, err
	}
	result := MigrationResult{
		ConfigPath: filepath.Join(dir, ConfigFile),
		Mode:       "scaffold",
		DryRun:     dryRun,
		Steps:      []MigrationStep{},
	}
	if cf.PlanningRepo == "" {
		if cf.ID != "" {
			return result, nil
		}
		mintedID := newID()
		added, err := backfillID(result.ConfigPath, mintedID, dryRun)
		if err != nil {
			return MigrationResult{}, err
		}
		if added {
			result.Steps = append(result.Steps, MigrationStep{
				Kind: MigrationRepoID, Key: "id", Value: mintedID,
			})
		}
		return result, nil
	}

	result.Mode = "pointer"
	if cf.PlanningRepoID != "" {
		return result, nil
	}
	root, err := resolvePlanningRepo(dir, cf.PlanningRepo, "")
	if err != nil {
		return MigrationResult{}, err
	}
	targetID := planningRepoID(root)
	if targetID == "" {
		return MigrationResult{}, fmt.Errorf(
			"%w: planning_repo %q has no durable id — run `tskflwctl -C %s config migrate` there first",
			domain.ErrValidation, cf.PlanningRepo, root,
		)
	}
	added, err := backfillPointerID(result.ConfigPath, cf, targetID, dryRun)
	if err != nil {
		return MigrationResult{}, err
	}
	if added {
		result.Steps = append(result.Steps, MigrationStep{
			Kind: MigrationPlanningRepoID, Key: "planning_repo_id", Value: targetID,
		})
	}
	return result, nil
}

// migrationConfig locates and parses the repository config without requiring it to be
// current. Discover remains the authority for the walk-up and pointer validation; its
// returned Dir identifies the config that governed the resolution.
func migrationConfig(start string) (string, configFile, error) {
	cfg, err := Discover(start)
	if err != nil {
		return "", configFile{}, err
	}
	if cfg.Dir == "" {
		return "", configFile{}, fmt.Errorf(
			"%w: no %s governs %s — run `tskflwctl init` first",
			domain.ErrValidation, ConfigFile, start,
		)
	}
	cf, err := readConfigFile(filepath.Join(cfg.Dir, ConfigFile))
	if err != nil {
		return "", configFile{}, err
	}
	return cfg.Dir, cf, nil
}
