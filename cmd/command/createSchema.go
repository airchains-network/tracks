package command

import (
	"fmt"
	logs "github.com/airchains-network/tracks/log"
	"github.com/airchains-network/tracks/node/shared"
	"github.com/spf13/cobra"
)

func parseCmdArgs(cmd *cobra.Command) (*StationArgs, error) {
	args := &StationArgs{}
	var err error

	args.accountName = cmd.Flag("accountName").Value.String()
	args.accountPath = cmd.Flag("accountPath").Value.String()
	args.jsonRPC = cmd.Flag("jsonRPC").Value.String()
	args.tracks, err = cmd.Flags().GetStringSlice("tracks")
	if err != nil {
		return nil, fmt.Errorf(" Failed to get 'tracks' flag values: %w", err)
	}
	args.bootstrapNode, err = cmd.Flags().GetStringSlice("bootstrapNode")
	if err != nil {
		return nil, fmt.Errorf(" Failed to get 'bootstarpNode' flag values: %w", err)
	}

	return args, nil
}

var CreateStation = &cobra.Command{
	Use:   "create-station",
	Short: "Create station from generated wallet",
	Run: func(cmd *cobra.Command, args []string) {
		conf, err := shared.LoadConfig()
		if err != nil {
			logs.Log.Error("Failed to load conf info")
			return
		}

		stationArgs, err := parseCmdArgs(cmd)
		if err != nil {
			logs.Log.Error(err.Error())
			return
		}
