package junction

import (
	"context"
	"fmt"
	"github.com/airchains-network/decentralized-sequencer/junction/types"
	logs "github.com/airchains-network/decentralized-sequencer/log"
	"github.com/airchains-network/decentralized-sequencer/node/shared"
	utilis "github.com/airchains-network/decentralized-sequencer/utils"
	"github.com/ignite/cli/v28/ignite/pkg/cosmosaccount"
	"github.com/ignite/cli/v28/ignite/pkg/cosmosclient"
	"time"
)

func SubmitCurrentPod() (success bool) {
	jsonRpc, stationId, accountPath, accountName, addressPrefix, tracks, err := GetJunctionDetails()
	_ = tracks
	if err != nil {
		logs.Log.Error("can not get junctionDetails.json data: " + err.Error())
		return false
	}

	podNumber := shared.GetPodState().LatestPodHeight

	// get latest pod hash
	LatestPodStatusHash := shared.GetPodState().LatestPodHash
	var LatestPodStatusHashStr string
	LatestPodStatusHashStr = string(LatestPodStatusHash)

	// previous pod hash
	PreviousPodHash := shared.GetPodState().PreviousPodHash
	var PreviousPodStatusHashStr string
	if PreviousPodHash == nil {
		PreviousPodStatusHashStr = ""
	} else {
		PreviousPodStatusHashStr = string(PreviousPodHash)
	}

	// get witness
	witnessByte := shared.GetPodState().LatestPublicWitness

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

	ctx := context.Background()
	gas := utilis.GenerateRandomWithFavour(510, 1000, [2]int{520, 700}, 0.7)
	gasFees := fmt.Sprintf("%damf", gas)

	logs.Log.Info(fmt.Sprintf("Gas Fees Used for submitPod transaction is: %s\n", gasFees))
	accountClient, err := cosmosclient.New(ctx, cosmosclient.WithAddressPrefix(addressPrefix), cosmosclient.WithNodeAddress(jsonRpc), cosmosclient.WithHome(accountPath), cosmosclient.WithGas("auto"), cosmosclient.WithFees(gasFees))
	if err != nil {
		logs.Log.Error("Error creating account client")
		return false
	}

	unixTime := time.Now().Unix()
	currentTime := fmt.Sprintf("%d", unixTime)

	//Error Solve kar
	//pMrh := shared.GetPodState().PreviousPodHash

	msg := types.MsgSubmitPod{
		Creator:                newTempAddr,
		StationId:              stationId,
		PodNumber:              podNumber,
		MerkleRootHash:         LatestPodStatusHashStr,
		PreviousMerkleRootHash: PreviousPodStatusHashStr, // bytes.NewBuffer(pMrh).String(), // PreviousPodStatusHashStr,
		PublicWitness:          witnessByte,
		Timestamp:              currentTime,
	}

	txRes, errTxRes := accountClient.BroadcastTx(ctx, newTempAccount, &msg)
	if errTxRes != nil {
		logs.Log.Error("error in transaction" + errTxRes.Error())
		return false
	}
	logs.Log.Info("Transaction Hash for SubmitPod: " + txRes.TxHash)
	return true

}
