package main

import (
	"fmt"
	"github.com/airchains-network/decentralized-sequencer/cmd/command"
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

	rootCmd.AddCommand(command.StationCmd)
	rootCmd.AddCommand(command.InitCmd)
	rootCmd.AddCommand(command.KeyGenCmd)
	rootCmd.AddCommand(command.ProverGenCMD)
	rootCmd.AddCommand(command.CreateStation)

	command.KeyGenCmd.AddCommand(keys.JunctionKeyGenCmd)
	command.ProverGenCMD.AddCommand(zkpCmd.V1ZKP)

	keys.JunctionKeyGenCmd.Flags().String("accountName", "", "Account Name")
	keys.JunctionKeyGenCmd.Flags().String("accountPath", "", "Account Path")
	keys.JunctionKeyGenCmd.MarkFlagRequired("accountName")
	keys.JunctionKeyGenCmd.MarkFlagRequired("accountPath")

	command.InitCmd.Flags().String("moniker", "", "Moniker for the Tracks")
	command.InitCmd.Flags().String("stationType", "", "Station Type for the Tracks (evm | cosmwasm | svm)")
	command.InitCmd.Flags().String("daType", "mock", "DA Type for the Tracks (avail | celestia | eigen | mock)")
	command.InitCmd.Flags().String("daRpc", "", "DA RPC for the Tracks")
	command.InitCmd.Flags().String("daKey", "", "DA Key for the Tracks")
	command.InitCmd.Flags().String("stationRpc", "", "Station RPC for the Tracks")
	command.InitCmd.Flags().String("stationAPI", "", "Station API for the Tracks")
	command.InitCmd.MarkFlagRequired("moniker")
	command.InitCmd.MarkFlagRequired("daRpc")
	command.InitCmd.MarkFlagRequired("daKey")
	command.InitCmd.MarkFlagRequired("stationType")
	command.InitCmd.MarkFlagRequired("stationRpc")
	command.InitCmd.MarkFlagRequired("stationAPI")

	command.CreateStation.Flags().String("info", "", "Station information")
	command.CreateStation.Flags().String("accountName", "", "Station Account Name")
	command.CreateStation.Flags().String("accountPath", "", "Station Account Path")
	command.CreateStation.Flags().String("jsonRPC", "", "Station JSON RPC")
	//command.CreateStation.Flags().String("tracks", "", "tracks array for this station")
	command.CreateStation.Flags().StringSlice("tracks", []string{}, "tracks array for this station")
	command.CreateStation.Flags().StringSlice("bootstrapNode", []string{}, "Bootstrap Node for the Tracks")

	command.CreateStation.MarkFlagRequired("info")
	command.CreateStation.MarkFlagRequired("accountName")
	command.CreateStation.MarkFlagRequired("accountPath")
	command.CreateStation.MarkFlagRequired("jsonRPC")
	command.CreateStation.MarkFlagRequired("tracks")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

}
