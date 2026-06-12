// Command tskflwctl is the local-first planning CLI.
package main

import (
	"os"

	"github.com/andy-esch/taskflow/internal/cli"
)

func main() {
	root := cli.NewRootCmd(os.Stdout, os.Stderr)
	if err := root.Execute(); err != nil {
		// Under --json, errors are a machine-readable envelope on stderr
		// (stdout stays empty on failure); prose otherwise.
		asJSON, _ := root.PersistentFlags().GetBool("json")
		cli.WriteError(os.Stderr, err, asJSON)
		os.Exit(cli.ExitCode(err)) // semantic codes: 10 not-found … 14 conflict
	}
}
