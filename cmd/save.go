package cmd

import (
	"time"

	"github.com/sago/snapordie/internal/docker"
	"github.com/sago/snapordie/internal/output"
	"github.com/sago/snapordie/internal/snapshot"
	"github.com/spf13/cobra"
)

var saveCmd = &cobra.Command{
	Use:     "save [name]",
	Aliases: []string{"sv"},
	Short:   "Save a database snapshot",
	Long: `Save the current database state as a snapshot.

If no name is given, uses the current timestamp.

Examples:
  snapordie save
  snapordie sv before-migration
  snapordie save bug-1234`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := ""
		if len(args) > 0 {
			name = args[0]
		}

		withContainer(containerFlag, func(cName, dataDir string) {
			output.Headerf("Save")

			output.Bulletf("Container  %s", cName)
			output.Infof("Data dir   %s", dataDir)

			s := output.NewStep("Stopping container")
			if err := docker.Stop(cName); err != nil {
				s.Fail()
				output.Errorf("stop: %s", err)
				return
			}
			s.Done()

			mgr := snapshot.NewManager(dataDir, cName)

			s = output.NewStep("Saving snapshot")
			snap, err := mgr.Save(name)
			if err != nil {
				s.Fail()
				output.Errorf("%s", err)
				docker.Start(cName)
				return
			}
			s.Donef(snap.Name + "  " + snapshot.HumanSize(snap.SizeBytes))

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

			output.Successf("Snapshot %q ready", snap.Name)
		})
		return nil
	},
}
