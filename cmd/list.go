package cmd

import (
	"fmt"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/sago/snapordie/internal/output"
	"github.com/sago/snapordie/internal/snapshot"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "Listar snapshots",
	Long: `Muestra todos los snapshots guardados con nombre, fecha y tamaño.

Ejemplos:
  snapordie list
  snapordie ls`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		withContainer(containerFlag, func(cName, dataDir string) {
			man := snapshot.NewManager(dataDir, cName)
			snaps, err := man.List()
			if err != nil || len(snaps) == 0 {
				output.Errorf("No hay snapshots guardados")
				output.Infof("Creá uno con: snapordie save")
				output.Infof("  snapordie save inicial")
				return
			}

			output.Headerf("Snapshots")
			output.Infof("Directorio  %s/.snapordie", dataDir)

			headerStyle := lipgloss.NewStyle().Bold(true)
			t := table.New().
				Headers("Nombre", "Tamaño", "Antigüedad", "Container").
				Border(lipgloss.NormalBorder()).
				BorderHeader(true).
				StyleFunc(func(row, col int) lipgloss.Style {
					if row == table.HeaderRow {
						return headerStyle
					}
					return lipgloss.NewStyle()
				})

			for _, s := range snaps {
				t.Row(s.Name, snapshot.HumanSize(s.SizeBytes),
					humanAge(s.Created), s.Container)
			}

			fmt.Fprintln(output.Writer(), "\n"+t.String())
			output.Infof("Para restaurar: snapordie reset <nombre>")
		})
		return nil
	},
}

func humanAge(t time.Time) string {
	d := time.Since(t).Round(time.Second)
	switch {
	case d.Hours() >= 48:
		return fmt.Sprintf("%.0fd", d.Hours()/24)
	case d.Hours() >= 2:
		return fmt.Sprintf("%.0fh", d.Hours())
	case d.Minutes() >= 2:
		return fmt.Sprintf("%.0fm", d.Minutes())
	default:
		return fmt.Sprintf("%.0fs", d.Seconds())
	}
}
