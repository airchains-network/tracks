package command

import (
	"fmt"
	"github.com/spf13/cobra"
)

var KeyGenCmd = &cobra.Command{
	Use:                "keys",
	Short:              "Initialize the keys sequencer nodes",
	Run:                runKeysCommand,
	DisableFlagParsing: true,
}

func runKeysCommand(cmd *cobra.Command, args []string) {
	const message = `Keyring management commands. These keys may be in any format supported by the
Tendermint crypto library and can be used by light-clients, full nodes, or any other application
that needs to sign with a private key`

	fmt.Println(message)

	// show help for the command
	if err := cmd.Help(); err != nil {
		fmt.Println("Failed to print command help:", err)
	}
}
