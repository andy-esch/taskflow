// Command docgen regenerates the CLI reference under docs/cli/ from the cobra
// command tree — one markdown page per command, so the reference can never drift
// from the actual flags/examples. Run it with `just docs` (or
// `go run ./internal/tools/docgen`); CI runs `just docs-check` to fail on drift.
//
// It only introspects command metadata (Use/Short/Long/Flags/Example) — cobra's
// doc generator never executes RunE/PersistentPreRunE — so it needs no planning
// repo and is safe to run anywhere.
package main

import (
	"flag"
	"log"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"

	"github.com/andy-esch/taskflow/internal/cli"
)

func main() {
	out := flag.String("out", "docs/cli", "output directory for the generated reference")
	flag.Parse()

	root := cli.NewRootCmd(os.Stdin, os.Stdout, os.Stderr)
	// Reproducible output: drop cobra's "Auto generated … on <date>" footer so
	// regeneration is a no-op unless a command/flag actually changed (the whole
	// point of the CI drift check). DisableAutoGenTag isn't inherited, so set it
	// on every command in the tree.
	disableAutoGenTag(root)

	if err := os.MkdirAll(*out, 0o755); err != nil {
		log.Fatalf("docgen: %v", err)
	}
	if err := doc.GenMarkdownTree(root, *out); err != nil {
		log.Fatalf("docgen: %v", err)
	}
}

func disableAutoGenTag(c *cobra.Command) {
	c.DisableAutoGenTag = true
	for _, sub := range c.Commands() {
		disableAutoGenTag(sub)
	}
}
