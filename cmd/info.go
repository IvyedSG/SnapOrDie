package cmd

import (
	"fmt"
	"time"

	"github.com/sago/snapordie/internal/output"
	"github.com/sago/snapordie/internal/snapshot"
	"github.com/spf13/cobra"
)

var infoCmd = &cobra.Command{
	Use:     "info <name>",
	Aliases: []string{"show"},
	Short:   "Show snapshot details",
	Long: `Display detailed information about a saved snapshot.

Examples:
  snapordie info bug-1234
  snapordie show bug-1234`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		withContainer(containerFlag, func(cName, dataDir string) {
			man := snapshot.NewManager(dataDir, cName)
			snap, err := man.Info(name)
			if err != nil {
				output.Errorf("Snapshot %q not found", name)
				output.Infof("Run %q to see available snapshots", "snapordie list")
				return
			}

			output.Headerf("Snapshot  %s", snap.Name)
			output.Infof("Created    %s", snap.Created.Format(time.RFC822))
			output.Infof("Size       %s", snapshot.HumanSize(snap.SizeBytes))
			output.Infof("Container  %s", snap.Container)
			output.Infof("Data dir   %s", snap.DataDir)
			output.Infof("Location   %s/.snapordie/%s", snap.DataDir, snap.Name)

			fmt.Fprintln(output.Writer())
			output.Infof("Run %q to restore", "snapordie reset "+snap.Name)
		})
		return nil
	},
}
