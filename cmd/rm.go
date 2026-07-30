package cmd

import (
	"github.com/sago/snapordie/internal/output"
	"github.com/sago/snapordie/internal/snapshot"
	"github.com/spf13/cobra"
)

var rmCmd = &cobra.Command{
	Use:     "rm <nombre>",
	Aliases: []string{"del"},
	Short:   "Eliminar un snapshot",
	Long: `Elimina un snapshot guardado permanentemente.

Ejemplos:
  snapordie rm bug-1234
  snapordie del snapshot-viejo`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		withContainer(containerFlag, func(cName, dataDir string) {
			man := snapshot.NewManager(dataDir, cName)
			if err := man.Remove(name); err != nil {
				output.Errorf("Snapshot %q no encontrado", name)
				output.Infof("Ejecutá %q para ver los disponibles", "snapordie list")
				return
			}

			output.Successf("Snapshot %q eliminado", name)
		})
		return nil
	},
}
