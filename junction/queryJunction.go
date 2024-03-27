package junction

import (
	"context"
	"github.com/airchains-network/decentralized-sequencer/junction/types"
	logs "github.com/airchains-network/decentralized-sequencer/log"
	"github.com/airchains-network/decentralized-sequencer/node/shared"
	"github.com/airchains-network/decentralized-sequencer/utilis"
	"github.com/ignite/cli/v28/ignite/pkg/cosmosclient"
)

func QueryVRF() (vrfRecord *types.VrfRecord) {

	jsonRpc, stationId, _, _, _, err := utilis.GetJunctionDetails()
	if err != nil {
		logs.Log.Error("can not get junctionDetails.json data: " + err.Error())
		return nil
	}

	podNumber := shared.GetPodState().LatestPodHeight

	ctx := context.Background()
	client, err := cosmosclient.New(ctx, cosmosclient.WithNodeAddress(jsonRpc))
	if err != nil {
		logs.Log.Error("Client connection error: " + err.Error())
		return nil
	}

	queryClient := types.NewQueryClient(client.Context())
	queryResp, err := queryClient.FetchVrn(ctx, &types.QueryFetchVrnRequest{
		PodNumber: podNumber,
		StationId: stationId,
	})
	if err != nil {
		logs.Log.Error("Error fetching VRF: " + err.Error())
		return nil
	}

	return queryResp.Details
}
