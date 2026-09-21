package main

import (
	"encoding/json"
	"image/color"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"charm.land/lipgloss/v2"

	"github.com/andy-esch/taskflow/internal/design"
	"github.com/andy-esch/taskflow/internal/theme"
	"github.com/andy-esch/taskflow/internal/wire"
)

// buildBinary compiles the real tskflwctl once per test run. Everything else
// in the suite tests packages in-process; only here are the os.Exit wiring and
// the semantic exit codes exercised through an actual process.
var (
	buildOnce sync.Once
	binPath   string
	buildErr  error
)

func binary(t *testing.T) string {
	t.Helper()
	buildOnce.Do(func() {
		dir, err := os.MkdirTemp("", "tskflwctl-smoke-*")
		if err != nil {
			buildErr = err
			return
		}
		binPath = filepath.Join(dir, "tskflwctl")
		out, err := exec.Command("go", "build", "-o", binPath, ".").CombinedOutput()
		if err != nil {
			buildErr = err
			t.Logf("build output:\n%s", out)
		}
	})
	if buildErr != nil {
		t.Fatalf("build binary: %v", buildErr)
	}
	return binPath
}

// run executes the binary against root, returning combined output and the real
// process exit code.
func run(t *testing.T, root string, args ...string) (string, int) {
	t.Helper()
	cmd := exec.Command(binary(t), append([]string{"-C", root}, args...)...)
	out, err := cmd.CombinedOutput()
	if err == nil {
		return string(out), 0
	}
	if ee, ok := err.(*exec.ExitError); ok {
		return string(out), ee.ExitCode()
	}
	t.Fatalf("run %v: %v\n%s", args, err, out)
	return "", -1
}

// runStreams is the process-boundary helper for the machine contract: successful
// payloads use stdout, fatal --json envelopes use stderr, and neither stream may
// silently absorb the other.
func runStreams(t *testing.T, root string, args ...string) (string, string, int) {
	t.Helper()
	cmd := exec.Command(binary(t), append([]string{"-C", root}, args...)...)
	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err == nil {
		return stdout.String(), stderr.String(), 0
	}
	if ee, ok := err.(*exec.ExitError); ok {
		return stdout.String(), stderr.String(), ee.ExitCode()
	}
	t.Fatalf("run %v: %v\nstdout=%s\nstderr=%s", args, err, stdout.String(), stderr.String())
	return "", "", -1
}

