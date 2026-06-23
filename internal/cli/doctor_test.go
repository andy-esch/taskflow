package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
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
