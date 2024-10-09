package trackgate

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/airchains-network/tracks/junction/trackgate/types"
	"github.com/ignite/cli/v28/ignite/pkg/cosmosaccount"
	"github.com/ignite/cli/v28/ignite/pkg/cosmosclient"
)

func SchemaCreation(addr string, client cosmosclient.Client, ctx context.Context, account cosmosaccount.Account, stationId string) {

	creator := addr

	version := "1.0.15"

	structDef := types_stellafig.SchemaDef{
		Fields: map[string]interface{}{
			"name":   "string",
			"age":    "int",
			"status": "uint",
			"origin": map[string]interface{}{
				"place": "string",
				"year":  "int",
			},
		},
	}

	randomschemaByte, err := json.Marshal(structDef)
	if err != nil {
		log.Fatal(err)
	}

	msg := &types.MsgSchemaCreation{
		Creator:           creator,
		ExtTrackStationId: stationId,
		Version:           version,
		Schema:            randomschemaByte,
	}

	// Broadcast a transaction from account `charlie` with the message
	// to create a post store response in txResp
	txResp, err := client.BroadcastTx(ctx, account, msg)
	if err != nil {
		fmt.Println("txResp above")
		fmt.Println(txResp)
		fmt.Println("txResp below")
		log.Fatal(err.Error())
	}

	// Print response from broadcasting a transaction
	fmt.Print("MsgCreatePost:\n\n")
	fmt.Println(txResp)

	randomschemaByteInd, err := json.MarshalIndent(msg, "", "    ")
	if err != nil {
		log.Fatalf("Failed to marshal JSON: %v", err)
	}
	fmt.Println("SchemaDetails: ", string(randomschemaByteInd))

	queryClient := types.NewQueryClient(client.Context())

	//RETRIEVE SCHEMA KEY
	queryResp, err := queryClient.RetrieveSchemaKey(ctx, &types.QueryRetrieveSchemaKeyRequest{ExtTrackStationId: stationId, SchemaVersion: version})
	if err != nil {
		log.Fatal(err)
	}

	schemaKey := queryResp.SchemaKey
	//track := queryResp.Track

	fmt.Println("schemaKey : ", schemaKey)
	//trackInd, err := json.MarshalIndent(track, "", "    ")
	//if err != nil {
	//	log.Fatalf("Failed to marshal JSON: %v", err)
	//}
	//fmt.Println("Track: ", string(trackInd))
	//
	//var schemaData internalCodeTypes.FigSchema
	////proverDataErr := json.Unmarshal(track.Schema, &schemaData)
	////if proverDataErr != nil {
	////	log.Fatal(proverDataErr)
	////}
	//schemaDataInd, err := json.MarshalIndent(schemaData, "", "    ")
	//if err != nil {
	//	log.Fatalf("Failed to marshal JSON: %v", err)
	//}
	//fmt.Println("SchemaDetails: ", string(schemaDataInd))

}
