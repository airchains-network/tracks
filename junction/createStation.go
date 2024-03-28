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

func CreateStation(extraArg junctionTypes.StationArg, stationId string, stationInfo types.StationInfo, accountName, accountPath, jsonRPC string, verificationKey groth16.VerifyingKey, addressPrefix string, tracks []string) bool {

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

	// Voting powers: almost equal.
	var tracksVotingPower []uint64
	totalPower := uint64(100)
	numTracks := len(tracks)
	// Calculate the equal share for each track
	equalShare := totalPower / uint64(numTracks)
	// Calculate the remainder
	remainder := totalPower % uint64(numTracks)
	// Distribute the equal share to each track
	for i := 0; i < numTracks; i++ {
		if remainder > 0 {
			// For each track, until the remainder is exhausted,
			// add an extra unit of power to make the total sum 100.
			tracksVotingPower = append(tracksVotingPower, equalShare+1)
			remainder-- // Decrement the remainder until it's 0
		} else {
			// Once the remainder is exhausted, append the equal share.
			tracksVotingPower = append(tracksVotingPower, equalShare)
		}
	}

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
	logs.Log.Info("tx hash: " + txResp.TxHash)

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

	success = utilis.SetJunctionDetails(jsonRPC, stationId, accountPath, accountName, addressPrefix, tracks)
	if !success {
		logs.Log.Error("Failed to create data/vrfPubKey.txt properly")
		return false
	}
	logs.Log.Info("data/vrfPubKey.txt Created")

	return true
}