func TestSmoke_PublishedExitTaxonomyMatchesProcessBehavior(t *testing.T) {
	root := t.TempDir()
	if out, code := run(t, root, "init", "--path", root, "--no-register"); code != 0 {
		t.Fatalf("init: exit %d\n%s", code, out)
	}
	if out, code := run(t, root, "epic", "new", "Exit Taxonomy", "--description", "test process exits"); code != 0 {
		t.Fatalf("epic new: exit %d\n%s", code, out)
	}
	for range 2 {
		if out, code := run(t, root, "task", "new", "Same Task", "--epic", "01-exit-taxonomy",
			"--description", "force an ambiguous selector", "--tags", "test"); code != 0 {
			t.Fatalf("task new: exit %d\n%s", code, out)
		}
	}

	firstSpace, secondSpace := t.TempDir(), t.TempDir()
	for _, space := range []string{firstSpace, secondSpace} {
		if out, code := run(t, space, "init", "--path", space, "--no-register"); code != 0 {
			t.Fatalf("space init: exit %d\n%s", code, out)
		}
	}
	const collisionID = "exit-taxonomy-process-test"
	if stdout, stderr, code := runStreams(t, root, "space", "add", firstSpace, "--id", collisionID, "--json"); code != 0 {
		t.Fatalf("first space registration: exit %d\nstdout=%s\nstderr=%s", code, stdout, stderr)
	}
	defer func() {
		if out, code := run(t, root, "space", "forget", collisionID); code != 0 {
			t.Errorf("forget process-test space: exit %d\n%s", code, out)
		}
	}()

	schemaOut, schemaErr, code := runStreams(t, root, "schema", "--json")
	if code != 0 || schemaErr != "" {
		t.Fatalf("schema --json: exit %d\nstdout=%s\nstderr=%s", code, schemaOut, schemaErr)
	}
	var schema wire.SchemaEnvelope
	if err := json.Unmarshal([]byte(schemaOut), &schema); err != nil {
		t.Fatalf("decode schema contract: %v\n%s", err, schemaOut)
	}
	published := make(map[int]wire.SchemaExitCode, len(schema.ExitCodes))
	for _, row := range schema.ExitCodes {
		published[row.Code] = row
	}

	tests := []struct {
		name     string
		args     []string
		wantCode int
		wantName string
	}{
		{"success", []string{"version"}, 0, "ok"},
		{"generic", []string{"--badflag", "--json"}, 1, "error"},
		{"not found", []string{"task", "show", "ghost", "--json"}, 10, "not-found"},
		{"validation", []string{"task", "list", "--status", "bogus", "--json"}, 11, "validation"},
		{"ambiguous", []string{"task", "show", "same-task", "--json"}, 13, "ambiguous"},
		{"conflict", []string{"space", "add", secondSpace, "--id", collisionID, "--json"}, 14, "conflict"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			stdout, stderr, code := runStreams(t, root, tc.args...)
			if code != tc.wantCode {
				t.Fatalf("exit = %d, want %d\nstdout=%s\nstderr=%s", code, tc.wantCode, stdout, stderr)
			}
			row, ok := published[code]
			if !ok || row.State != wire.ExitCodeStateActive || row.Name != tc.wantName {
				t.Fatalf("process exit %d/%q is not the published active row: %+v", code, tc.wantName, row)
			}
			if code == 0 {
				if stderr != "" {
					t.Fatalf("successful command wrote stderr: %q", stderr)
				}
				return
			}
			if stdout != "" {
				t.Fatalf("fatal --json command wrote stdout: %q", stdout)
			}
			var envelope wire.ErrorEnvelope
			if err := json.Unmarshal([]byte(stderr), &envelope); err != nil {
				t.Fatalf("decode error envelope: %v\n%s", err, stderr)
			}
			if envelope.Error.Code != tc.wantName {
				t.Fatalf("error.code = %q, want %q", envelope.Error.Code, tc.wantName)
			}
		})
	}

	reserved, ok := published[12]
	if !ok || reserved.Name != "invalid-transition" || reserved.State != wire.ExitCodeStateReserved {
		t.Fatalf("retired code 12 must remain explicitly reserved: %+v", reserved)
	}

	// Pin a real unclassified filesystem failure to the active 1/error row. This
	// also guards the reservation: filesystem detail must never make code 12 live.
	tasksDir := filepath.Join(root, "tasks")
	if err := os.Chmod(tasksDir, 0o500); err != nil {
		t.Fatal(err)
	}
	stdout, stderr, code := runStreams(t, root, "task", "new", "Blocked Write", "--epic", "01-exit-taxonomy", "--tags", "test", "--json")
	if err := os.Chmod(tasksDir, 0o700); err != nil {
		t.Fatal(err)
	}
	if code != 1 || stdout != "" {
		t.Fatalf("filesystem failure = exit %d, want 1; stdout=%s stderr=%s", code, stdout, stderr)
	}
	var filesystemEnvelope wire.ErrorEnvelope
	if err := json.Unmarshal([]byte(stderr), &filesystemEnvelope); err != nil {
		t.Fatalf("decode filesystem error envelope: %v\n%s", err, stderr)
	}
	if filesystemEnvelope.Error.Code != "error" || filesystemEnvelope.Error.Filesystem == nil {
		t.Fatalf("filesystem failure must be 1/error with typed detail: %+v", filesystemEnvelope.Error)
	}
}

