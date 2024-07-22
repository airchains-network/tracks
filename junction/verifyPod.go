package junction

import (
	"context"
	"fmt"
	"github.com/airchains-network/tracks/junction/types"
	logs "github.com/airchains-network/tracks/log"
	"github.com/airchains-network/tracks/node/shared"
	utilis "github.com/airchains-network/tracks/utils"
	"github.com/ignite/cli/v28/ignite/pkg/cosmosaccount"
	"github.com/ignite/cli/v28/ignite/pkg/cosmosclient"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
	"time"
)

func VerifyCurrentPod() (success bool) {
	zerolog.TimeFieldFormat = time.RFC3339
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	jsonRpc, stationId, accountPath, accountName, addressPrefix, _, err := GetJunctionDetails()
	if err != nil {
		logs.Log.Error("can not get junctionDetails.json data: " + err.Error())
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

	ctx := context.Background()
	gas := utilis.GenerateRandomWithFavour(510, 1000, [2]int{520, 700}, 0.7)
	gasFees := fmt.Sprintf("%damf", gas)
	log.Info().Str("module", "junction").Str("Gas Fees Used to Validate VRF", gasFees)

	_, err = cosmosclient.New(ctx, cosmosclient.WithAddressPrefix(addressPrefix), cosmosclient.WithNodeAddress(jsonRpc), cosmosclient.WithHome(accountPath), cosmosclient.WithGas("auto"), cosmosclient.WithFees(gasFees))
	if err != nil {
		logs.Log.Error("Switchyard client connection error")
		logs.Log.Error(err.Error())

		return false
	}

	currentPodState := shared.GetPodState()

	podNumber := currentPodState.LatestPodHeight
	LatestPodProof := currentPodState.LatestPodProof

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

	verifyPodStruct := types.MsgVerifyPod{
		Creator:                newTempAddr,
		StationId:              stationId,
		PodNumber:              podNumber,
		MerkleRootHash:         LatestPodStatusHashStr,
		PreviousMerkleRootHash: PreviousPodStatusHashStr,
		ZkProof:                LatestPodProof,
	}

	// check if pod is already verified
	podDetails := QueryPod(podNumber)
	if podDetails == nil {
		// pod already submitted
		log.Debug().Str("module", "junction").Msg("TxError: Pod not submitted, can not verify")
		return false
	} else if podDetails.IsVerified == true {
		// pod already verified
		log.Debug().Str("module", "junction").Msg("Pod already verified")
		return true
	}

	for {
		ctx := context.Background()
		gas := utilis.GenerateRandomWithFavour(510, 1000, [2]int{620, 900}, 0.7)

		for {
			gasFees := fmt.Sprintf("%damf", gas)
			log.Info().Str("module", "junction").Str("Gas Fees Used to Verify Pod", gasFees)

			accountClient, err := cosmosclient.New(ctx, cosmosclient.WithAddressPrefix(addressPrefix), cosmosclient.WithNodeAddress(jsonRpc), cosmosclient.WithHome(accountPath), cosmosclient.WithGas("auto"), cosmosclient.WithFees(gasFees))
			if err != nil {
				logs.Log.Error("Switchyard client connection error")
				logs.Log.Error(err.Error())
				return false
			}

			txRes, errTxRes := accountClient.BroadcastTx(ctx, newTempAccount, &verifyPodStruct)
			if errTxRes != nil {
				errTxResStr := errTxRes.Error()
				log.Error().Str("module", "junction").Str("Error", errTxResStr).Msg("Error in VerifyPod transaction")
				log.Debug().Str("module", "junction").Msg("Retrying VerifyPod transaction after 10 seconds..")
				time.Sleep(10 * time.Second)

				// Increase gas and update gasFees
				gas += 200
				log.Info().Str("module", "junction").Str("Updated Gas Fees for Retry", fmt.Sprintf("%damf", gas))
			} else {
				VerifyPodTxHash := txRes.TxHash
				currentPodState.VerifyPodTxHash = VerifyPodTxHash
				shared.SetPodState(currentPodState)
				log.Info().Str("module", "junction").Str("txHash", txRes.TxHash).Msg("Pod Verification Tx Success")
				return true
			}
		}
	}

}
