package trackgate

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/airchains-network/tracks/config"
	trackgateTypes "github.com/airchains-network/tracks/junction/trackgate/types"
	logs "github.com/airchains-network/tracks/log"
	"github.com/airchains-network/tracks/types"
	//"github.com/airchains-network/tracks/junction/trackgate/types"
	"github.com/ignite/cli/v28/ignite/pkg/cosmosaccount"
	"github.com/ignite/cli/v28/ignite/pkg/cosmosclient"
)

func SchemaEngage(conf *config.Config, podNum int, EspressoTxResponse *types.EspressoSchemaV1) bool {

	schemaObjectByte, err := json.Marshal(EspressoTxResponse)
	if err != nil {
		logs.Log.Error(fmt.Sprintf("Error  marshaling JSON: %v", err))
		return false
	}

	ctx := context.Background()

	accountName := conf.Junction.AccountName
	accountPath := conf.Junction.AccountPath
	addressPrefix := conf.Junction.AddressPrefix
	junctionRPC := conf.Junction.JunctionRPC
	stationId := conf.Junction.StationId

	client, err := cosmosclient.New(ctx, cosmosclient.WithAddressPrefix(addressPrefix), cosmosclient.WithNodeAddress(junctionRPC), cosmosclient.WithHome(accountPath), cosmosclient.WithGas("1000000"))
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

	msg := &trackgateTypes.MsgSchemaEngage{
		Operator:          creator,
		ExtTrackStationId: stationId,
		SchemaObject:      schemaObjectByte,
		StateRoot:         "stateroot-1",
		PodNumber:         uint64(podNum),
	}

	// Broadcast a transaction from account `charlie` with the message
	// to create a post store response in txResp
	txResp, err := client.BroadcastTx(ctx, newTempAccount, msg)
	if err != nil {
		logs.Log.Error("Error in broadcasting transaction")
		logs.Log.Error(err.Error())
		return false
	}

	// Print response from broadcasting a transaction
	fmt.Print("MsgCreatePost:\n\n")
	fmt.Println(txResp)

	return true
}
