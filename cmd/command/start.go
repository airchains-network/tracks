package command

import (
	"github.com/airchains-network/decentralized-sequencer/config"
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
		podState := shared.InitializePodState()

		// TODO: make changes in default values from config.yaml

		// node config
		shared.NewNode(conf, podState)

		// start node
		node.Start()
	},
}