func TestSmoke_LifecycleAndExitCodes(t *testing.T) {
	root := t.TempDir()

	// init scaffolds the tree.
	if out, code := run(t, root, "init", "--path", root); code != 0 {
		t.Fatalf("init: exit %d\n%s", code, out)
	}
	if out, code := run(t, root, "epic", "new", "Smoke Epic", "--description", "smoke"); code != 0 {
		t.Fatalf("epic new: exit %d\n%s", code, out)
	}
	out, code := run(t, root, "task", "new", "Smoke Task",
		"--epic", "01-smoke-epic", "--description", "smoke", "--tags", "smoke", "--json")
	if code != 0 {
		t.Fatalf("task new: exit %d\n%s", code, out)
	}
	var created struct {
		SchemaVersion string `json:"schema_version"`
		Created       struct {
			ID string `json:"id"`
		} `json:"created"`
	}
	if err := json.Unmarshal([]byte(out), &created); err != nil || created.Created.ID == "" {
		t.Fatalf("task new --json should return the resolved id: %v\n%s", err, out)
	}
	slug := created.Created.ID

	// Lifecycle: start → complete, then lint must be clean.
	if out, code := run(t, root, "task", "start", slug); code != 0 {
		t.Fatalf("task start: exit %d\n%s", code, out)
	}
	// The scaffold's acceptance criterion is unticked and unexplained, so completing is
	// refused (exit 11) — the task counterpart of `audit close` refusing while findings
	// are open. Through a real process, so the gate is proven at the exit-code boundary.
	if out, code := run(t, root, "task", "complete", slug); code != 11 {
		t.Fatalf("complete with an unexplained criterion should exit 11, got %d\n%s", code, out)
	}
	if out, code := run(t, root, "task", "ac", slug, "--check", "1"); code != 0 {
		t.Fatalf("task ac --check: exit %d\n%s", code, out)
	}
	if out, code := run(t, root, "task", "complete", slug); code != 0 {
		t.Fatalf("task complete: exit %d\n%s", code, out)
	}
	if out, code := run(t, root, "lint"); code != 0 {
		t.Fatalf("lint after lifecycle: exit %d\n%s", code, out)
	}

	// Semantic exit codes through a real process exit.
	if out, code := run(t, root, "task", "show", "no-such-task"); code != 10 {
		t.Errorf("not-found should exit 10, got %d\n%s", code, out)
	}
	if out, code := run(t, root, "task", "list", "--status", "bogus"); code != 11 {
		t.Errorf("invalid status filter should exit 11, got %d\n%s", code, out)
	}
	if out, code := run(t, root, "task", "move", slug, "limbo"); code != 11 {
		t.Errorf("invalid move target should exit 11, got %d\n%s", code, out)
	}

	// Errors go to stderr prefixed "error:" (main's contract).
	cmd := exec.Command(binary(t), "-C", root, "task", "show", "ghost")
	var stderr strings.Builder
	cmd.Stderr = &stderr
	_ = cmd.Run()
	if !strings.Contains(stderr.String(), "error:") {
		t.Errorf("errors should print to stderr with the error: prefix, got %q", stderr.String())
	}

	// Under --json, the failure is a machine-readable envelope on stderr and
	// stdout stays empty — agents must never parse prose (schema 1.1).
	cmd = exec.Command(binary(t), "-C", root, "task", "show", "ghost", "--json")
	var jsonOut, jsonErr strings.Builder
	cmd.Stdout = &jsonOut
	cmd.Stderr = &jsonErr
	_ = cmd.Run()
	if jsonOut.Len() != 0 {
		t.Errorf("stdout must stay empty on a --json failure, got %q", jsonOut.String())
	}
	var env struct {
		SchemaVersion string `json:"schema_version"`
		Error         struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal([]byte(jsonErr.String()), &env); err != nil {
		t.Fatalf("--json error should be a JSON envelope: %v\n%s", err, jsonErr.String())
	}
	if env.SchemaVersion == "" || env.Error.Code != "not-found" || env.Error.Message == "" {
		t.Errorf("error envelope wrong: %+v", env)
	}

	// Raw argv still carries the output request when cobra stops at an earlier
	// unknown flag. Error formatting must not depend on flag order.
	cmd = exec.Command(binary(t), "-C", root, "--badflag", "--json")
	jsonOut.Reset()
	jsonErr.Reset()
	cmd.Stdout = &jsonOut
	cmd.Stderr = &jsonErr
	_ = cmd.Run()
	if jsonOut.Len() != 0 {
		t.Errorf("stdout must stay empty on an early --json parse failure, got %q", jsonOut.String())
	}
	env = struct {
		SchemaVersion string `json:"schema_version"`
		Error         struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}{}
	if err := json.Unmarshal([]byte(jsonErr.String()), &env); err != nil {
		t.Fatalf("--json after an unknown flag should still select the JSON error envelope: %v\n%s", err, jsonErr.String())
	}
	if env.SchemaVersion == "" || env.Error.Message == "" {
		t.Errorf("early parse error envelope wrong: %+v", env)
	}

	// If a preceding value-taking flag consumes the literal `--`, cobra can
	// still parse a later --json even though the lightweight routing scanner
	// correctly treats `--` as a terminator. The error writer prefers that parsed
	// truth so the non-TTY machine path does not regress.
	cmd = exec.Command(binary(t), "-C", root, "task", "list", "--epic", "--", "--json", "-c", "nope")
	jsonOut.Reset()
	jsonErr.Reset()
	cmd.Stdout = &jsonOut
	cmd.Stderr = &jsonErr
	_ = cmd.Run()
	if jsonOut.Len() != 0 {
		t.Errorf("stdout must stay empty when cobra parsed a later --json, got %q", jsonOut.String())
	}
	if err := json.Unmarshal([]byte(jsonErr.String()), &env); err != nil {
		t.Fatalf("cobra-parsed --json should select the JSON error envelope: %v\n%s", err, jsonErr.String())
	}
}

func TestSmoke_VersionStamp(t *testing.T) {
	out, code := run(t, t.TempDir(), "version")
	if code != 0 || strings.TrimSpace(out) == "" {
		t.Errorf("version: exit %d output %q", code, out)
	}
}

// TestUseFang guards the machine contract at the gate: fang's styled human path
// must be closed for every machine context (non-TTY, or any --json run), so
// piped/agent output falls through to the original path and stays byte-identical.
// This is the load-bearing check — if useFang ever returns true under --json or
// off a TTY, fang could reshape the --json envelope / exit codes. (The
// fall-through path itself is exercised end-to-end by TestSmoke above, which runs
// the binary in a subprocess where stderr is not a TTY.)
func TestUseFang(t *testing.T) {
	cases := []struct {
		name string
		args []string
		tty  bool
		want bool
	}{
		{"tty, no flags", nil, true, true},
		{"tty, plain subcommand", []string{"task", "list"}, true, true},
		{"non-tty closes it (pipe/redirect/CI)", []string{"task", "list"}, false, false},
		{"--json closes it even on a tty", []string{"task", "list", "--json"}, true, false},
		{"--json=true closes it", []string{"--json=true"}, true, false},
		{"--json=1 closes it", []string{"--json=1"}, true, false},
		{"--json=t closes it", []string{"--json=t"}, true, false},
		{"--json=T closes it", []string{"--json=T"}, true, false},
		{"--json=TRUE closes it", []string{"--json=TRUE"}, true, false},
		{"--json=True closes it", []string{"--json=True"}, true, false},
		{"--json=false stays human", []string{"--json=false"}, true, true},
		{"last false spelling restores human path", []string{"--json", "--json=false"}, true, true},
		{"last true spelling closes human path", []string{"--json=false", "--json"}, true, false},
		{"invalid json value stays on cobra validation path", []string{"--json=not-a-bool"}, true, true},
		{"--json on a non-tty", []string{"--json"}, false, false},
		{"literal --json after -- does not count", []string{"task", "new", "--", "--json"}, true, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := useFang(tc.args, tc.tty); got != tc.want {
				t.Errorf("useFang(%q, tty=%v) = %v, want %v", tc.args, tc.tty, got, tc.want)
			}
		})
	}
}

