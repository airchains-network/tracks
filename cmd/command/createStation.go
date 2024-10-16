package command

import (
	"fmt"
	"github.com/airchains-network/tracks/junction/junction"
	junctionTypes "github.com/airchains-network/tracks/junction/junction/types"
	"github.com/airchains-network/tracks/junction/trackgate"
	logs "github.com/airchains-network/tracks/log"
	"github.com/airchains-network/tracks/node/shared"
	"github.com/airchains-network/tracks/types"
	v1 "github.com/airchains-network/tracks/zk/v1EVM"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

type StationArgs struct {
	stationName   string
	accountName   string
	accountPath   string
	jsonRPC       string
	tracks        []string
	operators     []string
	bootstrapNode []string
}

func parseCmdArgs(cmd *cobra.Command) (*StationArgs, error) {
	args := &StationArgs{}
	var err error

	args.stationName = cmd.Flag("stationName").Value.String()
	args.accountName = cmd.Flag("accountName").Value.String()
	args.accountPath = cmd.Flag("accountPath").Value.String()
	args.jsonRPC = cmd.Flag("jsonRPC").Value.String()
	args.tracks, err = cmd.Flags().GetStringSlice("tracks")
	if err != nil {
		return nil, fmt.Errorf(" Failed to get 'tracks' flag values: %w", err)
	}
	args.operators, err = cmd.Flags().GetStringSlice("operators")
	if err != nil {
		return nil, fmt.Errorf(" Failed to get 'operators' flag values: %w", err)
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

		addressPrefix := "air"

		if conf.Sequencer.SequencerType == "espresso" {

			success := trackgate.InitStation(stationArgs.accountName, stationArgs.accountPath, stationArgs.jsonRPC, stationArgs.bootstrapNode, addressPrefix, stationArgs.stationName, stationArgs.operators)
			if !success {
				logs.Log.Error("Failed to create new station due to above error")
				return
			}
			successSchema := trackgate.SchemaCreation()
			if !successSchema {
				logs.Log.Error("Failed to deploy schema  due to above error")
				return
			}
		} else {
			_, verificationKey, err := v1.GetVkPk()
			if err != nil {
				logs.Log.Error("Failed to read Proving Key & Verification key: " + err.Error())
				return
			}

			stationInfo := types.StationInfo{
				StationType: conf.Station.StationType,
			}

			extraArg := junctionTypes.StationArg{
				TrackType: conf.Sequencer.SequencerType,
				DaType:    conf.DA.DaType,
				Prover:    conf.Prover.ProverType,
			}
			success := junction.CreateStation(extraArg, uuid.New().String(), stationInfo, stationArgs.accountName, stationArgs.accountPath, stationArgs.jsonRPC, verificationKey, addressPrefix, stationArgs.tracks, stationArgs.bootstrapNode)
			if !success {
				logs.Log.Error("Failed to create new station due to above error")
				return
			}
		}

		logs.Log.Info("Successfully created station")
	},
}
