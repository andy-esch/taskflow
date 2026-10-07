package domain

import (
	"strings"
	"testing"
)

func TestFindingHeaderClassificationMatrix(t *testing.T) {
	for _, tc := range []struct {
		name, body string
		want       FindingHeaderDisposition
	}{
		{"canonical", "#### H1. Issue · **Status:** open\n", FindingHeaderCanonical},
		{"repairable", "#### BTA-01 — Issue\n\n**Status:** open\n", FindingHeaderRepairable},
		{"long letters", "#### CRITICAL-1. Issue · **Status:** open\n", FindingHeaderRepairable},
		{"long digits", "#### H-1234. Issue · **Status:** open\n", FindingHeaderRepairable},
		{"ambiguous", "## Findings\n\n#### H1 Missing metadata\n", FindingHeaderAmbiguous},
		{"ordinary S3", "### S3 Storage Architecture\n", FindingHeaderOrdinary},
		{"ordinary S3 in findings", "## Findings\n\n### S3 Storage Architecture\n", FindingHeaderOrdinary},
		{"S3 status without finding context", "### S3 Storage Architecture\n\n**Status:** open\n", FindingHeaderAmbiguous},
		{"ordinary V2", "### V2 Migration Guide\n", FindingHeaderOrdinary},
		{"ordinary MP3", "### MP3 Audio Support\n", FindingHeaderOrdinary},
		{"ordinary Top3", "### Top3 Recommendations\n", FindingHeaderOrdinary},
		{"no borrowed status", "### S3 Storage Architecture\n\n#### H1. Real finding · **Status:** open\n", FindingHeaderOrdinary},
		{"no fenced status", "### S3 Storage Architecture\n\n```md\n**Status:** open\n```\n", FindingHeaderOrdinary},
		{"no quoted status", "### S3 Storage Architecture\n\nProse mentions **Status:** open\n", FindingHeaderOrdinary},
		{"inline code is not authority", "## Findings\n#### H1 Example\n\nInline example: `· **Status:** fixed ` is documentation.\n", FindingHeaderAmbiguous},
		{"quoted dot is not authority", "## Findings\n#### H1 Example\n\n> · **Status:** fixed\n", FindingHeaderAmbiguous},
		{"prose dot is not authority", "## Findings\n#### H1 Example\n\nThe example is · **Status:** fixed\n", FindingHeaderAmbiguous},
		{"heading code is not authority", "## Findings\n#### H1 Example `· **Status:** fixed`\n", FindingHeaderAmbiguous},
		{"indented code is not authority", "## Findings\n#### H1 Example\n\n    **Status:** fixed\n", FindingHeaderAmbiguous},
		{"multiline code is not authority", "## Findings\n#### H1 Example\n\n`example\n**Status:** fixed\nend`\n", FindingHeaderAmbiguous},
	} {
		t.Run(tc.name, func(t *testing.T) {
			classes := ClassifyFindingHeaders(tc.body)
			if len(classes) == 0 || classes[0].Disposition != tc.want {
				t.Fatalf("classification=%+v want=%s", classes, tc.want)
			}
			fixed, changes := CanonicalizeFindingHeaders(tc.body)
			if tc.want == FindingHeaderRepairable {
				if len(changes) != 1 || len(ParseFindings(fixed)) != 1 {
					t.Fatalf("repair did not produce one canonical finding: %q %+v", fixed, changes)
				}
			} else if fixed != tc.body || len(changes) != 0 {
				t.Fatalf("non-repairable content changed: %q %+v", fixed, changes)
			}
			if tc.want == FindingHeaderOrdinary && len(LintFindingHeaders(tc.body)) != 0 {
				t.Fatal("ordinary prose caused a diagnostic")
			}
		})
	}
}

func TestFindingHeaderRepairPreservesCRLFAndRefusesDuplicateCanonicalization(t *testing.T) {
	body := "# Audit\r\n\r\n#### h-1: Issue · **Status:** open\r\n\r\nKeep prose.\r\n"
	fixed, changes := CanonicalizeFindingHeaders(body)
	if fixed != strings.Replace(body, "#### h-1:", "#### H1.", 1) || len(changes) != 1 {
		t.Fatalf("repair changed bytes outside header: %q %+v", fixed, changes)
	}
	for _, duplicate := range []string{
		"#### H1. Real · **Status:** fixed\n\n#### h-1: Duplicate · **Status:** open\n",
		"#### H-1. First · **Status:** open\n\n#### h1: Second · **Status:** open\n",
	} {
		fixed, changes := CanonicalizeFindingHeaders(duplicate)
		if fixed != duplicate || len(changes) != 0 || len(LintFindingHeaders(duplicate)) == 0 {
			t.Fatalf("duplicate repair was not diagnostic-only: %q %+v", fixed, changes)
		}
	}
}

func TestFindingHeaderWriteGuardUsesEvidenceNotOnlyHeaderText(t *testing.T) {
	before := "## Findings\n\n#### H1 Issue\n"
	if err := NearMissWriteError("audit", IntroducedNearMissHeaders("", before)); err != nil {
		t.Fatalf("diagnostic-only ambiguity blocked a prose write: %v", err)
	}
	after := before + "\n**Status:** open\n"
	if err := NearMissWriteError("audit", IntroducedNearMissHeaders(before, after)); err == nil {
		t.Fatal("adding finding evidence evaded guard through an unchanged heading")
	}
}
