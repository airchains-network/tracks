package command

import (
	"github.com/airchains-network/decentralized-sequencer/config"
	logs "github.com/airchains-network/decentralized-sequencer/log"
	"github.com/airchains-network/decentralized-sequencer/p2p"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
)

var moniker string
var stationType string
var daType string
var daRPC string
var stationRPC string
var stationAPI string
var daKey string

var InitCmd = &cobra.Command{
	Use:   "init",
	Short: "initialize the sequencer nodes",
	Run: func(cmd *cobra.Command, args []string) {

		moniker, _ = cmd.Flags().GetString("moniker")
		stationType, _ = cmd.Flags().GetString("stationType")
		daType, _ = cmd.Flags().GetString("daType")
		daRPC, _ = cmd.Flags().GetString("daRpc")
		daKey, _ = cmd.Flags().GetString("daKey")
		stationRPC, _ = cmd.Flags().GetString("stationRpc")
		stationAPI, _ = cmd.Flags().GetString("stationAPI")

		homeDir, err := os.UserHomeDir()
		if err != nil {
			panic(err) // Handle error appropriately
		}
		tracksDir := filepath.Join(homeDir, config.DefaultTracksDir)

		conf := config.DefaultConfig()
		nodeId := p2p.GeneratePeerID()

		conf.RootDir = tracksDir
		conf.DA.DaType = daType
		conf.DA.DaRPC = daRPC
		conf.DA.DaKey = daKey
		conf.Station.StationType = stationType
		conf.Station.StationRPC = stationRPC
		conf.Station.StationAPI = stationAPI
		conf.P2P.NodeId = nodeId
		conf.SetRoot(conf.RootDir)

		success := config.CreateConfigFile(conf.RootDir, conf)
		if !success {
			logs.Log.Warn("Unable to generate a config file because of the mentioned error. Please check the error and try again.")
			return
		}

		logs.Log.Info("Track initialization successful")
	},
}
