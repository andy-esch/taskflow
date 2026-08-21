package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/andy-esch/taskflow/internal/userconfig"
	"github.com/andy-esch/taskflow/internal/wire"
)

// linkedPair scaffolds a planning repo and an impl pointing at it. linkBack
// controls whether the planning repo tracks the impl back.
func linkedPair(t *testing.T, linkBack bool) (planning, impl string) {
	t.Helper()
	parent := t.TempDir()
	planning = filepath.Join(parent, "planning")
	impl = filepath.Join(parent, "impl")
	if err := os.MkdirAll(impl, 0o755); err != nil {
		t.Fatal(err)
	}
	runRoot(t, "init", "--path", planning)
	args := []string{"init", "--path", impl, "--planning-repo", "../planning"}
	if !linkBack {
		args = append(args, "--no-link-back")
	}
	runRoot(t, args...)
	return planning, impl
}

func TestDoctor_Clean(t *testing.T) {
	_, impl := linkedPair(t, true)
	out := runRoot(t, "-C", impl, "doctor")
	if !strings.Contains(out, "consistent") {
		t.Errorf("a consistent pair should report clean: %q", out)
	}
}

func TestDoctor_OneSidedExits11(t *testing.T) {
	_, impl := linkedPair(t, false) // one-sided
	out, err := runRootRC(t, "-C", impl, "doctor", "--json")
	if err == nil || ExitCode(err) != 11 {
		t.Fatalf("a one-sided link should exit 11, got %v", err)
	}
	var env struct {
		Problems []struct{ Repo, Message string } `json:"problems"`
	}
	if jerr := json.Unmarshal([]byte(out), &env); jerr != nil {
		t.Fatalf("invalid doctor json: %v\n%s", jerr, out)
	}
	if len(env.Problems) != 1 || !strings.Contains(env.Problems[0].Message, "one-sided") {
		t.Errorf("expected one one-sided problem, got %+v", env.Problems)
	}
}

// TestAmbientLinkWarning: a normal command emits the ⚠ to stderr (never stdout),
// and TSKFLW_NO_LINK_WARN silences it.
func TestAmbientLinkWarning(t *testing.T) {
	_, impl := linkedPair(t, false) // one-sided → a warning to emit

	var out, errOut bytes.Buffer
	cmd := NewRootCmd(strings.NewReader(""), &out, &errOut)
	cmd.SetArgs([]string{"-C", impl, "task", "list", "-q"})
	_ = cmd.Execute()
	if !strings.Contains(errOut.String(), "one-sided") {
		t.Errorf("expected ambient ⚠ on stderr:\n%q", errOut.String())
	}
	if strings.Contains(out.String(), "⚠") {
		t.Errorf("the warning must not land on stdout:\n%q", out.String())
	}

	t.Setenv("TSKFLW_NO_LINK_WARN", "1")
	var out2, errOut2 bytes.Buffer
	cmd2 := NewRootCmd(strings.NewReader(""), &out2, &errOut2)
	cmd2.SetArgs([]string{"-C", impl, "task", "list", "-q"})
	_ = cmd2.Execute()
	if strings.Contains(errOut2.String(), "one-sided") {
		t.Errorf("TSKFLW_NO_LINK_WARN should suppress the warning:\n%q", errOut2.String())
	}
}

// TestDoctor_NoAmbientDoubleWarn: doctor reports on stdout and must NOT also emit
// the ambient stderr ⚠ (its own PreRunE skips it).
func TestDoctor_NoAmbientDoubleWarn(t *testing.T) {
	_, impl := linkedPair(t, false)
	var out, errOut bytes.Buffer
	cmd := NewRootCmd(strings.NewReader(""), &out, &errOut)
	cmd.SetArgs([]string{"-C", impl, "doctor"})
	_ = cmd.Execute()
	if strings.Contains(errOut.String(), "one-sided") {
		t.Errorf("doctor should not also emit the ambient stderr warning:\n%q", errOut.String())
	}
	if !strings.Contains(out.String(), "one-sided") {
		t.Errorf("doctor should report the problem on stdout:\n%q", out.String())
	}
}

// TestDoctor_RegistrySection proves the home registry is audited beside linkbacks. A
// missing entry is actionable and exits 11; an empty but valid planning repo is healthy,
// remains counted, and does not appear in the problem array.
func TestDoctor_RegistrySection(t *testing.T) {
	spaceConfigHome(t)
	root := initializedSpaceRepo(t)
	empty := initializedSpaceRepo(t)
	missing := filepath.Join(t.TempDir(), "gone")
	for _, space := range []userconfig.Space{
		{ID: "empty", Path: empty},
		{ID: "missing", Path: missing},
	} {
		if added, _, err := userconfig.AddSpace(space, false); err != nil || !added {
			t.Fatalf("register %s: added=%v err=%v", space.ID, added, err)
		}
	}

	out, err := runRootRC(t, "-C", root, "doctor", "--json")
	if err == nil || ExitCode(err) != 11 {
		t.Fatalf("broken registry should exit 11, got err=%v code=%d\n%s", err, ExitCode(err), out)
	}
	var env wire.DoctorEnvelope
	if err := json.Unmarshal([]byte(out), &env); err != nil {
		t.Fatalf("decode doctor json: %v\n%s", err, out)
	}
	if env.Registry.Checked != 2 || len(env.Registry.Problems) != 1 {
		t.Fatalf("registry audit = %+v", env.Registry)
	}
	problem := env.Registry.Problems[0]
	if problem.ID != "missing" || problem.Kind != wire.SpaceStateMissing || problem.Remedy == "" {
		t.Errorf("missing-space diagnosis = %+v", problem)
	}
	spaces, readErr := userconfig.Spaces()
	if readErr != nil || len(spaces) != 2 {
		t.Errorf("doctor auto-forgot an entry: spaces=%v err=%v", spaces, readErr)
	}

	human, humanErr := runRootRC(t, "-C", root, "doctor", "--color=never")
	if humanErr == nil || !strings.Contains(human, "Space registry") ||
		!strings.Contains(human, "missing") || !strings.Contains(human, "remedy:") {
		t.Errorf("human registry diagnosis missing: err=%v\n%s", humanErr, human)
	}
}
