package cli

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/andy-esch/taskflow/internal/domain"
)

// A dirty audit (a closed audit with a still-open finding) makes `audit lint`
// exit 11 — the agent-routable contract, mirroring `lint`.
func TestAuditLint_DirtyExits11(t *testing.T) {
	root := setupRepo(t)
	dir := filepath.Join(root, "audits", "closed")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	body := "---\narea: x\ndate: 2026-01-01\n---\n#### S1. t\n**Status:** open\n"
	if err := os.WriteFile(filepath.Join(dir, "2026-01-01-x.md"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	cmd := NewRootCmd(&out, &out)
	cmd.SetArgs([]string{"-C", root, "audit", "lint"})
	if err := cmd.Execute(); !errors.Is(err, domain.ErrValidation) {
		t.Errorf("a dirty audit should make `audit lint` wrap ErrValidation (exit 11), got %v", err)
	}
}

// A clean repo (no audit findings issues) passes and exits 0.
func TestAuditLint_CleanPasses(t *testing.T) {
	root := setupRepo(t) // tasks only, no audits → nothing to flag
	var out bytes.Buffer
	cmd := NewRootCmd(&out, &out)
	cmd.SetArgs([]string{"-C", root, "audit", "lint"})
	if err := cmd.Execute(); err != nil {
		t.Errorf("a clean repo should pass audit lint, got %v", err)
	}
}
