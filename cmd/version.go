package cmd

import (
	"fmt"

	"github.com/sago/snapordie/internal/output"
	"github.com/spf13/cobra"
)

var (
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
)

func shortCommit(c string) string {
	if len(c) > 7 {
		return c[:7]
	}
	return c
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Mostrar versión",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Fprintf(output.Writer(), "snapordie %s (%s) %s\n",
			Version, shortCommit(Commit), Date)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
