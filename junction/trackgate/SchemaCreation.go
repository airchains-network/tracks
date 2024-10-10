package trackgate

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/airchains-network/tracks/config"
	logs "github.com/airchains-network/tracks/log"
	"os"
	"path/filepath"

	"github.com/airchains-network/tracks/junction/trackgate/types"
	"github.com/ignite/cli/v28/ignite/pkg/cosmosaccount"
	"github.com/ignite/cli/v28/ignite/pkg/cosmosclient"
)

func SchemaCreation(conf *config.Config) bool {

	ctx := context.Background()

	accountName := conf.Junction.AccountName
	accountPath := conf.Junction.AccountPath
	addressPrefix := conf.Junction.AddressPrefix
	junctionRPC := conf.Junction.JunctionRPC

	client, err := cosmosclient.New(ctx, cosmosclient.WithAddressPrefix(addressPrefix), cosmosclient.WithNodeAddress(junctionRPC), cosmosclient.WithHome(accountPath))
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

	version := "1.0.0"
	structDef := SchemaDef{
		Fields: map[string]interface{}{
			"espresso_tx_response_v_1": map[string]interface{}{
				"transaction": map[string]interface{}{
					"namespace": "int",
					"payload":   "string",
				},
				"hash":  "string",
				"index": "int",
				"proof": map[string]interface{}{
					"tx_index":        "string",
					"payload_num_txs": "string",
					"payload_proof_num_txs": map[string]interface{}{
						"proofs":       "string",
						"prefix_bytes": "bytes",
						"suffix_bytes": "bytes",
					},
					"payload_tx_table_entries": "string",
					"payload_proof_tx_table_entries": map[string]interface{}{
						"proofs":       "string",
						"prefix_bytes": "bytes",
						"suffix_bytes": "bytes",
					},
					"payload_proof_tx": map[string]interface{}{
						"proofs":       "string",
						"prefix_bytes": "bytes",
						"suffix_bytes": "bytes",
					},
				},
				"block_hash":   "string",
				"block_height": "int",
			},
			"station_id": "string",
		},
	}

	schemaByte, err := json.Marshal(structDef)
	if err != nil {
		logs.Log.Error(fmt.Sprintf("Error in Marshal of schemaByte: %v", err))
		return false
	}
	stationId := conf.Junction.StationId

	msg := &types.MsgSchemaCreation{
		Creator:           creator,
		ExtTrackStationId: stationId,
		Version:           version,
		Schema:            schemaByte,
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

	queryClient := types.NewQueryClient(client.Context())

	//RETRIEVE SCHEMA KEY
	queryResp, err := queryClient.RetrieveSchemaKey(ctx, &types.QueryRetrieveSchemaKeyRequest{ExtTrackStationId: stationId, SchemaVersion: version})
	if err != nil {
		logs.Log.Error(fmt.Sprintf("Error in retrieving schemaKey: %v", err))

	}

	schemaKey := queryResp.SchemaKey
	//track := queryResp.Track

	fmt.Println("schemaKey : ", schemaKey)

	homeDir, err := os.UserHomeDir()
	if err != nil {
		logs.Log.Error("Error in getting home dir path: " + err.Error())
		return false
	}

	ConfigFilePath := filepath.Join(homeDir, config.DefaultTracksDir, config.DefaultConfigDir, config.DefaultConfigFileName)
	conf.Station.StationSchemaKey = schemaKey

	// Marshal the struct to TOML
	f, err := os.Create(ConfigFilePath)
	if err != nil {
		logs.Log.Error("Error creating file")
		return false
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			logs.Log.Error("Error closing file")
		}
	}(f)
	newData := toml.NewEncoder(f)
	if err := newData.Encode(conf); err != nil {
		logs.Log.Error(fmt.Sprintf("Error encoding config: %v", err))
		return false
	}

	return true

}
