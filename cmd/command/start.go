package command

import (
	"github.com/airchains-network/decentralized-sequencer/blocksync"
	"github.com/airchains-network/decentralized-sequencer/config"
	logs "github.com/airchains-network/decentralized-sequencer/log"
	"github.com/airchains-network/decentralized-sequencer/node"
	"github.com/airchains-network/decentralized-sequencer/node/shared"
	"github.com/spf13/cobra"
)

var StationCmd = &cobra.Command{
	Use:   "start",
	Short: "start the sequencer nodes",
	Run: func(cmd *cobra.Command, args []string) {

		// default values
		conf := config.DefaultConfig()
		if !blocksync.InitDb() {
			logs.Log.Error("Error in initializing db")
			return
		}
		logs.Log.Info("Initialized the database")

		//shared.InitializePodState()
		// node config
		shared.NewNode(conf)

		// start node
		node.Start()
	},
}
