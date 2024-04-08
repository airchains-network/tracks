package zkpCmd

import (
	v1 "github.com/airchains-network/decentralized-sequencer/zk/v1EVM"
	"github.com/spf13/cobra"
)

func runV1ZKPCommand(_ *cobra.Command, _ []string) {
	v1.CreateVkPkNew()
}

var V1ZKP = &cobra.Command{
	Use:   "v1EVM",
	Short: "Initialize the Version 1 Zero Knowledge Prover",
	Run:   runV1ZKPCommand,
}
