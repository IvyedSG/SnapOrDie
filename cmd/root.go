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
	Short:             "Instant Docker database snapshots",
	SilenceUsage:      true,
	SilenceErrors:     true,
	Long: `SnapOrDie saves and restores Docker database snapshots
using copy-on-write (APFS on macOS, reflink on Linux).

Save a snapshot of your MySQL/MariaDB state, then reset
to it instantly — no more waiting 3 minutes for a SQL import.

Commands:
  save [name]     Save a snapshot
  reset [name]    Reset database to a snapshot
  list            List all snapshots
  info <name>     Show snapshot details
  rm <name>       Delete a snapshot`,
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
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "disable colored output")
	rootCmd.PersistentFlags().StringVar(&containerFlag, "container", "",
		"Docker container name (auto-detected if empty)")

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
			output.Errorf("Container %q not found", name)
			output.Infof("Check the name: docker ps")
			return false
		}
		dataDir := docker.DataDir(c)
		if dataDir == "" {
			output.Errorf("Container %q has no MySQL data dir mounted", name)
			output.Infof("Expected a bind mount for /var/lib/mysql")
			return false
		}
		fn(c.Name, dataDir)
		return true
	}

	c, err := docker.Detect()
	if err != nil {
		output.Errorf("No MySQL/MariaDB container detected")
		output.Infof("Make sure your database container is running:")
		output.Infof("  docker compose up -d mysql")
	fmt.Fprintln(output.Writer())
	output.Infof("Or specify the container manually:")
		output.Infof("  snapordie save --container <name>")
		return false
	}

	dataDir := docker.DataDir(c)
	if dataDir == "" {
		output.Errorf("Container %q has no MySQL data dir mounted", c.Name)
		output.Infof("This container needs a bind mount for /var/lib/mysql")
		return false
	}

	fn(c.Name, dataDir)
	return true
}
