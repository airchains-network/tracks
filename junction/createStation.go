package junction

import (
	"context"
	"encoding/json"
	"fmt"
	junctionTypes "github.com/airchains-network/decentralized-sequencer/junction/types"
	logs "github.com/airchains-network/decentralized-sequencer/log"
	"github.com/airchains-network/decentralized-sequencer/types"
	"github.com/airchains-network/decentralized-sequencer/utilis"
	"github.com/consensys/gnark/backend/groth16"
	"github.com/ignite/cli/v28/ignite/pkg/cosmosaccount"
	"github.com/ignite/cli/v28/ignite/pkg/cosmosclient"
	"time"
)

func CreateStation(extraArg junctionTypes.StationArg, stationId string, stationInfo types.StationInfo, accountName, accountPath, jsonRPC string, verificationKey groth16.VerifyingKey) bool {

	// convert station info to string
	stationJsonBytes, err := json.Marshal(stationInfo)
	if err != nil {
		logs.Log.Error("Error marshaling to JSON: " + err.Error())
		return false
	}
	stationInfoStr := string(stationJsonBytes)

	verificationKeyByte, err := json.Marshal(verificationKey)
	if err != nil {
		logs.Log.Error("Failed to unmarshal Verification key" + err.Error())
		return false
	}

	extraArgBytes, err := json.Marshal(extraArg)
	if err != nil {
		logs.Log.Error("Error marshalling extra arg")
		return false
	}

	addressPrefix := "air"
	registry, err := cosmosaccount.New(cosmosaccount.WithHome(accountPath))
	if err != nil {
		logs.Log.Error(fmt.Sprintf("Error creating account registry: %v", err))
		return false
	}

	newTempAccount, err := registry.GetByName(accountName)
	if err != nil {
		logs.Log.Error(fmt.Sprintf("Error getting account: %v", err))
		return false
	}

	newTempAddr, err := newTempAccount.Address(addressPrefix)
	if err != nil {
		logs.Log.Error(fmt.Sprintf("Error getting address: %v", err))
		return false
	}
	logs.Log.Info("tracks address: " + newTempAddr)

	success, amount, err := CheckBalance(jsonRPC, newTempAddr)
	if err != nil || !success {
		logs.Log.Error("Error checking balance")
		return false
	}
	if amount < 100 {
		logs.Log.Error("Not enough balance on " + newTempAddr + " to create station")
		return false
	}
	amountStr := fmt.Sprintf("%damf", amount)
	logs.Log.Info("Currently user have " + amountStr)

	var tracksVotingPower []uint64
	power := uint64(100)
	for i := 0; i < 1; i++ {
		tracksVotingPower = append(tracksVotingPower, power)
	}
	var tracks []string
	tracks = append(tracks, newTempAddr)

	ctx := context.Background()
	gas := utilis.GenerateRandomWithFavour(611, 1200, [2]int{612, 1000}, 0.7)
	gasFees := fmt.Sprintf("%damf", gas)
	accountClient, err := cosmosclient.New(ctx, cosmosclient.WithAddressPrefix(addressPrefix), cosmosclient.WithNodeAddress(jsonRPC), cosmosclient.WithHome(accountPath), cosmosclient.WithGas("auto"), cosmosclient.WithFees(gasFees))
	if err != nil {
		logs.Log.Error("Error creating account client")
		return false
	}

	stationData := junctionTypes.MsgInitStation{
		Creator:           newTempAddr,
		Tracks:            tracks,
		VerificationKey:   verificationKeyByte,
		StationId:         stationId,
		StationInfo:       stationInfoStr,
		TracksVotingPower: tracksVotingPower,
		ExtraArg:          extraArgBytes,
	}

	txResp, err := accountClient.BroadcastTx(ctx, newTempAccount, &stationData)
	if err != nil {
		logs.Log.Error("Error in broadcasting transaction")
		logs.Log.Error(err.Error())
		return false
	}

	timestamp := time.Now().String()
	successGenesis := utilis.CreateGenesisJson(stationInfo, verificationKey, stationId, tracks, tracksVotingPower, txResp.TxHash, timestamp, extraArg, newTempAddr)
	if !successGenesis {
		return false
	}

	// create VRF Keys
	vrfPrivateKey, vrfPublicKey := utilis.NewKeyPair()
	vrfPrivateKeyHex := vrfPrivateKey.String()
	vrfPublicKeyHex := vrfPublicKey.String()
	if vrfPrivateKeyHex != "" {
		utilis.SetVRFPrivKey(vrfPrivateKeyHex)
	} else {
		logs.Log.Error("Error saving VRF private key")
	}
	if vrfPublicKeyHex != "" {
		utilis.SetVRFPubKey(vrfPublicKeyHex)
	} else {
		logs.Log.Error("Error saving VRF public key")
	}
	logs.Log.Info("Successfully Created VRF public and private Keys")

	return true
}

//  go run cmd/main.go create-station --accountName noob --accountPath ./accounts/keys --jsonRPC "http://34.131.189.98:26657"
