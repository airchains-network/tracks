package junction

import (
	"context"
	"fmt"
	"github.com/airchains-network/tracks/junction/types"
	logs "github.com/airchains-network/tracks/log"
	"github.com/airchains-network/tracks/node/shared"
	mainTypes "github.com/airchains-network/tracks/types"
	utilis "github.com/airchains-network/tracks/utils"
	"github.com/ignite/cli/v28/ignite/pkg/cosmosaccount"
	"github.com/ignite/cli/v28/ignite/pkg/cosmosclient"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
	"strings"
	"time"
)

func ValidateVRF(addr string) bool {
	zerolog.TimeFieldFormat = time.RFC3339
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	jsonRpc, stationId, accountPath, accountName, addressPrefix, tracks, err := GetJunctionDetails()
	if err != nil {
		logs.Log.Error("can not get junctionDetails.json data: " + err.Error())
		return false
	}
	upperBond := uint64(len(tracks))

	registry, err := cosmosaccount.New(cosmosaccount.WithHome(accountPath))
	if err != nil {
		logs.Log.Error(fmt.Sprintf("Error creating account registry: %v", err))
		return false
	}

	rc := mainTypes.RequestCommitmentV2Plus{
		BlockNum:         1,
		StationId:        stationId,
		UpperBound:       upperBond,
		RequesterAddress: addr,
	}

	serializedRC, err := SerializeRequestCommitmentV2Plus(rc)
	if err != nil {
		logs.Log.Error(err.Error())
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
	msg := types.MsgValidateVrf{
		Creator:      newTempAddr,
		StationId:    stationId,
		PodNumber:    podNumber,
		SerializedRc: serializedRC,
	}

	latestVerifiedBatch := QueryLatestVerifiedBatch()
	if latestVerifiedBatch+1 != podNumber {
		log.Debug().Str("module", "junction").Msg("Incorrect pod number")
		if latestVerifiedBatch+1 < podNumber {
			log.Debug().Str("module", "junction").Msg("Rollback required")
			return false
		} else if latestVerifiedBatch+1 > podNumber {
			log.Debug().Str("module", "junction").Msg("Pod number at Switchyard is ahead of the current pod number")
			return true
		}
	}

	for {
		ctx := context.Background()
		gas := utilis.GenerateRandomWithFavour(510, 1000, [2]int{520, 700}, 0.7)

		for {
			gasFees := fmt.Sprintf("%damf", gas)
			log.Info().Str("module", "junction").Str("Gas Fees Used to Validate VRF", gasFees)

			accountClient, err := cosmosclient.New(ctx, cosmosclient.WithAddressPrefix(addressPrefix), cosmosclient.WithNodeAddress(jsonRpc), cosmosclient.WithHome(accountPath), cosmosclient.WithGas("auto"), cosmosclient.WithFees(gasFees))
			if err != nil {
				logs.Log.Error("Switchyard client connection error")
				logs.Log.Error(err.Error())
				return false
			}

			txRes, errTxRes := accountClient.BroadcastTx(ctx, newTempAccount, &msg)
			if errTxRes != nil {
				errStr := errTxRes.Error()
				if strings.Contains(errStr, VRFValidatedErrorContains) {
					log.Debug().Str("module", "junction").Msg("VRF already verified for this pod number")
					return true
				} else if strings.Contains(errStr, "insufficient fees") {
					log.Error().Str("module", "junction").Str("Error", errStr).Msg("Error in ValidateVRF transaction")
					// retry transaction with increased gas fees
					log.Debug().Str("module", "junction").Msg("Retrying ValidateVRF transaction after 10 seconds..")
					time.Sleep(10 * time.Second)

					// Increase gas and update gasFees
					gas += 200
					log.Info().Str("module", "junction").Str("Updated Gas Fees for Retry", gasFees)
				} else {
					log.Error().Str("module", "junction").Str("Error", errStr).Msg("Unexpected error in ValidateVRF transaction")
					return false
				}
			} else {
				// update VRN verified hash
				currentPodState.VRFValidationTxHash = txRes.TxHash
				shared.SetPodState(currentPodState)
				log.Info().Str("module", "junction").Str("txHash", txRes.TxHash).Msg("VRF Validated Tx Success")
				return true
			}
		}
	}

}
