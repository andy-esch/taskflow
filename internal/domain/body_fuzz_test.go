package domain

import (
	"strings"
	"testing"
)

// These fuzz targets feed random input into the hand-rolled line/index parsers in
// body.go — Section, the acceptance scanner/flip, and the lint guard. That's the
// panic-prone surface (index math on untrusted markdown; a real off-by-one in
// flipCheckbox already slipped through once). Beyond no-panic they assert the
// load-bearing invariants: the tally and the list agree, a flip preserves the
// criterion count and sets the target state and is idempotent, and Section never
// fabricates a line. The seed corpus also runs under a normal `go test`.

func bodySeeds() []string {
	return []string{
		"",
		"# Title\n\n## Acceptance criteria\n\n- [ ] a\n- [x] b\n- [X] c\n",
		"## Acceptance criteria\r\n\r\n- [x] crlf\r\n- [ ] two\r\n",
		"## Acceptance criteria\n\n```\n- [ ] fenced example\n```\n\n- [x] real\n",
		"## acceptance CRITERIA\n- [] botched\n- [ x] botched2\n- [-] partial\n- [1] cite\n",
		"## Acceptance\n\n- [X] cap only\n",
		"#### deep first\n\n## Acceptance criteria\n\n- [ ] x\n\n## Notes\n\n- [ ] not ac\n",
		"## A\n## B\n## C\n",
		"```go\n## Acceptance criteria\n- [ ] hidden in fence\n```\n",
		"## Acceptance criteria\n\n- [ ] uni é \U0001F600 中文\n",
		"~~~\n## Acceptance criteria\n~~~\n## Acceptance criteria\n- [x] real after fence\n",
		"## Acceptance criteria\n- [ ] a\n## Progress notes on acceptance criteria\n- [ ] b\n",
		"## Acceptance criteria",
		"- [ ] no heading at all\n",
		"#\n##\n###\n",
	}
}

func normLF(s string) string {
	return strings.ReplaceAll(strings.ReplaceAll(s, "\r\n", "\n"), "\r", "\n")
}

func FuzzSection(f *testing.F) {
	for _, s := range bodySeeds() {
		f.Add(s, "acceptance")
		f.Add(s, "")
		f.Add(s, "progress")
	}
	f.Fuzz(func(t *testing.T, body, name string) {
		sec, ok := Section(body, name)
		if !ok {
			if sec != "" {
				t.Fatalf("Section not-ok but returned text %q", sec)
			}
			return
		}
		// A found section never fabricates content: every line it returns is a line of
		// the newline-normalized body.
		have := map[string]bool{}
		for _, ln := range strings.Split(normLF(body), "\n") {
			have[ln] = true
		}
		for _, ln := range strings.Split(sec, "\n") {
			if !have[ln] {
				t.Fatalf("Section fabricated a line %q absent from the body", ln)
			}
		}
	})
}

func FuzzAcceptanceCriteria(f *testing.F) {
	for _, s := range bodySeeds() {
		for _, n := range []int{-1, 0, 1, 2, 1 << 20} {
			f.Add(s, n)
		}
	}
	f.Fuzz(func(t *testing.T, body string, n int) {
		// 1. The tally and the list agree — they share the scanner, so pin it.
		count := CountAcceptanceCriteria(body)
		list := ListAcceptanceCriteria(body)
		if count.Total != len(list) {
			t.Fatalf("Count.Total=%d but len(list)=%d", count.Total, len(list))
		}
		checked := 0
		for _, c := range list {
			if c.Checked {
				checked++
			}
		}
		if count.Checked != checked {
			t.Fatalf("Count.Checked=%d but list has %d checked", count.Checked, checked)
		}

		// 2. Lint never panics.
		_ = LintAcceptanceCriteria(body)

		// 3. A flip either errors cleanly (out-of-range / no section) or preserves the
		//    criterion count, sets the target state, and is idempotent.
		out, err := SetAcceptanceCriterion(body, n, true)
		if err != nil {
			return
		}
		got := ListAcceptanceCriteria(out)
		if len(got) != len(list) {
			t.Fatalf("flip changed the criterion count: %d -> %d", len(list), len(got))
		}
		if n >= 1 && n <= len(got) && !got[n-1].Checked {
			t.Fatalf("flip-to-checked left criterion %d unchecked", n)
		}
		out2, err := SetAcceptanceCriterion(out, n, true)
		if err != nil {
			t.Fatalf("a second identical flip errored: %v", err)
		}
		if out2 != out {
			t.Fatalf("flip to the current state is not idempotent")
		}
	})
}
