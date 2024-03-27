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
	"github.com/airchains-network/decentralized-sequencer/utilis"
	"github.com/ignite/cli/v28/ignite/pkg/cosmosaccount"
	"github.com/ignite/cli/v28/ignite/pkg/cosmosclient"
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/group/edwards25519"
)

func InitVRF() (bool, []byte) {
	jsonRpc, stationId, accountPath, accountName, addressPrefix, err := utilis.GetJunctionDetails()
	if err != nil {
		logs.Log.Error("can not get junctionDetails.json data: " + err.Error())
		return false, nil
	}

	registry, err := cosmosaccount.New(cosmosaccount.WithHome(accountPath))
	if err != nil {
		logs.Log.Error(fmt.Sprintf("Error creating account registry: %v", err))
		return false, nil
	}

	newTempAccount, err := registry.GetByName(accountName)
	if err != nil {
		logs.Log.Error(fmt.Sprintf("Error getting account: %v", err))
		return false, nil
	}

	newTempAddr, err := newTempAccount.Address(addressPrefix)
	if err != nil {
		logs.Log.Error(fmt.Sprintf("Error getting address: %v", err))
		return false, nil
	}

	ctx := context.Background()
	gasFees := fmt.Sprintf("%damf", 213)
	logs.Log.Warn(fmt.Sprintf("Gas Fees Used for init VRF transaction is: %s\n", gasFees))
	accountClient, err := cosmosclient.New(ctx, cosmosclient.WithAddressPrefix(addressPrefix), cosmosclient.WithNodeAddress(jsonRpc), cosmosclient.WithHome(accountPath), cosmosclient.WithGas("auto"), cosmosclient.WithFees(gasFees))
	if err != nil {
		logs.Log.Error("Error creating account client")
		return false, nil
	}
	// getting the account and creating client codes --> End

	// get variables required to generate or call verifiable random number
	suite := edwards25519.NewBlakeSHA256Ed25519()

	podNumber := shared.GetPodState().LatestPodHeight

	privateKeyStr := utilis.GetVRFPrivateKey()
	if privateKeyStr == "" {
		return false, nil
	}

	privateKey, err := LoadHexPrivateKey(privateKeyStr)
	if err != nil {
		logs.Log.Error("Error in loading private key: " + err.Error())
		return false, nil
	}
	publicKey := utilis.GetVRFPubKey()
	if publicKey == "" {
		return false, nil
	}

	rc := mainTypes.RequestCommitmentV2Plus{
		BlockNum:         1,
		StationId:        stationId,
		UpperBound:       1,
		RequesterAddress: newTempAddr,
	}

	serializedRC, err := utilis.SerializeRequestCommitmentV2Plus(rc)
	if err != nil {
		logs.Log.Error(err.Error())
		return false, nil
	}

	proof, vrfOutput, err := utilis.GenerateVRFProof(suite, privateKey, serializedRC, int64(rc.BlockNum))
	if err != nil {
		fmt.Printf("Error generating unique proof: %v\n", err)
		return false, nil
	}

	extraArg := types.ExtraArg{
		SerializedRc: serializedRC,
		Proof:        proof,
		VrfOutput:    vrfOutput,
	}

	// marshal
	extraArgsByte, err := json.Marshal(extraArg)
	if err != nil {
		logs.Log.Error(err.Error())
		return false, nil
	}

	var defaultOccupancy uint64
	defaultOccupancy = 1 // todo: for multinode sequencer =node.DefaultOccupancy
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
		return false, nil
	}

	logs.Log.Info("Transaction Hash: " + txRes.TxHash)

	return true, serializedRC

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
