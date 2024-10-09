package junction

import (
	"context"
	types2 "github.com/airchains-network/tracks/junction/junction/types"
	logs "github.com/airchains-network/tracks/log"
	"github.com/airchains-network/tracks/node/shared"
	"github.com/ignite/cli/v28/ignite/pkg/cosmosclient"
)

func QueryVRF() (vrfRecord *types2.VrfRecord) {

	jsonRpc, stationId, _, _, _, _, err := GetJunctionDetails()
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

	queryClient := types2.NewQueryClient(client.Context())
	queryResp, err := queryClient.FetchVrn(ctx, &types2.QueryFetchVrnRequest{
		PodNumber: podNumber,
		StationId: stationId,
	})
	if err != nil {
		//logs.Log.Error("Error fetching VRF: " + err.Error())
		return nil
	}

	return queryResp.Details
}

func QueryPod(podNumber uint64) (pod *types2.Pods) {

	jsonRpc, stationId, _, _, _, _, err := GetJunctionDetails()
	if err != nil {
		logs.Log.Error("can not get junctionDetails.json data: " + err.Error())
		return nil
	}

	ctx := context.Background()
	client, err := cosmosclient.New(ctx, cosmosclient.WithNodeAddress(jsonRpc))
	if err != nil {
		logs.Log.Error("Client connection error: " + err.Error())
		return nil
	}

	queryClient := types2.NewQueryClient(client.Context())
	queryResp, err := queryClient.GetPod(ctx, &types2.QueryGetPodRequest{StationId: stationId, PodNumber: podNumber})
	if err != nil {
		//logs.Log.Error("Error fetching VRF: " + err.Error())
		return nil
	}

	return queryResp.Pod
}

func QueryLatestVerifiedBatch() uint64 {
	jsonRpc, stationId, _, _, _, _, err := GetJunctionDetails()
	if err != nil {
		logs.Log.Error("can not get junctionDetails.json data: " + err.Error())
		return 0
	}

	ctx := context.Background()
	client, err := cosmosclient.New(ctx, cosmosclient.WithNodeAddress(jsonRpc))
	if err != nil {
		logs.Log.Error("Client connection error: " + err.Error())
		return 0
	}

	queryClient := types2.NewQueryClient(client.Context())
	queryResp, err := queryClient.GetLatestVerifiedPodNumber(ctx, &types2.QueryGetLatestVerifiedPodNumberRequest{StationId: stationId})
	if err != nil {
		return 0
	}

	return queryResp.PodNumber
}
