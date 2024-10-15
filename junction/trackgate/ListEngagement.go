package trackgate

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/airchains-network/tracks/config"
	"github.com/airchains-network/tracks/junction/trackgate/types"
	logs "github.com/airchains-network/tracks/log"
	"github.com/ignite/cli/v28/ignite/pkg/cosmosaccount"
	"github.com/ignite/cli/v28/ignite/pkg/cosmosclient"
)

func ListEngagements(conf *config.Config, order string, offset uint64, limit uint64) bool {

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

	queryClient := types.NewQueryClient(client.Context())

	params := &types.QueryListTrackEngagementsRequest{
		ExtTrackStationId: stationId,
		Pagination: &types.TrackgatePaginationRequest{
			Offset: offset,
			Limit:  limit,
			Order:  order,
		},
	}

	schemas, err := queryClient.ListTrackEngagements(ctx, params)
	if err != nil {
		logs.Log.Error(fmt.Sprintf("Error getting track engagements: %v", err))
		return false
	}

	jsonData, err := json.MarshalIndent(schemas, "", "    ")
	if err != nil {
		logs.Log.Error(fmt.Sprintf("Error marshalling track engagements: %v", err))
		return false
	}
	fmt.Println(string(jsonData))

	return true
}
