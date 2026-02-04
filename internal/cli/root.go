package cli

import (
	"context"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "taskflow",
	Short: "AI-Native Project Management",
	Long: `TaskFlow is a local-first project management tool designed 
to bridge the gap between human intuition and AI automation.`,
	// Run: func(cmd *cobra.Command, args []string) { }, 
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute(ctx context.Context) error {
	return rootCmd.ExecuteContext(ctx)
}

func init() {
	// Global flags can be defined here
}
