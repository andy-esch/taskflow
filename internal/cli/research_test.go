package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// researchPath resolves the flat id-led file a CLI-created research doc landed at
// (research/<minted-id>-<slug>.md) — the id is minted from `created`, so the file is
// found by its slug suffix.
func researchPath(t *testing.T, root, slug string) string {
	t.Helper()
	m, err := filepath.Glob(filepath.Join(root, "research", "*-"+slug+".md"))
	if err != nil {
		t.Fatalf("glob research %q: %v", slug, err)
	}
	if len(m) != 1 {
		t.Fatalf("expected exactly one research file for %q, got %v", slug, m)
	}
	return m[0]
}

func TestResearchNew_CreatesIDLedFile(t *testing.T) {
	root := freshRepo(t)
	runRoot(t, "-C", root, "research", "new", "Compare theming libraries",
		"--tags", "tui,color", "--description", "Weighed three libs", "--created", "2026-06-23")

	path := researchPath(t, root, "compare-theming-libraries")
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	got := string(content)
	for _, want := range []string{`created: "2026-06-23"`, "description: Weighed three libs", "tags: [tui, color]", "# Compare theming libraries"} {
		if !strings.Contains(got, want) {
			t.Errorf("created file missing %q:\n%s", want, got)
		}
	}
	// The absences ARE the contract (epic 28): no lifecycle, no cross-references.
	for _, absent := range []string{"status:", "bucket:", "epic:"} {
		if strings.Contains(got, absent) {
			t.Errorf("research must not carry %q:\n%s", absent, got)
		}
	}
}

// `init` scaffolds research/ alongside the other entity dirs, so a fresh repo can
// take a research doc without a manual mkdir.
func TestInit_ScaffoldsResearchDir(t *testing.T) {
	root := freshRepo(t)
	if fi, err := os.Stat(filepath.Join(root, "research")); err != nil || !fi.IsDir() {
		t.Errorf("init must scaffold research/: err=%v", err)
	}
}

func TestResearchList_NewestFirstAndJSON(t *testing.T) {
	root := freshRepo(t)
	runRoot(t, "-C", root, "research", "new", "Older work", "--created", "2026-01-03")
	runRoot(t, "-C", root, "research", "new", "Newer work", "--created", "2026-06-23")

	// -o name is the byte-stable order check: newest first.
	out := runRoot(t, "-C", root, "research", "list", "-o", "name")
	want := "newer-work\nolder-work\n"
	if out != want {
		t.Errorf("list -o name = %q, want %q", out, want)
	}

	js := runRoot(t, "-C", root, "research", "list", "--json")
	var env struct {
		SchemaVersion string `json:"schema_version"`
		Research      []struct {
			ID, Slug, Created string
		} `json:"research"`
	}
	if err := json.Unmarshal([]byte(js), &env); err != nil {
		t.Fatalf("invalid --json: %v\n%s", err, js)
	}
	if len(env.Research) != 2 {
		t.Fatalf("want 2 docs, got %d: %s", len(env.Research), js)
	}
	if env.Research[0].Slug != "newer-work" {
		t.Errorf("--json order = %q first, want newer-work", env.Research[0].Slug)
	}
	// Ids are minted from created, so id order matches date order.
	if env.Research[1].ID >= env.Research[0].ID {
		t.Errorf("id order must follow created order: %+v", env.Research)
	}
	if env.SchemaVersion == "" {
		t.Error("envelope missing schema_version")
	}
}

func TestResearchList_TagFilter(t *testing.T) {
	root := freshRepo(t)
	runRoot(t, "-C", root, "research", "new", "Tui doc", "--tags", "tui")
	runRoot(t, "-C", root, "research", "new", "Core doc", "--tags", "core")

	out := runRoot(t, "-C", root, "research", "list", "--tag", "tui", "-o", "name")
	if out != "tui-doc\n" {
		t.Errorf("--tag tui = %q, want tui-doc", out)
	}
}

func TestResearchShow_AndPath(t *testing.T) {
	root := freshRepo(t)
	runRoot(t, "-C", root, "research", "new", "Storage model options", "--created", "2026-01-15")

	// Fuzzy prefix resolution, like tasks/audits.
	out := runRoot(t, "-C", root, "research", "show", "storage-model")
	if !strings.Contains(out, "storage-model-options") || !strings.Contains(out, "2026-01-15") {
		t.Errorf("show missing metadata:\n%s", out)
	}

	// Compared by basename: `research path` emits the symlink-resolved absolute path
	// (/private/var/… on macOS) while the glob returns the unresolved /var/… form.
	p := strings.TrimSpace(runRoot(t, "-C", root, "research", "path", "storage-model"))
	if filepath.Base(p) != filepath.Base(researchPath(t, root, "storage-model-options")) {
		t.Errorf("path = %q, want the resolved storage-model-options file", p)
	}
}

