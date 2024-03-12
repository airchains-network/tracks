package command

import (
	"github.com/airchains-network/decentralized-sequencer/node"
	"github.com/spf13/cobra"
)

var (
	Node *node.NodeS
)

var StationCmd = &cobra.Command{
	Use:   "start",
	Short: "start the sequencer nodes",
	Run: func(cmd *cobra.Command, args []string) {
		Node = node.NewNode(conf, connections, podState)
		Node.Start()
	},
}
