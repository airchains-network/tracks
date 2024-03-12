package main

import (
	"fmt"
	command2 "github.com/airchains-network/decentralized-sequencer/cmd/command"
	"github.com/airchains-network/decentralized-sequencer/cmd/command/keys"
	"github.com/airchains-network/decentralized-sequencer/cmd/command/zkpCmd"
	"github.com/spf13/cobra"
	"os"
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "station-trackd",
		Short: "Decentralized Sequencer for StaionApps",
	}

	rootCmd.AddCommand(command2.StationCmd)
	rootCmd.AddCommand(command2.InitCmd)
	rootCmd.AddCommand(command2.KeyGenCmd)
	rootCmd.AddCommand(command2.ProverGenCMD)
	command2.KeyGenCmd.AddCommand(keys.JunctionKeyGenCmd)
	command2.ProverGenCMD.AddCommand(zkpCmd.V1ZKP)
	keys.JunctionKeyGenCmd.Flags().StringVarP(&keys.AcountName, "accountName", "n", "", "Account Name")
	keys.JunctionKeyGenCmd.Flags().StringVarP(&keys.AccountPath, "accountPath", "p", "", "Account Path")
	keys.JunctionKeyGenCmd.MarkFlagRequired("accountName")
	keys.JunctionKeyGenCmd.MarkFlagRequired("accountPath")
	command2.InitCmd.Flags().String("moniker", "", "Moniker for the sequencer")
	command2.InitCmd.Flags().String("stationType", "", "Station Type for the sequencer (evm | cosmwasm | svm)")
	command2.InitCmd.Flags().String("daType", "mock", "DA Type for the sequencer (avail | celestia | eigen | mock)")
	command2.InitCmd.MarkFlagRequired("moniker")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
