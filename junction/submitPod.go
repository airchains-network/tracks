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
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
	"time"
)

func SubmitCurrentPod() (success bool) {
	zerolog.TimeFieldFormat = time.RFC3339
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	jsonRpc, stationId, accountPath, accountName, addressPrefix, tracks, err := GetJunctionDetails()
	_ = tracks
	if err != nil {
		logs.Log.Error("can not get junctionDetails.json data: " + err.Error())
		return false
	}
	currentPodState := shared.GetPodState()

	podNumber := currentPodState.LatestPodHeight

	// get latest pod hash
	LatestPodStatusHash := currentPodState.LatestPodHash
	var LatestPodStatusHashStr string
	LatestPodStatusHashStr = string(LatestPodStatusHash)

	// previous pod hash
	PreviousPodHash := currentPodState.PreviousPodHash
	var PreviousPodStatusHashStr string
	if PreviousPodHash == nil {
		PreviousPodStatusHashStr = ""
	} else {
		PreviousPodStatusHashStr = string(PreviousPodHash)
	}

	// get witness
	witnessByte := currentPodState.LatestPublicWitness

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
	log.Info().Str("module", "junction").Str("Gas Fees Used to Validate VRF", gasFees)
	accountClient, err := cosmosclient.New(ctx, cosmosclient.WithAddressPrefix(addressPrefix), cosmosclient.WithNodeAddress(jsonRpc), cosmosclient.WithHome(accountPath), cosmosclient.WithGas("auto"), cosmosclient.WithFees(gasFees))
	if err != nil {
		logs.Log.Error("Switchyard client connection error")
		logs.Log.Error(err.Error())

		return false
	}

	unixTime := time.Now().Unix()
	currentTime := fmt.Sprintf("%d", unixTime)

	msg := types.MsgSubmitPod{
		Creator:                newTempAddr,
		StationId:              stationId,
		PodNumber:              podNumber,
		MerkleRootHash:         LatestPodStatusHashStr,
		PreviousMerkleRootHash: PreviousPodStatusHashStr, // bytes.NewBuffer(pMrh).String(), // PreviousPodStatusHashStr,
		PublicWitness:          witnessByte,
		Timestamp:              currentTime,
	}

	// check if pod is already submitted
	podDetails := QueryPod(podNumber)
	if podDetails != nil {
		// pod already submitted
		log.Debug().Str("module", "junction").Msg("Pod already submitted")
		return true
	}

	for {
		txRes, errTxRes := accountClient.BroadcastTx(ctx, newTempAccount, &msg)
		if errTxRes != nil {
			errStr := errTxRes.Error()
			log.Error().Str("module", "junction").Str("Error", errStr).Msg("Error in SubmitPod Transaction")

			log.Debug().Str("module", "junction").Msg("Retrying SubmitPod transaction after 10 seconds..")
			time.Sleep(10 * time.Second)
			//return false
		} else {
			// update txHash of submit pod in pod state
			currentPodState.InitPodTxHash = txRes.TxHash
			shared.SetPodState(currentPodState)
			log.Info().Str("module", "junction").Str("txHash", txRes.TxHash).Msg("Pod submitted successfully")
			return true
		}
	}
}
