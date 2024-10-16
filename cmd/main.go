package main

import (
	"fmt"
	"github.com/airchains-network/tracks/cmd/command/query"
	"os"

	"github.com/airchains-network/tracks/cmd/command"
	"github.com/airchains-network/tracks/cmd/command/keys"
	"github.com/airchains-network/tracks/cmd/command/zkpCmd"
	"github.com/ethereum/go-ethereum/log"
	"github.com/spf13/cobra"
)

// These variables will be set by the linker during the build process
var (
	Version = "dev"
	Build   = "none"
	Date    = "unknown"
	Branch  = "unknown"
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "track",
		Short: "Decentralized Sequencer for Stations",
	}

	// Define version command
	var versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Print the version number, commit number, date of release, and branch name",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Version: %s\nCommit: %s\nDate: %s\nBranch: %s\n", Version, Build, Date, Branch)
		},
	}

	// Add commands to rootCmd
	rootCmd.AddCommand(command.StationCmd)
	rootCmd.AddCommand(command.InitCmd)
	rootCmd.AddCommand(command.KeyGenCmd)
	rootCmd.AddCommand(command.ProverGenCMD)
	rootCmd.AddCommand(command.CreateStation)
	rootCmd.AddCommand(command.QueryCmd)
	rootCmd.AddCommand(command.CreateSchema)
	//rootCmd.AddCommand(command.SchemaEngage)
	rootCmd.AddCommand(command.Rollback)
	rootCmd.AddCommand(versionCmd) // Add version command

	//Add subcommands to query
	command.QueryCmd.AddCommand(query.ListStationEngagements)
	command.QueryCmd.AddCommand(query.ListStationSchemas)
	command.QueryCmd.AddCommand(query.ListStation)
	// Add subcommands to keygen and provergen
	command.KeyGenCmd.AddCommand(keys.JunctionKeyGenCmd)
	command.KeyGenCmd.AddCommand(keys.JunctionKeyImportCmd)
	command.ProverGenCMD.AddCommand(zkpCmd.V1ZKP)
	command.ProverGenCMD.AddCommand(zkpCmd.V1ZKPWasm)

	// Define flags for JunctionKeyGenCmd
	keys.JunctionKeyGenCmd.Flags().String("accountName", "", "Account Name")
	keys.JunctionKeyGenCmd.Flags().String("accountPath", "", "Account Path")
	keys.JunctionKeyGenCmd.MarkFlagRequired("accountName")
	keys.JunctionKeyGenCmd.MarkFlagRequired("accountPath")

	// Define flags for JunctionKeyImportCmd
	keys.JunctionKeyImportCmd.Flags().String("accountName", "", "Account Name")
	keys.JunctionKeyImportCmd.Flags().String("accountPath", "", "Account Path")
	keys.JunctionKeyImportCmd.Flags().String("mnemonic", "", "Mnemonic Key")
	keys.JunctionKeyImportCmd.MarkFlagRequired("accountName")
	keys.JunctionKeyImportCmd.MarkFlagRequired("accountPath")
	keys.JunctionKeyImportCmd.MarkFlagRequired("mnemonic")

	// Define flags for InitCmd
	command.InitCmd.Flags().String("moniker", "", "Moniker for the Tracks")

	command.InitCmd.Flags().String("stationType", "", "Station Type for the Tracks (evm | cosmwasm | svm)")
	command.InitCmd.Flags().String("stationRpc", "", "Station RPC for the Tracks")
	command.InitCmd.Flags().String("stationAPI", "", "Station API for the Tracks")

	command.InitCmd.Flags().String("daType", "mock", "DA Type for the Tracks (avail | celestia | eigen | mock)")
	command.InitCmd.Flags().String("daVersion", "mock", "DA Version")
	command.InitCmd.Flags().String("daName", "mock", "DA Name")
	command.InitCmd.Flags().String("daRpc", "", "DA RPC for the Tracks")
	command.InitCmd.Flags().String("daKey", "", "DA Key for the Tracks")

	command.InitCmd.Flags().String("sequencerType", "mock", "sequencer Type for the Tracks")
	command.InitCmd.Flags().String("sequencerVersion", "mock", "sequencer Version")
	command.InitCmd.Flags().String("sequencerRpc", "", "sequencer RPC for the Tracks")
	command.InitCmd.Flags().String("sequencerKey", "", "sequencer Key for the Tracks")
	command.InitCmd.Flags().String("sequencerNamespace", "mock", "Sequencer Namespace for the Tracks")

	command.InitCmd.Flags().String("proverType", "mock", "prover Type for the Tracks")
	command.InitCmd.Flags().String("proverVersion", "mock", "prover Version")
	command.InitCmd.Flags().String("proverRpc", "", "prover RPC for the Tracks")
	command.InitCmd.Flags().String("proverKey", "", "prover Key for the Tracks")

	command.InitCmd.MarkFlagRequired("moniker")
	command.InitCmd.MarkFlagRequired("sequencerType")
	command.InitCmd.MarkFlagRequired("daRpc")
	command.InitCmd.MarkFlagRequired("daName")
	command.InitCmd.MarkFlagRequired("daKey")
	command.InitCmd.MarkFlagRequired("stationType")
	command.InitCmd.MarkFlagRequired("stationRpc")
	command.InitCmd.MarkFlagRequired("stationAPI")

	// Define flags for CreateStation
	command.CreateStation.Flags().String("info", "", "Station information")
	command.CreateStation.Flags().String("stationName", "test", "Station Name ")
	command.CreateStation.Flags().String("accountName", "", "Station Account Name")
	command.CreateStation.Flags().String("accountPath", "", "Station Account Path")
	command.CreateStation.Flags().String("jsonRPC", "", "Station JSON RPC")
	command.CreateStation.Flags().StringSlice("tracks", []string{}, "tracks array for this station")
	command.CreateStation.Flags().StringSlice("bootstrapNode", []string{}, "Bootstrap Node for the Tracks")

	command.CreateStation.MarkFlagRequired("stationName")
	command.CreateStation.MarkFlagRequired("info")
	command.CreateStation.MarkFlagRequired("accountName")
	command.CreateStation.MarkFlagRequired("accountPath")
	command.CreateStation.MarkFlagRequired("jsonRPC")
	command.CreateStation.MarkFlagRequired("tracks")

	query.ListStationEngagements.Flags().String("offset", "0", "offset for the list")
	query.ListStationEngagements.Flags().String("limit", "100", "limit for the list")
	query.ListStationEngagements.Flags().String("order", "asc", "order of the list (asc | desc)")

	query.ListStationSchemas.Flags().String("offset", "0", "offset for the list")
	query.ListStationSchemas.Flags().String("limit", "100", "limit for the list")
	query.ListStationSchemas.Flags().Bool("reverse", false, "reverse the order of the list (true | false)")

	query.ListStation.Flags().String("offset", "0", "offset for the list")
	query.ListStation.Flags().String("limit", "100", "limit for the list")
	query.ListStation.Flags().Bool("reverse", false, "reverse the order of the list (true | false)")

	if err := rootCmd.Execute(); err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
}
