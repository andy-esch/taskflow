package cli

import (
	"strings"
	"testing"

	"github.com/andy-esch/taskflow/internal/config"
)

func TestUIRefusesNonInteractiveAndDryRunInvocation(t *testing.T) {
	repo := t.TempDir()
	if _, err := config.Init(repo, "", false); err != nil {
		t.Fatal(err)
	}
	_, err := runRootRC(t, "-C", repo, "ui")
	if err == nil || ExitCode(err) != 11 || !strings.Contains(err.Error(), "interactive terminal") {
		t.Fatalf("non-interactive ui should fail clearly with validation: %v", err)
	}
	_, err = runRootRC(t, "-C", repo, "ui", "--dry-run")
	if err == nil || ExitCode(err) != 11 || !strings.Contains(err.Error(), "dry-run") {
		t.Fatalf("ui dry-run should be rejected explicitly: %v", err)
	}
}
