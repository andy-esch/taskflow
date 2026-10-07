package main

import (
	"encoding/json"
	"maps"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/andy-esch/taskflow/internal/testutil"
	"github.com/andy-esch/taskflow/internal/wire"
)

func TestSmoke_InitEmittedRepairCommandRetainsTargetAndPolicy(t *testing.T) {
	for _, selection := range []string{"chdir", "path", "space", "environment"} {
		t.Run(selection, func(t *testing.T) {
			home, caller := t.TempDir(), t.TempDir()
			t.Setenv("TSKFLW_CONFIG_HOME", home)
			t.Setenv("TSKFLW_SPACE", "")
			t.Setenv("TSKFLW_NO_REGISTER", "")
			t.Setenv("PATH", filepath.Dir(binary(t))+string(os.PathListSeparator)+os.Getenv("PATH"))
			// Shell metacharacters in both arguments must remain literal paths.
			target := filepath.Join(t.TempDir(), "selected ' $HOME $(touch leaked-target) `id` ")
			tree := "planning ' $HOME $(touch leaked-root) `id`"
			runFrom := func(args ...string) []byte {
				t.Helper()
				command := exec.Command(binary(t), args...)
				command.Dir = caller
				out, err := command.CombinedOutput()
				if err != nil {
					t.Fatalf("%v: %v\n%s", args, err, out)
				}
				return out
			}
			runFrom("init", "--path", target, "--taskflow-root", tree, "--no-register", "--json")
			if selection == "space" || selection == "environment" {
				runFrom("space", "add", target, "--id", "repair-target")
			}
			threads := filepath.Join(target, tree, "threads")
			if err := os.Remove(filepath.Join(threads, ".gitkeep")); err != nil {
				t.Fatal(err)
			}
			if err := os.Remove(threads); err != nil {
				t.Fatal(err)
			}
			callerBefore, homeBefore := testutil.SnapshotTree(t, caller), testutil.SnapshotTree(t, home)
			targetBefore := testutil.SnapshotTree(t, target)
			args := []string{"init", "--no-register", "--json"}
			switch selection {
			case "chdir":
				args = append(args, "-C", target)
			case "path":
				args = append(args, "--path", target)
			case "space":
				args = append(args, "--space", "repair-target")
			case "environment":
				t.Setenv("TSKFLW_SPACE", "repair-target")
				// Capture opt-out from the environment rather than the flag.
				t.Setenv("TSKFLW_NO_REGISTER", "1")
				args = []string{"init", "--json"}
			}
			var receipt wire.InitEnvelope
			if err := json.Unmarshal(runFrom(args...), &receipt); err != nil {
				t.Fatal(err)
			}
			if !receipt.ScaffoldRepairAvailable || receipt.ScaffoldRepairCommand == "" || !maps.Equal(targetBefore, testutil.SnapshotTree(t, target)) {
				t.Fatalf("topology read mutated target or omitted repair advice: %+v", receipt)
			}
			// Do not help the emitted command by injecting a selector or opt-out.
			// Even different ambient defaults cannot redirect its explicit target.
			t.Setenv("TSKFLW_SPACE", "missing")
			t.Setenv("TSKFLW_NO_REGISTER", "")
			command := exec.Command("sh", "-c", receipt.ScaffoldRepairCommand)
			command.Dir = caller
			if out, err := command.CombinedOutput(); err != nil {
				t.Fatalf("execute emitted advice verbatim: %v\n%s", err, out)
			}
			if _, err := os.Stat(threads); err != nil {
				t.Fatalf("selected target was not repaired: %v", err)
			}
			if !maps.Equal(callerBefore, testutil.SnapshotTree(t, caller)) || !maps.Equal(homeBefore, testutil.SnapshotTree(t, home)) {
				t.Fatal("emitted repair advice changed the caller/registry or executed path contents")
			}
		})
	}
}
