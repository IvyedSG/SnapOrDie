package cmd

import (
	"fmt"
	"time"

	"github.com/sago/snapordie/internal/output"
	"github.com/sago/snapordie/internal/snapshot"
	"github.com/spf13/cobra"
)

var infoCmd = &cobra.Command{
	Use:     "info <nombre>",
	Aliases: []string{"show"},
	Short:   "Ver detalle de un snapshot",
	Long: `Muestra información detallada de un snapshot guardado.

Ejemplos:
  snapordie info bug-1234
  snapordie show bug-1234`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		withContainer(containerFlag, func(cName, dataDir string) {
			man := snapshot.NewManager(dataDir, cName)
			snap, err := man.Info(name)
			if err != nil {
				output.Errorf("Snapshot %q no encontrado", name)
				output.Infof("Ejecutá %q para ver los disponibles", "snapordie list")
				return
			}

			output.Headerf("Snapshot  %s", snap.Name)
			output.Infof("Creado     %s", snap.Created.Format(time.RFC822))
			output.Infof("Tamaño     %s", snapshot.HumanSize(snap.SizeBytes))
			output.Infof("Container  %s", snap.Container)
			output.Infof("Data dir   %s", snap.DataDir)
			output.Infof("Ubicación  %s/.snapordie/%s", snap.DataDir, snap.Name)

			fmt.Fprintln(output.Writer())
			output.Infof("Para restaurar: snapordie reset %s", snap.Name)
		})
		return nil
	},
}
