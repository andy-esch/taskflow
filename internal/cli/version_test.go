package cli

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestVersion_Subcommand(t *testing.T) {
	out := runRoot(t, "version") // works with no planning repo
	if !strings.Contains(out, "tskflwctl") {
		t.Errorf("version output missing name: %q", out)
	}
}

func TestVersion_JSON(t *testing.T) {
	out := runRoot(t, "--json", "version")
	var got struct {
		SchemaVersion string `json:"schema_version"`
		Version       string `json:"version"`
	}
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("bad json: %v\n%s", err, out)
	}
	if got.SchemaVersion == "" || got.Version == "" {
		t.Errorf("incomplete version json: %+v", got)
	}
}

func TestVersion_Flag(t *testing.T) {
	// cobra's --version short-circuits before repo discovery, so it works anywhere.
	out := runRoot(t, "--version")
	if !strings.Contains(out, "tskflwctl") {
		t.Errorf("--version output: %q", out)
	}
}

func TestWantColor_Precedence(t *testing.T) {
	// --no-color and --color=never force off even with FORCE_COLOR set.
	t.Setenv("FORCE_COLOR", "1")
	if wantColor("always", true, nil) {
		t.Error("--no-color must override everything")
	}
	if wantColor("never", false, nil) {
		t.Error("--color=never must be off")
	}
	// FORCE_COLOR turns auto on even for a non-terminal writer.
	if !wantColor("auto", false, nil) {
		t.Error("FORCE_COLOR should force color on in auto mode")
	}
}

func TestWantColor_NoColorEnv(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	if wantColor("auto", false, nil) {
		t.Error("NO_COLOR should disable color in auto mode")
	}
	// An explicit --color=always still wins over NO_COLOR.
	if !wantColor("always", false, nil) {
		t.Error("--color=always should win over NO_COLOR")
	}
}
