package userconfig

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestSetPreferenceDryRunDoesNotCreateHomeConfig(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "not-created")
	t.Setenv(DirEnv, dir)
	value := `"catppuccin"`
	path, changed, err := SetPreference(PreferenceThemeName, &value, true)
	if err != nil || !changed || path != filepath.Join(dir, FileName) {
		t.Fatalf("path=%q changed=%v err=%v", path, changed, err)
	}
	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		t.Fatalf("dry-run created config directory: %v", err)
	}
}

func TestSetPreferencePreservesCommentsAndUnknownKeys(t *testing.T) {
	dir := t.TempDir()
	t.Setenv(DirEnv, dir)
	path := filepath.Join(dir, FileName)
	input := "# keep\nunknown = 7\n\n[pager]\nenabled = true # note\n\n[other]\nvalue = \"untouched\"\n"
	if err := os.WriteFile(path, []byte(input), 0o644); err != nil {
		t.Fatal(err)
	}
	value := "false"
	if _, changed, err := SetPreference(PreferencePagerEnabled, &value, false); err != nil || !changed {
		t.Fatalf("changed=%v err=%v", changed, err)
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"# keep", "unknown = 7", "enabled = false # note", "[other]", `value = "untouched"`} {
		if !strings.Contains(string(got), want) {
			t.Errorf("missing %q:\n%s", want, got)
		}
	}
}

func TestSetPreferenceDryRunStillValidatesMalformedTOML(t *testing.T) {
	dir := t.TempDir()
	t.Setenv(DirEnv, dir)
	if err := os.WriteFile(filepath.Join(dir, FileName), []byte("[pager\nenabled = true\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	value := "false"
	if _, _, err := SetPreference(PreferencePagerEnabled, &value, true); err == nil {
		t.Fatal("dry-run must validate the resulting TOML")
	}
}

func TestSetPreferencePreservesSymlinkAndPermissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink and Unix permission contract is exercised on release platforms")
	}
	dir := t.TempDir()
	t.Setenv(DirEnv, dir)
	target := filepath.Join(t.TempDir(), "config.toml")
	if err := os.WriteFile(target, []byte("# dotfiles target\n[theme]\nname = \"neon\"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(dir, FileName)
	if err := os.Symlink(target, link); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}

	value := `"catppuccin"`
	if _, changed, err := SetPreference(PreferenceThemeName, &value, false); err != nil || !changed {
		t.Fatalf("changed=%v err=%v", changed, err)
	}
	if info, err := os.Lstat(link); err != nil || info.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("user config symlink was replaced: info=%v err=%v", info, err)
	}
	if info, err := os.Stat(target); err != nil || info.Mode().Perm() != 0o600 {
		t.Fatalf("user config permissions changed: info=%v err=%v", info, err)
	}
	b, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), `name = "catppuccin"`) {
		t.Errorf("symlink target was not edited:\n%s", b)
	}
}
