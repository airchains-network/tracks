package keys

import (
	logs "github.com/airchains-network/tracks/log"
	"github.com/airchains-network/tracks/utils"
	"github.com/spf13/cobra"
)

var JunctionKeyImportCmd = &cobra.Command{
	Use:   "import",
	Short: "Import the junction keys needed by the sequencer nodes",
	Run:   runImportKeyCommand,
}

func runImportKeyCommand(cmd *cobra.Command, _ []string) {

	accountFlag := cmd.Flag("accountName")
	if accountFlag == nil {
		logs.Log.Error(`flag "accountName" does not exist`)
		return
	}

	pathFlag := cmd.Flag("accountPath")
	if pathFlag == nil {
		logs.Log.Error(`flag "accountPath" does not exist`)
		return
	}

	mnemonicFlag := cmd.Flag("mnemonic")
	if mnemonicFlag == nil {
		logs.Log.Error(`flag "mnemonic" does not exist`)
		return
	}

	accountName := accountFlag.Value.String()
	accountPath := pathFlag.Value.String()
	mnemonic := mnemonicFlag.Value.String()

	utils.ImportAccountByMnemonic(accountName, accountPath, mnemonic)
}
