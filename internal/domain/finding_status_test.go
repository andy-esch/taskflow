package domain

import (
	"strings"
	"testing"
)

func TestFindingStatusAuthorityAndSpan(t *testing.T) {
	for _, example := range []string{
		"Inline example: `· **Status:** fixed ` is documentation.",
		"Inline example: `` ` · **Status:** fixed `` is documentation.",
		"> · **Status:** fixed",
		"> **Status:** fixed",
		"The example is · **Status:** fixed",
		"    **Status:** fixed",
		"\t**Status:** fixed",
		"`example\n**Status:** fixed\nend`",
	} {
		for _, eol := range []string{"\n", "\r\n"} {
			body := strings.ReplaceAll("#### H1. Metadata examples\n\n"+example+"\n\n**Status:** open (real metadata)\n", "\n", eol)
			finding := ParseFindings(body)[0]
			if finding.Status != "open" || body[finding.StatusSpan.Start:finding.StatusSpan.End] != "open (real metadata)" {
				t.Fatalf("example became authoritative or shifted the real span: %q %+v", body, finding)
			}
			fixed, err := SetFindingStatus(body, "H1", "fixed")
			if err != nil || fixed != strings.Replace(body, "**Status:** open (real metadata)", "**Status:** fixed", 1) {
				t.Fatalf("status write damaged an example: err=%v\n%s", err, fixed)
			}
			withoutStatus := strings.Replace(body, "**Status:** open (real metadata)", "No metadata here.", 1)
			if got := ParseFindings(withoutStatus)[0]; got.Status != "" || !got.StatusSpan.Empty() {
				t.Fatalf("example manufactured a parsed status: %+v", got)
			}
		}
	}
	const title = "#### H1. Example `· **Status:** fixed` · **Status:** open\n"
	if finding := ParseFindings(title)[0]; finding.Title != "Example `· **Status:** fixed`" || finding.Status != "open" {
		t.Fatalf("inline example damaged title or hid real metadata: %+v", finding)
	}
	const unmatched = "#### H1. Unmatched ` is literal\n\n**Status:** open\n\nLater ` is another paragraph.\n"
	if finding := ParseFindings(unmatched)[0]; finding.Status != "open" {
		t.Fatalf("unmatched delimiter consumed a later paragraph's real metadata: %+v", finding)
	}
	const decorated = "#### H1. Issue · **Status:** tracked by `6g392b0rps7w`\r\n"
	if fixed, err := SetFindingStatus(decorated, "H1", "fixed"); err != nil || fixed != "#### H1. Issue · **Status:** fixed\r\n" {
		t.Fatalf("restamping stranded code-formatted decoration or damaged CRLF: err=%v %q", err, fixed)
	}
}
