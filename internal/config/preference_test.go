package config

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
)

func TestSetPresentationPreservesTextAndSupportsInheritance(t *testing.T) {
	root := t.TempDir()
	mustMkdir(t, filepath.Join(root, "tasks"))
	path := filepath.Join(root, ConfigFile)
	input := "# keep\ntaskflow_root = \".\"\nid = \"space-id\"\nunknown = 7\n\n[theme]\n  name = \"neon\" # personal note\n\n[other]\nvalue = \"untouched\"\n"
	if err := os.WriteFile(path, []byte(input), 0o644); err != nil {
		t.Fatal(err)
	}
	value := `"catppuccin"`
	if _, changed, err := SetPresentation(root, PresentationThemeName, &value, false); err != nil || !changed {
		t.Fatalf("changed=%v err=%v", changed, err)
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"# keep", "unknown = 7", `  name = "catppuccin" # personal note`, "[other]", `value = "untouched"`} {
		if !strings.Contains(string(got), want) {
			t.Errorf("missing %q:\n%s", want, got)
		}
	}
	if _, changed, err := SetPresentation(root, PresentationThemeName, nil, false); err != nil || !changed {
		t.Fatalf("unset: changed=%v err=%v", changed, err)
	}
	got, _ = os.ReadFile(path)
	if strings.Contains(string(got), "name =") || !strings.Contains(string(got), "[theme]") {
		t.Fatalf("unset should remove only the scoped key:\n%s", got)
	}
}

func TestSetPresentationDryRunAndConcurrentFields(t *testing.T) {
	root := t.TempDir()
	if _, err := Init(root, "", false); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(root, ConfigFile)
	before, _ := os.ReadFile(path)
	theme := `"catppuccin"`
	if _, changed, err := SetPresentation(root, PresentationThemeName, &theme, true); err != nil || !changed {
		t.Fatalf("dry-run: changed=%v err=%v", changed, err)
	}
	if after, _ := os.ReadFile(path); string(after) != string(before) {
		t.Fatal("dry-run changed repository config")
	}

	command := `"delta"`
	start := make(chan struct{})
	errs := make(chan error, 2)
	var wg sync.WaitGroup
	for _, edit := range []func() error{
		func() error { _, _, err := SetPresentation(root, PresentationThemeName, &theme, false); return err },
		func() error {
			_, _, err := SetPresentation(root, PresentationPagerCommand, &command, false)
			return err
		},
	} {
		wg.Add(1)
		go func(edit func() error) {
			defer wg.Done()
			<-start
			errs <- edit()
		}(edit)
	}
	close(start)
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
	cfg, err := Discover(root)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Theme.Name != "catppuccin" || cfg.Pager.Command != "delta" {
		t.Fatalf("concurrent scoped edits lost a value: theme=%q command=%q", cfg.Theme.Name, cfg.Pager.Command)
	}
}

func TestConfigurationMutationsPreserveSymlinkAndPermissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink and Unix permission contract is exercised on release platforms")
	}
	root := t.TempDir()
	mustMkdir(t, filepath.Join(root, "tasks"))
	target := filepath.Join(t.TempDir(), "repo-config.toml")
	input := "# dotfiles target\ntaskflow_root = \".\"\n\n[theme]\nname = \"neon\"\n"
	if err := os.WriteFile(target, []byte(input), 0o600); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(root, ConfigFile)
	if err := os.Symlink(target, link); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}

	// Migration and preference edits share writeFileAtomic; exercise both public
	// paths so neither can regress to replacing the link or widening the target.
	if _, err := migrateWithIDGen(root, false, func() string { return "6g27symlinkid" }); err != nil {
		t.Fatal(err)
	}
	value := `"catppuccin"`
	if _, changed, err := SetPresentation(root, PresentationThemeName, &value, false); err != nil || !changed {
		t.Fatalf("changed=%v err=%v", changed, err)
	}
	if info, err := os.Lstat(link); err != nil || info.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("repository config symlink was replaced: info=%v err=%v", info, err)
	}
	if info, err := os.Stat(target); err != nil || info.Mode().Perm() != 0o600 {
		t.Fatalf("repository config permissions changed: info=%v err=%v", info, err)
	}
	b, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{`id = "6g27symlinkid"`, `name = "catppuccin"`} {
		if !strings.Contains(string(b), want) {
			t.Errorf("symlink target missing %q:\n%s", want, b)
		}
	}
}
