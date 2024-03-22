package junction

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/airchains-network/decentralized-sequencer/junction/types"
	logs "github.com/airchains-network/decentralized-sequencer/log"
	"github.com/airchains-network/decentralized-sequencer/utilis"
	"github.com/ignite/cli/v28/ignite/pkg/cosmosaccount"
	"github.com/ignite/cli/v28/ignite/pkg/cosmosclient"
	"os"
)

func CreateStation(stationId, stationInfo, accountName, accountPath, jsonRPC string, verificationKey []byte) bool {

	// extra args
	extraArg := types.StationArg{
		TrackType: "Airchains Sequencer",
		DaType:    "Eigen",
		Prover:    "Airchains",
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
	fmt.Println(newTempAddr)
	os.Exit(0)

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
	accountClient, err := cosmosclient.New(ctx, cosmosclient.WithAddressPrefix("air"), cosmosclient.WithNodeAddress(jsonRPC), cosmosclient.WithHome(accountPath), cosmosclient.WithGas("auto"), cosmosclient.WithFees(gasFees))
	if err != nil {
		logs.Log.Error("Error creating account client")
		return false
	}

	newStationData := types.MsgInitStation{
		Creator:           newTempAddr,
		Tracks:            tracks,
		VerificationKey:   verificationKey,
		StationId:         stationId,
		StationInfo:       stationInfo,
		TracksVotingPower: tracksVotingPower,
		ExtraArg:          extraArgBytes,
	}

	txResp, err := accountClient.BroadcastTx(ctx, newTempAccount, &newStationData)
	if err != nil {
		logs.Log.Error("Error in broadcasting transaction")
		logs.Log.Error(err.Error())
		return false
	}

	logs.Log.Info(txResp.TxHash)
	return true
}