func TestJSONFlagActive(t *testing.T) {
	for _, tc := range []struct {
		name string
		args []string
		want bool
	}{
		{"absent", []string{"task", "list"}, false},
		{"bare", []string{"task", "list", "--json"}, true},
		{"unknown before true", []string{"--badflag", "--json"}, true},
		{"last valid false wins", []string{"--json", "--json=0"}, false},
		{"last valid true wins", []string{"--json=false", "--json=True"}, true},
		{"invalid preserves prior valid value", []string{"--json", "--json=nope"}, true},
		{"invalid alone is not active", []string{"--json=nope"}, false},
		{"terminator", []string{"task", "new", "--", "--json"}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := jsonFlagActive(tc.args); got != tc.want {
				t.Fatalf("jsonFlagActive(%q) = %v, want %v", tc.args, got, tc.want)
			}
		})
	}
}

// TestRepoColorScheme is a smoke test: the scheme builds without panicking and
// sets the accent slots, so a future fang ColorScheme change can't silently drop
// our palette wiring.
func TestRepoColorScheme(t *testing.T) {
	// Simulate a dark terminal: the LightDarkFunc returns the dark variant.
	dark := func(_, dark color.Color) color.Color { return dark }
	cs := repoColorScheme(design.Default(), dark)
	d := design.Default().Dark
	if cs.Title != d.Accent.Color() {
		t.Errorf("Title should be the theme accent (%v), got %v", d.Accent.Color(), cs.Title)
	}
	if cs.Flag != d.Of(theme.ColorGreen).Color() {
		t.Errorf("Flag should be the palette's green, got %v", cs.Flag)
	}
	if cs.ErrorHeader[1] != d.Danger.Color() {
		t.Errorf("error badge bg should be the palette's danger, got %v", cs.ErrorHeader[1])
	}
	if cs.ErrorHeader[0] == nil || cs.Base == nil {
		t.Fatal("badge fg / base left unset")
	}
	// Codeblock is fang's only BACKGROUND role. Assigning it a foreground hue is
	// the defect that made the USAGE box unreadable (DimmedArgument landed on an
	// identically-colored background), so pin it to the surface token.
	if cs.Codeblock != d.Surface.Color() {
		t.Errorf("Codeblock must be the surface token (%v), got %v", d.Surface.Color(), cs.Codeblock)
	}
	if cs.Codeblock == cs.DimmedArgument {
		t.Error("Codeblock (background) equals DimmedArgument (foreground): text is invisible")
	}
	// Base stays the terminal default: fang shares one Base between the help body
	// text and the codeblock text, so pinning it would recolor all help prose.
	if cs.Base != (lipgloss.NoColor{}) {
		t.Errorf("Base should stay the terminal default, got %v", cs.Base)
	}
}

