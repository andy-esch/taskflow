package domain

import (
	"strings"
	"testing"
)

// Every shape that ParseFindings silently drops must be recognised, and each must
// canonicalise to a header the parser then accepts. The canonical output is fed
// back through ParseFindings so the repair is proven, not just asserted.
func TestNearMissFindingHeaders_CatchesEverySilentLossShape(t *testing.T) {
	cases := []struct {
		name, line, wantCanonical string
	}{
		{"colon after code", "#### H1: a title", "#### H1. a title"},
		{"hyphenated code", "#### H-1. a title", "#### H1. a title"},
		{"lowercase code", "#### h1. a title", "#### H1. a title"},
		{"no period", "#### H1 a title", "#### H1. a title"},
		{"bolded code", "#### **H1.** a title", "#### H1. a title"},
		{"bolded bare code", "#### **H1** a title", "#### H1. a title"},
		{"underscore code", "#### H_1. a title", "#### H1. a title"},
		{"em-dash separator", "#### M2 — a title", "#### M2. a title"},
		{"hyphen separator", "#### M2 - a title", "#### M2. a title"},
		{"multi-letter code", "### BTA-01. a title", "### BTA1. a title"},
		{"depth 2", "## L10: a title", "## L10. a title"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			body := "## Findings\n\n" + tc.line + "\n\n**Status:** open\n"
			if got := ParseFindings(body); len(got) != 0 {
				t.Fatalf("fixture is not actually a near miss: parsed %d findings", len(got))
			}
			hits := NearMissFindingHeaders(body)
			if len(hits) != 1 {
				t.Fatalf("want 1 near-miss hit, got %d", len(hits))
			}
			if hits[0].Canonical != tc.wantCanonical {
				t.Errorf("canonical = %q, want %q", hits[0].Canonical, tc.wantCanonical)
			}
			// The repair must actually make the parser see it.
			fixed, changed := CanonicalizeFindingHeaders(body)
			if len(changed) != 1 {
				t.Fatalf("want 1 repair, got %d", len(changed))
			}
			parsed := ParseFindings(fixed)
			if len(parsed) != 1 {
				t.Fatalf("the repaired body still does not parse: %q", fixed)
			}
			if parsed[0].Status != "open" {
				t.Errorf("repaired finding lost its status: %+v", parsed[0])
			}
		})
	}
}

// The narrowing that makes the warning high-confidence: ordinary document structure
// must never be claimed. These are real heading shapes from the audit corpus.
func TestNearMissFindingHeaders_IgnoresOrdinaryHeadings(t *testing.T) {
	for _, line := range []string{
		"### 1. Lifecycle and Dependency Semantics",
		"### 2. Coverage of Every First-Party Start / Status Path",
		"#### 3.1 Data Flow Architecture",
		"## Context: how it stayed red",
		"## Findings",
		"### Reviewer report",
		"#### H1. already canonical · **Status:** open",
		"## 2026-09-05 review",
		"### 4. Which repairs can be safely inferred?",
	} {
		if hits := NearMissFindingHeaders(line + "\n"); len(hits) != 0 {
			t.Errorf("false positive on %q: %+v", line, hits)
		}
	}
}

// Fenced example syntax is documentation, not a defect — the audit scaffold itself
// ships a fenced `#### H1.` template, and briefs quote drifted shapes on purpose.
func TestNearMissFindingHeaders_IgnoresFencedExamples(t *testing.T) {
	body := "## Findings\n\n```\n#### H-1. a fenced example\n#### h2: another\n```\n\nprose\n"
	if hits := NearMissFindingHeaders(body); len(hits) != 0 {
		t.Errorf("fenced examples must not be flagged, got %+v", hits)
	}
	fixed, changed := CanonicalizeFindingHeaders(body)
	if len(changed) != 0 || fixed != body {
		t.Errorf("a fenced example must not be rewritten:\n%s", fixed)
	}
}

// Repair is idempotent and byte-exact on an already-canonical body: `lint --fix`
// over a clean corpus must be a no-op.
func TestCanonicalizeFindingHeaders_IdempotentOnCanonicalBody(t *testing.T) {
	body := "## Findings\n\n#### H1. a title · **Status:** open\n\nBody.\n\n#### M2. another · **Status:** fixed\n"
	fixed, changed := CanonicalizeFindingHeaders(body)
	if len(changed) != 0 {
		t.Errorf("a canonical body reports repairs: %+v", changed)
	}
	if fixed != body {
		t.Errorf("a canonical body was rewritten:\n%q", fixed)
	}
	again, changed2 := CanonicalizeFindingHeaders(fixed)
	if again != fixed || len(changed2) != 0 {
		t.Error("repair is not idempotent")
	}
}

// The title survives verbatim, including punctuation the code-token repair must not
// touch — an em-dash, a `·` separator, and an inline status.
func TestCanonicalizeFindingHeaders_PreservesTitleAndDepth(t *testing.T) {
	body := "##### h-3: a title — with · punctuation · **Status:** open\n"
	fixed, changed := CanonicalizeFindingHeaders(body)
	if len(changed) != 1 {
		t.Fatalf("want 1 repair, got %d", len(changed))
	}
	want := "##### H3. a title — with · punctuation · **Status:** open\n"
	if fixed != want {
		t.Errorf("got  %q\nwant %q", fixed, want)
	}
	parsed := ParseFindings(fixed)
	if len(parsed) != 1 || parsed[0].Status != "open" || !strings.Contains(parsed[0].Title, "punctuation") {
		t.Errorf("repaired finding lost data: %+v", parsed)
	}
}

// The lint message must name the line and the exact replacement, so a human or an
// agent can act on it without re-deriving the grammar.
func TestLintFindingHeaders_MessageNamesLineAndReplacement(t *testing.T) {
	body := "## Findings\n\nprose\n\n#### H-1. a title\n"
	issues := LintFindingHeaders(body)
	if len(issues) != 1 {
		t.Fatalf("want 1 issue, got %d", len(issues))
	}
	for _, want := range []string{"line 5", `"#### H-1. a title"`, `"#### H1. a title"`} {
		if !strings.Contains(issues[0].Message, want) {
			t.Errorf("message missing %q: %s", want, issues[0].Message)
		}
	}
}
