package cli

import (
	"strings"
	"testing"
)

// `tskflwctl list` suggested `lint` and nothing else: a validator offered in place
// of a read, on a repository the caller may not own. The answer must name the
// noun-qualified forms instead.
func TestBareVerb_ListNamesTheNounQualifiedForms(t *testing.T) {
	root := setupRepo(t)

	_, err := runRootRC(t, "-C", root, "list")
	if err == nil {
		t.Fatal("a bare verb should be an error")
	}
	if got := ExitCode(err); got != 11 {
		t.Errorf("exit = %d, want 11", got)
	}
	msg := err.Error()
	for _, want := range []string{"task list", "board"} {
		if !strings.Contains(msg, want) {
			t.Errorf("the hint should name %q: %s", want, msg)
		}
	}
	if strings.Contains(msg, "lint") {
		t.Errorf("`lint` is a validator, not what `list` meant: %s", msg)
	}
}

func TestBareVerb_CoversTheCommonSlips(t *testing.T) {
	root := setupRepo(t)
	for verb, wantForm := range map[string]string{
		"show":     "task show",
		"new":      "task new",
		"start":    "task start",
		"complete": "task complete",
		"edit":     "task edit",
	} {
		_, err := runRootRC(t, "-C", root, verb)
		if err == nil {
			t.Errorf("%q should be an error", verb)
			continue
		}
		if !strings.Contains(err.Error(), wantForm) {
			t.Errorf("%q should point at %q: %v", verb, wantForm, err)
		}
	}
}

// A usage error must not depend on repo discovery: outside a planning tree the
// caller would otherwise be told to run `init`, which is not what they got wrong.
func TestBareVerb_AnswersOutsideAPlanningRepo(t *testing.T) {
	_, err := runRootRC(t, "-C", t.TempDir(), "list")
	if err == nil {
		t.Fatal("a bare verb should still be an error")
	}
	if strings.Contains(err.Error(), "not a taskflow planning repo") {
		t.Errorf("the usage hint should win over repo discovery: %v", err)
	}
	if !strings.Contains(err.Error(), "task list") {
		t.Errorf("want the noun-qualified hint, got: %v", err)
	}
}

// The redirects are hidden: they answer a mistake, they are not part of the surface.
func TestBareVerb_StaysOffTheHelpSurface(t *testing.T) {
	out := runRoot(t, "--help")
	for _, line := range strings.Split(out, "\n") {
		trimmed := strings.TrimSpace(line)
		for verb := range bareVerbRedirects {
			if strings.HasPrefix(trimmed, verb+" ") {
				t.Errorf("%q should not appear in help: %q", verb, line)
			}
		}
	}
}

// Real typos still reach cobra's distance matching — the redirects intercept only
// the exact bare verbs.
func TestBareVerb_LeavesOrdinaryTyposToCobra(t *testing.T) {
	root := setupRepo(t)

	_, err := runRootRC(t, "-C", root, "boadr")
	if err == nil {
		t.Fatal("an unknown command should still fail")
	}
	if !strings.Contains(err.Error(), "unknown command") {
		t.Errorf("want cobra's unknown-command error, got: %v", err)
	}
}
