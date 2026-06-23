package cli

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/andy-esch/taskflow/internal/config"
)

func boolp(b bool) *bool { return &b }

// TestPaged_NonTTYPassThrough is the agent-contract pin: when Out is not a real
// TTY (a buffer, a pipe, a redirect), paged writes straight to Out — no subprocess,
// no hang, byte-identical to no pager — even when paging is forced on with
// --paginate. Every buffer-capturing CLI test relies on this.
func TestPaged_NonTTYPassThrough(t *testing.T) {
	for _, tc := range []struct {
		name string
		app  *App
	}{
		{"default", &App{}},
		{"paginate-forced", &App{Paginate: true}},
		{"json", &App{JSON: true}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			tc.app.Out = &buf
			err := tc.app.paged(func(w io.Writer) error {
				if w != &buf {
					t.Fatalf("paged must hand the original writer through off a TTY, got a wrapper")
				}
				_, e := io.WriteString(w, "hello\n")
				return e
			})
			if err != nil {
				t.Fatalf("paged: %v", err)
			}
			if buf.String() != "hello\n" {
				t.Errorf("output = %q, want %q", buf.String(), "hello\n")
			}
		})
	}
}

func TestPagerActive_GateAlwaysWins(t *testing.T) {
	// A buffer is never a TTY, so the gate is shut regardless of the on/off flags —
	// the machine path can't be coerced into paging.
	var buf bytes.Buffer
	for _, app := range []*App{
		{Out: &buf, Paginate: true},
		{Out: &buf},
		{Out: &buf, JSON: true},
		{Out: &buf, NoInput: true},
	} {
		if app.pagerActive() {
			t.Errorf("pagerActive must be false off a TTY (app=%+v)", app)
		}
	}
}

func TestPagerWanted(t *testing.T) {
	cfg := func(enabled *bool) *config.Config {
		return &config.Config{Pager: config.PagerConfig{Enabled: enabled}}
	}
	for _, tc := range []struct {
		name string
		app  App
		want bool
	}{
		{"default (no flags, no config) → on", App{}, true},
		{"nil config → on", App{Cfg: cfg(nil)}, true},
		{"config enabled=true → on", App{Cfg: cfg(boolp(true))}, true},
		{"config enabled=false → off", App{Cfg: cfg(boolp(false))}, false},
		{"--no-pager beats config on", App{NoPager: true, Cfg: cfg(boolp(true))}, false},
		{"--paginate beats config off", App{Paginate: true, Cfg: cfg(boolp(false))}, true},
		{"--no-pager beats --paginate", App{NoPager: true, Paginate: true}, false},
	} {
		if got := tc.app.pagerWanted(); got != tc.want {
			t.Errorf("%s: pagerWanted = %v, want %v", tc.name, got, tc.want)
		}
	}
}

// TestPipeToPager_StreamsToSubprocess exercises the real spawn/pipe/wait path
// (which the TTY gate hides from the buffer-based tests): a capturing "pager" must
// receive exactly what render wrote to its stdin.
func TestPipeToPager_StreamsToSubprocess(t *testing.T) {
	capture := filepath.Join(t.TempDir(), "out.txt")
	app := &App{Out: io.Discard, ErrOut: io.Discard}
	err := app.pipeToPager("cat > "+capture, func(w io.Writer) error {
		_, e := io.WriteString(w, "paged body\n")
		return e
	})
	if err != nil {
		t.Fatalf("pipeToPager: %v", err)
	}
	got, err := os.ReadFile(capture)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "paged body\n" {
		t.Errorf("pager stdin = %q, want %q", got, "paged body\n")
	}
}

// TestPipeToPager_BrokenPipeIsNotAnError pins the quit-early behavior: a pager that
// reads nothing and exits (here `true`) makes our writes fail with EPIPE once the
// kernel pipe buffer fills — that's the user quitting, not a command failure.
func TestPipeToPager_BrokenPipeIsNotAnError(t *testing.T) {
	app := &App{Out: io.Discard, ErrOut: io.Discard}
	big := strings.Repeat("x", 1<<20) // 1 MiB > pipe buffer, so a write will hit EPIPE
	err := app.pipeToPager("true", func(w io.Writer) error {
		_, e := io.WriteString(w, big)
		return e
	})
	if err != nil {
		t.Errorf("a pager that exits early (broken pipe) must not surface an error, got %v", err)
	}
}

// TestPipeToPager_StdoutReachesOut closes the loop: the pager's stdout is wired to
// a.Out, so a pass-through "pager" (cat) echoes our rendered bytes onto Out — in
// production that Out is the terminal.
func TestPipeToPager_StdoutReachesOut(t *testing.T) {
	var out bytes.Buffer
	app := &App{Out: &out, ErrOut: io.Discard}
	err := app.pipeToPager("cat", func(w io.Writer) error {
		_, e := io.WriteString(w, "kept\n")
		return e
	})
	if err != nil {
		t.Fatalf("pipeToPager: %v", err)
	}
	if out.String() != "kept\n" {
		t.Errorf("pager stdout → Out = %q, want %q", out.String(), "kept\n")
	}
}

func TestPagerProgram(t *testing.T) {
	cfgCmd := func(cmd string) *config.Config {
		return &config.Config{Pager: config.PagerConfig{Command: cmd}}
	}
	t.Run("default less -FRX", func(t *testing.T) {
		t.Setenv("TSKFLW_PAGER", "")
		t.Setenv("PAGER", "")
		if got := (&App{}).pagerProgram(); got != "less -FRX" {
			t.Errorf("default = %q, want %q", got, "less -FRX")
		}
	})
	t.Run("$PAGER over default", func(t *testing.T) {
		t.Setenv("TSKFLW_PAGER", "")
		t.Setenv("PAGER", "more")
		if got := (&App{}).pagerProgram(); got != "more" {
			t.Errorf("= %q, want %q", got, "more")
		}
	})
	t.Run("[pager].command over $PAGER", func(t *testing.T) {
		t.Setenv("TSKFLW_PAGER", "")
		t.Setenv("PAGER", "more")
		if got := (&App{Cfg: cfgCmd("bat")}).pagerProgram(); got != "bat" {
			t.Errorf("= %q, want %q", got, "bat")
		}
	})
	t.Run("TSKFLW_PAGER wins over all", func(t *testing.T) {
		t.Setenv("TSKFLW_PAGER", "delta")
		t.Setenv("PAGER", "more")
		if got := (&App{Cfg: cfgCmd("bat")}).pagerProgram(); got != "delta" {
			t.Errorf("= %q, want %q", got, "delta")
		}
	})
}
