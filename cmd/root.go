package cmd

import (
	"fmt"
	"os"

	"github.com/sago/snapordie/internal/docker"
	"github.com/sago/snapordie/internal/output"
	"github.com/spf13/cobra"
)

var (
	containerFlag string
	noColor       bool
)

var rootCmd = &cobra.Command{
	Use:   "snapordie",
	Short: "Snapshots instantáneos de bases Docker",
	Long: `SnapOrDie guarda y restaura snapshots de bases de datos Docker
usando copy-on-write (APFS en macOS, reflink en Linux).

Guardá el estado de tu MySQL/MariaDB, y volvé a él
al instante — sin esperar 3 minutos por un import SQL.

Comandos:
  save [name]     Guardar un snapshot
  reset [name]    Restaurar base de datos a un snapshot
  list            Listar snapshots
  info <name>     Ver detalle de un snapshot
  rm <name>       Eliminar un snapshot`,
	SilenceUsage:      true,
	SilenceErrors:     true,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		output.Init()
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(output.Writer(), err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "desactivar colores")
	rootCmd.PersistentFlags().StringVar(&containerFlag, "container", "",
		"nombre del container Docker (se auto-detecta si está vacío)")

	rootCmd.AddCommand(saveCmd)
	rootCmd.AddCommand(resetCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(infoCmd)
	rootCmd.AddCommand(rmCmd)
}

func withContainer(name string, fn func(container, dataDir string)) bool {
	if name != "" {
		c, err := docker.Inspect(name)
		if err != nil {
			output.Errorf("Container %q no encontrado", name)
			output.Infof("Verificá el nombre con: docker ps")
			return false
		}
		dataDir := docker.DataDir(c)
		if dataDir == "" {
			output.Errorf("El container %q no tiene un directorio MySQL montado", name)
			output.Infof("Se necesita un bind mount para /var/lib/mysql")
			return false
		}
		fn(c.Name, dataDir)
		return true
	}

	c, err := docker.Detect()
	if err != nil {
		output.Errorf("No se detectó ningún container MySQL/MariaDB corriendo")
		output.Infof("Asegurate de tener la base levantada:")
		output.Infof("  docker compose up -d mysql")
		fmt.Fprintln(output.Writer())
		output.Infof("O especificá el container manualmente:")
		output.Infof("  snapordie save --container <nombre>")
		return false
	}

	dataDir := docker.DataDir(c)
	if dataDir == "" {
		output.Errorf("El container %q no tiene un directorio MySQL montado", c.Name)
		output.Infof("Este container necesita un bind mount para /var/lib/mysql")
		return false
	}

	fn(c.Name, dataDir)
	return true
}
