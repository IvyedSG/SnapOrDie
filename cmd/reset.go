package cmd

import (
	"time"

	"github.com/sago/snapordie/internal/docker"
	"github.com/sago/snapordie/internal/output"
	"github.com/sago/snapordie/internal/snapshot"
	"github.com/spf13/cobra"
)

var resetCmd = &cobra.Command{
	Use:     "reset [nombre]",
	Aliases: []string{"rs"},
	Short:   "Restaurar base de datos a un snapshot",
	Long: `Restaura la base de datos a un snapshot guardado previamente.

Si no se especifica nombre, usa el más reciente.

Ejemplos:
  snapordie reset
  snapordie rs antes-de-migration
  snapordie reset bug-1234`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := ""
		if len(args) > 0 {
			name = args[0]
		}

		withContainer(containerFlag, func(cName, dataDir string) {
			output.Headerf("Restaurar")

			man := snapshot.NewManager(dataDir, cName)
			snaps, err := man.List()
			if err != nil || len(snaps) == 0 {
				output.Errorf("No hay snapshots guardados")
				output.Infof("Creá uno con: snapordie save")
				output.Infof("  snapordie save inicial")
				return
			}

			resetName := name
			if resetName == "" {
				resetName = snaps[len(snaps)-1].Name
				output.Bulletf("Snapshot  %s  (el más reciente)", resetName)
			} else {
				output.Bulletf("Snapshot  %s", resetName)
			}

			output.Infof("Container %s", cName)

			s := output.NewStep("Deteniendo container")
			if err := docker.Stop(cName); err != nil {
				s.Fail()
				output.Errorf("stop: %s", err)
				return
			}
			s.Done()

			s = output.NewStep("Restaurando snapshot")
			if err := man.Reset(resetName); err != nil {
				s.Fail()
				output.Errorf("reset: %s", err)
				docker.Start(cName)
				return
			}
			s.Done()

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

			output.Successf("Base de datos restaurada a %q", resetName)
			output.Infof("Ejecutá migrations si cambió el schema")
		})
		return nil
	},
}
