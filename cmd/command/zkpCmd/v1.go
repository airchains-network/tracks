package zkpCmd

import (
	v1 "github.com/airchains-network/decentralized-sequencer/zk/v1"
	"github.com/spf13/cobra"
)

var V1ZKP = &cobra.Command{
	Use:   "v1",
	Short: "initialize the version 1 zero knowledge prover",
	Run: func(cmd *cobra.Command, args []string) {
		v1.CreateVkPkNew()

	},
}
