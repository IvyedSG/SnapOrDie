package cmd

import (
	"time"

	"github.com/sago/snapordie/internal/docker"
	"github.com/sago/snapordie/internal/output"
	"github.com/sago/snapordie/internal/snapshot"
	"github.com/spf13/cobra"
)

var resetCmd = &cobra.Command{
	Use:     "reset [name]",
	Aliases: []string{"rs"},
	Short:   "Reset database to a snapshot",
	Long: `Restore the database to a previously saved snapshot.

If no name is given, uses the most recent snapshot.

Examples:
  snapordie reset
  snapordie rs before-migration
  snapordie reset bug-1234`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := ""
		if len(args) > 0 {
			name = args[0]
		}

		withContainer(containerFlag, func(cName, dataDir string) {
			output.Headerf("Reset")

			man := snapshot.NewManager(dataDir, cName)
			snaps, err := man.List()
			if err != nil || len(snaps) == 0 {
				output.Errorf("No snapshots found")
				output.Infof("Run %q to create one", "snapordie save")
				output.Infof("  snapordie save initial")
				return
			}

			resetName := name
			if resetName == "" {
				resetName = snaps[len(snaps)-1].Name
				output.Bulletf("Snapshot  %s  (most recent)", resetName)
			} else {
				output.Bulletf("Snapshot  %s", resetName)
			}

			output.Infof("Container %s", cName)

			s := output.NewStep("Stopping container")
			if err := docker.Stop(cName); err != nil {
				s.Fail()
				output.Errorf("stop: %s", err)
				return
			}
			s.Done()

			s = output.NewStep("Restoring snapshot")
			if err := man.Reset(resetName); err != nil {
				s.Fail()
				output.Errorf("reset: %s", err)
				docker.Start(cName)
				return
			}
			s.Done()

			s = output.NewStep("Starting container")
			if err := docker.Start(cName); err != nil {
				s.Fail()
				output.Errorf("start: %s", err)
				return
			}

			if err := docker.WaitForHealthy(cName, 30*time.Second); err != nil {
				output.Infof("health check: %s", err)
			}
			s.Done()

			output.Successf("Database restored to %q", resetName)
			output.Infof("Run migrations if schema changed")
		})
		return nil
	},
}
