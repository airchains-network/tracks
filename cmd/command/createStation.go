package command

import (
	"encoding/json"
	"github.com/airchains-network/decentralized-sequencer/junction"
	logs "github.com/airchains-network/decentralized-sequencer/log"
	v1 "github.com/airchains-network/decentralized-sequencer/zk/v1"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

var CreateStation = &cobra.Command{
	Use:   "create-station",
	Short: "Create station from generated wallet",
	Run: func(cmd *cobra.Command, args []string) {

		stationInfo, _ := cmd.Flags().GetString("info")
		accountName := cmd.Flag("accountName").Value.String()
		accountPath := cmd.Flag("accountPath").Value.String()
		jsonRPC := cmd.Flag("jsonRPC").Value.String()
		stationId := uuid.New().String()
		provingKey, verificationKey, err := v1.GetVkPk()
		_ = provingKey // currently unused here
		if err != nil {
			logs.Log.Error("Failed to read Proving Key & Verification key" + err.Error())
			return
		}
		verificationKeyByte, err := json.Marshal(verificationKey)
		if err != nil {
			logs.Log.Error("Failed to unmarshal Verification key" + err.Error())
			return
		}

		junction.CreateStation(stationId, stationInfo, accountName, accountPath, jsonRPC, verificationKeyByte)
	},
}

// go run cmd/main.go create-station --accountName alice --accountPath ./accounts/junction --info "information" --jsonRPC "http://34.131.189.98:26657"
