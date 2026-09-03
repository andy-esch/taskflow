package cli

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// Generated docs prove help is synchronized with the SOURCE STRING; they cannot
// notice that the string itself became false. `audit close` went on promising to
// move audits into a `closed/` directory for weeks after ADR-0003 §4 made bucket a
// frontmatter value and deleted those directories — accurate documentation of an
// obsolete claim. This test reads the help text for meaning instead.
//
// The vocabulary here is narrow on purpose: only status and bucket names spelled as
// DIRECTORIES. The entity directories (tasks/, audits/, research/, threads/) are
// real and must keep appearing.
var retiredLayoutTerms = []string{
	"closed/", "open/", "deferred/", "in-progress/", "next-up/",
	"ready-to-start/", "completed/", "deprecated/",
}

// promisesDirectory reports whether text uses term as a PATH rather than as one
// element of a slash-separated enumeration. The vocabulary words are also written as
// lists — "deferred/wontfix/tracked/n-a", "completed/deprecated/deferred" — and those
// are correct prose, so a check that flags every occurrence is noise rather than a
// signal. A trailing slash followed by another word is a list; followed by a space,
// punctuation, or end of string it is a directory claim.
//
// Narrow on purpose, matching how the repo's own lint rules stay high-confidence.
func promisesDirectory(text, term string) bool {
	for i := 0; ; {
		at := strings.Index(text[i:], term)
		if at < 0 {
			return false
		}
		end := i + at + len(term)
		if end >= len(text) || !isWordByte(text[end]) {
			return true
		}
		i = end
	}
}

func isWordByte(b byte) bool {
	return b == '-' || (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || (b >= '0' && b <= '9')
}

func TestHelp_CarriesNoRetiredLayoutVocabulary(t *testing.T) {
	root := NewRootCmd(strings.NewReader(""), nil, nil)

	var walk func(cmd *cobra.Command)
	walk = func(cmd *cobra.Command) {
		surfaces := map[string]string{
			"Short":   cmd.Short,
			"Long":    cmd.Long,
			"Example": cmd.Example,
		}
		for name, text := range surfaces {
			for _, term := range retiredLayoutTerms {
				if promisesDirectory(text, term) {
					t.Errorf("%s: %s help promises the retired directory %q — status and bucket live in frontmatter (ADR-0003 §4)",
						cmd.CommandPath(), name, term)
				}
			}
		}
		cmd.Flags().VisitAll(func(f *pflag.Flag) {
			for _, term := range retiredLayoutTerms {
				if promisesDirectory(f.Usage, term) {
					t.Errorf("%s: --%s usage promises the retired directory %q", cmd.CommandPath(), f.Name, term)
				}
			}
		})
		for _, child := range cmd.Commands() {
			walk(child)
		}
	}
	walk(root)
}

// The `-c` help said "implies -o table" unconditionally, which reads as "-c means
// table" and steers an agent away from `--json -c` — the documented token-cheap
// machine path, and a combination the resolver explicitly supports.
func TestHelp_ColumnsDoesNotForbidJSONProjection(t *testing.T) {
	root := setupRepo(t)

	out := runRoot(t, "-C", root, "task", "list", "--help")

	if strings.Contains(out, "implies -o table)") {
		t.Errorf("the -c help should not read as an unconditional implication:\n%s", out)
	}
	// The behaviour the wording must not contradict.
	projected := runRoot(t, "-C", root, "--json", "task", "list", "-c", "slug,status")
	if !strings.Contains(projected, "\"slug\"") {
		t.Errorf("--json -c should project an envelope:\n%s", projected)
	}
}

// The audit lifecycle text is derived from the transition registry, so a new verb
// cannot ship with stale or missing prose.
func TestAuditVerbHelp_IsDerivedFromTheRegistry(t *testing.T) {
	root := setupRepo(t)
	out := runRoot(t, "-C", root, "audit", "--help")

	for _, want := range []string{"closed bucket", "open bucket", "deferred bucket"} {
		if !strings.Contains(out, want) {
			t.Errorf("audit help should name the %s destination:\n%s", want, out)
		}
	}
}
