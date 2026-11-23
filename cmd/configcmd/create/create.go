package create

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/sagan/ptool/cmd/configcmd"
	"github.com/sagan/ptool/config"
)

var command = &cobra.Command{
	Use:   "create",
	Short: "Create initial pool config file.",
	Long:  `Create initial pool config file.`,
	Args:  cobra.MatchAll(cobra.ExactArgs(0), cobra.OnlyValidArgs),
	RunE:  create,
}

func init() {
	configcmd.Command.AddCommand(command)
}

func create(cmd *cobra.Command, args []string) error {
	configFile := filepath.Join(config.ConfigDir, config.ConfigFile)
	if config.InShell && cmd.Flags().Changed("config") {
		path, err := cmd.Flags().GetString("config")
		if err != nil {
			return err
		}
		if !filepath.IsAbs(path) {
			path = filepath.Join(config.ConfigDir, path)
		}
		configFile = path
	}
	fmt.Printf("Creating config file %s\n", configFile)
	err := config.CreateDefaultConfigAt(configFile)
	if err == nil {
		fmt.Printf("Successfully created config file. Now edit it to add your sites / clients / others.\n")
	}
	return err
}
