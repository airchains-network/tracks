package command

import (
	"github.com/airchains-network/decentralized-sequencer/node"
	"github.com/airchains-network/decentralized-sequencer/p2p"
	"github.com/spf13/cobra"
)

var StationCmd = &cobra.Command{
	Use:   "start",
	Short: "start the sequencer nodes",
	Run: func(cmd *cobra.Command, args []string) {
		node.Node()
	},
}
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
		p2p.InititateIdentity(daType, moniker, stationType)
	},
}
