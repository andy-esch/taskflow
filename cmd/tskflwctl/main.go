// Command tskflwctl is the local-first planning CLI.
package main

import (
	"fmt"
	"os"

	"github.com/andy-esch/taskflow/internal/cli"
)

func main() {
	root := cli.NewRootCmd(os.Stdout, os.Stderr)
	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(cli.ExitCode(err)) // semantic codes: 10 not-found … 14 conflict
	}
}