func TestResearchShow_FrontmatterOnly(t *testing.T) {
	root := freshRepo(t)
	runRoot(t, "-C", root, "research", "new", "Some doc")

	out := runRoot(t, "-C", root, "research", "show", "some-doc", "--frontmatter-only")
	if strings.Contains(out, "## Question") {
		t.Errorf("--frontmatter-only must omit the body:\n%s", out)
	}
	if !strings.Contains(out, "some-doc") {
		t.Errorf("--frontmatter-only must keep the metadata:\n%s", out)
	}
}

// A non-id-led file in research/ is a FileProblem (the ADR-0003 carveout gate), which
// `research list` must surface rather than silently skip — this is exactly the
// pre-migration state of the 28 legacy docs.
func TestResearchList_NonIDLedFileIsReported(t *testing.T) {
	root := freshRepo(t)
	mustWrite(t, filepath.Join(root, "research", "2026-06-28-legacy.md"), "---\ncreated: 2026-06-28\n---\n# Legacy\n")

	res, err := runRootStreams(t, "-C", root, "research", "list")
	if err == nil {
		t.Fatal("a non-id-led research file must make `research list` exit non-zero")
	}
	// The aggregate error carries the count; the per-file line names the fix. It is
	// a diagnostic, so stderr is where it belongs — asserting the stream is what
	// makes this test able to notice it moving onto the payload stream.
	if !strings.Contains(res.Err, "no leading id") || !strings.Contains(res.Err, "meta/") {
		t.Errorf("stderr should name the carveout gate and the fix, got:\n%s", res.Err)
	}
	if strings.Contains(res.Out, "no leading id") {
		t.Errorf("the diagnostic must not reach the payload stream:\n%s", res.Out)
	}
}

func TestResearchNew_DryRunWritesNothing(t *testing.T) {
	root := freshRepo(t)
	runRoot(t, "-C", root, "--dry-run", "research", "new", "Not written")

	m, _ := filepath.Glob(filepath.Join(root, "research", "*.md"))
	if len(m) != 0 {
		t.Errorf("dry run wrote files: %v", m)
	}
}

// The ambiguous-match error is user-facing prose: "research" is a mass noun, so it must
// not read "matches 2 researchs".
func TestResearchShow_AmbiguousWordingIsNotResearchs(t *testing.T) {
	root := freshRepo(t)
	runRoot(t, "-C", root, "research", "new", "Dup doc", "--created", "2026-02-02")
	runRoot(t, "-C", root, "research", "new", "Dup doc", "--created", "2026-02-02")

	res, err := runRootStreams(t, "-C", root, "research", "show", "dup-doc")
	if err == nil {
		t.Fatal("a duplicate slug must be ambiguous")
	}
	// The ambiguity surfaces as the RETURNED error (the exit-code path) rather than
	// on either stream. The harness can now say so instead of merging both and
	// hoping: assert the wording where it actually lives, and that the payload
	// stream stays empty of it.
	if strings.Contains(err.Error(), "researchs") {
		t.Errorf("mass-noun plural is wrong:\n%s", err.Error())
	}
	if !strings.Contains(err.Error(), "matches 2 research docs") {
		t.Errorf("want 'matches 2 research docs', got:\n%s", err.Error())
	}
	if strings.Contains(res.Out, "researchs") {
		t.Errorf("mass-noun plural leaked onto stdout:\n%s", res.Out)
	}
}

func TestResearchSet_UpdatesFieldsSurgically(t *testing.T) {
	root := freshRepo(t)
	runRoot(t, "-C", root, "research", "new", "Theming libs", "--created", "2026-06-23")
	path := researchPath(t, root, "theming-libs")
	// A vestigial key like the legacy corpus carries, added by hand.
	orig, _ := os.ReadFile(path)
	if err := os.WriteFile(path, []byte(strings.Replace(string(orig), "tags: []", "tags: []\nstatus: reference", 1)), 0o644); err != nil {
		t.Fatal(err)
	}

	runRoot(t, "-C", root, "research", "set", "theming-libs", "--description", "Weighed three libs", "--tags", "tui,color")

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	s := string(got)
	for _, want := range []string{"description: Weighed three libs", "tags: [tui, color]", "status: reference", "updated_at:"} {
		if !strings.Contains(s, want) {
			t.Errorf("missing %q after set:\n%s", want, s)
		}
	}
	// created is untouched, and so is the id it was minted from.
	if !strings.Contains(s, `created: "2026-06-23"`) {
		t.Errorf("created must not change:\n%s", s)
	}
}

