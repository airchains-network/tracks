package blocksync

import (
	"context"
	"fmt"
	stationConfig "github.com/airchains-network/decentralized-sequencer/config"
	logs "github.com/airchains-network/decentralized-sequencer/log"
	"github.com/ethereum/go-ethereum/ethclient"
	"log"
	"os"
)

func StartIndexer() {

	StationType := "EVM"
	if StationType == "EVM" {
		// Intialize the Database
		response := InitDb()
		if !response {
			log.Fatal("Error in initializing db")
			os.Exit(0)
		}
		logs.Log.Info("Initialized the database")
		// Connect to the Ethereum client
		client, err := ethclient.Dial(stationConfig.StationRPC)
		fmt.Println(client)
		if err != nil {
			log.Fatal("Failed to connect to the Ethereum client:", err)
			os.Exit(0)
		}
		ctx := context.Background()
		blockDatabaseConnection := GetBlockDbInstance()
		txnDatabaseConnection := GetTxDbInstance()

		StoreEVMBlock(client, ctx, 0, blockDatabaseConnection, txnDatabaseConnection)

	}
}
