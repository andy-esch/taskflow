package configstore

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/testutil"
	"github.com/andy-esch/taskflow/internal/userconfig"
)

func TestConfigurationPoliciesDenyBeforeEffectsIncludingPreview(t *testing.T) {
	denied := errors.New("configuration refused")
	for _, tc := range []struct {
		name   string
		policy core.MutationPolicy
		want   error
	}{
		{"guarded", core.GuardedMutations(func() error { return denied }), denied},
		{"read-only", core.ReadOnlyMutations(), core.ErrReadOnlyPersistence},
		{"zero adapter", core.MutationPolicy{}, core.ErrInvalidMutationPolicy},
	} {
		for _, dryRun := range []bool{true, false} {
			t.Run(tc.name+map[bool]string{true: "/preview", false: "/commit"}[dryRun], func(t *testing.T) {
				base := t.TempDir()
				start := filepath.Join(base, "missing-repository")
				t.Setenv(userconfig.DirEnv, filepath.Join(base, "missing-home"))
				adapter := &FS{}
				if tc.name != "zero adapter" {
					adapter = testutil.Must(New(tc.policy))
				}
				if _, err := adapter.MigrateConfiguration(start, dryRun); !errors.Is(err, tc.want) {
					t.Fatalf("migration error=%v; want %v", err, tc.want)
				}
				for _, scope := range []core.ConfigScope{core.ConfigScopeUser, core.ConfigScopeRepository} {
					if _, err := adapter.SetPreference(start, core.PreferenceChange{Scope: scope, Field: core.PreferenceTheme, Value: "neon"}, dryRun); !errors.Is(err, tc.want) {
						t.Fatalf("preference error=%v; want %v", err, tc.want)
					}
				}
				if entries, err := os.ReadDir(base); err != nil || len(entries) != 0 {
					t.Fatalf("refusal touched storage: entries=%v err=%v", entries, err)
				}
			})
		}
	}
}