// `created` is immutable because the id encodes it — the single most important guard on
// this command, so it is asserted at the CLI boundary too.
func TestResearchSet_CreatedIsNotSettable(t *testing.T) {
	root := freshRepo(t)
	runRoot(t, "-C", root, "research", "new", "Doc", "--created", "2026-06-23")

	out, err := runRootRC(t, "-C", root, "research", "set", "doc", "--set", "created=2020-01-01")
	if err == nil {
		t.Fatal("setting created must fail")
	}
	msg := err.Error() + out
	if !strings.Contains(msg, "created") || !strings.Contains(msg, "id") {
		t.Errorf("error should explain the id/created coupling: %s", msg)
	}
	// And the file is unchanged.
	got, _ := os.ReadFile(researchPath(t, root, "doc"))
	if !strings.Contains(string(got), `created: "2026-06-23"`) {
		t.Errorf("created changed despite the refusal:\n%s", got)
	}
}

func TestResearchAppend_AddsToBodyAndStamps(t *testing.T) {
	root := freshRepo(t)
	runRoot(t, "-C", root, "research", "new", "Doc", "--created", "2026-06-23")

	runRoot(t, "-C", root, "research", "append", "doc", "--body", "## Addendum\n\nlipgloss v2 shipped.")

	got, _ := os.ReadFile(researchPath(t, root, "doc"))
	s := string(got)
	if !strings.Contains(s, "## Addendum") || !strings.Contains(s, "## Question") {
		t.Errorf("append should add to the body, not replace it:\n%s", s)
	}
	if !strings.Contains(s, "updated_at:") {
		t.Errorf("append should stamp updated_at:\n%s", s)
	}
}

// `research edit` is interactive with no preview, so it must refuse --dry-run rather than
// open an editor whose save is silently discarded.
func TestResearchEdit_RejectsDryRunAndNonTTY(t *testing.T) {
	root := freshRepo(t)
	runRoot(t, "-C", root, "research", "new", "Doc")

	// Assert the MESSAGE names dry-run: the harness has no TTY, so the Gate check errors
	// anyway and a bare "did it error?" assertion passes even with the guard deleted.
	dryOut, dryErr := runRootRC(t, "-C", root, "--dry-run", "research", "edit", "doc")
	if dryErr == nil {
		t.Fatal("`research edit --dry-run` must be rejected")
	}
	if msg := dryErr.Error() + dryOut; !strings.Contains(msg, "dry-run") {
		t.Errorf("the refusal must name --dry-run, not just any error: %s", msg)
	}
	// Non-interactive (test harness has no TTY): must point at the scriptable path.
	out, err := runRootRC(t, "-C", root, "research", "edit", "doc")
	if err == nil {
		t.Fatal("`research edit` without a TTY must be rejected")
	}
	if msg := err.Error() + out; !strings.Contains(msg, "research set") && !strings.Contains(msg, "research append") {
		t.Errorf("error should point at the non-interactive alternative: %s", msg)
	}
}

func TestResearchSet_DryRunWritesNothing(t *testing.T) {
	root := freshRepo(t)
	runRoot(t, "-C", root, "research", "new", "Doc", "--created", "2026-06-23")
	path := researchPath(t, root, "doc")
	before, _ := os.ReadFile(path)

	runRoot(t, "-C", root, "--dry-run", "research", "set", "doc", "--description", "preview only")

	after, _ := os.ReadFile(path)
	if string(before) != string(after) {
		t.Errorf("--dry-run wrote to disk:\n%s", after)
	}
}

// The mutation envelope must carry dry_run so a preview is distinguishable from a write.
func TestResearchSet_JSONEnvelope(t *testing.T) {
	root := freshRepo(t)
	runRoot(t, "-C", root, "research", "new", "Doc", "--created", "2026-06-23")

	js := runRoot(t, "-C", root, "--dry-run", "research", "set", "doc", "--description", "d", "--json")
	var env struct {
		SchemaVersion string `json:"schema_version"`
		DryRun        bool   `json:"dry_run"`
		Research      struct {
			Slug, Description string
		} `json:"research"`
	}
	if err := json.Unmarshal([]byte(js), &env); err != nil {
		t.Fatalf("invalid --json: %v\n%s", err, js)
	}
	if !env.DryRun {
		t.Errorf("dry_run must be true for a preview: %s", js)
	}
	if env.Research.Slug != "doc" || env.SchemaVersion == "" {
		t.Errorf("envelope wrong: %s", js)
	}
}

