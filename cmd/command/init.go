package command

import (
	"github.com/airchains-network/decentralized-sequencer/config"
	logs "github.com/airchains-network/decentralized-sequencer/log"
	"github.com/airchains-network/decentralized-sequencer/p2p"
	"github.com/airchains-network/decentralized-sequencer/utilis"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
)

var moniker string
var stationType string
var daType string
var InitCmd = &cobra.Command{
	Use:   "init",
	Short: "initialize the sequencer nodes",
	Run: func(cmd *cobra.Command, args []string) {
		moniker, _ := cmd.Flags().GetString("moniker")
		stationType, _ := cmd.Flags().GetString("stationType")
		daType, _ := cmd.Flags().GetString("daType")

		homeDir, err := os.UserHomeDir()
		if err != nil {
			panic(err) // Handle error appropriately
		}
		tracksDir := filepath.Join(homeDir, config.DefaultTracksDir)
		genFile := filepath.Join(homeDir, config.DefaultGenesisFileName)
		conf := config.DefaultConfig()
		if utilis.FileExists(genFile) {
			logs.Log.Info("Found genesis file")
		} else {

		}
		conf.RootDir = tracksDir
		conf.SetRoot(conf.RootDir)
		config.EnsureRoot(conf.RootDir)
		p2p.InititateIdentity(daType, moniker, stationType)
	},
}
