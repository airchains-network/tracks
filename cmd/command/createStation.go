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

//air1h25pqnxkv8g50n5nlrdv94wktjupfu4ujevsc8
/*
 sh restart.sh;
 go run cmd/main.go create-station --accountName noob --accountPath ./accounts/keys --jsonRPC "http://34.131.189.98:26657" --info "some info" --tracks air1dqf8xx42e8tlcwpd4ucwf60qeg4k6h7mzpnkf7,air1h25pqnxkv8g50n5nlrdv94wktjupfu4ujevsc8
 touch data/stationData.json; touch data/vrfPrivKey.txt; touch data/vrfPubKey.txt;
 go run cmd/main.go init --daRpc "mock-rpc" --daType "mock"  --moniker "monkey" --stationRpc "http://34.131.189.98:26657" --stationType "evm"
*/

/*
 sh restart.sh;
 touch data/genesis.json; touch data/stationData.json; touch data/vrfPrivKey.txt; touch data/vrfPubKey.txt;
 copy paste the above files to other nodes
*/
/*
 start both nodes simultaniously
  go run cmd/main.go start
  go run cmd/main.go start /ip4/192.168.1.24/tcp/2300/p2p/12D3KooWKAvMmJqu7A53UHC36ViycTS5M5V5wSB7Qf7pMuAeh7HK
*/