// Pins that `research new` uses the MINTABLE-range validator, not just the date shape.
// Without this, swapping ValidateMintableDate for ValidateDate in NewResearch leaves the
// whole suite green — the domain-level boundary tests can't see which one the caller uses,
// and a malformed-date case like "2026-6-3" is rejected by both so it can't distinguish.
func TestResearchNew_RejectsUnmintableCreated(t *testing.T) {
	root := freshRepo(t)
	for _, date := range []string{"1026-06-15", "9026-06-15", "1969-12-31"} {
		out, err := runRootRC(t, "-C", root, "research", "new", "Doc "+date, "--created", date)
		if err == nil {
			t.Errorf("--created %s is outside the encodable range and must be rejected", date)
			continue
		}
		if msg := err.Error() + out; !strings.Contains(msg, "outside the range") {
			t.Errorf("--created %s: want the mintable-range error, got %s", date, msg)
		}
	}
	// A shape-valid, in-range date still works.
	runRoot(t, "-C", root, "research", "new", "Fine", "--created", "2026-06-23")
}

// The three input guards on `research set`/`append`, none of which were exercised.
func TestResearchSet_InputGuards(t *testing.T) {
	root := freshRepo(t)
	runRoot(t, "-C", root, "research", "new", "Doc")

	cases := []struct {
		name, wantMsg string
		args          []string
	}{
		{"--set without =", "key=value", []string{"research", "set", "doc", "--set", "novalue"}},
		{"--set with empty key", "key=value", []string{"research", "set", "doc", "--set", "=v"}},
		{"same key set and unset", "both set and unset", []string{"research", "set", "doc", "--description", "x", "--unset", "description"}},
		{"unknown key on unset", "unknown research field", []string{"research", "set", "doc", "--unset", "bogsu"}},
		{"empty append body", "nothing to append", []string{"research", "append", "doc", "--body", "   "}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out, err := runRootRC(t, append([]string{"-C", root}, tc.args...)...)
			if err == nil {
				t.Fatalf("%s must be rejected", tc.name)
			}
			if msg := err.Error() + out; !strings.Contains(msg, tc.wantMsg) {
				t.Errorf("want %q in the error, got %s", tc.wantMsg, msg)
			}
		})
	}
}

// An agent must be able to DISCOVER research's field set from the contract instead of
// triggering an error and parsing prose — the reason epic_fields was added when `epic set`
// landed. research_fields shipped missing for two releases.
func TestSchema_ExposesResearchFields(t *testing.T) {
	js := runRoot(t, "schema", "--json")
	var c struct {
		ResearchFields []struct{ Name, Type string } `json:"research_fields"`
		TaskFields     []struct{ Name, Type string } `json:"task_fields"`
	}
	if err := json.Unmarshal([]byte(js), &c); err != nil {
		t.Fatalf("invalid schema --json: %v", err)
	}
	if len(c.ResearchFields) == 0 {
		t.Fatal("schema --json must expose research_fields")
	}
	got := map[string]string{}
	for _, f := range c.ResearchFields {
		got[f.Name] = f.Type
	}
	// Types matter: an agent writing `--set tags=a,b` needs to know it's a list.
	for name, typ := range map[string]string{"created": "date", "description": "string", "tags": "list", "updated_at": "date"} {
		if got[name] != typ {
			t.Errorf("research_fields[%s] = %q, want %q", name, got[name], typ)
		}
	}
	// Same inclusion rule as task_fields: storage machinery stays out.
	for _, absent := range []string{"id", "schema"} {
		if _, ok := got[absent]; ok {
			t.Errorf("research_fields must not carry %q (task_fields doesn't either)", absent)
		}
	}
}

// The unknown-field error is the other half of discovery, and it must name only fields the
// write path will accept — previously it listed created/id/schema/updated_at, all refused.
func TestResearchSet_UnknownFieldErrorListsOnlySettable(t *testing.T) {
	root := freshRepo(t)
	runRoot(t, "-C", root, "research", "new", "Doc")

	out, err := runRootRC(t, "-C", root, "research", "set", "doc", "--set", "tier=2")
	if err == nil {
		t.Fatal("an unknown field must be rejected without --force")
	}
	msg := err.Error() + out
	if !strings.Contains(msg, "settable: description, tags") {
		t.Errorf("want only the settable fields advertised, got: %s", msg)
	}
	for _, protected := range []string{"created", "updated_at", "schema"} {
		if strings.Contains(msg, protected+",") || strings.Contains(msg, ", "+protected) {
			t.Errorf("error advertises protected field %q, which set refuses: %s", protected, msg)
		}
	}
}
