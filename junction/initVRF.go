package junction

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/airchains-network/decentralized-sequencer/junction/types"
	logs "github.com/airchains-network/decentralized-sequencer/log"
	"github.com/airchains-network/decentralized-sequencer/node/shared"
	mainTypes "github.com/airchains-network/decentralized-sequencer/types"
	"github.com/ignite/cli/v28/ignite/pkg/cosmosaccount"
	"github.com/ignite/cli/v28/ignite/pkg/cosmosclient"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/group/edwards25519"
	"os"
	"time"
)

func InitVRF() (success bool, addr string) {
	zerolog.TimeFieldFormat = time.RFC3339
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	jsonRpc, stationId, accountPath, accountName, addressPrefix, tracks, err := GetJunctionDetails()
	if err != nil {
		logs.Log.Error("can not get junctionDetails.json data: " + err.Error())
		return false, ""
	}
	upperBond := uint64(len(tracks))

	registry, err := cosmosaccount.New(cosmosaccount.WithHome(accountPath))
	if err != nil {
		logs.Log.Error(fmt.Sprintf("Error creating account registry: %v", err))
		return false, ""
	}

	newTempAccount, err := registry.GetByName(accountName)
	if err != nil {
		logs.Log.Error(fmt.Sprintf("Error getting account: %v", err))
		return false, ""
	}

	newTempAddr, err := newTempAccount.Address(addressPrefix)
	if err != nil {
		logs.Log.Error(fmt.Sprintf("Error getting address: %v", err))
		return false, ""
	}

	ctx := context.Background()
	gasFees := fmt.Sprintf("%damf", 213)
	log.Info().Str("module", "junction").Str("Gas Fees Used for Vrf Initialization", gasFees)
	accountClient, err := cosmosclient.New(ctx, cosmosclient.WithAddressPrefix(addressPrefix), cosmosclient.WithNodeAddress(jsonRpc), cosmosclient.WithHome(accountPath), cosmosclient.WithGas("auto"), cosmosclient.WithFees(gasFees))
	if err != nil {
		logs.Log.Error("Error creating account client")
		return false, ""
	}

	// get variables required to generate or call verifiable random number
	suite := edwards25519.NewBlakeSHA256Ed25519()

	currentPodState := shared.GetPodState()
	podNumber := currentPodState.LatestPodHeight

	privateKeyStr := GetVRFPrivateKey()
	if privateKeyStr == "" {
		return false, ""
	}

	privateKey, err := LoadHexPrivateKey(privateKeyStr)
	if err != nil {
		logs.Log.Error("Error in loading private key: " + err.Error())
		return false, ""
	}
	publicKey := GetVRFPubKey()
	if publicKey == "" {
		return false, ""
	}

	rc := mainTypes.RequestCommitmentV2Plus{
		BlockNum:         1,
		StationId:        stationId,
		UpperBound:       upperBond,
		RequesterAddress: newTempAddr,
	}

	serializedRC, err := SerializeRequestCommitmentV2Plus(rc)
	if err != nil {
		logs.Log.Error(err.Error())
		return false, ""
	}

	proof, vrfOutput, err := GenerateVRFProof(suite, privateKey, serializedRC, int64(rc.BlockNum))
	if err != nil {
		fmt.Printf("Error generating unique proof: %v\n", err)
		return false, ""
	}

	extraArg := types.ExtraArg{
		SerializedRc: serializedRC,
		Proof:        proof,
		VrfOutput:    vrfOutput,
	}

	extraArgsByte, err := json.Marshal(extraArg)
	if err != nil {
		logs.Log.Error(err.Error())
		return false, ""
	}

	var defaultOccupancy uint64
	defaultOccupancy = 1
	msg := types.MsgInitiateVrf{
		Creator:        newTempAddr,
		PodNumber:      podNumber,
		StationId:      stationId,
		Occupancy:      defaultOccupancy,
		CreatorsVrfKey: publicKey,
		ExtraArg:       extraArgsByte,
	}

	txRes, errTxRes := accountClient.BroadcastTx(ctx, newTempAccount, &msg)
	if errTxRes != nil {
		logs.Log.Error("error in transaction" + errTxRes.Error())
		return false, ""
	}

	log.Info().Str("module", "junction").Str("Transaction Hash", txRes.TxHash)

	// update transaction hash in current pod
	currentPodState.VRFInitiationTxHash = txRes.TxHash
	// update pod state: update tx hash
	shared.SetPodState(currentPodState)
	log.Info().Str("module", "junction").Str("hash", txRes.TxHash).Msg("Vrf Initiation Tx Hash")

	log.Info().Str("module", "junction").Msg(txRes.TxHash)
	return true, newTempAddr

}

func LoadHexPrivateKey(hexPrivateKey string) (privateKey kyber.Scalar, err error) {
	// Initialize the Kyber suite for Edwards25519 curve
	// Convert the hexadecimal string to a byte slice
	privateKeyBytes, err := hex.DecodeString(hexPrivateKey)
	if err != nil {
		fmt.Printf("Error decoding private key: %v\n", err)
		return nil, err
	}

	// Initialize the Kyber suite for Edwards25519 curve
	suite := edwards25519.NewBlakeSHA256Ed25519()

	// Convert the byte slice into a Kyber scalar
	privateKey = suite.Scalar().SetBytes(privateKeyBytes)
	return privateKey, nil
}
