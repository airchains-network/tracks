package keys

import (
	"github.com/airchains-network/decentralized-sequencer/utilis"
	"github.com/spf13/cobra"
)

var AcountName string
var AccountPath string

var JunctionKeyGenCmd = &cobra.Command{
	Use:   "junction",
	Short: "initialize the junction keys needed by the sequencer nodes",
	Run: func(cmd *cobra.Command, args []string) {
		accountName := cmd.Flag("accountName").Value.String()
		accountPath := cmd.Flag("accountPath").Value.String()
		utilis.CreateAccount(accountName, accountPath)

	},
}
