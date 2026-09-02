package cli

import (
	"strings"
	"testing"
)

// `task depend add A B` is the shape the verb itself reads like, and cobra answered
// it with "accepts 1 arg(s), received 2" — an arity count that says nothing about
// where the second argument was supposed to go.
func TestDependAdd_ExtraPositionalNamesTheFlagForm(t *testing.T) {
	root := setupRepo(t)

	_, err := runRootRC(t, "-C", root, "task", "depend", "add", "alpha", "beta")
	if err == nil {
		t.Fatal("two positional args should be rejected")
	}
	if got := ExitCode(err); got != 11 {
		t.Errorf("exit = %d, want 11", got)
	}
	msg := err.Error()
	if !strings.Contains(msg, "task depend add alpha --on beta") {
		t.Errorf("the error should spell the flag form out: %s", msg)
	}
	if strings.Contains(msg, "accepts 1 arg") {
		t.Errorf("the bare arity error should be gone: %s", msg)
	}
}

// Several trailing prerequisites are the same mistake, and each needs its own --on.
func TestDependAdd_SeveralExtrasEachGetTheirOwnFlag(t *testing.T) {
	root := setupRepo(t)

	_, err := runRootRC(t, "-C", root, "task", "depend", "add", "alpha", "beta", "gamma")
	if err == nil {
		t.Fatal("three positional args should be rejected")
	}
	if !strings.Contains(err.Error(), "--on beta --on gamma") {
		t.Errorf("each prerequisite needs its own --on: %v", err)
	}
}

// remove shares the builder, so it must share the guidance.
func TestDependRemove_ExtraPositionalNamesTheFlagForm(t *testing.T) {
	root := setupRepo(t)

	_, err := runRootRC(t, "-C", root, "task", "depend", "remove", "alpha", "beta")
	if err == nil {
		t.Fatal("two positional args should be rejected")
	}
	if !strings.Contains(err.Error(), "task depend remove alpha --on beta") {
		t.Errorf("remove should name its own flag form: %v", err)
	}
}

func TestDependAdd_MissingTaskNamesTheShape(t *testing.T) {
	root := setupRepo(t)

	_, err := runRootRC(t, "-C", root, "task", "depend", "add")
	if err == nil {
		t.Fatal("no positional arg should be rejected")
	}
	if got := ExitCode(err); got != 11 {
		t.Errorf("exit = %d, want 11", got)
	}
	if !strings.Contains(err.Error(), "--on <prereq>") {
		t.Errorf("the error should name the full shape: %v", err)
	}
}

// The correct form still works — the guard only rejects extra positionals.
func TestDependAdd_FlagFormStillWorks(t *testing.T) {
	root := setupRepo(t)

	runRoot(t, "-C", root, "task", "depend", "add", "beta", "--on", "alpha")

	out := runRoot(t, "-C", root, "task", "blockers", "beta")
	if !strings.Contains(out, "alpha") {
		t.Errorf("the dependency should have been recorded:\n%s", out)
	}
}
