//go:build !threadspike

package cli

import "github.com/spf13/cobra"

// addThreadSpikeCommand is intentionally empty in ordinary builds. The manual
// ADR-0006 play surface is available only in binaries built with
// `-tags threadspike`.
func addThreadSpikeCommand(_ *cobra.Command, _ *App) {}
