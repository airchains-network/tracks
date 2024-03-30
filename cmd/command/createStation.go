package command

import (
	"github.com/airchains-network/decentralized-sequencer/config"
	"github.com/airchains-network/decentralized-sequencer/junction"
	junctionTypes "github.com/airchains-network/decentralized-sequencer/junction/types"
	logs "github.com/airchains-network/decentralized-sequencer/log"
	"github.com/airchains-network/decentralized-sequencer/node/shared"
	"github.com/airchains-network/decentralized-sequencer/types"
	v1 "github.com/airchains-network/decentralized-sequencer/zk/v1"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

var CreateStation = &cobra.Command{
	Use:   "create-station",
	Short: "Create station from generated wallet",
	Run: func(cmd *cobra.Command, args []string) {

		var conf config.Config
		var err error

		conf, err = shared.LoadConfig()
		if err != nil {
			logs.Log.Error("Failed to load conf info")
			return
		}

		// station info
		// todo change station info. types and inputs both.
		stationInfo := types.StationInfo{
			StationType: conf.Station.StationType,
		}

		// extra args
		extraArg := junctionTypes.StationArg{
			TrackType: "Airchains Sequencer",
			DaType:    conf.DA.DaType,
			Prover:    "Airchains",
		}

		accountName := cmd.Flag("accountName").Value.String()
		accountPath := cmd.Flag("accountPath").Value.String()
		jsonRPC := cmd.Flag("jsonRPC").Value.String()
		tracks, err := cmd.Flags().GetStringSlice("tracks")
		if err != nil {
			logs.Log.Error("Failed to get 'tracks' flag values: " + err.Error())
			return
		}

		stationId := uuid.New().String()
		provingKey, verificationKey, err := v1.GetVkPk()
		_ = provingKey // currently unused here
		if err != nil {
			logs.Log.Error("Failed to read Proving Key & Verification key" + err.Error())
			return
		}

		addressPrefix := "air"
		success := junction.CreateStation(extraArg, stationId, stationInfo, accountName, accountPath, jsonRPC, verificationKey, addressPrefix, tracks)
		if !success {
			logs.Log.Error("Failed to create new station due to above error")
			return
		} else {
			logs.Log.Info("Successfully created station")
			return
		}
	},
}
