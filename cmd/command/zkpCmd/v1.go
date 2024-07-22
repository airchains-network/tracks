package zkpCmd

import (
	v1 "github.com/airchains-network/tracks/zk/v1EVM"
	v1Wasm "github.com/airchains-network/tracks/zk/v1WASM"
	"github.com/spf13/cobra"
)

func runV1ZKPCommand(_ *cobra.Command, _ []string) {
	v1.CreateVkPkNew()
}

func runV1WasmZKPCommand(_ *cobra.Command, _ []string) {
	v1Wasm.CreateVkPkWasm()

}

var V1ZKP = &cobra.Command{
	Use:   "v1EVM",
	Short: "Initialize the EVM Version 1  Zero Knowledge Prover",
	Run:   runV1ZKPCommand,
}
var V1ZKPWasm = &cobra.Command{
	Use:   "v1WASM",
	Short: "Initialize the Wasm Version 1  Zero Knowledge Prover",
	Run:   runV1WasmZKPCommand,
}
