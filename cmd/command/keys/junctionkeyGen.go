package keys

import (
	"errors"
	"github.com/airchains-network/tracks/utils"
	"github.com/spf13/cobra"
)

type JunctionArgs struct {
	accountName string
	accountPath string
}

func parseJunctionArgs(cmd *cobra.Command) (*JunctionArgs, error) {
	accountFlag := cmd.Flag("accountName")
	if accountFlag == nil {
		return nil, errors.New(`flag "accountName" does not exist`)
	}

	pathFlag := cmd.Flag("accountPath")
	if pathFlag == nil {
		return nil, errors.New(`flag "accountPath" does not exist`)
	}

	args := &JunctionArgs{
		accountName: cmd.Flag("accountName").Value.String(),
		accountPath: cmd.Flag("accountPath").Value.String(),
	}

	return args, nil
}

var JunctionKeyGenCmd = &cobra.Command{
	Use:   "junction",
	Short: "Initialize the junction keys needed by the sequencer nodes",
	Run:   runJunctionCommand,
}

func runJunctionCommand(cmd *cobra.Command, _ []string) {
	args, err := parseJunctionArgs(cmd)
	if err != nil {
		cmd.Println("Unable to parse command args:", err)
		return
	}
	utils.CreateAccount(args.accountName, args.accountPath)
}
