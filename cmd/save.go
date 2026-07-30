package cmd

import (
	"time"

	"github.com/sago/snapordie/internal/docker"
	"github.com/sago/snapordie/internal/output"
	"github.com/sago/snapordie/internal/snapshot"
	"github.com/spf13/cobra"
)

var saveCmd = &cobra.Command{
	Use:     "save [nombre]",
	Aliases: []string{"sv"},
	Short:   "Guardar un snapshot",
	Long: `Guarda el estado actual de la base de datos como un snapshot.

Si no se especifica nombre, usa la fecha y hora actual.

Ejemplos:
  snapordie save
  snapordie sv antes-de-migration
  snapordie save bug-1234`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := ""
		if len(args) > 0 {
			name = args[0]
		}

		withContainer(containerFlag, func(cName, dataDir string) {
			output.Headerf("Guardar")

			output.Bulletf("Container  %s", cName)
			output.Infof("Data dir   %s", dataDir)

			s := output.NewStep("Deteniendo container")
			if err := docker.Stop(cName); err != nil {
				s.Fail()
				output.Errorf("stop: %s", err)
				return
			}
			s.Done()

			mgr := snapshot.NewManager(dataDir, cName)

			s = output.NewStep("Guardando snapshot")
			snap, err := mgr.Save(name)
			if err != nil {
				s.Fail()
				output.Errorf("%s", err)
				docker.Start(cName)
				return
			}
			s.Donef(snap.Name + "  " + snapshot.HumanSize(snap.SizeBytes))

			s = output.NewStep("Iniciando container")
			if err := docker.Start(cName); err != nil {
				s.Fail()
				output.Errorf("start: %s", err)
				return
			}

			if err := docker.WaitForHealthy(cName, 30*time.Second); err != nil {
				output.Infof("health check: %s", err)
			}
			s.Done()

			output.Successf("Snapshot %q listo", snap.Name)
		})
		return nil
	},
}
