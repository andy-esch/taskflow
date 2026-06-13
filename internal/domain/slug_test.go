package domain

import (
	"strings"
	"testing"
	"unicode/utf8"
)

func TestSlugify(t *testing.T) {
	cases := map[string]string{
		"Shell autocomplete: command + slug completion": "shell-autocomplete-command-slug-completion",
		"  Trim  and   collapse_spaces  ":               "trim-and-collapse-spaces",
		"Add create verbs (task new and epic new)":      "add-create-verbs-task-new-and-epic-new",
		// Punctuation is a WORD BREAK, not silently stripped (allowlist policy,
		// 2026-06-13): UI/UX → ui-ux, not uiux. The old strip-without-break
		// behavior was Python-pm parity; the pm is retired and break-on-punct
		// produces the more readable slug.
		"UPPER/Mixed Case":                 "upper-mixed-case",
		"--leading and trailing--":         "leading-and-trailing",
		`a:b\c*d?e<f>g|h"i/j`:              "a-b-c-d-e-f-g-h-i-j", // Windows-invalid chars become breaks
		"Done.":                            "done",                // no trailing dot (Windows-unsafe)
		"...dots...":                       "dots",
		"don't shorten Tasks' apostrophes": "dont-shorten-tasks-apostrophes",
		"v1.2 point release":               "v1.2-point-release",
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

// TestSlugify_RuneSafeTruncation pins the over-cap cut to a rune boundary: a
// long non-ASCII title must never slice mid-rune into an invalid-UTF-8 slug.
func TestSlugify_RuneSafeTruncation(t *testing.T) {
	long := strings.Repeat("日本語タイトル", 10) // 3-byte runes, no dashes, > 80 bytes
	got := Slugify(long)
	if !utf8.ValidString(got) {
		t.Fatalf("slug must be valid UTF-8, got %q", got)
	}
	if len(got) > 80 {
		t.Errorf("slug should be capped at ~80 bytes, got %d", len(got))
	}
	// A dashed multibyte title still trims back to a dash boundary.
	dashed := strings.Repeat("héllo-", 20)
	got = Slugify(dashed)
	if !utf8.ValidString(got) || strings.HasSuffix(got, "-") {
		t.Errorf("dashed multibyte slug should be valid and dash-trimmed, got %q", got)
	}
}

// TestSlugify_UnicodePunctuation pins the dogfooding bug (2026-06-13): an
// em-dash title produced `…-—-…` because slugPunct is ASCII-only. Unicode
// dashes become word breaks; other unicode punctuation/symbols are stripped.
func TestSlugify_UnicodePunctuation(t *testing.T) {
	cases := map[string]string{
		"Web companion — serve and more": "web-companion-serve-and-more",
		"En–dash and horizontal ― bar":   "en-dash-and-horizontal-bar",
		"“Smart quotes” and ellipsis…":   "smart-quotes-and-ellipsis",
		"Trademark™ and €uro":            "trademark-and-uro",
	}
	for in, want := range cases {
		if got := Slugify(in); got != want {
			t.Errorf("Slugify(%q) = %q, want %q", in, got, want)
		}
	}
}
