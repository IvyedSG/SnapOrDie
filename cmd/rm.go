package cmd

import (
	"github.com/sago/snapordie/internal/output"
	"github.com/sago/snapordie/internal/snapshot"
	"github.com/spf13/cobra"
)

var rmCmd = &cobra.Command{
	Use:     "rm <name>",
	Aliases: []string{"del"},
	Short:   "Delete a snapshot",
	Long: `Delete a saved snapshot permanently.

Examples:
  snapordie rm bug-1234
  snapordie del old-snapshot`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		withContainer(containerFlag, func(cName, dataDir string) {
			man := snapshot.NewManager(dataDir, cName)
			if err := man.Remove(name); err != nil {
				output.Errorf("Snapshot %q not found", name)
				output.Infof("Run %q to see available snapshots", "snapordie list")
				return
			}

			output.Successf("Snapshot %q removed", name)
		})
		return nil
	},
}
