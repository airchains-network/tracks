package command

import (
	"fmt"
	"github.com/spf13/cobra"
)

var KeyGenCmd = &cobra.Command{
	Use:   "keys",
	Short: "initialize the keys sequencer nodes",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Keyring management commands. These keys may be in any format supported by the\nTendermint crypto library and can be used by light-clients, full nodes, or any other application\nthat needs to sign with a private key")

		cmd.Help()
	},
	DisableFlagParsing: true,
}