// TestRepoColorSchemeFollowsSelectedTheme is the regression guard for the second
// defect: chrome used design.Default() unconditionally, so a catppuccin user got
// neon help. Every themed slot must move with the passed theme.
func TestRepoColorSchemeFollowsSelectedTheme(t *testing.T) {
	dark := func(_, dark color.Color) color.Color { return dark }
	other, ok := design.Lookup("catppuccin")
	if !ok {
		t.Fatal("catppuccin should be registered")
	}
	def := repoColorScheme(design.Default(), dark)
	alt := repoColorScheme(other, dark)
	for _, c := range []struct {
		name     string
		def, alt color.Color
	}{
		{"Title", def.Title, alt.Title},
		{"Command", def.Command, alt.Command},
		{"Flag", def.Flag, alt.Flag},
		{"Program", def.Program, alt.Program},
		{"Codeblock", def.Codeblock, alt.Codeblock},
	} {
		if c.def == c.alt {
			t.Errorf("%s did not follow the selected theme: both %v", c.name, c.def)
		}
	}
	if alt.Codeblock != other.Dark.Surface.Color() {
		t.Errorf("Codeblock should be catppuccin's surface (%v), got %v", other.Dark.Surface.Color(), alt.Codeblock)
	}
}

// TestRepoColorSchemeFollowsBackground pins the OTHER axis: fang's LightDarkFunc
// picks the background-appropriate variant, so a light terminal must get the light
// surface. The two palettes layer in opposite directions (neon lifts to base01,
// Latte rises to white), so a mapping that ignored the func would be invisible on
// one of them.
func TestRepoColorSchemeFollowsBackground(t *testing.T) {
	onDark := repoColorScheme(design.Default(), func(_, dark color.Color) color.Color { return dark })
	onLight := repoColorScheme(design.Default(), func(light, _ color.Color) color.Color { return light })
	if onDark.Codeblock != design.Default().Dark.Surface.Color() {
		t.Errorf("dark Codeblock = %v, want %v", onDark.Codeblock, design.Default().Dark.Surface.Color())
	}
	if onLight.Codeblock != design.Default().Light.Surface.Color() {
		t.Errorf("light Codeblock = %v, want %v", onLight.Codeblock, design.Default().Light.Surface.Color())
	}
	if onDark.Codeblock == onLight.Codeblock {
		t.Error("Codeblock ignored the background: same surface on light and dark")
	}
}
