package domain

import "testing"

func TestSlugify(t *testing.T) {
	cases := map[string]string{
		"Shell autocomplete: command + slug completion": "shell-autocomplete-command-slug-completion",
		"  Trim  and   collapse_spaces  ":               "trim-and-collapse-spaces",
		"Add create verbs (task new and epic new)":      "add-create-verbs-task-new-and-epic-new",
		"UPPER/Mixed Case":                              "uppermixed-case",
		"--leading and trailing--":                      "leading-and-trailing",
	}
	for in, want := range cases {
		if got := Slugify(in); got != want {
			t.Errorf("Slugify(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestSlugify_TruncatesOnBoundary(t *testing.T) {
	long := "this is a very long title that keeps going well past the eighty character limit for slugs"
	got := Slugify(long)
	if len(got) > 80 {
		t.Errorf("slug too long (%d): %q", len(got), got)
	}
	if got[len(got)-1] == '-' {
		t.Errorf("slug should not end on '-': %q", got)
	}
}
