package trackgate

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/airchains-network/tracks/config"
	trackgateTypes "github.com/airchains-network/tracks/junction/trackgate/types"
	logs "github.com/airchains-network/tracks/log"
	"github.com/airchains-network/tracks/node/shared"
	//utils "github.com/airchains-network/tracks/p2p"
	"net/http"
	"time"

	//"github.com/airchains-network/tracks/junction/trackgate/types"
	"github.com/ignite/cli/v28/ignite/pkg/cosmosaccount"
	"github.com/ignite/cli/v28/ignite/pkg/cosmosclient"
)

func SchemaEngage(conf *config.Config, podNum int, schemaObjectByte []byte) bool {

	ctx := context.Background()

	accountName := conf.Junction.AccountName
	accountPath := conf.Junction.AccountPath
	addressPrefix := conf.Junction.AddressPrefix
	junctionRPC := conf.Junction.JunctionRPC
	stationId := conf.Junction.StationId

	client, err := cosmosclient.New(ctx, cosmosclient.WithAddressPrefix(addressPrefix), cosmosclient.WithNodeAddress(junctionRPC), cosmosclient.WithHome(accountPath), cosmosclient.WithGas("auto"), cosmosclient.WithFees("1000amf"))
	if err != nil {
		logs.Log.Error(fmt.Sprintf("Error in connecting client: %v", err))
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
	creator := newTempAddr

	sequencerDetails := &trackgateTypes.SequencerDetails{
		Name:      conf.Sequencer.SequencerType,
		Version:   VersionNameV1,
		NameSpace: conf.Sequencer.SequencerNamespace,
		Address:   "nil",
	}

	sequencerDetailsBytes, err := json.Marshal(sequencerDetails)
	if err != nil {
		logs.Log.Error(fmt.Sprintf("Error marshalling sequencer details: %v", err))
		return false
	}

	msg := &trackgateTypes.MsgSchemaEngage{
		Operator:            creator,
		ExtTrackStationId:   stationId,
		SchemaObject:        schemaObjectByte,
		AcknowledgementHash: "acknowledgementHash",
		PodNumber:           uint64(podNum),
		SequencerDetails:    sequencerDetailsBytes,
	}

	fmt.Println("podNum", podNum)

	// Broadcast a transaction from account `charlie` with the message
	// to create a post store response in txResp
	for {
		txResp, err := client.BroadcastTx(ctx, newTempAccount, msg)
		_ = txResp
		if err != nil {
			logs.Log.Error("Error in broadcasting transaction")
			logs.Log.Error(err.Error())
			logs.Log.Error("Broadcast transaction failed, retrying in 5 seconds...")
			time.Sleep(5 * time.Second)

			// latest pod number  pod 1 -> 2
			//trackgateTypes.

			continue
		} else {
			//utils.UpdateTrackgateTxState(shared.TxEVCUpdate)

			//os.Exit(0)

			break
		}

	}

	// update pod number
	//fmt.Println("schemaObjetByte", schemaObjectByte)
	//
	//for {
	//	espressoDataSubmitSuccess := SubmitEspressoTx(schemaObjectByte)
	//	if !espressoDataSubmitSuccess {
	//		logs.Log.Error("Gin server call failed, retrying in 5 seconds...")
	//		time.Sleep(5 * time.Second)
	//		continue
	//	} else {
	//
	//		break
	//	}
	//}

	podState := shared.GetPodState()
	podState.LatestPodHeight = uint64(podNum + 1)
	shared.SetPodState(podState)

	return true
}

func SubmitEspressoTx(schemaObjectByte []byte) bool {

	//gin server api
	submitURL := fmt.Sprintf("http://192.168.1.18:8080/track/espresso")

	resp, err := http.Post(submitURL, "application/json", bytes.NewBuffer(schemaObjectByte))
	if err != nil {
		logs.Log.Error(fmt.Sprintf("Error submitting espresso transaction : %v", err))
		return false
	}
	defer resp.Body.Close()
	return true
}
