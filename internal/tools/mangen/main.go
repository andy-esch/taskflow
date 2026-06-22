// Command mangen regenerates the roff manpage for tskflwctl from the cobra
// command tree, using mango — the same generator fang's runtime `man` command
// uses. We generate it here, as a build tool, rather than via that runtime
// command because the styled human surface (and thus fang's `man` command) is
// TTY-gated in main: in CI / goreleaser the `man` command wouldn't exist, so a
// dedicated tool is the reliable path. Mirrors internal/tools/docgen.
//
// Run with `just man` (or `go run ./internal/tools/mangen -out manpages`);
// goreleaser's before-hook regenerates it so every release archive ships an
// up-to-date `tskflwctl.1`. It only introspects command metadata (never executes
// RunE/PersistentPreRunE), so it needs no planning repo and is safe anywhere.
package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"

	mcobra "github.com/muesli/mango-cobra"
	"github.com/muesli/roff"

	"github.com/andy-esch/taskflow/internal/cli"
)

func main() {
	out := flag.String("out", "manpages", "output directory for the generated manpage")
	flag.Parse()

	root := cli.NewRootCmd(os.Stdin, os.Stdout, os.Stderr)
	page, err := mcobra.NewManPage(1, root)
	if err != nil {
		log.Fatalf("mangen: %v", err)
	}
	if err := os.MkdirAll(*out, 0o755); err != nil {
		log.Fatalf("mangen: %v", err)
	}
	path := filepath.Join(*out, "tskflwctl.1")
	if err := os.WriteFile(path, []byte(page.Build(roff.NewDocument())), 0o644); err != nil {
		log.Fatalf("mangen: %v", err)
	}
}
