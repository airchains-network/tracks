package command

import (
	"fmt"
	"github.com/spf13/cobra"
)

var ProverGenCMD = &cobra.Command{
	Use:   "prover",
	Short: "Select the ZKP for your sequencer",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Some SOME BLAB BLAB")
		cmd.Help()
	},
	DisableFlagParsing: true,
}
