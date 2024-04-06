package command

import (
	"github.com/spf13/cobra"
)

var ProverGenCMD = &cobra.Command{
	Use:                "prover",
	Short:              "Select the ZKP for your sequencer",
	DisableFlagParsing: true,
	Run:                runProverCommand,
}

func runProverCommand(cmd *cobra.Command, _ []string) {
	if err := cmd.Help(); err != nil {
		cmd.Println("Unable to display help:", err)
	}
}
